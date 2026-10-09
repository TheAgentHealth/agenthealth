# Phase 14 published-release verification

Verified on October 9, 2026, after [PR #41](https://github.com/TheAgentHealth/agenthealth/pull/41)
merged with all 17 CI checks passing. The six original Copilot findings were
fixed; no additional Copilot review was requested or awaited.

Both immutable component tags identify merge commit
`6c2000745b05969e0f7f8b46e77571e3964a5946`:

- [CLI v0.11.1](https://github.com/TheAgentHealth/agenthealth/releases/tag/v0.11.1): regular Latest release, 17 downloadable assets.
- [Helm chart 0.1.0](https://github.com/TheAgentHealth/agenthealth/releases/tag/helm-v0.1.0): independent component release, two downloadable assets; `appVersion: v0.11.1`.

The [CLI release workflow](https://github.com/TheAgentHealth/agenthealth/actions/runs/37889162974)
and [chart release workflow](https://github.com/TheAgentHealth/agenthealth/actions/runs/37889822748)
both passed. Chart publication followed verification of both CLI registries.

## Download verification

All 19 published files were downloaded from their public releases. GitHub CLI
attestation verification enforced the repository, expected signer workflow,
exact source commit and tag reference, and GitHub-hosted runners.

| Artifact | Verification result |
|---|---|
| CLI assets | All 17 build-provenance signatures passed; all 16 entries in the signed checksum manifest matched. |
| Archive SBOMs | All five archive-bound CycloneDX signatures passed; signed predicates matched downloaded SBOMs, including executable hashes, platform, release version and Go 1.27.2. |
| Archives | Exact executable, Apache license and installation-guide contents; expected ELF, Mach-O and PE architectures and normalized ownership/modes/timestamps. |
| DEB/RPM | All four architecture/version records and exact three-file payloads checked; executable bytes matched signed Linux archives, certificate dependencies were present, and no installation hooks were included. |
| Homebrew/Scoop | Versions, download URLs and hashes matched the corresponding signed archives. |
| Chart | Package and checksum-file provenance signatures passed; package SHA-256 matched; files matched tagged source, with chart version 0.1.0 and CLI appVersion v0.11.1. Strict lint and all three workload modes passed. |

## Containers and Linux behavior

Both registries allowed anonymous manifest access and contained Linux AMD64 and
ARM64 images. Every extracted image executable matched its signed CLI archive.
Nonroot image configuration, source/version labels and native AMD64 `version`
and help behavior passed. GHCR's signed index provenance passed; both registries'
embedded BuildKit provenance and SPDX SBOMs were checked for both architectures,
including manifest/blob hashes and subject bindings.

| Registry | Verified multi-platform index digest |
|---|---|
| `ghcr.io/theagenthealth/agenthealth:v0.11.1` | `sha256:61c70483db3d8ac8323b6b732985fa7b1460cabd45239f71e1b87cd7a3a86405` |
| `docker.io/theagenthealth/agenthealth:v0.11.1` | `sha256:1e92da396a68997a86d1f36d3dfe0c53a1fb96dba4a1f9f86312d3ef27b91a4f` |

The downloaded Linux AMD64 executable passed `version`, help, `ping`, `doctor`,
all health exit codes 0–5 and tool-error exit 6 against controlled HTTP fixtures.
JSON/YAML results passed schema validation. AHP serving passed live/readiness,
authenticated detailed results, rejection of unauthenticated requests, token
redaction and graceful SIGTERM shutdown.

## Kubernetes and released links

On Linux AMD64 with kind 0.27.0, Kubernetes 1.32.2 and Helm 3.17.3:

- The tag-pinned Kustomize example completed with a healthy direct agent and all four dependencies.
- The published-image scenario suite passed process startup during dependency outage, init blocking/recovery, HTTP and exec probes, sidecar, accepted/rejected Secret credentials, Jobs/CronJobs, all downloaded-chart modes and readiness recovery without restarts.
- Helm installed the downloaded chart in serve, Job and CronJob modes; Deployment readiness and validation/scheduled Jobs passed.
- The optional application-image example ran the CLI as a nonroot user and retained the signed archive's exact executable bytes.
- All six CLI/chart tag-pinned manifest, chart-source and usage-example links resolved; both release descriptions included their exact source paths and chart download/usage instructions.

## Validation limits

The release workflow smoke-tested native archives on all five advertised
platforms. Local runtime checks used Linux AMD64; extracted ARM64 container
executables were compared, not locally executed. Kubernetes 1.29 is the declared
API floor, not an additional tested runtime. The HTTP fixtures do not certify
real agent, gateway or router interoperability.

GitHub keyless provenance signatures are available. Native executable
signing/notarization, package repository signatures, real package-manager
installation matrices and Helm GPG `.prov` files remain future work. Docker Hub's
embedded BuildKit records were verified alongside signed-archive binary identity;
they are not a separate GitHub keyless signature on its index.
