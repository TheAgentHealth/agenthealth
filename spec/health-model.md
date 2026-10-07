# Agent Health Specification (AHS)

> Status: Draft. This is the most mature layer of the project (see [Specification → Protocol → Implementation](../README.md#specification--protocol--implementation)).

The Agent Health Specification defines the vocabulary and health model used by every AgentHealth target type, adapter, and output format. It defines **what health means**, independent of any transport, language, or vendor.

## Health States

AgentHealth defines a common health vocabulary:

| Status | Meaning |
|---|---|
| `HEALTHY` | Target is operating normally |
| `DEGRADED` | Target works but one or more non-fatal conditions are impaired |
| `UNHEALTHY` | Target is reachable but cannot perform required functionality |
| `UNREACHABLE` | Target cannot be contacted |
| `MISCONFIGURED` | Configuration prevents a valid health determination or operation |
| `UNKNOWN` | Health cannot currently be determined |

Adapters may expose additional diagnostic detail, but MUST map their overall result onto one of these six states.

## Severity Order

Health states have a canonical severity order, from least to most severe:

```text
HEALTHY < DEGRADED < UNHEALTHY < UNREACHABLE < MISCONFIGURED < UNKNOWN
```

This order is used to:

- resolve a target's own overall status across multiple dimension results (take the most severe),
- resolve dependency contributions during [Status Aggregation](#status-aggregation),
- select a single CLI exit code across multiple targets in a batch run (see [exit-codes.md](exit-codes.md)).

`UNKNOWN` is deliberately treated as the most severe state: an indeterminate result MUST NOT be treated as better than a confirmed failure, since automation cannot safely assume health it cannot verify.

## Health Dimensions

Each dimension answers a narrower question about a target.

### Reachability

Can AgentHealth establish communication with the target? (TCP/HTTP connectivity, MCP transport connectivity, A2A endpoint availability, model API availability, database connectivity.)

### Protocol Health

Does the target correctly support the expected protocol? (MCP initialization, protocol version compatibility, A2A discovery, required endpoint validation, capability negotiation.)

### Authentication

Can the configured identity authenticate successfully? (API keys, OAuth credentials, service credentials, workload identity, token validity.) Secrets MUST NOT be exposed in AgentHealth output.

### Capability Health

Does the target expose the capabilities required by the workload? (MCP tools/resources, agent skills, model availability, expected API operations, required runtime features.)

### Functional Health

Can the target successfully perform a minimal operation that actually exercises it (not merely list what it claims to support — that's Capability Health)? Examples: a minimal model inference request, a lightweight API operation, database `SELECT 1`, a vector store read operation.

Functional checks are classified **active** per [Passive vs Active Checks](../README.md#passive-vs-active-checks): they MUST NOT run by default for any target type and require explicit opt-in by being listed in `checks` (see [configuration.md § Default Checks](configuration.md#default-checks)). Even when opted in, functional checks SHOULD remain minimal, safe, and non-destructive.

### Dependency Health

Are the target's required dependencies healthy? An agent may be reachable while a dependency is not; overall status reflects dependency propagation rules (critical vs optional — see [Status Aggregation](#status-aggregation)).

### Latency Health

Response latency compared against configurable thresholds (e.g. `latency_ms` vs `thresholds.latency_ms`).

### Configuration Health

Detects configuration problems that can be identified *before* attempting to reach or operate the target — missing endpoint, missing credentials, invalid or unsupported protocol selection, malformed configuration values — and reports them as `MISCONFIGURED` rather than misclassifying the target as unreachable.

Problems that can only be discovered by actually contacting the target — such as an incompatible protocol version returned during negotiation, or a required capability/model the target does not expose — are Protocol Health and Capability Health concerns respectively, not Configuration Health, and resolve to `UNHEALTHY` per [Error Classification](#error-classification).

## Target Model

AgentHealth targets are typed (`agent`, `mcp`, `a2a`, `model`, `http`, `database`, `vector-store`, and others). See [target-model.md](target-model.md) for the full list, required fields, and which health dimensions apply to which target type.

## Error Classification

Adapters and the core engine MUST classify failures consistently, so the same underlying problem produces the same status regardless of which adapter encountered it. This table is the single source of truth; [adapter-spec.md](adapter-spec.md) and [docs/architecture.md](../docs/architecture.md#error-model) reference it rather than restating it.

| Failure mode | Resulting status | Rationale |
|---|---|---|
| DNS resolution failure | `UNREACHABLE` | Cannot establish communication |
| TCP connection refused / connection reset | `UNREACHABLE` | Cannot establish communication |
| TLS handshake failure | `UNREACHABLE` | Cannot establish communication |
| Request timeout with no response received | `UNREACHABLE` | Treated as a connectivity failure, not an indeterminate one |
| Partial/ambiguous response received before a timeout (target may still be processing) | `UNKNOWN` | Genuinely indeterminate — distinct from a confirmed connectivity failure |
| Authentication rejected (missing, invalid, or expired credentials) | `MISCONFIGURED` | Authentication is a [Configuration Health](#configuration-health) concern |
| Malformed or invalid configuration detected before a check runs | `MISCONFIGURED` | The check could not be meaningfully attempted |
| Protocol negotiation failure / incompatible protocol version | `UNHEALTHY` | Target is reachable but cannot perform required functionality |
| Required capability missing | `UNHEALTHY` | Target is reachable but cannot perform required functionality |
| Optional capability missing | `DEGRADED` | Non-fatal gap in functionality |
| Functional check fails (minimal operation errors) | `UNHEALTHY` | Target is reachable but cannot perform required functionality |
| Latency exceeds configured threshold | `DEGRADED` | Non-fatal performance impairment |
| Dependency failure (critical or optional) | see [Status Aggregation](#status-aggregation) | Propagation rules differ from a target's own direct failures |
| Adapter/engine error, but the engine still produces a result document | `UNKNOWN` | The target's actual health is genuinely unknown; reflected in the result itself, not just the exit code (see [exit-codes.md](exit-codes.md)) |
| Adapter/engine error prevents producing any result document at all | *(no result produced)* | Surfaces only as CLI exit code `6` (internal error), never as a target status |

## Check Prerequisites

Some dimensions can only be meaningfully evaluated if an earlier dimension already succeeded. AgentHealth defines the following prerequisite chain:

1. **Configuration** is evaluated first, before attempting to reach the target at all. If it fails, no other dimension runs: `status` is `MISCONFIGURED` and `checks` contains only the `configuration` entry.
2. **Reachability** is attempted next. If it fails, **Protocol**, **Authentication**, **Capability**, **Functional**, and **Latency** are all blocked and skipped.
3. **Protocol** and **Authentication** each only require Reachability to have succeeded; they do not block each other.
4. **Capability** and **Functional** each require Reachability to have succeeded, and Authentication to have succeeded if authentication applies to the target type.
5. **Latency** can be measured whenever Reachability succeeded, regardless of the outcome of any other dimension.
6. **Dependency** is independent of all of the above: declared dependencies always execute and aggregate on their own, regardless of the parent target's own prerequisite chain (see [configuration.md § Dependency Execution](configuration.md#dependency-execution)).

### Skipped checks are omitted, not `UNKNOWN`

A dimension blocked by a failed prerequisite MUST be **omitted entirely** from `checks` — it MUST NOT be recorded with a status of `UNKNOWN` or any other value. [`own_status`](#step-1--own-status) is computed only over the dimensions that actually ran and produced a `checks` entry.

This matters because `UNKNOWN` is the most severe state in the [Severity Order](#severity-order): if a blocked Protocol/Authentication/Capability check were recorded as `UNKNOWN` merely because Reachability failed first, `own_status` would incorrectly become `UNKNOWN` instead of `UNREACHABLE`, masking the actual, more specific cause.

This is distinct from the empty-set rule in [Status Aggregation § Step 1](#step-1--own-status) (`own_status = UNKNOWN` when *zero* checks ran at all, e.g. `checks: []` was explicitly configured): a target whose Reachability check ran and failed has produced one check result, not zero, so Step 1 correctly yields `UNREACHABLE` in that case.

## Status Aggregation

When a target has dependencies, its overall status is computed in three steps.

### Step 1 — Own status

A target's **own status** is the most severe result (per [Severity Order](#severity-order)) across its own direct dimension checks (reachability, protocol, authentication, capability, functional, latency, configuration). It does not yet factor in dependencies.

If zero dimension checks actually ran for a target (e.g. `checks: []` was explicitly configured, disabling every dimension), `own_status` is `UNKNOWN`: the absence of any evidence is itself an indeterminate result, not a `HEALTHY` default. `max_severity` is otherwise undefined over an empty set; this is the one specified exception. (This is distinct from checks that ran but were *blocked* by a failed prerequisite — see [Check Prerequisites](#check-prerequisites).)

### Step 2 — Dependency contribution

Each dependency contributes to the parent's overall status based on its own status and its `critical` flag. **`critical` defaults to `true` when omitted** — a dependency is assumed load-bearing unless explicitly marked optional, so an omitted `critical` field cannot silently hide a real failure.

| Dependency status | `critical: true` (or omitted) | `critical: false` |
|---|---|---|
| `HEALTHY` | no contribution | no contribution |
| `DEGRADED` | `DEGRADED` | `DEGRADED` |
| `UNHEALTHY` | `UNHEALTHY` | `DEGRADED` |
| `UNREACHABLE` | `UNHEALTHY` | `DEGRADED` |
| `MISCONFIGURED` | `UNHEALTHY` | `DEGRADED` |
| `UNKNOWN` | `UNKNOWN` | `DEGRADED` |

Critical dependency failures contribute `UNHEALTHY` to the parent — **not** the dependency's literal status. A parent with a down critical dependency is still reachable and configured correctly itself; it simply cannot fully perform its function, which is exactly what `UNHEALTHY` means. Literally propagating a dependency's `UNREACHABLE` status to a parent that is itself perfectly reachable would contradict the definition of `UNREACHABLE`.

A critical dependency in `UNKNOWN` state contributes `UNKNOWN`: if a load-bearing dependency's health cannot be determined, the parent's ability to perform its function is equally indeterminate, so it would be inaccurate to call the parent confidently `HEALTHY` or `UNHEALTHY`.

Optional (`critical: false`) dependency failures of any kind contribute `DEGRADED` — the parent is not blocked, only diminished.

### Step 3 — Overall status

```text
overall_status = max_severity(own_status, contribution_1, contribution_2, ..., contribution_n)
```

using the [Severity Order](#severity-order).

### Worked example: reachable parent, failed critical dependency

A `research-agent` target itself passes every direct check (`own_status = HEALTHY`), but its critical `vector-store` dependency is `UNREACHABLE`:

```text
own_status                              = HEALTHY
vector-store (critical, UNREACHABLE)  → contributes UNHEALTHY
overall_status = max(HEALTHY, UNHEALTHY) = UNHEALTHY
```

The parent is reported `UNHEALTHY`, never `UNREACHABLE` — the agent itself was reachable; the failure was in what it depends on. See this same example worked through the full result document in [result-schema.md](result-schema.md#recursive-shape).

### Nested dependencies

A dependency may itself have dependencies (see the recursive shape in [result-schema.md](result-schema.md#recursive-shape)). Each level of the tree applies Steps 1–3 independently and bottom-up: a dependency's own `status`, as already recorded in the result tree, reflects its own dependency aggregation before it is used as an input to its parent's Step 2.

## Relationship to other spec documents

- Target types and per-type applicable dimensions: [target-model.md](target-model.md)
- Wire-level result representation: [result-schema.md](result-schema.md)
- Declarative target/check/dependency configuration: [configuration.md](configuration.md)
- How adapters must implement these semantics for a given technology: [adapter-spec.md](adapter-spec.md)
- CLI exit code mapping: [exit-codes.md](exit-codes.md)
- Security requirements for any conforming implementation: [security.md](security.md)
- How these semantics could be exposed/exchanged over a network: [protocol.md](protocol.md) (experimental)
