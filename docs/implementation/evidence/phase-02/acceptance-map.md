# Phase 2 — Acceptance Map

Phase 2 exit criteria (Goal 2 Phase 2 + phase-plan row 2 + blueprint 03/05) → proof.

| # | Criterion | Proof |
|---|---|---|
| 1 | `kernel/database`: pgx pool, TxManager, TenantDB, Platform door | `kernel/database/{database,txmanager,context,errors,migrate}.go`; unit tests `database_test.go` |
| 2 | Tenant-bound transaction API with SET LOCAL | `Manager.WithTenant/WithTenantRO` bind `app.tenant_id` via `set_config(...,true)`; per-tx `SET LOCAL ROLE` (SEC-11) |
| 3 | RLS helpers + fail-closed tenant identity | `app_tenant_id()` (00001, no missing_ok → ERROR when unset); `testkit.AssertRLSIsolation`; `TestIntegrationNoTenantContextFails` (#5) |
| 4 | Kernel migrations 000–001 + runner | `migrations/00001_bootstrap.sql`, `00002_core_identity.sql`; `database.Migrate` per-source history (D-0027); `make migrate` live |
| 5 | tenant/user/access tables | 00002 (tenants, users, user_tenant_access; citext email, uta_active partial unique); `TestIntegrationKernelTablesExist` |
| 6 | testkit DB helpers + `AssertRLSIsolation` | `testkit/{db,asserts,probe}.go` + `fakes/`; 8 integration + fakes tests, all green live and in-container |
| 7 | **Fresh DB migrates idempotently** (acceptance #2) | `TestIntegrationMigrateFreshAndIdempotent` (empty DB: Applied>0 then Applied==0); live `make migrate` ×2 → version 2 both |
| 8 | **Tenant-scoped query without tenant context fails** (acceptance #5) | `TestIntegrationNoTenantContextFails` (`ErrNoTenantContext` at the door) + raw-runtime-query ERROR in `AssertRLSIsolation` |
| 9 | **RLS isolation** (acceptance #22) | `AssertRLSIsolation`: cross-tenant invisibility, WITH CHECK block, self-visibility; `TestIntegrationRLSProbe`; `TestIntegrationRoleReassertedPerTx` (SEC-11) |
| 10 | Non-superuser RLS identity enforced | `WithConnRLSGuard` (connect) + `WithRLSGuard` (per-tx) refuse superuser/BYPASSRLS; SEC-12 scratch probe fired; D-0023 |
| 11 | `kernel/model` base primitives | `kernel/model/model.go` (04 §3 verbatim) + `IDGen`/UUIDv7; `model_test.go` |
| 12 | Config DB section + process narrowing | `config.DB`/`config.Pool`; `app` RuntimeDB/MigrateDB; `TestViewsRequireTheirDSN`, `TestViewsCarryOnlyTheirDSN` |
| 13 | CLI secret-redaction snapshot (carried Phase 1 item, closed) | `internal/cli/config_redaction_test.go` over the new `db.dsn` Secret field (blueprint 12 §5) |
| 14 | No package cycles; boundary lint incl. internal/tools | `scripts/lint_boundaries.sh` → OK (governs internal/cli + internal/tools) |
| 15 | Container-first verification | `make ci` + `make ci-container` (Phase 1) still green; `make test-integration` green host + tools container |
| 16 | Evidence bundle + reviews | this directory; review-findings.md (R1 security + architecture agents, 2 reproduced criticals fixed) |
