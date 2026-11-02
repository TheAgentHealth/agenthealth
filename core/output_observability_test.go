package core

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"testing"
)

func TestObservabilityGraphAndRedaction(t *testing.T) {
	root := ciResult(Healthy)
	root.Target.ID = "private-root"
	dep := ciResult(Unhealthy)
	dep.Target.ID = "private-peer"
	latency := 125.0
	dep.LatencyMS = &latency
	yes, no := true, false
	dep.Critical = &yes
	dep.Relationship = "downstream"
	other := dep
	other.Critical = &no
	other.Relationship = "path"
	root.Dependencies = []Result{dep, other}
	records, err := observations([]Result{root})
	if err != nil {
		t.Fatal(err)
	}
	nodes, edges, checks := 0, 0, 0
	for _, r := range records {
		switch r.Kind {
		case "node":
			nodes++
		case "edge":
			edges++
		case "check":
			checks++
		}
	}
	if nodes != 2 || edges != 2 || checks != 2 {
		t.Fatalf("lost or duplicated evidence: %+v", records)
	}
	for _, format := range []string{"prometheus", "logs", "otlp", "webhook"} {
		t.Run(format, func(t *testing.T) {
			var out bytes.Buffer
			if err := WriteObservability(&out, []Result{root}, format); err != nil {
				t.Fatal(err)
			}
			for _, secret := range []string{"private", "secret", "password", "injected"} {
				if strings.Contains(out.String(), secret) {
					t.Fatalf("leaked %s", secret)
				}
			}
			switch format {
			case "prometheus":
				if strings.Count(out.String(), "agenthealth_latency_seconds{node=") != 1 || !strings.Contains(out.String(), "} 0.125\n") || !strings.Contains(out.String(), `critical="false"`) {
					t.Fatal(out.String())
				}
			case "logs":
				if strings.Count(out.String(), "\n") != len(records) {
					t.Fatal(out.String())
				}
			case "webhook":
				var payload struct {
					Version string              `json:"export_version"`
					Records []ExportObservation `json:"records"`
				}
				if json.Unmarshal(out.Bytes(), &payload) != nil || payload.Version != "v1" || len(payload.Records) != len(records) {
					t.Fatal(out.String())
				}
			case "otlp":
				var payload struct {
					ResourceLogs []struct {
						ScopeLogs []struct {
							LogRecords []struct {
								Observed string `json:"observedTimeUnixNano"`
								Body     struct {
									Value string `json:"stringValue"`
								} `json:"body"`
							} `json:"logRecords"`
						} `json:"scopeLogs"`
					} `json:"resourceLogs"`
				}
				if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
					t.Fatal(err)
				}
				logs := payload.ResourceLogs[0].ScopeLogs[0].LogRecords
				if len(logs) != len(records) || logs[0].Observed == "" {
					t.Fatal(out.String())
				}
				var record ExportObservation
				if json.Unmarshal([]byte(logs[0].Body.Value), &record) != nil || record.Kind != "node" {
					t.Fatal(out.String())
				}
			}
		})
	}
}

func TestObservabilityValidationAndFailures(t *testing.T) {
	for _, format := range []string{"prometheus", "logs", "otlp", "webhook", "unknown"} {
		for _, results := range [][]Result{nil, {{}}, {ciResult(Status("bad\nlabel"))}} {
			var out bytes.Buffer
			if WriteObservability(&out, results, format) == nil || out.Len() != 0 {
				t.Fatal("invalid export emitted")
			}
		}
		if WriteObservability(ciFailWriter{}, []Result{ciResult(Healthy)}, format) == nil {
			t.Fatal("writer error lost")
		}
	}
	for _, status := range statuses {
		for _, format := range []string{"prometheus", "logs", "otlp", "webhook"} {
			if err := WriteObservability(io.Discard, []Result{ciResult(status)}, format); err != nil {
				t.Fatal(err)
			}
		}
	}
	// Anonymous nodes must never collapse just because their names match.
	records, err := observations([]Result{ciResult(Healthy), ciResult(Healthy)})
	if err != nil || len(records) != 4 || records[0].Node == records[2].Node {
		t.Fatal(records, err)
	}
}

// Writers violating io.Writer's short-write contract must still fail the export.
type exportShortWriter struct{}

func (exportShortWriter) Write(p []byte) (int, error) { return len(p) - 1, nil }
func TestObservabilityShortWrites(t *testing.T) {
	for _, format := range []string{"prometheus", "logs", "otlp", "webhook"} {
		if err := WriteObservability(exportShortWriter{}, []Result{ciResult(Healthy)}, format); err != io.ErrShortWrite {
			t.Fatalf("%s: %v", format, err)
		}
	}
}
