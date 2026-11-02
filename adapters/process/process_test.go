package process

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/TheAgentHealth/agenthealth/core"
)

func TestWorker(t *testing.T) {
	// A copied test executable acts as the protocol peer on every native platform.
	if !strings.Contains(filepath.Base(os.Args[0]), "worker") {
		return
	}
	var req WireRequest
	if json.NewDecoder(os.Stdin).Decode(&req) != nil {
		os.Exit(2)
	}
	switch req.Endpoint {
	case "hang":
		time.Sleep(time.Minute)
	case "partial":
		os.Stdout.WriteString("{\"kind\":\"response\"}\n")
		time.Sleep(time.Minute)
	case "oversized":
		os.Stdout.WriteString(strings.Repeat(" ", maxBytes) + "{\"kind\":\"result\",\"status\":\"HEALTHY\"}")
	case "trailing":
		os.Stdout.WriteString("{\"kind\":\"result\",\"status\":\"HEALTHY\"}" + strings.Repeat(" ", maxBytes))
	case "extra":
		os.Stdout.WriteString("{\"kind\":\"result\",\"status\":\"HEALTHY\",\"message\":\"secret\"}")
	case "duplicate":
		os.Stdout.WriteString("{\"kind\":\"result\",\"status\":\"HEALTHY\"}\n{}")
	case "environment":
		if os.Getenv("PHASE23_SECRET") != "" || req.Credential != "explicit" {
			os.Exit(3)
		}
		os.Stdout.WriteString("{\"kind\":\"result\",\"status\":\"HEALTHY\"}")
	default:
		os.Stdout.WriteString("{\"kind\":\"response\"}\n{\"kind\":\"result\",\"status\":\"HEALTHY\"}")
	}
	os.Exit(0)
}

func worker(t *testing.T) Manifest {
	t.Helper()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(executable)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "worker.test")
	if err := os.WriteFile(path, data, 0700); err != nil {
		t.Fatal(err)
	}
	hash := sha256.Sum256(data)
	return Manifest{Protocol: "v1", Command: path, SHA256: hex.EncodeToString(hash[:]), Metadata: core.Metadata{Name: "test", Version: "1", CompatibilityVersion: "v1", TargetTypes: []string{"custom"}, Dimensions: []string{"configuration", "reachability"}}}
}

func TestProtocol(t *testing.T) {
	m := worker(t)
	for _, endpoint := range []string{"success", "environment", "extra", "duplicate", "oversized", "trailing", "hang", "partial"} {
		t.Run(endpoint, func(t *testing.T) {
			t.Setenv("PHASE23_SECRET", "must-not-inherit")
			a, err := New(m)
			if err != nil {
				t.Fatal(err)
			}
			budget := 3 * time.Second
			if endpoint == "hang" || endpoint == "partial" {
				budget = 500 * time.Millisecond
			}
			ctx, cancel := context.WithTimeout(context.Background(), budget)
			defer cancel()
			marked := false
			obs, err := a.Check(ctx, core.Request{Target: core.Target{Endpoint: endpoint}, Credential: "explicit", MarkResponse: func() { marked = true }}, "reachability")
			valid := endpoint == "success" || endpoint == "environment"
			if (endpoint == "hang" || endpoint == "partial") && !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("lost deadline classification: %v", err)
			}
			if (err == nil) != valid {
				t.Fatalf("valid=%v error=%v", valid, err)
			}
			if valid && obs.Check.Status != core.Healthy {
				t.Fatal(obs)
			}
			if endpoint == "partial" && (!marked || !obs.ResponseReceived) {
				t.Fatal("partial response lost")
			}
		})
	}
}

func TestManifestAndPin(t *testing.T) {
	m := worker(t)
	a, err := New(m)
	if err != nil {
		t.Fatal(err)
	}
	m.Metadata.Dimensions[0] = "bad"
	metadata := a.Metadata()
	metadata.TargetTypes[0] = "http"
	if a.Metadata().TargetTypes[0] != "custom" || a.Metadata().Dimensions[0] != "configuration" {
		t.Fatal("mutable metadata")
	}
	if err := os.WriteFile(m.Command, []byte("changed"), 0700); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Check(context.Background(), core.Request{}, "configuration"); err == nil {
		t.Fatal("changed executable accepted")
	}
	for _, change := range []func(*Manifest){
		func(m *Manifest) { m.Protocol = "v2" }, func(m *Manifest) { m.Command = "relative" }, func(m *Manifest) { m.SHA256 = "bad" },
		func(m *Manifest) { m.SHA256 = strings.Repeat("A", 64) }, func(m *Manifest) { m.Metadata.CompatibilityVersion = "v2" },
	} {
		candidate := a.manifest
		change(&candidate)
		if _, err := New(candidate); err == nil {
			t.Fatal("invalid manifest accepted")
		}
	}
	path := filepath.Join(t.TempDir(), "manifest.json")
	raw, _ := json.Marshal(a.manifest)
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err != nil {
		t.Fatal(err)
	}
	for _, raw := range []string{string(raw) + "{}", `{"unknown":true}`, strings.Repeat(" ", maxBytes+1)} {
		os.WriteFile(path, []byte(raw), 0600)
		if _, err := Load(path); err == nil {
			t.Fatal("invalid JSON accepted")
		}
	}
}

func TestNoGroupSignalDuringOrAfterReaping(t *testing.T) {
	groups, direct := 0, 0
	control := processControl{group: func() error { groups++; return nil }, direct: func() error { direct++; return nil }}
	if err := control.cancel(); err != nil {
		t.Fatal(err)
	}
	control.beginWait()
	for i := 0; i < 3; i++ {
		if err := control.cancel(); err != nil {
			t.Fatal(err)
		}
	}
	if groups != 1 || direct != 3 {
		t.Fatalf("group=%d direct=%d", groups, direct)
	}
}
