// Round 2 Playwright prototype — improvement prototype 1: clean shutdown.
//
// Same flags, folders, screenshot path and output lines as the original.
// What changed:
//   - Ctrl+C or closing the window now closes Chromium and deletes the temp
//     profile (before, Chromium kept running and runs\playwright\profile-* stayed).
//   - Profile folders left behind by earlier crashed runs are cleaned up at start.
//   - Uses the tab the browser opens with instead of leaving an extra blank tab.
//   - Each step prints "step N/5 ok" and the run ends with a PASS/FAIL summary.
//   - -install says when it finished; "Press Enter to stop." only shows when it waits.
//   - -shot NAME picks the screenshot file name (default playwright-google.png),
//     so it can run next to the original without overwriting its screenshot.
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/mxschmitt/playwright-go"
)

var smoke = flag.Bool("smoke", false, "deterministic local pages, then close automatically")
var headless = flag.Bool("headless", false, "hide the browser during smoke verification")
var install = flag.Bool("install", false, "install pinned browser and driver into the project cache")
var shot = flag.String("shot", "playwright-google.png", "screenshot file name inside ..\\artifacts")

const totalSteps = 5

// cleanup closes the browser and removes the temp profile exactly once,
// whether the run ends normally, fails, or is interrupted.
type cleanup struct {
	once    sync.Once
	pw      *playwright.Playwright
	ctx     playwright.BrowserContext
	profile string
}

func (c *cleanup) run() {
	c.once.Do(func() {
		if c.ctx != nil {
			_ = c.ctx.Close()
		}
		if c.pw != nil {
			_ = c.pw.Stop()
		}
		if c.profile != "" {
			// Chromium can hold files for a moment after it exits (Windows).
			for i := 0; i < 10; i++ {
				if err := os.RemoveAll(c.profile); err == nil {
					break
				}
				time.Sleep(300 * time.Millisecond)
			}
		}
	})
}

func main() {
	flag.Parse()
	c := &cleanup{}

	// Ctrl+C, and on Windows closing the console window, arrive here.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sig
		fmt.Println("\nInterrupted: closing the browser and removing the temp profile...")
		c.run()
		fmt.Println("Cleanup done.")
		os.Exit(130)
	}()

	err := run(c)
	c.run()
	if err != nil {
		log.Print("Playwright prototype failed: ", err)
		os.Exit(1)
	}
}

func step(n int, what string) { fmt.Printf("step %d/%d ok: %s\n", n, totalSteps, what) }

func run(c *cleanup) error {
	root, err := filepath.Abs(filepath.Join("..", "managed", "playwright"))
	if err != nil {
		return err
	}
	if err := os.Setenv("PLAYWRIGHT_BROWSERS_PATH", root); err != nil {
		return err
	}
	driver, err := filepath.Abs(filepath.Join("..", "managed", "driver"))
	if err != nil {
		return err
	}
	if *install {
		if err := playwright.Install(&playwright.RunOptions{DriverDirectory: driver, Browsers: []string{"chromium"}}); err != nil {
			return err
		}
		fmt.Printf("Install finished: driver in %s, browser in %s\n", driver, root)
		return nil
	}
	pw, err := playwright.Run(&playwright.RunOptions{DriverDirectory: driver})
	if err != nil {
		return fmt.Errorf("Playwright driver missing: %w", err)
	}
	c.pw = pw
	bin := pw.Chromium.ExecutablePath()
	abs, err := filepath.Abs(bin)
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(root, abs)
	if err != nil || (rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator))) || filepath.IsAbs(rel) {
		return fmt.Errorf("refusing executable outside project-owned Playwright browser cache")
	}
	if st, err := os.Stat(abs); err != nil || st.IsDir() {
		return fmt.Errorf("missing managed Playwright Chromium; no system browser fallback")
	}
	profiles := filepath.Join("..", "runs", "playwright")
	if err := os.MkdirAll(profiles, 0700); err != nil {
		return err
	}
	sweepStaleProfiles(profiles)
	profile, err := os.MkdirTemp(profiles, "profile-")
	if err != nil {
		return err
	}
	c.profile = profile
	ctx, err := pw.Chromium.LaunchPersistentContext(profile, playwright.BrowserTypeLaunchPersistentContextOptions{Headless: headless, ExecutablePath: playwright.String(abs)})
	if err != nil {
		return err
	}
	c.ctx = ctx
	step(1, "managed Chromium launched with a fresh temp profile")

	// The browser starts with one tab; use it rather than opening a second, blank one.
	var page playwright.Page
	if ps := ctx.Pages(); len(ps) > 0 {
		page = ps[0]
	} else if page, err = ctx.NewPage(); err != nil {
		return err
	}

	firstURL := "https://www.google.com"
	if *smoke {
		firstURL = "data:text/html,<title>Alpha smoke fixture</title><h1>Local browser smoke</h1>"
	}
	if _, err = page.Goto(firstURL); err != nil {
		return err
	}
	title, err := page.Title()
	if err != nil {
		return err
	}
	fmt.Printf("Real page title: %s\nURL: %s\n", title, page.URL())
	if err := os.MkdirAll(filepath.Join("..", "artifacts"), 0700); err != nil {
		return err
	}
	path := filepath.Join("..", "artifacts", filepath.Base(*shot))
	if _, err = page.Screenshot(playwright.PageScreenshotOptions{Path: playwright.String(path)}); err != nil {
		return err
	}
	step(2, "first page opened, title read, screenshot saved")

	// Local deterministic click/type exercise: no private data sent to external pages.
	fixture := "data:text/html," + url.PathEscape(`<html><title>Before</title><input id="query"><button id="send" onclick="document.title=document.querySelector('#query').value">Apply</button></html>`)
	if _, err = page.Goto(fixture); err != nil {
		return fmt.Errorf("local action page: %w", err)
	}
	if err = page.Locator("#query").Fill("go-browser-ready"); err != nil {
		return fmt.Errorf("fill selector: %w", err)
	}
	if err = page.Locator("#send").Click(); err != nil {
		return fmt.Errorf("click selector: %w", err)
	}
	actionTitle, err := page.Title()
	if err != nil {
		return err
	}
	if actionTitle != "go-browser-ready" {
		return fmt.Errorf("unexpected title after click: %q", actionTitle)
	}
	step(3, "local type + click changed the title")

	if err := secondTab(ctx); err != nil {
		return err
	}
	step(4, "second tab opened, read and closed")

	if n := len(ctx.Pages()); n != 1 {
		return fmt.Errorf("expected 1 open tab at the end, found %d", n)
	}
	step(5, "exactly one tab left open")

	fmt.Println("Local selector/type/click and second-tab operations executed.")
	fmt.Printf("Managed Playwright browser screenshot: %s\n", path)
	fmt.Printf("RESULT: PASS (%d/%d steps)\n", totalSteps, totalSteps)
	if !*smoke {
		fmt.Println("Press Enter to stop.")
		fmt.Scanln()
	}
	return nil
}

func secondTab(ctx playwright.BrowserContext) error {
	tab, err := ctx.NewPage()
	if err != nil {
		return fmt.Errorf("new tab: %w", err)
	}
	defer tab.Close() // closed even if a step below fails
	secondURL := "https://example.com"
	if *smoke {
		secondURL = "data:text/html,<title>Second local tab</title>"
	}
	if _, err = tab.Goto(secondURL); err != nil {
		return fmt.Errorf("navigate second tab: %w", err)
	}
	secondTitle, err := tab.Title()
	if err != nil {
		return err
	}
	if secondTitle == "" {
		return errors.New("second site returned an empty title")
	}
	fmt.Println("Second page title: ", secondTitle)
	if err = tab.Close(); err != nil {
		return fmt.Errorf("close second tab: %w", err)
	}
	fmt.Println("Second tab closed.")
	return nil
}

// sweepStaleProfiles removes profile-* folders left by runs that were killed.
// Only folders older than 10 minutes are touched, so a run in progress is safe.
func sweepStaleProfiles(dir string) {
	ents, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range ents {
		if !e.IsDir() || !strings.HasPrefix(e.Name(), "profile-") {
			continue
		}
		info, err := e.Info()
		if err != nil || time.Since(info.ModTime()) < 10*time.Minute {
			continue
		}
		p := filepath.Join(dir, e.Name())
		if os.RemoveAll(p) == nil {
			fmt.Println("Removed leftover profile from an earlier run:", p)
		}
	}
}
