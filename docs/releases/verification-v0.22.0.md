# v0.22.0 publication verification

Release: [AgentHealth v0.22.0](https://github.com/TheAgentHealth/agenthealth/releases/tag/v0.22.0),
published November 9, 2026 as the regular Latest release. Source commit:
`d7d85027606baad31b83a5614a1c7e0314643bf5`. Follow-up inspection: November 13, 2026.
The published tag and artifacts were not changed by this documentation update.

## Publication and CI evidence

The [Release workflow](https://github.com/TheAgentHealth/agenthealth/actions/runs/42685435523)
passed its CI gate, validation, release assets, container, OCI packages, npm,
PyPI and final publication jobs, including its Docker Hub deployment record.
Recovery jobs were skipped because ordinary publication succeeded.
The [exact-source pre-release CI](https://github.com/TheAgentHealth/agenthealth/actions/runs/42685458393)
also passed. This records the workflow results; it does not substitute for
independent verification of every signature or platform at follow-up.

## Public downloads and registries

All 21 GitHub release assets were downloaded anonymously and their SHA-256
digests checked against the release metadata. Every entry in `checksums.txt`
also matched its downloaded file. The release contains five platform archives,
five CycloneDX SBOMs, four DEB/RPM packages, Homebrew/Scoop manifests, Helm chart,
Python wheel and source archive, npm tarball and checksum manifest.

PyPI 0.22.0 wheel/source downloads were byte-identical to the release assets.
The npm 0.22.0 tarball was byte-identical to its release asset and matched
registry SHA-512 integrity. npm's `latest` tag resolved to 0.22.0.

| Package | Published digest |
|---|---|
| `ghcr.io/theagenthealth/agenthealth:v0.22.0` | `sha256:88d69d66515e1a428967400c313950343be2dcb93137e41662f441d4f2b25975` |
| `docker.io/theagenthealth/agenthealth:v0.22.0` | `sha256:4f2c9bd17bd1c2d631ad26becdfd676601e43c955b81167fc8b8a9c7da2f52a1` |
| `ghcr.io/theagenthealth/charts/agenthealth:0.22.0` | `sha256:b2bec01e28b2ec0df51082c64dd745de9b1615c4e5148a9916f245ee6584eb5e` |
| `ghcr.io/theagenthealth/agenthealth-binaries:v0.22.0` | `sha256:3d7781a35771b3653ad8c7c1b856c2dcabe16dedafac131887be555d029474e1` |

Anonymous GHCR manifest/config access confirmed Linux AMD64/ARM64 images,
version v0.22.0 and the exact source revision. All Helm and binary-bundle
layers were downloaded anonymously and matched their digests; the chart layer
and bundled release files match the GitHub asset hashes.
Docker Hub's public tag API confirmed the recorded digest and both platforms.
Direct anonymous Docker Hub manifest retrieval was rate-limited (HTTP 429), so
that registry's blobs were not independently downloaded during follow-up.

## Runtime checks and verification limits

The published Linux AMD64 archive executed successfully: `version` reported
v0.22.0; `conformance test` accepted `healthy.json` and `dependency.json` with
exit 0, and rejected `inconsistent.json` with `inconsistent_status` and exit 6.
Other platforms were checked through downloadable hashes and pre-release CI,
not executed locally during follow-up. Container runtime, installed SDK,
DEB/RPM installation and deployed Kubernetes checks were not repeated locally.
Cryptographic provenance, SBOM attestations and registry signatures were not
independently reverified during this follow-up; checksum equality is not a
replacement for signature verification. Use the [installation verification
instructions](../installation.md#verify-provenance-and-sboms).

The packaged SDK READMEs and PyPI description still refer to v0.21.0 although
package metadata and runtime version constants are 0.22.0. Current source
instructions are corrected separately; existing immutable package bytes retain
the older prose. A future release must carry those corrections.

## Security follow-up

The [November 13 scheduled CI run](https://github.com/TheAgentHealth/agenthealth/actions/runs/43445872962)
passed Reference conformance but failed the Go vulnerability scan. It reported
GO-2026-6859 in `net/http` at Go 1.27.3, with Go 1.27.4 identified as the fix.
The released Linux AMD64 SBOM records Go 1.27.3. Publication availability does
not imply that this later security finding has been resolved. Current source advances CI/release builds and the pinned Docker builder to
Go 1.27.4. Publish a new synchronized patch release after validating that source;
do not replace v0.22.0 or move its tag.

Phase 24 provides bounded offline result validation and reference behavioral
evidence. Certification, independent product interoperability and remaining
Phase 26 hardening are outside that baseline.

Remediation follow-up: [v0.22.1](verification-v0.22.1.md) ships the Go 1.27.4
build and corrected SDK descriptions; the v0.22.0 artifacts above remain unchanged.
