package core

import (
	"context"
	"fmt"
	"testing"
)

// Wide shared DAGs exercise node deduplication, edge projection and scheduling.
func BenchmarkLargeDAG(b *testing.B) {
	for _, size := range []int{100, 1000, 5000} {
		b.Run(fmt.Sprint(size), func(b *testing.B) {
			registry := NewRegistry()
			adapter := &testAdapter{fn: func(context.Context, Request, string) (Observation, error) {
				return Observation{Check: CheckResult{Status: Healthy}}, nil
			}}
			if err := registry.Register(adapter); err != nil {
				b.Fatal(err)
			}
			nodes := make([]Target, size)
			for i := range nodes {
				nodes[i] = targetForTest("configuration")
				nodes[i].ID = fmt.Sprintf("node-%d", i)
				if i > 0 {
					nodes[i].Dependencies = []Dependency{{Ref: "node-0"}}
				}
			}
			config := Config{Version: "v1", Targets: nodes}
			if err := config.Validate(); err != nil {
				b.Fatal(err)
			}
			engine := NewEngine(registry)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := engine.Run(context.Background(), config); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
