# Specification Security Requirements

> Status: Draft. Defines the security requirements that the Agent Health Specification (and any conforming implementation) MUST satisfy. This is distinct from [SECURITY.md](../SECURITY.md), which covers reporting vulnerabilities in this repository's own code.

## Why this is part of the specification

Health tooling routinely touches credentials and production infrastructure. If the specification does not constrain what a conforming implementation is allowed to expose, every adapter and SDK re-derives its own, possibly inconsistent, possibly unsafe answer. These requirements apply to **any** implementation of AHS/AHP, not just this repository's reference implementation.

## Requirements

A conforming implementation MUST:

- never include secrets, API keys, tokens, or credentials in a [result](result-schema.md), log line, or diagnostic message, even in verbose/debug modes;
- redact authentication material before it reaches any output format (terminal, JSON, AHP responses);
- treat [functional checks](health-model.md#functional-health) as non-destructive/safe by default;
- require explicit, opt-in configuration before running a check capable of side effects (see [Passive vs Active Checks](../docs/project-reference.md#passive-vs-active-checks));
- verify TLS by default when checking HTTPS/TLS-based targets, and require explicit opt-out rather than defaulting to insecure;
- support configurable timeouts on every check, so a slow/unresponsive target cannot hang the caller indefinitely;
- avoid amplifying checks into a denial-of-service vector against the target (e.g. bounded concurrency, bounded retries).

A conforming implementation SHOULD:

- distinguish information that is safe to expose publicly (e.g. `status: DEGRADED`) from information that requires authorization to view (e.g. which specific dependency/endpoint failed, internal hostnames);
- make the distinction between active and passive checks visible in configuration and output, not just in documentation.

## Relationship to AHP

The experimental [Agent Health Protocol](protocol.md) defines bearer authorization for detailed health evidence and public aggregate summaries. See [protocol.md § Security](protocol.md#execution-and-security).

## Relationship to this repository's own security policy

Requirements in this document constrain what the specification allows a conformant implementation to do. [SECURITY.md](../SECURITY.md) in the repository root covers how to report a vulnerability in this repository's actual code. Both apply, but they answer different questions.

## Phase 10 graph safety

Graph references resolve only explicitly configured IDs; advertised names never authorize endpoints or topology discovery. The engine rejects missing references, duplicate IDs, cycles and excessive depth/tree projections before adapter work. Node budgets, bounded per-run concurrency and existing active opt-in/non-retry rules apply. Shared active probes execute once per ID per run.

IDs, names and relationship structure can reveal sensitive topology. Avoid secrets in identifiers and restrict result access appropriately. Known credential values in IDs are replaced by `redacted`, which can collapse identities in sanitized output; consumers must not infer shared identity from that placeholder. Experimental AHP topology access requires bearer authorization; finer scopes remain planned. See the [graph contract](../docs/rfcs/phase-10-dependency-graph.md).
