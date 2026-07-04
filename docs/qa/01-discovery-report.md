# wowapi — QA Discovery Report

**Effort:** Comprehensive Regression Testing & Framework Reliability Hardening (`goal-test.md`).
**Date:** 2026-07-04. **Baseline commit:** `eb99fa8` (Goal 2 complete, Phases 0–12).
**Method:** grounded in the actual codebase — `go list`, `go test -cover`, `go tool cover -func`, and
direct source inspection. No behavior is assumed.

## 1. Architecture & layout (actual)

wowapi is a domain-neutral Go platform kernel distributed as a third-party dependency. Public package
layers (import law enforced by `scripts/lint_boundaries.sh` + `wowapi lint boundaries`):

- `kernel/*` — the framework core (no imports of module/app/adapters/testkit). 30 sub-packages:
  config, secrets, logging, model, database, errors, httpx, validation, pagination, filtering,
  auth, authz, policy, relationship, resource, outbox, jobs, rules, workflow, document, storage,
  comment, attachment, notify, webhook, integration, observability.
- `module` — the public module contract (`Module` interface + capability-scoped `Context`).
- `app` — the sole composition root (`kernel.New` + `App.Boot`, worker, run hooks, readiness).
- `adapters/*` — concrete third-party integrations (secrets/envprovider, metrics/prometheus).
- `testkit` — the external-consumer test harness (fixtures, RLS asserts, contract runner, pools).
- `migrations` — embedded goose kernel migrations 00001–00011.
- `cmd/wowapi` — the CLI; `internal/cli`, `internal/tools/{migrate,benchbudget}`, `internal/e2e`.

## 2. Test inventory (measured)

- 43 buildable packages; **477** `Test`/`Benchmark`/`Example` functions across **73** `_test.go` files.
- kernel: 48 test files over 80 source files; app 5/7; adapters 2/2; testkit 8/10; internal 7/18.
- Test kinds present: unit, component, integration (real Postgres via testkit templates), API/contract
  (`testkit.RunModuleContract`, `TestIntegrationScratchConsumer`), E2E (`internal/e2e` scaffold→build→
  migrate→/healthz), security suite (`make test-security`), benchmarks + budget gate, race gate.
- Fixtures/helpers: `testkit` (Admin/Runtime/Platform pools + PlatformTxM, CreateTenant/User/Capacity/
  Org/Role/GrantRole/CreateResource[Type], TenantCtx, AssertRLSIsolation), `testkit/fakes`.

## 3. Per-package statement coverage (measured, DB-backed)

| Band | Packages |
|---|---|
| 100% | model, secrets, migrations, testkit/fakes, adapters/secrets/envprovider |
| 90–99% | adapters/metrics/prometheus (97), storage (92), filtering (92) |
| 80–89% | errors (89), logging (88.5), auth (88), observability (87), pagination (85), authz (84), integration (84) |
| 70–79% | config (79), webhook (79), notify (78), policy (76.5), attachment (74.5), comment (74), validation (71) |
| 60–69% | internal/cli (67.6), document (65), httpx (63.5), seeds (63.3), jobs (63.2), outbox (63) |
| 50–59% | relationship (58.3), rules (58.4), workflow (53.3), testkit (53) |
| < 40% | app (39.8), resource (23.5), database (14.3) |
| 0% (no _test.go) | kernel (root `kernel.go`), module, cmd/wowapi, internal/tools/{migrate,benchbudget}, internal/testmodules/requests |

Coverage % is a *pointer*, not the target. Low numbers in `database` reflect infra exercised
indirectly by every integration test; `app`/`kernel` root reflect boot paths hit indirectly. The real
work is finding UNTESTED behavior that affects correctness/security/data-integrity/regression — see the
Coverage Matrix (03) and the confirmed gaps below.

## 4. Confirmed meaningful gaps (function-level, verified non-duplicative)

Verified via `go tool cover -func` (0.0% funcs) cross-checked against existing `_test.go` names:

| ID | Area | Untested behavior | Risk | Class |
|---|---|---|---|---|
| G1 | `kernel/database` RLS guard | `WithConnRLSGuard` / `WithRLSGuard` — the fail-closed rejection of a superuser/BYPASSRLS runtime connection (RLS is defeated for such roles) has NO direct test | **HIGH** | security / tenancy |
| G2 | `kernel/workflow` runtime | `CompleteTask`, `Delegate`, `Override` (runtime, incl. the Phase-7 authz gate), `OpenTasksFor`, gateway-step routing (`gatewayTarget`) — none exercised | HIGH | workflow correctness |
| G3 | `kernel/workflow` SLA | `parseISODuration` (ISO-8601 SLA durations) — no direct test of parse/edge cases | MED | parsing / edge |
| G4 | `kernel/relationship` | `Relate` (the ReBAC relationship write path) — untested directly | MED | data integrity / authz |
| G5 | `kernel/resource` registry | `Register` / duplicate / key validation / `Err` accumulation — the registration contract has no direct unit test | MED | framework contract |
| G6 | `kernel/database` idempotency | `IdemStore.Begin`/`Complete`/`Discard` at the DB layer — the atomic claim (data-integrity core) is covered only indirectly via httpx | MED | data integrity / concurrency |
| G7 | `kernel` root | `kernel.New` wiring — no direct assertion that every registry/service field is non-nil and correctly wired | LOW | wiring regression |
| G8 | `internal/tools/benchbudget` | budget-file parsing + threshold logic — 0% (the CI perf gate's own logic is untested) | LOW | tooling reliability |

These are the targets. Each is behavior that affects framework correctness, security, data integrity,
or a public contract — not coverage-chasing and not duplicative of existing tests.

## 5. Non-goals / honest scope notes

- Existing strong coverage (RLS isolation sweeps, deny-by-default authz, outbox atomicity, job DLQ,
  document security, notify/webhook, config redaction, CLI, E2E) is NOT re-implemented — see 02.
- `cmd/wowapi` / `internal/tools/migrate` mains are thin wrappers (exercised by E2E / integration);
  not targeted for direct unit tests.
- Graphify semantic extract remains blocked on an LLM key (R11) — environmental, not a test gap.
