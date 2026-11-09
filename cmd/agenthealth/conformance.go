package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/TheAgentHealth/agenthealth/core"
)

func runConformance(args []string, out, diagnostics io.Writer) int {
	if len(args) != 2 || args[0] != "test" {
		fmt.Fprintln(diagnostics, "usage: agenthealth conformance test <result.json>")
		return 6
	}
	// Reject devices and pipes before opening, since opening a pipe may block.
	info, err := os.Stat(args[1])
	if err != nil || !info.Mode().IsRegular() {
		fmt.Fprintln(diagnostics, "conformance evidence must be a regular file")
		return 6
	}
	file, err := os.Open(args[1])
	if err != nil {
		fmt.Fprintln(diagnostics, "cannot open conformance evidence")
		return 6
	}
	defer file.Close()
	info, err = file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		fmt.Fprintln(diagnostics, "conformance evidence must be a regular file")
		return 6
	}
	report, err := core.TestConformance(file)
	if err != nil {
		fmt.Fprintln(diagnostics, "cannot read conformance evidence")
		return 6
	}
	if json.NewEncoder(out).Encode(report) != nil {
		fmt.Fprintln(diagnostics, "cannot write conformance report")
		return 6
	}
	if !report.Passed {
		return 6
	}
	return 0
}
