# Phase 19: read-only vector-store health

[RFC #57](https://github.com/TheAgentHealth/agenthealth/issues/57) records the
user-authorized additive v1 extension. Initial dialects are Qdrant REST and
Pinecone dense-vector data-plane REST, version 2025-10.

`vector-store` already exists in the target vocabulary. Its required
`vector_store` block selects `provider: qdrant` with `collection`, or
`provider: pinecone` with an optional `namespace` (the endpoint selects the
index host). Endpoints are absolute HTTP(S) bases, optionally with a path
prefix, without userinfo, queries or fragments. Collection names are bounded
ASCII identifiers; named/sparse vectors and arbitrary query text are excluded.

Existing dimensions, health states, result v1 and exits remain unchanged.
Metadata access is passive even where the API requires POST. Functional work
requires both `checks` containing `functional` and `functional.safe: true` plus
1–4096 finite numeric vector components, at least one nonzero. Query cost is
explicit: one top-1 request, no payload, metadata or vector returns, no writes,
no embedding provider and no active retries. An empty successful result is
healthy; semantic relevance and corpus completeness are not asserted.

Credentials use `auth.bearer_env` as a generic existing environment reference;
Qdrant sends `api-key`, Pinecone sends `Api-Key`. Headers are validated before
network access. Verified TLS, redirect denial, context budgets and response
tracking use the engine transport. Responses are bounded to 1 MiB and error
bodies are discarded. Only engine-owned diagnostics leave the adapter; metadata,
query vectors and retrieved objects never become result evidence. Per-run
metadata is cached only inside one target node.

The [alignment audit](../phase-19-alignment.md) covers phases 0–26. Additional
providers and product matrices remain future work rather than inferred from
successful HTTP mocks or PostgreSQL login.

For Go/schema parity, empty unused provider fields are normalized as omitted:
Qdrant may specify `namespace: ""`; Pinecone may specify `collection: ""`.
Nonempty values in those unused fields are rejected.
