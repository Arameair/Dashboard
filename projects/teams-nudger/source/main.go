// ChatGPT Teams nudger — round2 Playwright prototype.
//
// Run from prototypes\round2\playwright:
//
//	go run . -install          one time: driver + Chromium into ..\managed\playwright
//	go run . -dry-run -once    check all 4 chats, send nothing
//	go run .                   keep going every 10 minutes (Ctrl+C to stop)
//
// Uses only the project-owned Chromium (no system browser fallback).
// Sign in once with -login (plain browser, Google-safe). Login is kept in ..\runs\playwright\chatgpt-profile so you sign in once.
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/mxschmitt/playwright-go"
)

const nudge = "Check your ID card and latest GitHub updates. Continue your authorized work, coordinate with your team, and report results or blockers. If another conversation needs attention, provide the intended recipient and the next copy-and-paste nudge."

type seat struct{ id, role, url string }

var seats = []seat{
	{"01", "Project Leader", "https://chatgpt.com/g/g-p-6acbbf0aee748191b5bd7415efff6b35-teams/c/6acbf8b2-1dc0-83ea-8914-2beff94a0c18"},
	{"02", "Team Co-Lead", "https://chatgpt.com/g/g-p-6acbbf0aee748191b5bd7415efff6b35-teams/c/6acbf8c8-dd98-83ea-a992-3f7a8c49749b"},
	{"03", "Browser Builder", "https://chatgpt.com/g/g-p-6acbbf0aee748191b5bd7415efff6b35-teams/c/6acbf8da-1a70-83e9-ac76-ed94e79461b5"},
	{"04", "Independent QA", "https://chatgpt.com/g/g-p-6acbbf0aee748191b5bd7415efff6b35-teams/c/6acbf8ff-525c-83e9-81e9-cb48a8b55a60"},
}

const (
	selEditor = `#prompt-textarea, div[contenteditable="true"]`
	selStop   = `form button[aria-label="Stop"], [data-testid="stop-button"]`
	selLogin  = `[data-testid="login-button"]`
	loadErr   = "Could not load this ChatGPT conversation"
)

func main() {
	install := flag.Bool("install", false, "download Playwright driver + Chromium into ..\\managed\\playwright, then exit")
	dry := flag.Bool("dry-run", false, "check chats, send nothing")
	once := flag.Bool("once", false, "one round, then exit")
	every := flag.Duration("every", 10*time.Minute, "time between rounds")
	gap := flag.Duration("gap", 20*time.Second, "pause between chats (avoids ChatGPT rate limits)")
	demo := flag.Bool("demo", false, "watch: open ChatGPT logged out, click Try it for free, send one test chat")
	login := flag.Bool("login", false, "open a plain browser window to sign in to ChatGPT (no automation)")
	skip := flag.String("skip", "", `seats to leave alone, e.g. "01" or "01,03"`)
	flag.Parse()
	log.SetFlags(log.Ltime)

	demo_ := *demo
	if err := run(*install, *login, demo_, *dry, *once, *every, *gap, *skip); err != nil {
		log.Print("Playwright prototype failed: ", err)
		os.Exit(1)
	}
}

func run(install, login, demo, dry, once bool, every, gap time.Duration, skip string) error {
	root, err := filepath.Abs(filepath.Join("..", "managed", "playwright"))
	if err != nil {
		return err
	}
	if err := os.Setenv("PLAYWRIGHT_BROWSERS_PATH", root); err != nil {
		return err
	}
	// Keep the Node driver project-owned too (..\managed\driver), not in %LOCALAPPDATA%.
	driverDir, err := filepath.Abs(filepath.Join("..", "managed", "driver"))
	if err != nil {
		return err
	}
	opts := &playwright.RunOptions{
		Browsers:        []string{"chromium"},
		NoInstallShell:  true,
		DriverDirectory: driverDir,
	}

	// FIX 1: playwright.Run() never downloads anything. Install must be called explicitly.
	if install {
		if err := playwright.Install(opts); err != nil {
			return fmt.Errorf("install: %w", err)
		}
		log.Printf("installed Chromium into %s", root)
		return nil
	}

	pw, err := playwright.Run(opts)
	if err != nil {
		return fmt.Errorf("Playwright driver missing (run with -install first): %w", err)
	}
	defer pw.Stop()

	abs, err := filepath.Abs(pw.Chromium.ExecutablePath())
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(root, abs)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return errors.New("refusing executable outside project-owned Playwright browser cache")
	}
	if st, err := os.Stat(abs); err != nil || st.IsDir() {
		return errors.New("missing managed Playwright Chromium (run with -install first); no system browser fallback")
	}

	// FIX 2: a persistent profile, not a temp one, so the ChatGPT login survives between runs.
	profile, err := filepath.Abs(filepath.Join("..", "runs", "playwright", "chatgpt-profile"))
	if err != nil {
		return err
	}
	if err := os.MkdirAll(profile, 0o700); err != nil {
		return err
	}

	if demo {
		return runDemo(pw, abs)
	}

	// Sign-in mode: open the same managed Chromium as a PLAIN browser (no Playwright,
	// no automation flags) so Google allows the sign-in. You type everything yourself.
	if login {
		pw.Stop()
		log.Print("Opening a normal browser window for sign-in. Use 'Continue with Google' on ChatGPT,")
		log.Print("finish signing in, then CLOSE that window. Your session is saved for Run.")
		c := exec.Command(abs,
			"--user-data-dir="+profile,
			"--no-first-run", "--no-default-browser-check",
			"https://chatgpt.com/auth/login")
		if err := c.Run(); err != nil {
			return fmt.Errorf("sign-in browser: %w", err)
		}
		log.Print("Sign-in window closed. You can click Run now.")
		return nil
	}

	ctx, err := pw.Chromium.LaunchPersistentContext(profile, playwright.BrowserTypeLaunchPersistentContextOptions{
		Headless:       playwright.Bool(false),
		ExecutablePath: playwright.String(abs),
		NoViewport:     playwright.Bool(true),
	})
	if err != nil {
		if strings.Contains(err.Error(), "ProcessSingleton") || strings.Contains(err.Error(), "already in use") {
			return errors.New("the sign-in window is still open; close it, then click Run again")
		}
		return err
	}
	defer ctx.Close()

	page := ctx.Pages()[0]
	if _, err := page.Goto("https://chatgpt.com/"); err != nil {
		return fmt.Errorf("open chatgpt.com: %w", err)
	}
	time.Sleep(4 * time.Second)
	// Don't sign in inside the automated window (Google blocks it). Use Sign in mode instead.
	if n, _ := page.Locator(selLogin).Count(); n > 0 {
		return errors.New("not signed in to ChatGPT yet: click 'Sign in' in the launcher first")
	}
	log.Print("signed in to ChatGPT.")

	skipped := map[string]bool{}
	for _, s := range strings.Split(skip, ",") {
		if s = strings.TrimSpace(s); s != "" {
			skipped[s] = true
		}
	}
	identitySent := map[string]bool{} // first nudge per seat (this run) includes its identity

	for {
		log.Println("---- round ----")
		for i, s := range seats {
			if i > 0 {
				time.Sleep(gap)
			}
			if skipped[s.id] {
				log.Printf("Seat %s: skipped (-skip)", s.id)
				continue
			}
			log.Printf("Seat %s: %s", s.id, nudgeSeat(page, s, identitySent, dry))
		}
		if once {
			return nil
		}
		log.Printf("next round at %s", time.Now().Add(every).Format("3:04 PM"))
		time.Sleep(every)
	}
}

func nudgeSeat(page playwright.Page, s seat, identitySent map[string]bool, dry bool) string {
	if err := openChat(page, s.url); err != nil {
		return "skipped: " + err.Error()
	}
	time.Sleep(3 * time.Second) // let a running reply show its Stop square
	if ok, _ := page.Locator(selStop).First().IsVisible(); ok {
		return "running (Stop square), left alone"
	}

	msg := nudge
	first := !identitySent[s.id]
	if first {
		msg = fmt.Sprintf("You are Seat %s — %s (Continuation). %s", s.id, s.role, nudge)
	}
	if dry {
		return fmt.Sprintf("idle, would send %q (dry run)", short(msg))
	}

	ed := page.Locator(selEditor).First()
	if err := ed.Click(); err != nil {
		return "skipped: can't click message box: " + err.Error()
	}
	kb := page.Keyboard()
	_ = kb.Press("Control+A")
	_ = kb.Press("Delete")
	if err := kb.InsertText(msg); err != nil {
		return "skipped: can't type: " + err.Error()
	}
	if got, _ := ed.InnerText(); strings.TrimSpace(got) != msg {
		_ = kb.Press("Control+A")
		_ = kb.Press("Delete")
		return "skipped: message box didn't match exactly, cleared it"
	}
	if err := kb.Press("Enter"); err != nil {
		return "skipped: Enter failed: " + err.Error()
	}
	if err := page.Locator(selStop).First().WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible, Timeout: playwright.Float(20000),
	}); err != nil {
		return "sent, but didn't see it start; check this chat"
	}
	identitySent[s.id] = true
	if first {
		return "sent identity + nudge"
	}
	return "sent nudge"
}

// openChat loads a chat; on ChatGPT's "Could not load" (rate limit) it waits 60s and retries once.
func openChat(page playwright.Page, url string) error {
	for try := 1; try <= 2; try++ {
		if _, err := page.Goto(url, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded}); err != nil {
			return err
		}
		ed := page.Locator(selEditor).First()
		bad := page.GetByText(loadErr).First()
		if err := ed.Or(bad).First().WaitFor(playwright.LocatorWaitForOptions{
			State: playwright.WaitForSelectorStateVisible, Timeout: playwright.Float(30000),
		}); err != nil {
			return errors.New("chat never finished loading")
		}
		if ok, _ := bad.IsVisible(); !ok {
			return nil
		}
		if try == 1 {
			log.Print("  ChatGPT couldn't load the chat (rate limit?), waiting 60s")
			time.Sleep(60 * time.Second)
		}
	}
	return errors.New("ChatGPT still couldn't load it, next round")
}

func short(s string) string {
	if len(s) > 70 {
		return s[:70] + "…"
	}
	return s
}

// runDemo opens ChatGPT logged out in a fresh throwaway profile, slowed down so you
// can watch, clicks the "try it" button if one shows, and sends one test message.
func runDemo(pw *playwright.Playwright, exe string) error {
	tmp, err := os.MkdirTemp(filepath.Join("..", "runs", "playwright"), "demo-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	ctx, err := pw.Chromium.LaunchPersistentContext(tmp, playwright.BrowserTypeLaunchPersistentContextOptions{
		Headless:       playwright.Bool(false),
		ExecutablePath: playwright.String(exe),
		NoViewport:     playwright.Bool(true),
		SlowMo:         playwright.Float(600), // slow enough to watch each step
	})
	if err != nil {
		return err
	}
	defer ctx.Close()
	page := ctx.Pages()[0]

	log.Print("demo: opening chatgpt.com (logged out)")
	if _, err := page.Goto("https://chatgpt.com/", playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded}); err != nil {
		return err
	}
	time.Sleep(4 * time.Second)

	tryButtons := []string{"Try it for free", "Try it first", "Stay logged out", "Try ChatGPT"}
	clickTry := func() bool {
		for _, t := range tryButtons {
			b := page.GetByText(t, playwright.PageGetByTextOptions{Exact: playwright.Bool(false)}).First()
			if ok, _ := b.IsVisible(); ok {
				log.Printf("demo: clicking %q", t)
				if err := b.Click(); err == nil {
					time.Sleep(2 * time.Second)
					return true
				}
			}
		}
		return false
	}
	if !clickTry() {
		log.Print("demo: no 'try' button showing, going straight to the message box")
	}

	ed := page.Locator(selEditor).First()
	if err := ed.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible, Timeout: playwright.Float(30000)}); err != nil {
		return errors.New("demo: message box never appeared (ChatGPT may be showing a login wall or a human check)")
	}
	before, _ := page.Locator(`[data-message-author-role="assistant"]`).Count()

	msg := "This is a test message from the Teams launcher. Please reply with: test received."
	log.Printf("demo: typing %q", msg)
	if err := ed.Click(); err != nil {
		return err
	}
	if err := page.Keyboard().Type(msg, playwright.KeyboardTypeOptions{Delay: playwright.Float(25)}); err != nil {
		return err
	}
	log.Print("demo: pressing Enter")
	if err := page.Keyboard().Press("Enter"); err != nil {
		return err
	}

	// A "stay logged out" popup sometimes appears after the first message.
	time.Sleep(3 * time.Second)
	clickTry()

	log.Print("demo: waiting for the reply...")
	deadline := time.Now().Add(90 * time.Second)
	for time.Now().Before(deadline) {
		time.Sleep(2 * time.Second)
		n, _ := page.Locator(`[data-message-author-role="assistant"]`).Count()
		busy, _ := page.Locator(selStop).First().IsVisible()
		if n > before && !busy {
			txt, _ := page.Locator(`[data-message-author-role="assistant"]`).Last().InnerText()
			log.Printf("demo: ChatGPT replied: %s", strings.TrimSpace(txt))
			log.Print("demo: done. The window closes in 15 seconds.")
			time.Sleep(15 * time.Second)
			return nil
		}
	}
	return errors.New("demo: no reply within 90s (check the window for a popup or human check)")
}
