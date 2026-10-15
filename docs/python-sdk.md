# Python SDK

Phase 15 implements a Python 3.10+ wrapper around the Go CLI. Published SDK and CLI use
software v0.14.0; publication is recorded separately in [capability status](phase-status.md).
The SDK consumes AHS v1 and never reimplements adapter execution or aggregation.

## Installation

Install the published version:

```bash
pip install agenthealth==0.14.0
```

Install the matching CLI v0.14.0 separately using the
[verified binary installation](installation.md#verify-provenance-and-sboms) instructions.
Select its exact platform archive and verify both checksums and signed
provenance before execution. The SDK never downloads binaries, changes PATH,
or executes code during import. For source development:

```bash
pip install -e sdk/python
```

## Synchronous and asynchronous checks

```python
from agenthealth import Client, HealthStatus, load_config

client = Client(binary="/usr/local/bin/agenthealth", timeout=30)
report = client.check(load_config("agenthealth.yaml"))
print(report.status, report.exit_code)
for result in report.results:
    print(result.target.id, result.target.name, result.status)
    for dependency in result.dependencies:
        print(dependency.relationship, dependency.critical, dependency.target.name)
```

```python
from agenthealth import async_check, check

report = check(target_type="http", endpoint="http://localhost:8080/health")

async def probe():
    return await async_check("agenthealth.yaml", timeout=15)
```

`check` returns a `CheckReport` for single or multiple targets. `report.result`
is convenient for one root; batches use ordered `report.results`.
`report.status` and `exit_code` reflect the Go aggregate health result. Health
exits 0–5 return results. `HealthResult` exposes target ID/type/name, status,
latency, checks, recursive dependencies, relationship and critical metadata.
Each `CheckEntry` exposes status, message, diagnostic code and transport steps.
`to_dict()` returns an independent copy, preserving unknown additive fields.
Direct/composite roles remain contextual; no new role fields are invented.

All implemented adapters are accessed through `target_type` or the existing
configuration. `options` supplies additional target configuration, such as
`auth`, `mcp`, `a2a`, `agent`, `gateway`, `router` or `dependencies`. It cannot
override the named target arguments. Active checks require the same explicit
configuration opt-in as the CLI. OAuth login remains a separate CLI operation.

`Configuration(data)` copies and validates JSON-compatible configuration.
`load_config(path)` reads bounded JSON/YAML and rejects duplicate mapping keys.
Schema validation is local; the Go CLI performs graph and semantic validation.
Relative paths resolve from the caller's working directory, matching CLI
behavior. Use `Configuration(data, base_dir=...)` to select another directory.
The SDK writes configuration into a private temporary directory and removes it
on completion, errors, timeout and cancellation.

## Errors and execution policy

| Exception | Meaning |
|---|---|
| `ConfigurationError` | Invalid/unreadable SDK configuration or conflicting arguments |
| `BinaryNotFoundError` | CLI missing or cannot start |
| `CompatibilityError` | CLI software version differs from the SDK |
| `ExecutionError` | CLI tool failure, including semantic configuration rejection (exit 6) |
| `CheckTimeoutError` | SDK deadline expired and the process was stopped |
| `OutputError` | Oversized, invalid or inconsistent AHS output |

All inherit `AgentHealthError`. Exception messages omit configuration values,
CLI stderr, response bodies and credentials. `asyncio.CancelledError` propagates
after cleanup. Use `await client.async_check()` inside an event loop; synchronous
`check` does not create a nested loop.

A deadline covers version verification and checking. Configuration is limited
to 4 MiB, stdout to 16 MiB, stderr to 64 KiB; output is captured in temporary
files and monitored while the CLI runs. These are SDK limits in addition to
engine transport, graph and deadline controls. There are no SDK retries.
Each POSIX run has a process group; timeout/cancellation terminates CLI and
stdio descendants. Windows termination covers the CLI process; detached stdio
children require operator lifecycle controls.

Configuration remains [trusted executable input](../SECURITY.md#trusted-configuration).
`environment` overrides selected inherited environment variables, useful for
credential references; credentials never appear in SDK exception messages.
The SDK does not sandbox programs, authorize tenants or register Python adapters.

## PyPI and GitHub distribution

The pending PyPI publisher must name project `agenthealth`, owner
`TheAgentHealth`, repository `agenthealth`, workflow `release.yml`, environment
`pypi`. The GitHub environment allows `v*` tags. A dedicated publishing job in
[the shared release workflow](../.github/workflows/release.yml) uses OIDC with
job-level `id-token: write` and PyPI attestations; API tokens are unnecessary.

Wheel and source archive use the same software version and immutable source tag
as CLI/Docker/Helm. They are built once, included in release checksums and GitHub
build provenance, uploaded to the draft GitHub Release, and published unchanged
to PyPI. The final GitHub Release waits for PyPI, Docker and OCI packages.
GitHub Packages has no Python/PyPI registry; Python distributions are attached
as GitHub Release assets. The existing GHCR binary bundle also contains them,
but requires an OCI artifact client rather than pip.

CI installs the built wheel on Linux, macOS and Windows with Python 3.10/3.13,
rebuilds the wheel from the source archive, and exercises real CLI integration,
error handling, cancellation, typed graph/path evidence and schema alignment.
The [phase impact review](phase-15-alignment.md) records compatibility and limits.
