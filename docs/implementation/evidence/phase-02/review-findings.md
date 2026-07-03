# Phase 2 — Review Findings

Two parallel critique agents reviewed the database/model/migrations/testkit slice (2026-07-03):
**S** = RLS/tenant-isolation security reviewer (risk R1), who ran live probes against the compose
Postgres and REPRODUCED three isolation holes; **A** = architecture/API reviewer. Fixes carry
regression tests; the two critical security fixes were reproduced failing first, then fixed.

| ID | Sev | Finding | Resolution | Status |
|---|---|---|---|---|
| SEC-11 | critical | `SET ROLE app_rt` is session-scoped and never re-asserted per tx: a module issuing `RESET ROLE` escalates in-tx (superuser login) and poisons the pooled connection for later tenants — reproduced (tenant saw 2 rows) | runtime now authenticates AS a non-superuser `app_rt` login (not superuser+SET ROLE); TxManager re-asserts `SET LOCAL ROLE` per tenant tx; testkit provisions the local app_rt login out-of-band. `TestIntegrationRoleReassertedPerTx` reproduces then passes. D-0023 revised | **fixed** |
| SEC-12 | critical | `NewPool` never verified the effective role enforces RLS; an over-privileged DSN (or no role set) silently disables RLS — reproduced | `database.WithConnRLSGuard()` (connect-time) + `Manager.WithRLSGuard()` (per tenant tx) refuse superuser/BYPASSRLS roles fail-closed; scratch probe confirms `NewPool` refuses a superuser DSN | **fixed** |
| SEC-13 | high | `app_rt` held SELECT/INSERT/UPDATE on the RLS-less global identity tables → any module could read/tamper the cross-tenant membership graph — reproduced (grants present) | 00002 grants those tables to `app_platform` only; kernel identity services get a platform pool in Phase 4 (D-0026). app_rt cannot touch the global spine in Phase 2 | **fixed** |
| SEC-14 | low | `QueryTimeout <= 0` silently disabled the statement ceiling; `NewManager` bypasses `config.Validate` | `NewManager` clamps `<= 0` to the compiled default | **fixed** |
| SEC-15 | info | migrate tool runs as the raw superuser DATABASE_URL | header warning added: product `app.RunMigrate` must use the app_migrate owner DSN, never this shortcut | **fixed (doc)** |
| ARCH-16 | high | migration runner used one goose history table → kernel `00001..` and a module's `0001..` collide; docstring promised multi-source support the code didn't implement | `Migrate(…, source)` uses a per-source history table (`goose_version_<source>`); `migrations.SourceName`; returns `MigrateResult{Version,Applied}` (D-0027) | **fixed** |
| ARCH-17 | medium | `RuntimeDB`/`MigrateDB` hand-copied `config.DB` fields → silent drift as DB config grows (already dropped MaxConns from migrate) | shared knobs moved to `config.Pool`, embedded in `config.DB` and both views (D-0029) | **fixed** |
| ARCH-18 | medium | idempotency test only checked version equality against an already-migrated clone; fresh path never asserted | `TestIntegrationMigrateFreshAndIdempotent` runs `Migrate` on a genuinely empty DB, asserts `Applied>0` then rerun `Applied==0` | **fixed** |
| ARCH-19 | low | actor binding optional, deviating from 05 §2 "error if absent", with no decision | recorded as D-0030: actor stays optional until Phase 4 introduces the actor/audit machinery that consumes it, then `WithTenant` requires it | **accepted (documented)** |
| ARCH-20 | low | `ExpectOneRow` mapped >1 row to `ErrVersionConflict`, masking a too-broad-WHERE bug as a benign 409 | 0 → conflict; >1 → distinct internal error (500) (D-0028) | **fixed** |
| ARCH-21 | low | template DBs never reclaimed; 32-bit content hash | hash widened to 64 bits; orphan-template sweep noted as a future `make` target (retention is intentional for reuse) | **fixed (hash) / accepted (retention)** |
| ARCH-22 | low | probe.go comment misattributed why RLS binds (said FORCE; really non-owner role) | comment corrected: app_rt is non-owner/non-super so ordinary RLS binds; FORCE is defense-in-depth for the owner case | **fixed** |
| ARCH-23 | info | `internal/tools` ungoverned by boundary lint | lint now restricts internal/tools to kernel + migrations | **fixed** |
| ARCH-24 | info | sqlc pgx/v5 will need `CopyFrom`/`SendBatch` on DBTX for `:copyfrom`/`:batch` | noted for Phase 5/6 (added centrally on the sealed facade when the COPY helper lands); sealing verified escape-proof today | **accepted (tracked)** |
| ARCH-25 | info | `MigrateDB` lacked MaxConns | subsumed by the ARCH-17 `config.Pool` embed | **fixed** |
| ARCH-26 | info | `make ci` never runs integration tests | `make test-integration` documented and run in host + container gates this phase; wiring it into an auto-skipping `ci` step tracked for Phase 11 hardening | **accepted (tracked)** |

Reviewer-confirmed solid: fail-closed `app_tenant_id()` (empty/unset both ERROR, probed);
tenant-binding order + rollback/panic coverage; SET LOCAL works in READ ONLY tx; TenantDB sealing
is escape-proof (unexported struct, no pgx.Tx satisfaction); testkit advisory-lock + name
sanitization; view DSN narrowing; migration DDL ordering (citext before citext column).

Residual risk:
- The non-superuser-login requirement (D-0023) is now enforced in code (guards) and modelled in
  testkit, but product **deployment docs** must state it plainly — tracked for Phase 10/12.
- `Manager.Platform` runs on the app_rt pool; it has no grants on global tables now, so the real
  platform pool (app_platform login) must land with the first kernel identity service (Phase 4).
- Integration tests are gated in host + container runs but not yet in an auto-skipping `ci` step
  (Phase 11).
