package core

import (
	"strings"
	"testing"
)

func TestHTTPConfiguration(t *testing.T) {
	prefix := "version: v1\ntargets:\n  - name: http\n    type: http\n    endpoint: https://example.com\n"
	valid := []string{
		"    http: {}\n",
		"    http:\n      expected_status: [200, 204]\n      headers:\n        Content-Type: application/json\n",
		"    checks: [functional]\n    http:\n      body_contains: ready\n      max_body_bytes: 1024\n",
	}
	for _, suffix := range valid {
		if _, err := LoadConfig(strings.NewReader(prefix + suffix)); err != nil {
			t.Fatalf("valid config rejected: %s %v", suffix, err)
		}
	}
	invalid := []string{
		"    http: null\n",
		"    http:\n      unknown: true\n",
		"    http:\n      expected_status: []\n",
		"    http:\n      expected_status: [199]\n",
		"    http:\n      expected_status: [600]\n",
		"    http:\n      expected_status: ['200']\n",
		"    http:\n      body_contains: ready\n",
		"    http:\n      max_body_bytes: 1048577\n",
		"    http:\n      headers:\n        X-Ready: true\n",
		"    http:\n      headers:\n        X-Ready: yes\n        x-ready: no\n",
		"    http:\n      headers:\n        'bad name': value\n",
		"    checks: [reachability]\n    http:\n      expected_status: [200]\n",
	}
	for _, suffix := range invalid {
		if _, err := LoadConfig(strings.NewReader(prefix + suffix)); err == nil {
			t.Fatalf("invalid config accepted: %s", suffix)
		}
	}
	if _, err := LoadConfig(strings.NewReader(strings.Replace(prefix, "type: http", "type: mcp", 1) + valid[0])); err == nil {
		t.Fatal("HTTP options accepted on MCP")
	}
}
