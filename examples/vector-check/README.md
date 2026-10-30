# Vector-store health examples

Phase 19 supports Qdrant REST and Pinecone dense-vector data-plane REST.
Use a CLI built from this source until its synchronized release is verified.
These services are supporting dependencies, not composite agents.

## Disposable Qdrant matrix

Qdrant **1.15.5** binds only loopback. The key below is a disposable test key.
Provision the collection separately: the health adapter never creates data.

```bash
docker compose -p agenthealth-phase19 -f examples/vector-check/compose.yaml up -d
# Wait until this authenticated metadata request succeeds:
curl --fail --retry 30 --retry-all-errors --retry-connrefused --retry-delay 1 \
  -H 'api-key: phase19-local-key' http://127.0.0.1:16333/collections
curl --fail -X PUT -H 'api-key: phase19-local-key' \
  -H 'Content-Type: application/json' \
  -d '{"vectors":{"size":3,"distance":"Cosine"}}' \
  http://127.0.0.1:16333/collections/docs
export VECTOR_KEY=phase19-local-key
agenthealth check examples/vector-check/agenthealth.yaml
agenthealth check examples/vector-check/query.yaml
AGENTHEALTH_QDRANT_ENDPOINT=http://127.0.0.1:16333 \
  go test -race ./adapters/vectorstore -run TestRealQdrant -v
docker compose -p agenthealth-phase19 -f examples/vector-check/compose.yaml down -v
```

Expected: passive metadata and explicit top-1 query are healthy on the empty
collection. The Go matrix also checks missing collection and wrong-key denial.
No insertion is required; an empty query result validates execution, not relevance.
Curl retries above are provisioning only; active health queries never retry.

## Pinecone

Replace `YOUR_INDEX_HOST` in [pinecone.yaml](pinecone.yaml) with the actual
index host, set `PINECONE_API_KEY`, then run `agenthealth check` with that file.
Do not use the control-plane `api.pinecone.io` host. Add a `namespace` only if
its existence is required. The checked dialect is dense vectors, API 2025-10.
Cloud account interoperability is not established by the local mock suite.

To opt into a query, add `checks` containing `functional` and
`vector_store.functional: {safe: true, vector: [1, 0, 0]}` using your actual
index dimension. Each query may incur read charges. Never put production
vectors or API keys directly into committed examples.

See the [guide](../../docs/vector-store-adapter.md) and
[alignment audit](../../docs/phase-19-alignment.md).
