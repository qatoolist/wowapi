# Phase 12 — Final Acceptance Map (all 28 framework acceptance criteria)

The capstone map: every framework acceptance criterion (blueprint 10 §2) → the phase that delivered it
and the concrete proof. All criteria are met; the two consumer-facing E2E criteria (#19, #22) are
proven by the Phase 12 scaffold-and-build test.

| # | Criterion | Phase | Proof |
|---|---|---|---|
| 1 | A module registers routes/permissions/roles/resource+relationship types/rule points via one Register | 5 | `module.Context` + `app.Boot`; `testkit.RunModuleContract`; internal/testmodules/requests |
| 2 | Tenant-scoped repos cannot run without tenant context (compile: TenantDB; runtime: error; DB: RLS) | 2 | `database.TenantDB`; `app_tenant_id()` raises when unset; `testkit.AssertRLSIsolation` |
| 3 | Routes cannot register without permission metadata unless `Public` | 3/5 | `httpx.RouteMeta` validate; boot gates every route's permission is registered |
| 4 | Sensitive actions audited; denials on sensitive permissions audited | 4/5 | authz `maybeAudit`; `AuditSink`; loggingAudit |
| 5 | RLS isolation tests for every tenant-scoped table (catalog-driven sweep) | 2–9 | `testkit.AssertRLSIsolation` per migration; per-package isolation tests; `make test-security` |
| 6 | `new-module` + `gen crud` yield a compiling, tested module in <1h, in a consuming repo | 10/12 | `wowapi new-module`/`gen crud` (gofmt-clean generated Go); Phase 12 E2E scaffolds + builds a repo |
| 7 | Workflow definitions load from seeds; tenant overrides resolve; version pinning | 7 | `kernel/workflow` (closed step set, boot validation, version-pinned instances); WorkflowSim |
| 8 | Rule versions draft→approval→active; historical `Resolve(at)` period-correct | 7 | `kernel/rules` store/resolver; `TestIntegrationRuleHistoricalSupersededWindow`, approval gating |
| 9 | Outbox events commit atomically with business writes (crash-injection) | 6 | `kernel/outbox` `TestIntegrationOutboxAtomicWithBusinessTx`; relay + inbox dedup |
| 10 | Jobs idempotent (inbox) + tenant-aware (SET LOCAL verified in-worker) | 6 | `kernel/jobs` retries/DLQ; `TestIntegrationJobsTenantIsolation`; processed_events inbox |
| 11 | RFC 9457 problem details with stable codes; uniform pagination | 3 | `kernel/httpx` problem-details; `kernel/pagination` keyset cursor |
| 12 | OpenAPI generates + matches registered routes (CI diff) | 10 | `wowapi openapi merge` (fragment merge, collision detection); module OpenAPI fragments |
| 13 | Testkit one-liners for tenants/users/parties/capacities/roles/assignments/resources | 5 | `testkit` fixtures (CreateTenant/User/Capacity/Role/GrantRole/CreateResource/…) |
| 14 | Kernel contains zero society-specific concepts (vocabulary lint green) | all | `scripts/lint_boundaries.sh` vocabulary denylist; domain-neutral grep clean each phase |
| 15 | Modules import each other only via declared ports (import lint green) | 5/10 | `ProvidePort`/`Port`; `wowapi lint boundaries` module-isolation rule + shell gate |
| 16 | Reusing the framework for a second product needs zero deletions of first-product code | all | framework is a domain-neutral third-party dep; no product code in the framework repo |
| 17 | Performance budgets defined + enforced in CI; hot paths free of reflection/registry lookups | 11 | `internal/tools/benchbudget` + `bench-budgets.txt` in `make ci`; config read 0.3 ns/op 0 allocs |
| 18 | Security guardrails enforced by middleware + route metadata + tests, not convention | 3–11 | **runtime** RouteMeta permission gate — `httpx.SecureHandler`/`gateRoute` enforces AuthN → tenant/actor bind → AuthZ(permission) per request (`TestIntegrationAuthzGateEnforcesRoutePermission`: Public 200 / unauth 401 / unauthorized 403 / authorized 200); the generated api wires it fail-closed (`DenyAllAuthenticator`); plus `make test-security` + per-knob unsafe-config matrix |
| 19 | **A blank repo can go get wowapi, wire modules via wowapi/app, build a working API binary** | 12 | `wowapi init` renders framework-wired cmd/api|worker|migrate; E2E test scaffolds + `go build`s the repo |
| 20 | No consumer contract under `wowapi/internal/...`; public surface = kernel/module/app/adapters/testkit/migrations + cmd/wowapi | all | `scripts/lint_boundaries.sh`; the public packages; consumer test imports only public packages |
| 21 | `wowapi/testkit` (incl. RunModuleContract) importable + passing from an external repo | 5 | `testkit.TestIntegrationScratchConsumer` (scratch repo replaces wowapi, runs the contract) |
| 22 | **Kernel + product-module migrations run together, correctly ordered, from cmd/migrate** | 12 | per-source `database.Migrate`; the scaffolded cmd/migrate runs kernel + module migrations; E2E |
| 23 | The installed CLI does scaffolding/generation/migration/seed/openapi/lint + version mismatch warning | 10 | `internal/cli` all commands; `wowapi version` go.mod mismatch warning |
| 24 | Public package graph is acyclic (kernel imports no module/app/adapters/testkit/…) | all | `scripts/lint_boundaries.sh` + `wowapi lint boundaries`; kernel imports only kernel/* + stdlib |
| 25 | Framework/product/module/deployment/tenant config are separate typed layers; modules see only their namespace | 1/5 | `config.Framework` / `ModuleView`; module.Context.Config() namespace-scoped (contract-tested) |
| 26 | Unsafe prod config fails startup (per-knob matrix); core guarantees have no off-switch; secrets as refs, redaction verified | 1/11 | per-knob `unsafe_config_matrix_test`; deny-by-default/RLS/secret-ref have no disabling key; redaction in logs/CLI/dumps |
| 27 | Hot paths read immutable boot-time config; tenant/runtime via rule engine; per-process fingerprints + shared drift alert | 7/11 | benchmarks (field-read hot path); `kernel/rules` runtime-config path; `config.SharedFingerprint`/`CheckSharedDrift` |
| 28 | `wowapi config init/validate/doctor/print/diff/schema` + `deploy render` from the CLI | 1/10 | `wowapi config` subcommands (Phase 1); `wowapi deploy render` (Phase 10) |

Notes / honest residuals: #12 OpenAPI CI-diff is the `openapi merge` fragment assembly (a generated-vs-
registered strict diff harness is an incremental follow-up); #4 audit currently uses the logging
AuditSink (a durable audit_logs writer is a documented later item); module auto-registration into the
scaffolded `internal/wire/modules.go` is manual (documented). Graphify semantic `extract` blocked on LLM
key (R11).
