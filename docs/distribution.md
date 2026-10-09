# Releases, Packages and synchronized versions

The agreed next release is **v0.12.0**. All implemented artifacts will use that
software version and the same immutable source commit. This replaces the earlier
independent chart/SDK version policy. CLI v0.11.1 and chart 0.1.0 remain published
and immutable; historical release notes and verification records remain valid.

This document records the target publication layout. The existing workflows
still publish CLI and chart releases separately. OCI chart/binary publication
and coordinated completion are **not implemented or published yet**. A new
release must implement those changes rather than simply republish old assets
with different filenames.

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
| Helm | `oci://ghcr.io/theagenthealth/charts/agenthealth`, version `0.12.0` | Helm | OCI publication planned |
| Binaries | `ghcr.io/theagenthealth/agenthealth-binaries:v0.12.0` | ORAS or a compatible OCI artifact client | OCI publication planned |

Link every package to `TheAgentHealth/agenthealth`, record source/revision/version
metadata and verify its public visibility. Publishing to GHCR does not establish
anonymous access by itself; new packages can require a maintainer to change
visibility in GitHub settings. Verify each actual package before announcement.
See [GitHub's Container registry documentation](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry).

The binary package is an additional OCI artifact containing the release files;
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
and chart 0.1.0 installation remains documented in
[installation](installation.md) and [Kubernetes](kubernetes.md).

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

- [ ] Update release workflows and the Helm helper to consume the same software tag; avoid separate release creation and conflicting checksum uploads.
- [ ] Advance chart metadata and all current image/install references to 0.12.0/v0.12.0 before tagging.
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
