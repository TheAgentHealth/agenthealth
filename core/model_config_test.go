package core

import (
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestModelConfigurationFixtures(t *testing.T) {
	for _, kind := range []string{"valid", "invalid"} {
		paths, err := filepath.Glob("../tests/spec/fixtures/" + kind + "/configuration.model*.json")
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
	prefix := "version: v1\ntargets:\n  - name: llm\n    type: model\n    endpoint: https://example.com\n"
	for _, suffix := range []string{"    model: null\n", "    model:\n      api: 123\n", "    model:\n      required_models: [123]\n", "    model:\n      unknown: true\n", "    checks: [functional]\n    model:\n      functional:\n        safe: true\n        model: gpt\n        prompt: 123\n", "    checks: [functional]\n    model:\n      functional:\n        safe: true\n        model: [gpt]\n        prompt: OK\n", "    checks: [functional]\n    model:\n      functional:\n        safe: \"true\"\n        model: gpt\n        prompt: OK\n", "    checks: [functional]\n    model:\n      functional:\n        safe: true\n        model: gpt\n        prompt: OK\n        max_output_tokens: \"8\"\n"} {
		if _, err := LoadConfig(strings.NewReader(prefix + suffix)); err == nil {
			t.Fatalf("accepted invalid YAML: %s", suffix)
		}
	}
}
