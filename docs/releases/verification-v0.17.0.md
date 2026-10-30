# v0.17.0 publication verification

Verified on October 30, 2026 (America/Chicago). The regular Latest
[GitHub Release](https://github.com/TheAgentHealth/agenthealth/releases/tag/v0.17.0),
[PyPI](https://pypi.org/project/agenthealth/0.17.0/) and
[npm](https://www.npmjs.com/package/@theagenthealth/agenthealth/v/0.17.0) are public.

## Source and gates

Immutable tag `v0.17.0` points to `a45f24ceffb775fda13f491df3fd230bb3f34025`, the merge commit of
[PR #58](https://github.com/TheAgentHealth/agenthealth/pull/58). Its 17 focused
commits implement the Qdrant/Pinecone Phase 19 baseline and realign phases 0–26.
All 42 [PR CI jobs](https://github.com/TheAgentHealth/agenthealth/actions/runs/41033837424)
passed before the user-authorized approval bypass was used.

The first Copilot review was assessed. Valid schema zero-value parity and
packaged SDK README findings were fixed; Compose exclusion was already fixed.
The proposed Qdrant array decoder was rejected with a direct Qdrant 1.15.5
`result.points` response and official query-endpoint documentation.
No second Copilot review was requested or awaited.

[Fresh tagged-source CI](https://github.com/TheAgentHealth/agenthealth/actions/runs/41034796628)
passed all 41 test jobs, including five native archive platforms, real Qdrant
and native databases, both SDK matrices and Kubernetes 1.31.9/1.32.2/1.33.1.
Only the PR-only path detector was intentionally skipped.
The [Release workflow](https://github.com/TheAgentHealth/agenthealth/actions/runs/41034782605) succeeded on attempt 2.

npm publication first became visible through the per-version endpoint while
the install index still returned ETARGET. After public npm install visibility
was confirmed, only failed jobs and dependents were retried. The retry reused
the identical npm tarball, immutable tag, signed assets and existing packages;
nothing was rebuilt, replaced or republished under a changed source.

## Downloads, signatures and runtime

- Downloaded all 21 files anonymously. GitHub digests and all 20 SHA-256
  checksum entries matched the complete required synchronized distribution set.
- Verified all 21 build-provenance attestations against the immutable source
  tag/commit and hosted release workflow, plus five CycloneDX attestations.
  Verified signed OCI provenance for the image, chart and binary bundle.
- Inspected five archives and executable/SBOM hashes. The downloaded Linux
  AMD64 executable reports v0.17.0 and passed a real Qdrant functional query
  and denied-key checks without disclosure. Other platforms ran in native CI.
- PyPI wheel/sdist bytes and SHA-256 match signed assets. Both PyPI attestations
  verified with pypi-attestations. A fresh public installation passed all 18
  SDK integration tests against the downloaded matching CLI.
- Public npm tarball bytes and SHA-512 integrity match the signed asset.
  Fresh registry installation passed signature/attestation auditing and
  installed-package ESM/TypeScript/CLI smoke checks.
- Anonymous GHCR pull passed version/help, non-root image configuration and
  HTTP-check smoke tests. Both image architectures bind the exact source and
  version; GHCR and Docker Hub filesystem/runtime identities match.
- Downloaded public chart and public image passed serve, Job and CronJob modes
  on disposable Kubernetes 1.32.2. Earlier local source smoke covered init
  outage/recovery, exec/HTTP probes, sidecars, Secrets and readiness recovery.

## Public manifests

| Package | Immutable manifest digest |
|---|---|
| GHCR image | `sha256:fedb1bbb852713b7bca6e5721cb00d7f6675c5ad3eaf86d20d26f1234994ca68` |
| Docker Hub mirror | `sha256:58d46986656edd3c9d176a5ac869853d282e0ec212abccb7e5961cbf99a0f0db` |
| `ghcr.io/theagenthealth/charts/agenthealth:0.17.0` | `sha256:7f1e26deb3018650318c159fd38b017c2c924a751503f51acff73f9338437165` |
| `ghcr.io/theagenthealth/agenthealth-binaries:v0.17.0` | `sha256:5329c8a3f217c646ab82d11572bc4600addc285ec2e01c73ffff4fa35b64467f` |

Anonymous OCI verification checked every chart/bundle blob, manifest,
configuration hash, source/version annotation and both Linux image platforms.
The OCI binary bundle contains the same 21 Release files. Chart version is
0.17.0, appVersion is v0.17.0, and image defaults match.

## Scope and limits

Controlled vector interoperability covers Qdrant 1.15.5. Pinecone dense-vector
REST 2025-10 uses official API fixtures; live Pinecone cloud accounts are not
claimed as tested. Weaviate, Milvus, pgvector, OpenSearch, Elasticsearch, named/
sparse vectors and broad provider/cluster/failover matrices remain planned.
Query execution does not prove relevance or corpus completeness. PostgreSQL
login/SELECT 1 does not verify pgvector retrieval. No security audit is claimed.
See the [vector guide](../vector-store-adapter.md) and
[all-phase alignment audit](../phase-19-alignment.md).
