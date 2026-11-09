package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"unicode/utf8"
)

// ConformanceMaxBytes bounds an offline evidence document before decoding.
const ConformanceMaxBytes = 4 << 20

// ConformanceReport describes evidence validation, not certification or health.
// It deliberately contains no input identities, messages, or parser errors.
type ConformanceReport struct {
	Version string `json:"conformance_version"`
	Profile string `json:"profile"`
	Passed  bool   `json:"passed"`
	Results int    `json:"results"`
	Code    string `json:"code"`
}

// TestConformance validates one JSON v1 result or batch without executing it.
// Unknown extension fields are tolerated, as in the v1 result schema.
func TestConformance(reader io.Reader) (ConformanceReport, error) {
	report := ConformanceReport{Version: "v1", Profile: "ahs-result", Code: "invalid_document"}
	data, err := io.ReadAll(io.LimitReader(reader, ConformanceMaxBytes+1))
	if err != nil {
		return report, errors.New("cannot read conformance evidence")
	}
	if len(data) > ConformanceMaxBytes {
		report.Code = "document_too_large"
		return report, nil
	}
	if !utf8.Valid(data) {
		return report, nil
	}
	// Token walking rejects duplicate keys, even inside ignored extensions.
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := uniqueJSON(decoder, 0); err != nil {
		return report, nil
	}
	if _, err := decoder.Token(); err != io.EOF {
		return report, nil
	}
	var document map[string]json.RawMessage
	if json.Unmarshal(data, &document) != nil || document == nil {
		return report, nil
	}
	var version string
	if json.Unmarshal(document["spec_version"], &version) != nil || version != "v1" {
		report.Code = "unsupported_spec_version"
		return report, nil
	}
	items := []json.RawMessage{data}
	if batch, ok := document["results"]; ok {
		if _, ambiguous := document["target"]; ambiguous {
			return report, nil
		}
		if json.Unmarshal(batch, &items) != nil || len(items) == 0 {
			return report, nil
		}
	}
	for _, item := range items {
		result, ok := conformanceResult(item, 0)
		if !ok {
			report.Code = "invalid_result"
			return report, nil
		}
		if !conformanceStatus(result) {
			report.Code = "inconsistent_status"
			return report, nil
		}
	}
	report.Passed, report.Results, report.Code = true, len(items), "conformant"
	return report, nil
}

func uniqueJSON(decoder *json.Decoder, depth int) error {
	if depth > 256 {
		return errors.New("depth")
	}
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delim, composite := token.(json.Delim)
	if !composite {
		return nil
	}
	switch delim {
	case '{':
		keys := map[string]bool{}
		for decoder.More() {
			key, err := decoder.Token()
			if err != nil {
				return err
			}
			name, ok := key.(string)
			if !ok || keys[name] {
				return errors.New("duplicate key")
			}
			keys[name] = true
			if err := uniqueJSON(decoder, depth+1); err != nil {
				return err
			}
		}
	case '[':
		for decoder.More() {
			if err := uniqueJSON(decoder, depth+1); err != nil {
				return err
			}
		}
	default:
		return errors.New("delimiter")
	}
	_, err = decoder.Token()
	return err
}

func conformanceResult(data []byte, depth int) (Result, bool) {
	var result Result
	if depth > maxResultDepth {
		return result, false
	}
	var fields map[string]json.RawMessage
	if json.Unmarshal(data, &fields) != nil || fields == nil {
		return result, false
	}
	for _, key := range []string{"target", "status", "checks", "dependencies"} {
		if value, ok := fields[key]; !ok || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return result, false
		}
	}
	// Go's decoder accepts null for scalar fields; the wire contract does not.
	for _, key := range []string{"critical", "relationship"} {
		if value, ok := fields[key]; ok && bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return result, false
		}
	}
	if value, ok := fields["relationship"]; ok {
		var relationship string
		if json.Unmarshal(value, &relationship) != nil || relationship == "" {
			return result, false
		}
	}
	var target map[string]json.RawMessage
	if json.Unmarshal(fields["target"], &target) != nil {
		return result, false
	}
	if json.Unmarshal(target["name"], &result.Target.Name) != nil || json.Unmarshal(target["type"], &result.Target.Type) != nil {
		return result, false
	}
	if value, ok := target["id"]; ok {
		var id string
		if json.Unmarshal(value, &id) != nil || !graphID(id) {
			return result, false
		}
	}
	// Decode only exact wire keys: encoding/json otherwise treats differently
	// cased extension keys as aliases and could overwrite validated evidence.
	decode := func(values map[string]json.RawMessage, key string, destination any) bool {
		value, ok := values[key]
		return !ok || json.Unmarshal(value, destination) == nil
	}
	if !decode(target, "id", &result.Target.ID) ||
		!decode(fields, "status", &result.Status) ||
		!decode(fields, "critical", &result.Critical) ||
		!decode(fields, "relationship", &result.Relationship) ||
		!decode(fields, "latency_ms", &result.LatencyMS) {
		return result, false
	}
	var checks map[string]map[string]json.RawMessage
	if json.Unmarshal(fields["checks"], &checks) != nil {
		return result, false
	}
	result.Checks = make(map[string]CheckResult, len(checks))
	for dimension, check := range checks {
		if check == nil {
			return result, false
		}
		for _, key := range []string{"message", "code", "steps"} {
			if value, ok := check[key]; ok && bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
				return result, false
			}
		}
		if value, ok := check["code"]; ok {
			var code string
			if json.Unmarshal(value, &code) != nil || !diagnosticCodePattern.MatchString(code) {
				return result, false
			}
		}
		var entry CheckResult
		if json.Unmarshal(check["status"], &entry.Status) != nil ||
			!decode(check, "message", &entry.Message) ||
			!decode(check, "code", &entry.Code) ||
			!decode(check, "steps", &entry.Steps) {
			return result, false
		}
		result.Checks[dimension] = entry
	}
	var children []json.RawMessage
	if json.Unmarshal(fields["dependencies"], &children) != nil {
		return result, false
	}
	result.Dependencies = make([]Result, 0, len(children))
	for _, child := range children {
		parsed, ok := conformanceResult(child, depth+1)
		if !ok {
			return result, false
		}
		result.Dependencies = append(result.Dependencies, parsed)
	}
	if ValidateResult(result) != nil {
		return result, false
	}
	return result, true
}

func conformanceStatus(result Result) bool {
	critical := make([]bool, len(result.Dependencies))
	for i, child := range result.Dependencies {
		if !conformanceStatus(child) {
			return false
		}
		critical[i] = child.Critical == nil || *child.Critical
	}
	// A pre-check configuration failure may legitimately have no check entries.
	if len(result.Checks) == 0 && result.Status == Misconfigured {
		return result.Status == Aggregate(map[string]CheckResult{"configuration": {Status: Misconfigured}}, result.Dependencies, critical)
	}
	return result.Status == Aggregate(result.Checks, result.Dependencies, critical)
}
