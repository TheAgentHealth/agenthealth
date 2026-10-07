# A2A peer examples

These configurations require your own running A2A 0.3.0 JSON-RPC peer.
Adjust the endpoint and expected `health` skill ID to match its agent card.
They do not start a server. Run the commands from the repository root with
AgentHealth v0.3.0 or newer.

| Configuration | Purpose |
|---|---|
| [agenthealth.yaml](agenthealth.yaml) | Default passive discovery, protocol/authentication checks, a required skill, and latency threshold |
| [bearer.yaml](bearer.yaml) | Passive checks using the `AGENT_TOKEN` environment reference |
| [custom-card.yaml](custom-card.yaml) | A card at a custom absolute URL on the configured origin |
| [functional.yaml](functional.yaml) | An explicitly safe text interaction with passive checks |

## Passive discovery

The standard card location is `/.well-known/agent-card.json` at the configured
origin. The card's `url` must point to a JSON-RPC endpoint on that same origin.
Authentication/protocol checks use read-only `tasks/get` lookups with fresh IDs;
a valid task-not-found response passes without creating a task.

```bash
agenthealth ping a2a http://localhost:9000
agenthealth doctor a2a http://localhost:9000
agenthealth check examples/a2a-check/agenthealth.yaml --format json
agenthealth doctor examples/a2a-check/agenthealth.yaml
agenthealth check examples/a2a-check/custom-card.yaml --format yaml
```

The custom-card example expects `/metadata/agent-card.json`. Change both URLs
when moving to a different origin. A missing required `health` skill yields
`DEGRADED` and exit code 1. The capability check inspects metadata; it does not
execute that skill. See [capability expectations](../../docs/a2a-adapter.md#discovery-and-passive-checks)
for optional boolean capability requirements.

## Bearer credentials

Set `AGENT_TOKEN` in the process environment before running the bearer example;
keep the credential value out of configuration files. Replace its HTTPS endpoint
with the origin of your authenticated peer. Missing/empty references fail
configuration before networking; HTTP 401/403 yields `MISCONFIGURED` and exit
code 4. These checks validate acceptance for the passive lookup.

```bash
agenthealth check examples/a2a-check/bearer.yaml --format json
```

## Minimal interaction

Use the functional example with a peer supporting default `text/plain` input
and output. Confirm its prompt is non-destructive for your agent before using
`safe: true`; the declaration is an operator assertion. The probe sends one
blocking `message/send` request and never retries it. Add the same `auth` block
as the bearer example if your peer requires a token.

```bash
agenthealth check examples/a2a-check/functional.yaml --format json
```

Valid agent text replies or completed tasks pass. Failed/rejected/canceled tasks
are `UNHEALTHY` (exit 2); pending tasks and input requirements are `UNKNOWN`
(exit 5). The adapter does not poll, continue, or automatically cancel tasks,
and does not expose the reply body. A successful interaction validates completion,
not answer semantics. See the [adapter guide](../../docs/a2a-adapter.md) for the
full supported scope, limits, and classifications.

For a current v1 peer, use [the v1 configuration](v1.yaml). Legacy examples
explicitly select 0.3.0. Unconfigured ping/doctor now select v1.
