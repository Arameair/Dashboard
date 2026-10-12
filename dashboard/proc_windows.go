package main

import (
	"os/exec"
	"strconv"
)

func killTree(c *exec.Cmd) error {
	if c == nil || c.Process == nil {
		return nil
	}
	return exec.Command("taskkill", "/T", "/F", "/PID", strconv.Itoa(c.Process.Pid)).Run()
}

func openPath(p string) error { return exec.Command("explorer", p).Start() }

func openURL(u string) error { return exec.Command("rundll32", "url.dll,FileProtocolHandler", u).Start() }
