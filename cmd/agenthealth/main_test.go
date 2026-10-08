package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
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
			json.NewEncoder(w).Encode(map[string]any{"supportedInterfaces": []any{map[string]any{"url": server.URL + "/rpc", "protocolBinding": "JSONRPC", "protocolVersion": "1.0"}}, "name": "peer", "description": "Health peer", "version": "1", "url": server.URL + "/rpc", "capabilities": map[string]any{}, "skills": []any{}, "defaultInputModes": []string{"text/plain"}, "defaultOutputModes": []string{"text/plain"}})
			return
		}
		var request struct{ ID, Method string }
		json.NewDecoder(r.Body).Decode(&request)
		if request.Method != "GetTask" {
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

func TestAgentCommands(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "POST" {
			t.Error("passive command invoked a task")
		}
		w.Write([]byte(`{"version":"v1","name":"first","live":true,"ready":true,"capabilities":[],"dependencies":[]}`))
	}))
	defer server.Close()
	path := filepath.Join(t.TempDir(), "agent.yaml")
	if err := os.WriteFile(path, []byte("version: v1\ntargets:\n  - name: first\n    type: agent\n    endpoint: "+server.URL+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"ping", "agent", server.URL}, {"ping", "multi-agent", server.URL}, {"check", path}, {"doctor", path}} {
		for _, format := range []string{"terminal", "json", "yaml"} {
			var out, diagnostic bytes.Buffer
			cmd := append(append([]string{}, args...), "--format", format)
			if code := run(context.Background(), cmd, &out, &diagnostic); code != 0 || diagnostic.Len() != 0 {
				t.Fatalf("%v code=%d %s", cmd, code, diagnostic.String())
			}
		}
	}
}

func TestHelpSupportedTargets(t *testing.T) {
	var out, diagnostic bytes.Buffer
	if code := run(context.Background(), []string{"--help"}, &out, &diagnostic); code != 0 {
		t.Fatalf("code=%d %s", code, diagnostic.String())
	}
	for _, value := range []string{"agent, multi-agent, http, api, mcp, a2a", "1.0 JSON-RPC", "0.3.0 compatibility", "safe probe handler"} {
		if !strings.Contains(out.String(), value) {
			t.Fatalf("help missing %q: %s", value, out.String())
		}
	}
}

func TestGatewayPingFormats(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("unexpected health method: %s", r.Method)
		}
		w.WriteHeader(200)
	}))
	defer server.Close()
	for _, format := range []string{"terminal", "json", "yaml"} {
		t.Run(format, func(t *testing.T) {
			var out, diagnostic bytes.Buffer
			code := run(context.Background(), []string{"ping", "gateway", server.URL, "--format", format}, &out, &diagnostic)
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

func TestRouterPingFormats(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("unexpected health method: %s", r.Method)
		}
		w.WriteHeader(200)
	}))
	defer server.Close()
	for _, format := range []string{"terminal", "json", "yaml"} {
		t.Run(format, func(t *testing.T) {
			var out, diagnostic bytes.Buffer
			code := run(context.Background(), []string{"ping", "router", server.URL, "--format", format}, &out, &diagnostic)
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

func TestRouterCheckAndDoctor(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(503) }))
	defer server.Close()
	path := filepath.Join(t.TempDir(), "router.yaml")
	if err := os.WriteFile(path, []byte("version: v1\ntargets:\n  - name: router\n    type: router\n    endpoint: "+server.URL+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, command := range []string{"check", "doctor"} {
		var out, diagnostic bytes.Buffer
		code := run(context.Background(), []string{command, path, "--format", "json"}, &out, &diagnostic)
		if code != 2 || diagnostic.Len() != 0 {
			t.Fatalf("%s code=%d: %s", command, code, diagnostic.String())
		}
		var document map[string]any
		if err := json.Unmarshal(out.Bytes(), &document); err != nil {
			t.Fatal(err)
		}
		if document["status"] != "UNHEALTHY" {
			t.Fatalf("unexpected result: %s", out.String())
		}
	}
}

func TestGraphAgentReferencesAndPathEvidence(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/first" {
			w.Write([]byte(`{"version":"v1","name":"first","live":true,"ready":true,"capabilities":[],"dependencies":["peer"]}`))
			return
		}
		if r.URL.Path == "/path" && r.Method == "POST" {
			w.Write([]byte(`{"completed":true,"success":false,"downstream":"peer"}`))
			return
		}
		w.Write([]byte(`{"version":"v1","name":"peer","live":true,"ready":true,"capabilities":[],"dependencies":[]}`))
	}))
	defer server.Close()
	config := "version: v1\nconcurrency: 2\ntargets:\n- id: first\n  name: first\n  type: agent\n  endpoint: " + server.URL + "/first\n  dependencies:\n  - ref: peer\n    relationship: downstream\n  - ref: path\n    relationship: path\n- id: peer\n  name: peer\n  type: agent\n  endpoint: " + server.URL + "/peer\n- id: path\n  name: communication\n  type: agent\n  endpoint: " + server.URL + "/path\n  checks: [configuration, functional]\n  agent:\n    functional:\n      safe: true\n      text: health\n      downstream: peer\n"
	path := filepath.Join(t.TempDir(), "graph.yaml")
	if err := os.WriteFile(path, []byte(config), 0600); err != nil {
		t.Fatal(err)
	}
	for _, format := range []string{"terminal", "json", "yaml"} {
		var out, diagnostic bytes.Buffer
		if code := run(context.Background(), []string{"check", path, "--format", format}, &out, &diagnostic); code != 2 {
			t.Fatal(code, diagnostic.String(), out.String())
		}
		if format == "terminal" {
			if !strings.Contains(out.String(), "[path]") {
				t.Fatal(out.String())
			}
			continue
		}
		var envelope struct {
			Results []core.Result `json:"results"`
		}
		data := out.Bytes()
		if format == "yaml" {
			var doc any
			if err := yaml.Unmarshal(data, &doc); err != nil {
				t.Fatal(err)
			}
			data, _ = json.Marshal(doc)
		}
		if err := json.Unmarshal(data, &envelope); err != nil {
			t.Fatal(err)
		}
		first := envelope.Results[0]
		if first.Status != core.Unhealthy || first.Checks["capability"].Status != core.Healthy || first.Dependencies[0].Status != core.Healthy || first.Dependencies[1].Checks["functional"].Status != core.Unhealthy {
			t.Fatal(first)
		}
	}
}

func TestGraphGatewayRouterSharedBackendAndFailedRoutes(t *testing.T) {
	var counts [4]atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/gateway":
			counts[0].Add(1)
		case "/router":
			counts[1].Add(1)
		case "/backend":
			counts[2].Add(1)
		case "/route":
			counts[3].Add(1)
			w.WriteHeader(503)
		}
	}))
	defer server.Close()
	declarations := []string{
		"- id: gateway\n  name: gateway\n  type: gateway\n  endpoint: " + server.URL + "/gateway\n  checks: [configuration, protocol]\n",
		"- id: router\n  name: router\n  type: router\n  endpoint: " + server.URL + "/router\n  checks: [configuration, protocol]\n",
		"- id: backend\n  name: backend\n  type: http\n  endpoint: " + server.URL + "/backend\n  checks: [configuration, protocol]\n",
		"- id: route\n  name: route\n  type: http\n  endpoint: " + server.URL + "/route\n  checks: [configuration, protocol]\n",
	}
	path := filepath.Join(t.TempDir(), "signals.yaml")
	// Compare request counts with isolated execution: multiple graph references
	// must not issue additional backend probes or erase failed route evidence.
	var baseline [4]int32
	for i, node := range declarations {
		if err := os.WriteFile(path, []byte("version: v1\ntargets:\n"+node), 0600); err != nil {
			t.Fatal(err)
		}
		var out, diagnostic bytes.Buffer
		expected := 0
		if i == 3 {
			expected = 2
		}
		if code := run(context.Background(), []string{"check", path, "--format", "json"}, &out, &diagnostic); code != expected {
			t.Fatal(code, diagnostic.String())
		}
		baseline[i] = counts[i].Swap(0)
	}
	edges := "  dependencies:\n  - ref: backend\n    relationship: downstream\n  - ref: route\n    relationship: path\n"
	graph := "version: v1\nconcurrency: 2\ntargets:\n" + declarations[0] + edges + declarations[1] + edges + declarations[2] + declarations[3]
	if err := os.WriteFile(path, []byte(graph), 0600); err != nil {
		t.Fatal(err)
	}
	var out, diagnostic bytes.Buffer
	if code := run(context.Background(), []string{"check", path, "--format", "json"}, &out, &diagnostic); code != 2 {
		t.Fatal(code, diagnostic.String())
	}
	var envelope struct {
		Results []core.Result `json:"results"`
	}
	if err := json.Unmarshal(out.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		r := envelope.Results[i]
		if r.Status != core.Unhealthy || r.Checks["protocol"].Status != core.Healthy || r.Dependencies[0].Status != core.Healthy || r.Dependencies[1].Status != core.Unhealthy {
			t.Fatal(r)
		}
	}
	for i := range counts {
		if counts[i].Load() != baseline[i] {
			t.Fatalf("signal %d: isolated=%d graph=%d", i, baseline[i], counts[i].Load())
		}
	}
}
