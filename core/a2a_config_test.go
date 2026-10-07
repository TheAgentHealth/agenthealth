package core

import (
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestA2AConfigurationFixtures(t *testing.T) {
	for _, kind := range []string{"valid", "invalid"} {
		paths, err := filepath.Glob("../tests/spec/fixtures/" + kind + "/configuration.a2a*.json")
		if err != nil {
			t.Fatal(err)
		}
		for _, path := range paths {
			t.Run(filepath.Base(path), func(t *testing.T) {
				raw, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				_, err = LoadConfig(strings.NewReader(string(raw)))
				if (err == nil) != (kind == "valid") {
					t.Fatalf("unexpected validation: %v", err)
				}
				var c Config
				if yaml.Unmarshal(raw, &c) == nil && kind == "valid" {
					if err := c.Validate(); err != nil {
						t.Fatal(err)
					}
				}
			})
		}
	}
	prefix := "version: v1\ntargets:\n  - name: peer\n    type: a2a\n    endpoint: https://example.com\n"
	for _, suffix := range []string{"    a2a: null\n", "    a2a:\n      card_url: 123\n", "    a2a:\n      required_skills: [123]\n", "    a2a:\n      unknown: true\n", "    checks: [functional]\n    a2a:\n      functional:\n        safe: true\n        text: 123\n"} {
		if _, err := LoadConfig(strings.NewReader(prefix + suffix)); err == nil {
			t.Fatalf("accepted invalid YAML: %s", suffix)
		}
	}
}
