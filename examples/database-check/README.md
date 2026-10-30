# Native database health

These disposable services use PostgreSQL 17.6, MySQL 8.4.6, Redis 7.4.5 and
MongoDB 8.0.13. Docker Compose v2 and Go 1.26+ are required. Ports bind only to
loopback. Credentials below are public test values; remove volumes after use.

```bash
docker compose -p agenthealth-phase18 -f examples/database-check/compose.yaml up -d --wait --wait-timeout 240
export AGENTHEALTH_DB_PASSWORD=health-test-password
go run ./cmd/agenthealth check examples/database-check/agenthealth.yaml --format json
AGENTHEALTH_DATABASE_TESTS=1 go test -race ./adapters/database -run TestRealDatabases -v
docker compose -p agenthealth-phase18 -f examples/database-check/compose.yaml down -v
```

The suite expects healthy read-only probes, then `MISCONFIGURED` for incorrect
passwords, and checks password suppression in JSON. The configuration explicitly
disables TLS for these loopback fixtures; production defaults verify TLS.
Use a least-privilege database user in production, not this MongoDB root fixture.
The [guide](../../docs/database-adapter.md) explains scope and limits.
