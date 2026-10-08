# Security Policy

## Reporting a vulnerability

Please do not open a public GitHub issue for security vulnerabilities. Instead, report them privately via GitHub's private vulnerability reporting feature on [github.com/TheAgentHealth/agenthealth](https://github.com/TheAgentHealth/agenthealth/security/advisories/new), or by contacting the maintainers directly.

Include as much detail as possible:

- affected version/commit
- reproduction steps
- potential impact
- any suggested fix

We will acknowledge reports and work with you on disclosure timing. Since the project is pre-1.0, there is no formal bug bounty program at this time.

## Supported versions

AgentHealth is currently pre-1.0 (see [Project Status](README.md#project-status)). Security fixes are applied to the `main` branch; there is no long-term support branch yet. This will be revisited once a stable 1.0 is released — see [Phase 26 — Production Hardening](ROADMAP.md#phase-26--production-hardening) in the roadmap, which includes a security audit before 1.0.

## Trusted configuration

AgentHealth configuration is trusted executable input. Review endpoints, MCP stdio commands and arguments, environment credential references, and OAuth token-file paths before execution. Never run a configuration from an untrusted source, including an external pull request, without reviewing it first.

Configuration can execute a configured stdio program with selected environment credentials, contact loopback or private-network services (including cloud metadata endpoints), acquire OAuth tokens, and write configured token files. Passive checks and active-check opt-in do not sandbox subprocesses or restrict network destinations. Run with only the credentials, filesystem permissions, and network access needed for the intended checks.

AHP service mode remains planned. Its design must address network egress and target allowlists, stdio enablement, authorization, topology redaction, freshness, and rate limiting before accepting remote requests.

## Security principles

Because AgentHealth routinely interacts with credentials and production infrastructure, the project follows these principles (see also [Security Principles](README.md#security-principles)):

- Never print secrets or authentication material in output, logs, or diagnostics.
- Redact credentials even in verbose/debug modes.
- Health checks are non-destructive by default.
- Checks that could have side effects (invoking a tool, running inference, writing data) must be explicit, opt-in **active** checks — see [Passive vs Active Checks](README.md#passive-vs-active-checks).
- Secure defaults: TLS verification enabled by default, no insecure fallback without explicit opt-in.
- Configurable timeouts to avoid checks hanging indefinitely or being used as a denial-of-service vector against targets.
- Adapters must document the permissions/credentials their checks require (see [Adapter Contract](spec/adapter-spec.md)).

## Scope

This policy covers the `agenthealth` core engine, CLI, official adapters, official SDKs (Python/JavaScript) and official Docker images/Helm charts when officially published from this repository. Community adapters outside this repository are out of scope and should publish their own security policy.

## Future: Agent Health Protocol (AHP)

Once [AHP](spec/protocol.md) moves beyond proposed/experimental status, its own security requirements (information disclosure, authentication of health endpoints, dependency redaction, rate limiting) will be defined in [spec/protocol.md](spec/protocol.md) and enforced through conformance testing.
