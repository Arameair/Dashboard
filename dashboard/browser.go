package main

// One Project = one real Chrome/Edge window with its own profile folder.
//
// The browser is started as a normal browser with a local control port
// (--remote-debugging-port). The dashboard watches it over plain HTTP
// (title / URL, no session attached) and only attaches a DevTools session
// when you ask it to do something (send, read, screenshot, focus, tile).
// "Sign in" mode never attaches, so Google sign-in sees an ordinary browser.
//
// If the dashboard is closed, the browsers stay open; the next start
// reconnects to them.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/browser"
	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/cdproto/page"
	cdpruntime "github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

const (
	stStopped  = "stopped"
	stStarting = "starting"
	stRunning  = "running"
	stStopping = "stopping"
	maxLog     = 1000
)

type Project struct {
	m *Manager

	mu        sync.Mutex
	cfg       ProjectCfg
	state     string
	err       string
	signin    bool
	gen       int // bumps on every start; stale goroutines check it and quit
	port      int
	cmd       *exec.Cmd
	title     string
	curURL    string
	targetID  string
	busy      bool // site is generating
	hasInput  bool
	action    string // what a user action is doing right now
	nextNudge time.Time
	lastReply uint64
	logs      []string
	logFile   *os.File

	actMu sync.Mutex // one browser action at a time (user actions wait; background checks skip)

	attachMu  sync.Mutex
	allocCtx  context.Context
	allocStop context.CancelFunc
	tabCtx    context.Context
	tabStop   context.CancelFunc
	tabTarget string
	tabPort   int

	shotMu sync.Mutex
	shot   []byte
	shotAt time.Time

	shown   bool // window is on the screen (otherwise parked off-screen, seen through the dashboard)
	viewMu  sync.Mutex
	castMu  sync.Mutex // start/stop of the stream, so a leaving viewer can't stop a new one
	viewers map[chan []byte]struct{}
}

// ---------- logging ----------

func (p *Project) logf(format string, a ...any) {
	ts := time.Now().Format("15:04:05")
	var lines []string
	for _, l := range strings.Split(strings.TrimRight(fmt.Sprintf(format, a...), "\n"), "\n") {
		lines = append(lines, ts+"  "+strings.TrimRight(l, "\r"))
	}
	p.mu.Lock()
	p.logs = append(p.logs, lines...)
	if len(p.logs) > maxLog {
		p.logs = append([]string(nil), p.logs[len(p.logs)-maxLog:]...)
	}
	f, id := p.logFile, p.cfg.ID
	p.mu.Unlock()
	if f != nil {
		_, _ = f.WriteString(strings.Join(lines, "\n") + "\n")
	}
	p.m.pubLog(id, lines)
}

func (p *Project) clearLog() {
	p.mu.Lock()
	p.logs = nil
	p.mu.Unlock()
}

// ---------- start / stop ----------

func (p *Project) Start(signin bool) error {
	p.mu.Lock()
	if p.state != stStopped {
		s := p.state
		p.mu.Unlock()
		return fmt.Errorf("already %s", s)
	}
	p.state, p.err, p.signin = stStarting, "", signin
	p.gen++
	gen, cfg := p.gen, p.cfg
	p.mu.Unlock()
	p.m.dirty()
	go p.launch(gen, cfg, signin)
	return nil
}

func (p *Project) launch(gen int, cfg ProjectCfg, signin bool) {
	profile := p.m.profileDir(cfg.ID)
	if port := readPort(profile); port > 0 && alive(port) {
		p.logf("browser was already open, reconnected")
		p.up(gen, nil, port)
		return
	}
	exe, err := findBrowser(p.m.browserPath())
	if err != nil {
		p.fail(gen, err)
		return
	}
	if err := os.MkdirAll(profile, 0o700); err != nil {
		p.fail(gen, err)
		return
	}
	_ = os.Remove(filepath.Join(profile, "DevToolsActivePort"))
	start := cfg.URL
	if signin {
		start = signinURL(cfg.URL)
	}
	if start == "" {
		start = "about:blank"
	}
	args := []string{
		"--user-data-dir=" + profile,
		"--remote-debugging-port=0", // Chrome picks a free port and writes it to DevToolsActivePort
		"--no-first-run", "--no-default-browser-check",
		"--disable-background-timer-throttling",
		"--disable-backgrounding-occluded-windows",
		"--disable-renderer-backgrounding",
		"--disable-features=CalculateNativeWinOcclusion",
		fmt.Sprintf("--window-size=%d,%d", winW, winH),
	}
	if signin {
		args = append(args, fmt.Sprintf("--window-position=%d,%d", shownX, shownY))
	} else { // the window stays off-screen; the dashboard shows it
		args = append(args, fmt.Sprintf("--window-position=%d,%d", hiddenX, hiddenY))
	}
	args = append(args, "--new-window", start)
	cmd := exec.Command(exe, args...)
	if err := cmd.Start(); err != nil {
		p.fail(gen, fmt.Errorf("couldn't start %s: %w", filepath.Base(exe), err))
		return
	}
	if signin {
		p.logf("opening %s for sign-in (not automated). Sign in, then click Done.", filepath.Base(exe))
	} else {
		p.logf("opening %s", filepath.Base(exe))
	}
	exited := make(chan error, 1)
	go func() { exited <- cmd.Wait() }()
	deadline := time.Now().Add(30 * time.Second)
	// Edge (and Chrome, when another process owns the profile) relaunches itself:
	// the process we started exits at once and a new one becomes the browser.
	// So an early exit isn't a failure; keep waiting for the control port.
	handedOff := false
	var exitErr error
	for {
		if handedOff {
			time.Sleep(200 * time.Millisecond)
		} else {
			select {
			case exitErr = <-exited:
				handedOff = true
				if d := time.Now().Add(15 * time.Second); d.Before(deadline) {
					deadline = d
				}
			case <-time.After(200 * time.Millisecond):
			}
		}
		if port := readPort(profile); port > 0 && alive(port) {
			if handedOff {
				p.up(gen, nil, port) // monitor() notices when this one closes
			} else if p.up(gen, cmd, port) {
				go func() {
					<-exited
					// A relaunch after the port came up leaves the browser running.
					if port := readPort(profile); port > 0 && alive(port) {
						return
					}
					p.gone(gen, "browser closed")
				}()
			}
			return
		}
		if time.Now().After(deadline) {
			if handedOff {
				p.fail(gen, fmt.Errorf("browser closed right away (%v). Is this profile open in another window?", exitErr))
				return
			}
			_ = killTree(cmd)
			p.fail(gen, errors.New("browser didn't open its control port within 30s"))
			return
		}
	}
}

// up marks the browser running; false if this start was cancelled meanwhile.
func (p *Project) up(gen int, cmd *exec.Cmd, port int) bool {
	p.mu.Lock()
	if p.gen != gen || p.state != stStarting {
		p.mu.Unlock()
		go closeBrowser(port, cmd)
		return false
	}
	p.state, p.port, p.cmd = stRunning, port, cmd
	signin := p.signin
	p.mu.Unlock()
	p.m.dirty()
	p.logf("browser ready")
	_ = p.placeWindow(signin) // also covers a browser that was already open
	go p.monitor(gen)
	return true
}

func (p *Project) fail(gen int, err error) {
	p.mu.Lock()
	if p.gen != gen {
		p.mu.Unlock()
		return
	}
	p.state, p.err, p.signin = stStopped, err.Error(), false
	p.mu.Unlock()
	p.logf("error: %v", err)
	p.m.dirty()
}

// gone cleans up after the browser exits (closed by you, by Stop, or crashed).
func (p *Project) gone(gen int, why string) {
	p.mu.Lock()
	if p.gen != gen || p.state == stStopped {
		p.mu.Unlock()
		return
	}
	p.state, p.port, p.cmd, p.signin, p.shown = stStopped, 0, nil, false, false
	p.title, p.curURL, p.targetID, p.busy, p.hasInput, p.action = "", "", "", false, false, ""
	p.mu.Unlock()
	p.detach()
	p.shotMu.Lock()
	p.shot = nil
	p.shotMu.Unlock()
	p.logf("%s", why)
	p.m.dirty()
}

func (p *Project) Stop() {
	p.mu.Lock()
	switch p.state {
	case stStarting:
		p.gen++ // launch() sees this and closes what it opened
		p.state, p.signin = stStopped, false
		p.mu.Unlock()
		p.m.dirty()
		return
	case stRunning:
	default:
		p.mu.Unlock()
		return
	}
	p.state = stStopping
	gen, port, cmd := p.gen, p.port, p.cmd
	p.mu.Unlock()
	p.m.dirty()
	p.logf("closing browser")
	closeBrowser(port, cmd)
	p.gone(gen, "browser closed")
}

// closeBrowser asks the browser to close (like closing its window), then forces it if needed.
func closeBrowser(port int, cmd *exec.Cmd) {
	if port > 0 {
		if wsURL, err := browserWS(port); err == nil {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			if conn, _, _, err := ws.Dial(ctx, wsURL); err == nil {
				_ = wsutil.WriteClientText(conn, []byte(`{"id":1,"method":"Browser.close"}`))
				_, _ = wsutil.ReadServerText(conn)
				conn.Close()
			}
			cancel()
		}
		for i := 0; i < 30 && alive(port); i++ {
			time.Sleep(200 * time.Millisecond)
		}
		if !alive(port) {
			return
		}
	}
	_ = killTree(cmd)
}

// ---------- watching ----------

type pageTarget struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

var httpc = &http.Client{Timeout: 2 * time.Second}

func listPages(port int) ([]pageTarget, error) {
	res, err := httpc.Get(fmt.Sprintf("http://127.0.0.1:%d/json/list", port))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var all []pageTarget
	if err := json.NewDecoder(res.Body).Decode(&all); err != nil {
		return nil, err
	}
	var pages []pageTarget
	for _, t := range all {
		if t.Type == "page" && !strings.HasPrefix(t.URL, "devtools://") {
			pages = append(pages, t)
		}
	}
	return pages, nil
}

func (p *Project) monitor(gen int) {
	t := time.NewTicker(2 * time.Second)
	defer t.Stop()
	fails, tick := 0, 0
	for range t.C {
		p.mu.Lock()
		if p.gen != gen || p.state != stRunning {
			p.mu.Unlock()
			return
		}
		port, cur := p.port, p.targetID
		p.mu.Unlock()

		pages, err := listPages(port)
		if err != nil {
			if fails++; fails >= 2 {
				p.gone(gen, "browser closed")
				return
			}
			continue
		}
		fails = 0
		if len(pages) == 0 {
			closeBrowser(port, nil)
			p.gone(gen, "all windows closed")
			return
		}
		pick := pages[0]
		for _, pg := range pages {
			if pg.ID == cur {
				pick = pg
			}
		}
		p.mu.Lock()
		if p.targetID != pick.ID || p.title != pick.Title || p.curURL != pick.URL {
			p.targetID, p.title, p.curURL = pick.ID, pick.Title, pick.URL
			p.m.dirty()
		}
		cfg, signin := p.cfg, p.signin
		p.mu.Unlock()

		if signin {
			continue
		}
		if cfg.Console && !p.attached() {
			_, _ = p.tab()
		}
		if tick++; tick%5 == 0 && (cfg.AutoNudge || cfg.SaveReplies) {
			p.watch(cfg)
		}
	}
}

type pageInfo struct {
	Busy  bool   `json:"busy"`
	Input bool   `json:"input"`
	Count int    `json:"count"`
	Last  string `json:"last"`
}

func (p *Project) info(ctx context.Context) (pageInfo, error) {
	s := p.m.selectors()
	js := fmt.Sprintf(`(() => {
  const q = s => { try { return [...document.querySelectorAll(s)] } catch (e) { return [] } };
  const vis = el => !!(el.offsetWidth || el.offsetHeight || el.getClientRects().length);
  const r = q(%s);
  return { busy: q(%s).some(vis), input: q(%s).some(vis), count: r.length,
           last: r.length ? (r[r.length - 1].innerText || "").trim() : "" };
})()`, jsStr(s.Reply), jsStr(s.Busy), jsStr(s.Input))
	var pi pageInfo
	tctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	err := chromedp.Run(tctx, chromedp.Evaluate(js, &pi))
	if err == nil {
		p.mu.Lock()
		if p.busy != pi.Busy || p.hasInput != pi.Input {
			p.busy, p.hasInput = pi.Busy, pi.Input
			p.m.dirty()
		}
		p.mu.Unlock()
	}
	return pi, err
}

// watch: background check every ~10s. Saves new replies and sends due auto-nudges.
func (p *Project) watch(cfg ProjectCfg) {
	if !p.actMu.TryLock() {
		return // a user action is running
	}
	defer p.actMu.Unlock()
	ctx, err := p.tab()
	if err != nil {
		return
	}
	pi, err := p.info(ctx)
	if err != nil {
		return
	}
	if cfg.SaveReplies && !pi.Busy && pi.Last != "" {
		p.saveIfNew(pi.Last)
	}
	p.mu.Lock()
	due := cfg.AutoNudge && !p.nextNudge.IsZero() && time.Now().After(p.nextNudge)
	p.mu.Unlock()
	if !due {
		return
	}
	next := time.Duration(cfg.NudgeMin) * time.Minute
	switch {
	case pi.Busy:
		p.logf("auto-nudge: still working, trying again in 1 min")
		next = time.Minute
	case !pi.Input:
		p.logf("auto-nudge: no message box on this page (signed out?), skipped")
	default:
		if err := p.send(ctx, cfg.Nudge); err != nil {
			p.logf("auto-nudge failed: %v", err)
		} else {
			p.logf("auto-nudge sent")
		}
	}
	p.mu.Lock()
	p.nextNudge = time.Now().Add(next)
	p.mu.Unlock()
	p.m.dirty()
}

func (p *Project) saveIfNew(txt string) {
	h := fnv(txt)
	p.mu.Lock()
	if h == p.lastReply {
		p.mu.Unlock()
		return
	}
	p.lastReply = h
	cfg := p.cfg
	p.mu.Unlock()
	dir := p.m.repliesDir(cfg.ID)
	_ = os.MkdirAll(dir, 0o700)
	_ = os.WriteFile(filepath.Join(dir, ".last"), []byte(strconv.FormatUint(h, 16)), 0o600)
	now := time.Now()
	f := filepath.Join(dir, now.Format("2006-01-02_150405")+".md")
	body := fmt.Sprintf("# %s\n\n- saved: %s\n- page: %s\n\n%s\n", cfg.Name, now.Format(time.RFC3339), p.snapshotURL(), txt)
	if err := os.WriteFile(f, []byte(body), 0o600); err != nil {
		p.logf("couldn't save reply: %v", err)
		return
	}
	p.logf("new reply saved (%d chars): %s", len([]rune(txt)), short(oneLine(txt), 120))
}

func (p *Project) snapshotURL() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.curURL
}

// ---------- attaching ----------

func (p *Project) attached() bool {
	p.attachMu.Lock()
	defer p.attachMu.Unlock()
	return p.tabCtx != nil && p.tabCtx.Err() == nil
}

// tab returns a DevTools session on the project's tab, attaching if needed.
func (p *Project) tab() (context.Context, error) {
	p.mu.Lock()
	state, signin, port, want := p.state, p.signin, p.port, p.targetID
	p.mu.Unlock()
	if state != stRunning {
		return nil, errors.New("browser isn't running")
	}
	if signin {
		return nil, errors.New("in sign-in mode: click Done after signing in")
	}
	if want == "" {
		pages, err := listPages(port)
		if err != nil || len(pages) == 0 {
			return nil, errors.New("no open tab")
		}
		want = pages[0].ID
	}
	p.attachMu.Lock()
	defer p.attachMu.Unlock()
	if p.tabCtx != nil && p.tabCtx.Err() == nil && p.tabTarget == want && p.tabPort == port {
		return p.tabCtx, nil
	}
	// The old tab is gone (or the browser restarted). Cancelling a chromedp tab
	// context closes that tab, which is fine here: it no longer exists.
	if p.tabStop != nil {
		go p.tabStop()
		p.tabCtx, p.tabStop = nil, nil
	}
	if p.allocStop != nil && p.tabPort != port {
		p.allocStop()
		p.allocCtx, p.allocStop = nil, nil
	}
	if p.allocCtx == nil {
		wsURL, err := browserWS(port)
		if err != nil {
			return nil, err
		}
		p.allocCtx, p.allocStop = chromedp.NewRemoteAllocator(context.Background(), wsURL, chromedp.NoModifyURL)
	}
	ctx, stop := chromedp.NewContext(p.allocCtx, chromedp.WithTargetID(target.ID(want)))
	done := make(chan error, 1)
	go func() { done <- chromedp.Run(ctx) }() // first Run must not carry a timeout (it owns the connection)
	select {
	case err := <-done:
		if err != nil {
			stop()
			return nil, fmt.Errorf("couldn't attach: %w", err)
		}
	case <-time.After(10 * time.Second):
		go stop()
		return nil, errors.New("couldn't attach: timed out")
	}
	chromedp.ListenTarget(ctx, func(ev any) { p.onEvent(ctx, ev) })
	p.tabCtx, p.tabStop, p.tabTarget, p.tabPort = ctx, stop, want, port
	return ctx, nil
}

func (p *Project) detach() {
	p.attachMu.Lock()
	defer p.attachMu.Unlock()
	if p.tabStop != nil {
		go p.tabStop()
	}
	if p.allocStop != nil {
		go p.allocStop()
	}
	p.tabCtx, p.tabStop, p.allocCtx, p.allocStop, p.tabTarget, p.tabPort = nil, nil, nil, nil, "", 0
}

func (p *Project) onEvent(ctx context.Context, ev any) {
	if e, ok := ev.(*page.EventScreencastFrame); ok {
		go func() { _ = run(ctx, 3*time.Second, page.ScreencastFrameAck(e.SessionID)) }()
		p.pubFrame(e)
		return
	}
	p.mu.Lock()
	on := p.cfg.Console
	p.mu.Unlock()
	if !on {
		return
	}
	switch e := ev.(type) {
	case *cdpruntime.EventConsoleAPICalled:
		var parts []string
		for _, a := range e.Args {
			switch {
			case len(a.Value) > 0:
				var s string
				if json.Unmarshal(a.Value, &s) == nil {
					parts = append(parts, s)
				} else {
					parts = append(parts, string(a.Value))
				}
			case a.Description != "":
				parts = append(parts, a.Description)
			default:
				parts = append(parts, string(a.Type))
			}
		}
		go p.logf("console.%s: %s", e.Type, short(strings.Join(parts, " "), 600))
	case *cdpruntime.EventExceptionThrown:
		d := e.ExceptionDetails
		msg := d.Text
		if d.Exception != nil && d.Exception.Description != "" {
			msg = d.Exception.Description
		}
		go p.logf("page error: %s", short(oneLine(msg), 600))
	}
}

// ---------- actions ----------

// Do runs one dashboard action and returns text for the dashboard (or an error).
func (p *Project) Do(a, arg string) (res string, err error) {
	switch a {
	case "start":
		return "", p.Start(false)
	case "signin":
		return "", p.Start(true)
	case "stop":
		go p.Stop()
		return "", nil
	case "done":
		p.mu.Lock()
		was := p.signin
		p.signin = false
		p.mu.Unlock()
		if was {
			p.logf("sign-in mode off: the dashboard can control this browser now")
			p.m.dirty()
			_ = p.placeWindow(false)
		}
		return "", nil
	case "clear":
		p.clearLog()
		return "", nil
	}

	p.actMu.Lock()
	defer p.actMu.Unlock()
	p.setAction(a)
	defer p.setAction("")
	defer func() {
		if err != nil {
			p.logf("%s failed: %v", a, err)
		}
	}()
	ctx, err := p.tab()
	if err != nil {
		return "", err
	}
	p.mu.Lock()
	cfg := p.cfg
	p.mu.Unlock()

	switch a {
	case "hide":
		return "", p.placeWindow(false)
	case "show":
		if err := p.placeWindow(true); err != nil {
			return "", err
		}
		return "", run(ctx, 5*time.Second, chromedp.ActionFunc(func(ctx context.Context) error {
			if id, b, err := browser.GetWindowForTarget().Do(ctx); err == nil && b.WindowState == browser.WindowStateMinimized {
				_ = browser.SetWindowBounds(id, &browser.Bounds{WindowState: browser.WindowStateNormal}).Do(ctx)
			}
			return page.BringToFront().Do(ctx)
		}))
	case "reload":
		return "", run(ctx, 5*time.Second, chromedp.ActionFunc(func(ctx context.Context) error {
			return page.Reload().Do(ctx)
		}))
	case "back":
		return "", run(ctx, 5*time.Second, chromedp.Evaluate(`history.back()`, nil))
	case "home", "goto":
		u := cfg.URL
		if a == "goto" {
			u = arg
		}
		u, err := normURL(u)
		if err != nil {
			return "", err
		}
		p.logf("go to %s", u)
		return "", run(ctx, 15*time.Second, chromedp.ActionFunc(func(ctx context.Context) error {
			_, _, txt, _, err := page.Navigate(u).Do(ctx)
			if err == nil && txt != "" {
				err = errors.New(txt)
			}
			return err
		}))
	case "send", "nudge":
		msg := arg
		if a == "nudge" {
			msg = cfg.Nudge
		}
		if strings.TrimSpace(msg) == "" {
			return "", errors.New("nothing to send")
		}
		pi, err := p.info(ctx)
		if err != nil {
			return "", err
		}
		if pi.Busy {
			return "", errors.New("the page is still generating; try again when it finishes")
		}
		if err := p.send(ctx, msg); err != nil {
			return "", err
		}
		p.logf("sent: %s", short(oneLine(msg), 140))
		if a == "nudge" && cfg.AutoNudge {
			p.mu.Lock()
			p.nextNudge = time.Now().Add(time.Duration(cfg.NudgeMin) * time.Minute)
			p.mu.Unlock()
			p.m.dirty()
		}
		return "sent", nil
	case "read":
		pi, err := p.info(ctx)
		if err != nil {
			return "", err
		}
		if pi.Last == "" {
			return "", errors.New("no reply found on this page")
		}
		if pi.Busy {
			p.logf("latest reply (still generating):\n%s", pi.Last)
		} else {
			p.logf("latest reply:\n%s", pi.Last)
			p.saveIfNew(pi.Last)
		}
		return pi.Last, nil
	}
	return "", fmt.Errorf("unknown action %q", a)
}

func (p *Project) setAction(a string) {
	p.mu.Lock()
	p.action = a
	p.mu.Unlock()
	p.m.dirty()
}

func run(ctx context.Context, d time.Duration, acts ...chromedp.Action) error {
	tctx, cancel := context.WithTimeout(ctx, d)
	defer cancel()
	return chromedp.Run(tctx, acts...)
}

// send types msg into the message box, checks it landed exactly, and presses Enter.
func (p *Project) send(ctx context.Context, msg string) error {
	sel := p.m.selectors().Input
	clear := fmt.Sprintf(`(() => { const el = [...document.querySelectorAll(%s)].find(e => e.offsetWidth || e.offsetHeight);
  if (!el) return false; el.focus(); if ('value' in el && el.tagName !== 'DIV') { el.select(); } else { document.execCommand('selectAll'); }
  document.execCommand('delete'); return true; })()`, jsStr(sel))
	read := fmt.Sprintf(`(() => { const el = [...document.querySelectorAll(%s)].find(e => e.offsetWidth || e.offsetHeight);
  return el ? ('value' in el && el.tagName !== 'DIV' ? el.value : el.innerText) : null; })()`, jsStr(sel))
	var ok bool
	var got *string
	err := run(ctx, 20*time.Second,
		chromedp.WaitVisible(sel, chromedp.ByQuery),
		chromedp.Evaluate(clear, &ok),
		chromedp.ActionFunc(func(ctx context.Context) error {
			if !ok {
				return errors.New("no message box on this page")
			}
			return input.InsertText(msg).Do(ctx)
		}),
		chromedp.Sleep(300*time.Millisecond),
		chromedp.Evaluate(read, &got),
	)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return errors.New("no message box on this page (signed out or still loading?)")
		}
		return err
	}
	if got == nil || squash(*got) != squash(msg) {
		_ = run(ctx, 3*time.Second, chromedp.Evaluate(clear, &ok))
		return errors.New("the message box didn't take the text exactly; cleared it, nothing sent")
	}
	return run(ctx, 5*time.Second, chromedp.KeyEvent("\r"))
}

// Shot returns a small JPEG of the tab (cached for 2s).
func (p *Project) Shot() ([]byte, error) {
	p.shotMu.Lock()
	defer p.shotMu.Unlock()
	if p.shot != nil && time.Since(p.shotAt) < 2*time.Second {
		return p.shot, nil
	}
	ctx, err := p.tab()
	if err != nil {
		return nil, err
	}
	var buf []byte
	err = run(ctx, 5*time.Second, chromedp.ActionFunc(func(ctx context.Context) error {
		_, _, _, _, vv, _, err := page.GetLayoutMetrics().Do(ctx)
		if err != nil {
			return err
		}
		scale := 1.0
		if vv.ClientWidth > 560 {
			scale = 560 / vv.ClientWidth
		}
		buf, err = page.CaptureScreenshot().WithFormat(page.CaptureScreenshotFormatJpeg).WithQuality(60).
			WithClip(&page.Viewport{X: vv.PageX, Y: vv.PageY, Width: vv.ClientWidth, Height: vv.ClientHeight, Scale: scale}).Do(ctx)
		return err
	}))
	if err != nil {
		return nil, err
	}
	p.shot, p.shotAt = buf, time.Now()
	return buf, nil
}

// ---------- small helpers ----------

func readPort(profile string) int {
	b, err := os.ReadFile(filepath.Join(profile, "DevToolsActivePort"))
	if err != nil {
		return 0
	}
	n, _ := strconv.Atoi(strings.TrimSpace(strings.SplitN(string(b), "\n", 2)[0]))
	return n
}

func alive(port int) bool {
	res, err := httpc.Get(fmt.Sprintf("http://127.0.0.1:%d/json/version", port))
	if err != nil {
		return false
	}
	res.Body.Close()
	return res.StatusCode == 200
}

func browserWS(port int) (string, error) {
	res, err := httpc.Get(fmt.Sprintf("http://127.0.0.1:%d/json/version", port))
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	var v struct {
		WS string `json:"webSocketDebuggerUrl"`
	}
	if err := json.NewDecoder(res.Body).Decode(&v); err != nil || v.WS == "" {
		return "", errors.New("browser didn't report a control address")
	}
	return v.WS, nil
}

func signinURL(u string) string {
	if strings.Contains(u, "chatgpt.com") {
		return "https://chatgpt.com/auth/login"
	}
	return u
}

func normURL(u string) (string, error) {
	u = strings.TrimSpace(u)
	if u == "" {
		return "", errors.New("no address")
	}
	if !strings.Contains(u, "://") && !strings.HasPrefix(u, "about:") {
		host := strings.SplitN(strings.SplitN(u, "/", 2)[0], ":", 2)[0]
		if host == "localhost" || net.ParseIP(host) != nil {
			u = "http://" + u
		} else {
			u = "https://" + u
		}
	}
	pu, err := url.Parse(u)
	if err != nil || (pu.Scheme != "http" && pu.Scheme != "https" && pu.Scheme != "about" && pu.Scheme != "file") {
		return "", fmt.Errorf("not a web address: %s", u)
	}
	return u, nil
}

func jsStr(s string) string { b, _ := json.Marshal(s); return string(b) }

func squash(s string) string { return strings.Join(strings.Fields(s), " ") }

func oneLine(s string) string { return squash(s) }

func short(s string, n int) string {
	if r := []rune(s); len(r) > n {
		return string(r[:n]) + "…"
	}
	return s
}

func fnv(s string) uint64 {
	h := uint64(14695981039346656037)
	for i := 0; i < len(s); i++ {
		h ^= uint64(s[i])
		h *= 1099511628211
	}
	return h
}
