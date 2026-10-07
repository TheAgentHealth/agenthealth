//go:build linux

package mcp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/TheAgentHealth/agenthealth/core"
)

func TestStdioCancellationTerminatesDescendants(t *testing.T) {
	t.Setenv("STDIO_TEST_TOKEN", "child-secret-value")
	pidFile := filepath.Join(t.TempDir(), "child.pid")
	t.Setenv("STDIO_TREE_PIDFILE", pidFile)
	target := stdioTarget(t, "descendants")
	target.MCP.Stdio.Env["PID_FILE"] = "STDIO_TREE_PIDFILE"
	target.Checks = []string{"protocol"}
	target.MCP.Functional = nil
	timeout := 1000
	target.TimeoutMS = &timeout
	result := runTarget(t, target)
	if result.Status != core.Unreachable {
		t.Fatalf("%+v", result)
	}
	pid, err := os.ReadFile(pidFile)
	if err != nil {
		t.Fatal("helper never started its child")
	}
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		stat, err := os.ReadFile("/proc/" + strings.TrimSpace(string(pid)) + "/stat")
		if os.IsNotExist(err) {
			return
		}
		if err == nil {
			end := strings.LastIndex(string(stat), ")")
			if end >= 0 && len(stat) > end+2 && stat[end+2] == 'Z' {
				return
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("MCP descendant survived cancellation")
}

func TestStdioCommandRelativeToDirectory(t *testing.T) {
	t.Setenv("STDIO_TEST_TOKEN", "child-secret-value")
	target := stdioTarget(t, "modern")
	directory := t.TempDir()
	if err := os.Symlink(target.MCP.Stdio.Command, filepath.Join(directory, "server")); err != nil {
		t.Fatal(err)
	}
	target.MCP.Stdio.Directory = directory
	target.MCP.Stdio.Command = "./server"
	if result := runTarget(t, target); result.Status != core.Healthy {
		t.Fatalf("relative command rejected: %+v", result)
	}
}
