package core

import (
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
