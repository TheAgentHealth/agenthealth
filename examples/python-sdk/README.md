# Python SDK example

Install the SDK from source with `pip install -e sdk/python` and build a CLI
matching the SDK's software version. Install SDK/CLI v0.15.0
from their verified published artifacts. Start the services referenced by your
reviewed configuration before running checks.

From the repository root:

```bash
python examples/python-sdk/check.py examples/core-check/agenthealth.yaml --binary /path/to/agenthealth
python examples/python-sdk/check.py examples/graph-check/agenthealth.yaml --binary /path/to/agenthealth --async
```

The graph example requires configured direct/composite agents, a router and an
explicitly authorized safe path probe; see [graph setup](../graph-check/README.md).
The example preserves CLI health exits 0–5 and returns 6 for SDK/tool failures.
It prints typed dependency relationships without treating supporting dependencies
as composite agents. See the [SDK guide](../../docs/python-sdk.md).

Published software: v0.15.0. Requires the matching CLI; model/LLM
configuration is supported through the canonical schema.
