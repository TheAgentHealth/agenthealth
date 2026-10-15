# v0.14.0 publication verification

Verified on 2026-10-15. [GitHub Release](https://github.com/TheAgentHealth/agenthealth/releases/tag/v0.14.0),
[PyPI](https://pypi.org/project/agenthealth/0.14.0/) and
[npm](https://www.npmjs.com/package/@theagenthealth/agenthealth/v/0.14.0) are public.

## Source and publication

The immutable source tag is `v0.14.0`, commit
`4993f41db99c340ad383c848e9ce86cd0944c3a1`.
[The source release build](https://github.com/TheAgentHealth/agenthealth/actions/runs/38692893491)
produced the 21 signed assets. All assets were retained unchanged during
[infrastructure recovery](https://github.com/TheAgentHealth/agenthealth/actions/runs/38695570652),
which completed successfully on its second attempt. The
[tagged-source CI](https://github.com/TheAgentHealth/agenthealth/actions/runs/38695583275)
passed, including native binary, Kubernetes and six JavaScript SDK matrix jobs.

The initial unpublished tag was recreated with explicit maintainer authorization.
The final source tag has not moved after publication. Recovery corrected npm
workflow identity and publication invocation; it did not rebuild release payloads.
npm trusted publishing uses `release.yml`, environment `npm`, and OIDC without
an npm token. Its registry provenance identifies infrastructure commit
`8231d8aa337c8da2ff6ec7ae04d4a90d90524c9b` on main; the published tarball exactly
matches the signed artifact built from the source tag above. npm accepted the
package before its registry metadata propagated; verification was rerun after
propagation and skipped publication of the existing identical package. Future
verification now waits up to ten minutes for registry propagation.

PR #50 was merged after green CI and the first Copilot review. All three valid
findings were addressed; no second review was requested. Subsequent infrastructure
and verification documentation changes were pushed directly to main as authorized.

## Download integrity, contents and runtime

- All 21 assets downloaded anonymously from the public release and matched the
  previously verified staging bytes and GitHub asset digests.
- All 20 SHA-256 entries in signed `checksums.txt` matched.
- All 21 build provenance attestations and five CycloneDX attestations verified
  against the exact source tag/commit and hosted release workflow.
- Five platform archives passed safe-content, license, permissions, platform-header,
  executable/SBOM hash and Go 1.27.2 toolchain checks. Four DEB/RPM packages matched
  their archive binaries, platform/version metadata and certificate dependencies,
  without installation hooks. Homebrew/Scoop URLs, versions and hashes matched.
- Downloaded Linux AMD64 CLI passed version/help, health exit codes 0–5, tool
  failure exit 6, JSON/YAML validation, ping/doctor, AHP authentication,
  capabilities/dependencies, graceful shutdown and credential redaction checks.
- Public PyPI wheel and sdist matched signed release bytes; PyPI attestations
  verified. A fresh public installation passed 15 SDK integration tests against
  the downloaded CLI.
- Public npm installation passed native Linux CLI invocation. All seven installed
  dependency registry signatures and the SDK provenance attestation verified with
  `npm audit signatures`. Public tarball bytes and SHA-512 integrity matched the
  signed release tarball. Installed tarball ESM and TypeScript smoke checks passed.

npm tarball SHA-256:
`6952c91363913a8b68fbc41e745cc4de33c84dac2d68dab6d58c05d0a8cd0811`.

## Containers and OCI packages

| Public package | Immutable manifest digest |
| --- | --- |
| `ghcr.io/theagenthealth/agenthealth:v0.14.0` | `sha256:4222306310bdcd478d534f1aa5fa26039fc1ec5584dbcc5b486c358454058eb6` |
| `theagenthealth/agenthealth:v0.14.0` | `sha256:36a503ce079a124ba5fd3fce9637e1d3753c33b09cfacedf0f2b72fed7f5f765` |
| `ghcr.io/theagenthealth/charts/agenthealth:0.14.0` | `sha256:889dfc267f5438e51229961e1e12b870ad56bb05b9962615d71cf7b6ae2ba482` |
| `ghcr.io/theagenthealth/agenthealth-binaries:v0.14.0` | `sha256:61365b4fdc01629848e34c95e47e89d74c054b2e905e234c774c50e2ec5171ac` |

Anonymous Docker, Helm and ORAS retrieval passed. Both image architectures matched
signed archive binaries; native AMD64 version/help and non-root execution passed.
BuildKit SBOM/provenance and GHCR signed provenance verified. OCI manifests,
configs and every layer matched their digests; the binary bundle contained the
same 21 assets. The downloaded chart matched its source/version/image metadata,
passed strict lint in all three modes, and ran serve, Job and CronJob modes in a
disposable Kubernetes cluster using the public image. The cluster was deleted.

## Limits

Local runtime verification used Linux AMD64. Other platform binaries were checked
by archive inspection and native CI; local ARM64 execution is not claimed.
Executable OS signing, macOS notarization, Helm GPG signing, APT/YUM repositories,
and managed Homebrew/Scoop repositories remain outside this release. Windows SDK
cancellation terminates the direct CLI process; POSIX process-group cleanup has
broader descendant coverage. SDK installation does not download a CLI binary.
