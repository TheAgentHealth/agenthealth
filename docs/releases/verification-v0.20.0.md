# AgentHealth v0.20.0 publication verification

## Source and release gates

- Immutable source tag: `v0.20.0`, commit `6b42d213d3ff530b77eb7839bb5bac4e42d2c645`.
- [Implementation PR #66](https://github.com/TheAgentHealth/agenthealth/pull/66)
  preserves sixteen focused commits. Its first Copilot review identified a genuine
  premature SDK publication claim; both SDK READMEs were corrected. No second
  Copilot review was requested or awaited.
- [PR CI](https://github.com/TheAgentHealth/agenthealth/actions/runs/41387879406)
  and [tagged-source release CI](https://github.com/TheAgentHealth/agenthealth/actions/runs/41388751637)
  passed, including all five native distribution jobs, SDK matrices, independent
  Prometheus/OTLP parser validation and the required Kubernetes matrix.
- [Release workflow](https://github.com/TheAgentHealth/agenthealth/actions/runs/41388739645)
  completed successfully. Initial npm post-publication installation hit stale
  registry metadata; only failed/downstream jobs were rerun after propagation.
  Existing signed assets and matching registry bytes were reused; no source tag
  or published artifact was moved, replaced or rebuilt for that retry.
- [Public regular Latest release](https://github.com/TheAgentHealth/agenthealth/releases/tag/v0.20.0); pre-1.0 maturity remains unchanged.

## Download and runtime evidence

- All 21 public release files match GitHub SHA-256 digests; all 20 manifest entries
  match downloaded bytes. All 21 build-provenance attestations bind this source.
- Five signed CycloneDX attestations, archive contents and binary/SBOM hashes
  verified. The downloaded Linux AMD64 CLI reports v0.20.0 and passes real HTTP
  checks in JSON, JUnit, GitHub and all four observability formats. Other native
  platforms were exercised by CI rather than cross-platform local execution.
- The public CLI passes independent Prometheus/protobuf OTLP parsing, redaction,
  shared-node deduplication and repeated-edge uniqueness checks.
- Anonymous GHCR image pull and version/help/non-root/real HTTP smoke passed.
  GHCR and Docker Hub platform/runtime identities match this source/version.
- Downloaded public chart and public image pass serve, Job and CronJob modes on
  disposable Kubernetes 1.33.1. Source runtime smoke additionally covered outage,
  recovery, probes, sidecars, Secrets and failed Jobs.
- Public PyPI wheel/sdist bytes match signed assets; both PyPI attestations pass.
  A fresh public SDK/framework installation passes 48 tests with the public CLI.
- Public npm bytes and SHA-512 integrity match the signed tarball. Fresh registry
  installation, registry signatures/attestations and real CLI smoke pass.
- All three GHCR identities have signed source/version bindings and anonymous
  access. OCI chart/binary manifests, layers and every bundled file match signed
  release assets.

## Immutable registry identities

| Package | Verified digest |
|---|---|
| GHCR Docker `ghcr.io/theagenthealth/agenthealth:v0.20.0` | `sha256:c1ee476905fa3336a6e8ec05c456fcd42637ae5fea0ba54f58731f8848e993ba` |
| Docker Hub `theagenthealth/agenthealth:v0.20.0` | `sha256:8ba2474117af457b35ec8a16ce72d8ae890f6c066984100b2d7bf86c8e221f7a` |
| Helm `ghcr.io/theagenthealth/charts/agenthealth:0.20.0` | `sha256:9b5a3f6a530763ea94b6e22a0c3b1e4b15e0535b3dbded3a114fa799bfa7236e` |
| OCI binaries `ghcr.io/theagenthealth/agenthealth-binaries:v0.20.0` | `sha256:e3bdb191b0a6ec10a86ab96e85e42606d2e9a56a73ffb1c6e56777a5f1270172` |

Actual collector/hosted-webhook product interoperability remains unverified.
The export baseline provides finite payloads and explicit external delivery;
no scrape server, durable queue, gRPC exporter, tenant isolation or conformance
certification is claimed. Native executable signing and GPG Helm `.prov` files
remain deferred; the verified signatures above are keyless attestations.
