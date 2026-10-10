package agenthealthref

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TheAgentHealth/agenthealth/adapters/agent"
	"github.com/TheAgentHealth/agenthealth/core"
)

func TestHandlerContractWithAdapter(t *testing.T) {
	server := httptest.NewServer(Handler(func() Document {
		return Document{Name: "reference", Live: true, Ready: true, Capabilities: []string{"health"}}
	},
		func(_ context.Context, task Task) (Outcome, error) {
			return Outcome{Completed: true, Success: task.Text == "health" && task.Downstream == ""}, nil
		}))
	defer server.Close()
	registry := core.NewRegistry()
	if err := registry.Register(agent.Adapter{}); err != nil {
		t.Fatal(err)
	}
	config := core.Config{Version: "v1", Targets: []core.Target{{Name: "reference", Type: "agent", Endpoint: server.URL, Checks: []string{"protocol", "capability", "functional"}, Agent: &core.AgentOptions{RequiredCapabilities: []string{"health"}, Functional: &core.AgentTask{Safe: true, Text: "health"}}}}}
	results, err := core.NewEngine(registry).Run(context.Background(), config)
	if err != nil || results[0].Status != core.Healthy {
		t.Fatal(results, err)
	}
}
func TestRejectUnsafeAndExtraTaskDocuments(t *testing.T) {
	h := Handler(func() Document { return Document{} }, func(context.Context, Task) (Outcome, error) { t.Fatal("unsafe probe executed"); return Outcome{}, nil })
	for _, body := range []string{`{"safe":false,"text":"health"}`, `{"safe":true,"text":"health"} {}`, `{"safe":true,"text":"health","command":"sh"}`} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("POST", "/health", strings.NewReader(body)))
		if w.Code != 400 {
			t.Fatal(w.Code)
		}
	}
}

func TestTaskSizeBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name, text, downstream string
		status                 int
	}{
		{"maximum ASCII", strings.Repeat("x", 65536), "", 200},
		{"maximum escaped text", strings.Repeat("\x00", 65536), strings.Repeat("😀", 1024), 200},
		{"maximum UTF8", strings.Repeat("😀", 16384), "", 200},
		{"text too long", strings.Repeat("x", 65537), "", 400},
		{"UTF8 too long", strings.Repeat("😀", 16384) + "x", "", 400},
		{"downstream too long", "health", strings.Repeat("😀", 1025), 400},
		{"blank text", " ", "", 400},
		{"blank downstream", "health", " ", 400},
	} {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			h := Handler(func() Document { return Document{} }, func(_ context.Context, task Task) (Outcome, error) {
				called = true
				return Outcome{Completed: true, Success: true, Downstream: task.Downstream}, nil
			})
			body, err := json.Marshal(Task{Safe: true, Text: tc.text, Downstream: tc.downstream})
			if err != nil {
				t.Fatal(err)
			}
			w := httptest.NewRecorder()
			h.ServeHTTP(w, httptest.NewRequest("POST", "/health", bytes.NewReader(body)))
			if w.Code != tc.status || called != (tc.status == 200) {
				t.Fatalf("status=%d called=%v bodyBytes=%d", w.Code, called, len(body))
			}
		})
	}
	h := Handler(func() Document { return Document{} }, func(context.Context, Task) (Outcome, error) {
		t.Fatal("oversized envelope executed")
		return Outcome{}, nil
	})
	for _, body := range []string{strings.Repeat(" ", maxTaskBodyBytes+1), `{"safe":true,"text":"health"}` + strings.Repeat(" ", maxTaskBodyBytes)} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("POST", "/health", strings.NewReader(body)))
		if w.Code != 413 {
			t.Fatal(w.Code)
		}
	}
}

func TestMaximumTaskThroughAgentAdapter(t *testing.T) {
	for _, text := range []string{strings.Repeat("x", 65536), strings.Repeat("\x00", 65536)} {
		called := false
		server := httptest.NewServer(Handler(func() Document { return Document{Name: "reference", Live: true, Ready: true} }, func(_ context.Context, task Task) (Outcome, error) {
			if task.Text != text {
				t.Error("task text changed in transit")
			}
			called = true
			return Outcome{Completed: true, Success: true}, nil
		}))
		registry := core.NewRegistry()
		if err := registry.Register(agent.Adapter{}); err != nil {
			t.Fatal(err)
		}
		config := core.Config{Version: "v1", Targets: []core.Target{{Name: "maximum-task", Type: "agent", Endpoint: server.URL, Checks: []string{"functional"}, Agent: &core.AgentOptions{Functional: &core.AgentTask{Safe: true, Text: text}}}}}
		results, err := core.NewEngine(registry).Run(context.Background(), config)
		server.Close()
		if err != nil || !called || len(results) != 1 || results[0].Status != core.Healthy {
			t.Fatal(results, err, called)
		}
	}
}
