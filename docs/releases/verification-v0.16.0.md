# v0.16.0 publication verification

Verified on October 29, 2026 (America/Chicago). The
[GitHub Release](https://github.com/TheAgentHealth/agenthealth/releases/tag/v0.16.0),
[PyPI](https://pypi.org/project/agenthealth/0.16.0/) and
[npm](https://www.npmjs.com/package/@theagenthealth/agenthealth/v/0.16.0) are public.
The GitHub Release is a regular Latest release with 21 signed files.

## Source and gates

Immutable tag `v0.16.0` points to `2cf8a59c470dd03a6329c1605067e1e61d7a7dbd`, the merge commit of
[PR #55](https://github.com/TheAgentHealth/agenthealth/pull/55). Its 25 focused
commits implement Phase 18 and realign phases 0–26. The first Copilot review was
assessed and valid findings addressed, including schema boundaries, transport
classification and bundled SDK READMEs. MySQL readiness was strengthened to an
authenticated query; its alleged startup blocker was not reproduced. No second
Copilot review was requested or awaited. The authorized approval bypass was used
only after all 41 [PR checks](https://github.com/TheAgentHealth/agenthealth/actions/runs/40994768122) passed.

[Fresh tagged-source CI](https://github.com/TheAgentHealth/agenthealth/actions/runs/40995650144)
passed all 40 test jobs, including five native archive platforms, the four native
databases, both SDK matrices and Kubernetes 1.31.9/1.32.2/1.33.1. The PR-only path
detector was intentionally skipped. Go 1.26 minimum tests and the compatibility
status required by the existing protected branch both passed.

The [Release workflow](https://github.com/TheAgentHealth/agenthealth/actions/runs/40995638244)
succeeded on its second attempt. npm publication initially preceded propagation
to its install index, causing an `ETARGET` during the final clean installation.
The public index was checked before retrying only failed jobs and dependents.
The retry verified and reused the identical npm tarball; no source tag, signed
asset or already published registry version was rebuilt, moved or overwritten.

## Downloads, signatures and runtime

- Downloaded all 21 files anonymously; GitHub asset digests and all 20 SHA-256
  checksum entries matched the complete required distribution set.
- Verified all 21 build-provenance attestations against the exact source tag,
  commit and hosted release workflow, plus all five CycloneDX attestations.
  Verified signed OCI provenance for the image, chart and binary bundle.
- Inspected all five archives and archive-bound executable/SBOM hashes. The
  downloaded Linux AMD64 executable reports `v0.16.0` and passed all four native
  database functional probes and wrong-password checks without disclosure.
  Other platforms were validated by native release CI, not local execution.
- PyPI wheel/sdist public bytes and SHA-256 matched signed release assets. Both
  PyPI attestations verified with `pypi-attestations` against this repository.
  A fresh public PyPI installation passed all 22 SDK tests against the downloaded CLI.
- Public npm tarball bytes and SHA-512 integrity matched the signed asset. A
  fresh installation verified seven registry signatures and two attestations.
  Installed-tarball ESM and TypeScript checks passed against the downloaded CLI.
- An anonymous public GHCR image pull passed version/help, non-root runtime
  configuration and HTTP-check smoke tests. Both image architectures identify
  the exact source/version; GHCR and Docker Hub filesystem/runtime identities match.
- The downloaded public chart and public image passed serve, Job and CronJob
  modes on a disposable Kubernetes 1.32.2 cluster. Earlier local source smoke
  also covered init outage/recovery, probes, sidecars, Secrets and readiness.

## Public manifests

| Package | Immutable manifest digest |
|---|---|
| `GHCR Docker image` | `sha256:ec147b0368f76bd45d104ff3c3249257ce893b711649d6d9d0e1ce171249f606` |
| `Docker Hub mirror` | `sha256:bb9db2f7f81a3dd45778f2cce19d0b6021c19717646d7b1fefef33794c97fc0c` |
| `ghcr.io/theagenthealth/charts/agenthealth:0.16.0` | `sha256:13c7588c35f6a829cd7a909947ac644783996b43bb27fd55aea29347e68569ad` |
| `ghcr.io/theagenthealth/agenthealth-binaries:v0.16.0` | `sha256:20b1ff516d50f6d73a136aaa3c57d8e5479d4bfc4b3e288ccee99de54a6fc6e2` |

Anonymous registry verification checked every chart/bundle layer, manifests,
configuration hashes, source/version metadata and both Linux image platforms.
The OCI binary bundle contains the same 21 Release files. Chart version is
`0.16.0`, appVersion is `v0.16.0`, and default image pins match.

## Scope and limits

Controlled native tests use PostgreSQL 17.6, MySQL 8.4.6, Redis 7.4.5 and MongoDB
8.0.13 with explicit plaintext loopback fixtures. Verified TLS is the adapter's
default, and untrusted-certificate rejection has regression coverage; broad
server/TLS/cloud matrices, topology/failover discovery and cloud/IAM login remain
future work. Login/PING and constant-query success do not prove application data
permissions, freshness or retrieval readiness. No security audit is claimed.
Source builds require Go 1.26+, while release binaries use patched Go 1.27.3.
See the [database guide](../database-adapter.md) and
[all-phase alignment review](../phase-18-alignment.md).
