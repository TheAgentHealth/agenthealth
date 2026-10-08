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

`ping <type> <endpoint>` creates one target named `ping` with the engine's default passive checks. Agent/multi-agent, gateway/router HTTP signals, HTTP/API, MCP Streamable HTTP/stdio, and A2A 1.0 JSON-RPC with explicit 0.3.0 compatibility are supported. Other specification target types produce a `MISCONFIGURED` result until their adapters are implemented. An invalid target type is an invocation error.

`check <configuration.yaml>` loads strict YAML and runs the configured targets and dependencies. `doctor <type> <endpoint>` runs the same passive defaults as ping. `doctor <configuration.yaml>` performs the same checks and adds status-based troubleshooting advice to terminal output. It does not enable additional functional checks. Failed checks and dependency evidence appear in all formats. Doctor advice is general guidance rather than a claim that a particular root cause has been identified.

Use `--format terminal`, `--format json`, or `--format yaml` before or after positional arguments. Terminal is the default. JSON and YAML use identical specification fields, with a single-result document for one target or an ordered `results` batch for multiple targets. Doctor's machine output is the same health document as check. `--` ends option parsing for paths beginning with a dash.

`version` and `--version` print the build version (`dev` by default). Release builds can set it with:

```bash
go build -ldflags '-X main.version=v0.5.0' -o agenthealth ./cmd/agenthealth
```

## Exit behavior

Health commands return the most severe top-level status after dependency aggregation: healthy `0`, degraded `1`, unhealthy `2`, unreachable `3`, misconfigured `4`, unknown `5`. Invocation, unreadable/invalid configuration, and execution/output failures return `6` and write diagnostics to stderr. A target's invalid endpoint or missing adapter produces a health result and returns `4`. See the [exit-code specification](../spec/exit-codes.md).

Build a binary when testing exit codes: `go run` reports its own process exit code when the program exits unsuccessfully.

## Configuration trust boundary

Configuration is trusted executable input: it can launch MCP stdio programs, pass referenced credentials, contact private-network endpoints, acquire OAuth tokens, and write token files. Review these settings before execution, including configurations supplied through external pull requests. Active-check opt-in does not sandbox commands or network access. See [trusted configuration](../SECURITY.md#trusted-configuration).

## Safety and scope

The CLI inherits the [engine safety policies](../core/README.md#policies-and-safety): verified TLS, no redirect following, secret redaction, passive defaults, bounded timeouts and retries. Functional checks require explicit configuration opt-in. Interrupt and termination signals cancel the run through the engine context.

The [HTTP adapter](http-adapter.md) uses HEAD for passive connectivity, authentication, and protocol checks. Default protocol checks require a 2xx response; response headers and accepted statuses can be configured. Body matching requires explicit `functional` opt-in. The [MCP adapter](mcp-adapter.md) supports legacy initialization, modern stateless discovery, configured stdio, OAuth acquisition, and inventory checks. The [A2A adapter](a2a-adapter.md) discovers agent cards, validates protocol/authentication with read-only task lookups, checks expected skills, and supports opt-in safe text interactions.

## A2A checks

AgentHealth v0.4.0 supports A2A 1.0 JSON-RPC with explicit 0.3.0 compatibility. Point `ping` or `doctor` at
the v1 peer’s origin; discovery uses `/.well-known/agent-card.json`. A configured target
can supply an absolute card URL on the same origin, bearer credentials, required
skill IDs/capabilities, latency thresholds, and an explicitly safe interaction.

```bash
agenthealth ping a2a http://localhost:9000
agenthealth doctor a2a http://localhost:9000
agenthealth check examples/a2a-check/agenthealth.yaml --format json
agenthealth check examples/a2a-check/bearer.yaml --format yaml
agenthealth check examples/a2a-check/functional.yaml --format json
```

Examples require your running peer and matching metadata; the bearer example
requires `AGENT_TOKEN`. Default checks only discover metadata and perform
read-only task lookups. `functional` requires configuration opt-in and
`safe: true`; it sends one text interaction without retries or polling.
`doctor` uses the same check policy and adds terminal advice. See the
[A2A guide](a2a-adapter.md) and [example setup](../examples/a2a-check/README.md).

## MCP OAuth login

`login <configuration.yaml> <target-name>` authenticates a configured top-level
MCP target using authorization code + PKCE. It requires `mcp.oauth.grant:
authorization_code`, a pre-registered client ID, and a private token file.
The command prints an authorization URL for the user to open, waits for a
loopback callback for up to five minutes, then saves credentials without
printing tokens. Login uses terminal output and returns 0 on success or 6 on
failure. Health commands remain noninteractive and use acquired or refreshed
credentials. See [OAuth setup](mcp-adapter.md#oauth-acquisition-and-login).

For legacy A2A peers, use a configuration containing
`a2a.protocol_version: '0.3.0'`; unconfigured `ping` and `doctor` select v1.

## Agent checks

`ping agent <health-resource>` and `ping multi-agent <health-resource>` run passive
checks. `check` and `doctor` also accept required capabilities and explicit safe
task/path probes. See the [agent adapter](agent-adapter.md) and
[example](../examples/agent-check/agenthealth.yaml).

Phase 8 Agentgateway integration is implemented in source. See [the gateway guide](../docs/agentgateway.md). Gateway checks use explicitly configured read-only HTTP signals and separate backend/path dependencies.

Phase 9 Agent Router HTTP integration is released in v0.7.0.
See [the router guide](agent-router.md). Named router signals, direct backends and configured paths retain separate evidence.

## Phase 10 dependency graph

Use v0.8.0 or newer for explicit graph fields; v0.7.0 binaries do not include them.

```bash
agenthealth check examples/graph-check/agenthealth.yaml --format json
agenthealth check examples/graph-check/agenthealth.yaml --format yaml
agenthealth doctor examples/graph-check/agenthealth.yaml
```

Terminal output labels IDs, relationships and optional edges. JSON/YAML retain
`target.id`, incoming-edge `relationship` and `critical` alongside the existing
recursive result. Graph validation failures return tool-failure exit code 6
before adapter calls; health outcomes retain exit codes 0–5.
See the [graph guide](dependency-graph.md) and [contract](rfcs/phase-10-dependency-graph.md).
Existing nested configurations remain valid.

## Experimental AHP serving

`agenthealth serve <configuration.yaml>` exposes snapshot health over HTTP.
See the [AHP guide](ahp.md) for authorization, freshness and deployment.
