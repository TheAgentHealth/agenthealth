# Executable adapter extensions

Phase 23 implements an opt-in local executable transport. Install and review the
program yourself; AgentHealth never discovers or downloads executable adapters.
See the [contract and security boundary](rfcs/phase-23-adapters.md) and
[working example](../examples/plugin-check/README.md).

Build the example, calculate its SHA-256, and write a local manifest:

```json
{
  "protocol": "v1",
  "command": "/absolute/path/to/probe",
  "sha256": "REPLACE_WITH_LOWERCASE_SHA256",
  "metadata": {
    "name": "community-probe",
    "version": "1.0.0",
    "compatibility_version": "v1",
    "target_types": ["custom"],
    "dimensions": ["configuration", "reachability"],
    "active_checks": []
  }
}
```

```sh
agenthealth check examples/plugin-check/agenthealth.yaml --adapter /absolute/path/to/manifest.json --format json
```

Repeat `--adapter` for disjoint target types. Invalid manifests and duplicate
registrations are tool failures (exit 6); child execution failures become health
results. Existing health exit codes 0–5 and all output formats still apply.
Built-in target registrations cannot be replaced through this flag. `serve` and
SDK APIs do not load manifests. Compiled Go applications can explicitly register
`process.New` adapters in their own registry before constructing the engine.

Only identity, endpoint, dimension and the explicitly configured bearer credential
are sent to the child. Rich protocol options and additional credential references
are unsupported. Configuration must validate locally without networking. Report
remote responses immediately; emit only normalized status, never arbitrary text,
capability names, result fields, topology, labels or response bodies. Functional
checks must be declared active and explicitly requested. The engine owns latency,
dependency evidence and observability projections.

Keep agent role and path evidence distinct: a custom check named `direct-agent`
can observe a user-facing runtime; `composite-agent` observes a downstream runtime.
Neither proves communication between them. Use separate configured path probes
(or the existing agent safe peer task), and separate built-in A2A, gateway/router
and backend targets. See [agent examples](../examples/agent-check/README.md),
[gateway examples](../examples/gateway-check/README.md), and
[router examples](../examples/router-check/README.md). No role-specific target type
or inferred dependency discovery is introduced.

Executable metadata and status reports are trusted assertions from local code.
Process termination is implemented, but filesystem/network permissions remain
those of the invoking user. Keep binaries and manifests in trusted directories.
Hashes pin executable bytes; they do not pin shared libraries or prevent concurrent
replacement in a writable directory. No community certification is claimed.

On Windows, Go may supply the OS-required `SystemRoot` variable; application
environment variables remain excluded. Unix group termination is used before
reaping starts. After stdout closes and reaping starts, cancellation uses the
lifetime-aware direct-child handle; descendants that close inherited pipes are
not guaranteed to be terminated. OS containment remains outside this protocol.

## Phase 24 conformance evidence

Manifest compatibility, hash pinning, bounded streams, response markers, cancellation and normalization are included in the reference catalog. This is not community certification or an OS sandbox.
See the [conformance guide](conformance.md) and [cross-phase review](phase-24-alignment.md).
