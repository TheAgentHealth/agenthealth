# Specification Security Requirements

> Status: Draft. Defines the security requirements that the Agent Health Specification (and any conforming implementation) MUST satisfy. This is distinct from [SECURITY.md](../SECURITY.md), which covers reporting vulnerabilities in this repository's own code.

## Why this is part of the specification

Health tooling routinely touches credentials and production infrastructure. If the specification does not constrain what a conforming implementation is allowed to expose, every adapter and SDK re-derives its own, possibly inconsistent, possibly unsafe answer. These requirements apply to **any** implementation of AHS/AHP, not just this repository's reference implementation.

## Requirements

A conforming implementation MUST:

- never include secrets, API keys, tokens, or credentials in a [result](result-schema.md), log line, or diagnostic message, even in verbose/debug modes;
- redact authentication material before it reaches any output format (terminal, JSON, future AHP responses);
- treat [functional checks](health-model.md#functional-health) as non-destructive/safe by default;
- require explicit, opt-in configuration before running a check capable of side effects (see [Passive vs Active Checks](../README.md#passive-vs-active-checks));
- verify TLS by default when checking HTTPS/TLS-based targets, and require explicit opt-out rather than defaulting to insecure;
- support configurable timeouts on every check, so a slow/unresponsive target cannot hang the caller indefinitely;
- avoid amplifying checks into a denial-of-service vector against the target (e.g. bounded concurrency, bounded retries).

A conforming implementation SHOULD:

- distinguish information that is safe to expose publicly (e.g. `status: DEGRADED`) from information that requires authorization to view (e.g. which specific dependency/endpoint failed, internal hostnames);
- make the distinction between active and passive checks visible in configuration and output, not just in documentation.

## Relationship to AHP

Once the [Agent Health Protocol](protocol.md) moves beyond proposed/experimental status, it will additionally need to define authentication/authorization **of the health endpoint itself**, since AHP exposes health over the network rather than only running checks locally. See [protocol.md § Security](protocol.md#security).

## Relationship to this repository's own security policy

Requirements in this document constrain what the specification allows a conformant implementation to do. [SECURITY.md](../SECURITY.md) in the repository root covers how to report a vulnerability in this repository's actual code. Both apply, but they answer different questions.
