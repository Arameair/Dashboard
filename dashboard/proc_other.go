//go:build !windows

package main

import (
	"os/exec"
	"runtime"
)

func killTree(c *exec.Cmd) error {
	if c == nil || c.Process == nil {
		return nil
	}
	return c.Process.Kill()
}

func opener() string {
	if runtime.GOOS == "darwin" {
		return "open"
	}
	return "xdg-open"
}

func openPath(p string) error { return exec.Command(opener(), p).Start() }

func openURL(u string) error { return exec.Command(opener(), u).Start() }
