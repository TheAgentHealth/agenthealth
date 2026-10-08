# A2A health adapter

The adapter supports [A2A 1.0 JSON-RPC](https://a2a-protocol.org/v1.0.1/specification/) by default and explicitly selected [A2A 0.3.0 compatibility](https://a2a-protocol.org/v0.3.0/specification/). It extends Phase 6. gRPC, REST, streaming, push notifications, extended-card retrieval, OAuth acquisition, and remote card registries are not implemented. Advertised streaming and push capabilities can be inspected without invoking them.

```bash
agenthealth ping a2a https://agent.example.com
agenthealth doctor a2a https://agent.example.com
agenthealth check examples/a2a-check/agenthealth.yaml --format json
```

The [Phase 6 example guide](../examples/a2a-check/README.md) includes passive,
bearer, custom-card, and functional configurations requiring your own peer.

## Configuration options

| Option | Behavior |
|---|---|
| `endpoint` | Establishes the trusted origin for discovery and the advertised RPC URL |
| `auth.bearer_env` | Resolves an existing bearer token from the environment |
| `a2a.card_url` | Overrides the well-known card path with an absolute same-origin URL |
| `a2a.protocol_version` | `1.0` (also the empty default) or explicit `0.3.0`; other pins are rejected locally |
| `a2a.required_skills` | Unique, nonblank skill IDs expected in the card |
| `a2a.required_capabilities` | Unique names from `streaming`, `pushNotifications`, `extendedAgentCard` (v1), and `stateTransitionHistory` (0.3) |
| `a2a.functional` | Requires `safe: true`, nonblank `text`, and explicit `functional` in `checks` |

Explicit `checks` must include `capability` when skill/capability expectations
are present. Configuration options do not propagate to dependencies; each peer
is configured independently. See the [configuration contract](../spec/configuration.md#a2a-expectations-phase-6-draft).

## Discovery and passive checks

The configured endpoint establishes the trusted origin. Discovery fetches `/.well-known/agent-card.json` at that origin; `a2a.card_url` can supply a different absolute card URL on the same origin. For v1 the first compatible `supportedInterfaces` entry selects a JSONRPC endpoint at `protocolVersion: 1.0`; its `tenant`, if present, is included in RPC parameters. Legacy mode uses the card's `url`. Discovered endpoints must have the same scheme, hostname, and effective port as the configured endpoint. Redirects are never followed, TLS is verified, and URL credentials and fragments are rejected.

Reachability measures the bounded card fetch, including its body. Receiving an HTTP error still establishes reachability; subsequent checks classify the error. Card data is cached for one target run. Protocol validation checks required agent metadata, skill identities, supported version, capabilities, and JSON-RPC transport. In legacy mode an omitted `preferredTransport` means JSONRPC. Input/output modes must parse as MIME media types; matching uses the normalized type and subtype, ignoring case and parameters.

Authentication and protocol checks share one read-only `GetTask` (v1) or `tasks/get` (0.3) request with a fresh random nonexistent task ID. A valid `TaskNotFoundError` (`-32001`) establishes protocol readiness without starting a task; an existing valid task response is also accepted. Authentication success means the configured credential was accepted for this passive endpoint. It does not establish authorization for every skill. Public card access alone is insufficient evidence of RPC authentication.

Credentials use `auth.bearer_env`. Public agents and declared HTTP bearer security alternatives are supported. A required scheme needing credentials or an unsupported scheme produces `MISCONFIGURED`. OAuth tokens may be supplied as bearer environment references; the adapter does not acquire them. Missing credentials fail before network access. Agent cards and RPC errors are never printed.

Capability checks validate required skill IDs and optional boolean capabilities: `streaming`, `pushNotifications`, `extendedAgentCard` (v1), and `stateTransitionHistory` (0.3). Missing requirements yield `UNHEALTHY`. Unknown required protocol extensions are unsupported.

<!-- spec-example: configuration -->
```yaml
version: v1
targets:
  - name: peer
    type: a2a
    endpoint: https://agent.example.com
    auth:
      bearer_env: AGENT_TOKEN
    a2a:
      protocol_version: '1.0'
      required_skills: [health]
      required_capabilities: [streaming]
    thresholds:
      latency_ms: 1000
```

## Opt-in interaction

A minimal interaction sends one `SendMessage` (v1) or `message/send` (0.3) request with a text message, a fresh message ID, blocking mode, and no requested history. The user must explicitly select a non-destructive prompt appropriate for the agent and declare `safe: true`. This is an assertion by the operator; the adapter cannot prevent an agent from acting on a prompt. It requires advertised default `text/plain` input and output modes.

<!-- spec-example: configuration -->
```yaml
version: v1
targets:
  - name: peer
    type: a2a
    endpoint: https://agent.example.com
    checks: [reachability, protocol, authentication, capability, functional, latency]
    a2a:
      required_skills: [health]
      functional:
        safe: true
        text: 'Report your readiness as text without invoking tools or making external changes.'
```

A structurally valid agent text message or completed task yields `HEALTHY`. Failed, rejected, and canceled tasks yield `UNHEALTHY`. Authentication-required tasks yield `MISCONFIGURED`; pending tasks, input requirements, and unknown task state yield `UNKNOWN`. There is no polling, continuation, cancellation request, or retry of the active interaction. No task output is exposed. Completion validates the interaction, not semantic correctness of the agent's answer.

## Limits and errors

Card and RPC bodies are capped at 1 MiB. Interaction text is capped at 64 KiB. Context cancellation and configured per-check deadlines apply to network and body reads. Passive retries apply only before any response bytes arrive; active checks never retry. Latency thresholds use the engine's normal `DEGRADED` policy.

HTTP 401/403 yields `MISCONFIGURED`; other unexpected HTTP statuses, malformed metadata/envelopes, and unexpected RPC errors yield `UNHEALTHY`. Unsupported server versions/transports and required protocol extensions yield `UNHEALTHY`; origin violations yield `MISCONFIGURED`. Oversized or interrupted bodies yield `UNKNOWN`; connection and TLS failures before a response yield `UNREACHABLE`. Diagnostics are engine-owned and exclude raw response data and errors.

## Version migration and interoperability

v0.4.0 changes the unpinned A2A default from 0.3.0 to 1.0. Existing legacy peers
must set `a2a.protocol_version: '0.3.0'`. Versions are selected before RPC calls;
there is no automatic downgrade after a v1 failure. JSON-RPC v1 requests carry
`A2A-Version: 1.0`, PascalCase methods, proto-JSON roles and text parts, and
`returnImmediately: false`. Send results contain exactly one task or message.
Task states accept canonical enum names or their numeric protobuf values.

See the [SDK interoperability matrix](interoperability.md) for the actual
servers tested and limits of the evidence. Card signatures are not verified;
this adapter checks operational compatibility, not agent identity certification.

## Dependency graph integration

[Phase 10](dependency-graph.md) adds explicit node IDs and references while preserving this adapter’s checks and safety contract. References to one A2A node share card/passive evidence and any explicitly opted-in task per run. Distinct route nodes retain their own evidence. A healthy A2A endpoint still does not prove first-agent communication. Critical/optional policy belongs to each edge. The [graph example](../examples/graph-check/README.md) shows the configuration pattern. This extension is included in v0.8.0.
