# Phase 9 — Agent Router integration

Specification discussion: [RFC #23](https://github.com/TheAgentHealth/agenthealth/issues/23).

The user-authorized Phase 9 implementation adds the `router` target vocabulary
and permits existing HTTP options on top-level and nested router targets. It
uses read-only GET signals and explicit named route/backend dependencies.
See the [integration contract](../agent-router.md).

## Earlier phase impact

Phase 1: additive v1 configuration vocabulary, configuration/result type schemas and adapter applicability
updates with valid/invalid fixtures. Health states, dimensions, result structure
and exit codes retain their contracts. Older binaries reject `router`; existing
configurations remain compatible. No topology fields are introduced.
Phase 2: extend Go validation and reuse independent dependency execution,
critical/optional aggregation, bounded HTTP probes and transport safety.
Phase 3: register the adapter for ping, check and doctor in all output formats.
Phases 4–8: preserve HTTP HEAD behavior, gateway GET behavior, MCP/A2A contracts
and Phase 7 first-runtime probes. Binary packaging includes CLI registration.

## Later phase impact

Phase 10 must preserve named router signals, direct backends and individual
route results, including multiple routes to a backend; shared identity and graph
scheduling remain future design. Phase 11 AHP, SDKs and observability must retain
these distinctions and credential/topology redaction. Phases 24–26 must test
independent failures, safe opt-in, deadlines and actual product interoperability.
No remote response authorizes topology discovery or arbitrary active probes.

## Compatibility and scope

The generic interface avoids guessing a router product or admin wire contract.
Product-specific discovery is unsupported. Configured route availability uses
existing HTTP/MCP/A2A protocol evidence; it does not prove membership, failover
or communication from an agent runtime. This additive implementation is not a
release; publication and maintainer specification review remain separate steps.
