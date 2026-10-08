# Configuration Specification

> Status: Draft. Describes the declarative format used by `agenthealth check <file>` and `agenthealth doctor <file>`.

## Example

<!-- spec-example: configuration -->
```yaml
version: v1

targets:
  - name: research-agent
    type: agent
    endpoint: https://agent.example.com

    checks:
      - reachability
      - protocol
      - authentication
      - capability
      - functional
      - latency

    agent:
      functional:
        safe: true
        text: Return a fixed health acknowledgement without tools or writes.

    thresholds:
      latency_ms: 1000

    dependencies:
      - name: github-mcp
        type: mcp
        endpoint: http://github-mcp:3000
        critical: true

      - name: vector-store
        type: vector-store
        endpoint: http://vector-db:6333
        critical: true

      - name: analytics-api
        type: http
        endpoint: https://analytics.example.com/health
        critical: false
```

## Top-level fields

| Field | Type | Description |
|---|---|---|
| `version` | string | Configuration schema version (e.g. `v1`) |
| `targets` | list (min. 1 item) | One or more targets to check. A single target produces a single [Result](result-schema.md#result-object-recursive) document; more than one target produces a [batch envelope](result-schema.md#batch--check-run-envelope) |

A JSON Schema for this configuration format is available at [spec/schemas/configuration.schema.json](schemas/configuration.schema.json).

## Target fields

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Human-readable identifier for the target |
| `type` | string | yes | One of the [target types](target-model.md) |
| `endpoint` | string | yes | Address used to reach the target |
| `checks` | list | no | Which [health dimensions](health-model.md#health-dimensions) to run. If omitted, see [Default Checks](#default-checks) |
| `thresholds` | map | no | Numeric thresholds, e.g. `latency_ms`. See [Threshold Semantics](#threshold-semantics) |
| `dependencies` | list | no | Nested targets this target depends on. See [Dependency Inheritance](#dependency-inheritance) |

## Dependency fields

Each entry under `dependencies` accepts the same `name`/`type`/`endpoint`/`checks`/`thresholds`/`dependencies` fields as a target (dependencies may themselves have dependencies, recursively), plus:

| Field | Type | Required | Description |
|---|---|---|---|
| `critical` | boolean | no, defaults to `true` | Whether failure of this dependency should propagate to the parent target's overall status (see [Status Aggregation](health-model.md#status-aggregation)) |

## Default Checks

If `checks` is omitted for a target or dependency, AgentHealth runs the default set of **passive** dimensions applicable to that entry's `type`, per the [applicable-dimensions table](target-model.md#applicable-dimensions-per-target-type) in target-model.md: `reachability`, `protocol`, `authentication`, `capability`, `dependency`, `latency`, and `configuration`.

`functional` is **never** included in the default set, regardless of target type. [Functional Health](health-model.md#functional-health) checks exercise the target (running inference, executing a query, invoking an operation) and are classified **active** under [Passive vs Active Checks](../README.md#passive-vs-active-checks), so they require explicit opt-in. Listing `functional` in a target's or dependency's `checks` *is* that opt-in — no separate flag is needed.

For example, an `http` target with no `checks` listed runs reachability, protocol (HTTP status/header expectations), authentication, latency, and configuration (all passive and applicable to `http`), but never `functional` unless explicitly requested.

## Dependency Execution

The main example above declares `dependencies:` (`github-mcp`, `vector-store`, `analytics-api`) but omits `dependency` from the parent's `checks` list. This is intentional and well-defined, not an oversight:

**Every entry declared under a target's or dependency's `dependencies:` list always executes and always contributes to [Status Aggregation](health-model.md#status-aggregation) — this is driven entirely by the presence of entries in `dependencies:`, never by whether `dependency` appears in `checks`.** There is no way to declare a dependency and have it silently skipped; if you don't want a dependency evaluated, remove it from `dependencies:` rather than omitting `dependency` from `checks`.

Listing `dependency` in `checks` controls only one thing: whether a summary `checks.dependency` diagnostic entry (e.g. `{"status": "UNHEALTHY", "message": "critical dependency vector-store is UNREACHABLE"}`) is included in the result's `checks` map for human/machine-readable explanation. Omitting it omits that summary line — it has **no effect** on whether dependencies run, nor on the `dependencies` array in the result, nor on the parent's `overall_status`.

This is also why [target-model.md](target-model.md#applicable-dimensions-per-target-type)'s applicable-dimensions table marks the Dependency column — for types like `mcp` and `http` — as indicating only whether a `checks.dependency` summary entry is part of that type's *default* checks, not whether that type is allowed to declare `dependencies:`. Any target type may declare `dependencies:`, regardless of that table.

## Dependency Inheritance

Dependencies do **not** inherit `checks` or `thresholds` from their parent target. Each dependency entry is evaluated independently:

- if the dependency declares its own `checks`, those are used;
- otherwise, [Default Checks](#default-checks) applies based on the dependency's own `type` (not the parent's type).

This matters because a dependency's `type` is frequently different from its parent's (e.g. an `agent` target depending on a `vector-store`), so inheriting the parent's dimension list would often be meaningless.

## Threshold Semantics

`thresholds` apply only to the target or dependency entry on which they are declared — they are never inherited by that entry's own dependencies (consistent with [Dependency Inheritance](#dependency-inheritance)).

For `thresholds.latency_ms` specifically: if the `latency` dimension is run and the measured `latency_ms` exceeds the configured threshold, the `latency` check's status is `DEGRADED` (see [Error Classification](health-model.md#error-classification)). If `thresholds.latency_ms` is omitted, the `latency` dimension (if run) only measures and reports `latency_ms` without evaluating it against a threshold, and defaults to `HEALTHY` unless no response was received at all (in which case `UNREACHABLE` applies per the reachability/timeout rules, independent of any latency threshold).

## Execution policies (Phase 2 draft)

Each target and dependency accepts these fields; policies and credentials are never inherited by dependencies.

| Field | Type | Default | Meaning |
|---|---|---|---|
| `timeout_ms` | integer, 1–300000 | 5000 | Deadline for each check attempt |
| `check_timeouts_ms` | map of dimension to integer, 1–300000 | empty | Overrides `timeout_ms` for named dimensions, including implicit prerequisite gates; `dependency` is not allowed |
| `retries` | integer, 0–3 | 0 | Additional attempts after a passive check fails `UNREACHABLE` without receiving a response |
| `retry_delay_ms` | integer, 0–60000 | 100 | Cancellation-aware delay between attempts |
| `auth.bearer_env` | non-empty environment-variable name | absent | Environment variable containing a bearer credential; literal secrets are not accepted |

These additions are implemented as the reference engine's draft policy contract, not a stable 1.0 standard. Unknown fields, null fields, multiple YAML documents, and scalar type coercions are rejected. Only `version: v1` is currently supported. The engine also limits dependency depth to 64.

<!-- spec-example: configuration -->
```yaml
version: v1
targets:
  - name: private-api
    type: http
    endpoint: https://api.example.com/health
    auth:
      bearer_env: AGENTHEALTH_API_TOKEN
    timeout_ms: 5000
    check_timeouts_ms:
      reachability: 1000
    retries: 2
    retry_delay_ms: 100
    checks: [reachability, authentication, latency]
```

Missing or empty referenced credentials yield a `MISCONFIGURED` configuration check before network access. The environment is read once per run; resolved credentials never enter the result. Endpoint userinfo is rejected by the reference HTTP adapter; use `auth.bearer_env` instead.

Active checks (including `functional`) never retry, even when `retries` is configured. A partial response before timeout yields `UNKNOWN` and never retries. Authentication rejection, functional failures, configuration failures, and generic adapter errors do not retry. Retry budgets apply separately to each check. Cancellation interrupts attempts and retry delays.

The reference engine applies a 60-second whole-run budget, or a shorter caller deadline, encompassing all targets, dependencies, checks, attempts, and delays. Exhaustion stops further adapter work and produces diagnostics for remaining targets. Individual checks have their own deadlines, and at most 16 adapter calls can remain in flight per engine. Dependencies execute sequentially in declaration order and retain independent per-check policies; graph scheduling and configurable concurrency remain Phase 10 work.

`latency_ms` measures the final successful reachability attempt, excluding local adapter-slot queue time, earlier failed attempts, and retry delays. It is null when reachability did not succeed. A latency check uses this measurement and does not issue another request.

## HTTP response expectations (Phase 4 draft)

HTTP/API targets and dependencies accept an optional `http` map. Options are not inherited by dependencies.

| Field | Type | Default | Meaning |
|---|---|---|---|
| `http.expected_status` | nonempty list of integers, 200–599 | any 2xx | Accepted final response statuses |
| `http.headers` | map of header name to string | empty | Required response headers; names are case-insensitive, at least one value must match exactly |
| `http.body_contains` | string | empty | Literal UTF-8 text required in the GET response body; nonempty values require explicit `functional` in `checks` |
| `http.max_body_bytes` | integer, 1–1048576 | 65536 | Maximum body size accepted for body matching |

`protocol` is a passive HEAD check for status and response headers. It runs by default for HTTP/API. Status/header expectations require `protocol` or `functional` when an explicit check list is supplied, so configured expectations cannot silently be skipped. `functional` uses GET and applies the same expectations, plus optional body matching. No automatic HEAD-to-GET fallback is performed. A HEAD-incompatible endpoint can accept 405 explicitly or use an opt-in functional check.

A status/header/body mismatch is `UNHEALTHY`. Authentication rejection (401/403) is `MISCONFIGURED` even if listed in accepted statuses. An oversized or incomplete body is `UNKNOWN`, with no retry. Bodies are read only for body matching, up to the configured limit plus one byte. Expected values and response content never appear in diagnostics. Redirects are checked as responses and never followed.

<!-- spec-example: configuration -->
```yaml
version: v1
targets:
  - name: ready-api
    type: http
    endpoint: http://localhost:8080/health
    checks: [reachability, protocol, authentication, functional, latency]
    http:
      expected_status: [200]
      headers:
        Content-Type: application/json
      body_contains: '"ready":true'
      max_body_bytes: 65536
```

Transport observations appear as optional `checks.reachability.steps` entries for `dns`, `tcp`, `tls`, and `http`; see [HTTP diagnostics](../docs/http-adapter.md). These describe the observed network path rather than additional health dimensions.

## Still to be defined

- Additional authentication schemes and secret-manager references
- General environment-variable interpolation beyond `auth.bearer_env`
- Stable 1.0 policy contract after implementation feedback

The implemented policy contract above resolves the Phase 2 timeout, retry, and environment credential-reference requirements. Further schemes remain future extensions.

## MCP expectations and safe invocation (Phase 5 draft)

MCP targets may specify `mcp` options on top-level targets or dependencies.
They use the [MCP adapter contract](../docs/mcp-adapter.md) for Streamable HTTP
or configured stdio:

- `protocol_version`: optional pin to `2025-03-26`, `2025-06-18`, `2025-11-25`, or `2026-07-28`.
  An omitted or empty value probes the modern protocol and permits legacy
  fallback as described by the adapter. A pin disables cross-era fallback.
- `required_tools`, `required_resources`, `required_prompts`: optional arrays
  of unique, nonblank identities. Tools/prompts match names; resources match
  URIs. Missing requirements produce `DEGRADED`. Explicit checks must include
  `capability` when nonempty requirements are specified.
- `functional`: optional invocation with nonblank `tool`, `safe: true`, and
  optional `arguments_json` (empty or a JSON object string, at most 64 KiB).
  The Go loader additionally parses this string; JSON Schema does not parse
  embedded JSON. Invocation requires explicit `functional` in `checks`, and
  every MCP functional check requires invocation options. The operator must
  verify safety; tool annotations must additionally declare read-only and
  non-destructive behavior. Functional calls never retry.

MCP options on other target types, unsupported version pins, duplicate/blank
requirements, and missing functional safety declarations are invalid
configuration. The shared result schema and exit codes are unchanged.

<!-- spec-example: configuration -->
```yaml
version: v1
targets:
  - name: local-mcp
    type: mcp
    endpoint: http://localhost:3000/mcp
    checks: [protocol, capability, functional, latency]
    mcp:
      required_tools: [search]
      functional:
        tool: search
        safe: true
        arguments_json: '{"query":"health"}'
```

### MCP transports and OAuth (Phase 5 draft extension)

`mcp.transport` is `http` (default) or `stdio`. Stdio requires `mcp.stdio` with
nonblank `command`, optional string `args`, optional `directory`, and optional
`env` mapping child variable names to host environment references. Commands,
arguments, and directories cannot contain NUL bytes. Stdio options require
stdio transport, and stdio targets cannot use `auth` or `mcp.oauth`.
The adapter requires a `stdio://<identity>` endpoint and invokes the command
without a shell. Each target run owns one process shared across dimensions and cleans it up
after checking; canceled attempts discard the process.

HTTP OAuth targets use `mcp.oauth` instead of `auth`. Required fields are
nonblank `issuer`, `client_id`, and `grant` (`client_credentials`,
`refresh_token`, or `authorization_code`). Optional fields are
`client_secret_env`, `refresh_token_env`, `token_file`, unique ASCII OAuth
`scopes`, and `redirect_port` (integer 0–65535).

- Client credentials require `client_secret_env`.
- Refresh token grants require `refresh_token_env` or `token_file`.
- Authorization code grants require `token_file`; the separate `login` command
  obtains credentials through PKCE and a loopback callback.

Environment references must be nonblank and contain neither `=` nor NUL.
Configured references are snapshotted per run, must resolve to nonempty values,
and participate in redaction. Acquired credentials are registered dynamically
for redaction and cached only within the target's run. Token files are bound to
the issuer, resource, and client ID, with private file permissions.
OAuth configuration never triggers interactive login during a health check.

The adapter enforces secure issuer/metadata/token URLs, issuer/resource
validation, modern per-request metadata and mirrored HTTP headers, legacy
negotiation rules, and bounded subprocess cleanup described in its contract.
No new health dimensions or result fields are introduced.

## A2A expectations (Phase 6 draft)

Targets of type `a2a` may include `a2a` options. `protocol_version` accepts
`1.0` (the empty default) or explicit `0.3.0`. v0.4.0 changes the
reference adapter default; legacy peers must pin `0.3.0`. `card_url` overrides well-known card discovery;
the adapter validates that it is an absolute HTTP(S) URL on the target origin.
`required_skills` lists nonempty unique skill IDs. `required_capabilities` lists
unique names from `streaming`, `pushNotifications`, `extendedAgentCard` (v1),
and `stateTransitionHistory` (legacy).
Expectations require the capability dimension when `checks` is explicit.

`a2a.functional` requires explicit `functional` in `checks`, `safe: true`, and
nonblank `text` of at most 65536 UTF-8 bytes. A functional check on an A2A target
requires these interaction options. The safety declaration asserts that the
operator has selected a non-destructive interaction. Credentials use the existing
`auth.bearer_env` reference. Options apply independently to nested dependencies.
Unknown fields, null options, duplicate requirements, and options on other target
types are rejected. See the [A2A adapter guide](../docs/a2a-adapter.md) for examples
and runtime behavior.

## Agent expectations (Phase 7 draft)

`agent` and `multi-agent` targets and dependencies accept `agent` options:
`required_capabilities` is a unique list of nonempty names and requires the
capability dimension when checks are explicit. `functional` requires explicit
functional checks, `safe: true`, nonempty `text` up to 64 KiB and optional
nonempty `downstream` up to 1024 characters. The downstream selector is sent to
the first runtime; it is not a discovered URL. Names-only discovery validates
that advertised dependencies are configured; execution uses the existing
independent nested dependency policies. Defaults include the dependency
summary for both agent types. See the [normative adapter interface](../docs/agent-adapter.md)
and [impact/compatibility proposal](../docs/rfcs/phase-7-agent-health.md).

<!-- spec-example: configuration -->
```yaml
version: v1
targets:
  - name: first
    type: agent
    endpoint: http://localhost:8080/health
    checks: [protocol, capability, functional]
    agent:
      required_capabilities: [answer]
      functional:
        safe: true
        text: Return a fixed acknowledgement without tools or writes.
```

Legacy agent configurations listing functional without task options remain
syntactically valid but yield MISCONFIGURED before networking; they previously
had no executable agent adapter. Add a safe task to enable functional execution.
