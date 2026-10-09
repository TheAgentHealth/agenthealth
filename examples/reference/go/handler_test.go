package agenthealthref

import (
	"context"
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
