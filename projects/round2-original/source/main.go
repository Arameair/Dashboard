package main

import (
	"flag"
	"fmt"
	"github.com/mxschmitt/playwright-go"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

var smoke = flag.Bool("smoke", false, "deterministic local pages, then close automatically")
var headless = flag.Bool("headless", false, "hide the browser during smoke verification")
var install = flag.Bool("install", false, "install pinned browser and driver into the project cache")

func main() {
	flag.Parse()
	if err := run(); err != nil {
		log.Print("Playwright prototype failed: ", err)
		os.Exit(1)
	}
}
func run() error {
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
		return playwright.Install(&playwright.RunOptions{DriverDirectory: driver, Browsers: []string{"chromium"}})
	}
	pw, err := playwright.Run(&playwright.RunOptions{DriverDirectory: driver})
	if err != nil {
		return fmt.Errorf("Playwright driver missing: %w", err)
	}
	defer pw.Stop()
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
	profile, err := os.MkdirTemp(profiles, "profile-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(profile)
	ctx, err := pw.Chromium.LaunchPersistentContext(profile, playwright.BrowserTypeLaunchPersistentContextOptions{Headless: headless, ExecutablePath: playwright.String(abs)})
	if err != nil {
		return err
	}
	defer ctx.Close()
	page, err := ctx.NewPage()
	if err != nil {
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
	path := filepath.Join("..", "artifacts", "playwright-google.png")
	if _, err = page.Screenshot(playwright.PageScreenshotOptions{Path: playwright.String(path)}); err != nil {
		return err
	}
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
	tab, err := ctx.NewPage()
	if err != nil {
		return fmt.Errorf("new tab: %w", err)
	}
	secondURL := "https://example.com"
	if *smoke {
		secondURL = "data:text/html,<title>Second local tab</title>"
	}
	if _, err = tab.Goto(secondURL); err != nil {
		tab.Close()
		return fmt.Errorf("navigate second tab: %w", err)
	}
	secondTitle, err := tab.Title()
	if err != nil {
		tab.Close()
		return err
	}
	if secondTitle == "" {
		tab.Close()
		return fmt.Errorf("second site returned an empty title")
	}
	fmt.Println("Second page title: ", secondTitle)
	if err = tab.Close(); err != nil {
		return fmt.Errorf("close second tab: %w", err)
	}
	fmt.Println("Second tab closed.")
	fmt.Println("Local selector/type/click and second-tab operations executed.")
	fmt.Printf("Managed Playwright browser screenshot: %s\nPress Enter to stop.\n", path)
	if !*smoke {
		fmt.Scanln()
	}
	return nil
}
