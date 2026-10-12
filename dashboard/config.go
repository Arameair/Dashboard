package main

// config.json lives in the data folder and is edited from the dashboard.
// Each project is one browser with its own profile (sign-ins, cookies).

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

type Selectors struct {
	Input string `json:"input"` // message box
	Reply string `json:"reply"` // assistant replies (last one is "latest")
	Busy  string `json:"busy"`  // visible while the site is still generating
}

type ProjectCfg struct {
	ID          string `json:"id"` // fixed once created; names the profile folder
	Name        string `json:"name"`
	URL         string `json:"url"`
	Color       string `json:"color"`
	Nudge       string `json:"nudge"`
	NudgeMin    int    `json:"nudge_minutes"`
	AutoNudge   bool   `json:"auto_nudge"`
	SaveReplies bool   `json:"save_replies"`
	Console     bool   `json:"console"` // copy the page's console errors/logs into this project's log
	AutoStart   bool   `json:"auto_start"`
}

type Config struct {
	BrowserPath string       `json:"browser_path"` // empty = find Chrome / Edge automatically
	Selectors   Selectors    `json:"selectors"`
	Projects    []ProjectCfg `json:"projects"`
}

var palette = []string{"#3b82f6", "#16a34a", "#ea580c", "#9333ea", "#db2777", "#0891b2", "#ca8a04", "#64748b"}

func defaultSelectors() Selectors {
	return Selectors{
		Input: `#prompt-textarea, div[contenteditable="true"], textarea`,
		Reply: `[data-message-author-role="assistant"]`,
		Busy:  `[data-testid="stop-button"], button[aria-label="Stop streaming"], form button[aria-label="Stop"]`,
	}
}

// A new dashboard starts empty. Projects are added from the page (+ Project) or
// come from definition files (see mergeDefinitions).
func defaultConfig() *Config { return &Config{Selectors: defaultSelectors()} }

// mergeDefinitions adds projects listed in definition files that the dashboard
// doesn't have yet. A definition file is {"projects": [...]} with the same
// fields as config.json and can live anywhere (for example in the repository
// of the work those browsers belong to). paths is a ";"-separated list of files
// or folders (every *.json in a folder); missing ones are skipped. Projects you
// already have (same id) are left alone, so changes made in the dashboard win.
func mergeDefinitions(c *Config, paths string) (added []string, errs []error) {
	have := map[string]bool{}
	for _, p := range c.Projects {
		have[p.ID] = true
	}
	var files []string
	for _, p := range strings.Split(paths, ";") {
		p = strings.Trim(strings.TrimSpace(p), `"`)
		fi, err := os.Stat(p)
		if p == "" || err != nil {
			continue
		}
		if !fi.IsDir() {
			files = append(files, p)
			continue
		}
		m, _ := filepath.Glob(filepath.Join(p, "*.json"))
		files = append(files, m...)
	}
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		var d struct {
			Projects []ProjectCfg `json:"projects"`
		}
		if err := json.Unmarshal(b, &d); err != nil {
			errs = append(errs, fmt.Errorf("%s is not valid JSON (%v)", f, err))
			continue
		}
		for _, p := range d.Projects {
			if p.ID = slug(p.ID); p.ID == "" {
				p.ID = slug(p.Name)
			}
			if p.ID == "" || have[p.ID] {
				continue
			}
			have[p.ID] = true
			c.Projects = append(c.Projects, p)
			added = append(added, p.Name)
		}
	}
	c.normalize()
	return added, errs
}

func (c *Config) normalize() {
	d := defaultSelectors()
	if strings.TrimSpace(c.Selectors.Input) == "" {
		c.Selectors.Input = d.Input
	}
	if strings.TrimSpace(c.Selectors.Reply) == "" {
		c.Selectors.Reply = d.Reply
	}
	if strings.TrimSpace(c.Selectors.Busy) == "" {
		c.Selectors.Busy = d.Busy
	}
	seen := map[string]bool{}
	out := c.Projects[:0]
	for i, p := range c.Projects {
		p.ID = slug(p.ID)
		if p.ID == "" {
			p.ID = slug(p.Name)
		}
		if p.ID == "" || seen[p.ID] {
			continue
		}
		seen[p.ID] = true
		if p.Name == "" {
			p.Name = p.ID
		}
		if p.Color == "" {
			p.Color = palette[i%len(palette)]
		}
		if p.NudgeMin <= 0 {
			p.NudgeMin = 10
		}
		out = append(out, p)
	}
	c.Projects = out
}

func loadConfig(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		c := defaultConfig()
		return c, saveConfig(path, c)
	}
	if err != nil {
		return nil, err
	}
	c := &Config{}
	if err := json.Unmarshal(b, c); err != nil {
		return nil, fmt.Errorf("%s is not valid JSON (%v). Fix it or delete it to start fresh", path, err)
	}
	c.normalize()
	return c, nil
}

func saveConfig(path string, c *Config) error {
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

var slugRe = regexp.MustCompile(`[^a-z0-9]+`)

func slug(s string) string {
	s = slugRe.ReplaceAllString(strings.ToLower(s), "-")
	s = strings.Trim(s, "-")
	if len(s) > 40 {
		s = strings.Trim(s[:40], "-")
	}
	return s
}

// findBrowser returns a Chrome/Edge/Chromium executable.
func findBrowser(override string) (string, error) {
	if override = strings.Trim(strings.TrimSpace(override), `"`); override != "" {
		if fi, err := os.Stat(override); err == nil && !fi.IsDir() {
			return override, nil
		}
		return "", fmt.Errorf("browser path %q doesn't exist (fix it in Settings)", override)
	}
	var cands []string
	switch runtime.GOOS {
	case "windows":
		for _, env := range []string{"ProgramFiles", "ProgramFiles(x86)", "LOCALAPPDATA"} {
			if b := os.Getenv(env); b != "" {
				cands = append(cands, filepath.Join(b, `Google\Chrome\Application\chrome.exe`))
			}
		}
		for _, env := range []string{"ProgramFiles(x86)", "ProgramFiles"} {
			if b := os.Getenv(env); b != "" {
				cands = append(cands, filepath.Join(b, `Microsoft\Edge\Application\msedge.exe`))
			}
		}
	case "darwin":
		cands = []string{"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge",
			"/Applications/Chromium.app/Contents/MacOS/Chromium"}
	default:
		for _, n := range []string{"google-chrome", "google-chrome-stable", "chromium", "chromium-browser", "microsoft-edge", "brave-browser"} {
			if p, err := exec.LookPath(n); err == nil {
				cands = append(cands, p)
			}
		}
	}
	for _, c := range cands {
		if fi, err := os.Stat(c); err == nil && !fi.IsDir() {
			return c, nil
		}
	}
	return "", fmt.Errorf("couldn't find Chrome or Edge; set the browser path in Settings")
}
