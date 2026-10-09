# Documentation guide

See the [all-phase status index](phase-status.md) for implementation, documentation,
and example coverage across Phases 0–26.

Phases 0–6 and foundation stabilization are implemented in
[v0.4.0](releases/v0.4.0.md). The specification remains
pre-1.0. Use this index to find the current contracts, behavior, and examples;
see the [roadmap](../ROADMAP.md) for limitations and later phases.

| Phase | Documentation | Examples or supporting files |
|---|---|---|
| 0: Foundation | [Architecture](architecture.md), [contributing](../CONTRIBUTING.md), [governance](../GOVERNANCE.md), [security policy](../SECURITY.md), [releases](../RELEASING.md) | [CI](../.github/workflows/ci.yml), [installation](installation.md) |
| 1: Specification | [Specification index](../spec/README.md): health states, target model, configuration, results, exit codes, adapter contract, security | [Schema fixtures](../tests/spec/fixtures/) and marked examples in the specification |
| 2: Engine | [Core API and execution guarantees](../core/README.md) | [Go API example](../examples/core-check/main.go), [configuration](../examples/core-check/agenthealth.yaml) |
| 3: CLI | [Commands, formats, doctor, OAuth login, and exit codes](cli.md), [installation](installation.md) | [Run the CLI](../README.md#run-the-cli), [example commands](../examples/README.md) |
| 4: HTTP/API | [Status, headers, bearer authentication, body checks, transport diagnostics, and safety](http-adapter.md), [HTTP configuration](../spec/configuration.md#http-response-expectations-phase-4-draft) | [HTTP readiness](../examples/core-check/agenthealth.yaml), complete configuration examples in the specification |
| 5: MCP | [HTTP/stdio, protocol selection, inventories, OAuth, functional probes, and limits](mcp-adapter.md), [login](cli.md#mcp-oauth-login) | [Discovery, stateless, stdio, functional, and all OAuth flows](../examples/README.md) |
| 6: A2A | [A2A 1.0 JSON-RPC with explicit 0.3.0 compatibility discovery, authentication, skills, and opt-in interactions](a2a-adapter.md) | [A2A v1 configuration](../examples/a2a-check/v1.yaml), [0.3.0 compatibility](../examples/a2a-check/agenthealth.yaml) |
| 7: Agent health (v0.5.0) | [Health resource, capabilities, safe tasks and runtime handler prerequisites](agent-adapter.md), [impact proposal](rfcs/phase-7-agent-health.md) | [Agent, peer and path example](../examples/agent-check/README.md) |
| 8: Agentgateway (v0.6.0) | [Gateway signals and backend/path evidence](agentgateway.md), [impact proposal](rfcs/phase-8-agentgateway.md) | [Gateway configuration](../examples/gateway-check/README.md) |
| 9: Agent Router (v0.7.0) | [Router signals and route/backend evidence](agent-router.md), [impact proposal](rfcs/phase-9-agent-router.md) | [Router configuration and setup](../examples/router-check/README.md) |
| 10: Dependency graph (v0.8.0) | [Shared nodes, edge policies and budgets](dependency-graph.md), [impact review](rfcs/phase-10-dependency-graph.md) | [Graph configuration](../examples/graph-check/README.md), [earlier-phase alignment](phase-10-alignment.md) |
| 11: AHP (experimental, v0.9.0) | [HTTP serving](ahp.md), [wire contract](../spec/protocol.md), [impact review](rfcs/phase-11-ahp.md) | [Serving example](../examples/ahp-check/README.md), [coverage and alignment](phase-11-alignment.md) |
| 12: Docker (v0.10.0) | [Docker distribution](docker.md), [container usage](installation.md#container-image) | [Dockerfile](../Dockerfile), [CI container job](../.github/workflows/ci.yml), [release publish job](../.github/workflows/release.yml) |
| 13: Standalone binaries (v0.11.0) | [Package installation](installation.md#package-channels-phase-13-v0110), [coverage and alignment](phase-13-alignment.md) | [Distribution example](../examples/distribution/README.md), [package builder](../scripts/package_release.py), [native smoke checks](../scripts/test_release_binary.py) |
| 14: Kubernetes (released chart 0.1.0) | [Deployment guide](kubernetes.md), [coverage and alignment](phase-14-alignment.md) | [Manifests](../deploy/kubernetes/), [Helm chart](../deploy/helm/agenthealth/), [scenarios](../examples/kubernetes/README.md), [tests](../tests/kubernetes/) |

Phase 7 is released in [v0.5.0](releases/v0.5.0.md): see the [agent adapter](agent-adapter.md) and
[agent/peer/path example](../examples/agent-check/agenthealth.yaml).

Phase 10 is included in [v0.8.0](releases/v0.8.0.md); older v0.7.0 binaries do not include it.
Experimental AHP HTTP serving is included in v0.9.0; see [AHP](ahp.md). [Phase 12 — Docker Distribution](../ROADMAP.md#phase-12--docker-distribution)
is released in v0.10.0; see [Docker](docker.md).

Configuration examples are templates requiring your services and credentials.
The documentation validator checks marked configuration/result blocks and
standalone health configuration examples and configuration embedded in
Kubernetes resources against the AHS schema; Helm overrides use the chart schema.
Kubernetes manifests and chart rendering also have dedicated tests. Validation
does not establish connectivity to the placeholder servers or prove every command example works.

Agent/multi-agent, gateway/router HTTP signals, HTTP/API, MCP, and A2A 1.0 JSON-RPC with explicit 0.3.0 compatibility are the available adapters. Other adapter types,
SDKs remain planned; Kubernetes integrations and chart 0.1.0 are released ([guide](kubernetes.md)); a [container image](installation.md#container-image) is
published starting in v0.10.0. Historical
release notes describe their tagged release. Use v0.2.0 or newer for Phase 5
MCP functionality and v0.3.0 or newer for Phase 6 A2A functionality.

Foundation stabilization in v0.4.0 adds current A2A, shared probes and release
hardening. See [interoperability](interoperability.md) and
[v0.4.0 migration notes](releases/v0.4.0.md).

Phase 8 Agentgateway integration is released in [v0.6.0](releases/v0.6.0.md). See [the gateway guide](../docs/agentgateway.md). Gateway checks use explicitly configured read-only HTTP signals and separate backend/path dependencies.

Phase 9 Agent Router HTTP integration is released in v0.7.0.
See [the router guide](agent-router.md). Named router signals, direct backends and configured paths retain separate evidence.

Phase 9 release and migration details: [v0.7.0 notes](releases/v0.7.0.md).

Phase 11 release and migration details: [v0.9.0 notes](releases/v0.9.0.md).

- [Kubernetes integration](kubernetes.md) — manifests, chart, probes, Secrets and deployment validation.
- [Phase 14 alignment](phase-14-alignment.md) — coverage, compatibility and validation limits.

## Release and package distribution

See the [synchronized distribution plan](distribution.md) for the agreed
v0.12.0 policy, Releases/Packages layout, availability and implementation
checklist. OCI chart/binary publishing is implemented in source; publication is pending.
