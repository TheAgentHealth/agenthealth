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
6. Address review feedback. PRs are merged once approved by a maintainer.

## Commit messages

Use clear, descriptive commit messages. Conventional prefixes (`feat:`, `fix:`, `docs:`, `spec:`, `chore:`) are encouraged but not yet strictly enforced.

## Code of conduct

Participation in this project is expected to be respectful and harassment-free. A formal `CODE_OF_CONDUCT.md` will be added as the project matures; until then, standard open-source community norms apply.

## Security issues

Do not open public issues for security vulnerabilities. See [SECURITY.md](SECURITY.md) for how to report them.

## Questions

Open a GitHub Discussion or issue on [github.com/TheAgentHealth/agenthealth](https://github.com/TheAgentHealth/agenthealth).
