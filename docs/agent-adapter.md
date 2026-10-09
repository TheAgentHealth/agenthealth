# Agent health adapter

The `agent` target checks an individual agent; `multi-agent` checks an exposed
cooperating-system endpoint using the same interface. Direct means the user's
first agent; composite means a downstream agent. Roles are expressed by the
configuration's named result tree rather than new target types or result fields.

## Passive interface

Point `endpoint` at a runtime's health resource. It must support HEAD for
reachability and GET for this JSON document (`Content-Type: application/json`):

```json
{"version":"v1","name":"assistant","live":true,"ready":true,"capabilities":["answer"],"dependencies":["peer"]}
```

All six fields are required. Names must be nonempty; capability and dependency
names must be unique. Additional fields are ignored for forward compatibility.
The health resource GET must be read-only. The document is fetched once per target execution and shared by authentication,
protocol and capability checks. Authentication rejects HTTP 401/403 as
MISCONFIGURED. Other unsuccessful HTTP statuses are UNHEALTHY. Protocol checks
validate metadata, liveness and readiness: a valid document with either flag
false is UNHEALTHY. Malformed metadata is UNHEALTHY; oversized or incomplete
responses are UNKNOWN. Reachability alone proves only that an HTTP response
arrived, regardless of status. Liveness and readiness are runtime declarations;
passive evidence does not prove successful task execution.

Capability checks compare `agent.required_capabilities` with the document.
Missing required capabilities are UNHEALTHY. Discovered dependency names must
match explicitly configured immediate dependency names; an unconfigured name
is UNKNOWN. Discovery never follows addresses or creates targets. Configured
dependencies execute independently, including after parent failure, whether or
not the runtime advertises them. Each has its own credentials and policies.
Critical failures contribute UNHEALTHY; optional failures contribute DEGRADED.

## Safe task and communication probes

`checks: [functional]` and `agent.functional: {safe: true, text: ...}` opt in
to one POST to the configured health resource. The resource must implement a
bounded, non-destructive probe handler. It receives JSON containing `safe`,
`text`, and optionally `downstream`. A successful response is:

```json
{"completed":true,"success":true}
```

For a path probe, `downstream` identifies the peer the first runtime must
contact using its own configured communication mechanism, directly or via A2A.
The handler must actually perform the safe peer exchange before returning
`{"completed":true,"success":true,"downstream":"peer"}`. A different peer or
`success: false` is UNHEALTHY. Missing success/completion or incomplete work is
UNKNOWN; there is no polling or retry. POST response bodies and task text are
never included in results. A remote handler's claim is trusted evidence, not
independent tracing of its implementation; use a runtime integration that
implements this contract. The CLI does not implement arbitrary framework task
APIs or claim that checking an A2A endpoint proves a first-to-peer path.

Use separate named dependencies for peer endpoint health and the first-to-peer
probe to distinguish failures. A2A peers retain `type: a2a` and their existing
card/skill/task configuration. A path probe is a separate `agent` target pointed
at the first runtime, with `agent.functional.downstream` selecting the peer.
See the [example](../examples/agent-check/agenthealth.yaml).

## Bounds and compatibility

Responses are limited to 64 KiB; task text to 64 KiB of UTF-8 bytes and downstream IDs to 1024
Unicode characters. Engine deadlines, cancellation, TLS verification, redirect refusal,
bearer environment references, credential redaction, and concurrency limits
apply. `safe: true` declares operator authorization and handler safety; the CLI
cannot sandbox a remote runtime. The interface is a Phase 7 adapter contract,
not the planned AHP serving protocol. Existing v1 configurations/results and
exit codes remain compatible; `agent` options are an additive extension.

```bash
agenthealth ping agent http://localhost:8080/health
agenthealth check examples/agent-check/agenthealth.yaml --format json
agenthealth doctor examples/agent-check/agenthealth.yaml --format yaml
```

An empty `downstream` selector is equivalent to omission and probes the agent
itself. Whitespace-only selectors are invalid. JSON Schema string lengths
count Unicode characters; Go additionally enforces the task text byte limit.

## Dependency graph integration

[Phase 10](dependency-graph.md) adds explicit node IDs and references while preserving this adapter’s checks and safety contract. Advertised dependency names match only explicitly configured inline nodes or resolved references. References never authorize remote discovery. Keep the peer node and the safe first-runtime communication probe as distinct IDs. Critical/optional policy belongs to each edge. The [graph example](../examples/graph-check/README.md) shows the configuration pattern. This extension is included in v0.8.0.

## Kubernetes deployment

The Kubernetes fixture distinguishes direct and composite agent health from supporting gateway/router/route evidence. It does not supply a first-runtime safe downstream handler. Existing functional configurations remain opt-in and may repeat in serving mode.
See [Phase 14 deployment guidance](kubernetes.md) for ConfigMaps, Secrets,
probe policy and validation limits. No adapter or configuration migration is required.
