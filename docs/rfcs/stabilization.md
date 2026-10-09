# Stabilization impact review

The maintainer requested reconciliation of the security review and implementation
of the confirmed improvements. This document records the additive specification
proposal for review under CONTRIBUTING and GOVERNANCE; it does not imply external
approval or a new published release.

Public contract proposal: [RFC #46](https://github.com/TheAgentHealth/agenthealth/issues/46).

## Passive HTTP method

Add optional `http.method: HEAD | GET` to the v1 configuration contract. Preserve
HEAD when omitted for HTTP/API and GET for gateway/router. Functional body checks
always use GET. No silent fallback, redirect following, body-inspection opt-in
change, new target type, health state or result field is introduced.

Phase 4 needs Go validation, JSON Schema, valid/invalid fixtures, regression tests
and documentation. Phases 8/9 retain existing defaults and can select an explicit
passive method. Phases 10/11/14 use the same per-node option; no inherited method or
remote configuration submission is added. Existing released configs remain valid;
new method fields require the updated binary. Future SDKs must preserve this field
and old binaries must reject it rather than ignore it.

## Security, middleware and sequencing

Current security documentation must describe the released AHP implementation.
Reference application handlers implement the existing Phase 7 document/task
contract; they do not introduce a framework-independent proof of task health.
Framework probes remain explicit and application-authorized. Phase 20's full
framework integration remains distinct from reference middleware.

Stabilization and broader interoperability precede Phases 15–19 expansion.
Phases 23/26 retain responsibility for out-of-process adapter isolation and
production hardening. Do not mark Phase 25 complete based on controlled SDK tests
or static Kubernetes rendering. Historical releases retain their original scope.
