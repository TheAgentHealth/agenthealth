# Stabilization validation record

Validated locally on 2026-10-09 against the source changes on
`fix/security-and-stabilization`. These results do not publish a release or
replace CI for the proposed commit. Published software remains v0.12.0.
The additive HTTP method contract is proposed in [RFC #46](https://github.com/TheAgentHealth/agenthealth/issues/46).

| Check | Result | Scope |
|---|---|---|
| Go vet and race tests | Passed | All Go packages, including reference handler contract tests |
| Production statement coverage | 86.27%; 80% gate passed | CLI/core/adapters; executable examples excluded from the aggregate |
| Python repository tests | 203 passed, 4 skipped | Skips are three dedicated SDK tests and one dedicated real-MCP test |
| Official current/legacy SDK suites | 3 passed per suite | Controlled MCP HTTP/stdio and A2A SDK applications |
| Maintained MCP filesystem server | Passed | 2026.8.31 passive protocol/inventory checks; no tool execution |
| Reference middleware | Passed | Go adapter contract, 3 FastAPI/compiled-LangGraph tests, Express typecheck and runtime test |
| Agentgateway | Passed | v1.6.0 checksum-verified binary, readiness HTTP signal |
| LiteLLM router | Passed | 1.104.2 process liveness; empty model list, isolated fixture environment and ephemeral master key |
| Released configuration compatibility | Passed | v0.9.0/current/v0.9.0 and v0.12.0/current/v0.12.0 with one unchanged HTTP configuration |
| Parser fuzz smoke | Passed | Five targets, five seconds each, two workers; YAML, JSON-RPC, SSE, OAuth discovery, A2A cards |
| DAG benchmarks/profile | Passed | 100, 1,000 and 5,000 nodes, each non-root referencing a shared node; within the 10,000-projection bound |
| Kubernetes static validation | Passed | Strict Helm lint and 18 manifest/chart tests |
| Kubernetes runtime | Passed | Disposable 1.32.2 cluster: init outage/recovery, probes, sidecar, Secrets, Jobs, CronJobs, Helm and readiness recovery |
| Documentation | Passed | Links, 22 marked examples, 28 standalone YAML examples, generated status/distribution tables |
| Workflow structure | Passed | YAML parsing and shell syntax; existing release-gate/path-detection tests |

The test-skip summarizer produces reason counts and individual test identities;
CI now uploads JUnit reports, skip summaries, coverage profiles and DAG profiles.
The local disposable cluster and product processes were removed after validation.

## Interpretation and remaining work

These are targeted implementation checks, not a security audit, product
certification, broad migration guarantee or proof of multi-tenant safety.
The [threat model](threat-model.md) distinguishes engine controls from operator
network/identity responsibilities. [Stabilization](stabilization.md) tracks broader
fault scenarios, long fuzz campaigns, full A2A framework interoperability,
router decisions, deployment migration and out-of-process isolation.

Additional CI Kubernetes versions 1.31.9 and 1.33.1 use images published in the
[kind v0.28.0 release](https://github.com/kubernetes-sigs/kind/releases/tag/v0.28.0).
Those matrix runs were configured but not executed locally. Do not claim those
versions are validated until runtime CI succeeds. Existing 1.32.2 release-gate
job identity is preserved.

The router smoke deliberately uses the lightweight liveness endpoint documented
by [LiteLLM](https://docs.litellm.ai/docs/proxy/health), rather than a model health
endpoint that triggers inference. Gateway/router HTTP checks still do not verify
routing decisions or automatically discover route tables.
