package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/TheAgentHealth/agenthealth/core"
	"gopkg.in/yaml.v3"
)

func TestPingFormats(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "HEAD" {
			t.Errorf("unexpected active request: %s", r.Method)
		}
		w.WriteHeader(200)
	}))
	defer server.Close()
	for _, format := range []string{"terminal", "json", "yaml"} {
		t.Run(format, func(t *testing.T) {
			var out, diagnostic bytes.Buffer
			code := run(context.Background(), []string{"ping", "http", server.URL, "--format", format}, &out, &diagnostic)
			if code != 0 || diagnostic.Len() != 0 {
				t.Fatalf("code %d: %s", code, diagnostic.String())
			}
			if format == "terminal" {
				return
			}
			var document map[string]any
			var err error
			if format == "json" {
				err = json.Unmarshal(out.Bytes(), &document)
			} else {
				err = yaml.Unmarshal(out.Bytes(), &document)
			}
			if err != nil || document["spec_version"] != "v1" || document["status"] != "HEALTHY" {
				t.Fatalf("invalid document: %s (%v)", out.String(), err)
			}
			if _, ok := document["latency_ms"]; !ok {
				t.Fatal("missing wire field latency_ms")
			}
		})
	}
}

func TestBatchAndDoctor(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("version: v1\ntargets:\n  - name: unsupported\n    type: custom\n    endpoint: http://localhost\n  - name: inconclusive\n    type: http\n    endpoint: http://localhost\n    checks: []\n"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, command := range []string{"check", "doctor"} {
		var out, diagnostic bytes.Buffer
		if code := run(context.Background(), []string{command, path, "--format=json"}, &out, &diagnostic); code != 5 {
			t.Fatalf("code %d: %s", code, diagnostic.String())
		}
		var document struct {
			Results []core.Result `json:"results"`
		}
		if err := json.Unmarshal(out.Bytes(), &document); err != nil {
			t.Fatal(err)
		}
		if len(document.Results) != 2 || document.Results[0].Status != core.Misconfigured || document.Results[1].Status != core.Unknown {
			t.Fatalf("unexpected results: %+v", document)
		}
	}
	var out, diagnostic bytes.Buffer
	if code := run(context.Background(), []string{"doctor", path}, &out, &diagnostic); code != 5 || !bytes.Contains(out.Bytes(), []byte("Doctor:")) {
		t.Fatal("missing doctor advice")
	}
}

func TestInvocationFailures(t *testing.T) {
	for _, args := range [][]string{nil, {"bad"}, {"ping"}, {"check", "missing.yaml"}, {"ping", "bad", "http://localhost"}, {"version", "extra"}, {"version", "--format=bad"}, {"check", "--format"}, {"--unknown"}} {
		var out, diagnostic bytes.Buffer
		if code := run(context.Background(), args, &out, &diagnostic); code != 6 || out.Len() != 0 || diagnostic.Len() == 0 {
			t.Fatalf("%v: code %d stdout %s stderr %s", args, code, out.String(), diagnostic.String())
		}
	}
}

func TestExitCodes(t *testing.T) {
	for code, status := range []core.Status{core.Healthy, core.Degraded, core.Unhealthy, core.Unreachable, core.Misconfigured, core.Unknown} {
		if exitCode(status) != code {
			t.Fatalf("wrong exit code for %s", status)
		}
	}
}

type brokenWriter struct{}

func (brokenWriter) Write([]byte) (int, error) { return 0, os.ErrPermission }

func TestOutputFailure(t *testing.T) {
	var diagnostic bytes.Buffer
	if code := run(context.Background(), []string{"version"}, brokenWriter{}, &diagnostic); code != 6 {
		t.Fatalf("code %d", code)
	}
}

func TestMCPPingAndDoctor(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "HEAD" {
			w.WriteHeader(405)
			return
		}
		var request struct {
			ID     int    `json:"id"`
			Method string `json:"method"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
			w.WriteHeader(400)
			return
		}
		if request.Method == "server/discover" {
			w.WriteHeader(400)
			return
		}
		if request.Method == "notifications/initialized" {
			w.WriteHeader(202)
			return
		}
		if request.Method != "initialize" {
			t.Errorf("unexpected passive method %s", request.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": request.ID, "result": map[string]any{"protocolVersion": "2025-11-25", "serverInfo": map[string]string{"name": "test", "version": "1"}, "capabilities": map[string]any{}}})
	}))
	defer server.Close()
	for _, command := range []string{"ping", "doctor"} {
		for _, format := range []string{"terminal", "json", "yaml"} {
			var out, diagnostic bytes.Buffer
			if code := run(context.Background(), []string{command, "mcp", server.URL, "--format", format}, &out, &diagnostic); code != 0 || diagnostic.Len() != 0 {
				t.Fatalf("%s/%s code=%d diagnostics=%s output=%s", command, format, code, &diagnostic, &out)
			}
			if command == "doctor" && format == "terminal" && !bytes.Contains(out.Bytes(), []byte("Doctor:")) {
				t.Fatal("missing advice")
			}
		}
	}
}

func TestLoginInvocation(t *testing.T) {
	for _, args := range [][]string{{"login"}, {"login", "missing.yaml", "target"}, {"login", "config.yaml", "target", "--format", "json"}} {
		var out, diagnostic bytes.Buffer
		if code := run(context.Background(), args, &out, &diagnostic); code != 6 {
			t.Fatalf("%v code=%d", args, code)
		}
	}
}

func TestA2APingAndDoctor(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			json.NewEncoder(w).Encode(map[string]any{"protocolVersion": "0.3.0", "name": "peer", "description": "Health peer", "version": "1", "url": server.URL + "/rpc", "capabilities": map[string]any{}, "skills": []any{}, "defaultInputModes": []string{"text/plain"}, "defaultOutputModes": []string{"text/plain"}})
			return
		}
		var request struct{ ID, Method string }
		json.NewDecoder(r.Body).Decode(&request)
		if request.Method != "tasks/get" {
			t.Errorf("unexpected active request %s", request.Method)
		}
		json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": request.ID, "error": map[string]any{"code": -32001, "message": "not found"}})
	}))
	defer server.Close()
	for _, command := range []string{"ping", "doctor"} {
		for _, format := range []string{"terminal", "json", "yaml"} {
			t.Run(command+"/"+format, func(t *testing.T) {
				var out, diagnostics bytes.Buffer
				code := run(context.Background(), []string{command, "a2a", server.URL, "--format", format}, &out, &diagnostics)
				if code != 0 || diagnostics.Len() != 0 {
					t.Fatalf("code %d: %s %s", code, out.String(), diagnostics.String())
				}
				if !bytes.Contains(out.Bytes(), []byte("HEALTHY")) {
					t.Fatalf("missing health: %s", out.String())
				}
			})
		}
	}
}
