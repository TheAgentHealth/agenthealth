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

Run both locally before opening a PR:

```bash
pip install -r tests/requirements.txt
python3 -m pytest tests/spec -v
python3 scripts/check_markdown_links.py
```

If you change `spec/configuration.md` or `spec/result-schema.md`, update `spec/schemas/*.json` and add/update fixtures under `tests/spec/fixtures/` in the same PR. Once a Go core engine lands ([Phase 2](ROADMAP.md#phase-2--core-engine)), its own test suite will be added as additional required CI jobs alongside these.

## Commit messages

Use clear, descriptive commit messages. Conventional prefixes (`feat:`, `fix:`, `docs:`, `spec:`, `chore:`) are encouraged but not yet strictly enforced.

## Code of conduct

Participation in this project is governed by the [Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold it.

## Security issues

Do not open public issues for security vulnerabilities. See [SECURITY.md](SECURITY.md) for how to report them.

## Questions

Open a GitHub Discussion or issue on [github.com/TheAgentHealth/agenthealth](https://github.com/TheAgentHealth/agenthealth).
