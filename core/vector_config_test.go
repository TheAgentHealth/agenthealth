package core

import (
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVectorStoreFixtures(t *testing.T) {
	for _, kind := range []string{"valid", "invalid"} {
		paths, err := filepath.Glob("../tests/spec/fixtures/" + kind + "/configuration.vector*.json")
		if err != nil || len(paths) == 0 {
			t.Fatal("missing vector fixtures", err)
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
			})
		}
	}
}
func TestVectorStoreRejectsCoercions(t *testing.T) {
	prefix := "version: v1\ntargets:\n- name: vector\n  type: vector-store\n  endpoint: https://example.com\n"
	for _, suffix := range []string{
		"  vector_store: null\n",
		"  vector_store: {provider: 123, collection: docs}\n",
		"  vector_store: {provider: qdrant, collection: 123}\n",
		"  vector_store: {provider: qdrant, collection: docs, extra: true}\n",
		"  checks: [functional]\n  vector_store: {provider: qdrant, collection: docs, functional: {safe: true, vector: ['1']}}\n",
		"  checks: [functional]\n  vector_store: {provider: qdrant, collection: docs, functional: {safe: 'true', vector: [1]}}\n",
		"  checks: [functional]\n  vector_store: {provider: qdrant, collection: docs, functional: {safe: true, vector: [.nan]}}\n",
	} {
		if _, err := LoadConfig(strings.NewReader(prefix + suffix)); err == nil {
			t.Fatalf("accepted %s", suffix)
		}
	}
	for _, v := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		target := Target{Name: "vector", Type: "vector-store", Endpoint: "https://example.com", Checks: []string{"functional"}, VectorStore: &VectorStoreOptions{Provider: "qdrant", Collection: "docs", Functional: &VectorQuery{Safe: true, Vector: []float64{v}}}}
		if err := (Config{Version: "v1", Targets: []Target{target}}).Validate(); err == nil {
			t.Fatal("accepted nonfinite vector")
		}
	}
}
