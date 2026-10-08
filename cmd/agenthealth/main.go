package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	a2aadapter "github.com/TheAgentHealth/agenthealth/adapters/a2a"
	agentadapter "github.com/TheAgentHealth/agenthealth/adapters/agent"
	gatewayadapter "github.com/TheAgentHealth/agenthealth/adapters/gateway"
	httpadapter "github.com/TheAgentHealth/agenthealth/adapters/http"
	mcpadapter "github.com/TheAgentHealth/agenthealth/adapters/mcp"
	routeradapter "github.com/TheAgentHealth/agenthealth/adapters/router"
	"github.com/TheAgentHealth/agenthealth/core"
)

// Set at release build time with -ldflags '-X main.version=<version>'.
var version = "dev"

const usage = `Usage:
  agenthealth ping <type> <endpoint> [--format terminal|json|yaml]
  agenthealth check <configuration.yaml> [--format terminal|json|yaml]
  agenthealth doctor <type> <endpoint> [--format terminal|json|yaml]
  agenthealth doctor <configuration.yaml> [--format terminal|json|yaml]
  agenthealth login <configuration.yaml> <target-name>
  agenthealth version

Available target types: agent, multi-agent, http, api, mcp, a2a, gateway, router.
A2A defaults to 1.0 JSON-RPC with explicit 0.3.0 compatibility.
Agent task/path probes require the application to implement the safe probe handler.
Other target types require future adapters.
Checks are passive by default; functional checks require opt-in in configuration.
`

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	os.Exit(run(ctx, os.Args[1:], os.Stdout, os.Stderr))
}

func run(ctx context.Context, args []string, out, diagnostics io.Writer) int {
	fail := func(message string) int {
		fmt.Fprintln(diagnostics, message)
		return 6
	}
	format := "terminal"
	var positional []string
	options := true
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if options && arg == "--" {
			options = false
			continue
		}
		if options && (arg == "--help" || arg == "-h") {
			if _, err := io.WriteString(out, usage); err != nil {
				return fail("cannot write help")
			}
			return 0
		}
		if options && arg == "--version" {
			arg = "version"
		}
		if options && (arg == "--format" || strings.HasPrefix(arg, "--format=")) {
			if arg == "--format" {
				i++
				if i == len(args) {
					return fail("--format requires terminal, json, or yaml")
				}
				format = args[i]
			} else {
				format = strings.TrimPrefix(arg, "--format=")
			}
			continue
		}
		if options && strings.HasPrefix(arg, "-") {
			return fail("unknown option; use --help")
		}
		positional = append(positional, arg)
	}
	if format != "terminal" && format != "json" && format != "yaml" {
		return fail("unsupported output format; use terminal, json, or yaml")
	}
	if len(positional) == 0 {
		return fail(usage)
	}
	command := positional[0]
	if command == "version" {
		if len(positional) != 1 {
			return fail("version takes no arguments")
		}
		if _, err := fmt.Fprintf(out, "agenthealth %s\n", version); err != nil {
			return fail("cannot write version")
		}
		return 0
	}
	var config core.Config
	switch command {
	case "ping", "doctor":
		if command == "doctor" && len(positional) == 2 {
			var err error
			config, err = core.LoadConfigFile(positional[1])
			if err != nil {
				return fail(err.Error())
			}
			break
		}
		if len(positional) != 3 {
			return fail("usage: agenthealth " + command + " <type> <endpoint>")
		}
		config = core.Config{Version: "v1", Targets: []core.Target{{Name: command, Type: positional[1], Endpoint: positional[2]}}}
		if err := config.Validate(); err != nil {
			return fail("invalid target; use a specification target type and a nonempty endpoint")
		}
	case "login":
		if len(positional) != 3 || format != "terminal" {
			return fail("usage: agenthealth login <configuration.yaml> <target-name>")
		}
		loaded, err := core.LoadConfigFile(positional[1])
		if err != nil {
			return fail(err.Error())
		}
		for _, target := range loaded.Targets {
			if target.Name == positional[2] {
				loginCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
				defer cancel()
				if err := mcpadapter.Login(loginCtx, target, out); err != nil {
					return fail("OAuth login failed; verify issuer, client registration, callback and token file configuration")
				}
				return 0
			}
		}
		return fail("login target not found in configuration")
	case "check":
		if len(positional) != 2 {
			return fail("usage: agenthealth " + command + " <configuration.yaml>")
		}
		var err error
		config, err = core.LoadConfigFile(positional[1])
		if err != nil {
			return fail(err.Error())
		}
	default:
		return fail("unknown command; use --help")
	}
	registry := core.NewRegistry()
	if err := registry.Register(routeradapter.Adapter{}); err != nil {
		return fail("cannot register router adapter")
	}
	if err := registry.Register(gatewayadapter.Adapter{}); err != nil {
		return fail("cannot register gateway adapter")
	}
	if err := registry.Register(agentadapter.Adapter{}); err != nil {
		return fail("cannot register agent adapter")
	}
	if err := registry.Register(httpadapter.Adapter{}); err != nil {
		return fail("cannot register HTTP adapter")
	}
	if err := registry.Register(mcpadapter.Adapter{}); err != nil {
		return fail("cannot register MCP adapter")
	}
	if err := registry.Register(a2aadapter.Adapter{}); err != nil {
		return fail("cannot register A2A adapter")
	}
	results, err := core.NewEngine(registry).Run(ctx, config)
	if err != nil {
		return fail("cannot execute health checks")
	}
	switch format {
	case "json":
		err = core.WriteJSON(out, results)
	case "yaml":
		err = core.WriteYAML(out, results)
	default:
		err = core.WriteHuman(out, results)
		if err == nil && command == "doctor" {
			err = writeAdvice(out, results)
		}
	}
	if err != nil {
		return fail("cannot write health results")
	}
	states := make([]core.Status, len(results))
	for i, result := range results {
		states[i] = result.Status
	}
	return exitCode(core.Worst(states...))
}

func exitCode(status core.Status) int {
	switch status {
	case core.Healthy:
		return 0
	case core.Degraded:
		return 1
	case core.Unhealthy:
		return 2
	case core.Unreachable:
		return 3
	case core.Misconfigured:
		return 4
	default:
		return 5
	}
}

func writeAdvice(w io.Writer, results []core.Result) error {
	advice := map[core.Status]string{
		core.Healthy:       "Checks passed. Review the check list to confirm it covers your readiness requirements.",
		core.Degraded:      "Inspect latency thresholds and optional dependencies marked as impaired.",
		core.Unhealthy:     "Inspect failed checks and critical dependencies before using this target.",
		core.Unreachable:   "Verify the endpoint, service availability, network access, and timeout policy.",
		core.Misconfigured: "Verify target configuration, adapter availability, and referenced environment credentials.",
		core.Unknown:       "Review requested checks and adapter support; the available evidence is inconclusive.",
	}
	for _, result := range results {
		if _, err := fmt.Fprintf(w, "Doctor: %s\n", advice[result.Status]); err != nil {
			return err
		}
		if err := writeAdvice(w, result.Dependencies); err != nil {
			return err
		}
	}
	return nil
}
