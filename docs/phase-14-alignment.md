# Phase 14 coverage and alignment

Kubernetes integration is implemented in source and remains unpublished pending
review. Chart 0.1.0 consumes the existing released CLI v0.11.0. No CLI release is
needed for this deployment-only change. Component publication remains separate.

## Coverage

| Deliverable | Implementation and evidence |
|---|---|
| Readiness and startup | AHP Deployment with `/ready` and `/live`; executable probe scenario with strict CLI exits |
| Init containers | External dependency gate, tested before application startup |
| Jobs and CronJobs | No retry, bounded deadline, forbidden schedule overlap; healthy and outage runtime checks |
| Sidecar | Conventional Deployment container, localhost check, shared Pod readiness |
| ConfigMap and Secrets | Read-only mounts, external Secret environment references, chart token entry validation |
| Helm | Independently versioned chart; serve/job/cronjob modes, schema validation, registry/digest overrides |
| Deployment validation | Pytest manifests/configuration/templates, kind server validation and runtime scenarios |
| Agent and path examples | Direct/composite fixture health and separate supporting gateway/router/route evidence; no first-to-peer claim |
| Distribution | Chart package/checksum helper and separate component tag workflow; no publication during implementation |
| Discovery | Guide, example instructions and prepared tag-pinned component notes |

See [the guide](kubernetes.md), [tests](../tests/kubernetes/test_kubernetes.py),
[runtime scenarios](../scripts/test_kubernetes.py), [CI](../.github/workflows/ci.yml)
and [component workflow](../.github/workflows/helm-release.yml).

## Earlier phases

| Phases | Impact or no-impact rationale |
|---|---|
| 0: Foundation | README, roadmap, indexes, release policy and CI reflect source availability and independent chart release. Governance unchanged. |
| 1: AHS | No schema, status, target, dimension, aggregation or exit-code changes. Embedded configuration validated against existing v1 schema. |
| 2: Engine | Existing timeout/budget/retry policies bound exec checks. No execution or diagnostic changes. |
| 3: CLI | Existing check/serve commands used directly with absolute exec path. Nonzero health exits fail probes/Jobs; tool failures remain exit 6. |
| 4–6: HTTP, MCP, A2A | No adapter changes. HTTP fixtures provide deterministic deployment evidence; MCP stdio/token-file use still requires suitable images/private writable storage. |
| 7: Agent health | Direct and composite health remain distinct from runtime communication. Passive examples add no functional authorization. |
| 8–9: Gateway/router | Separate signals and route evidence use existing HTTP contracts; no product discovery or certification added. |
| 10: Graph | Nested configurations remain valid. No graph model changes; explicit graphs can be supplied in a ConfigMap. |
| 11: AHP | Existing freshness and aggregate readiness preserved; startup/liveness use process health. No endpoint, authentication or refresh changes. |
| 12: Docker | Both public v0.11.0 registry references verified anonymously; image build/entrypoint/security remain unchanged. |
| 13: Distribution | No archive/package changes; chart packages and release tags are independent of CLI assets. |

## Later phases

Phases 15–16 SDKs need no new contract: chart packaging does not provide SDK
interfaces. Phase 21 CI integrations can reuse the explicit-context disposable
cluster scenarios and failed-Job gate. Phase 25 must add real framework,
A2A/MCP, gateway/router and Kubernetes-version interoperability rather than
claiming fixture certification. Phase 26 must cover chart provenance/signing,
registry digest maintenance, topology-specific network policy and broader
cluster/security validation. Other future adapters, observability and plugin
phases have no contract impact because Phase 14 only deploys existing commands.
Operator, CRD, admission and rollout automation remain deferred.

## Validation and remaining release work

Local checks passed on Linux AMD64: Helm 3.17.3 lint/template/package variants,
configuration schemas, Python tests, Go vet/race tests, documentation validation,
Kustomize server validation, derived-image binary execution, and Kubernetes
1.32.2 runtime checks in kind 0.27.0 with both the released image and a freshly
built source-image override. The runtime suite exercises
both readiness and Job failure during a dependency outage, accepted and
rejected Secret credentials without disclosure, init blocking/recovery, and
AHP/exec readiness recovery without restarts. Kubernetes 1.29 is
the declared API floor, not a tested runtime version. No cloud cluster or other
architecture was exercised.

The implementation is prepared on a review branch, including selected
contributor improvements. The existing roadmap layout edit was preserved.
No chart tag or component release was published. Release URL resolution and installation from the
published chart await authorized publication and cannot be claimed from local
source validation. Both image manifests were available anonymously on
October 8, 2026; this does not guarantee future mirror availability.

## Contributor integration

Selected ideas from [PR #15](https://github.com/TheAgentHealth/agenthealth/pull/15)
by Sharath K (`sharath568`) complement this implementation: Kustomize base and
fixture overlay, derived application image with the CLI, shared-dependency
readiness guidance, source-image loading into kind, and init outage/recovery
checks. The original outdated image references and incorrect optional-dependency
readiness advice were not copied; existing CLI exits are preserved. Contributor
credit is recorded here, in the guide and in the integration commit.
