# Test Strategy

Blueprint source: [08-testing-and-tooling.md](../blueprint/08-testing-and-tooling.md); Goal 2 §Testing Requirements.

## Layers & where they run

| Layer | What | Runner | Make target | From phase |
|---|---|---|---|---|
| Unit | pure funcs: config validation/redaction, cursor codecs, backoff math, policy eval, app registration/topo-sort | `go test ./...` (no services) | `test-unit` | 0 |
| Integration | service+repo+kernel against real Postgres (testcontainers; template-DB clone per test) | containers required | `test-integration` | 2 |
| Contract | `testkit.RunModuleContract` against `internal/testmodules/requests` + generated modules | containers | `test-contract` | 5 |
| Security | RLS isolation sweep, authz matrix, route-metadata completeness, secret-leak scan, unsafe-prod config matrix | containers | `test-security` | 2/4 |
| External-consumer | scratch product repo (created in CI tmpdir via `wowapi init` + `go mod edit -replace`) builds, registers a module, passes contract tests | containers + Go | `test-consumer` (part of `ci`) | 5 |
| Race | `go test -race ./...` | host or container | `test-race` | 0 |
| Bench | hot-path budgets (httpx chain, authz eval, config access) with 2× regression gate | host or container | `bench` | 3/11 |
| E2E | api+worker+migrate against compose stack; smoke workflows | compose | part of `ci` (Phase 12) | 12 |
| Golden | CLI + generator output snapshots; generated code must compile & test green | Go | `test` (Phase 10) | 10 |

## Non-negotiables
- Real Postgres for anything RLS/repo-shaped — no repository mocks (CLAUDE.md + Goal 2).
- Fakes only at true external boundaries: mail/SMS/push, malware scanner, IdP token minting,
  third-party webhooks, object storage fake alongside MinIO integration runs.
- Fake clock + deterministic ID gen injected via the same constructors production uses.
- Every blueprint acceptance criterion maps to a named test/assertion (tracked per phase in
  `evidence/phase-XX/acceptance-map.md`).
- All suites runnable in containers (`make shell` → same commands); host runs are a convenience.

## Container execution
`deployments/compose.yaml` provides postgres/minio/mailpit plus a `tools` service (Go image with
repo mounted) so `docker compose run tools make test` works without host Go. Testcontainers-based
suites run either on host Docker or inside the tools container via the mounted Docker socket
(documented in the compose file).
