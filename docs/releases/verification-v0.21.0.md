# v0.21.0 publication verification

Release: [AgentHealth v0.21.0](https://github.com/TheAgentHealth/agenthealth/releases/tag/v0.21.0). Regular public Latest release;
source commit `ee6edbdc47731abe0cbffa8555dbaf90d081c389`. No published tag was moved or replaced.

All 21 required release assets were downloaded anonymously and checked
against published digests and the complete checksum manifest. Five platform
archives, matching CycloneDX SBOMs, four DEB/RPM packages, channel manifests,
Helm chart, Python wheel/sdist and npm tarball are present. Signed GitHub
provenance was verified against the exact source SHA/tag for every asset;
archive SBOM attestations were verified independently.

The Linux AMD64 archive was executed: version, HTTP health checks, canonical JSON,
JUnit/GitHub and all four observability projections passed. Other platform archive
contents/hashes were checked; CI provides their native runtime coverage.

Public GHCR Docker digest: `sha256:b8816f75e6a5806c1d7463cb19853dd264e19eaae5104452bc9751fa9ebdef67`.
Docker Hub mirror digest: `sha256:6ef9b73598c0f007a8027db13af58526f95ebce45221d9df73a1b5b656aa3a8d`. Image source/version and
platform identities match. Signed OCI provenance and anonymous layer access were
verified for the Docker image, OCI chart and OCI binary bundle. OCI chart/binary
contents match the signed release assets.

- `ghcr.io/theagenthealth/charts/agenthealth:0.21.0`: `sha256:3c92f08e4c3a1803b9c0bc54a7e29a0a0bc0485dbf8a3378f96fdf7cf40f62b9`
- `ghcr.io/theagenthealth/agenthealth-binaries:v0.21.0`: `sha256:2d6143165707bda548f3165a3b6efa0c5e6f047e1fae94e43ac30db526757d47`

PyPI wheel/sdist and npm registry tarball bytes, digests/integrity and provenance
match the signed release artifacts. See the release workflow for the exact-tag
pre-release CI gate and publication jobs.

Local validation covered Go vet/full race tests, process/CLI regression tests,
Python schema and SDK tests, JavaScript tests/types, all marked documentation
examples, Helm lint/static tests, native archive and container runtime smoke,
and disposable Kubernetes 1.33.1 deployment scenarios. The initial Copilot review's
two genuine findings were fixed and resolved; no second review was requested.

Phase 23 remains an operator-trusted local process extension. It does not provide
OS sandboxing, tenant isolation, automatic downloads, community certification or
real third-party adapter interoperability. See [alignment](../phase-23-alignment.md)
and [transport limitations](../rfcs/phase-23-adapters.md).

Public runtime follow-up: the downloaded chart and public image passed serve,
Job and CronJob modes on disposable Kubernetes 1.33.1. A fresh PyPI install
passed all 48 SDK/framework tests against the public CLI. The public npm install
passed its version/CLI check and npm 11 registry-signature/attestation audit.
PyPI publisher attestations for both wheel and sdist also verified.
