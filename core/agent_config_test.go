package core

import (
	"gopkg.in/yaml.v3"
	"strings"
	"testing"
)

func TestAgentConfiguration(t *testing.T) {
	base := `version: v1
targets:
  - name: first
    type: agent
    endpoint: http://localhost/health
    checks: [functional]
    agent:
      functional:
        safe: true
        text: Reply OK
`
	for _, tc := range []struct {
		value string
		valid bool
	}{
		{base, true}, {strings.Replace(base, "safe: true", "safe: false", 1), false},
		{strings.Replace(base, "safe: true", "safe: 'true'", 1), false},
		{strings.Replace(base, "type: agent", "type: http", 1), false},
		{strings.Replace(base, "checks: [functional]", "checks: [protocol]", 1), false},
		{strings.Replace(base, "text: Reply OK", "text: ''", 1), false},
		{strings.Replace(base, "text: Reply OK", "text: null", 1), false},
	} {
		_, err := LoadConfig(strings.NewReader(tc.value))
		if (err == nil) != tc.valid {
			t.Fatalf("valid=%v err=%v", tc.valid, err)
		}
	}
}

func TestAgentUnicodeBounds(t *testing.T) {
	for _, tc := range []struct {
		name, text, downstream string
		valid                  bool
	}{
		{"empty selector", "health", "", true},
		{"selector unicode boundary", "health", strings.Repeat("界", 1024), true},
		{"selector too long", "health", strings.Repeat("界", 1025), false},
		{"selector whitespace", "health", "  ", false},
		{"text byte boundary", strings.Repeat("é", 32768), "", true},
		{"text above byte boundary", strings.Repeat("é", 32769), "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			target := Target{Name: "first", Type: "agent", Endpoint: "http://localhost/health", Checks: []string{"functional"}, Agent: &AgentOptions{Functional: &AgentTask{Safe: true, Text: tc.text, Downstream: tc.downstream}}}
			err := (Config{Version: "v1", Targets: []Target{target}}).Validate()
			if (err == nil) != tc.valid {
				t.Fatalf("valid=%v err=%v", tc.valid, err)
			}
			// Exercise strict YAML loading as well as Go-built configurations.
			encoded, err := yaml.Marshal(map[string]any{"version": "v1", "targets": []any{map[string]any{"name": target.Name, "type": target.Type, "endpoint": target.Endpoint, "checks": target.Checks, "agent": map[string]any{"functional": target.Agent.Functional}}}})
			if err != nil {
				t.Fatal(err)
			}
			_, err = LoadConfig(strings.NewReader(string(encoded)))
			if (err == nil) != tc.valid {
				t.Fatalf("YAML valid=%v err=%v", tc.valid, err)
			}
		})
	}
}
