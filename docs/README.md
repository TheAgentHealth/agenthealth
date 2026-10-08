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
| 8: Agentgateway (source) | [Gateway signals and backend/path evidence](agentgateway.md), [impact proposal](rfcs/phase-8-agentgateway.md) | [Gateway configuration](../examples/gateway-check/README.md) |
| 7: Agent health (v0.5.0) | [Health resource, capabilities, safe tasks and runtime handler prerequisites](agent-adapter.md), [impact proposal](rfcs/phase-7-agent-health.md) | [Agent, peer and path example](../examples/agent-check/README.md) |

Phase 7 is released in [v0.5.0](releases/v0.5.0.md): see the [agent adapter](agent-adapter.md) and
[agent/peer/path example](../examples/agent-check/agenthealth.yaml).

The next planned phases are:

- [Phase 9 — Agent Router Integration](../ROADMAP.md#phase-9--agent-router-integration): independent router health and route/backend diagnostics.
- [Phase 10 — Dependency Graph](../ROADMAP.md#phase-10--dependency-graph): graph execution and failure propagation.
- [Phase 11 — Agent Health Protocol (AHP)](../ROADMAP.md#phase-11--agent-health-protocol-ahp): standardized health exchange.

These phases are not yet implemented; their roadmap entries define planned scope.

Configuration examples are templates requiring your services and credentials.
The documentation validator checks marked configuration/result blocks and
all standalone example YAML files against the schemas. It does not establish
connectivity to the placeholder servers or prove every command example works.

Agent/multi-agent, gateway HTTP signals, HTTP/API, MCP, and A2A 1.0 JSON-RPC with explicit 0.3.0 compatibility are the available adapters. Other adapter types,
SDKs, container images, and Kubernetes integrations remain planned. Historical
release notes describe their tagged release. Use v0.2.0 or newer for Phase 5
MCP functionality and v0.3.0 or newer for Phase 6 A2A functionality.

Foundation stabilization in v0.4.0 adds current A2A, shared probes and release
hardening. See [interoperability](interoperability.md) and
[v0.4.0 migration notes](releases/v0.4.0.md).

Phase 8 Agentgateway integration is implemented in source. See [the gateway guide](../docs/agentgateway.md). Gateway checks use explicitly configured read-only HTTP signals and separate backend/path dependencies.
