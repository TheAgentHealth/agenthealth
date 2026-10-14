# v0.13.2 publication verification

Verified on 2026-10-14 UTC (2026-10-13 America/Chicago) after
[Release run 38490405098](https://github.com/TheAgentHealth/agenthealth/actions/runs/38490405098) succeeded.

- [GitHub Release v0.13.2](https://github.com/TheAgentHealth/agenthealth/releases/tag/v0.13.2) is public, regular and Latest.
- Immutable source: `cd8e85d4bda39e754668e50ad68de899a5a81345`.
- [Main CI](https://github.com/TheAgentHealth/agenthealth/actions/runs/38489812404) and [fresh tagged-source CI](https://github.com/TheAgentHealth/agenthealth/actions/runs/38490414756) passed: six Python SDK jobs, five native binary jobs, Go tests, upgrade/rollback and Kubernetes 1.31.9/1.32.2/1.33.1 runtime checks.
- Downloaded all 20 public assets anonymously. GitHub digests and bytes match signed staging files; all 19 checksum entries match.
- Verified 20 source-bound GitHub build provenance attestations, five signed CycloneDX predicates and three signed GHCR package provenance statements. Verification requires the release workflow and exact source commit and rejects self-hosted runners.
- Inspected five archives: executable platform headers, version-specific INSTALL.md, license, permissions, timestamps, executable hashes and Go 1.27.2 SBOMs. Four DEB/RPM payloads match archives; metadata, certificate dependencies and absence of install hooks passed. Homebrew/Scoop URLs, versions and hashes match.
- Downloaded Linux AMD64 CLI passed version/help, health exits 0–5, tool exit 6, JSON/YAML schema validation, ping/doctor, AHP authorization, dependencies, capabilities, shutdown and credential redaction.
- Both container registries permit anonymous access. Linux AMD64/ARM64 image executables match signed archives; manifest/layer digests, source/version labels, runtime identity and BuildKit SBOM/provenance passed. Native AMD64 version/help passed.
- Chart 0.13.2/appVersion v0.13.2, image pin, source revision and packaged files match tagged source. Strict lint and serve/job/cronjob rendering passed. Anonymous OCI chart and complete binary-bundle manifest/config/all-blob checks passed.
- [PyPI agenthealth 0.13.2](https://pypi.org/project/agenthealth/0.13.2/) exposes exactly the wheel and source archive. Downloads match PyPI SHA-256 and signed GitHub bytes. Both PyPI attestations cryptographically verify the GitHub publisher, release.yml workflow and pypi environment.
- A fresh virtual environment installed the public PyPI package and passed all 15 SDK integration tests against the downloaded matching CLI. CI also rebuilt/installed the source archive and checked repeat-build byte identity.
- Local tests: 235 passed, five explicitly skipped environment-dependent cases; CI covers the platform/runtime gates separately. Documentation links and canonical examples passed.

## Immutable package references

| Package | Digest |
|---|---|
| ghcr.io/theagenthealth/agenthealth | `sha256:2f333b296fed27a301c1be0b18208dae95ee8dfa4329512fb18adec22f395b3d` |
| theagenthealth/agenthealth | `sha256:73134420a033c30c04c118d693c21737dfa2baa2d69617d5741a015c16d7133e` |
| ghcr.io/theagenthealth/charts/agenthealth | `sha256:23939bc8cf2a8fc07baaa6e54420aea5f53d39b68d51b791ceb09bdfce12d7f8` |
| ghcr.io/theagenthealth/agenthealth-binaries | `sha256:9920594c8db9bf40fa0423a47b39c8c6fe61d6911ffb9ab842aaeb77d9817da3` |

## Attempt history and limits

v0.13.0 staging stopped when verification found stale embedded installation references.
v0.13.1 staging failed because the PyPA action pin was an annotated tag object
without a matching container image. Its infrastructure-only recovery then caught
a macOS SDK cleanup race. Both immutable tags and staged artifacts are preserved;
neither became a final GitHub or PyPI release. v0.13.2 fixes installation rendering,
the publisher commit pin and cleanup, accepting a permission error only after
confirming CLI exit.

The first Copilot review's reproducibility and stderr credential findings were
addressed. No second review was requested or awaited. Final fixes were pushed to
main under the maintainer's explicit authorization.

Native OS signing/notarization, Helm GPG .prov, APT/YUM repositories, public
Homebrew/Scoop hosting and cloud-cluster certification remain deferred. Local
ARM64 container execution was not tested; native ARM64 archive CI passed.
Windows detached stdio descendants need external lifecycle controls. GitHub
Packages has no Python registry: the wheel/source archive are ordinary GitHub
Release assets and also included in the signed OCI file bundle.

Local verification reports are under `/tmp/agenthealth-release-verification/v0.13.2`;
that temporary directory is not a public download channel.
