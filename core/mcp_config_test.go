package core

import (
	"strings"
	"testing"
)

func TestMCPConfiguration(t *testing.T) {
	prefix := "version: v1\ntargets:\n  - name: mcp\n    type: mcp\n    endpoint: http://localhost:3000/mcp\n"
	for _, suffix := range []string{"    mcp: {}\n", "    mcp:\n      protocol_version: '2025-06-18'\n      required_tools: [search]\n", "    checks: [functional]\n    mcp:\n      functional:\n        tool: search\n        safe: true\n        arguments_json: '{\"query\":\"health\"}'\n"} {
		if _, err := LoadConfig(strings.NewReader(prefix + suffix)); err != nil {
			t.Fatalf("valid config: %v", err)
		}
	}
	for _, suffix := range []string{"    mcp: null\n", "    mcp:\n      protocol_version: '2024-11-05'\n", "    mcp:\n      required_tools: [search, search]\n", "    mcp:\n      required_resources: [' ']\n", "    checks: [reachability]\n    mcp:\n      required_tools: [search]\n", "    checks: [functional]\n", "    mcp:\n      functional:\n        tool: search\n        safe: true\n", "    checks: [functional]\n    mcp:\n      functional:\n        tool: search\n        safe: false\n", "    checks: [functional]\n    mcp:\n      functional:\n        tool: search\n        safe: true\n        arguments_json: 'null'\n", "    mcp:\n      required_tools: [123]\n", "    mcp:\n      unknown: true\n"} {
		if _, err := LoadConfig(strings.NewReader(prefix + suffix)); err == nil {
			t.Fatalf("invalid accepted: %s", suffix)
		}
	}
	if _, err := LoadConfig(strings.NewReader(strings.Replace(prefix, "type: mcp", "type: http", 1) + "    mcp: {}\n")); err == nil {
		t.Fatal("MCP options accepted on HTTP")
	}
}
