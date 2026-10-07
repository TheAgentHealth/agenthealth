package core

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigurationFixtures(t *testing.T) {
	for _, kind := range []string{"valid", "invalid"} {
		paths, err := filepath.Glob("../tests/spec/fixtures/" + kind + "/configuration.*.yaml")
		if err != nil || len(paths) == 0 {
			t.Fatal("missing fixtures", err)
		}
		for _, path := range paths {
			t.Run(filepath.Base(path), func(t *testing.T) {
				_, err := LoadConfigFile(path)
				if (err == nil) != (kind == "valid") {
					t.Fatalf("unexpected validation outcome: %v", err)
				}
			})
		}
	}
}
func TestConfigControls(t *testing.T) {
	base := "version: v1\ntargets:\n  - name: example\n    type: http\n    endpoint: https://example.com\n"
	for _, suffix := range []string{"    checks: null\n", "    thresholds: {latency_ms: null}\n", "    name: 123\n", "    checks: [secret-value]\n", "    thresholds: {latency_ms: -1}\n", "---\nversion: v1\n", "    critical: false\n", "    endpoint: secret-value\n"} {
		_, err := LoadConfig(strings.NewReader(base + suffix))
		if err == nil {
			t.Fatalf("accepted invalid configuration: %s", suffix)
		}
		if strings.Contains(err.Error(), "secret-value") {
			t.Fatal("configuration diagnostic leaked supplied value")
		}
	}
	omitted, err := LoadConfig(strings.NewReader(base))
	if err != nil {
		t.Fatal(err)
	}
	empty, err := LoadConfig(strings.NewReader(base + "    checks: []\n"))
	if err != nil {
		t.Fatal(err)
	}
	if omitted.Targets[0].Checks != nil || empty.Targets[0].Checks == nil {
		t.Fatal("lost omitted versus explicitly empty checks")
	}
	if !(Dependency{}).IsCritical() {
		t.Fatal("critical must default to true")
	}
}
func TestSeverityAndDependencies(t *testing.T) {
	if Worst() != Unknown || Worst(Status("bad")) != Unknown {
		t.Fatal("absence or invalid evidence must be UNKNOWN")
	}
	criticalExpected := []Status{Healthy, Degraded, Unhealthy, Unhealthy, Unhealthy, Unknown}
	for i, status := range statuses {
		if DependencyContribution(status, true) != criticalExpected[i] {
			t.Fatal("critical mapping", status)
		}
		optional := Degraded
		if status == Healthy {
			optional = Healthy
		}
		if DependencyContribution(status, false) != optional {
			t.Fatal("optional mapping", status)
		}
		if Worst(Healthy, status) != status {
			t.Fatal("severity", status)
		}
	}
	checks := map[string]CheckResult{"reachability": {Status: Healthy}}
	deps := []Result{{Status: Unreachable}}
	if Aggregate(checks, deps, nil) != Unhealthy {
		t.Fatal("critical unreachable dependency must make parent UNHEALTHY")
	}
	if Aggregate(checks, deps, []bool{false}) != Degraded {
		t.Fatal("optional failure must degrade")
	}
	checks["dependency"] = CheckResult{Status: Unknown}
	if Aggregate(checks, nil, nil) != Healthy {
		t.Fatal("dependency summary must not count as own evidence")
	}
	if Aggregate(nil, nil, nil) != Unknown {
		t.Fatal("empty checks must remain UNKNOWN")
	}
}
func TestOutputs(t *testing.T) {
	data, err := os.ReadFile("../tests/spec/fixtures/valid/result.single.json")
	if err != nil {
		t.Fatal(err)
	}
	var result Result
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}
	for _, count := range []int{1, 2} {
		results := []Result{result}
		if count == 2 {
			results = append(results, result)
		}
		var buffer bytes.Buffer
		if err := WriteJSON(&buffer, results); err != nil {
			t.Fatal(err)
		}
		var wire map[string]json.RawMessage
		if err := json.Unmarshal(buffer.Bytes(), &wire); err != nil {
			t.Fatal(err)
		}
		if string(wire["spec_version"]) != `"v1"` {
			t.Fatal("missing spec version")
		}
		if (wire["results"] != nil) != (count == 2) {
			t.Fatal("wrong envelope")
		}
	}
	result.Target.Name = "example\x1b[31m\n"
	var buffer bytes.Buffer
	if err := WriteHuman(&buffer, []Result{result}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(buffer.String(), "\x1b") {
		t.Fatal("terminal escape leaked")
	}
	result.Checks = nil
	if WriteJSON(&buffer, []Result{result}) == nil {
		t.Fatal("accepted null checks")
	}
}
