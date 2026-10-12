package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLogSinceHandlesTrimmedLines(t *testing.T) {
	r := &run{Started: time.Unix(123, 456)}
	for i := 0; i < 4006; i++ {
		r.add("message")
	}
	if len(r.Log) != 4000 || r.LogStart != 6 {
		t.Fatalf("log buffer: len=%d start=%d", len(r.Log), r.LogStart)
	}
	lines, next, reset, token := logSince(r, 0, "")
	if !reset || len(lines) != 4000 || next != 4006 || token == "" {
		t.Fatalf("initial poll: reset=%v lines=%d next=%d token=%q", reset, len(lines), next, token)
	}
	lines, next, reset, _ = logSince(r, next, token)
	if reset || len(lines) != 0 || next != 4006 {
		t.Fatalf("caught-up poll: reset=%v lines=%d next=%d", reset, len(lines), next)
	}
	r.add("latest")
	lines, next, reset, _ = logSince(r, next, token)
	if reset || len(lines) != 1 || !strings.Contains(lines[0], "latest") || next != 4007 {
		t.Fatalf("after rollover: reset=%v lines=%v next=%d", reset, lines, next)
	}
}

func TestLogSinceResetsOnNewRun(t *testing.T) {
	previous := &run{Started: time.Unix(1, 1)}
	previous.add("old")
	_, _, _, oldToken := logSince(previous, 0, "")
	current := &run{Started: time.Unix(2, 1)}
	current.add("new")
	lines, next, reset, token := logSince(current, 1, oldToken)
	if !reset || next != 1 || token == oldToken || len(lines) != 1 || !strings.Contains(lines[0], "new") {
		t.Fatalf("new run cursor: reset=%v next=%d lines=%v", reset, next, lines)
	}
}

func TestFindProjectsDirSourceAndRelease(t *testing.T) {
	root := t.TempDir()
	release := filepath.Join(root, "dist")
	if err := os.MkdirAll(filepath.Join(root, "projects"), 0o755); err != nil { t.Fatal(err) }
	if err := os.MkdirAll(release, 0o755); err != nil { t.Fatal(err) }
	wd, err := os.Getwd()
	if err != nil { t.Fatal(err) }
	t.Cleanup(func() { _ = os.Chdir(wd) })
	t.Setenv("DASHBOARD_PROJECTS_DIR", "")
	if err := os.Chdir(root); err != nil { t.Fatal(err) }
	if got := findProjectsDir(release); got != filepath.Join(root, "projects") {
		t.Fatalf("source path = %s", got)
	}
	if err := os.MkdirAll(filepath.Join(release, "projects"), 0o755); err != nil { t.Fatal(err) }
	if got := findProjectsDir(release); got != filepath.Join(release, "projects") {
		t.Fatalf("release path = %s", got)
	}
	custom := filepath.Join(root, "custom")
	t.Setenv("DASHBOARD_PROJECTS_DIR", custom)
	if got := findProjectsDir(release); got != custom { t.Fatalf("override = %s", got) }
}

func TestLoadProjectsEmptyArrayAndMissingFolder(t *testing.T) {
	oldDir := projDir
	t.Cleanup(func() { projDir = oldDir })
	projDir = t.TempDir()
	ps, problems := loadProjects()
	if len(problems) != 0 { t.Fatalf("unexpected errors: %v", problems) }
	b, err := json.Marshal(ps)
	if err != nil || string(b) != "[]" { t.Fatalf("empty list JSON = %s (err %v)", b, err) }
	projDir = filepath.Join(t.TempDir(), "missing")
	_, problems = loadProjects()
	if len(problems) == 0 { t.Fatal("expected missing folder error") }
}
