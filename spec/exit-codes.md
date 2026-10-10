# Exit Codes

> Status: Draft. Defines the exit code contract for `agenthealth check` and `agenthealth ping`, so automation (CI/CD, Kubernetes probes, shell scripts) can branch on outcome without parsing output. Referenced from [docs/architecture.md](../docs/architecture.md#exit-codes) and [Phase 3 — Universal CLI](../ROADMAP.md#phase-3--universal-cli).

## Mapping

| Exit code | Health state | Meaning |
|---|---|---|
| `0` | `HEALTHY` | Target (and all critical dependencies) healthy |
| `1` | `DEGRADED` | Target works but one or more non-fatal conditions are impaired |
| `2` | `UNHEALTHY` | Target reachable but cannot perform required functionality |
| `3` | `UNREACHABLE` | Target could not be contacted |
| `4` | `MISCONFIGURED` | Configuration prevented a valid check |
| `5` | `UNKNOWN` | A result document was produced, but the target's health could not be conclusively determined (see [Error Classification](health-model.md#error-classification)) |
| `6` | *(no target status)* | AgentHealth itself failed before producing any valid result document (e.g. invalid CLI invocation, unreadable configuration file, internal panic) |

Codes `0`–`5` correspond exactly to the six [health states](health-model.md#health-states), in [Severity Order](health-model.md#severity-order). Code `6` is reserved entirely for AgentHealth's own failures and is never used to represent a target's health.

## Rules

- Exit code reflects the **overall** status after [dependency aggregation](health-model.md#status-aggregation), not just the target's own direct checks.
- Code `5` (`UNKNOWN`) means AgentHealth ran successfully and produced a result, but that result's status is genuinely indeterminate (e.g. a caught adapter error, or an ambiguous response received before timeout). This differs from code `6`, which means no result was produced at all.
- Code `6` must never be used to represent a target's health state, so automation can distinguish "the target's health is unknown" (`5`, a valid but inconclusive result exists) from "the tool itself broke" (`6`, no result exists).
- `agenthealth ping` and `agenthealth check` use the same mapping so scripts behave consistently across both commands.
- This mapping is considered part of the public contract once [AgentHealth 1.0](../docs/roadmap-details.md#agenthealth-10) ships; changes after that point follow the [backward compatibility policy](../RELEASING.md#backward-compatibility-policy).

## Multiple targets

When `agenthealth check <config>` runs against a configuration with more than one top-level target, the output is a [batch envelope](result-schema.md#batch--check-run-envelope) containing one Result per target. The process exit code is the **most severe** code (per [Severity Order](health-model.md#severity-order)) across all `results` entries — one unhealthy target among ten healthy ones still produces a non-zero exit code. Code `6` is used only if the run itself could not produce a batch envelope at all (e.g. the configuration file failed to parse), never because one target within an otherwise successful run was unhealthy.

## Example

```bash
agenthealth check production.yaml
case $? in
  0) echo "healthy" ;;
  1) echo "degraded" ;;
  2) echo "unhealthy" ;;
  3) echo "unreachable" ;;
  4) echo "misconfigured" ;;
  5) echo "unknown / inconclusive" ;;
  6) echo "agenthealth internal error — no result produced" ;;
esac
```
