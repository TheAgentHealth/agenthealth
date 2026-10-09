# All-phase documentation status

This index follows the source-tree scope in [ROADMAP.md](../ROADMAP.md).
Phases 0–7 have released baselines; Phase 8 is released in v0.6.0; Phase 9 in v0.7.0.
Phase 10 is included in v0.8.0; see the [alignment audit](phase-10-alignment.md).
Phase 11 experimental HTTP serving is included in v0.9.0; see its
[coverage and alignment audit](phase-11-alignment.md).
Phase 12 Docker distribution is released in v0.10.0.
Phase 13 package-channel tooling is included in v0.11.0.
Planned phases have roadmap designs, not shipped adapters, packages or runnable
integration examples. Existing repository CI and protocol interoperability tests
do not establish completion of the broader Phases 21, 25 or 26.

| Phase | Status | Documentation | Examples or validation |
|---|---|---|---|
| [0: Project Foundation](../ROADMAP.md#phase-0--project-foundation) | Released baseline | [Architecture](architecture.md) | [CI](../.github/workflows/ci.yml) |
| [1: Agent Health Specification](../ROADMAP.md#phase-1--agent-health-specification) | Released baseline | [Specification](../spec/README.md) | [Schema fixtures](../tests/spec/fixtures/) |
| [2: Core Engine](../ROADMAP.md#phase-2--core-engine) | Released baseline | [Core engine](../core/README.md) | [Core example](../examples/core-check/agenthealth.yaml) |
| [3: Universal CLI](../ROADMAP.md#phase-3--universal-cli) | Released baseline | [CLI](cli.md) | [Examples](../examples/README.md) |
| [4: Generic HTTP Health Adapter](../ROADMAP.md#phase-4--generic-http-health-adapter) | Released baseline | [HTTP adapter](http-adapter.md) | [HTTP readiness](../examples/core-check/agenthealth.yaml) |
| [5: MCP Health Adapter](../ROADMAP.md#phase-5--mcp-health-adapter) | Released baseline | [MCP adapter](mcp-adapter.md) | [MCP examples](../examples/README.md) |
| [6: A2A Health Adapter](../ROADMAP.md#phase-6--a2a-health-adapter) | Released baseline | [A2A adapter](a2a-adapter.md) | [A2A examples](../examples/a2a-check/README.md) |
| [7: Agent Health](../ROADMAP.md#phase-7--agent-health) | Released baseline | [Agent adapter](agent-adapter.md) | [Agent/peer/path example](../examples/agent-check/README.md) |
| [8: Agentgateway Integration](../ROADMAP.md#phase-8--agentgateway-integration) | Released in v0.6.0 | [Agentgateway](agentgateway.md) | [Gateway example](../examples/gateway-check/README.md) |
| [9: Agent Router Integration](../ROADMAP.md#phase-9--agent-router-integration) | Released in v0.7.0 | [Agent Router](agent-router.md) | [Router example](../examples/router-check/README.md) |
| [10: Dependency Graph](../ROADMAP.md#phase-10--dependency-graph) | Included in v0.8.0 | [Dependency graph](dependency-graph.md) | [Graph contract and validation](rfcs/phase-10-dependency-graph.md) |
| [11: Agent Health Protocol (AHP)](../ROADMAP.md#phase-11--agent-health-protocol-ahp) | Included in v0.9.0; experimental HTTP v1 | [AHP](ahp.md) | [Serving example](../examples/ahp-check/README.md) |
| [12: Docker Distribution](../ROADMAP.md#phase-12--docker-distribution) | Released in v0.10.0 | [Docker](docker.md) | [Dockerfile](../Dockerfile), [CI job](../.github/workflows/ci.yml), [release publish job](../.github/workflows/release.yml), [static Dockerfile tests](../tests/test_container_image.py) |
| [13: Standalone Binaries](../ROADMAP.md#phase-13--standalone-binaries) | Included in v0.11.0 | [Installation](installation.md), [release policy](../RELEASING.md), [alignment audit](phase-13-alignment.md) | [Distribution example](../examples/distribution/README.md), [release builder](../scripts/build_release.py), [package tests](../tests/test_package_release.py) |
| [14: Kubernetes Integration](../ROADMAP.md#phase-14--kubernetes-integration) | Released as helm-v0.1.0 | Chart 0.1.0; CLI v0.11.1 | [Kubernetes scenarios](../examples/kubernetes/README.md), [alignment audit](phase-14-alignment.md) |
| [15: Python SDK / PyPI](../ROADMAP.md#phase-15--python-sdk--pypi) | Planned | Roadmap design only | No implementation examples yet |
| [16: JavaScript / TypeScript SDK](../ROADMAP.md#phase-16--javascript--typescript-sdk) | Planned | Roadmap design only | No implementation examples yet |
| [17: Model / LLM Adapters](../ROADMAP.md#phase-17--model--llm-adapters) | Planned | Roadmap design only | No implementation examples yet |
| [18: Database Adapters](../ROADMAP.md#phase-18--database-adapters) | Planned | Roadmap design only | No implementation examples yet |
| [19: Vector Store Adapters](../ROADMAP.md#phase-19--vector-store-adapters) | Planned | Roadmap design only | No implementation examples yet |
| [20: Agent Runtime / Framework Integrations](../ROADMAP.md#phase-20--agent-runtime--framework-integrations) | Planned | Roadmap design only | No implementation examples yet |
| [21: CI/CD Integrations](../ROADMAP.md#phase-21--cicd-integrations) | Planned | Roadmap design only | No implementation examples yet |
| [22: Observability Export](../ROADMAP.md#phase-22--observability-export) | Planned | Roadmap design only | No implementation examples yet |
| [23: Plugin / Adapter Ecosystem](../ROADMAP.md#phase-23--plugin--adapter-ecosystem) | Planned | Roadmap design only | No implementation examples yet |
| [24: AgentHealth Conformance](../ROADMAP.md#phase-24--agenthealth-conformance) | Planned | Roadmap design only | No implementation examples yet |
| [25: Ecosystem Interoperability Testing](../ROADMAP.md#phase-25--ecosystem-interoperability-testing) | Planned | Roadmap design only | No implementation examples yet |
| [26: Production Hardening](../ROADMAP.md#phase-26--production-hardening) | Planned | Roadmap design only | No implementation examples yet |

Historical [release notes](README.md) describe their tagged versions. Phase 8
does not retroactively add gateway support to v0.5.0 binaries. For current adapter
limits, use the linked guides and the roadmap capability checklists.
