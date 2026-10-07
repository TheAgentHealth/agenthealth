# Go core engine

The reference engine implements [Phase 2](../ROADMAP.md#phase-2--core-engine). The [Phase 3 CLI](../docs/cli.md) uses this engine; the [MCP adapter](../docs/mcp-adapter.md) implements Phase 5 and the [A2A adapter](../docs/a2a-adapter.md) implements Phase 6 for A2A 0.3.0 JSON-RPC. Other adapters remain later roadmap work.

## Run it

The [HTTP/API adapter](../docs/http-adapter.md) provides passive HEAD-based connectivity, authentication, status and response-header checks, transport-stage diagnostics, and explicit opt-in GET body matching. Default status expectations are 2xx. TLS is verified and redirects are not followed.

Start a local service with a `/health` endpoint and run:

```bash
go run ./examples/core-check examples/core-check/agenthealth.yaml
```

The example emits JSON; it is an API demonstration, not the Phase 3 CLI or an automation exit-code implementation.

## API

`LoadConfigFile` loads strict YAML; `Config.Validate` validates configurations built in Go. `NewRegistry().Register(adapter)` checks metadata compatibility and duplicate target types. `NewEngine(registry).Run(ctx, config)` returns results in declaration order. Unknown target adapters yield `MISCONFIGURED` when checks are requested. Explicit `checks: []` produces `UNKNOWN` with no adapter calls; dependencies still run.

Configuration and Reachability gate downstream work. Authentication gates Capability and Functional, while Protocol and Latency remain independent of authentication success. Passing implicit gates are hidden; failed gates appear as evidence and blocked checks are omitted. Latency measures the final successful Reachability attempt, excluding local adapter-slot queue time; threshold violations yield `DEGRADED`.

Each nested dependency runs independently, even after the parent fails. Critical defaults to true; optional failures contribute `DEGRADED`. Dependencies do not inherit checks, thresholds, credentials, or policies. Dependency summary checks are optional diagnostics, separate from aggregation. `Aggregate` also accepts already-evaluated dependency results for external consumers.

`WriteJSON` and `WriteYAML` emit a single result or ordered batch with the same wire fields. `WriteHuman` displays the same result tree and latency with deterministic check ordering and terminal control characters removed. Both validate result shapes and mask recognizable URLs and credential assignments. Output rejects dependency trees deeper than 64 levels, including cyclic caller-built result slices, before writing any bytes. Results initialize Checks and Dependencies to empty collections, not nil.

## Policies and safety

See [execution policies](../spec/configuration.md#execution-policies-phase-2-draft) for the draft fields, defaults, and bounds. Per-attempt deadlines default to five seconds, retries to zero, and the whole-run budget to 60 seconds (a shorter caller deadline takes precedence). Passive no-response connectivity failures may retry at most three additional times. Active checks and partial responses never retry. Retry delays and pool acquisition honor cancellation.

Credentials resolve from environment references once per run. MCP OAuth credentials are acquired noninteractively and cached within the target run; dynamically acquired access/refresh tokens participate in redaction. Configured stdio processes receive only selected runtime variables and explicit environment references. Missing/empty credentials fail Configuration before network requests. Raw adapter errors, messages, and panic values never reach results or engine logs: the engine uses canonical diagnostics and allowlisted transport stages and emits no internal logs. Resolved credentials are redacted recursively from identities and results; callers can safely log engine results through the formatters. `Redactor` is available for other structured logging paths. Arbitrary manually constructed results must supply known secret values to a Redactor before output; generic patterns cannot identify every possible secret.

Adapters declare active dimensions; none execute by default. Functional must be active and explicitly requested. Trusted adapters must keep probes non-destructive and avoid their own unsafe logging. The supplied HTTP client verifies TLS without an insecure option and stops redirects to prevent credential forwarding. Per-attempt context/client deadlines govern TLS handshakes and response headers without a separate fixed transport cap. Invalid HTTP header control bytes in credentials fail Configuration before networking.

Calls are limited to 16 in flight per engine. Panics and invalid adapter states normalize to `UNKNOWN`. `Request.MarkResponse` records partial responses even when the caller's deadline wins a race with adapter completion. Adapters must honor contexts: in-process Go code cannot be forcibly terminated, so an uncooperative call occupies a bounded pool slot until it returns. See the [adapter interface contract](../spec/adapter-spec.md#reference-go-interface-phase-2-draft).

## Validation

```bash
go test -race ./...
go vet ./...
python3 -m pytest tests -q
python3 scripts/check_markdown_links.py
python3 scripts/validate_doc_examples.py
```

Tests cover configuration fixtures, prerequisite gates, independent dependencies, timeout/partial-response classification, bounded retries and calls, active controls, panic recovery, secret suppression, TLS verification, redirect safety, HTTP authentication, and JSON envelopes. A Python integration test validates real example output against the result schema. CI runs all these checks.
