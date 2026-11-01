# v0.19.0 publication verification

Verified November 1, 2026 (America/Chicago). The regular Latest
[GitHub Release](https://github.com/TheAgentHealth/agenthealth/releases/tag/v0.19.0),
[PyPI](https://pypi.org/project/agenthealth/0.19.0/) and
[npm](https://www.npmjs.com/package/@theagenthealth/agenthealth/v/0.19.0) are public.

## Source and gates

Immutable tag `v0.19.0` identifies `bd65d3a0017ef26ea625f6b0d4f331996fe680b5`, the merge of
[PR #63](https://github.com/TheAgentHealth/agenthealth/pull/63), preserving 22 focused commits.
All 45 [PR CI jobs](https://github.com/TheAgentHealth/agenthealth/actions/runs/41276974690)
passed before the explicitly authorized approval-rule bypass. All four genuine
first Copilot findings were fixed, regression-tested and resolved: old-tag CI
recovery, shell configuration/report overlap, symlink/hard-link cleanup and
fail-closed report I/O. No second review was requested or awaited.

[Fresh tagged-source CI](https://github.com/TheAgentHealth/agenthealth/actions/runs/41277689223)
passed the required tests, including all five native archive platforms, both SDK
matrices, all three Action platforms and Kubernetes 1.31.9/1.32.2/1.33.1.
The PR-only path detector is intentionally skipped in release CI.
The [Release workflow](https://github.com/TheAgentHealth/agenthealth/actions/runs/41277682025)
completed synchronized publication from that immutable source.

## Downloads and runtime evidence

- All 21 release files matched GitHub digests and all 20 checksum entries.
- All 21 build-provenance attestations bind the source tag/commit. Five signed
  CycloneDX attestations and all archive executable/SBOM hashes verified.
- Downloaded Linux AMD64 CLI reports v0.19.0 and passes real HTTP checks in
  JSON, JUnit and GitHub formats. Other native platforms were exercised by CI.
- Anonymous GHCR image pull passed version/help, non-root configuration and
  real HTTP runtime smoke. Both architectures bind source/version; Docker Hub
  and GHCR runtime/filesystem identities match.
- Public chart/image passed serve, Job and CronJob modes on disposable
  Kubernetes 1.33.1. Source smoke also passed outage/recovery, probes, sidecars,
  Secrets and failed-Job scenarios.
- Public PyPI wheel/sdist match signed release bytes; both PyPI attestations
  verified. Fresh installed SDK/framework tests passed against the public CLI.
- Public npm bytes and SHA-512 integrity match the signed tarball; fresh public
  installation, signature auditing and ESM/TypeScript/CLI smoke passed.
- OCI chart and binary-bundle signatures, source/version metadata, anonymous
  manifests/layers and every bundled file matched signed release assets.

## Immutable registry identities

| Package | Verified digest |
|---|---|
| GHCR Docker `ghcr.io/theagenthealth/agenthealth:v0.19.0` | `sha256:dcdaf7daf2033d284367ed9107d1cdd2473ed2f18caa911e827a88092be1e4a5` |
| Docker Hub `theagenthealth/agenthealth:v0.19.0` | `sha256:7efa3b839f113d26a5f025044125605d941409e5a80d1faa14b8898e757e89d3` |
| Helm `ghcr.io/theagenthealth/charts/agenthealth:0.19.0` | `sha256:8408939037ea427ff0d03d801c983700e61536df3c70dcfcdb1bb7d3bccd41c1` |
| OCI binaries `ghcr.io/theagenthealth/agenthealth-binaries:v0.19.0` | `sha256:e1afd559f93c3c175fc5c4ff135b8d7a92e4c2df43927f6b5a19483efcfa701a` |

Hosted GitLab/Jenkins/Azure execution matrices, broader product compatibility,
SARIF, native executable signing and GPG Helm `.prov` files remain deferred.
This verification does not establish tenant isolation or conformance certification.
