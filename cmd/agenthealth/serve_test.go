package main

import (
	"bytes"
	"context"
	"testing"
)

func TestServeInvocation(t *testing.T) {
	for _, args := range [][]string{{"serve"}, {"serve", "missing.yaml"}, {"serve", "missing.yaml", "--unknown"}, {"serve", "missing.yaml", "--format", "json"}} {
		var out, diagnostics bytes.Buffer
		if code := run(context.Background(), args, &out, &diagnostics); code != 6 || diagnostics.Len() == 0 {
			t.Fatalf("%v: %d %s", args, code, diagnostics.String())
		}
	}
	var out, diagnostics bytes.Buffer
	if code := run(context.Background(), []string{"serve", "--help"}, &out, &diagnostics); code != 0 || out.Len() == 0 {
		t.Fatal(code)
	}
}
