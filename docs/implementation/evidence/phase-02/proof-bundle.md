# Phase 2 — Proof Bundle

Scope (phase-plan row 2): `kernel/database` (pool, TxManager/TenantDB, RLS binding, migration
runner), `kernel/model`, kernel migrations 000–001, tenant/user/access tables, testkit DB helpers +
`AssertRLSIsolation`. Date: 2026-07-03.

## 1. Decision evidence
Pre-code: D-0020 (kernel/model ships now), D-0021 (DSN validated at narrowing), D-0022 (env-DSN +
template clone, no testcontainers), D-0023 (RLS identity), D-0024 (TenantDB growth), D-0025 (RLS
proven on probe tables). Post-review: D-0023 **revised** (non-superuser login mandatory), D-0026
(global tables → app_platform), D-0027 (per-source migrations), D-0028 (ExpectOneRow), D-0029
(config.Pool), D-0030 (actor binding deviation). All in `docs/implementation/decisions.md`;
blueprint 03 §5 / 12 §4 updated where behavior changed.

## 2. Discussion evidence
- Testcontainers vs env-DSN template clone: chose env-DSN (compose already provides PG; avoids the
  single largest dependency) — D-0022.
- The central RLS debate: the security review REPRODUCED that `SET ROLE` from a superuser login is
  escapable (RESET ROLE) and that an over-privileged DSN silently disables RLS. Resolution: the
  runtime must be a genuine non-superuser login; SET ROLE is not a security boundary. Both connect-
  time and per-tx guards added; testkit models the production login. This reversed the original
  D-0023 mechanism — recorded, not silently changed.
- Migration multi-source: one history table can't support kernel+module numbering; moved to
  per-source history tables before the contract shipped to any consumer (D-0027).

## 3. Critique/review evidence
`review-findings.md`: 16 findings (2 critical + 1 high security, 1 high architecture, 2 medium,
rest low/info). Two criticals (SEC-11, SEC-12) and one high (SEC-13) were reproduced against live
Postgres before fixing; SEC-11 has a reproduce-then-pass regression test; SEC-12 has a live guard
probe. Every finding fixed or accepted with rationale + tracking.

## 4. Implementation evidence
New: `kernel/database/` (database, txmanager, context, errors, migrate + tests), `kernel/model/`,
`migrations/` (00001, 00002, embed + tests), `testkit/` (db, asserts, probe, fakes + tests),
`internal/tools/migrate/`. Changed: `kernel/config/config.go` (DB + Pool), `app/views.go` (DB
narrowing + section fingerprints), `internal/cli/config_redaction_test.go`, `Makefile` (migrate,
test-integration), `scripts/lint_boundaries.sh` (internal/tools + internal/cli), Dockerfile (Phase
1 cgo). Deps: pgx/v5, google/uuid, goose/v3, shopspring/decimal.
Team: 3 parallel implementation agents (model / migrations / testkit) + lead (database, config,
views, all security fixes); 2 parallel review agents (R1 security + architecture).

## 5. Verification evidence
`command-log.md`: unit suites per package, live `make migrate` idempotency, 8 testkit integration
tests (host #8 and tools-container #10), SEC-11 reproduce→fix (#11–#12), SEC-12 guard probe (#13),
full `make ci` + `make test-integration` after fixes (#14). Graphify updated at phase end.

## 6. Acceptance evidence
`acceptance-map.md`: all 16 Phase 2 exit criteria mapped to code, named tests, and command-log
entries; acceptance #2/#5/#22 each to a specific integration assertion. Carried forward:
- Platform pool as app_platform login → Phase 4 (first kernel identity service, D-0026).
- Actor binding becomes required in `WithTenant` → Phase 4 (D-0030).
- `ci` auto-running integration when a DB is present → Phase 11 (ARCH-26).
- Non-superuser-login deployment doc → Phase 10/12.
- sqlc CopyFrom/SendBatch on the sealed facade → Phase 5/6 (ARCH-24).
- Graphify semantic `extract` still blocked on LLM key (R11).
