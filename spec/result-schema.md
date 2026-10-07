# Result Schema

> Status: Draft. Defines the canonical machine-readable health result produced by `agenthealth check <configuration.yaml> --format json` (see [Phase 3 — Universal CLI](../ROADMAP.md#phase-3--universal-cli)), and referenced by the proposed [AHP response envelope](protocol.md#proposed-response-envelope-non-normative).

## Recursive Shape

A single target's result (a **Result** object) has the same shape whether it is the top-level target being checked or an entry in another Result's `dependencies` list:

```json
{
  "target": { "name": "github-mcp", "type": "mcp" },
  "status": "HEALTHY",
  "latency_ms": 84,
  "checks": {
    "reachability": { "status": "HEALTHY" }
  },
  "dependencies": []
}
```

The top-level document wraps one Result with a `spec_version`. This example shows the [Status Aggregation worked example](health-model.md#worked-example-reachable-parent-failed-critical-dependency): a reachable parent (`research-agent`) whose critical `vector-store` dependency is `UNREACHABLE`, so the parent's own status is `UNHEALTHY`, not `UNREACHABLE`:

<!-- spec-example: result -->
```json
{
  "spec_version": "v1",
  "target": { "name": "research-agent", "type": "agent" },
  "status": "UNHEALTHY",
  "latency_ms": 182,
  "checks": {
    "reachability": { "status": "HEALTHY" },
    "authentication": { "status": "HEALTHY" },
    "protocol": { "status": "HEALTHY" },
    "capability": { "status": "HEALTHY" },
    "functional": { "status": "HEALTHY" },
    "latency": { "status": "HEALTHY" },
    "dependency": {
      "status": "UNHEALTHY",
      "message": "critical dependency vector-store is UNREACHABLE"
    }
  },
  "dependencies": [
    {
      "target": { "name": "github-mcp", "type": "mcp" },
      "status": "HEALTHY",
      "latency_ms": 84,
      "checks": {
        "reachability": { "status": "HEALTHY" }
      },
      "dependencies": []
    },
    {
      "target": { "name": "vector-store", "type": "vector-store" },
      "status": "UNREACHABLE",
      "latency_ms": null,
      "checks": {
        "reachability": {
          "status": "UNREACHABLE",
          "message": "connection refused"
        }
      },
      "dependencies": []
    }
  ]
}
```

A JSON Schema enforcing this recursive shape is available at [spec/schemas/result.schema.json](schemas/result.schema.json).

## Fields

### Result object (recursive)

| Field | Type | Required | Description |
|---|---|---|---|
| `target.name` | string | yes | Identifier of the target being checked |
| `target.type` | string | yes | One of the [target types](target-model.md) |
| `status` | string | yes | One of the [health states](health-model.md#health-states) |
| `latency_ms` | number \| null | no | Measured latency in milliseconds; `null` if no response was received (e.g. `UNREACHABLE`) |
| `checks` | map | yes | Per-[dimension](health-model.md#health-dimensions) result, keyed by dimension name. MAY be an empty object `{}` only when `status` is `MISCONFIGURED` (problem detected before any check could run) or `UNKNOWN` (zero checks were configured to run, see [health-model.md § Status Aggregation](health-model.md#status-aggregation)); for every other status at least one entry MUST be present, since that status can only have been derived from evidence a check produced. Dimensions blocked by a failed prerequisite (see [health-model.md § Check Prerequisites](health-model.md#check-prerequisites)) are omitted, not recorded as `UNKNOWN` |
| `dependencies` | list of Result | yes (may be empty `[]`) | One entry per configured dependency, recursively shaped like this same Result object |

### Per-check entry (`checks.<dimension>`)

| Field | Type | Required | Description |
|---|---|---|---|
| `status` | string | yes | One of the [health states](health-model.md#health-states) for this specific dimension |
| `code` | string | no | Stable diagnostic identifier matching `^[a-z][a-z0-9_]{0,63}$`; consumers MUST tolerate unknown codes and absent codes. |
| `message` | string | no | Human-readable detail (e.g. `"connection refused"`, `"latency 1842ms exceeds threshold 1000ms"`). MUST NOT contain secrets. |

### Top-level document

| Field | Type | Required | Description |
|---|---|---|---|
| `spec_version` | string | yes | Version of this result schema (e.g. `v1`) |
| *(all Result fields)* | — | yes | The top-level document is a Result object plus `spec_version` |

## Batch / Check-Run Envelope

`agenthealth check <config>` runs against a configuration that MAY declare multiple top-level targets (see [configuration.md](configuration.md)). The output for a multi-target run is a **batch envelope**, not a single Result:

<!-- spec-example: result -->
```json
{
  "spec_version": "v1",
  "results": [
    {
      "target": { "name": "research-agent", "type": "agent" },
      "status": "HEALTHY",
      "latency_ms": 90,
      "checks": {
        "reachability": { "status": "HEALTHY" },
        "authentication": { "status": "HEALTHY" }
      },
      "dependencies": []
    },
    {
      "target": { "name": "billing-agent", "type": "agent" },
      "status": "DEGRADED",
      "latency_ms": 1200,
      "checks": {
        "reachability": { "status": "HEALTHY" },
        "latency": {
          "status": "DEGRADED",
          "message": "latency 1200ms exceeds threshold 1000ms"
        }
      },
      "dependencies": []
    }
  ]
}
```

| Field | Type | Description |
|---|---|---|
| `spec_version` | string | Version of this result schema |
| `results` | list of Result | One entry per top-level target declared in the configuration, in declaration order |

`agenthealth ping <type> <target>` (a single ad-hoc target, no configuration file) always produces a single top-level Result document, not a batch envelope, since there is exactly one target.

A JSON Schema for the batch envelope is available at [spec/schemas/result.schema.json](schemas/result.schema.json).

## Rules

- `status` MUST be one of the six values defined in [health-model.md](health-model.md#health-states).
- `checks` keys SHOULD correspond to the dimensions actually run for the target; see [configuration.md § Default Checks](configuration.md#default-checks) for what runs when `checks` is omitted.
- `checks` MUST be empty (`{}`) only when `status` is `MISCONFIGURED` or `UNKNOWN`; every other status requires at least one `checks` entry as evidence (see the `checks` field description above).
- Dimensions blocked by a failed prerequisite MUST be omitted from `checks` entirely, never recorded as `UNKNOWN` (see [health-model.md § Check Prerequisites](health-model.md#check-prerequisites)).
- Secrets and credentials MUST NOT appear anywhere in this structure, including in `message` fields.
- `dependencies` entries recursively follow the exact same Result shape (see [Recursive Shape](#recursive-shape)) so generic tooling can walk the tree without special-casing the root.
- For a batch run, the CLI's single process exit code reflects the **most severe** status across all `results` entries, per [Severity Order](health-model.md#severity-order) — see [exit-codes.md § Multiple targets](exit-codes.md#multiple-targets).

## Open items

- Extension fields / vendor-specific metadata namespace
- Whether `message` should be structured (error code + human-readable string) rather than free text

## Transport diagnostics

Check entries may include an optional `steps` map with `dns`, `tcp`, `tls`, and `http` keys and health-state values. The HTTP adapter reports observed transport stages under reachability. Missing stages mean they were not observed (for example, IP literals omit DNS and reused connections may omit TCP/TLS). In-progress stages at timeout are `UNKNOWN`; a failed stage is `UNREACHABLE`. These diagnostics do not independently contribute to aggregation.

## Diagnostic codes

Codes are independent of human message wording and do not replace status or exit
codes. The reference engine emits only allowlisted codes; arbitrary adapter code
and message text are suppressed. Checks without a canonical diagnostic omit code.
Other implementations may use additional identifiers conforming to the pattern.

| Code | Meaning |
|---|---|
| `a2a_card` | invalid A2A agent card metadata |
| `a2a_version` | unsupported A2A protocol version or transport |
| `a2a_origin` | A2A discovered endpoint must use the configured origin |
| `a2a_auth` | A2A authentication rejected or declared scheme is unsupported or missing credentials |
| `a2a_http` | A2A endpoint returned an unexpected HTTP status |
| `a2a_limit` | A2A response exceeded the safety size limit |
| `a2a_protocol` | invalid A2A JSON-RPC response |
| `a2a_rpc` | A2A server rejected the protocol request |
| `a2a_required` | required A2A skill or capability is missing |
| `a2a_functional` | A2A interaction failed |
| `a2a_pending` | A2A interaction requires continuation or has not completed |
| `mcp_process` | MCP subprocess could not start or exited before responding |
| `mcp_oauth` | MCP OAuth configuration, discovery or token acquisition failed |
| `mcp_login` | MCP OAuth login is required; use agenthealth login |
| `mcp_input` | MCP response requires unsupported input or continuation |
| `mcp_auth` | MCP authentication rejected (401 or 403) |
| `mcp_http` | MCP endpoint returned an unexpected HTTP status |
| `mcp_protocol` | invalid MCP initialization or JSON-RPC response |
| `mcp_version` | MCP protocol version is unsupported or differs from the pinned version |
| `mcp_limit` | MCP response or discovery exceeded safety limits |
| `mcp_rpc` | MCP server rejected the protocol request |
| `mcp_required` | required MCP tool, resource or prompt is missing |
| `mcp_unsafe` | functional tool lacks explicit read-only and non-destructive annotations |
| `mcp_functional` | MCP functional invocation failed |
| `http_status` | HTTP status did not match expected status |
| `http_headers` | required HTTP response header did not match |
| `http_body` | HTTP response body did not contain required text |
| `http_body_limit` | HTTP response body exceeded the configured size limit |
| `http_auth` | HTTP authentication rejected (401 or 403) |
