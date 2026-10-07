# Contributing to AgentHealth

Thanks for your interest in contributing to AgentHealth, the reference implementation and specification maintained by the TheAgentHealth organization.

## Before you start

AgentHealth is pre-1.0 and under active design (see [Project Status](README.md#project-status)). Interfaces, schemas, and commands may still change. For anything non-trivial, please open an issue or discussion before writing a large PR.

## Ways to contribute

- Core engine implementation
- Protocol adapters (MCP, A2A, HTTP, model, database, vector store)
- SDKs (Python, JavaScript/TypeScript)
- Kubernetes and CI/CD integrations
- Specification design (see [spec/](spec/README.md))
- Interoperability testing across real agent stacks
- Documentation and examples
- Security reviews
- Testing

## Specification changes vs implementation changes

These follow different review bars:

- **Implementation changes** (CLI, core engine, adapters, SDKs, integrations): standard pull request review by maintainers.
- **Specification changes** (anything in [spec/](spec/README.md), especially [health-model.md](spec/health-model.md) and [protocol.md](spec/protocol.md)): require an RFC-style discussion first, since downstream adapters and SDKs depend on spec stability. See [GOVERNANCE.md](GOVERNANCE.md).

## Development workflow

1. Open an issue describing the problem or proposal (skip for small fixes/typos).
2. Fork the repository and create a branch from `main`.
3. Make your change, including tests where applicable.
4. Ensure existing tests pass and add new tests for new behavior.
5. Open a pull request against `main` describing what changed and why.
6. Address review feedback. PRs are merged once approved by a maintainer **and CI is green**.

## CI checks

Every pull request runs the [CI workflow](.github/workflows/ci.yml):

- **Spec schema tests** (`tests/spec/`) — validates `spec/schemas/*.json` against fixture examples (both valid and intentionally-invalid), so a spec/schema change that breaks the contract fails CI instead of being caught in manual review.
- **Markdown link check** (`scripts/check_markdown_links.py`) — verifies every relative markdown link and `#anchor` in the repository actually resolves.
- **Doc example validation** (`scripts/validate_doc_examples.py`) — validates fenced JSON/YAML examples in README.md, ROADMAP.md, and spec/*.md that are explicitly marked `<!-- spec-example: result -->` or `<!-- spec-example: configuration -->` against the schemas, so prose examples can't silently drift from the schemas the way they did before this check existed. An example can only be exempted with `<!-- spec-example: skip reason="..." -->` — a documented, intentional reason is required.

The **Go core tests** job also checks formatting, runs `go vet ./...`, and runs `go test -race ./...`.

Run the checks locally before opening a PR:

```bash
go test -race ./...
go vet ./...
pip install -r tests/requirements.txt
python3 -m pytest tests/spec -v
python3 scripts/check_markdown_links.py
python3 scripts/validate_doc_examples.py
```

If you change `spec/configuration.md` or `spec/result-schema.md`, update `spec/schemas/*.json` and add/update fixtures under `tests/spec/fixtures/` in the same PR. If you add or change a canonical example in README.md, ROADMAP.md, or spec/*.md, mark it with `<!-- spec-example: result -->` or `<!-- spec-example: configuration -->` so CI validates it too. The [Phase 2 Go core](core/README.md) test suite runs alongside these checks.

## Commit messages

Use clear, descriptive commit messages. Conventional prefixes (`feat:`, `fix:`, `docs:`, `spec:`, `chore:`) are encouraged but not yet strictly enforced.

## Code of conduct

Participation in this project is governed by the [Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold it.

## Security issues

Do not open public issues for security vulnerabilities. See [SECURITY.md](SECURITY.md) for how to report them.

## Questions

Open a GitHub Discussion or issue on [github.com/TheAgentHealth/agenthealth](https://github.com/TheAgentHealth/agenthealth).
