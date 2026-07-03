# Phase 2 — Command Log

Exact commands with exit status; summarized output where long. Date: 2026-07-03.

| # | Command | Exit | Output summary |
|---|---|---|---|
| 1 | `go get pgx/v5 google/uuid goose/v3 shopspring/decimal` | 0 | Phase 2 dependencies pinned (R8: latest stable) |
| 2 | `go test ./kernel/config/ ./app/... ./internal/cli/ -count=1` (after config.DB section) | 0 | DB section added without breaking existing suites |
| 3 | `go build ./kernel/database/` | 0 | pool/TxManager/migrate runner compile |
| 4 | `go test ./kernel/model/ -count=1` (agent) | 0 | 4 tests (ActiveAt boundaries, UUIDv7 ordering, Statused, decimal Money) |
| 5 | `go test ./migrations/ -count=1` (agent) | 0 | 4 tests (embed inventory, goose markers, fail-closed app_tenant_id, no RLS on globals, no passwords) |
| 6 | `go test ./app/... ./internal/cli/ -count=1` (views DB narrowing + CLI redaction snapshot) | 0 | D-0021 narrowing tests + blueprint 12 §5 CLI secret-snapshot test (carried Phase 1 item — closed) |
| 7 | `make migrate` ×2 against live compose DB | 0/0 | kernel migrations applied; version 2 both runs — live idempotency (acceptance #2) |
| 8 | `DATABASE_URL=… go test ./testkit/... -count=1 -v` (agent, live PG 16.14) | 0 | 6 integration tests: MigrateIdempotent, RLSProbe (4 isolation properties), NoTenantContextFails, KernelTablesExist(+no rowsecurity on globals), ReadOnlyTx, VersionConflictHelper; fakes 4 tests `-race` clean |
| 9 | `gofmt -l . && make ci && make test-integration` | 0 | host CI (vet, boundaries, unit, race, build) + integration suite green |
| 10 | `docker compose run --rm tools make test-integration` | 0 | integration suite green INSIDE the tools container against the compose postgres service (container-first gate) |
| 11 | SEC-11 reproduction: `TestIntegrationRoleReassertedPerTx` before the fix | FAIL | tenant saw 2 rows after `RESET ROLE` — confirmed the superuser-login escalation the review reported |
| 12 | After fix (non-superuser app_rt login + per-tx role/guard): `make test-integration` | 0 | SEC-11 test passes; all testkit + database integration tests green |
| 13 | SEC-12 probe (scratch module, superuser DSN + `WithConnRLSGuard`) | 0 (guard fired) | `NewPool` refused: "effective role is superuser or BYPASSRLS; RLS would not be enforced" — fail-closed confirmed |
| 14 | `gofmt -l . && sh scripts/lint_boundaries.sh && make ci && make test-integration` (after all review fixes) | 0 | boundary lint OK (now governs internal/tools + internal/cli); host CI + integration green |
