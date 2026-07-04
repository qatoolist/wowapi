# wowapi — Test Execution Report & Regression Execution Guide

## A. Execution report (this effort)

Date: 2026-07-04. Environment: darwin/arm64, Go 1.26, Postgres via
`DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable`, and the
container toolbox (`make ci-container`).

### New tests — all PASS (21/21)

| Suite | Result |
|---|---|
| `kernel/database` RLS guard (2) | PASS |
| `kernel/workflow` runtime lifecycle (5) | PASS |
| `kernel/workflow` SLA parse (2) | PASS |
| `kernel/relationship` Relate (3) | PASS |
| `kernel/resource` registry (6) | PASS |
| `internal/tools/benchbudget` parser (4) | PASS |

Passed: 21. Failed: 0. Skipped: 0 (the RLS-guard reject test is env-conditional but ran — the test
cluster's login role is superuser/BYPASSRLS, so the reject path executed). Flaky: 0. Blocked: 0.

### Full regression — all PASS, no regressions

- `make ci` (host: vet + boundary lint + unit + **race** + **bench-budget** + build) → exit 0.
- `make ci-container` (parallel, in-container) → exit 0, zero failures, zero role/concurrent-update
  flakes (the Phase-11 testkit fix holds).
- The touched packages pass under `-race`.

### Coverage delta (measured)

| Package | Before | After |
|---|---|---|
| kernel/resource | 23.5% | 97.1% |
| kernel/relationship | 58.3% | 91.7% |
| kernel/workflow | 53.3% | 68.1% |
| kernel/database | 14.3% | 31.3% |
| internal/tools/benchbudget | 0% | 61.3% |

The remaining `database` uncovered lines are pool/tx plumbing exercised indirectly by every integration
test (per-package attribution artifact), not untested behavior.

### Defects found

No new framework defects. One design finding (D1 — `relationship.Relate` is platform-only and had no
callers/tests) is documented and now regression-protected; it is not a bug (the privilege split is
correct), so no code change was required. See 06.

## B. Regression execution guide (actual project commands)

All commands run from the repo root. Integration/E2E/security suites need a Postgres DSN — start the
dev stack with `make up` (or export `DATABASE_URL`).

| Intent | Command |
|---|---|
| **Full CI gate** (vet, boundary lint, unit, race, perf budgets, build) | `make ci` |
| **Full CI in the container** (authoritative regression) | `make ci-container` |
| Unit tests only | `make test-unit` |
| Race detector | `make test-race` |
| Integration (real Postgres) | `make test-integration` |
| **Security suite** (RLS, authz, secrets, unsafe-config) | `make test-security` |
| Performance budgets | `make bench` then `make bench-budget` |
| Serial in-container (avoids parallel template contention if ever needed) | `docker compose -f deployments/compose.yaml run --rm tools go test -p 1 ./...` |
| A single package | `DATABASE_URL=… go test -count=1 ./kernel/<pkg>/` |
| A single test | `DATABASE_URL=… go test -run TestName -v -count=1 ./kernel/<pkg>/` |
| **This effort's new tests** | `DATABASE_URL=… go test -run 'ConnRLSGuard|CompleteTask|Delegate|Override|GatewayRouting|ParseISODuration|Relate|Registry|ValidTypeKey|BaseName|LoadBudgets|ParseBenchOutput' ./kernel/database/ ./kernel/workflow/ ./kernel/relationship/ ./kernel/resource/ ./internal/tools/benchbudget/` |
| Coverage (per package) | `DATABASE_URL=… go test -cover ./...` |
| Coverage profile + funcs | `go test -coverprofile=cov.out ./kernel/<pkg>/ && go tool cover -func=cov.out` |
| E2E (scaffold→build→migrate→/healthz) | `DATABASE_URL=… go test -run E2E ./internal/e2e/` |
| External consumer contract | `DATABASE_URL=… go test -run ScratchConsumer ./testkit/` |

**Suggested cadence:** pre-commit → `make ci`; pre-merge → `make ci-container` + `make test-integration`;
release → the full table + `make bench-budget`.
