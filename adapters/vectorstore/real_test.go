package vectorstore

import (
	"github.com/TheAgentHealth/agenthealth/core"
	"os"
	"testing"
)

// Provisioning is deliberately external to the adapter: health checks never write.
func TestRealQdrant(t *testing.T) {
	endpoint := os.Getenv("AGENTHEALTH_QDRANT_ENDPOINT")
	if endpoint == "" {
		t.Skip("disposable Qdrant matrix not configured")
	}
	t.Setenv("VECTOR_KEY", "phase19-local-key")
	v := target(endpoint, "qdrant", true)
	v.Auth = &core.AuthReference{BearerEnv: "VECTOR_KEY"}
	if result := run(t, v); result.Status != core.Healthy {
		t.Fatalf("healthy query: %+v", result)
	}
	v.VectorStore.Collection = "absent-collection"
	if result := run(t, v); result.Status != core.Unhealthy {
		t.Fatalf("missing collection: %+v", result)
	}
	v.VectorStore.Collection = "docs"
	t.Setenv("VECTOR_KEY", "wrong-key")
	if result := run(t, v); result.Status != core.Misconfigured {
		t.Fatalf("auth denial: %+v", result)
	}
}
