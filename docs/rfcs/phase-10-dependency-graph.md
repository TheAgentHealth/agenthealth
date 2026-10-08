# Phase 10: explicit dependency graph

The implementation requested for Phase 10 extends v1 additively. Existing
nested targets, health states, exit codes, checks and adapter contracts remain
valid. This document records the specification discussion and impact review
for maintainer review with the implementation. The public specification discussion is [RFC #26](https://github.com/TheAgentHealth/agenthealth/issues/26).

## Contract

Targets may declare a globally unique `id` (1–64 ASCII letters, digits, dots,
underscores or hyphens, starting with a letter or digit). Dependency entries
may use `ref` to select that ID anywhere in the explicit configuration.
A reference accepts only `ref`, `critical`, and `relationship`; it cannot
override a node's checks, credentials, endpoint, budget, or dependencies.
Inline dependencies retain their existing fields and may also declare IDs.
No endpoints or topology are fetched from advertised dependency names.

Edges may declare `relationship`: `supporting`, `downstream`, `path`, `gateway`,
or `router`. Omission preserves the legacy unlabeled dependency behavior.
`downstream` denotes composite agents. A `path` selects a separately configured
probe that supplies communication evidence, such as an explicitly safe first
agent task. An endpoint check alone cannot establish first-to-peer communication.
Shared backends and multiple paths have separate IDs and evidence.

Each explicit ID executes once per run, including its active operation and
cleanup. Anonymous inline targets execute independently. Dependency traversal
and adapter checks run in parallel, retain declaration order in output, and
continue independently after parent failure. The whole-run limit remains
60 seconds. `concurrency` defaults to 16 and accepts 1–16; it limits adapter
calls, including cleanup, per run under the existing engine-wide 16-call bound.
An uncooperative adapter retains both slots until it actually returns.

`budget_ms` accepts 1–60000 and bounds all of a node's own checks, retries and
queue waits. Dependencies have independent budgets; a short parent budget
does not cancel a shared backend. Existing per-attempt deadlines still apply.
Cycles, duplicate IDs, missing references, depth over 64, and over 10000 tree
projections fail configuration before any adapter execution.

Results retain the v1 recursive shape. `target.id` identifies shared evidence;
dependency entries carry `critical` and optional `relationship`. Repeated tree
projections contain the same node evidence but different edge policies.
Node checks retain own evidence while `status` includes dependency aggregation.
Human output labels relationship edges; JSON and YAML carry equivalent fields.
Optional failures contribute DEGRADED and critical failures preserve AHS rules.

## Impact review

Earlier phases: extend Phase 1 configuration/result schemas and fixtures,
Phase 2 scheduling, validation, cleanup bounds and propagation, and Phase 3
existing human/JSON/YAML output. Phases 7–9 adapters retain their protocols;
agent advertised names match only explicitly resolved configured dependencies.
HTTP/API, MCP and A2A protocols, credential schemes, request safety and functional opt-in remain unchanged; their guides explain shared graph lifecycle. Existing nested examples remain valid. Distribution commands and exit codes require no change. See the [phase-by-phase alignment audit](../phase-10-alignment.md) for affected documents and no-impact rationales.

Later phases: AHP must exchange node identity and edge relationship/policy,
redact topology, and avoid treating repeated projections as repeated probes.
SDKs must preserve IDs and edge metadata. Observability exports and conformance
must distinguish node failures, communication paths and shared-node evidence.
Production hardening must exercise large DAGs, bounded projections, cancellation,
and uncooperative adapter isolation. These remain future-phase deliverables.

Compatibility: additive v1 configuration/result fields require no migration
for existing files. Consumers with closed local result schemas should update
their schema. Parallel execution changes timing, never declaration-order output.
No AHP wire binding or automatic topology discovery is introduced.

Cleanup hooks remain serialized across nodes, runs and engines sharing a registry. Waiting for the serialization gate honors the cleanup deadline; an uncooperative hook holds the gate and its adapter-call slots until it actually returns.
