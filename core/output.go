package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"sort"
	"strings"
)

// ValidateResult verifies the wire shape before emitting it. Adapters remain
// responsible for redacting messages before handing them to the output layer.
func ValidateResult(r Result) error {
	if r.Target.Name == "" || !contains(targetTypes, r.Target.Type) || !r.Status.Valid() {
		return errors.New("invalid result identity or status")
	}
	if r.Checks == nil || r.Dependencies == nil {
		return errors.New("checks and dependencies must be initialized")
	}
	if len(r.Checks) == 0 && r.Status != Unknown && r.Status != Misconfigured {
		return errors.New("result status requires check evidence")
	}
	if r.LatencyMS != nil && (*r.LatencyMS < 0 || math.IsNaN(*r.LatencyMS) || math.IsInf(*r.LatencyMS, 0)) {
		return errors.New("invalid result latency")
	}
	for dimension, check := range r.Checks {
		if !contains(dimensions, dimension) || !check.Status.Valid() {
			return errors.New("invalid check dimension or status")
		}
	}
	for _, dep := range r.Dependencies {
		if err := ValidateResult(dep); err != nil {
			return err
		}
	}
	return nil
}
func validateResults(results []Result) error {
	if len(results) == 0 {
		return errors.New("output requires at least one result")
	}
	for _, r := range results {
		if err := ValidateResult(r); err != nil {
			return err
		}
	}
	return nil
}

// WriteJSON emits a single result document or an ordered batch envelope.
func WriteJSON(w io.Writer, results []Result) error {
	if err := validateResults(results); err != nil {
		return err
	}
	sanitized := make([]Result, len(results))
	for i, r := range results {
		sanitized[i] = NewRedactor().Result(r)
	}
	results = sanitized
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if len(results) == 1 {
		return encoder.Encode(struct {
			SpecVersion string `json:"spec_version"`
			Result
		}{"v1", results[0]})
	}
	return encoder.Encode(struct {
		SpecVersion string   `json:"spec_version"`
		Results     []Result `json:"results"`
	}{"v1", results})
}

// WriteHuman presents the same result tree without terminal control characters.
func WriteHuman(w io.Writer, results []Result) error {
	if err := validateResults(results); err != nil {
		return err
	}
	for _, result := range results {
		if err := writeHumanResult(w, NewRedactor().Result(result), 0); err != nil {
			return err
		}
	}
	return nil
}
func printable(s string) string {
	return strings.Map(func(r rune) rune {
		if r < 32 || (r >= 127 && r <= 159) {
			return ' '
		}
		return r
	}, s)
}
func writeHumanResult(w io.Writer, r Result, depth int) error {
	indent := strings.Repeat("  ", depth)
	if _, err := fmt.Fprintf(w, "%s%s (%s): %s\n", indent, printable(r.Target.Name), r.Target.Type, r.Status); err != nil {
		return err
	}
	if r.LatencyMS != nil {
		if _, err := fmt.Fprintf(w, "%s  latency_ms: %.2f\n", indent, *r.LatencyMS); err != nil {
			return err
		}
	}
	keys := make([]string, 0, len(r.Checks))
	for key := range r.Checks {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		c := r.Checks[key]
		if _, err := fmt.Fprintf(w, "%s  %s: %s %s\n", indent, key, c.Status, printable(c.Message)); err != nil {
			return err
		}
	}
	for _, dep := range r.Dependencies {
		if err := writeHumanResult(w, dep, depth+1); err != nil {
			return err
		}
	}
	return nil
}
