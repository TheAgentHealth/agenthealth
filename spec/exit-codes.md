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
| `5` | internal error | AgentHealth itself failed (not the target) |

## Rules

- Exit code reflects the **overall** status after [dependency aggregation](health-model.md#status-aggregation), not just the target's own direct checks.
- Exit code `5` is reserved for AgentHealth failures (e.g. invalid CLI invocation, unreadable configuration file, internal panic). It must never be used to represent a target's health state, so automation can distinguish "the target is unhealthy" from "the tool itself broke."
- `agenthealth ping` and `agenthealth check` use the same mapping so scripts behave consistently across both commands.
- This mapping is considered part of the public contract once [AgentHealth 1.0](../ROADMAP.md#agenthealth-10) ships; changes after that point follow the [backward compatibility policy](../RELEASING.md#backward-compatibility-policy).

## Example

```bash
agenthealth check production.yaml
case $? in
  0) echo "healthy" ;;
  1) echo "degraded" ;;
  2) echo "unhealthy" ;;
  3) echo "unreachable" ;;
  4) echo "misconfigured" ;;
  5) echo "agenthealth internal error" ;;
esac
```
