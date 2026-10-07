# Documentation guide

Phases 0–5 are implemented on the Phase 5 branch. The specification remains
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

Configuration examples are templates requiring your services and credentials.
The documentation validator checks marked configuration/result blocks and
all standalone example YAML files against the schemas. It does not establish
connectivity to the placeholder servers or prove every command example works.

HTTP/API and MCP are the available adapters. A2A and other adapter types,
SDKs, container images, and Kubernetes integrations remain planned. Historical
release notes describe the tagged release, which may precede Phase 5; build
this branch to use changes that have not yet been released.
