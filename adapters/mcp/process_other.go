//go:build !linux && !darwin && !freebsd && !openbsd && !netbsd && !dragonfly

package mcp

import "os/exec"

func configureProcess(cmd *exec.Cmd)  {}
func killProcess(cmd *exec.Cmd) error { return cmd.Process.Kill() }
