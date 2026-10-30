# Database adapter

Phase 18 implements native PostgreSQL, MySQL, Redis and MongoDB checks.
Use the `database` target type. Data services are supporting dependencies,
not composite agents. The implementation uses trusted compiled Go drivers;
it does not add a SQL console or accept arbitrary commands.

| Endpoint scheme | Default port | Passive evidence | Explicit `functional` probe |
|---|---|---|---|
| `postgres` | 5432 | Native connection/login | `SELECT 1`, expect integer 1 |
| `mysql` | 3306 | Native connection/login | `SELECT 1`, expect 1 |
| `redis` | 6379 | Authenticated `PING`, expect `PONG` | Another `PING` |
| `mongodb` | 27017 | Authenticated `ping` command | Another `ping` command |

SQL query execution requires `functional` in `checks`. No writes, application
queries, collection reads, DDL, migrations or server-side scripts are supported.
A successful login/ping does not establish application permissions, table/index
availability, replication freshness or overall application read readiness.
Even `SELECT 1` only proves a minimal constant query can execute.

<!-- spec-example: configuration -->
```yaml
version: v1
targets:
  - name: primary-db
    type: database
    endpoint: postgres://db.example.com:5432/app
    checks: [reachability, authentication, protocol, capability, functional, latency]
    timeout_ms: 3000
    database:
      username: health
      password_env: AGENTHEALTH_DATABASE_PASSWORD
      tls: required
```

`database` accepts only `username`, `password_env`, and `tls`. SQL endpoints
require a username and a nonempty database path. Redis permits database numbers
0–15 (default 0). MongoDB uses the path as database and authentication source,
defaulting to `admin`. MongoDB connects to one explicit server directly;
SRV discovery, replica-set discovery and sharded topologies are outside this
initial scope. MongoDB username/password authentication uses SCRAM negotiation;
Redis supports password and username/password ACL authentication. SQL uses the
selected driver's supported password handshake. TLS client certificates and
cloud/IAM authentication are not implemented.

TLS defaults to `required`, verifies server names and trust chains, and requires
TLS 1.2+. There is no skip-verification or plaintext fallback. `disabled` is an
explicit plaintext mode intended for the disposable loopback examples. PostgreSQL
uses native TLS negotiation; MySQL uses its TLS handshake, and Redis/MongoDB use
TLS connections. URI credentials, query options, fragments, multiple hosts and
arbitrary driver settings are rejected. Use an explicit port when needed.

Passwords resolve once per run from `password_env`, participate in engine
redaction, and never enter a configured endpoint. Missing or empty referenced
passwords fail Configuration before any connection. Do not set `auth` on a
database target; `auth.bearer_env` belongs to other adapter transports. An absent
password reference permits a passwordless server configuration. Keep service
accounts minimally privileged and protect the configuration as trusted input.

Reachability measures a short TCP connection, including its latency. It does
not authenticate or verify TLS; native authentication/protocol/capability checks
perform that work under their check deadlines. A configuration containing only
`reachability` therefore establishes TCP reachability alone. Default checks also
request native protocol/authentication evidence. Passive native evidence is shared
within one target run; separate graph nodes and runs do not reuse it. Functional
SQL uses a fresh connection and runs exactly one constant query. Driver-level
operation retries are disabled; after native work begins the adapter marks the
attempt conservatively as response-received so engine retries cannot repeat it.

Recognized server authentication denials produce `MISCONFIGURED`; other native
protocol/query failures produce `UNHEALTHY`. TCP failures and deadline expiry use
the shared engine classification. PostgreSQL message bodies are capped at 4 MiB and Redis reply lines at 1 KiB.
MySQL and MongoDB retain their native driver framing limits. Raw driver errors, server messages, query
results and passwords never appear in output. A native probe's failure can affect
multiple dimensions because those dimensions consume the same evidence.
Connections close on success/failure/cancellation; cleanup contexts are bounded.

Use these targets as independent nested dependencies or shared graph nodes.
Existing critical/optional edge policies, graph deduplication, JSON/YAML results,
CLI exit codes and AHP snapshot freshness semantics apply unchanged. The Python
and JavaScript SDKs use their bundled canonical schemas and matching CLI; they
do not create independent database connections.

For Kubernetes, supply `password_env` using a Secret and the chart's existing
`envFrom`/environment facilities. Keep the total check budget below probe/Job
deadlines. `/live` remains process liveness; database readiness belongs to the
AHP snapshot and its dependency policy. Never place literal production passwords
in ConfigMaps, Helm values, command arguments or example files.

[Disposable examples](../examples/database-check/README.md) record exact tested
server versions and healthy/wrong-password expectations. This controlled native
matrix does not establish managed-cloud, cluster, proxy or broad version support.
See the [phase impact review](phase-18-alignment.md) and
[configuration contract](../spec/configuration.md#database-options-phase-18).
