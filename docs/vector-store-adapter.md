# Vector-store adapter

Phase 19 implements Qdrant REST and Pinecone dense-vector data-plane REST in
source. See [examples](../examples/vector-check/README.md),
[RFC #57](https://github.com/TheAgentHealth/agenthealth/issues/57) and the
[phase alignment audit](phase-19-alignment.md).

<!-- spec-example: configuration -->
```yaml
version: v1
targets:
  - name: retrieval
    type: vector-store
    endpoint: https://qdrant.example.com
    auth:
      bearer_env: VECTOR_API_KEY
    vector_store:
      provider: qdrant
      collection: documents
    thresholds:
      latency_ms: 1000
```

The endpoint is an HTTP(S) base with an optional path prefix. Userinfo, query
strings and fragments are rejected. Qdrant requires a collection identifier
of 1–255 ASCII letters, digits, underscores or hyphens, starting with a letter
or digit. Pinecone uses `provider: pinecone`, the actual **index host**, no
`collection`, and an optional namespace of at most 512 characters. A configured
namespace must appear in index statistics for Capability to pass. Omit it to
check index availability, including a new empty index. Empty namespace denotes
the default namespace and does not require it to be present in statistics.
The existing target vocabulary still accepts placeholder targets without options;
the implemented adapter reports them as MISCONFIGURED before network access.

`auth.bearer_env` resolves an environment variable; the wire header is `api-key`
for Qdrant and `Api-Key` for Pinecone. The reference name is retained for v1
compatibility; no bearer Authorization header is sent. Pinecone sends
`X-Pinecone-Api-Version: 2025-10`. Never put literal credentials in endpoints.
HTTPS certificates are verified and redirects are not followed. Explicit HTTP
is intended for controlled local networks and transmits credentials in plaintext.

| Dimension | Evidence |
|---|---|
| Configuration | Provider/options, endpoint and credential header safety |
| Reachability | HTTP response to metadata request, including error responses |
| Authentication | Metadata request accepted; 401/403 is MISCONFIGURED |
| Protocol | Valid provider envelope and dense-vector metadata |
| Capability | Collection/index metadata; Qdrant green is healthy, yellow/grey degraded, red unhealthy; configured Pinecone namespace exists |
| Functional | Explicit bounded top-1 vector query with valid result envelope |
| Latency | Engine measures Reachability and applies existing thresholds |

Qdrant uses `GET /collections/{collection}` and
`POST /collections/{collection}/points/query`. Pinecone uses read-only
`POST /describe_index_stats` and `POST /query`; POST metadata does not make
that read-only operation an active check. Passive dimensions share one bounded
metadata response per target run. HTTP 404 is UNHEALTHY, 429 DEGRADED, other
unexpected HTTP statuses UNHEALTHY. Invalid JSON/envelopes are UNHEALTHY;
responses over 1 MiB are UNKNOWN. Transport errors use engine classification.
Error bodies, retrieved objects and vector values never enter output.

## Opt-in query

<!-- spec-example: configuration -->
```yaml
version: v1
targets:
  - name: retrieval-query
    type: vector-store
    endpoint: https://qdrant.example.com
    checks: [configuration, reachability, authentication, protocol, capability, functional, latency]
    auth:
      bearer_env: VECTOR_API_KEY
    vector_store:
      provider: qdrant
      collection: documents
      functional:
        safe: true
        vector: [1, 0, 0]
```

Both explicit `functional` selection and `safe: true` are required. Supply
1–4096 finite numeric components, at least one nonzero, matching the metadata
vector dimension. A mismatch is MISCONFIGURED before query transmission.
The adapter requests one result with payload/metadata/vector returns disabled.
Queries never retry, never create or modify collections, and never call an
embedding service. A successful empty result is healthy; no relevance, accuracy,
expected-document or corpus-completeness claim follows. Read charges may apply.

## Integration and limits

Use `vector-store` as a supporting dependency of a direct agent; composite
agent and communication-path evidence remain separate. Nested dependencies,
shared graph nodes, check/doctor/serve, JSON/YAML, AHP and both CLI-wrapper SDKs
reuse the existing engine contracts. `ping vector-store <url>` cannot select a
provider/collection; use a configuration file. Kubernetes can inject the
referenced environment variable from a Secret; include all probe work in the
configured budget. Vector failure affects readiness, not AHP `/live`.

Controlled interoperability covers Qdrant 1.15.5. Pinecone uses official API
contract fixtures, without a live cloud account validation claim. Named/sparse
vectors, Weaviate, Milvus, pgvector, OpenSearch, Elasticsearch, cloud signing,
custom query text, vector generation and full version/cluster matrices remain
planned. PostgreSQL login/SELECT 1 does not verify pgvector retrieval.

API references: [Qdrant metadata](https://api.qdrant.tech/api-reference/collections/get-collection),
[Qdrant query](https://api.qdrant.tech/api-reference/search/query-points),
[Pinecone statistics](https://docs.pinecone.io/reference/api/2025-10/data-plane/describeindexstats)
and [Pinecone query](https://docs.pinecone.io/reference/api/2025-10/data-plane/query).
