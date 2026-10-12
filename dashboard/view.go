package main

// Live view: a project's tab is streamed into its dashboard card and the mouse
// and keyboard are passed back, so the browser window itself can stay off-screen.
// "Pop out" puts the real window on the screen when you need it (file pickers,
// right-click menus and other native popups aren't part of the stream).

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net"
	"time"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

const (
	hiddenX, hiddenY = -20000, -20000 // where a hidden window is parked
	shownX, shownY   = 80, 60
	winW, winH       = 1280, 900
)

// ---------- window placement ----------

// browserCmd sends one command on the browser's control socket (no page session,
// so it is safe in sign-in mode).
func browserCmd(port int, method string, params, out any) error {
	wsURL, err := browserWS(port)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	conn, _, _, err := ws.Dial(ctx, wsURL)
	if err != nil {
		return err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(3 * time.Second))
	req := map[string]any{"id": 1, "method": method}
	if params != nil {
		req["params"] = params
	}
	b, _ := json.Marshal(req)
	if err := wsutil.WriteClientText(conn, b); err != nil {
		return err
	}
	for {
		b, err := wsutil.ReadServerText(conn)
		if err != nil {
			return err
		}
		var res struct {
			ID     int             `json:"id"`
			Result json.RawMessage `json:"result"`
			Error  *struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if json.Unmarshal(b, &res) != nil || res.ID != 1 {
			continue
		}
		if res.Error != nil {
			return errors.New(res.Error.Message)
		}
		if out != nil {
			return json.Unmarshal(res.Result, out)
		}
		return nil
	}
}

// placeWindow puts the browser window on the screen (show) or parks it off-screen.
func (p *Project) placeWindow(show bool) error {
	p.mu.Lock()
	port, tid := p.port, p.targetID
	p.mu.Unlock()
	if port == 0 {
		return errors.New("browser isn't running")
	}
	if tid == "" {
		pages, err := listPages(port)
		if err != nil || len(pages) == 0 {
			return errors.New("no open tab")
		}
		tid = pages[0].ID
	}
	var w struct {
		WindowID int64 `json:"windowId"`
	}
	if err := browserCmd(port, "Browser.getWindowForTarget", map[string]any{"targetId": tid}, &w); err != nil {
		return err
	}
	set := func(bounds map[string]any) error {
		return browserCmd(port, "Browser.setWindowBounds", map[string]any{"windowId": w.WindowID, "bounds": bounds}, nil)
	}
	_ = set(map[string]any{"windowState": "normal"}) // un-maximize / un-minimize first
	x, y := hiddenX, hiddenY
	if show {
		x, y = shownX, shownY
	}
	if err := set(map[string]any{"left": x, "top": y, "width": winW, "height": winH}); err != nil {
		return err
	}
	p.mu.Lock()
	p.shown = show
	p.mu.Unlock()
	p.m.dirty()
	return nil
}

// ---------- streaming ----------

func (p *Project) curTab() context.Context {
	p.attachMu.Lock()
	defer p.attachMu.Unlock()
	if p.tabCtx != nil && p.tabCtx.Err() == nil {
		return p.tabCtx
	}
	return nil
}

func (p *Project) pubFrame(e *page.EventScreencastFrame) {
	if e.Metadata == nil {
		return
	}
	b, _ := json.Marshal(map[string]any{"w": e.Metadata.DeviceWidth, "h": e.Metadata.DeviceHeight, "d": e.Data})
	p.viewMu.Lock()
	for ch := range p.viewers {
		select {
		case ch <- b:
		default: // slow viewer: replace the stale frame
			select {
			case <-ch:
			default:
			}
			select {
			case ch <- b:
			default:
			}
		}
	}
	p.viewMu.Unlock()
}

type viewIn struct {
	T    string  `json:"t"` // m = mouse, k = key, t = text
	K    string  `json:"k"` // CDP event type
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
	B    string  `json:"b"`
	Bs   int64   `json:"bs"`
	C    int64   `json:"c"`
	M    int64   `json:"m"`
	DX   float64 `json:"dx"`
	DY   float64 `json:"dy"`
	Key  string  `json:"key"`
	Code string  `json:"code"`
	VK   int64   `json:"vk"`
	Text string  `json:"text"`
}

func (p *Project) viewInput(in viewIn) {
	ctx := p.curTab()
	if ctx == nil {
		return
	}
	switch in.T {
	case "m":
		switch in.K {
		case "mousePressed", "mouseReleased", "mouseMoved", "mouseWheel":
		default:
			return
		}
		if in.B == "" {
			in.B = "none"
		}
		a := input.DispatchMouseEvent(input.MouseType(in.K), in.X, in.Y).WithModifiers(input.Modifier(in.M)).
			WithButton(input.MouseButton(in.B)).WithButtons(in.Bs).WithClickCount(in.C)
		if in.K == "mouseWheel" {
			a = a.WithDeltaX(in.DX).WithDeltaY(in.DY)
		}
		_ = run(ctx, 3*time.Second, a)
	case "k":
		switch in.K {
		case "keyDown", "rawKeyDown", "keyUp":
		default:
			return
		}
		a := input.DispatchKeyEvent(input.KeyType(in.K)).WithModifiers(input.Modifier(in.M)).WithKey(in.Key).
			WithCode(in.Code).WithWindowsVirtualKeyCode(in.VK).WithNativeVirtualKeyCode(in.VK)
		if in.Text != "" {
			a = a.WithText(in.Text).WithUnmodifiedText(in.Text)
		}
		_ = run(ctx, 3*time.Second, a)
	case "t":
		if in.Text != "" {
			_ = run(ctx, 5*time.Second, input.InsertText(in.Text))
		}
	}
}

func firstFrame(ctx context.Context) []byte {
	var size []float64 // the picture includes the scrollbar, so measure the window, not the layout
	var img []byte
	err := run(ctx, 5*time.Second, chromedp.Evaluate(`[innerWidth, innerHeight]`, &size),
		chromedp.ActionFunc(func(ctx context.Context) (err error) {
			img, err = page.CaptureScreenshot().WithFormat(page.CaptureScreenshotFormatJpeg).WithQuality(70).Do(ctx)
			return err
		}))
	if err != nil || len(size) != 2 || size[0] == 0 {
		return nil
	}
	b, _ := json.Marshal(map[string]any{"w": size[0], "h": size[1], "d": base64.StdEncoding.EncodeToString(img)})
	return b
}

// serveView streams the tab to one dashboard card until the socket closes.
func (p *Project) serveView(conn net.Conn) {
	defer conn.Close()
	ch := make(chan []byte, 2)
	p.viewMu.Lock()
	if p.viewers == nil {
		p.viewers = map[chan []byte]struct{}{}
	}
	p.viewers[ch] = struct{}{}
	p.viewMu.Unlock()
	done := make(chan struct{})
	defer func() {
		close(done)
		p.viewMu.Lock()
		delete(p.viewers, ch)
		p.viewMu.Unlock()
		p.castMu.Lock()
		p.viewMu.Lock()
		left := len(p.viewers)
		p.viewMu.Unlock()
		if ctx := p.curTab(); left == 0 && ctx != nil {
			_ = run(ctx, 3*time.Second, page.StopScreencast())
		}
		p.castMu.Unlock()
	}()

	go func() {
		// Keeps the stream running on whichever tab is current (it changes when
		// the browser restarts or the tab is replaced), and writes the frames.
		var last context.Context
		start := func() {
			ctx, err := p.tab()
			if err != nil || ctx == last {
				return
			}
			p.castMu.Lock()
			defer p.castMu.Unlock()
			if run(ctx, 5*time.Second, emulation.SetFocusEmulationEnabled(true),
				page.StartScreencast().WithFormat(page.ScreencastFormatJpeg).WithQuality(70).WithMaxWidth(1600).WithMaxHeight(1200)) == nil {
				last = ctx
			}
			// The stream only sends a frame when the page repaints, so a page that
			// is sitting still needs one taken by hand.
			if b := firstFrame(ctx); b != nil {
				select {
				case ch <- b:
				default:
				}
			}
		}
		start()
		t := time.NewTicker(time.Second)
		defer t.Stop()
		for {
			select {
			case <-done:
				return
			case b := <-ch:
				_ = conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
				if wsutil.WriteServerText(conn, b) != nil {
					conn.Close()
					return
				}
			case <-t.C:
				start()
			}
		}
	}()

	for {
		b, _, err := wsutil.ReadClientData(conn)
		if err != nil {
			return
		}
		var in viewIn
		if json.Unmarshal(b, &in) == nil {
			p.viewInput(in)
		}
	}
}
