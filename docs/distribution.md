# Releases, Packages and synchronized versions

The agreed next release is **v0.12.0**. All implemented artifacts will use that
software version and the same immutable source commit. This replaces the earlier
independent chart/SDK version policy. CLI v0.11.1 and chart 0.1.0 remain published
and immutable; historical release notes and verification records remain valid.

This document records the target publication layout. The shared release workflow and OCI chart/binary publication are implemented
in source. v0.12.0 and its new OCI packages are **not published yet**. An
authorized shared tag builds new assets and coordinates all publication jobs;
it does not rename old downloads.

## One complete GitHub Release

Use one v0.12.0 release page containing:

- Five standalone archives: Linux AMD64/ARM64, macOS AMD64/ARM64 and Windows AMD64.
- Five matching CycloneDX SBOMs, four Linux DEB/RPM packages and Homebrew/Scoop manifests.
- Helm chart `agenthealth-0.12.0.tgz` and one complete asset checksum manifest.
- Package links for Docker, Helm and the binary OCI bundle, with exact published digests.
- Both configured container registry references, installation commands, verification instructions, source links and actual validation limits.

Container image references and digests belong in the release notes; the image
is stored in its registry. An offline OCI/container archive may be attached when
explicitly included in release scope. It is not currently an existing release
asset. Do not claim a downloadable image archive merely because a registry link
is present.

## GitHub Packages layout

| Package | Target reference | User tool | Availability |
|---|---|---|---|
| Docker | `ghcr.io/theagenthealth/agenthealth:v0.12.0` | Docker-compatible container client | Existing image package; new version not published yet |
| Helm | `oci://ghcr.io/theagenthealth/charts/agenthealth`, version `0.12.0` | Helm | Implemented in source; publication pending |
| Binaries | `ghcr.io/theagenthealth/agenthealth-binaries:v0.12.0` | ORAS or a compatible OCI artifact client | Implemented in source; publication pending |

Link every package to `TheAgentHealth/agenthealth`, record source/revision/version
metadata and verify its public visibility. Publishing to GHCR does not establish
anonymous access by itself; new packages can require a maintainer to change
visibility in GitHub settings. Verify each actual package before announcement.
See [GitHub's Container registry documentation](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry).

The binary package is an additional OCI artifact containing the complete release files, including the chart so the common checksum
manifest has no absent entries;
it is not a runnable Docker image or an SDK. Keep bundled files byte-identical
to their signed Release downloads and make checksum coverage match the bundle
contents. Ordinary GitHub downloads remain the easiest installation path for
users without an OCI artifact client. DEB/RPM files in this bundle do not create
an APT/YUM repository; channel manifests do not create a Homebrew tap or Scoop
bucket. Those hosting channels remain separate work.

## Version syntax

One software version does not require identical punctuation in every format:

| Field | v0.12.0 value |
|---|---|
| Git tag and GitHub Release | `v0.12.0` |
| CLI version and container tag | `v0.12.0` |
| Chart version and Helm OCI tag | `0.12.0` |
| Chart appVersion and default image | `v0.12.0` |
| DEB/RPM, Homebrew/Scoop and future SDK version metadata | `0.12.0` |
| OCI binary bundle tag | `v0.12.0` |

Helm derives its OCI tag from chart version metadata. See the
[Helm OCI registry guide](https://helm.sh/docs/topics/registries/).
AHS/AHP and configuration/result contract versions remain separate; synchronized
software releases do not rename `spec_version: v1`.

## Installation examples

The following commands describe the **planned v0.12.0 channels** and are not
available until publication and verification complete. Existing v0.11.1/archive
and chart 0.1.0 releases remain available. Source installation guides prepare
the v0.12.0 commands; substitute historical tags for currently published files.

Direct installation from the unified Release download:

```bash
helm upgrade --install agenthealth \
  https://github.com/TheAgentHealth/agenthealth/releases/download/v0.12.0/agenthealth-0.12.0.tgz \
  --namespace agenthealth --create-namespace \
  --set-file config=./agenthealth.yaml
```

Installation from the OCI chart package:

```bash
helm upgrade --install agenthealth \
  oci://ghcr.io/theagenthealth/charts/agenthealth \
  --version 0.12.0 \
  --namespace agenthealth --create-namespace \
  --set-file config=./agenthealth.yaml
```

The configuration must describe the user's services; the fixture's `demo`
Service is not automatically installed by the chart. For users who verify the
chart before installation, download it, verify its GitHub signed provenance and
checksum, then install the verified local package. Helm GPG `.prov` verification
is not currently provided.

Retrieve the additional binary OCI bundle:

```bash
oras pull ghcr.io/theagenthealth/agenthealth-binaries:v0.12.0 \
  --output ./agenthealth-v0.12.0
```

See [ORAS pull documentation](https://oras.land/docs/commands/oras_pull/).
Verify OCI provenance/digest and signed asset checksums before selecting and
extracting the correct OS/architecture archive. Do not use `docker run` or
`helm install` for the binary bundle. Binary download, checksum and signature
verification remain available directly through GitHub Releases.

## Implementation and publication checklist

- [x] Update release workflows and the Helm helper to consume the same software tag; avoid separate release creation and conflicting checksum uploads.
- [x] Advance chart metadata and all current image/install references to 0.12.0/v0.12.0 before tagging.
- [ ] Build all release files from the exact same tagged source; generate the final checksum manifest after all formats are packaged.
- [ ] Publish the Docker image, OCI chart and OCI binary bundle; attach all downloadable files to one GitHub Release.
- [ ] Link packages to this repository and establish public visibility; verify anonymous registry access and repository Packages discovery.
- [ ] Sign downloadable assets and OCI provenance; verify source tag/commit, digests, SBOMs, contents and byte identity across channels.
- [ ] Gate release completion on all required publication jobs, regardless of whether jobs execute in parallel.
- [ ] Download published files and verify native Linux behavior and actual chart installation; report other platforms according to evidence available.
- [ ] Preserve existing releases and explain that v0.12.0 supersedes v0.11.1/chart 0.1.0.

Skipping local test reruns at a maintainer's request does not mean that new
workflow changes have passed old CI. Report skipped checks and respect actual
GitHub merge requirements. Artifact builds, publication, signed provenance and
post-publication verification are distinct from regression-test reruns and are
still necessary before declaring the release complete.

Future SDKs adopt the common release version when implemented. Publish them to
supported ecosystem registries such as PyPI/npm and link their distributions
from the common release page; do not claim a planned SDK or unsupported package
format is already hosted in GitHub Packages.

## Workflow coordination and first-package recovery

The shared workflow validates source, builds and signs all 18 downloadable
files (17 content assets plus checksums), then stages one draft release.
The container job publishes both configured image registries. The packages job
consumes the signed build artifact, publishes the Helm chart and full binary
bundle, signs their OCI provenance and checks anonymous manifests/blob hashes.
Only the final coordinator publishes the complete GitHub Release and adds the
verified package digests. The old chart workflow is retired for new tags.

New GHCR packages can initially be private. If public verification fails after
package creation, make charts/agenthealth and agenthealth-binaries public in
GitHub package settings and use `gh run rerun <run-id> --failed`. Keep the draft,
original signed files and immutable tag. The publisher reuses matching existing
OCI versions and rejects different content. Do not rerun every job, recreate a
tag or overwrite a public release to resolve a visibility issue.
