# Threat model

This model covers the current CLI, Go engine, official adapters, experimental
AHP server, and Kubernetes deployments. It is a design review, not a penetration
test or certification. Review it when changing configuration, discovery,
authentication, scheduling, output, or adapter execution.

## Assets and trust boundaries

Assets include environment credentials, OAuth access/refresh tokens, token files,
local execution privileges, internal services, dependency topology, availability,
and the integrity of health evidence. Operators control configuration and the
binary. Remote endpoints, protocol documents, and response bodies are untrusted.
In-process adapters are trusted code. AHP clients are remote readers of a fixed
configuration, not configuration authors.

Configuration review is mandatory before running external pull-request examples.
An active-check flag authorizes a probe; it does not sandbox the configured
program or make a remote GET handler safe. Health declarations are evidence
from the target, not independent proof of correct behavior.

## Threats, controls and residual exposure

| Threat | Implemented controls | Operator responsibility and remaining exposure |
|---|---|---|
| SSRF and cloud metadata | TLS verification, redirect refusal, URL validation; A2A discovery restricts RPC URLs to the configured origin and OAuth discovery validates configured trust relationships | No global destination allowlist or IP-range block. Trusted endpoints and OAuth issuer configuration can contact loopback, private addresses and metadata services. Restrict egress at the network boundary, including IPv6 and DNS rebinding scenarios. |
| MCP stdio command execution | Explicit commands and selected environment references; bounded session lifecycle | Even passive stdio checks execute code. Review executables, arguments, PATH and supplied credentials. Use dedicated low-privilege processes/containers; a stdio program is not sandboxed by the engine. |
| OAuth token files | Bounded reads, regular-file checks, platform-specific private-file protections; temporary-file writes and rename; newly created directories use 0700 | Treat paths and parent directories as trusted. Do not share writable parents across users or tenants. Mount only the required private storage, review ownership/ACLs, and prevent concurrent writers. File checks do not establish a complete filesystem sandbox or eliminate every race. |
| Private-network access | Explicit targets; discovered dependencies never automatically become targets | Private access is often intentional. Deploy near required services with a least-privilege service identity and egress policy. A health checker with production credentials is a privileged workload. |
| AHP topology disclosure | Detailed dependency/capability endpoints require a configured bearer token; disabled without one. Public health/readiness routes return aggregate status and observation time | Public status is still information disclosure. Restrict ingress where needed. Authorized readers receive topology and evidence; secret redaction is not topology anonymization. |
| Bearer-token scope | Token syntax/length validation and constant-time comparison of token hashes | One AHP token grants all detailed evidence for that server. There are no per-target scopes, roles or tenant claims. Use HTTPS at a trusted proxy, restrict readers, rotate by updating the secret and restarting the process, and separate configurations with different access policies. |
| Expensive graphs and endpoints | Run/check budgets, bounded concurrency and response bodies; AHP refreshes serially and serves snapshots without request-triggered probes | Large graphs, repeated active probes, expensive remote work and large snapshot responses can consume resources. Review graph size and refresh policy; apply CPU/memory limits and proxy request/connection limits. Cancellation cannot undo remote side effects. No built-in AHP rate limiter. |
| Uncooperative in-process adapters | Caller deadlines, panic normalization, bounded engine/per-run slots retained until calls return | Go cannot forcibly terminate a goroutine. A stuck call can permanently retain a slot, resource or cleanup gate; enough calls can exhaust useful capacity. Only run trusted adapters in process. Third-party adapter execution needs a future out-of-process protocol and OS isolation. |
| Output/redaction | Canonical diagnostics, suppression of raw errors/bodies, known-credential redaction, format validation and terminal-control filtering | Unknown secrets, encoded/transformed credentials and sensitive names may evade redaction. Custom adapters can log outside engine control. Review collected output and supply secrets to the Redactor for caller-built results. Do not treat redaction as a DLP guarantee. |
| Kubernetes egress | Non-root/container hardening and Secret references in supplied deployments | Charts do not install a destination-specific egress policy. Apply a CNI-enforced NetworkPolicy allowing only DNS and required destinations; account for OAuth issuers and MCP dependencies. ClusterIP is not an authorization boundary. See the [example policy](../examples/kubernetes/network-policy.yaml). |
| Multi-tenant deployment | Fixed server configuration and protected detailed routes | No mutually untrusted tenant isolation, per-tenant quotas, credential separation or per-tenant authorization. Do not accept arbitrary tenant configurations in a shared process. Use separate workload identities, processes, secrets, volumes and network policies; container separation alone does not establish a complete tenant boundary. |

## Serving and deployment requirements

AHP is implemented since v0.9.0 and remains experimental. CLI defaults bind to
loopback; Helm serve deployments expose the process through Kubernetes networking.
Terminate remote HTTPS, restrict ingress, and enforce rate limits at a proxy.
Use passive configurations for continuous monitoring unless repeated active
probes are deliberately authorized. The server does not accept remote target or
command submission. An unauthenticated request cannot directly select arbitrary
network destinations or start a graph run.

Do not run an untrusted configuration merely because detailed AHP routes are
protected. Ingress authorization and outbound execution privileges are separate
boundaries. Give every deployment only the credentials, mounts and network access
its reviewed configuration needs.

## Validation and open work

Existing tests exercise authorization, freshness, redirect/TLS behavior, protocol
bounds, credential redaction and adapter cancellation. Official MCP/A2A SDK CI
adds controlled interoperability; it does not establish compatibility with every
product. The [stabilization plan](stabilization.md) tracks fuzzing, graph profiling,
network faults, broader interoperability and upgrade/rollback evidence.

Future work includes destination policy enforcement, scoped authorization,
out-of-process adapter execution and an external security audit. Do not advertise
multi-tenant safety or production certification before those guarantees exist.
