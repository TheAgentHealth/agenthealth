# Phase 8 — Agentgateway integration

The authorized Phase 8 implementation reuses the existing `gateway` type,
HTTP expectations and nested dependency results. No new topology or result
fields are introduced. The additive v1 configuration extension permits `http`
options on gateway targets, including nested gateway dependencies.

## Impact review

Earlier phases: extend configuration schema and Go validation, register the
GET-based gateway adapter in the CLI, reuse HTTP safety and diagnostics,
and validate independent backend execution. Existing HTTP HEAD behavior,
Phase 7 peer/path probes, status semantics, result schema and exit codes stay
compatible. Distribution packaging includes the adapter through the CLI.

Later phases: Phase 9 can reuse HTTP signals but must design router-specific
route evidence. Phase 10 must preserve named gateway, backend and path results
and independent critical/optional aggregation. AHP, SDKs, observability and
conformance must retain these distinctions and redaction. No automatic backend
inventory or topology discovery is authorized by a remote response.

## Evidence and limitations

See [the integration contract](../agentgateway.md). Gateway protocol checks
report the configured health signal; dependency results report independent
backends and configured communication paths. Aggregate status includes both.
A failing proxied request alone cannot identify whether the gateway or backend
caused it. No undocumented admin endpoint is parsed, and no model generation
or arbitrary tool invocation is performed by the gateway adapter.
