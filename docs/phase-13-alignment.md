# Phase 13 coverage and earlier-phase alignment

Phase 13 package tooling is included in v0.11.0.
The v0.10.0 baseline retains its five archives, SBOMs and checksums; v0.11.0
adds four Linux packages and two channel manifests. This audit covers
repository source alignment, not publication or native installation certification.

## Coverage checklist

| Layer | Coverage |
|---|---|
| README and roadmap | Source/release distinction, channel scope, remaining public hosting and native signing work |
| Documentation | [Installation](installation.md#package-channels-phase-13-v0110), [release runbook](../RELEASING.md#phase-13-package-assets), [agent instructions](../AGENTS.md), documentation and phase indexes |
| Examples | [Local distribution preparation](../examples/distribution/README.md); existing health configurations require no migration |
| Archive tooling | [Release builder](../scripts/build_release.py) retains names, normalized metadata, executable SBOMs and version injection |
| Package tooling | [Package builder](../scripts/package_release.py) verifies archive hashes, generates Homebrew/Scoop metadata and reuses Linux payloads for DEB/RPM |
| Tests | [Archive/SBOM tests](../tests/test_release_archives.py), [manifest and tampering tests](../tests/test_package_release.py), [platform-selection/tampering tests](../tests/test_release_smoke.py), [native smoke checks](../scripts/test_release_binary.py), [older-tag recovery tests](../tests/test_release_recovery.py) |
| CI and release | [CI](../.github/workflows/ci.yml) builds and smoke-tests five native platforms; [release](../.github/workflows/release.yml) gates publication on those checks and attests all generated assets |

## Earlier phases

| Phase | Alignment or no-impact rationale |
|---|---|
| 0: Foundation | README, indexes, agent runbook and release policy updated; new distribution CI added. Governance and repository identity need no change. |
| 1: AHS | No schema, health-state, dimension, target, protocol or exit-code changes. Specification docs and fixtures remain valid. |
| 2: Engine | Same executable and Go engine in every distribution. No execution, aggregation, cancellation, concurrency or diagnostic changes. Existing core tests remain applicable. |
| 3: CLI | Commands, flags, output and exit codes remain identical. Native archive smoke checks verify version/help and missing-config exit 6. |
| 4: HTTP/API | No adapter changes. Linux packages depend on CA certificates; TLS verification and endpoint prerequisites remain required. HTTP guides/configurations remain valid. |
| 5: MCP | OAuth/login, private token storage and stdio behavior unchanged. Installing the CLI does not install a trusted MCP server executable. Existing examples remain valid. |
| 6: A2A | Protocol selection, discovery, origins, credentials and active opt-in unchanged; existing guides and examples need no migration. |
| 7: Agent health | Direct agent, composite agent, path and supporting-dependency evidence unchanged. Runtime probe handlers remain operator prerequisites. |
| 8: Agentgateway | Signal/backend/path configuration and safety unchanged; packaging does not provision a gateway or its backends. |
| 9: Agent Router | Named signal and route/backend evidence unchanged; packaging does not provision a router. |
| 10: Dependency graph | Node/edge contracts and graph examples unchanged. Alignment guide points to current Docker/distribution status. |
| 11: Experimental AHP | Same serving command, authentication and freshness behavior. Packages create no daemon/service or configuration; serving example remains valid. |
| 12: Docker | Container builds and payload comparison unchanged. Corrected release policy to describe existing parallel container/binary publication accurately. Container MCP/token-file limitations remain valid. |

Historical releases and RFC proposals retain their original scope; Phase 13
source work does not add package assets to earlier published versions. Individual
adapter READMEs need no packaging edits because their behavior did not change.

At the first Phase 13 release, CLI formats published together while SDKs/charts
used independent versions. The [canonical release policy](../RELEASING.md#component-versions-and-synchronized-distribution)
now requires shared software versions starting with the planned v0.12.0 release.
See [the distribution plan](distribution.md) for the coordinated Releases/Packages
layout and outstanding OCI publication work; historical release evidence stays valid.

## Later phases and limitations

The roadmap records exact-version/platform verification requirements for SDKs
(15–16) and CI integrations (21), and installation/signing follow-ups for
hardening (26). Kubernetes (14) retains the CLI probe exit contract. Other later
adapters, observability, plugins and conformance designs need no contract changes
because Phase 13 adds installation metadata only.

Local validation covers Go vet/race tests, Python regressions, links and marked
configuration examples, five cross-compiled archives, four built packages,
asset hashes, native Linux AMD64 archive/DEB execution, and byte comparison of
both DEB payloads with their archives. Native macOS, Windows and Linux ARM64
runtime checks are configured in CI and have not been run locally.

Public taps/buckets and APT/YUM repositories, real package-manager installation,
upgrade/removal matrices, native signing/notarization and package-specific SBOM
attestations remain follow-up work. Archive SBOMs describe the reused executable;
build provenance covers package assets. Existing optional SDK integration tests
may be skipped locally and do not establish live interoperability.

## Phase 14 follow-through

Kubernetes integration is now implemented in source. Its
[alignment audit](phase-14-alignment.md) records existing contract compatibility
and validation. This earlier-phase baseline and historical release notes retain
the capabilities of their own released versions.
