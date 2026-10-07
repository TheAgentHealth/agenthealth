# AgentHealth CLI

Build with Go 1.23 or newer from the repository root:

```bash
go build -o agenthealth ./cmd/agenthealth
./agenthealth --help
./agenthealth ping http https://example.com
./agenthealth check examples/core-check/agenthealth.yaml --format json
./agenthealth doctor examples/core-check/agenthealth.yaml
./agenthealth version
```

`ping <type> <endpoint>` creates one target named `ping` with the engine's default passive checks. HTTP and API are currently supported by the reference HTTP adapter. Other specification target types, including MCP and A2A, produce a `MISCONFIGURED` result until their adapters are implemented. An invalid target type is an invocation error.

`check <configuration.yaml>` loads strict YAML and runs the configured targets and dependencies. `doctor <configuration.yaml>` performs the same checks and adds status-based troubleshooting advice to terminal output. It does not enable additional functional checks. Failed checks and dependency evidence appear in all formats. Doctor advice is general guidance rather than a claim that a particular root cause has been identified.

Use `--format terminal`, `--format json`, or `--format yaml` before or after positional arguments. Terminal is the default. JSON and YAML use identical specification fields, with a single-result document for one target or an ordered `results` batch for multiple targets. Doctor's machine output is the same health document as check. `--` ends option parsing for paths beginning with a dash.

`version` and `--version` print the build version (`dev` by default). Release builds can set it with:

```bash
go build -ldflags '-X main.version=0.1.0' -o agenthealth ./cmd/agenthealth
```

## Exit behavior

Health commands return the most severe top-level status after dependency aggregation: healthy `0`, degraded `1`, unhealthy `2`, unreachable `3`, misconfigured `4`, unknown `5`. Invocation, unreadable/invalid configuration, and execution/output failures return `6` and write diagnostics to stderr. A target's invalid endpoint or missing adapter produces a health result and returns `4`. See the [exit-code specification](../spec/exit-codes.md).

Build a binary when testing exit codes: `go run` reports its own process exit code when the program exits unsuccessfully.

## Safety and scope

The CLI inherits the [engine safety policies](../core/README.md#policies-and-safety): verified TLS, no redirect following, secret redaction, passive defaults, bounded timeouts and retries. Functional checks require explicit configuration opt-in. Interrupt and termination signals cancel the run through the engine context.

The [HTTP adapter](http-adapter.md) uses HEAD for passive connectivity, authentication, and protocol checks. Default protocol checks require a 2xx response; response headers and accepted statuses can be configured. Body matching requires explicit `functional` opt-in. Protocol-specific MCP/A2A adapters remain later phases.
