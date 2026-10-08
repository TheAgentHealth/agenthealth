# Model health adapter

The model adapter checks model and LLM inference endpoints used by agents
([Phase 17](../ROADMAP.md#phase-17--model--llm-adapters)). It handles `model`
and `llm` targets and supports two API dialects:

- `openai` (default): OpenAI-compatible APIs, including OpenAI, vLLM, Ollama,
  LiteLLM and other servers exposing `/models` and `/chat/completions`.
- `anthropic`: the Anthropic API (`/models` and `/messages`).

```bash
agenthealth ping llm http://localhost:11434/v1
agenthealth check examples/model-check/openai.yaml
agenthealth doctor examples/model-check/anthropic.yaml
```

Streaming, embeddings, tool calls, cloud-provider signing (Amazon Bedrock,
Google Vertex AI) and Azure `api-key` headers are not implemented.

## Configuration options

| Option | Behavior |
|---|---|
| `endpoint` | API base URL including its version path, such as `https://api.openai.com/v1`; a configured query is preserved on every request |
| `auth.bearer_env` | Credential reference: sent as `Authorization: Bearer` for `openai`, and as `x-api-key` for `anthropic` |
| `model.api` | `openai` (also the empty default) or `anthropic` |
| `model.required_models` | Unique, nonblank model IDs that the model listing must contain |
| `model.functional` | Opt-in minimal inference; requires `safe: true`, `model`, `prompt` (at most 4 KiB) and explicit `functional` in `checks` |
| `model.functional.max_output_tokens` | Output token bound from 1 to 1024; default 16 |
| `model.functional.token_parameter` | `max_tokens` (default) or `max_completion_tokens`; `anthropic` always uses `max_tokens` |

Explicit `checks` must include `capability` when `required_models` is set. See
the [configuration contract](../spec/configuration.md#model-expectations-phase-17-draft).

## Passive checks

Reachability, authentication, protocol and capability share a single
`GET <endpoint>/models` request per target run. The `anthropic` dialect adds
`limit=1000` and the `anthropic-version: 2023-06-01` header. Reachability
records `dns`, `tcp`, `tls` and `http` transport steps, and any HTTP response
establishes reachability. Later checks classify the response:

| Response | Status | Code |
|---|---|---|
| 401 or 403 | `MISCONFIGURED` | `model_auth` |
| 429 | `DEGRADED` | `model_rate_limit` |
| 529 | `DEGRADED` | `model_overloaded` |
| Other non-2xx, including a wrong base path | `UNHEALTHY` | `model_http` |
| Not a JSON object with a `data` array of models with nonblank `id` | `UNHEALTHY` | `model_protocol` |
| Listing larger than 4 MiB | `UNKNOWN` | `model_limit` |
| A required model missing | `UNHEALTHY` | `model_required` |
| A required model missing from a listing that reports `has_more` | `UNKNOWN` | `model_limit` |

A listed model means that the API advertises it to this credential. It does
not prove that the model can serve requests; use the functional check for that.

## Minimal inference (active)

The functional check sends one non-streaming request with one user message:
`POST /chat/completions` for `openai` or `POST /messages` for `anthropic`. It
runs only when `functional` is listed in `checks`, and the engine never retries
it. The prompt should ask for a short reply and must not trigger tools or
external actions; `safe: true` records that the operator reviewed it.

Responses are classified as for listings, except that 404 yields
`model_unavailable` (unknown or unserved model) and other errors yield
`model_functional`. A 2xx response must contain a non-empty `choices` array
(`openai`) or a `message` with `content` (`anthropic`).

Each run is a billable request. Keep `max_output_tokens` small and run
functional checks less often than passive ones.

### Output token bound

OpenAI-compatible servers disagree on the output limit field. Ollama 0.40, for
example, ignores `max_completion_tokens` and honors `max_tokens`, while
current OpenAI reasoning models reject `max_tokens`. The default `max_tokens`
fails visibly where it is rejected rather than generating without a limit. Set
`token_parameter: max_completion_tokens` for such models.

When a response reports more output tokens (`usage.completion_tokens` or
`usage.output_tokens`) than `max_output_tokens`, the server did not enforce the
bound, and the check reports `DEGRADED` with code `model_token_limit`. Switch
`token_parameter` to fix it.

## Safety

- Error response bodies are never read. Successful bodies are bounded (4 MiB
  for listings, 1 MiB for inference) and never printed.
- Credentials come only from environment references, are validated as header
  values before network access, and are redacted from all output.
- TLS is verified and redirects are not followed, as in all adapters.

## Validation

Unit tests cover both dialects, every classification above, shared listings,
no retries for inference, token parameter selection and redaction. The adapter
was also run against Ollama 0.40 with `qwen2.5:0.5b`. That run covered passive
checks, required-model detection, `model_unavailable` for an unknown model, a
bounded inference, and `model_token_limit` detection when
`max_completion_tokens` was ignored. It has not been run against the hosted
OpenAI or Anthropic APIs; their request shapes follow the public API references.
