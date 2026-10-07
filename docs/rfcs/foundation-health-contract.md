# Foundation health contract

Status: Implemented in [PR #9](https://github.com/TheAgentHealth/agenthealth/pull/9)
and released in [v0.4.0](../releases/v0.4.0.md).
[RFC issue #8](https://github.com/TheAgentHealth/agenthealth/issues/8) is closed.

## Problem and decision

MCP and A2A classify missing required capabilities as DEGRADED and incompatible
server versions as MISCONFIGURED, contradicting AHS. Align adapters with AHS:
invalid local pins remain MISCONFIGURED through configuration validation;
valid configuration meeting an incompatible server is UNHEALTHY; missing
explicitly required functionality is UNHEALTHY. Unsupported server transports
and required protocol extensions are protocol incompatibility failures.
Latency degradation and optional dependency aggregation retain DEGRADED.
There is no new optional capability expectation setting in this change.

Add an optional check-entry `code` containing a stable engine-owned diagnostic
identifier. Preserve existing protocol-prefixed identifiers and suppress unknown
adapter codes, including codes embedded directly in CheckResult. Codes describe
the reason; status continues to determine exit codes. This additive extension
keeps spec_version v1; consumers must tolerate new optional fields.

## Phase impact review

Earlier phases: update AHS clarification, result schema, Go model/output
validation, engine sanitization, MCP/A2A classification, documentation and
fixtures. CLI JSON/YAML receive the additive field without new flags.

Later phases: agent checks, graph diagnostics, AHP, SDKs, adapter development
and conformance must consume status and code without parsing messages. No new
agent roles, topology fields or adapter lifecycle are introduced.

Sequencing: complete these fixes before agent-health implementation. Current
A2A support, external interoperability, measured probe reuse and packaging
hardening remain separate follow-ups; no release is authorized by this RFC.

## Validation and migration

Regression tests cover incompatible servers, missing required inventory and
capabilities, code serialization, schema acceptance/rejection and suppression
of arbitrary adapter codes. Required-capability failures now return exit 2
instead of 1, and incompatible server versions return exit 2 instead of 4.

## v0.4.0 completion scope

The maintainer requested the full foundation milestone and v0.4.0 publication.
This extends the proposal to default A2A 1.0 JSON-RPC with explicit 0.3.0 mode,
shared per-target resources through an optional cleanup hook, real official SDK
checks, scanning and release supply-chain evidence. Local pins still fail before
networking; discovered endpoints must retain the trusted origin. The A2A default
change ships in the pre-1.0 minor release with explicit migration notes. The wire
envelope remains v1 and schema options are extended, without removing old pins.
The initial “no release authorized” note described the earlier request; publication
is now authorized after validation and review under the repository workflow.
