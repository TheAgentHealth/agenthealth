# v0.18.0 publication verification

Verified October 31, 2026 (America/Chicago). The regular Latest
[GitHub Release](https://github.com/TheAgentHealth/agenthealth/releases/tag/v0.18.0),
[PyPI](https://pypi.org/project/agenthealth/0.18.0/) and
[npm](https://www.npmjs.com/package/@theagenthealth/agenthealth/v/0.18.0) are public.

## Source and gates

Immutable tag `v0.18.0` points to `03c48df9ee5926e296d5f478ee973e5470029558`, the merge of
[PR #61](https://github.com/TheAgentHealth/agenthealth/pull/61), with 14 focused commits.
All 42 [PR CI jobs](https://github.com/TheAgentHealth/agenthealth/actions/runs/41178377707)
passed before the user-authorized approval bypass. The first Copilot review was
assessed: optional-extra installation coverage, async peer callback validation,
and conflicting documentation were corrected. No second review was requested
or awaited.

[Fresh tagged-source CI](https://github.com/TheAgentHealth/agenthealth/actions/runs/41178856669)
passed all 41 test jobs, including native archives on five platforms, both SDK
matrices, controlled framework tests and Kubernetes 1.31.9/1.32.2/1.33.1.
Only the PR-only path detector was intentionally skipped.
The [Release workflow](https://github.com/TheAgentHealth/agenthealth/actions/runs/41178847619)
succeeded on attempt 2. npm per-version bytes/provenance were available before
its install index caught up. After confirming public install visibility, only
the failed npm job and dependents were retried. The immutable tag, signed assets
and published package bytes were reused without rebuilding or replacement.

## Downloads and runtime evidence

- All 21 downloads matched GitHub digests and the complete 20-entry checksum set.
- All 21 build-provenance attestations bind the exact source tag/commit; five
  CycloneDX attestations and archive executable/SBOM hashes verified.
- Downloaded Linux AMD64 CLI reports v0.18.0 and returns HEALTHY for the live
  dedicated LangGraph example. Other native platforms were tested by CI.
- Public PyPI wheel/sdist bytes match signed assets and both PyPI attestations
  verified. Fresh installation of `agenthealth[langgraph]==0.18.0` passed dependency
  validation, 25 framework tests, and 23 SDK tests against the downloaded CLI.
- Public npm bytes and SHA-512 integrity match the signed tarball. Fresh npm
  installation passed registry signature/attestation auditing and installed
  ESM/TypeScript/CLI smoke checks.
- Anonymous GHCR image pull passed version/help, non-root configuration and a
  real HTTP health check. Both architectures bind the source/version; GHCR and
  Docker Hub filesystem/runtime identities match.
- Downloaded public chart and public image passed serve, Job and CronJob modes
  on disposable Kubernetes 1.33.1. Local source smoke also covered init outage/
  recovery, probes, sidecars, Secrets and readiness recovery. Clusters were deleted.

## Public manifests

| Package | Immutable manifest digest |
|---|---|
| GHCR image | `sha256:a96cb4580fdd121ee704fadaffc334392fefe866bf0f06b80ce598eff3e02c0f` |
| Docker Hub mirror | `sha256:7db0da5402235b582d8481ad23b071add4d613ab0d756410114e6e6fabd77291` |
| `ghcr.io/theagenthealth/charts/agenthealth:0.18.0` | `sha256:991d57ab0203ee53a909285c5bf21beca486afe7a8b021424ee65ca250898130` |
| `ghcr.io/theagenthealth/agenthealth-binaries:v0.18.0` | `sha256:b38e16b9f07f24bf04ace5a639d6bf6db7f0c441e3649bc54be4525fd08388e3` |

Anonymous verification checked every chart/bundle manifest and blob, source/
version annotations and both image platforms. The OCI bundle contains the same
21 Release files. Chart version is 0.18.0; appVersion/image are v0.18.0.

## Scope

The supported baseline is LangGraph 1.2.14/FastAPI 0.143.0, dedicated safe graphs,
allowlisted async peer callbacks and an SDK consumer node. Callbacks remain
trusted application code; authentication, ingress limits and cooperative
cancellation require application integration. Other framework products and native
JavaScript plugins remain planned. This does not complete Phase 25 or claim
product certification, a security audit or tenant isolation. See the
[runtime guide](../framework-integrations.md) and [alignment](../phase-20-alignment.md).
