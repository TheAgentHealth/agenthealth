# Release Process & Versioning

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
   - standalone binaries (once [Phase 11](ROADMAP.md#phase-11--standalone-binaries) lands),
   - Docker image (once [Phase 10](ROADMAP.md#phase-10--docker-distribution) lands),
   - PyPI package (once [Phase 13](ROADMAP.md#phase-13--python-sdk--pypi) lands),
   - npm package (once [Phase 14](ROADMAP.md#phase-14--javascript--typescript-sdk) lands).
4. Publish release notes summarizing changes, including any breaking changes and migration notes.

This process will be automated (CI-driven releases, signed artifacts, SBOM generation) as part of [Phase 24 — Production Hardening](ROADMAP.md#phase-24--production-hardening) ahead of a stable 1.0.

## Backward compatibility policy

Once 1.0 ships, the project commits to:

- no breaking changes to the result schema, configuration format, or CLI flags within a `MAJOR` version,
- a deprecation period (documented in release notes) before removing a flag, config field, or adapter capability,
- clear migration notes for any `MAJOR` version bump.

Before 1.0, no such guarantee is made; this document exists to make pre-1.0 expectations explicit rather than leaving them undocumented.
