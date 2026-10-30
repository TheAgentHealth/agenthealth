# TheAgentHealth

> Vendor-neutral health and readiness checks for agentic systems

**AgentHealth** is an open-source, vendor-neutral, framework-neutral, and language-neutral project for determining whether AI agents and the infrastructure they depend on are **reachable, correctly configured, responsive, ready, and operationally healthy**.

## Start here

**Latest release: [v0.16.0](https://github.com/TheAgentHealth/agenthealth/releases/tag/v0.16.0), pre-1.0.** Supported adapters: agent/multi-agent, HTTP/API, MCP HTTP/stdio, A2A, gateway, and router. Model/LLM OpenAI-compatible and Anthropic checks and native PostgreSQL/MySQL/Redis/MongoDB database checks are released; vector-store adapters remain planned. See the [model guide](docs/model-adapter.md) and [database guide](docs/database-adapter.md).

### Install

Choose how you want to run AgentHealth. Published distributions share **v0.16.0**:

<!-- generated: distributions:start -->
| Use it as | Get it here | Installation guide |
|---|---|---|
| Standalone CLI | [Release downloads](https://github.com/TheAgentHealth/agenthealth/releases/tag/v0.16.0): Linux AMD64/ARM64, macOS Intel/Apple Silicon, Windows AMD64 | [Download, verify and extract](docs/installation.md) — no Go required |
| Docker image | [Docker Hub](https://hub.docker.com/r/theagenthealth/agenthealth): `theagenthealth/agenthealth:v0.16.0`; GHCR: `ghcr.io/theagenthealth/agenthealth:v0.16.0` | [Docker usage](docs/docker.md) — Linux AMD64/ARM64 |
| Kubernetes / Helm | `oci://ghcr.io/theagenthealth/charts/agenthealth`, chart version `0.16.0`; chart archive also in release downloads | [Kubernetes setup and examples](docs/kubernetes.md) |
| Linux DEB / RPM | AMD64/ARM64 packages in release downloads | [Local package installation](docs/installation.md#package-channels-phase-13-v0110) |
| Homebrew / Scoop | `agenthealth.rb` / `agenthealth.json` in release downloads | [Manifest usage](docs/installation.md#package-channels-phase-13-v0110); official tap/bucket hosting is planned |
| OCI download bundle | `ghcr.io/theagenthealth/agenthealth-binaries:v0.16.0` | [Retrieve with ORAS](docs/distribution.md#installation-examples); contains release files |
| Python SDK / PyPI | [PyPI](https://pypi.org/project/agenthealth/0.16.0/): `pip install agenthealth==0.16.0`; wheel/sdist in Release downloads | [Python SDK](docs/python-sdk.md) — requires the matching CLI |
| JavaScript / TypeScript SDK / npm | [npm](https://www.npmjs.com/package/@theagenthealth/agenthealth/v/0.16.0): `npm install @theagenthealth/agenthealth@0.16.0`; signed tarball in Release downloads | [JavaScript SDK](docs/javascript-sdk.md) — Node 22.14+ ESM and matching CLI |
<!-- generated: distributions:end -->

**Run immediately with Docker:**

```bash
docker run --rm theagenthealth/agenthealth:v0.16.0 version
docker run --rm theagenthealth/agenthealth:v0.16.0 --help
```

The same commands work with `ghcr.io/theagenthealth/agenthealth:v0.16.0`.
Use an exact version tag to pin a release; `v0` follows stable major-version-0
releases and `latest` follows the latest stable release.

**Install with Helm:** save a configuration describing your services as
`agenthealth.yaml` (see the example below), then run against your selected cluster:

```bash
helm upgrade --install agenthealth \
  oci://ghcr.io/theagenthealth/charts/agenthealth --version 0.16.0 \
  --namespace agenthealth --create-namespace \
  --set-file config=./agenthealth.yaml
```

See the [installation guide](docs/installation.md) for checksums and signed
provenance, and the [Kubernetes guide](docs/kubernetes.md) for workload modes.
The [Python SDK](docs/python-sdk.md) is published on PyPI as `agenthealth==0.16.0`; install a matching CLI separately. The [JavaScript SDK](docs/javascript-sdk.md) is published on npm for Phase 16. Official
APT/YUM repositories are also planned; DEB/RPM downloads are available today.

### 60-second example

Run a passive HTTP check against an endpoint you control:

```bash
./agenthealth version
./agenthealth ping http http://localhost:8080/health
```

For a configuration file, save this as `agenthealth.yaml`:

<!-- spec-example: configuration -->
```yaml
version: v1
targets:
  - name: local-api
    type: http
    endpoint: http://localhost:8080/health
    checks:
      - reachability
      - protocol
```

```bash
./agenthealth check agenthealth.yaml --format json
./agenthealth doctor agenthealth.yaml
```

A reachable endpoint returning HTTP 200 should produce `HEALTHY` and exit code `0`; an unavailable endpoint produces `UNREACHABLE` and exit code `3`. See [CLI usage](docs/cli.md) for all statuses and exit codes.

**Configuration is trusted executable input.** Review commands, endpoints, credential references, and token-file paths before running a configuration. See [the security boundary](SECURITY.md#trusted-configuration).

## Supported capabilities

| Target | Available evidence | Limits |
|---|---|---|
| HTTP/API | Connectivity, status, headers, bearer authentication; opt-in GET body matching | HEAD remains the passive default; explicit GET selection is supported |
| MCP | HTTP/stdio, protocol negotiation, inventory expectations, OAuth, opt-in safe tool probes | Configured stdio commands execute trusted code; supported protocol revisions are explicit |
| A2A | 1.0 JSON-RPC and explicit 0.3.0 compatibility, card/skill expectations, safe task probes | No generic framework integration, gRPC or streaming task execution |
| Agent/multi-agent | Application health document and explicit safe-task contract | Application integration is required; see reference middleware |
| Gateway | Gateway HTTP health signals and configured dependency/path evidence | No product discovery or independently verified Agentgateway interoperability |
| Router | Router HTTP health signals and configured dependency/path evidence | No route-table validation or decision verification |
| Dependency graph | Shared node evidence, critical/optional edges and budgets | Graph configuration is trusted; large graphs need resource planning |
| AHP | Experimental snapshot serving, readiness/liveness and protected detailed evidence | Shared bearer scope; HTTPS and request limits belong at the deployment boundary |

## Project status

AgentHealth is pre-1.0. Published software is v0.16.0; changes described as
implemented in source require a newer build. Planned target vocabulary is not
an availability guarantee. [Capability status](docs/phase-status.md) is generated
from [one machine-readable source](docs/status.json).

## Run the CLI

`ping` checks one endpoint, `check` executes configuration, and `doctor` shows
dimension-level evidence. Use `--format json` or `--format yaml` for automation.
The [CLI guide](docs/cli.md) defines flags and the [exit-code contract](spec/exit-codes.md)
defines statuses. Serving stays alive when monitored targets become unhealthy.

## Security principles

Treat configuration as trusted executable input. Review MCP commands, endpoints,
credential references and OAuth paths. TLS verification and redirect refusal
are mandatory in the supplied engine. Active probes need explicit authorization.
Redaction suppresses known credentials and raw diagnostic material; it cannot
identify every unknown or transformed secret. See [SECURITY.md](SECURITY.md) and
the [threat model](docs/threat-model.md).

## Limitations

In-process adapters are trusted code. A timed-out adapter can continue running
and retain a concurrency slot; enough stuck calls can exhaust capacity. Third-party
adapters need future out-of-process isolation. A shared process is not safe for
mutually untrusted tenant configurations.

Health declarations and safe-task results are target-provided evidence, not
independent proof of agent correctness. Gateway/router HTTP signals do not
establish real product interoperability. Official MCP/A2A SDK CI runs controlled
servers; broader frameworks/products remain validation work.

AHP reads a fixed configuration and serves snapshots; clients do not submit
commands or trigger checks. Detailed routes require a token, public routes still
disclose status, and authorized readers see topology. Remote deployment requires
HTTPS, ingress restrictions and rate limits. Kubernetes needs CNI-enforced egress
rules for the destinations in your reviewed configuration.

## Application integration

The agent adapter requires a specific read-only health document and an optional
bounded safe-task handler. Start with [reference middleware](docs/reference-middleware.md)
for Go, FastAPI, Express and LangGraph. These examples are application integration
helpers; the [Python SDK](docs/python-sdk.md) provides typed CLI access. The [JavaScript SDK](docs/javascript-sdk.md) is published on npm for Phase 16.

## Specification — Protocol — Implementation

The [specification](spec/README.md) defines health states, configuration, results
and adapter contracts. [AHP](spec/protocol.md) defines the experimental serving
binding. [Core](core/README.md) and [adapter guides](docs/README.md) describe the
reference implementation and its bounds.

## Documentation

- [Installation and verification](docs/installation.md)
- [Adapter and deployment guides](docs/README.md)
- [Examples](examples/README.md)
- [Current, next and later roadmap](ROADMAP.md)
- [Testing and stabilization](docs/stabilization.md)
- [Historical design/reference material](docs/project-reference.md)
- [Releases](https://github.com/TheAgentHealth/agenthealth/releases)
- [Website](https://theagenthealth.github.io)

## Contributing

Follow [CONTRIBUTING.md](CONTRIBUTING.md), [GOVERNANCE.md](GOVERNANCE.md) and the
[release policy](RELEASING.md). Security reports belong in private vulnerability
reporting as described by [SECURITY.md](SECURITY.md). New behavior must keep Go
validation, schemas, fixtures and documented contracts aligned.

## Passive GET

HTTP/API checks default to HEAD. For a read-only endpoint that returns 405 or
501 to HEAD, v0.15.0 accepts an explicit GET method. Omission preserves the
HEAD default.

<!-- spec-example: configuration -->
```yaml
version: v1
targets:
  - name: get-only-api
    type: http
    endpoint: http://localhost:8080/health
    checks: [reachability, protocol]
    http:
      method: GET
```

Passive GET checks status and headers without body matching. Body inspection
still requires an explicit functional check. The engine does not silently retry
HEAD as GET or follow redirects. Only configure GET against a read-only handler.

## Development validation

Use Go 1.26 or newer and install the Python test requirements. CI runs race tests,
coverage gating, schema validation, documentation checks and protocol suites.
Reference application handlers have separate FastAPI/LangGraph and Express tests.

```bash
go test -race ./...
python3 -m pytest tests -q -ra
python3 scripts/generate_status.py --check
```

See [stabilization](docs/stabilization.md) for fuzzing, graph profiles, real-product
smokes, Kubernetes versions and compatibility checks. Passing one suite does not
establish every planned integration; consult its exact scope and limitations.

Phase 16 introduced the [JavaScript / TypeScript SDK](docs/javascript-sdk.md) in v0.14.0: `@theagenthealth/agenthealth`, Node 22.14+ ESM, with a matching
verified CLI. See the [phase alignment audit](docs/phase-16-alignment.md).

## Model / LLM health (Phase 17)

The [model adapter](docs/model-adapter.md) checks an agent's supporting model
dependencies with one bounded passive model listing. OpenAI-compatible and
Anthropic dialects support required model IDs, rate-limit/overload diagnostics,
latency thresholds and opt-in minimal inference. Included in v0.15.0; v0.14.0 has no model adapter.

```bash
agenthealth ping llm http://localhost:11434/v1
agenthealth check examples/model-check/anthropic.yaml --format json
```

See [model examples](examples/model-check/README.md) and the
[alignment audit](docs/phase-17-alignment.md). Cloud signing, streaming and
broader product interoperability remain future work.

## Database health (Phase 18)

Released in v0.16.0: native PostgreSQL/MySQL login, Redis/MongoDB authenticated
PING, verified TLS defaults and environment password references. Fixed SQL
`SELECT 1` queries require explicit functional opt-in. These are supporting
dependency checks; minimal-query success does not establish application data
readiness. See the [guide](docs/database-adapter.md),
[alignment audit](docs/phase-18-alignment.md),
[disposable examples](examples/database-check/README.md) and
[publication verification](docs/releases/verification-v0.16.0.md).

## Contributors

Thanks to [Sharath K (@sharath568)](https://github.com/sharath568) for proposing
and implementing the initial Phase 17 model/LLM adapter in
[RFC #14](https://github.com/TheAgentHealth/agenthealth/issues/14) and
[PR #17](https://github.com/TheAgentHealth/agenthealth/pull/17). His authored
commit is preserved in the integrated implementation.
