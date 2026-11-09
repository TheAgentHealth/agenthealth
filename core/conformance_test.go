package core

import (
	"errors"
	"strings"
	"testing"
)

const validConformance = `{"spec_version":"v1","target":{"name":"sample","type":"agent"},"status":"HEALTHY","checks":{"reachability":{"status":"HEALTHY"}},"dependencies":[]}`

func TestConformanceDocuments(t *testing.T) {
	for _, tc := range []struct { name, input string; passed bool }{
		{"valid", validConformance, true},
		{"extension", strings.Replace(validConformance, `"spec_version"`, `"extension":{"x":1},"spec_version"`, 1), true},
		{"missing", strings.Replace(validConformance, `"dependencies":[]`, `"other":[]`, 1), false},
		{"null_checks", strings.Replace(validConformance, `{"reachability":{"status":"HEALTHY"}}`, `null`, 1), false},
		{"null_message", strings.Replace(validConformance, `"reachability":{"status"`, `"reachability":{"message":null,"status"`, 1), false},
		{"null_critical", strings.Replace(validConformance, `"spec_version"`, `"critical":null,"spec_version"`, 1), false},
		{"empty_relationship", strings.Replace(validConformance, `"spec_version"`, `"relationship":"","spec_version"`, 1), false},
		{"empty_code", strings.Replace(validConformance, `"reachability":{"status"`, `"reachability":{"code":"","status"`, 1), false},
		{"empty_id", strings.Replace(validConformance, `"name":"sample"`, `"id":"","name":"sample"`, 1), false},
		{"duplicate", strings.Replace(validConformance, `"spec_version"`, `"status":"UNKNOWN","spec_version"`, 1), false},
		{"escaped_duplicate", strings.Replace(validConformance, `"spec_version"`, `"\u0073tatus":"UNKNOWN","spec_version"`, 1), false},
		{"trailing", validConformance+` {}`, false},
		{"version", strings.Replace(validConformance, `"v1"`, `"v2"`, 1), false},
		{"status", strings.Replace(validConformance, `"status":"HEALTHY"`, `"status":"UNHEALTHY"`, 1), false},
		{"empty_batch", `{"spec_version":"v1","results":[]}`, false},
		{"ambiguous", strings.Replace(validConformance, `"spec_version"`, `"results":[],"spec_version"`, 1), false},
		{"batch", `{"spec_version":"v1","results":[`+validConformance+`,`+validConformance+`]}`, true},
		{"unknown", strings.ReplaceAll(strings.Replace(validConformance, `{"reachability":{"status":"HEALTHY"}}`, `{}`, 1), `"HEALTHY"`, `"UNKNOWN"`), true},
		{"oversized", strings.Repeat(" ", ConformanceMaxBytes+1), false},
	} {
		t.Run(tc.name, func(t *testing.T) { report, err := TestConformance(strings.NewReader(tc.input)); if err != nil || report.Passed != tc.passed { t.Fatalf("report=%+v error=%v", report, err) } })
	}
}

func TestConformanceDependencies(t *testing.T) {
	child := strings.ReplaceAll(validConformance, "HEALTHY", "UNREACHABLE")
	parent := strings.Replace(validConformance, `"dependencies":[]`, `"dependencies":[`+child+`]`, 1)
	for _, tc := range []struct{ input string; passed bool }{
		{parent, false},
		{strings.Replace(parent, `"status":"HEALTHY"`, `"status":"UNHEALTHY"`, 1), true},
		{strings.Replace(parent, `"status":"HEALTHY"`, `"status":"UNREACHABLE"`, 1), false},
		{strings.Replace(strings.Replace(parent, `"status":"HEALTHY"`, `"status":"DEGRADED"`, 1), `"dependencies":[{`, `"dependencies":[{"critical":false,`, 1), true},
	} { report, err := TestConformance(strings.NewReader(tc.input)); if err != nil || report.Passed != tc.passed { t.Fatalf("report=%+v error=%v", report, err) } }
	deep := validConformance
	for i:=0; i<66; i++ { deep = strings.Replace(validConformance, `"dependencies":[]`, `"dependencies":[`+deep+`]`, 1) }
	if report, _ := TestConformance(strings.NewReader(deep)); report.Passed { t.Fatal("accepted excessive depth") }
}

type conformanceBrokenReader struct{}
func (conformanceBrokenReader) Read([]byte) (int,error) { return 0, errors.New("secret") }
func TestConformanceReadFailure(t *testing.T) {
	_, err := TestConformance(conformanceBrokenReader{})
	if err == nil || strings.Contains(err.Error(), "secret") { t.Fatal("reader error not sanitized") }
}
