package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDashboardRoot(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "dashboard")
	if err := os.Mkdir(source, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"go.mod", "page.go"} {
		if err := os.WriteFile(filepath.Join(source, name), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	executable := filepath.Join(t.TempDir(), "dashboard.exe")
	for _, cwd := range []string{root, source} {
		if got := dashboardRoot(executable, cwd); got != root {
			t.Errorf("from %s: got %s, want %s", cwd, got, root)
		}
	}
	if got := dashboardRoot(executable, t.TempDir()); got != filepath.Dir(executable) {
		t.Errorf("packaged root = %s", got)
	}
}

func TestExternalProjectsAndSeparateLogs(t *testing.T) {
	oldProjects, oldLogs := projDir, logDir
	t.Cleanup(func() { projDir, logDir = oldProjects, oldLogs })
	projDir, logDir = t.TempDir(), t.TempDir()
	folder := filepath.Join(projDir, "demo")
	if err := os.Mkdir(folder, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(folder, "project.json"), []byte(`{"name":"External demo","exe":"demo.exe"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	ps, problems := loadProjects()
	if len(problems) != 0 || len(ps) != 1 || ps[0].workDir() != folder {
		t.Fatalf("projects = %v; problems = %v", ps, problems)
	}
	if err := saveRunLog(ps[0].ID, "finished OK"); err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(filepath.Join(logDir, "demo", "last-run.log"))
	if err != nil || string(contents) != "finished OK\n" {
		t.Fatalf("log = %q; error = %v", contents, err)
	}
	if _, err := os.Stat(filepath.Join(folder, "last-run.log")); !os.IsNotExist(err) {
		t.Fatalf("dashboard log was written into the project: %v", err)
	}
}
