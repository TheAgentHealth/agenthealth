package core

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"regexp"
	"sync"
	"time"
)

var nodeIDPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]{0,63}$`)

func graphID(s string) bool { return nodeIDPattern.MatchString(s) }
func validRelationship(s string) bool {
	return s == "" || contains([]string{"supporting", "downstream", "path", "gateway", "router"}, s)
}
func emptyReferenceTarget(t Target) bool { return reflect.DeepEqual(t, Target{}) }

func graphNodes(c Config) (map[string]Target, error) {
	nodes := map[string]Target{}
	var visit func(Target) error
	visit = func(t Target) error {
		if t.ID != "" {
			if _, ok := nodes[t.ID]; ok {
				return errors.New("duplicate graph node id")
			}
			nodes[t.ID] = t
		}
		for _, d := range t.Dependencies {
			if d.Ref == "" {
				if err := visit(d.Target); err != nil {
					return err
				}
			}
		}
		return nil
	}
	for _, t := range c.Targets {
		if err := visit(t); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}
func validateGraph(c Config) error {
	if c.Concurrency < 0 || c.Concurrency > 16 {
		return errors.New("concurrency must be between 1 and 16 when specified")
	}
	nodes, err := graphNodes(c)
	if err != nil {
		return err
	}
	projections := 0
	var walk func(Target, map[string]bool, int) error
	walk = func(t Target, path map[string]bool, depth int) error {
		projections++
		if projections > 10000 {
			return errors.New("graph exceeds maximum 10000 result projections")
		}
		if depth > 64 {
			return errors.New("graph exceeds maximum depth 64")
		}
		if t.ID != "" {
			if path[t.ID] {
				return errors.New("dependency graph contains a cycle")
			}
			path[t.ID] = true
			defer delete(path, t.ID)
		}
		for _, d := range t.Dependencies {
			child := d.Target
			if d.Ref != "" {
				var ok bool
				child, ok = nodes[d.Ref]
				if !ok {
					return errors.New("unresolved dependency reference")
				}
			}
			if err := walk(child, path, depth+1); err != nil {
				return err
			}
		}
		return nil
	}
	for _, t := range c.Targets {
		if err := walk(t, map[string]bool{}, 0); err != nil {
			return err
		}
	}
	return nil
}

// runGraph evaluates each explicit identity once per run. Edges contribute
// independently; results remain ordered tree projections for v1 consumers.
func (e *Engine) runGraph(ctx context.Context, c Config, client *http.Client, credentials map[string]string, redactor *Redactor) []Result {
	nodes, _ := graphNodes(c)
	type entry struct {
		once   sync.Once
		result Result
	}
	cache := map[string]*entry{}
	for id := range nodes {
		cache[id] = &entry{}
	}
	concurrency := c.Concurrency
	if concurrency == 0 {
		concurrency = 16
	}
	executor := &Engine{registry: e.registry, slots: e.slots, limit: make(chan struct{}, concurrency)}
	var run func(Target) Result
	var compute func(Target) Result
	compute = func(t Target) Result {
		// Start dependencies independently of the parent's prerequisite checks.
		deps := make([]Result, len(t.Dependencies))
		critical := make([]bool, len(deps))
		var wg sync.WaitGroup
		for i, d := range t.Dependencies {
			wg.Add(1)
			go func(i int, d Dependency) {
				defer wg.Done()
				child := d.Target
				if d.Ref != "" {
					child = nodes[d.Ref]
				}
				deps[i] = run(child)
				deps[i].Relationship = d.Relationship
				required := d.IsCritical()
				critical[i] = required
				deps[i].Critical = &required
			}(i, d)
		}
		own := t
		own.Dependencies = append([]Dependency(nil), t.Dependencies...)
		for i, d := range own.Dependencies {
			if d.Ref != "" {
				own.Dependencies[i].Target = nodes[d.Ref]
			}
		}
		nodeCtx := ctx
		cancel := func() {}
		if t.BudgetMS != nil {
			nodeCtx, cancel = context.WithTimeout(ctx, time.Duration(*t.BudgetMS)*time.Millisecond)
		}
		defer cancel()
		result := executor.runOwnTarget(nodeCtx, own, client, credentials, redactor)
		wg.Wait()
		result.Target.ID = t.ID
		result.Dependencies = deps
		if _, ok := result.Checks["dependency"]; ok {
			contribution := Healthy
			for i, d := range deps {
				contribution = Worst(contribution, DependencyContribution(d.Status, critical[i]))
			}
			result.Checks["dependency"] = CheckResult{Status: contribution}
		}
		result.Status = Aggregate(result.Checks, deps, critical)
		return result
	}
	run = func(t Target) Result {
		if t.ID == "" {
			return compute(t)
		}
		item := cache[t.ID]
		item.once.Do(func() { item.result = compute(t) })
		return item.result
	}
	results := make([]Result, len(c.Targets))
	var wg sync.WaitGroup
	for i, t := range c.Targets {
		wg.Add(1)
		go func(i int, t Target) { defer wg.Done(); results[i] = run(t) }(i, t)
	}
	wg.Wait()
	// All nodes must finish secret registration before any final result is redacted.
	for i, result := range results {
		results[i] = redactor.Result(result)
	}
	return results
}
