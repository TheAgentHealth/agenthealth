# Phase 7 agent health contract

Status: accepted draft interface, merged in [PR #19](https://github.com/TheAgentHealth/agenthealth/pull/19)
and released in [v0.5.0](../releases/v0.5.0.md). Review discussion: [RFC issue #18](https://github.com/TheAgentHealth/agenthealth/issues/18).
This document records the RFC-style proposal and compatibility/impact
assessment. The maintainer authorized merging and releasing the reviewed implementation.

The problem is that healthy supporting dependencies or a healthy A2A peer do
not establish that the user's first agent accepts work, or that it can contact
that peer. Introduce the [agent adapter interface](../agent-adapter.md) for
`agent` and `multi-agent`: passive metadata/liveness/readiness, named capability
and dependency discovery, and explicitly safe bounded task/path probes.

Retain existing health states/dimensions: protocol carries metadata and
liveness/readiness, capability carries requirements and configured discovery,
functional carries completed task/path evidence, and dependency aggregation
retains its existing critical/optional semantics. Preserve the recursive result
shape; separate named agent, peer, path and supporting dependency results.
Do not add topology, role, route or graph scheduling fields before Phase 10.

Earlier phases: extend configuration schema, strict Go validation, compiled-in
adapter registration and CLI documentation; reuse HTTP safety/reachability,
A2A peer checks, engine per-target state and dependency execution. Add stable
agent diagnostics, schema fixtures, adapter/CLI regression tests and examples.
No distribution format or exit-code changes are needed. Existing configurations
continue to work. Previously unsupported agent targets now execute health
checks; software ships this additive v1 draft extension in a future minor
release. Publication is documented in the v0.5.0 release notes.

Later phases: gateway/router adapters must preserve independent backend/path
evidence; the graph phase must consume configured dependency IDs without
assuming a runtime advertisement authorizes networking. AHP/SDK/framework and
conformance work must carry runtime-declared readiness separately from active
probe evidence and verify that path handlers actually contact selected peers.
Do not infer backend health from a successful path, or a working path from a
healthy backend. Arbitrary framework APIs and gateway/router instrumentation
remain future integration work.

Safety: names-only discovery prevents remote configuration injection. All
probes use the supplied TLS-verifying, redirect-refusing client and deadlines;
responses and outgoing task text are bounded. Functional probes run once,
with opt-in and a safe-handler contract, and classify incomplete completion as
UNKNOWN. They trust the participating runtime, as other remote protocol
observations do; independent tracing is outside this phase.
