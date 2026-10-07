package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/TheAgentHealth/agenthealth/core"
)

func TestMCPStdioProcess(t *testing.T) {
	if len(os.Args) < 3 || os.Args[len(os.Args)-2] != "mcp-helper" {
		return
	}
	mode := os.Args[len(os.Args)-1]
	if mode == "tree-child" {
		time.Sleep(30 * time.Second)
		os.Exit(0)
	}
	if mode == "descendants" {
		executable, _ := os.Executable()
		child := exec.Command(executable, "-test.run=TestMCPStdioProcess", "--", "mcp-helper", "tree-child")
		if child.Start() != nil {
			os.Exit(9)
		}
		os.WriteFile(os.Getenv("PID_FILE"), []byte(fmt.Sprint(child.Process.Pid)), 0600)
	}
	if mode == "exit" {
		os.Exit(2)
	}
	reader := bufio.NewScanner(os.Stdin)
	reader.Buffer(make([]byte, 4096), maxBody+1)
	for reader.Scan() {
		var q struct {
			ID     int                        `json:"id"`
			Method string                     `json:"method"`
			Params map[string]json.RawMessage `json:"params"`
		}
		json.Unmarshal(reader.Bytes(), &q)
		if q.Method == "notifications/initialized" || q.Method == "notifications/cancelled" {
			continue
		}
		if mode == "partial" {
			fmt.Fprint(os.Stdout, "{")
			for reader.Scan() {
			}
			os.Exit(0)
		}
		if mode == "env-isolation" && os.Getenv("UNRELATED_MCP_SECRET") != "" {
			os.Exit(7)
		}
		if mode == "hang" || mode == "descendants" || mode == "silent-legacy" && q.Method == "server/discover" {
			continue
		}
		if mode == "oversize" {
			fmt.Println(strings.Repeat("x", maxBody+1))
			continue
		}
		if mode == "stderr" {
			fmt.Fprintln(os.Stderr, "child-secret-value")
		}
		if q.Method == "server/discover" && mode != "modern" {
			json.NewEncoder(os.Stdout).Encode(map[string]any{"jsonrpc": "2.0", "id": q.ID, "error": map[string]any{"code": -32602, "message": "not initialized"}})
			continue
		}
		var result any
		switch q.Method {
		case "server/discover":
			var meta map[string]json.RawMessage
			json.Unmarshal(q.Params["_meta"], &meta)
			if string(meta["io.modelcontextprotocol/protocolVersion"]) != `"2026-07-28"` {
				os.Exit(3)
			}
			result = map[string]any{"resultType": "complete", "supportedVersions": []string{modernVersion}, "capabilities": map[string]any{"tools": map[string]any{}}}
		case "initialize":
			result = map[string]any{"protocolVersion": "2025-11-25", "serverInfo": map[string]string{"name": "stdio-helper", "version": "1"}, "capabilities": map[string]any{"tools": map[string]any{}}}
		case "tools/list":
			result = map[string]any{"resultType": "complete", "tools": []any{map[string]any{"name": "search", "inputSchema": map[string]string{"type": "object"}, "annotations": map[string]bool{"readOnlyHint": true, "destructiveHint": false}}}}
		case "tools/call":
			result = map[string]any{"resultType": "complete", "content": []any{map[string]string{"type": "text", "text": os.Getenv("CHILD_TOKEN")}}}
		default:
			os.Exit(4)
		}
		json.NewEncoder(os.Stdout).Encode(map[string]any{"jsonrpc": "2.0", "id": q.ID, "result": result})
	}
	os.Exit(0)
}
func stdioTarget(t *testing.T, mode string) core.Target {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	return core.Target{Name: "local-process", Type: "mcp", Endpoint: "stdio://fixture", Checks: []string{"protocol", "capability", "functional", "latency"}, MCP: &core.MCPOptions{Transport: "stdio", Stdio: &core.MCPStdio{Command: executable, Args: []string{"-test.run=TestMCPStdioProcess", "--", "mcp-helper", mode}, Env: map[string]string{"CHILD_TOKEN": "STDIO_TEST_TOKEN"}}, Functional: &core.MCPInvocation{Tool: "search", Safe: true}}}
}
func TestStdioTransports(t *testing.T) {
	t.Setenv("STDIO_TEST_TOKEN", "child-secret-value")
	t.Setenv("UNRELATED_MCP_SECRET", "must-not-be-inherited")
	for _, mode := range []string{"modern", "legacy", "silent-legacy", "stderr", "env-isolation"} {
		t.Run(mode, func(t *testing.T) {
			target := stdioTarget(t, mode)
			result := runTarget(t, target)
			if result.Status != core.Healthy || result.LatencyMS == nil {
				t.Fatalf("%+v", result)
			}
		})
	}
	for _, mode := range []string{"hang", "exit", "oversize", "partial"} {
		t.Run(mode, func(t *testing.T) {
			target := stdioTarget(t, mode)
			timeout := 100
			if mode == "partial" {
				timeout = 1000
			}
			if mode == "oversize" {
				timeout = 2000
			}
			target.TimeoutMS = &timeout
			target.Checks = []string{"protocol"}
			target.MCP.Functional = nil
			started := time.Now()
			result := runTarget(t, target)
			expected := core.Unreachable
			if mode == "oversize" || mode == "partial" {
				expected = core.Unknown
			}
			if result.Status != expected || time.Since(started) > 3*time.Second {
				t.Fatalf("%+v elapsed=%s", result, time.Since(started))
			}
		})
	}
}
