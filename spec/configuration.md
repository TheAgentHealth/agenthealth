# Configuration Specification

> Status: Draft. Describes the declarative format used by `agenthealth check <file>` and `agenthealth doctor <file>`.

## Example

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
      - capabilities
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
| `targets` | list | One or more targets to check |

## Target fields

| Field | Type | Description |
|---|---|---|
| `name` | string | Human-readable identifier for the target |
| `type` | string | One of the [target types](health-model.md#target-model) |
| `endpoint` | string | Address used to reach the target |
| `checks` | list | Which [health dimensions](health-model.md#health-dimensions) to run |
| `thresholds` | map | Numeric thresholds, e.g. `latency_ms` |
| `dependencies` | list | Nested targets this target depends on |

## Dependency fields

Each entry under `dependencies` accepts the same `name`/`type`/`endpoint` fields as a target, plus:

| Field | Type | Description |
|---|---|---|
| `critical` | boolean | Whether failure of this dependency should propagate to the parent target's overall status (see [Status Aggregation](health-model.md#status-aggregation)) |

## Still to be defined

- Authentication references (how credentials are referenced without being embedded in plaintext)
- Retry policies
- Active vs passive check opt-in per target
- Per-check timeout overrides
- Environment variable / secret-manager interpolation syntax

These are tracked under [Phase 1 — Agent Health Specification](../ROADMAP.md#phase-1--agent-health-specification) in the roadmap.
