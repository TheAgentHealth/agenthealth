# Model / LLM examples

Requires a Phase 17 build (released in v0.15.0), a running model API and
credentials/model IDs appropriate to that server. These are templates, not
claims of tested hosted-provider compatibility.

| Configuration | Setup and behavior |
|---|---|
| [OpenAI-compatible](openai.yaml) | Set `OPENAI_API_KEY`; change the required model to one your account advertises. Passive listing only. |
| [Anthropic](anthropic.yaml) | Set `ANTHROPIC_API_KEY`; adjust required model ID. Uses `x-api-key`. |
| [Ollama](ollama.yaml) | Start a local OpenAI-compatible `/v1` endpoint and install the configured model. Includes active inference. |
| [Functional](functional.yaml) | Set `OPENAI_API_KEY`; choose a model supporting `max_completion_tokens`. May incur charges. |
| [Shared graph](graph.yaml) | Start the configured agents and path handler plus local model API. Passive model evidence is shared by agent dependencies. |

```bash
agenthealth ping llm http://localhost:11434/v1
agenthealth check examples/model-check/openai.yaml --format json
agenthealth doctor examples/model-check/anthropic.yaml
```

Functional checks require explicit opt-in and are never retried. The requested
token bound depends on the server honoring it; reported excess usage is degraded.
See the [model guide](../../docs/model-adapter.md) and
[alignment audit](../../docs/phase-17-alignment.md).
