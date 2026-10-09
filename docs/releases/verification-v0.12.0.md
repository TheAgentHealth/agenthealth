# v0.12.0 publication verification

Verified on 2026-10-09 after [Release run 37895572354](https://github.com/TheAgentHealth/agenthealth/actions/runs/37895572354) completed successfully.

- Release: [AgentHealth v0.12.0](https://github.com/TheAgentHealth/agenthealth/releases/tag/v0.12.0), public, regular and Latest.
- Source commit: `b42106ac0ecdcbca4176f33eeda44a371b6a3f51`.
- All five native archive smoke jobs succeeded: Linux AMD64/ARM64, macOS AMD64/ARM64 and Windows AMD64. Tagged-source validation also succeeded; no separate local source test suite was rerun.
- Downloaded all 18 public assets without GitHub credentials. Verified GitHub asset digests, 17 checksum entries and byte identity with the signed draft and complete ORAS bundle.
- Verified 23 downloaded-file attestations: 18 build provenance statements and five archive-bound CycloneDX SBOM predicates. Verified three signed GHCR OCI provenance statements against the release workflow and exact source commit; rejected self-hosted runners.
- Inspected all five archive contents, executable platform headers, license/install guide, SBOM version/platform/toolchain and executable hashes. All use Go 1.27.2.
- Inspected all four DEB/RPM payloads and architecture/version metadata, certificate dependencies and absence of install hooks. Packaged executables match signed archives. Homebrew/Scoop versions, download URLs and hashes match release archives.
- Downloaded Linux AMD64 CLI passed version/help, exit codes 0–6, JSON/YAML schema validation, ping/doctor, AHP authorization, dependency outage/recovery and token-redaction checks.
- GHCR and Docker Hub support Linux AMD64/ARM64. Both architectures' image executables match their signed archives. Anonymous manifest access, BuildKit SBOM/provenance contents and native AMD64 version/help passed. ARM64 image execution was not tested locally; its native archive workflow job passed.
- Anonymous Helm pull and ORAS pull passed. Their files match release downloads. Workflow anonymous verification checked every OCI manifest, config and asset blob hash.
- Chart metadata is version 0.12.0/appVersion v0.12.0, source revision matches the release, and chart contents match merged source. All three workload modes render correctly.
- Installed the public OCI chart in serve, job and cronjob modes on disposable Kubernetes 1.32.2 with Helm 3.17.3. Deployment rollout, Job completion and a manually triggered CronJob returned healthy results.
- The public repository-filtered Packages page lists Docker, charts/agenthealth and agenthealth-binaries. Existing v0.11.1 and helm-v0.1.0 releases remain unchanged.

## Immutable package references

| Package | Digest |
|---|---|
| GHCR Docker | `sha256:ad568974f50b11e4c9342b6dd17816e9f391397501ff922197d99f60ea70cd31` |
| Docker Hub | `sha256:f9ae3e1ceda1738d4b445504eed1ba155b1aa343988e7c1aefc3c186fd9ab7d2` |
| Helm OCI | `sha256:abb80ced1bc0a60a5c087a482ce35c7540877297fb11d22527a93c744dd1df43` |
| Binary OCI | `sha256:86e357192c00cf75440b8e0d83ad4c4db1ed22893045927301318ac1d882a143` |

## Release attempt history and limits

The initial v0.12.0 attempt failed before artifact publication on an obsolete Kubernetes image assertion. The maintainer explicitly authorized deleting/recreating that unpublished tag at corrected source. An intermediate v0.12.1 attempt was canceled before publication and its tag removed. The current published v0.12.0 is now immutable.

Native OS signing/notarization, Helm GPG `.prov`, APT/YUM repositories and public Homebrew/Scoop hosting remain deferred. Verification uses GitHub keyless attestations. Local reports and downloaded files are in `/tmp/agenthealth-release-verification/`; that temporary location is not a public distribution channel.
