# Official SDK interoperability

The suite launches actual official SDK servers on loopback and runs the Go CLI
against them. Fixture handlers return deterministic health text through the SDK;
they do not handcraft protocol envelopes. Run on each PR, main push and nightly
through [the interoperability workflow](../.github/workflows/interoperability.yml).

| Official Python SDK | Protocol | Verified behavior |
|---|---|---|
| MCP 2.3.0 | 2026-07-28 | HTTP and stdio discovery, inventory, one safe tool call, required-tool failure |
| MCP 1.30.0 | 2025-11-25 | HTTP and stdio legacy negotiation, inventory, one safe tool call, required-tool failure |
| A2A 1.2.2 | 1.0 JSON-RPC | Default discovery, read-only task lookup, skills, safe message, required-skill failure |
| A2A 0.3.26 | 0.3.0 JSON-RPC | Explicit legacy discovery, read-only task lookup, skills, safe message, required-skill failure |

Both pinned SDK sets pass locally on Linux AMD64. Core runtime tests also run
on Linux, macOS and Windows in CI; official SDK integration currently runs on
Linux. This matrix does not establish interoperability with every deployment,
external OAuth provider, other SDK language, gRPC/REST binding, or streaming.
The MCP OAuth fixture tests cover configured grants and token redaction separately.

## Reproduce

From the repository root, with Go and Python 3.12 available:

```bash
python3 -m venv /tmp/agenthealth-sdk-tests
/tmp/agenthealth-sdk-tests/bin/pip install -r tests/interop/requirements-current.txt
go build -o /tmp/agenthealth-sdk-tests/agenthealth ./cmd/agenthealth
AGENTHEALTH_INTEROP=1 AGENTHEALTH_BINARY=/tmp/agenthealth-sdk-tests/agenthealth /tmp/agenthealth-sdk-tests/bin/python -m pytest tests/interop -v
```

Use a separate virtual environment with `requirements-legacy.txt` to test legacy
servers. Regular tests skip this opt-in suite when the environment variable is
absent. Required capability failures assert exit 2 and stable diagnostic codes.
SDK versions are pinned so an upstream release cannot silently alter the matrix.

Sources: [MCP Python SDK](https://github.com/modelcontextprotocol/python-sdk),
[A2A Python SDK](https://github.com/a2aproject/a2a-python),
[A2A v1 specification](https://a2a-protocol.org/v1.0.1/specification/).
