# AgentHealth JavaScript / TypeScript SDK

Node.js ESM SDK wrapping the reference Go CLI. Node **22.14+** is required.
Browsers, CommonJS `require`, automatic binary downloads and install scripts are
not supported. All implemented adapters and graph configurations use the same
AHS v1 configuration and result contracts as the CLI.

```sh
npm install @theagenthealth/agenthealth@0.15.0
```

Install and verify CLI **v0.15.0** separately using the
[installation guide](../../docs/installation.md). The SDK never downloads or
executes binaries at import/install time. A mismatched CLI raises
`CompatibilityError` before running checks.

```js
import { check } from '@theagenthealth/agenthealth';

const report = await check({
  version: 'v1',
  targets: [{ name: 'api', type: 'http', endpoint: 'http://localhost:8080/health',
    checks: ['reachability'] }]
}, { binary: 'agenthealth', timeout: 30000 });
console.log(report.status, report.results);
```

`check(config, options)` and `new Client(options).check(config, { signal })`
return promises. Config can be an AHS configuration object, a `Configuration`,
or a JSON/YAML filename. `loadConfig(path)` validates without executing checks.
`new Configuration(data, { baseDir })` explicitly changes the working directory;
otherwise relative credential/token paths use the caller's cwd, like the CLI.
Configuration files are trusted input: MCP stdio and explicitly enabled active
checks can execute configured programs or interactions.

`timeout` is an overall subprocess budget in **milliseconds**, default 30000,
including version probing (capped at five seconds). Pass an `AbortSignal` for
cancellation and `environment` for explicit environment overrides. Temporary
configuration directories are private, with files mode 0600 on POSIX. The SDK
caps configuration at 4 MiB, stdout at 16 MiB and stderr at 64 KiB. Errors never
include raw CLI output or configured values. POSIX cleanup kills the subprocess
group, including descendants; Windows cleanup terminates the direct CLI process
and does not guarantee termination of MCP descendants.

Health exit codes **0–5 resolve** to `{ results, exitCode, status, spec_version }`;
tool failures (including exit 6), malformed results, deadline expiration and
cancellation reject with exported `AgentHealthError` subclasses. Root order,
node IDs, dependency relationship/critical metadata, named peer/path checks,
diagnostic codes and additional result evidence are preserved. `status` is the
CLI aggregate status; the SDK does not recompute dependency policies.

`TargetConfiguration` intentionally permits protocol options validated by the
bundled canonical schema. Result interfaces type shared fields; additional
adapter evidence stays available as `unknown`. Experimental AHP serving remains
a CLI/server capability, rather than an SDK HTTP server.

See the [SDK guide](../../docs/javascript-sdk.md),
[example](../../examples/javascript-sdk/README.md), and
[Phase 16 impact review](../../docs/phase-16-alignment.md).

Source release target: v0.15.0. Requires the matching CLI; model/LLM
configuration is supported through the canonical schema.
