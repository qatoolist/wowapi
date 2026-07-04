# wowapi — Coverage Matrix

Columns: Module/Flow · Components · Existing coverage · New coverage (this effort) · Happy · Negative ·
Integration · E2E · Regression priority · Gaps/notes. "✓✓" = strong, "✓" = present, "—" = n/a.
Grounded in the actual suite (measured coverage in 01 §3).

| Module / Flow | Components | Existing | New (this effort) | Happy | Neg | Integ | E2E | Reg pri | Gaps / notes |
|---|---|---|---|---|---|---|---|---|---|
| Tenancy / RLS isolation | database, testkit, every tenant table | ✓✓ AssertRLSIsolation sweep + per-pkg isolation | — | ✓✓ | ✓✓ | ✓✓ | ✓ | **HIGH** | Well covered |
| **RLS connection guard** | database.WithConnRLSGuard / WithRLSGuard | ✗ (0% direct) | **✓✓ rls_guard_test** (reject superuser/BYPASSRLS; admit app_rt) | ✓ | ✓✓ | ✓ | — | **HIGH** | **Gap closed (G1)** |
| Deny-by-default authz | authz, policy, relationship | ✓✓ evaluator + escalation | — | ✓✓ | ✓✓ | ✓✓ | — | HIGH | Well covered |
| **ReBAC edge write** | relationship.Relate | ✗ (unused/untested) | **✓✓ relationship_relate_test** (platform write→Has; app_rt denied; isolation) | ✓ | ✓✓ | ✓✓ | — | MED | **Gap closed (G4/D1)** |
| ReBAC edge check | relationship.Has | ✓✓ (live/expired/other/system) | — | ✓✓ | ✓ | ✓✓ | — | MED | Well covered |
| **Resource type registry** | resource.Registry | ✗ (23%→) | **✓✓ registry_test** (key shape, prefix, dup, accumulate) | ✓ | ✓✓ | — | — | MED | **Gap closed (G5)** 23→97% |
| Rules engine | rules store/resolver | ✓✓ precedence/historical/schema/approval | — | ✓✓ | ✓ | ✓✓ | — | HIGH | Well covered |
| Workflow definition | workflow definition validate | ✓✓ orphan/dangle/unreachable/fail-closed | — | ✓✓ | ✓✓ | — | — | HIGH | Well covered |
| **Workflow runtime** | CompleteTask/Delegate/Override/gateway | ✗ (0% these funcs) | **✓✓ runtime_lifecycle_test** (task complete+neg; delegate; override+neg; gateway route) | ✓✓ | ✓✓ | ✓✓ | — | **HIGH** | **Gap closed (G2)** 53→68% |
| Workflow decide/reject | runtime Decide | ✓✓ approval/reject/optimistic-lock | — | ✓✓ | ✓ | ✓✓ | — | HIGH | Well covered |
| **Workflow SLA parse** | parseISODuration | ✗ (0%) | **✓✓ sla_parse_test** (W/D/H/M/S valid; malformed) | ✓ | ✓✓ | — | — | MED | **Gap closed (G3)** |
| Workflow SLA sweep | sweeper | ✓ SweepSLA idempotent | — | ✓ | — | ✓ | — | MED | Sweep covered; parse now covered |
| Outbox / relay / inbox | outbox | ✓✓ atomicity/ordering/DLQ/dedup | — | ✓✓ | ✓✓ | ✓✓ | — | HIGH | Well covered |
| Jobs / worker | jobs, app.StartWorker | ✓✓ enqueue-atomic/retry-DLQ/reclaim/isolation | — | ✓✓ | ✓✓ | ✓✓ | ✓ | HIGH | Well covered |
| Idempotency (DB + http) | database.IdemStore, httpx | ✓✓ store/in-flight/concurrent + WithIdempotency | — | ✓✓ | ✓✓ | ✓✓ | — | HIGH | **Covered — not duplicated (G6 dropped)** |
| Documents / storage | document, storage | ✓✓ upload/scan/retention/grants/RLS | — | ✓✓ | ✓✓ | ✓✓ | — | HIGH | Well covered |
| **Document hooks** | document OnFileUpload/OnDocumentAccess | ✗ (nil hooks in every test) | **✓✓ hooks_fire_test** (upload hook aborts confirm; access hook denies download) | ✓ | ✓✓ | ✓✓ | — | MED | **Gap closed (G10)** |
| **Global (tenant-less) jobs** | jobs.EnqueueGlobal | ✗ | **✓✓ enqueue_global_test** (NULL-tenant row; nil/empty-kind rejected) | ✓ | ✓✓ | ✓✓ | — | MED | **Gap closed (G12)**; global RUN semantics noted |
| Seed catalog sync | seeds.Sync | ✓✓ via RunModuleContract (called 2× → idempotent) | — | ✓✓ | ✓ | ✓✓ | — | HIGH | **Covered — not duplicated (G11)** |
| HTTP list-query parse | httpx ParsePagination/Filters/Sort | ✓ via pagination/filtering primitives | — (thin adapters; accepted G9) | ✓ | ✓ | ✓ | — | LOW | Adapters over tested primitives |
| Comments / attachments | comment, attachment | ✓✓ CAS/author-guard/isolation | — | ✓✓ | ✓✓ | ✓✓ | — | MED | Well covered |
| Notify / webhook / integ | notify, webhook, integration | ✓✓ send-tx/backoff; verify/replay/breaker; provider | — | ✓✓ | ✓✓ | ✓✓ | — | HIGH | Well covered |
| Config / secrets | config, secrets, logging | ✓✓ layered/redaction/fingerprint/drift/unsafe-matrix | — | ✓✓ | ✓✓ | ✓ | — | HIGH | Well covered |
| HTTP primitives + health | httpx | ✓ problem-details/idem/etag/pagination/health | — | ✓✓ | ✓ | ✓ | — | MED | Well covered |
| Observability / perf | observability, prometheus, benchmarks | ✓✓ metrics/RED/AccessLog + 24 benchmarks | — | ✓✓ | ✓ | — | — | MED | Well covered |
| **Perf-budget CI gate** | internal/tools/benchbudget | ✗ (0%) | **✓✓ main_test** (baseName/loadBudgets±/parseBenchOutput) | ✓ | ✓✓ | — | — | LOW | **Gap closed (G8)** 0→61% |
| Module SDK / boot | module, app | ✓✓ RunModuleContract + boot gates + scratch-consumer | — | ✓✓ | ✓ | ✓ | ✓✓ | HIGH | Well covered |
| CLI tooling | internal/cli | ✓✓ all commands + golden | — | ✓✓ | ✓✓ | — | — | MED | Well covered |
| Delivery / E2E | migrations, internal/e2e | ✓✓ migrate ordering + scaffold→build→migrate→/healthz | — | ✓✓ | ✓ | ✓✓ | ✓✓ | HIGH | Well covered |
| kernel.New wiring | kernel | ✗ direct (indirect via app boot) | — (skipped: nil field fails app boot tests; low marginal value) | ✓ | — | ✓ | — | LOW | Indirectly covered (G7) |

Summary: the existing suite is strong across the framework. This effort closed the 6 confirmed
behavioral gaps (G1–G5, G8), correctly skipped the two non-gaps (G6 already covered, G7 indirect), and
surfaced one design finding (D1: `relationship.Relate` is platform-only and had no callers/tests).
