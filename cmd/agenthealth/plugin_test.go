package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	processadapter "github.com/TheAgentHealth/agenthealth/adapters/process"
	"github.com/TheAgentHealth/agenthealth/core"
)

func TestExecutableAdapterRegistration(t *testing.T) {
	path := filepath.Join(t.TempDir(), "adapter.json")
	executable, _ := os.Executable()
	manifest := processadapter.Manifest{Protocol: "v1", Command: executable, SHA256: "0000000000000000000000000000000000000000000000000000000000000000", Metadata: core.Metadata{Name: "test", Version: "1", CompatibilityVersion: "v1", TargetTypes: []string{"custom"}, Dimensions: []string{"configuration", "reachability"}}}
	raw, _ := json.Marshal(manifest)
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		args []string
		code int
	}{
		{[]string{"ping", "custom", "https://example.invalid", "--adapter", path, "--format", "json"}, 5},
		{[]string{"ping", "custom", "https://example.invalid", "--adapter", path, "--adapter", path}, 6},
		{[]string{"ping", "custom", "https://example.invalid", "--adapter", path + "missing"}, 6},
		{[]string{"ping", "custom", "https://example.invalid", "--adapter"}, 6},
		{[]string{"version", "--adapter", path}, 6},
	} {
		var out, diag bytes.Buffer
		if code := run(context.Background(), tc.args, &out, &diag); code != tc.code {
			t.Fatalf("args %v code=%d diag=%s", tc.args, code, diag.String())
		}
		if bytes.Contains(out.Bytes(), []byte(executable)) || bytes.Contains(diag.Bytes(), []byte(executable)) {
			t.Fatal("executable path leaked")
		}
	}
	manifest.Metadata.TargetTypes = []string{"http"}
	raw, _ = json.Marshal(manifest)
	os.WriteFile(path, raw, 0600)
	var out, diag bytes.Buffer
	if code := run(context.Background(), []string{"ping", "http", "https://example.invalid", "--adapter", path}, &out, &diag); code != 6 {
		t.Fatalf("built-in replacement accepted: %d", code)
	}
}
