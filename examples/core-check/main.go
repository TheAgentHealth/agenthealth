// This example exercises Phase 2 without implementing the Phase 3 CLI.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	httpadapter "github.com/TheAgentHealth/agenthealth/adapters/http"
	"github.com/TheAgentHealth/agenthealth/core"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: go run ./examples/core-check <configuration.yaml>")
		os.Exit(6)
	}
	config, err := core.LoadConfigFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(4)
	}
	registry := core.NewRegistry()
	if err := registry.Register(httpadapter.Adapter{}); err != nil {
		fmt.Fprintln(os.Stderr, "cannot register adapter")
		os.Exit(6)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	results, err := core.NewEngine(registry).Run(ctx, config)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(6)
	}
	if err := core.WriteJSON(os.Stdout, results); err != nil {
		fmt.Fprintln(os.Stderr, "cannot write results")
		os.Exit(6)
	}
}
