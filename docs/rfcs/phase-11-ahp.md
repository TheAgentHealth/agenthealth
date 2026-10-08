# Phase 11 AHP design and impact review

The user authorized Phase 11 implementation. This design records the experimental
HTTP v1 choices for maintainer review; it does not claim external standard approval.
The [wire contract](../../spec/protocol.md) defines the implemented scope.

## Earlier phases

Phases 1 and 7–10 retain their health, configuration and result contracts. AHP
wraps existing validated, redacted results in a separately versioned envelope.
Graph IDs, relationships and edge policies survive authenticated exchange.
Readiness accepts optional degradation; completed functional/path checks are
independent from declared capability and endpoint health. No adapters change.
Phases 2–3 gain a snapshot server and serve command. Existing command exit codes,
configuration schemas and output shapes remain compatible. Tests cover public
redaction, authorization, cold/stale snapshots, liveness and method handling.

## Later phases

Phase 12 containers and Phase 14 Kubernetes can run serve and query `/live` and
`/ready`; container packaging and deployment examples remain their deliverables.
Phases 15–16 SDKs must distinguish AHP envelopes from CLI results and enforce
version, freshness and authorization. Phase 20 framework integrations may expose
the same contract. Phase 22 exporters must preserve separate dimensions and avoid
public topology. Phase 24 independent conformance must use the envelope schema
and test status mapping, auth, version errors, staleness and graph evidence.
Phase 25 interoperability must test another implementation; local tests do not
claim this. Phase 26 owns proxy/load testing, TLS deployment and finer access
controls. No changes to database/model adapters or binary packaging are needed.

## Limitations

HTTP JSON only, one configuration per process, one bearer scope, serial refresh,
no automatic discovery, no per-client limiter or built-in TLS. Detailed capability
and dependency views intentionally share the recursive evidence document so
relationship context is retained. Independent interoperability remains unverified.
