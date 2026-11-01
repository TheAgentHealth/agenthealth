# CI/CD deployment validation

Phase 21 adds `--format junit`, `--format github`, a reusable
[GitHub Action](../integrations/github-action/action.yml) and
[provider templates](../examples/ci/README.md). Run health checks after deployment
and allow promotion only when the health command returns 0.

## Reports and exit codes

```bash
agenthealth check deployment.yaml --format junit > agenthealth-junit.xml
agenthealth check deployment.yaml --format github
```

These are separate checks: choose one format per invocation to avoid repeating
active probes. The Action runs once, writes JUnit and emits a generic annotation.
Both formats preserve the CLI health exit mapping (0 healthy; 1 degraded;
2 unhealthy; 3 unreachable; 4 misconfigured; 5 unknown). Tool failures return 6
and may produce no complete report. An empty report is never proof of health.

JUnit has one testcase per top-level target and nested dependency occurrence,
with deterministic positional names (`target-1/dependency-1`) and target-type
classnames. All non-HEALTHY statuses produce failures, including UNKNOWN.
Optional dependency failures remain visible as failed cases while the parent
retains engine-owned DEGRADED aggregation. Shared graph IDs are not deduplicated
because edge policies may differ. This output is a CI projection; JSON/YAML
remain authoritative for dimensions, relationships, IDs and detailed evidence.

GitHub annotations use only fixed statuses, validated target types and positional
paths. Neither CI format includes target names, URLs, messages, secrets or raw
transport errors. This also prevents workflow-command injection. Preserve a
separate JSON report only if its topology exposure and access policy are acceptable;
doing so by rerunning checks can repeat functional probes. SARIF is deferred:
health failures are not static-analysis security findings.

## Exact-version installation and Action

The [installer](../scripts/install_ci_binary.py) selects one exact `vX.Y.Z`
platform asset, verifies SHA-256 and GitHub build provenance bound to the source
tag and release workflow, then extracts only the executable. It requires Python
3.10+, a recent authenticated `gh` supporting `--source-ref` and
`--deny-self-hosted-runners`, and network access. It supports the five released
platforms and refuses to overwrite an existing destination. Verification failure
must fail the pipeline; never fall back to an unverified binary.

Use the [GitHub example](../examples/ci/github.yml) from this source checkout.
Application repositories can reference
`TheAgentHealth/agenthealth/integrations/github-action@<reviewed-full-commit-SHA>`;
replace the placeholder with an immutable reviewed commit. The Action requires
an already installed verified CLI, so installing the Action alone does not install
the CLI. Do not reference a mutable branch for production deployment gates.

| Input | Behavior |
|---|---|
| `configuration` | Required trusted YAML path |
| `version` | Required exact stable CLI version, checked before execution |
| `binary` | Installed executable; default `agenthealth` |
| `report` | Workspace-relative JUnit path; default `agenthealth-junit.xml` |
| `timeout` | Process deadline, 1–3600 seconds; default 120 |

The `exit-code` output preserves health codes or 6 for launch, version, timeout
and tool failures. The Action removes stale reports, rejects report paths outside
the workspace or overlapping configuration, and does not print CLI stderr.
Upload reports with `if: always()`; a failed Action must still block promotion.
Hosted Linux, macOS and Windows runners provide Python and Bash. Self-hosted
runners must supply both. Python kills the CLI on the process deadline; configure
trusted MCP child processes with their own lifecycle controls as well.

## Earlier-phase deployment evidence

Use the existing [agent examples](../examples/agent-check/README.md) for direct
agents and composite peers, [gateway](../examples/gateway-check/README.md) and
[router](../examples/router-check/README.md) examples for configured path/backend
evidence, or [graph examples](../examples/graph-check/README.md) for explicit
relationships. A healthy A2A endpoint alone does not establish peer communication.
Model/database/vector targets remain supporting dependencies. Reports do not
change active opt-in, dependency policies, retries or SDK contracts.

## Kubernetes gates

Reuse [manifest/schema tests](../tests/kubernetes/) and
[the explicit-context smoke suite](../scripts/test_kubernetes.py):

```bash
python3 -m pytest tests/kubernetes -q
python3 scripts/test_kubernetes.py --context kind-disposable-ci --image agenthealth:ci
```

Provision a disposable cluster and load that exact test image first. The suite
checks completed/failed Jobs, readiness, restart counts and failure scenarios;
creating a Job or installing a chart alone is insufficient. For application Jobs,
wait for completion with a bounded timeout, collect status and logs on failure,
and fail the deployment gate if the Job fails or times out. Keep `/live`
independent of dependency readiness and preserve the existing chart security
context, secret references and check budgets.

## Credentials and trust

Configuration is executable trusted input. Run deployment gates only on reviewed
configuration with scoped credentials, restricted egress, appropriate timeouts
and isolated runners. Never expose deployment secrets to untrusted fork changes
or execute their configuration using `pull_request_target`. Keep credentials in
CI secret stores and environment references, not parameters, endpoints or reports.
See the [threat model](threat-model.md) and [alignment review](phase-21-alignment.md).
