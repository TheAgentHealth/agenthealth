# Release Process & Versioning

See [AGENTS.md](AGENTS.md#prepare-and-publish-a-release) for the agent release
runbook, including preflight checks and verification of published downloads.

## Versioning scheme

AgentHealth follows [Semantic Versioning 2.0.0](https://semver.org/) for the CLI, core engine, adapters, and SDKs:

```text
MAJOR.MINOR.PATCH

MAJOR — breaking changes to the CLI, configuration format, result schema, or adapter interface
MINOR — backward-compatible functionality (new adapters, new checks, new flags)
PATCH — backward-compatible bug fixes
```

Until [AgentHealth 1.0](ROADMAP.md#agenthealth-10), the project is pre-1.0 (`0.y.z`): breaking changes may occur in minor releases, but will be called out in release notes.

## Release labels

Plain version tags such as `v0.2.0` publish regular GitHub releases marked
**Latest**. Pre-1.0 describes API maturity; it does not require GitHub's
pre-release label. Regular releases still carry the compatibility limitations
documented below.

Use a suffixed tag such as `v0.3.0-rc.1` for an intentional preview. The workflow
marks suffixed tags as pre-releases and does not make them Latest. To promote
a published preview without changing its tag or downloads:

```bash
gh release edit <tag> --prerelease=false --latest
```

Promotion requires an explicit maintainer decision. v0.2.0 was promoted to a
regular release; v0.1.0 remains a historical preview.

## Specification versioning

The specification documents under [spec/](spec/README.md) are versioned independently via `spec_version` (e.g. `v1`), separate from the software release version. A software release can ship `spec_version: v1` across several `MINOR`/`PATCH` releases without a spec change.

## Component versions and synchronized distribution

Version components by their public interfaces. Release every distribution format
of the same component together. A shared repository does not require every
component to have the same version or to publish on every release.

| Component or artifact | Version and publication policy |
|---|---|
| CLI, core engine and in-tree adapters | One CLI version, tagged `vX.Y.Z` |
| Standalone archives, DEB/RPM and Homebrew/Scoop manifests | Same CLI version; generate and publish together for every CLI release |
| GHCR container and configured Docker Hub mirror | Same CLI version and source; publish with every CLI release even when the Dockerfile is unchanged |
| Python SDK / PyPI (planned) | Independent SDK version; publish when its code/API or bundled/pinned CLI dependency changes |
| JavaScript/TypeScript SDK / npm (planned) | Independent SDK version; publish when its code/API or bundled/pinned CLI dependency changes |
| Helm chart | Independent chart version; `appVersion` records the CLI version and the image reference pins an explicit tag or digest |
| Kubernetes examples/manifests | Update with affected behavior; no separate package version unless a packaged component is introduced |
| AHS/AHP contracts | Contract versions remain independent of software versions |

For example, CLI `v0.11.0` publishes matching archives, Linux packages, channel
manifests and container images. An unchanged SDK may retain its existing version
if its documented compatibility covers that CLI. Updating an SDK's bundled or
pinned CLI requires a new SDK release. Updating a published chart's default image
requires a new chart version, even when the templates are unchanged.

Publish by component release intent, not by arbitrary repository changes or
Dockerfile-only path filters. Pull requests and merges run validation; they do
not publish packages. The current `v*` release workflow publishes the implemented
CLI formats only. Future SDK/chart workflows must use separate, component-specific
release triggers (for example `python-vX.Y.Z`, `javascript-vX.Y.Z`,
`helm-vX.Y.Z`) and must not trigger CLI publication. Python/JavaScript names are a design for future workflows; the Phase 14
`helm-v*` trigger is implemented for chart publication.

Release notes identify the component, artifacts and compatibility requirements.
Keep published versions immutable. Verify all required formats before declaring
a CLI release complete; binary and container jobs may run in parallel, but a
successful binary upload alone does not establish successful container publication.
Public taps/buckets and package repositories remain separate hosting work.

## Release steps (current, pre-1.0)

1. Changes land on `main` via reviewed pull requests.
2. When a release is cut, tag `main` as `vX.Y.Z`.
3. Build and publish artifacts for that tag:
   - standalone binaries and checksums (implemented early from [Phase 13](ROADMAP.md#phase-13--standalone-binaries)),
   - multi-platform container image with provenance, SBOM and signed attestation ([Phase 12](ROADMAP.md#phase-12--docker-distribution)), published in parallel with the binary release after source validation,
   - Linux packages and Homebrew/Scoop manifests once Phase 13 tooling is included.
   The implemented Helm chart publishes through its independent `helm-v*`
   component workflow; future SDKs will also use separate component releases.
   A CLI tag does not automatically publish them.
4. Publish release notes summarizing changes, including any breaking changes and migration notes.

**First container release only:** GitHub creates a new organization package as
private by default. After the very first `container` job publishes
`ghcr.io/theagenthealth/agenthealth`, a maintainer with package admin access
must open the package's settings on GitHub and change its visibility to
public; otherwise `docker run ghcr.io/theagenthealth/agenthealth` fails with
`unauthorized` for everyone except the publishing workflow. The release
workflow's smoke test logs out of GHCR first so this is caught as a release
failure instead of silently passing.

**Docker Hub mirror (optional):** the `container` job also pushes to
`docker.io/theagenthealth/agenthealth` once the `DOCKERHUB_USERNAME` and
`DOCKERHUB_ACCESS_TOKEN` repository secrets exist (generate a Docker Hub access token
under Account Settings → Personal access tokens, scope Read & Write). Until
those secrets are added, the Docker Hub steps are skipped and only GHCR is
published.

The [release workflow](.github/workflows/release.yml) validates tagged source, cross-compiles five platform archives using [scripts/build_release.py](scripts/build_release.py), gates publication on native archive smoke tests for all five platforms, generates package assets and verifies checksums, and publishes the GitHub release only after asset upload succeeds. Each tag needs release notes at `docs/releases/vX.Y.Z.md`. The first release is `v0.1.0`.

To prepare artifacts locally (the output directory must be empty):

```bash
python3 scripts/build_release.py v0.1.0 --output /tmp/agenthealth-v0.1.0-assets
```

After the release PR passes CI and is merged, tag that exact main commit:

```bash
git checkout main
git pull --ff-only
git tag -a v0.1.0 -m "AgentHealth v0.1.0"
git push origin v0.1.0
```

Watch the Release workflow and verify the downloads before announcing the release. Tags are immutable release identifiers; fixes ship under a new version rather than retagging. Plain version tags publish regular releases by default; suffixed tags publish pre-releases. v0.4.0 adds per-platform CycloneDX SBOMs, keyless signed provenance and SBOM attestations, SHA-pinned Actions, vulnerability scanning and normalized archive metadata. OS-native signing and notarization remain [Phase 26 — Production Hardening](ROADMAP.md#phase-26--production-hardening) work.

## Backward compatibility policy

Once 1.0 ships, the project commits to:

- no breaking changes to the result schema, configuration format, or CLI flags within a `MAJOR` version,
- a deprecation period (documented in release notes) before removing a flag, config field, or adapter capability,
- clear migration notes for any `MAJOR` version bump.

Before 1.0, no such guarantee is made; this document exists to make pre-1.0 expectations explicit rather than leaving them undocumented.

Release builds use Go 1.27.2 and govulncheck v1.8.0. Each published release
contains five archives, five matching SBOMs and checksums. Starting in v0.11.0, Phase 13 tooling also produces four DEB/RPM packages and two channel manifests. Verify provenance and
SBOM attestations following [installation instructions](docs/installation.md#verify-provenance-and-sboms).

## Infrastructure-only recovery

If artifacts were built and signed but publication stopped because of release
infrastructure, the Release workflow supports manual dispatch with
`release_tag` set to the existing immutable tag. It checks out that tag,
validates and rebuilds its source, and requires every rebuilt asset to match
the original signed provenance before adding SBOM attestations and publishing.
It preserves the original build provenance and tag. This recovery path requires
existing provenance for every asset; source changes require a new version.

v0.4.0 used this path after [PR #10](https://github.com/TheAgentHealth/agenthealth/pull/10)
corrected the action's detection of schema-valid CycloneDX documents without
the optional `serialNumber`. Published SBOMs use explicit CycloneDX predicates.

## Phase 13 package assets

After building archives, run `python3 scripts/package_release.py <tag> --output <directory>`.
This verifies existing checksums, reuses the Linux executable, license and install-document bytes from each verified archive, invokes
[nFPM](https://nfpm.goreleaser.com/docs/configuration/) v2.47.0 through Go (requires Go 1.26.4 or newer, or automatic toolchain download), and
adds four DEB/RPM packages plus Homebrew/Scoop manifests to `checksums.txt`.
The release workflow signs build provenance for every asset. Archive SBOMs
remain bound to their archive executable; packages reuse that executable but
do not yet have separate SBOM attestations. Verify package provenance before
installing. Native signing and package repository signatures remain Phase 26.

The release waits for native archive smoke tests on all five platforms.
Publishing a tap/bucket or APT/YUM repository is separate from generating
release assets and is not automated. Older tag recovery checks whether the tagged source contains the Phase 13
helpers before invoking them, preserving the original asset set. Do not generate package assets
for an already published release.

## Helm chart releases

Phase 14 introduces independent `helm-v<chart-version>` releases through
[the chart workflow](.github/workflows/helm-release.yml). CLI `v*` tags do not
publish charts. Chart 0.1.0 has `appVersion: v0.11.1` and pins that released
image. Change the chart version whenever templates or image defaults change.

Prepare locally with `python3 scripts/package_helm.py helm-v0.1.0 --output
/tmp/agenthealth-chart-release` (requires Helm 3 and an empty output directory).
This validates chart metadata, requires component release notes, packages
`agenthealth-0.1.0.tgz` and writes `checksums.txt`. Run the Kubernetes CI scenarios
before an authorized release. Publishing requires a separately authorized,
reviewed chart tag on the intended main commit; do not tag during implementation.

Notes must link to `deploy/kubernetes`, `deploy/helm/agenthealth` and
`examples/kubernetes` at the component tag, include the packaged chart download,
record CLI/chart versions and actually validated Kubernetes versions, and show
both same-version registry references. Verify anonymous image availability and
all release links before announcement. Chart releases are not GitHub Latest,
so they do not displace the CLI release. The chart workflow signs package/checksum build provenance through GitHub
Actions attestations. Verify those signatures with the chart signer workflow
and exact tag commit. GPG-signed Helm `.prov` files remain deferred; SHA-256
checksums alone do not establish publisher identity.
