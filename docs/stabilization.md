# Stabilization and validation

Prioritize current security boundaries and interoperability evidence before
expanding SDKs and model/database/vector adapters. Fixes may release under existing
review gates; this plan does not impose a blanket release freeze.

## Current work

- Reconcile SECURITY.md with released experimental serving; maintain the threat model.
- Preserve default methods and add explicit passive GET selection in source.
- Describe gateway/router as HTTP health signals and agent health as an application contract.
- Generate the capability table from `docs/status.json` and reject drift in CI.
- Gate production Go statement coverage at 80%; collect the race-enabled coverage profile.
- Publish pytest JUnit reports and summarized skip reasons, including dedicated SDK suites.
- Run parser fuzz smoke jobs and collect DAG benchmark/allocation profiles.

See the [local validation record](stabilization-validation.md) for measured results
and the exact limits of the evidence.

## Running validation

Use a supported installed Go toolchain (1.23 or newer):

```bash
go vet ./...
go test -race -coverprofile=coverage.out ./...
python3 scripts/check_coverage.py coverage.out
python3 -m pytest tests -q -ra --junitxml=pytest.xml
python3 scripts/summarize_test_skips.py pytest.xml test-skips.md
python3 scripts/generate_status.py --check
python3 scripts/check_markdown_links.py
python3 scripts/validate_doc_examples.py
```

Coverage excludes executable examples from the aggregate gate; they are still
compiled and tested. Coverage is an exercised-statement metric, not proof of
security. Increase the floor as meaningful tests improve it.

Fuzz targets cover YAML configuration, MCP JSON-RPC, SSE framing, A2A cards and
OAuth discovery. Inputs are bounded and transports are local or synthetic. A
short CI run catches regressions but does not replace longer scheduled campaigns.

Run DAG benchmarks with allocation and memory profiles:

```bash
go test ./core -run '^$' -bench BenchmarkLargeDAG -benchmem -count=3
go test ./core -run '^$' -bench BenchmarkLargeDAG -benchtime=1x -memprofile=dag-memory.out
go tool pprof -top dag-memory.out
```

## Interoperability evidence

Existing CI tests controlled servers using pinned official MCP and A2A SDKs,
current and legacy. Those are real SDK implementations, not universal product
coverage. Add independently maintained MCP servers, A2A frameworks and actual
gateway/router products with pinned versions and repeatable failure scenarios.
Record successful and unsuccessful runs; never mark interoperability complete from
HTTP fixtures alone.

The original Kubernetes runtime gate pins 1.32.2. An additional CI matrix now
configures 1.31.9 and 1.33.1 using digest-pinned kind v0.28.0 images. Successful
runtime evidence is required before claiming support for those versions.
Keep current release CI job identities intact when adding a version matrix, since
publication requires explicit successful checks for the tagged source.

## Remaining acceptance criteria

- Sustained fuzz campaigns with retained regression corpus and crash triage.
- Goroutine-leak accounting across cancellation, stdio teardown, cleanup and repeated runs.
- Broader DNS/TLS/reset/slow-body fault scenarios beyond existing timeout/redirect tests.
- Broader graph shapes and sustained resource budgets beyond the current 100/1,000/5,000-node shared-DAG profiles.
- Real framework/product fixtures with exact dependency versions.
- Multiple Kubernetes versions with runtime results, rather than rendering alone.
- Broader upgrade/rollback configurations and deployment secrets beyond the current unchanged-HTTP old/current/old binary smoke sequence.
- Out-of-process third-party adapters, scoped AHP authorization and external security audit.

Configuration compatibility must test actual released binaries: validating an old
fixture against the current schema alone does not establish upgrade or rollback.

## Product smoke coverage added in source

Dedicated CI now exercises the maintained MCP filesystem server at 2026.8.31,
Agentgateway v1.6.0 readiness, LiteLLM 1.104.2 process liveness, compiled LangGraph
reference handlers, and
configuration compatibility using released v0.9.0/v0.12.0 binaries. These focused
checks do not establish router decision correctness, Agentgateway route decisions,
full A2A framework interoperability or deployment-secret migration safety.
