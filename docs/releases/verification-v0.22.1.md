# v0.22.1 publication verification

Release: [AgentHealth v0.22.1](https://github.com/TheAgentHealth/agenthealth/releases/tag/v0.22.1). Regular public Latest release;
source commit `9e62542762c371fdf05d95e9b275e09d46047898`. Public verification completed 2026-11-13
(America/Chicago). Existing v0.22.0 tags and artifacts were preserved.

## Release gates and signed downloads

The [Release workflow](https://github.com/TheAgentHealth/agenthealth/actions/runs/43525793509)
passed all required publication jobs. Its [fresh exact-tag CI](https://github.com/TheAgentHealth/agenthealth/actions/runs/43525820680)
passed the required source, conformance, five-platform native archive, SDK and
Kubernetes gates. Missing/skipped mandatory jobs were not treated as passes.

All 21 public assets were downloaded anonymously and matched their release
SHA-256 digests and the complete checksum manifest: five platform archives,
five CycloneDX SBOMs, four DEB/RPM packages, two channel manifests, Helm chart,
Python wheel/source archive, npm tarball and checksums.

GitHub cryptographic build provenance was independently verified for every
asset against the exact source SHA, `refs/tags/v0.22.1`, repository release
workflow and GitHub-hosted runner policy. All five archive CycloneDX
attestations were verified separately. Every SBOM records Go 1.27.4.

## Public registries

| Package | Published digest |
|---|---|
| GHCR Docker image | `sha256:727c21fd5e547d53e88a0761c0c663e7bf6b2706326b58171abd046610701323` |
| Docker Hub mirror | `sha256:3b0804ad350a0585df68ef1d79baf7defd18b0c77b574340c93da691d810ddcc` |
| Helm OCI chart | `sha256:1bd0de74199a87671825d246b4dd70b3e7f24b45d08c62e0de2ae2fccb1aebe2` |
| OCI binary bundle | `sha256:b2779501258b3944be1238cb64e32ec198d942082fb153c8a62d46a754e4c782` |

GHCR anonymous manifest/config access established Linux AMD64/ARM64, the exact
source revision and software version. Chart and binary-bundle manifests,
metadata and every asset blob were downloaded anonymously and verified against
signed release hashes. All three GHCR package provenance attestations were
independently verified against the immutable source tag and workflow.
Anonymous Docker Hub manifests confirmed both platforms and filesystem/runtime equality with GHCR.

PyPI wheel/source bytes and npm tarball bytes match the signed GitHub assets;
registry digests/integrity match. Both PyPI trusted-publisher attestations were
verified against the publishing repository. npm registry signature/provenance
audit passed with npm 11. The public SDKs and CLI all identify 0.22.1.
GitHub Latest, PyPI, npm
`latest`, and both images' `latest`/`v0` aliases resolve to this release.

## Runtime evidence and limits

The public Linux AMD64 archive passed version/help/tool-error smoke and the
positive/negative conformance examples: valid evidence exits 0; inconsistent
status exits 6. The public container passed version, non-root/runtime metadata
and real HTTP health smoke; its native executable matches the signed archive.
Fresh public Python/npm installations passed SDK checks against that CLI.

The downloaded chart and public image passed disposable Kubernetes 1.32.2
runtime scenarios, including init outage/recovery, probes, sidecar, Secrets,
Jobs, CronJobs, chart modes and readiness recovery. The disposable cluster was
removed. A Helm release was also installed from the downloaded chart in job
mode and uninstalled. Other operating systems/architectures rely on exact-tag CI;
they were not
executed locally during this follow-up. DEB/RPM assets were verified as downloads;
no local package-manager installation was performed. OS-native executable
signing/notarization remains planned.

Go 1.27.4 replaces the affected Go 1.27.3 build toolchain for GO-2026-6859;
it does not modify v0.22.0 binaries. This is a patched reference release,
not a security audit or independent product certification. Phase 24 remains
bounded offline AHS result validation plus reference behavioral conformance.
