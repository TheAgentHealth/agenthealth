package core

import (
	"bytes"
	"context"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestGraphValidation(t *testing.T) {
	for _, text := range []string{
		"- ref: missing",
		"- ref: root",
		"- ref: root\n      endpoint: test",
		"- ref: root\n      relationship: secret",
	} {
		_, err := LoadConfig(strings.NewReader("version: v1\ntargets:\n- id: root\n  name: root\n  type: custom\n  endpoint: test\n  dependencies:\n    " + text + "\n"))
		if err == nil {
			t.Fatalf("accepted %s", text)
		}
	}
	_, err := LoadConfig(strings.NewReader("version: v1\nconcurrency: 0\ntargets:\n- name: root\n  type: custom\n  endpoint: test\n"))
	if err == nil {
		t.Fatal("accepted explicit zero concurrency")
	}
}
func TestGraphSharedEvidenceAndPolicies(t *testing.T) {
	var calls atomic.Int32
	a := &testAdapter{fn: func(_ context.Context, r Request, d string) (Observation, error) {
		status := Healthy
		if r.Target.ID == "backend" {
			calls.Add(1)
			status = Unhealthy
		}
		return Observation{Check: CheckResult{Status: status}}, nil
	}}
	registry := NewRegistry()
	if err := registry.Register(a); err != nil {
		t.Fatal(err)
	}
	backend := targetForTest("configuration")
	backend.ID = "backend"
	parent := targetForTest("configuration", "dependency")
	parent.ID = "first"
	optional := false
	parent.Dependencies = []Dependency{{Ref: "backend", Relationship: "downstream", Critical: &optional}, {Ref: "backend", Relationship: "path"}}
	c := Config{Version: "v1", Concurrency: 1, Targets: []Target{parent, backend}}
	results, err := NewEngine(registry).Run(context.Background(), c)
	if err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 1 || results[0].Status != Unhealthy || results[0].Dependencies[0].Relationship != "downstream" || results[0].Dependencies[0].IsCriticalForTest() {
		t.Fatal(results, calls.Load())
	}
	if err := ValidateResult(results[0]); err != nil {
		t.Fatal(err)
	}
	c.Targets[0].Dependencies = c.Targets[0].Dependencies[:1]
	results, err = NewEngine(registry).Run(context.Background(), c)
	if err != nil || results[0].Status != Degraded {
		t.Fatal(results, err)
	}
}
func (r Result) IsCriticalForTest() bool { return r.Critical == nil || *r.Critical }
func TestGraphParallelBoundAndBudget(t *testing.T) {
	var inFlight, peak atomic.Int32
	a := &testAdapter{fn: func(ctx context.Context, _ Request, _ string) (Observation, error) {
		n := inFlight.Add(1)
		defer inFlight.Add(-1)
		for old := peak.Load(); n > old; old = peak.Load() {
			if peak.CompareAndSwap(old, n) {
				break
			}
		}
		select {
		case <-time.After(20 * time.Millisecond):
			return Observation{Check: CheckResult{Status: Healthy}}, nil
		case <-ctx.Done():
			return Observation{}, ctx.Err()
		}
	}}
	registry := NewRegistry()
	registry.Register(a)
	targets := []Target{}
	for i := 0; i < 6; i++ {
		targets = append(targets, targetForTest("configuration"))
	}
	_, err := NewEngine(registry).Run(context.Background(), Config{Version: "v1", Concurrency: 2, Targets: targets})
	if err != nil {
		t.Fatal(err)
	}
	if peak.Load() != 2 {
		t.Fatal(peak.Load())
	}
	ms := 1
	targets[0].BudgetMS = &ms
	results, err := NewEngine(registry).Run(context.Background(), Config{Version: "v1", Targets: targets[:1]})
	if err != nil || results[0].Status == Healthy {
		t.Fatal(results, err)
	}
}

func TestGraphIndirectCyclesDuplicatesAndProjectionBound(t *testing.T) {
	a, b := targetForTest("configuration"), targetForTest("configuration")
	a.ID = "a"
	b.ID = "b"
	a.Dependencies = []Dependency{{Ref: "b"}}
	b.Dependencies = []Dependency{{Ref: "a"}}
	if err := (Config{Version: "v1", Targets: []Target{a, b}}).Validate(); err == nil {
		t.Fatal("accepted indirect cycle")
	}
	a.Dependencies = nil
	b.Dependencies = nil
	b.ID = "a"
	if err := (Config{Version: "v1", Targets: []Target{a, b}}).Validate(); err == nil {
		t.Fatal("accepted duplicate id")
	}
	targets := []Target{}
	previous := ""
	for i := 0; i < 14; i++ {
		node := targetForTest("configuration")
		node.ID = string(rune('a' + i))
		if previous != "" {
			node.Dependencies = []Dependency{{Ref: previous}, {Ref: previous}}
		}
		previous = node.ID
		targets = append(targets, node)
	}
	if err := (Config{Version: "v1", Targets: targets}).Validate(); err == nil {
		t.Fatal("accepted excessive projection")
	}
}
func TestGraphUncooperativeCallsRetainConcurrencySlots(t *testing.T) {
	release := make(chan struct{})
	defer close(release)
	var calls atomic.Int32
	adapter := &testAdapter{fn: func(context.Context, Request, string) (Observation, error) {
		calls.Add(1)
		<-release
		return Observation{Check: CheckResult{Status: Healthy}}, nil
	}}
	registry := NewRegistry()
	registry.Register(adapter)
	targets := []Target{}
	ms := 5
	for i := 0; i < 6; i++ {
		node := targetForTest("configuration")
		node.BudgetMS = &ms
		targets = append(targets, node)
	}
	_, err := NewEngine(registry).Run(context.Background(), Config{Version: "v1", Concurrency: 2, Targets: targets})
	if err != nil || calls.Load() != 2 {
		t.Fatal(err, calls.Load())
	}
}
func TestGraphSharedActiveCheckOnceAndFreshRuns(t *testing.T) {
	var calls atomic.Int32
	adapter := &testAdapter{fn: func(_ context.Context, _ Request, d string) (Observation, error) {
		if d == "functional" {
			calls.Add(1)
		}
		return Observation{Check: CheckResult{Status: Healthy}}, nil
	}}
	registry := NewRegistry()
	registry.Register(adapter)
	engine := NewEngine(registry)
	node := targetForTest("functional")
	node.ID = "probe"
	parent := targetForTest("configuration")
	parent.Dependencies = []Dependency{{Ref: "probe"}, {Ref: "probe"}}
	config := Config{Version: "v1", Targets: []Target{parent, node}}
	for i := 1; i <= 2; i++ {
		if _, err := engine.Run(context.Background(), config); err != nil {
			t.Fatal(err)
		}
		if calls.Load() != int32(i) {
			t.Fatal(calls.Load())
		}
	}
}

func TestGraphParentBudgetDoesNotCancelSharedDependency(t *testing.T) {
	adapter := &testAdapter{fn: func(ctx context.Context, r Request, _ string) (Observation, error) {
		delay := 20 * time.Millisecond
		if r.Target.ID == "parent" {
			delay = time.Second
		}
		select {
		case <-time.After(delay):
			return Observation{Check: CheckResult{Status: Healthy}}, nil
		case <-ctx.Done():
			return Observation{}, ctx.Err()
		}
	}}
	registry := NewRegistry()
	if err := registry.Register(adapter); err != nil {
		t.Fatal(err)
	}
	parent, child := targetForTest("configuration"), targetForTest("configuration")
	parent.ID = "parent"
	child.ID = "child"
	ms := 1
	parent.BudgetMS = &ms
	parent.Dependencies = []Dependency{{Ref: "child", Relationship: "supporting"}}
	results, err := NewEngine(registry).Run(context.Background(), Config{Version: "v1", Targets: []Target{parent, child}})
	if err != nil {
		t.Fatal(err)
	}
	if results[0].Checks["configuration"].Status == Healthy || results[0].Dependencies[0].Status != Healthy || results[1].Status != Healthy {
		t.Fatal(results)
	}
}

func TestGraphRedactsSecretsRegisteredByLaterNodes(t *testing.T) {
	secret := "late-acquired-secret"
	firstDone := make(chan struct{})
	adapter := &testAdapter{fn: func(_ context.Context, r Request, _ string) (Observation, error) {
		if r.Target.ID == "first" {
			close(firstDone)
		} else {
			<-firstDone
			r.RememberSecret(secret)
		}
		return Observation{Check: CheckResult{Status: Healthy}}, nil
	}}
	registry := NewRegistry()
	if err := registry.Register(adapter); err != nil {
		t.Fatal(err)
	}
	first, other := targetForTest("configuration"), targetForTest("configuration")
	first.ID = "first"
	first.Name = secret
	other.ID = "other"
	results, err := NewEngine(registry).Run(context.Background(), Config{Version: "v1", Targets: []Target{first, other}})
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := WriteJSON(&out, results); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), secret) {
		t.Fatal("later registered secret leaked")
	}
}
