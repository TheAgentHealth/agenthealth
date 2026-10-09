# Dependency graph

Phase 10 supports shared nodes through explicit IDs and dependency references.
See the [contract and impact review](rfcs/phase-10-dependency-graph.md).

<!-- spec-example: configuration -->
```yaml
version: v1
concurrency: 4
targets:
  - id: first
    name: Research Agent
    type: agent
    endpoint: http://localhost:8080/health
    budget_ms: 3000
    dependencies:
      - ref: peer
        relationship: downstream
      - ref: first-peer-path
        relationship: path
      - ref: router
        relationship: router
  - id: peer
    name: Search Agent
    type: agent
    endpoint: http://localhost:8081/health
  - id: first-peer-path
    name: Search communication
    type: agent
    endpoint: http://localhost:8080/health
    checks: [configuration, functional]
    agent:
      functional:
        safe: true
        text: Return a fixed health acknowledgement from the search peer.
        downstream: Search Agent
  - id: router
    name: Router
    type: router
    endpoint: http://localhost:8082/health
    dependencies:
      - ref: peer
        relationship: downstream
        critical: false
```

The [standalone example](../examples/graph-check/agenthealth.yaml) contains this configuration. Save it as `agenthealth.yaml` and run `agenthealth check agenthealth.yaml`
against compatible configured services. Functional probes require a handler
that actually contacts the selected peer, as in the
[agent example](../examples/agent-check/README.md).

The peer is checked once even though it appears at the root and through two
edges. Critical policy belongs to each edge. A failing path impairs the first
agent even when direct peer checks pass. Names advertised by agents do not
create nodes or discover endpoints; references resolve only configured IDs.

JSON/YAML preserve IDs, relationships, and critical policy. Human output
shows the dependency tree with relationship labels. Each node's check evidence
is separate from its aggregate status. Nested configurations continue to work.

`budget_ms` bounds a node's own work; shared dependencies keep their independent
budgets. `concurrency` bounds adapter calls from 1 to 16, defaulting to 16.
Missing references, duplicate IDs, cycles and excessive depth/projections are
rejected before probes. The engine accepts at most 10000 result tree projections.

Topology is explicit and may be sensitive. Only share machine output with
appropriate consumers. IDs must not contain secrets; known credential values
are redacted in output. Experimental AHP serving requires bearer authorization for topology; see [AHP](ahp.md).

## Kubernetes deployment

Existing explicit graphs can be mounted as reviewed configuration. Shared nodes and edge policies retain their aggregation and budgets. Optional-edge failures can yield DEGRADED, which still fails CLI exec probes; AHP readiness follows its own contract.
See [Phase 14 deployment guidance](kubernetes.md) for ConfigMaps, Secrets,
probe policy and validation limits. No adapter or configuration migration is required.
