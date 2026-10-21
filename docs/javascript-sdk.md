# JavaScript / TypeScript SDK

Source release target: **v0.15.0**. The commands below require its publication;
check the [release page](https://github.com/TheAgentHealth/agenthealth/releases/tag/v0.15.0)
before installing. Links to v0.14.0 verification describe that historical release.

Phase 16 releases [`@theagenthealth/agenthealth`](https://www.npmjs.com/package/@theagenthealth/agenthealth/v/0.15.0) for Node.js **22.14+**, ESM.
It wraps the reference Go CLI at the exact same software version and does not
fetch or run executables on installation/import. Browser execution is outside
this SDK's scope. The software remains pre-1.0.

```sh
npm install @theagenthealth/agenthealth@0.15.0
```

Install and verify CLI v0.15.0 separately following
[installation](installation.md). See the [package API](../sdk/javascript/README.md)
for configuration, cancellation, timeouts, limits and typed errors, and the
[runnable example](../examples/javascript-sdk/README.md).

```typescript
import { Client, loadConfig } from '@theagenthealth/agenthealth';
const client = new Client({ binary: 'agenthealth', timeout: 30000 });
const controller = new AbortController();
const config = await loadConfig('agenthealth.yaml');
const report = await client.check(config, { signal: controller.signal });
for (const root of report.results) {
  console.log(root.target.id, root.status, root.dependencies);
}
```

All existing adapters remain available through canonical configuration:
HTTP/API, MCP, A2A, direct/composite agents, gateway/router and graph targets.
The SDK preserves named peer/path checks, graph IDs, relationship and critical
metadata, transport steps and diagnostic codes. Supporting dependencies remain
distinct from composite agents. Health semantics and active-check opt-ins stay
in Go. No schema, exit-code or protocol changes are introduced by the wrapper.
AHP serving remains an experimental CLI capability.

Health statuses 0–5 resolve, including UNREACHABLE and MISCONFIGURED. CLI tool
failure 6, invalid output, incompatible CLI, timeout and cancellation reject
with sanitized typed errors. The timeout unit is milliseconds. POSIX cleanup
terminates process groups; Windows guarantees direct-process cleanup only.
Credentials use configured references and the caller's cwd unless a Configuration
explicitly selects a baseDir. Treat configuration as trusted executable input,
especially when it contains MCP stdio commands or enables active checks.

The synchronized release includes a reproducible npm tarball in signed GitHub
assets, the checksum set and the OCI binary bundle. The explicit `npm` job in `release.yml` publishes directly using the
configured OIDC identity and GitHub environment `npm`, alongside `pypi`.
The final `publish` job requires both. Verify downloaded
bytes against signed checksums and verify installed registry provenance:

```sh
npm install --ignore-scripts @theagenthealth/agenthealth@0.15.0
npm audit signatures
```

CI runs Node 22 and 24 on Linux/macOS/Windows, real CLI HTTP/batch/failure checks,
TypeScript compilation, schema parity, bounded error/cancellation tests,
deterministic packaging and installed-tarball import checks. This controlled
coverage does not complete broader Phase 25 ecosystem interoperability.
See the [impact review](phase-16-alignment.md).
