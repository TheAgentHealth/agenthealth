# AgentHealth Specification

This directory contains the specification documents for AgentHealth, introduced in the main [README](../README.md#specification--protocol--implementation).

| Layer | Document | Status |
|---|---|---|
| Agent Health Specification (AHS) | [health-model.md](health-model.md) | Draft — most mature layer |
| Target Model | [target-model.md](target-model.md) | Draft |
| Agent Health Protocol (AHP) | [protocol.md](protocol.md) | Proposed / experimental |
| Configuration | [configuration.md](configuration.md) | Draft |
| Result Schema | [result-schema.md](result-schema.md) | Draft |
| Exit Codes | [exit-codes.md](exit-codes.md) | Draft |
| Adapter Contract | [adapter-spec.md](adapter-spec.md) | Draft |
| Security Requirements | [security.md](security.md) | Draft |

## Relationship between documents

```text
health-model.md   -> defines WHAT health means (states, dimensions)
target-model.md   -> defines target types and per-type applicable dimensions
result-schema.md  -> defines how a single check result is represented
configuration.md  -> defines how targets/checks/dependencies are declared
exit-codes.md     -> defines the CLI exit code contract for automation
adapter-spec.md   -> defines the contract a technology-specific adapter must satisfy
security.md       -> defines security requirements any conforming implementation must satisfy
protocol.md       -> (experimental) defines HOW health is exposed/exchanged over the wire
```

## Versioning

Specification documents use a `spec_version` field (e.g. `v1`) independent of the AgentHealth software release version. Breaking changes to any document here require a version bump and a changelog entry in this directory.

## Proposing changes

Specification changes are more sensitive than implementation changes because downstream adapters, SDKs, and (eventually) AHP conformance tests depend on them. See [GOVERNANCE.md](../GOVERNANCE.md) for the RFC-style process used for spec changes, and [CONTRIBUTING.md](../CONTRIBUTING.md) for the general contribution workflow.
