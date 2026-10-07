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

This policy covers the `agenthealth` core engine, CLI, official adapters, official SDKs (Python/JavaScript), and official Docker images/Helm charts published from this repository. Community adapters outside this repository are out of scope and should publish their own security policy.

## Future: Agent Health Protocol (AHP)

Once [AHP](spec/protocol.md) moves beyond proposed/experimental status, its own security requirements (information disclosure, authentication of health endpoints, dependency redaction, rate limiting) will be defined in [spec/protocol.md](spec/protocol.md) and enforced through conformance testing.
