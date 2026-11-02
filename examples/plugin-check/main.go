// This executable demonstrates the v1 wire protocol without making network calls.
// It is a protocol fixture, not evidence of a real remote service's health.
package main

import (
	"encoding/json"
	"github.com/TheAgentHealth/agenthealth/adapters/process"
	"github.com/TheAgentHealth/agenthealth/core"
	"os"
)

func main() {
	var request process.WireRequest
	decoder := json.NewDecoder(os.Stdin)
	decoder.DisallowUnknownFields()
	if decoder.Decode(&request) != nil || request.Protocol != "v1" {
		os.Exit(1)
	}
	status := core.Healthy
	if request.Dimension != "configuration" {
		status = core.Unknown
	}
	if json.NewEncoder(os.Stdout).Encode(process.Event{Kind: "result", Status: status}) != nil {
		os.Exit(1)
	}
}
