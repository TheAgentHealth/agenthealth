# JavaScript SDK example

Requires Node 22.14+ and a verified AgentHealth CLI v0.14.0 on PATH, or
`AGENTHEALTH_BINARY` pointing to it. Install the published v0.14.0 package:

```sh
cd examples/javascript-sdk
npm install --ignore-scripts
node check.mjs http://localhost:8080/health
```

The example prints a typed SDK report and uses the health exit code (0–5).
Missing/mismatched CLI and invalid configuration reject rather than becoming
health results. For source development, install a locally built npm tarball.
See the [SDK guide](../../docs/javascript-sdk.md).
