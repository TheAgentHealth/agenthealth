package core

import (
	"bytes"
	"context"
	"runtime/pprof"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// Cancellation must release adapter work and pool slots across repeated runs.
func TestRepeatedCancellationReleasesCallsAndSlots(t *testing.T) {
	engineGoroutines := func() int {
		var stacks bytes.Buffer
		if err := pprof.Lookup("goroutine").WriteTo(&stacks, 2); err != nil {
			t.Fatal(err)
		}
		count := 0
		for _, stack := range strings.Split(stacks.String(), "\n\n") {
			if strings.Contains(stack, "core.(*Engine).attempt.func") || strings.Contains(stack, "core.(*Engine).closeTarget.func") || strings.Contains(stack, "core.(*Engine).runGraph.func") {
				count++
			}
		}
		return count
	}
	baseline := engineGoroutines()
	var active atomic.Int32
	adapter := &testAdapter{fn: func(ctx context.Context, _ Request, dimension string) (Observation, error) {
		if dimension == "configuration" {
			return Observation{Check: CheckResult{Status: Healthy}}, nil
		}
		active.Add(1)
		defer active.Add(-1)
		<-ctx.Done()
		return Observation{}, ctx.Err()
	}}
	registry := NewRegistry()
	if err := registry.Register(adapter); err != nil {
		t.Fatal(err)
	}
	engine := NewEngine(registry)
	timeout := 2
	target := targetForTest("reachability")
	target.TimeoutMS = &timeout
	for i := 0; i < 32; i++ {
		if _, err := engine.Run(context.Background(), Config{Version: "v1", Targets: []Target{target}}); err != nil {
			t.Fatal(err)
		}
	}
	deadline := time.Now().Add(time.Second)
	for active.Load() != 0 || len(engine.slots) != 0 || engineGoroutines() > baseline {
		if time.Now().After(deadline) {
			t.Fatalf("adapter work leaked: active=%d slots=%d", active.Load(), len(engine.slots))
		}
		time.Sleep(time.Millisecond)
	}
}
