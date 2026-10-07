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

For example, an `http` target with no `checks` listed runs reachability, authentication, latency, and configuration (all passive and applicable to `http`), but never `functional` unless explicitly requested.

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

## Still to be defined

- Authentication references (how credentials are referenced without being embedded in plaintext)
- Retry policies
- Per-check timeout overrides
- Environment variable / secret-manager interpolation syntax

These are tracked under [Phase 1 — Agent Health Specification](../ROADMAP.md#phase-1--agent-health-specification) in the roadmap.
