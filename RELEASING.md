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

## Specification versioning

The specification documents under [spec/](spec/README.md) are versioned independently via `spec_version` (e.g. `v1`), separate from the software release version. A software release can ship `spec_version: v1` across several `MINOR`/`PATCH` releases without a spec change.

## What gets its own version

| Component | Versioned independently? | Notes |
|---|---|---|
| `agenthealth` CLI / core engine | Yes | Primary release artifact |
| Adapters (MCP, A2A, HTTP, ...) | No, initially | Ship together with the core engine while adapters live in this repo |
| Python SDK (`agenthealth` on PyPI) | Yes | Tracks, but is not required to match, the CLI version |
| JavaScript SDK (`@agenthealth/sdk` on npm) | Yes | Tracks, but is not required to match, the CLI version |
| Docker image (`agenthealth/agenthealth`) | Yes, tagged to CLI version | Plus `latest` and major-version floating tags |
| Helm chart | Yes | Chart version and app version are tracked separately per Helm convention |

## Release steps (current, pre-1.0)

1. Changes land on `main` via reviewed pull requests.
2. When a release is cut, tag `main` as `vX.Y.Z`.
3. Build and publish artifacts for that tag:
   - standalone binaries and checksums (implemented early from [Phase 11](ROADMAP.md#phase-11--standalone-binaries)),
   - Docker image (once [Phase 10](ROADMAP.md#phase-10--docker-distribution) lands),
   - PyPI package (once [Phase 13](ROADMAP.md#phase-13--python-sdk--pypi) lands),
   - npm package (once [Phase 14](ROADMAP.md#phase-14--javascript--typescript-sdk) lands).
4. Publish release notes summarizing changes, including any breaking changes and migration notes.

The [release workflow](.github/workflows/release.yml) validates tagged source, cross-compiles five platform archives using [scripts/build_release.py](scripts/build_release.py), verifies the Linux AMD64 version and archive checksums, and publishes a GitHub prerelease only after asset upload succeeds. Each tag needs release notes at `docs/releases/vX.Y.Z.md`. The first release is `v0.1.0`.

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

Watch the Release workflow and verify the downloads before announcing the release. Tags are immutable release identifiers; fixes ship under a new version rather than retagging. The workflow publishes prereleases by default; stable release publication, signed artifacts, and SBOM generation remain [Phase 24 — Production Hardening](ROADMAP.md#phase-24--production-hardening) work.

## Backward compatibility policy

Once 1.0 ships, the project commits to:

- no breaking changes to the result schema, configuration format, or CLI flags within a `MAJOR` version,
- a deprecation period (documented in release notes) before removing a flag, config field, or adapter capability,
- clear migration notes for any `MAJOR` version bump.

Before 1.0, no such guarantee is made; this document exists to make pre-1.0 expectations explicit rather than leaving them undocumented.
