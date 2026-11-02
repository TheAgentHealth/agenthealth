package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"time"
)

// ExportObservation is a deliberately reduced view: source names, IDs, diagnostic
// messages, step names and endpoints never cross the export boundary.
// Aliases are local to one snapshot, not durable identities or anonymization.
type ExportObservation struct {
	Kind         string   `json:"kind"`
	Node         string   `json:"node"`
	Parent       string   `json:"parent,omitempty"`
	Edge         string   `json:"edge,omitempty"`
	Type         string   `json:"type"`
	Status       Status   `json:"status"`
	Dimension    string   `json:"dimension,omitempty"`
	LatencyMS    *float64 `json:"latency_ms,omitempty"`
	Relationship string   `json:"relationship,omitempty"`
	Critical     *bool    `json:"critical,omitempty"`
}

func observations(results []Result) ([]ExportObservation, error) {
	if err := validateResults(results); err != nil {
		return nil, err
	}
	records := []ExportObservation{}
	ids := map[string]string{}
	next := 0
	edges := 0
	var walk func(Result, string)
	walk = func(r Result, parent string) {
		alias, seen := ids[r.Target.ID]
		if r.Target.ID == "" {
			seen = false
		}
		if !seen {
			next++
			alias = fmt.Sprintf("node-%d", next)
			if r.Target.ID != "" {
				ids[r.Target.ID] = alias
			}
		}
		if parent != "" {
			edges++
			critical := true
			if r.Critical != nil {
				critical = *r.Critical
			}
			records = append(records, ExportObservation{Kind: "edge", Edge: fmt.Sprintf("edge-%d", edges), Node: alias, Parent: parent, Type: r.Target.Type, Status: r.Status, Relationship: r.Relationship, Critical: &critical})
		}
		if seen {
			return
		}
		records = append(records, ExportObservation{Kind: "node", Node: alias, Type: r.Target.Type, Status: r.Status, LatencyMS: r.LatencyMS})
		dimensions := make([]string, 0, len(r.Checks))
		for dimension := range r.Checks {
			dimensions = append(dimensions, dimension)
		}
		sort.Strings(dimensions)
		for _, dimension := range dimensions {
			records = append(records, ExportObservation{Kind: "check", Node: alias, Type: r.Target.Type, Status: r.Checks[dimension].Status, Dimension: dimension})
		}
		for _, dep := range r.Dependencies {
			walk(dep, alias)
		}
	}
	for _, r := range results {
		walk(r, "")
	}
	return records, nil
}

// WriteObservability emits a finite snapshot, never sends network requests and
// buffers validation/encoding before touching the destination. OTLP is an
// OTLP/HTTP JSON ExportLogsServiceRequest; webhook is a versioned JSON payload.
func WriteObservability(w io.Writer, results []Result, format string) error {
	if format != "prometheus" && format != "logs" && format != "otlp" && format != "webhook" {
		return errors.New("unsupported observability format")
	}
	records, err := observations(results)
	if err != nil {
		return err
	}
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	switch format {
	case "logs":
		for _, record := range records {
			if err := encoder.Encode(record); err != nil {
				return err
			}
		}
	case "webhook":
		err = encoder.Encode(struct {
			Version string              `json:"export_version"`
			Records []ExportObservation `json:"records"`
		}{"v1", records})
	case "otlp":
		// A single observed timestamp applies to the entire snapshot. No fabricated
		// trace/span context, host identity or endpoint attributes are attached.
		now := fmt.Sprint(time.Now().UnixNano())
		logs := make([]any, 0, len(records))
		for _, record := range records {
			body, marshalErr := json.Marshal(record)
			if marshalErr != nil {
				return marshalErr
			}
			logs = append(logs, map[string]any{"observedTimeUnixNano": now, "body": map[string]string{"stringValue": string(body)}})
		}
		err = encoder.Encode(map[string]any{"resourceLogs": []any{map[string]any{"scopeLogs": []any{map[string]any{"scope": map[string]string{"name": "agenthealth"}, "logRecords": logs}}}}})
	case "prometheus":
		buffer.WriteString("# HELP agenthealth_status Health status; exactly one state is 1.\n# TYPE agenthealth_status gauge\n# HELP agenthealth_latency_seconds Observed node check latency.\n# TYPE agenthealth_latency_seconds gauge\n")
		for _, record := range records {
			labels := fmt.Sprintf("kind=%q,node=%q,parent=%q,type=%q,dimension=%q,relationship=%q,critical=%q,edge=%q", record.Kind, record.Node, record.Parent, record.Type, record.Dimension, record.Relationship, criticalLabel(record.Critical), record.Edge)
			for _, status := range statuses {
				value := 0
				if status == record.Status {
					value = 1
				}
				fmt.Fprintf(&buffer, "agenthealth_status{%s,status=%q} %d\n", labels, status, value)
			}
		}
		for _, record := range records {
			if record.LatencyMS != nil {
				fmt.Fprintf(&buffer, "agenthealth_latency_seconds{node=%q,type=%q} %g\n", record.Node, record.Type, *record.LatencyMS/1000)
			}
		}
	}
	if err != nil {
		return err
	}
	_, err = w.Write(buffer.Bytes())
	return err
}

func criticalLabel(value *bool) string {
	if value == nil {
		return ""
	}
	if *value {
		return "true"
	}
	return "false"
}
