// Dashboard v5: control several real browsers from one page.
//
// Each project is its own Chrome/Edge window with its own profile (so each can
// stay signed in to a different chat). From the dashboard you can start/stop
// them, see what each is showing, navigate, send a message, read the latest
// reply, auto-nudge on a timer, tile the windows, and show/hide each log.
//
// Everything stays on this computer: the page is served on 127.0.0.1 only.
package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/chromedp/cdproto/browser"
	"github.com/chromedp/chromedp"
)

//go:embed index.html
var indexHTML []byte

const version = "v5"

// ---------- event hub (server-sent events) ----------

type Hub struct {
	mu   sync.Mutex
	subs map[chan []byte]struct{}
}

func (h *Hub) sub() chan []byte {
	c := make(chan []byte, 512)
	h.mu.Lock()
	h.subs[c] = struct{}{}
	h.mu.Unlock()
	return c
}

func (h *Hub) unsub(c chan []byte) {
	h.mu.Lock()
	delete(h.subs, c)
	h.mu.Unlock()
}

func (h *Hub) pub(b []byte) {
	h.mu.Lock()
	for c := range h.subs {
		select {
		case c <- b:
		default: // slow client: drop rather than block the browsers
		}
	}
	h.mu.Unlock()
}

// ---------- manager ----------

type Manager struct {
	mu       sync.Mutex
	cfg      *Config
	cfgPath  string
	dataDir  string
	projects map[string]*Project
	order    []string
	hub      *Hub
	isDirty  atomic.Bool
}

func (m *Manager) dirty() { m.isDirty.Store(true) }

func (m *Manager) pubLog(id string, lines []string) {
	b, _ := json.Marshal(map[string]any{"t": "log", "id": id, "lines": lines})
	m.hub.pub(b)
}

func (m *Manager) profileDir(id string) string { return filepath.Join(m.dataDir, "profiles", id) }
func (m *Manager) repliesDir(id string) string { return filepath.Join(m.dataDir, "replies", id) }

func (m *Manager) browserPath() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.cfg.BrowserPath
}

func (m *Manager) selectors() Selectors {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.cfg.Selectors
}

func (m *Manager) get(id string) *Project {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.projects[id]
}

func (m *Manager) list() []*Project {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]*Project, 0, len(m.order))
	for _, id := range m.order {
		out = append(out, m.projects[id])
	}
	return out
}

func (m *Manager) add(pc ProjectCfg) *Project {
	p := &Project{m: m, cfg: pc, state: stStopped}
	_ = os.MkdirAll(filepath.Join(m.dataDir, "logs"), 0o700)
	lp := filepath.Join(m.dataDir, "logs", pc.ID+".log")
	if fi, err := os.Stat(lp); err == nil && fi.Size() > 5<<20 {
		_ = os.Rename(lp, lp+".old")
	}
	p.logFile, _ = os.OpenFile(lp, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if b, err := os.ReadFile(filepath.Join(m.repliesDir(pc.ID), ".last")); err == nil {
		p.lastReply, _ = strconv.ParseUint(strings.TrimSpace(string(b)), 16, 64)
	}
	if pc.AutoNudge {
		p.nextNudge = time.Now().Add(time.Duration(pc.NudgeMin) * time.Minute)
	}
	m.projects[pc.ID] = p
	m.order = append(m.order, pc.ID)
	return p
}

// saveLocked writes config.json from the live projects. Caller holds m.mu.
func (m *Manager) saveLocked() error {
	m.cfg.Projects = m.cfg.Projects[:0]
	for _, id := range m.order {
		p := m.projects[id]
		p.mu.Lock()
		m.cfg.Projects = append(m.cfg.Projects, p.cfg)
		p.mu.Unlock()
	}
	return saveConfig(m.cfgPath, m.cfg)
}

func (m *Manager) upsert(in ProjectCfg) (string, error) {
	in.Name = strings.TrimSpace(in.Name)
	in.URL = strings.TrimSpace(in.URL)
	if in.Name == "" {
		return "", errors.New("give the project a name")
	}
	if in.URL != "" {
		u, err := normURL(in.URL)
		if err != nil {
			return "", err
		}
		in.URL = u
	}
	if in.NudgeMin <= 0 {
		in.NudgeMin = 10
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	p := m.projects[in.ID]
	if p == nil { // new
		base := slug(in.Name)
		if base == "" {
			base = "project"
		}
		id := base
		for i := 2; m.projects[id] != nil; i++ {
			id = fmt.Sprintf("%s-%d", base, i)
		}
		in.ID = id
		if in.Color == "" {
			in.Color = palette[len(m.order)%len(palette)]
		}
		m.add(in)
	} else {
		p.mu.Lock()
		if in.AutoNudge && (!p.cfg.AutoNudge || p.cfg.NudgeMin != in.NudgeMin) {
			p.nextNudge = time.Now().Add(time.Duration(in.NudgeMin) * time.Minute)
		}
		if !in.AutoNudge {
			p.nextNudge = time.Time{}
		}
		p.cfg = in
		p.mu.Unlock()
	}
	m.dirty()
	return in.ID, m.saveLocked()
}

func (m *Manager) remove(id string) error {
	p := m.get(id)
	if p == nil {
		return errors.New("no such project")
	}
	p.Stop()
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.projects, id)
	for i, x := range m.order {
		if x == id {
			m.order = append(m.order[:i], m.order[i+1:]...)
			break
		}
	}
	if p.logFile != nil {
		p.logFile.Close()
	}
	m.dirty()
	return m.saveLocked()
}

func (m *Manager) move(id string, delta int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, x := range m.order {
		if x == id {
			j := i + delta
			if j < 0 || j >= len(m.order) {
				return nil
			}
			m.order[i], m.order[j] = m.order[j], m.order[i]
			m.dirty()
			return m.saveLocked()
		}
	}
	return errors.New("no such project")
}

// tile arranges every controllable browser window in a grid on the screen.
func (m *Manager) tile() error {
	var ps []*Project
	var ctxs []context.Context
	for _, p := range m.list() {
		p.mu.Lock()
		ok := p.state == stRunning && !p.signin
		p.mu.Unlock()
		if !ok {
			continue
		}
		if ctx, err := p.tab(); err == nil {
			ps, ctxs = append(ps, p), append(ctxs, ctx)
		}
	}
	if len(ps) == 0 {
		return errors.New("no running browsers to arrange")
	}
	var scr struct{ W, H, X, Y float64 }
	if err := run(ctxs[0], 5*time.Second, chromedp.Evaluate(
		`({W: screen.availWidth, H: screen.availHeight, X: screen.availLeft || 0, Y: screen.availTop || 0})`, &scr)); err != nil {
		return err
	}
	n := len(ps)
	cols := int(math.Ceil(math.Sqrt(float64(n))))
	rows := int(math.Ceil(float64(n) / float64(cols)))
	w, h := int64(scr.W)/int64(cols), int64(scr.H)/int64(rows)
	for i, ctx := range ctxs {
		b := &browser.Bounds{Left: int64(scr.X) + int64(i%cols)*w, Top: int64(scr.Y) + int64(i/cols)*h, Width: w, Height: h}
		err := run(ctx, 5*time.Second, chromedp.ActionFunc(func(ctx context.Context) error {
			id, _, err := browser.GetWindowForTarget().Do(ctx)
			if err != nil {
				return err
			}
			_ = browser.SetWindowBounds(id, &browser.Bounds{WindowState: browser.WindowStateNormal}).Do(ctx)
			return browser.SetWindowBounds(id, b).Do(ctx)
		}))
		if err != nil {
			ps[i].logf("tile: %v", err)
		}
	}
	return nil
}

// ---------- state for the page ----------

type projView struct {
	ProjectCfg
	State     string `json:"state"`
	Err       string `json:"err"`
	Signin    bool   `json:"signin"`
	Title     string `json:"title"`
	CurURL    string `json:"cur_url"`
	Busy      bool   `json:"busy"`
	HasInput  bool   `json:"has_input"`
	Action    string `json:"action"`
	NextNudge int64  `json:"next_nudge"` // unix ms, 0 = none
}

func (m *Manager) stateJSON() []byte {
	var views []projView
	for _, p := range m.list() {
		p.mu.Lock()
		v := projView{ProjectCfg: p.cfg, State: p.state, Err: p.err, Signin: p.signin, Title: p.title,
			CurURL: p.curURL, Busy: p.busy, HasInput: p.hasInput, Action: p.action}
		if p.cfg.AutoNudge && !p.nextNudge.IsZero() {
			v.NextNudge = p.nextNudge.UnixMilli()
		}
		p.mu.Unlock()
		views = append(views, v)
	}
	m.mu.Lock()
	settings := map[string]any{"browser_path": m.cfg.BrowserPath, "selectors": m.cfg.Selectors}
	m.mu.Unlock()
	found, ferr := findBrowser(m.browserPath())
	settings["browser_found"] = found
	if ferr != nil {
		settings["browser_error"] = ferr.Error()
	}
	b, _ := json.Marshal(map[string]any{"t": "state", "projects": views, "settings": settings, "data": m.dataDir, "version": version})
	return b
}

// ---------- http ----------

func main() {
	port := flag.Int("port", 8777, "dashboard port on 127.0.0.1 (0 = any free port)")
	data := flag.String("data", "", "data folder (default: your local app data\\TeamsDashboard\\v5)")
	noOpen := flag.Bool("no-open", false, "don't open the dashboard in your browser")
	defs := flag.String("projects", "", "project definition files or folders, separated by ; (default: a projects folder next to the program)")
	flag.Parse()
	log.SetFlags(log.Ltime)

	dataDir := *data
	if dataDir == "" {
		base := os.Getenv("LOCALAPPDATA")
		if base == "" {
			base, _ = os.UserCacheDir()
		}
		dataDir = filepath.Join(base, "TeamsDashboard", version)
	}
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		log.Fatal(err)
	}
	cfgPath := filepath.Join(dataDir, "config.json")
	cfg, err := loadConfig(cfgPath)
	if err != nil {
		fmt.Println("Problem reading settings:", err)
		fmt.Println("Press Enter to close.")
		fmt.Scanln()
		os.Exit(1)
	}

	if *defs == "" {
		if exe, err := os.Executable(); err == nil {
			*defs = filepath.Join(filepath.Dir(exe), "projects")
		}
	}
	added, errs := mergeDefinitions(cfg, *defs)
	for _, err := range errs {
		fmt.Println("Skipped a project definition file:", err)
	}
	if len(added) > 0 {
		fmt.Println("Added from project definitions:", strings.Join(added, ", "))
		if err := saveConfig(cfgPath, cfg); err != nil {
			fmt.Println("Couldn't save settings:", err)
		}
	}

	m := &Manager{cfg: cfg, cfgPath: cfgPath, dataDir: dataDir, projects: map[string]*Project{}, hub: &Hub{subs: map[chan []byte]struct{}{}}}
	for _, pc := range cfg.Projects {
		m.add(pc)
	}

	// Listen first, so a second copy just opens the running dashboard.
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", *port))
	if err != nil {
		if res, e := httpc.Get(fmt.Sprintf("http://127.0.0.1:%d/api/ping", *port)); e == nil {
			b, _ := io.ReadAll(res.Body)
			res.Body.Close()
			if strings.HasPrefix(string(b), "teams-dashboard") {
				fmt.Println("The dashboard is already running; opening it.")
				_ = openURL(fmt.Sprintf("http://127.0.0.1:%d/", *port))
				return
			}
		}
		if ln, err = net.Listen("tcp", "127.0.0.1:0"); err != nil {
			log.Fatal(err)
		}
	}
	addr := ln.Addr().String()
	pageURL := "http://" + addr + "/"

	// Reconnect to browsers left open last time, and start auto-start projects.
	for _, p := range m.list() {
		p.mu.Lock()
		pc := p.cfg
		p.mu.Unlock()
		if port := readPort(m.profileDir(pc.ID)); (port > 0 && alive(port)) || pc.AutoStart {
			_ = p.Start(false)
		}
	}

	// Push state to the page whenever something changed (at most ~3x a second).
	go func() {
		for range time.Tick(300 * time.Millisecond) {
			if m.isDirty.Swap(false) {
				m.hub.pub(m.stateJSON())
			}
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write(indexHTML)
	})
	mux.HandleFunc("GET /api/ping", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "teams-dashboard "+version)
	})
	mux.HandleFunc("GET /api/state", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(m.stateJSON())
	})
	mux.HandleFunc("GET /api/events", func(w http.ResponseWriter, r *http.Request) {
		fl, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", 500)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-store")
		ch := m.hub.sub()
		defer m.hub.unsub(ch)
		logs := map[string][]string{}
		for _, p := range m.list() {
			p.mu.Lock()
			logs[p.cfg.ID] = append([]string(nil), p.logs...)
			p.mu.Unlock()
		}
		first, _ := json.Marshal(map[string]any{"t": "logs", "logs": logs})
		fmt.Fprintf(w, "data: %s\n\ndata: %s\n\n", m.stateJSON(), first)
		fl.Flush()
		keep := time.NewTicker(20 * time.Second)
		defer keep.Stop()
		for {
			select {
			case b := <-ch:
				fmt.Fprintf(w, "data: %s\n\n", b)
				for n := len(ch); n > 0; n-- { // batch whatever queued up
					fmt.Fprintf(w, "data: %s\n\n", <-ch)
				}
				fl.Flush()
			case <-keep.C:
				io.WriteString(w, ": keepalive\n\n")
				fl.Flush()
			case <-r.Context().Done():
				return
			}
		}
	})
	mux.HandleFunc("GET /api/shot", func(w http.ResponseWriter, r *http.Request) {
		p := m.get(r.URL.Query().Get("id"))
		if p == nil {
			http.NotFound(w, r)
			return
		}
		b, err := p.Shot()
		if err != nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.Header().Set("Content-Type", "image/jpeg")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write(b)
	})
	mux.HandleFunc("GET /api/log", func(w http.ResponseWriter, r *http.Request) {
		p := m.get(r.URL.Query().Get("id"))
		if p == nil {
			http.NotFound(w, r)
			return
		}
		p.mu.Lock()
		txt := strings.Join(p.logs, "\n")
		p.mu.Unlock()
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		io.WriteString(w, txt)
	})

	reply := func(w http.ResponseWriter, v any, err error) {
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(400)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "result": v})
	}
	body := func(r *http.Request, v any) error {
		if err := json.NewDecoder(io.LimitReader(r.Body, 256<<10)).Decode(v); err != nil && err != io.EOF {
			return fmt.Errorf("bad request: %w", err)
		}
		return nil
	}

	mux.HandleFunc("POST /api/act", func(w http.ResponseWriter, r *http.Request) {
		p := m.get(r.URL.Query().Get("id"))
		if p == nil {
			reply(w, nil, errors.New("no such project"))
			return
		}
		var in struct{ Arg string }
		if err := body(r, &in); err != nil {
			reply(w, nil, err)
			return
		}
		res, err := p.Do(r.URL.Query().Get("a"), in.Arg)
		reply(w, res, err)
	})
	mux.HandleFunc("POST /api/all", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("a") {
		case "start":
			for _, p := range m.list() {
				_ = p.Start(false)
			}
		case "stop":
			for _, p := range m.list() {
				go p.Stop()
			}
		case "tile":
			reply(w, nil, m.tile())
			return
		default:
			reply(w, nil, errors.New("unknown action"))
			return
		}
		reply(w, nil, nil)
	})
	mux.HandleFunc("POST /api/project", func(w http.ResponseWriter, r *http.Request) {
		var in ProjectCfg
		if err := body(r, &in); err != nil {
			reply(w, nil, err)
			return
		}
		id, err := m.upsert(in)
		reply(w, id, err)
	})
	mux.HandleFunc("POST /api/project/delete", func(w http.ResponseWriter, r *http.Request) {
		reply(w, nil, m.remove(r.URL.Query().Get("id")))
	})
	mux.HandleFunc("POST /api/project/move", func(w http.ResponseWriter, r *http.Request) {
		d, _ := strconv.Atoi(r.URL.Query().Get("d"))
		reply(w, nil, m.move(r.URL.Query().Get("id"), d))
	})
	mux.HandleFunc("POST /api/settings", func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			BrowserPath string    `json:"browser_path"`
			Selectors   Selectors `json:"selectors"`
		}
		if err := body(r, &in); err != nil {
			reply(w, nil, err)
			return
		}
		in.BrowserPath = strings.Trim(strings.TrimSpace(in.BrowserPath), `"`)
		if in.BrowserPath != "" {
			if _, err := findBrowser(in.BrowserPath); err != nil {
				reply(w, nil, err)
				return
			}
		}
		m.mu.Lock()
		m.cfg.BrowserPath, m.cfg.Selectors = in.BrowserPath, in.Selectors
		m.cfg.normalize()
		err := m.saveLocked()
		m.mu.Unlock()
		m.dirty()
		reply(w, nil, err)
	})
	mux.HandleFunc("POST /api/open", func(w http.ResponseWriter, r *http.Request) {
		dir := m.dataDir
		if id := r.URL.Query().Get("id"); id != "" {
			if m.get(id) == nil {
				reply(w, nil, errors.New("no such project"))
				return
			}
			dir = m.repliesDir(id)
		}
		_ = os.MkdirAll(dir, 0o700)
		reply(w, dir, openPath(dir))
	})
	mux.HandleFunc("POST /api/quit", func(w http.ResponseWriter, r *http.Request) {
		closeAll := r.URL.Query().Get("close") == "1"
		reply(w, nil, nil)
		go func() {
			if closeAll {
				var wg sync.WaitGroup
				for _, p := range m.list() {
					wg.Add(1)
					go func(p *Project) { defer wg.Done(); p.Stop() }(p)
				}
				wg.Wait()
			}
			fmt.Println("Dashboard closed.")
			time.Sleep(200 * time.Millisecond)
			os.Exit(0)
		}()
	})

	// Only this computer's own pages may talk to the dashboard.
	guard := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, port, _ := net.SplitHostPort(addr)
		if r.Host != addr && r.Host != "localhost:"+port {
			http.Error(w, "forbidden", 403)
			return
		}
		if r.Method != http.MethodGet {
			if o := r.Header.Get("Origin"); o != "" && o != "http://"+r.Host {
				http.Error(w, "forbidden", 403)
				return
			}
		}
		mux.ServeHTTP(w, r)
	})

	fmt.Println("Dashboard", version, "->", pageURL)
	fmt.Println("Settings, profiles, logs and saved replies:", dataDir)
	fmt.Println("Keep this window open while you use the dashboard. Closing it leaves the browsers open;")
	fmt.Println("the next start reconnects to them.")
	if !*noOpen {
		_ = openURL(pageURL)
	}
	log.Fatal(http.Serve(ln, guard))
}
