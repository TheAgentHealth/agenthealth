package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConformanceCLI(t *testing.T) {
	path := filepath.Join(t.TempDir(), "evidence.json")
	for _, tc := range []struct {
		evidence string
		code     int
	}{
		{`{"spec_version":"v1","target":{"name":"secret-input","type":"http"},"status":"UNREACHABLE","checks":{"reachability":{"status":"UNREACHABLE"}},"dependencies":[]}`, 0},
		{`{"secret-input":`, 6},
	} {
		if err := os.WriteFile(path, []byte(tc.evidence), 0600); err != nil {
			t.Fatal(err)
		}
		var out, diagnostics bytes.Buffer
		if code := run(context.Background(), []string{"conformance", "test", path}, &out, &diagnostics); code != tc.code {
			t.Fatalf("code=%d output=%s stderr=%s", code, &out, &diagnostics)
		}
		if strings.Contains(out.String()+diagnostics.String(), "secret-input") {
			t.Fatal("input leaked")
		}
	}
	for _, args := range [][]string{{"conformance"}, {"conformance", "test", "missing-secret-path"}, {"conformance", "test", t.TempDir()}, {"conformance", "test", path, "--adapter", "untrusted"}} {
		var out, diagnostics bytes.Buffer
		if run(context.Background(), args, &out, &diagnostics) != 6 {
			t.Fatal("accepted bad invocation")
		}
		if strings.Contains(diagnostics.String(), "missing-secret-path") {
			t.Fatal("path leaked")
		}
	}
	var diagnostics bytes.Buffer
	if runConformance([]string{"test", path}, brokenWriter{}, &diagnostics) != 6 {
		t.Fatal("output failure did not fail the command")
	}
	if diagnostics.String() != "cannot write conformance report\n" {
		t.Fatal("output failure was not diagnosed")
	}
}
