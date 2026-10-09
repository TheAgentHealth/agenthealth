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

Starting with the next synchronized release, **v0.12.0**, all implemented
AgentHealth artifacts and packages share one software version and one immutable
source tag. Publish them together, including unchanged distribution formats.
This supersedes the earlier independent chart/SDK version policy. Do not move,
rename or overwrite CLI `v0.11.1` or chart `helm-v0.1.0`; they remain historical
releases. AHS/AHP and `spec_version` contract versions remain independent.

| Artifact | Version for v0.12.0 | GitHub Releases | Package distribution |
|---|---|---|---|
| CLI, engine and adapters | `v0.12.0` | Five platform archives, five CycloneDX SBOMs and checksums | Additional GHCR OCI binary bundle |
| DEB/RPM and Homebrew/Scoop manifests | `0.12.0` package metadata | Four Linux packages and two channel manifests | Include in binary bundle; APT/YUM repositories and public taps/buckets remain separate work |
| Docker image | `v0.12.0` image tag | Registry links, exact digests and installation instructions; offline image archives are optional | GHCR container image and configured Docker Hub mirror |
| Helm chart | `version: 0.12.0`, `appVersion: v0.12.0` | Chart `.tgz`, checksums and verification instructions | GHCR OCI chart; Helm's OCI tag is `0.12.0` |
| Future SDKs | Same software version when implemented | Release notes and appropriate distributions/registry links | Use supported ecosystem registries; do not claim planned SDKs are published |

The leading `v` is a Git/source and container-tag convention, not a different
release version. Helm and package managers use their required numeric version
syntax. Every image default and binary download reference must select the same
release. A breaking change to any shipped component determines the shared
version increment; backward-compatible capabilities require a minor release.

Use one `vX.Y.Z` release as the discovery page for all downloads, registry
packages and installation/verification commands. Package workflows may be
separate jobs, but consume that same tag and source commit and form one
coordinated release. Do not create another `helm-v*` release for the new policy.
Do not announce success until every required distribution has published and its
public accessibility, provenance, hashes and contents have been verified.

**Implementation status:** synchronized publication is implemented in source in
[release.yml](.github/workflows/release.yml). It builds a complete checksum set,
stages a signed draft, publishes Docker/chart/binary OCI packages and gates
public release completion on all required jobs and anonymous package access.
v0.12.0 has not been published. See [distribution](docs/distribution.md) for
package names, commands, visibility recovery and remaining publication checks.

## Release steps (current, pre-1.0)

1. Changes land on `main` via reviewed pull requests.
2. When a release is cut, tag `main` as `vX.Y.Z`.
3. Build and publish artifacts for that tag:
   - standalone binaries and checksums (implemented early from [Phase 13](ROADMAP.md#phase-13--standalone-binaries)),
   - multi-platform container image with provenance, SBOM and signed attestation ([Phase 12](ROADMAP.md#phase-12--docker-distribution)), published in parallel with the binary release after source validation,
   - Linux packages and Homebrew/Scoop manifests once Phase 13 tooling is included.
   The next synchronized release must also attach the same-version Helm chart
   and publish the chart/binary OCI packages through coordinated jobs. The
   shared workflow stages the draft and waits for all required publication jobs.
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

The historical chart 0.1.0 was published as `helm-v0.1.0` with CLI
`v0.11.1`. Preserve that release and its signed files. Starting with the next
release, use the shared `vX.Y.Z` tag, chart `version: X.Y.Z`, CLI
`appVersion: vX.Y.Z` and the matching image. Attach the chart package to the same
GitHub release as the CLI and publish it as a GHCR OCI chart.

The [shared workflow](.github/workflows/release.yml) consumes the software tag.
The [chart helper](scripts/package_helm.py) requires matching chart/app/image
versions and normalizes the package for deterministic provenance-backed recovery.
The [asset helper](scripts/prepare_synchronized_release.py) verifies the complete
CLI set, adds the chart and regenerates one checksum manifest before signing.
The [OCI publisher](scripts/publish_oci.py) consumes those exact signed files,
links source/revision/version annotations, refuses to replace differing existing
versions and verifies anonymous access to all asset blobs. The historical
[chart workflow](.github/workflows/helm-release.yml) no longer responds to tags.

Release notes must link to `deploy/kubernetes`, `deploy/helm/agenthealth` and
`examples/kubernetes` at the shared tag and include both chart download and OCI
installation commands, the exact CLI/image version and tested Kubernetes
versions. Verify the chart package's signed provenance and the OCI chart's
published digest/source binding. GPG Helm `.prov` files remain deferred; use
GitHub keyless attestations rather than suggesting `helm --verify` works.

See [distribution and synchronized-release preparation](docs/distribution.md).
