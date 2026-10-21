# AgentHealth Python SDK

Typed synchronous and asynchronous access to AgentHealth's Go CLI. This package
preserves AHS v1 results, diagnostic codes, dependency relationships and graph
IDs. Health failures return results; tool failures raise exceptions.

Install the package with `pip install agenthealth` once published. Install a
verified AgentHealth CLI of the same software version separately and put it on
PATH, or pass its absolute path. Installation and imports never download or
execute a binary. This first SDK targets Python 3.10+ and CLI v0.15.0.

```python
from agenthealth import Client, HealthStatus

client = Client(binary="/usr/local/bin/agenthealth", timeout=30)
report = client.check(target_type="http", endpoint="http://localhost:8080/health")
print(report.status, report.exit_code)
if report.status == HealthStatus.HEALTHY:
    print(report.result.target.name)
```

```python
from agenthealth import async_check

async def probe():
    return await async_check("agenthealth.yaml")
```

Use `Configuration` for Python dictionaries or `load_config` for JSON/YAML.
The same configuration supports all implemented CLI adapters, explicit active
checks, nested dependencies and graph references. Credentials stay in configured
references. Configuration is trusted executable input: MCP stdio can launch
programs and endpoints can reach private networks. Loading validates schema;
Go validates graph semantics at execution.

The SDK checks the exact CLI version, bounds configuration and output, enforces
an overall deadline and stops checks on timeout/cancellation. POSIX checks use
a dedicated process group to clean up descendants. On Windows termination
covers the CLI process; detached stdio descendants require operator lifecycle
controls. CLI stderr is never included in exceptions.

The SDK does not aggregate health independently, run OAuth login, register
Python adapters, or auto-install the CLI. `report.results` preserves ordered
roots, including each dependency projection. `to_dict()` preserves additional
result fields. Relative paths retain the caller's working directory, matching
the CLI; `Configuration(data, base_dir=...)` selects a different directory.

Source implementation is pre-1.0; PyPI availability is recorded separately in
the repository capability status. See the repository's Python SDK guide for
installation, release setup, errors and examples.

Source release target: v0.15.0. Requires the matching CLI; model/LLM
configuration is supported through the canonical schema.
