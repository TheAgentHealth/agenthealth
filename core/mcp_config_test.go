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

func TestMCPTransportAndOAuthConfiguration(t *testing.T) {
	prefix := "version: v1\ntargets:\n  - name: mcp\n    type: mcp\n    endpoint: https://example.com/mcp\n"
	valid := []string{
		"    mcp:\n      protocol_version: '2026-07-28'\n",
		"    mcp:\n      transport: stdio\n      stdio:\n        command: python3\n        args: [server.py]\n        env:\n          MCP_TOKEN: TOKEN_REFERENCE\n",
		"    mcp:\n      oauth:\n        issuer: https://auth.example.com\n        client_id: registered\n        grant: client_credentials\n        client_secret_env: CLIENT_SECRET\n",
		"    mcp:\n      oauth:\n        issuer: https://auth.example.com\n        client_id: registered\n        grant: authorization_code\n        token_file: .credentials/token.json\n",
	}
	for _, suffix := range valid {
		if _, err := LoadConfig(strings.NewReader(prefix + suffix)); err != nil {
			t.Fatalf("valid rejected %v", err)
		}
	}
	invalid := []string{
		"    mcp:\n      transport: stdio\n",
		"    mcp:\n      stdio:\n        command: python3\n",
		"    mcp:\n      transport: stdio\n      stdio:\n        command: python3\n        args: [1]\n",
		"    mcp:\n      transport: stdio\n      stdio:\n        command: python3\n        env:\n          TOKEN: 1\n",
		"    mcp:\n      oauth:\n        issuer: https://auth.example.com\n        client_id: registered\n        grant: client_credentials\n",
		"    mcp:\n      oauth:\n        issuer: https://auth.example.com\n        client_id: registered\n        grant: authorization_code\n",
		"    mcp:\n      oauth:\n        issuer: https://auth.example.com\n        client_id: registered\n        grant: authorization_code\n        token_file: .credentials/token.json\n        redirect_port: 65536\n",
		"    mcp:\n      oauth: null\n",
	}
	for _, suffix := range invalid {
		if _, err := LoadConfig(strings.NewReader(prefix + suffix)); err == nil {
			t.Fatalf("invalid accepted %s", suffix)
		}
	}
	if _, err := LoadConfig(strings.NewReader(prefix + "    auth:\n      bearer_env: TOKEN\n" + valid[2])); err == nil {
		t.Fatal("two auth modes accepted")
	}
}
