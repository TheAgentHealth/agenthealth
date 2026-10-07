# Result Schema

> Status: Draft. Defines the canonical machine-readable health result produced by `agenthealth check --output json`, and referenced by the proposed [AHP response envelope](protocol.md#proposed-response-envelope-non-normative).

## Example

```json
{
  "spec_version": "v1",
  "target": {
    "name": "research-agent",
    "type": "agent"
  },
  "status": "DEGRADED",
  "latency_ms": 182,
  "checks": {
    "reachability": { "status": "HEALTHY" },
    "authentication": { "status": "HEALTHY" },
    "protocol": { "status": "HEALTHY" },
    "capabilities": { "status": "HEALTHY" },
    "dependencies": { "status": "DEGRADED" }
  },
  "dependencies": [
    {
      "name": "github-mcp",
      "type": "mcp",
      "status": "HEALTHY"
    },
    {
      "name": "vector-store",
      "type": "vector-store",
      "status": "UNREACHABLE"
    }
  ]
}
```

## Fields

| Field | Type | Description |
|---|---|---|
| `spec_version` | string | Version of this result schema (e.g. `v1`) |
| `target.name` | string | Identifier of the target being checked |
| `target.type` | string | One of the [target types](health-model.md#target-model) |
| `status` | string | One of the [health states](health-model.md#health-states) |
| `latency_ms` | number | Measured latency in milliseconds |
| `checks` | map | Per-[dimension](health-model.md#health-dimensions) result, keyed by dimension name |
| `dependencies` | list | Results for each dependency, recursively shaped like a target result |

## Rules

- `status` MUST be one of the six values defined in [health-model.md](health-model.md#health-states).
- `checks` keys SHOULD correspond to the dimensions actually run for the target (not all dimensions are required for every target type).
- Secrets and credentials MUST NOT appear anywhere in this structure.
- `dependencies` entries recursively follow this same shape so dependency trees can be walked by generic tooling.

## Open items

- Exit code mapping (tracked under [Phase 3 — Universal CLI](../ROADMAP.md#phase-3--universal-cli))
- Error representation for internal failures vs target-reported failures
- Extension fields / vendor-specific metadata namespace
