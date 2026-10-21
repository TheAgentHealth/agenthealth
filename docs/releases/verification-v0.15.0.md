# v0.15.0 publication verification

Verified on October 20, 2026 (America/Chicago). [GitHub Release](https://github.com/TheAgentHealth/agenthealth/releases/tag/v0.15.0),
[PyPI](https://pypi.org/project/agenthealth/0.15.0/) and
[npm](https://www.npmjs.com/package/@theagenthealth/agenthealth/v/0.15.0) are public.
The GitHub Release is a regular Latest release with 21 signed files.

## Source, contributor credit and gates

Immutable source tag `v0.15.0` points to
`7f1147928c434b79249df2dca9d7d8b5ff43f37a`, the merge commit of
[PR #52](https://github.com/TheAgentHealth/agenthealth/pull/52).
Its 13 focused commits include Sharath K's authored model-adapter contribution,
integrating [RFC #14](https://github.com/TheAgentHealth/agenthealth/issues/14)
and [PR #17](https://github.com/TheAgentHealth/agenthealth/pull/17).
[Sharath K (@sharath568)](https://github.com/sharath568) is acknowledged in the
README, phase alignment audit and release notes.

All 39 [PR CI checks](https://github.com/TheAgentHealth/agenthealth/actions/runs/39532877991)
passed before merge. The first Copilot review was addressed; no second review
was requested or awaited. The authorized approval bypass was used with green CI.
The successful [fresh tagged-source CI](https://github.com/TheAgentHealth/agenthealth/actions/runs/39534405077)
passed all required checks, including five native archives, Python/JavaScript SDK
matrices and Kubernetes 1.31.9, 1.32.2 and 1.33.1. Its path detector was
intentionally skipped for workflow dispatch; required Kubernetes jobs ran.

The [Release workflow](https://github.com/TheAgentHealth/agenthealth/actions/runs/39533632152)
completed successfully after infrastructure retries. An existing macOS Python
cancellation test initially observed a child PID immediately after group
termination; its unchanged retry and the fresh full gate passed. npm publication
and exact-byte verification succeeded before npm's install index propagated;
only the npm job and its dependents were rerun. The retry verified and reused the
identical published package. No tag, signed artifact or registry version was
moved, rebuilt for recovery or overwritten.

## Downloads, signatures and runtime

- Downloaded all 21 files anonymously from the public Release; asset digests and
  all 20 signed SHA-256 checksum entries matched.
- Verified all 21 GitHub build-provenance attestations and five archive-bound
  CycloneDX attestations against the exact source tag/commit and hosted release
  workflow. Verified signed OCI provenance for the image, chart and binary bundle.
- Inspected five archive contents, normalized ownership, version-specific
  installation guide, executable hashes and SBOMs. All use patched Go 1.27.3,
  fixing GO-2026-6809. The release includes four Linux packages and two channel
  manifests under the same shared version.
- The downloaded Linux AMD64 executable passed version/help/tool-error smoke
  checks, both model dialects, one-listing passive sharing, required-model and
  rate-limit failures, explicit bounded-request inference, schema validation
  and credential redaction against controlled servers.
- A fresh public PyPI installation passed 16 SDK integration tests against the
  downloaded CLI. Wheel/sdist public bytes and SHA-256 matched signed assets;
  both PyPI attestations verified the publishing repository.
- Public npm bytes and SHA-512 integrity matched the signed tarball. A fresh npm
  installation verified seven registry signatures and two package attestations.
  Installed-tarball ESM and TypeScript smoke checks passed against the downloaded CLI.
- The public container passed version/help, non-root/runtime configuration and
  HTTP-check smoke tests. AMD64 and ARM64 image binaries exactly matched their
  signed archive binaries; ARM64 was inspected without claiming local execution.

## Public containers and OCI packages

| Package | Immutable manifest digest |
|---|---|
| `GHCR Docker image` | `sha256:e648b3f169e502ab0adf914fa06c95607592941ddd2c8d538ed215e641eabc7c` |
| `Docker Hub mirror` | `sha256:b6f4212bc1d0e2bc27b690b13ac767dabc68dae510a4b439f4ee14f0a9c9bf3d` |
| `ghcr.io/theagenthealth/charts/agenthealth:0.15.0` | `sha256:13a7bae0bc01fec3ec30177be646428923a3acb8eac9e3fd0161437ec2243bc6` |
| `ghcr.io/theagenthealth/agenthealth-binaries:v0.15.0` | `sha256:4e240379141a48cf6def5f9f75a4c216cf0b712bdb29da621925eb290e7a56f2` |

Anonymous registry access verified source/version/platform metadata, OCI manifest
and config digests, and every chart/bundle layer. The binary bundle contains the
same 21 Release assets. GHCR and Docker Hub image runtime/layer identities match.

The downloaded chart matches version 0.15.0, appVersion v0.15.0 and matching image
pins. Local pre-merge Kubernetes smoke covered init outage/recovery, probes,
sidecars, Secrets, Jobs, CronJobs, Helm and readiness recovery. A final disposable
Kubernetes 1.32.2 check installed the published chart in serve, Job and CronJob
modes with the public v0.15.0 image; all three passed. Both disposable clusters
were deleted after validation.

## Limits

Controlled model HTTP servers verify dialect behavior, not hosted-provider
interoperability or inference quality. Sharath's Ollama 0.40 report remains
attributed evidence, not a rerun. Token enforcement depends on the remote server;
reported excess is detected after completion. Bedrock/Vertex signing,
Azure-specific authentication, streaming and embeddings remain planned.
Local executable runtime checks used Linux AMD64; other platforms rely on native
CI and archive inspection. AHP remains experimental. OS-native signing,
notarization, GPG Helm signatures and hosted package-channel repositories remain
outside this release. Historical releases and their records retain their scope.
