package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLogSinceSurvivesLogRollover(t *testing.T) {
	r := &run{Started: time.Unix(7, 1)}
	for i := 0; i < 4006; i++ {
		r.add("sample")
	}
	if len(r.Log) != 4000 || r.LogStart != 6 {
		t.Fatalf("retained=%d start=%d, want 4000 and 6", len(r.Log), r.LogStart)
	}
	initial, next, reset, token := logSince(r, 0, "")
	if !reset || len(initial) != 4000 || next != 4006 || token == "" {
		t.Fatalf("initial: reset=%v len=%d next=%d token=%q", reset, len(initial), next, token)
	}
	lines, after, changed, sameToken := logSince(r, next, token)
	if changed || after != 4006 || len(lines) != 0 || sameToken != token {
		t.Fatalf("caught up: reset=%v next=%d len=%d", changed, after, len(lines))
	}
	r.add("last-new-line")
	lines, after, changed, _ = logSince(r, next, token)
	if changed || after != 4007 || len(lines) != 1 || !strings.Contains(lines[0], "last-new-line") {
		t.Fatalf("new line after trim: reset=%v next=%d lines=%v", changed, after, lines)
	}
}

func TestLogSinceResetsWhenRunChanges(t *testing.T) {
	old := &run{Started: time.Unix(1, 1)}
	old.add("old")
	_, _, _, oldToken := logSince(old, 0, "")
	current := &run{Started: time.Unix(2, 1)}
	current.add("new")
	lines, next, reset, token := logSince(current, 1, oldToken)
	if !reset || token == oldToken || next != 1 || len(lines) != 1 || !strings.Contains(lines[0], "new") {
		t.Fatalf("new-run reset=%v next=%d token=%q lines=%v", reset, next, token, lines)
	}
}

func TestLogSinceNoRunReturnsEmptyArray(t *testing.T) {
	lines, next, reset, token := logSince(nil, 0, "")
	if lines == nil || len(lines) != 0 || next != 0 || reset || token != "" {
		t.Fatalf("empty log: lines=%v next=%d reset=%v token=%q", lines, next, reset, token)
	}
}

func TestLoadProjectsReportsUnavailableFolder(t *testing.T) {
	previous := projDir
	t.Cleanup(func() { projDir = previous })
	projDir = filepath.Join(t.TempDir(), "does-not-exist")
	ps, problems := loadProjects()
	if len(ps) != 0 || len(problems) == 0 {
		t.Fatalf("projects=%v problems=%v", ps, problems)
	}
	if _, err := os.Stat(projDir); !os.IsNotExist(err) {
		t.Fatalf("loadProjects must not create unavailable folders: %v", err)
	}
}
