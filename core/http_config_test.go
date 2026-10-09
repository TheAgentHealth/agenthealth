package core

import (
	"strings"
	"testing"
)

func TestHTTPConfiguration(t *testing.T) {
	prefix := "version: v1\ntargets:\n  - name: http\n    type: http\n    endpoint: https://example.com\n"
	valid := []string{
		"    http:\n      method: GET\n",
		"    http:\n      method: HEAD\n",
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
		"    http:\n      method: ''\n",
		"    http:\n      method: null\n",
		"    http:\n      method: POST\n",
		"    http:\n      method: get\n",
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

func TestRouterConfiguration(t *testing.T) {
	for _, typ := range []string{"router", "gateway"} {
		prefix := "version: v1\ntargets:\n  - name: signal\n    type: " + typ + "\n    endpoint: https://example.com/ready\n"
		for _, tc := range []struct {
			options string
			valid   bool
		}{
			{"    http:\n      expected_status: [200]\n", true},
			{"    checks: [functional]\n    http:\n      body_contains: ready\n      max_body_bytes: 1024\n", true},
			{"    http:\n      body_contains: ready\n", false},
			{"    checks: [functional]\n    http:\n      max_body_bytes: 1048577\n", false},
		} {
			_, err := LoadConfig(strings.NewReader(prefix + tc.options))
			if (err == nil) != tc.valid {
				t.Fatalf("%s valid=%v error=%v", typ, tc.valid, err)
			}
		}
	}
}
