# Offline conformance examples

Build the CLI with `go build -o agenthealth ./cmd/agenthealth`, then run from the
repository root (Windows executable names may use `.exe`):

```bash
./agenthealth conformance test examples/conformance/healthy.json
./agenthealth conformance test examples/conformance/dependency.json
./agenthealth conformance test examples/conformance/inconsistent.json
```

The first two commands return 0 with `passed: true`. The dependency example is
conformant even though the parent is unhealthy: its critical composite agent is
unreachable. This is endpoint evidence, not proof of first-to-downstream
communication. Supporting dependencies and paths remain separate evidence.

The last command intentionally returns 6 with `inconsistent_status`: its shape
is schema-valid, but its overall status contradicts the reachability check.
These are static fixtures, not observations of running agents. No network,
credentials or external service is needed. See the [guide](../../docs/conformance.md)
for limits and the separate reference behavioral suite.
