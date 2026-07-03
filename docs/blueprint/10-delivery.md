# 10 — NFR Matrix, Acceptance Criteria, Phase 0 Backlog, Boundary Check, First Files

## 1. NFR matrix

| NFR | Requirement | Design decision | Responsible component | Acceptance test | Risk if ignored |
|---|---|---|---|---|---|
| Security | deny-by-default authz; tenant isolation; audited sensitive actions | RLS+SET LOCAL; RouteMeta; layered evaluator; append-only audit | database, authz, httpx, audit | RLS isolation suite; route-metadata completeness test; denial-audit test | data breach, compliance failure |
| Performance | simple GET p95<50ms; authz<1ms cached; middleware<1ms | keyset pagination, allowlists, caches, pgx, budgets in CI | httpx, database, authz | `make bench` gates 2× regression | death by a thousand allocs |
| Scalability | 10× tenants without redesign | stateless api/worker, pooled PG, async via outbox/jobs, extraction seams | kernel at large | load test 1k rps read/100 rps write on reference hw | rewrite under growth |
| Reliability | no lost events/jobs; graceful deploys | transactional outbox, retries+DLQ, drain shutdown, migrations expand-contract | outbox, jobs | kill-worker-mid-batch test: zero loss/dupe effects | silent data loss |
| Maintainability | boundaries survive team growth | import lint, module contract tests, ADRs | platform, tools | `lint-boundaries` in CI | big ball of mud |
| Extensibility | new module w/o kernel edits | Module SDK + registries + seeds | platform | contract suite: sample module registers everything | fork-per-product |
| Observability | every request/job traceable to tenant+actor | slog+otel+metrics conventions | observability | log-field completeness test; trace continuity test | undebuggable prod |
| Testability | integration tests cheap & real | testkit, testcontainers, template-DB clone, fakes at edges | testkit | new module test suite runs <60s locally | mock-only false confidence |
| Dev experience | module in a day, CRUD in an hour | scaffolds, helpers, make targets | tools | timed walkthrough of `make new-module` → passing tests | boilerplate fatigue |
| Portability | no cloud lock-in in kernel | adapter ports (storage, secrets, mail), plain PG | adapters | compose-only local stack boots fully | vendor hostage |
| Compliance readiness | who/what/when/under-which-rule reconstructable | audit + rule provenance + temporal tables + retention jobs | audit, rules | historical resolution test; audit immutability test | failed audits |
| Operational simplicity | one DB, few processes, boring deploys | monolith, River-on-PG, managed PG | deployments | runbook rehearsal: deploy+rollback+restore | ops burnout |

## 2. Framework acceptance criteria (all executable via testkit)

1. A new module registers routes, permissions, roles, resource/relationship types, rule points,
   workflows, events, jobs, seeds, migrations **without modifying framework core** (contract suite).
2. Tenant-scoped repositories cannot execute without tenant context (compile-level: TenantDB; runtime: error; DB: RLS).
3. Routes cannot register without permission metadata unless explicitly `Public`.
4. Sensitive actions produce audit rows; permission denials on sensitive permissions are audited.
5. RLS isolation tests pass for every registered tenant-scoped table (catalog-driven sweep).
6. `make new-module` + CRUD gen yields a compiling, tested module in under an hour.
7. Workflow definitions load from seeds; tenant overrides resolve; version pinning holds.
8. Rule versions support draft→approval→active; historical `Resolve(at)` returns period-correct values.
9. Outbox events commit atomically with business writes (crash-injection test).
10. Jobs are idempotent (inbox) and tenant-aware (`SET LOCAL` verified in-worker).
11. All errors are RFC 9457 problem details with stable codes; pagination is uniform across modules.
12. OpenAPI generates and matches registered routes (CI diff).
13. Testkit creates tenants/users/parties/capacities/roles/assignments/resources in one-liners.
14. Kernel contains zero society-specific concepts (vocabulary lint green).
15. Modules import each other only via declared ports (import lint green).
16. Reusing the framework for a second product requires zero deletions of first-product code from core.
17. Performance budgets defined and enforced in CI; hot paths free of reflection/registry lookups.
18. Security guardrails enforced by middleware + route metadata + tests, not convention.

## 3. Phase 0 backlog (framework only — no product stories)

Each epic lists: stories → acceptance criteria (AC), dependencies (D), test coverage (T), risk (R).
Order is roughly dependency order; epics 1–6 are the critical path.

**E1. Project skeleton & CI** — repo layout, go.mod, Makefile, golangci-lint, boundary lint stub, Dockerfile, compose (pg+minio+mailpit), CI pipeline. AC: `make lint test` green in CI; compose boots. D: none. T: smoke. R: low.
**E2. Config & logging** — typed env config w/ validation-at-boot; slog JSON + redaction; request-id plumbing. AC: bad config fails boot with precise error; secrets never logged (test). D: E1. R: low.
**E3. Database & migration runner** — pgx pool, goose runner (kernel+module discovery), `cmd/migrate`, bootstrap migration (roles, extensions, `app_tenant_id()`). AC: fresh DB → migrate → idempotent re-run. D: E1. T: integration. R: low.
**E4. Tenant context & RLS** — tenants/users/user_tenant_access tables; tenant middleware; TxManager/TenantDB with SET LOCAL; RLS enablement pattern; `AssertRLSIsolation`. AC: criteria #2, #5. D: E3. T: isolation suite. R: **high — everything sits on this; build first, review hardest.**
**E5. Identity/auth foundation** — OIDC verifier middleware, JWKS cache, principal, revocation hook port, test-token issuer. AC: expired/wrong-aud rejected; fake IdP in tests. D: E4. R: medium.
**E6. Party/person/org/capacity foundation** — tables + kernel services + capacity middleware. AC: multi-capacity user must select; sole capacity implicit. D: E4. R: medium.
**E7. Resource registry & relationship framework** — type registries, resources mirror, relationships store, Checker. AC: registry sync at boot; temporal edge queries. D: E4. R: medium.
**E8. Authorization framework** — permissions/roles/assignments/policies tables, evaluator (RBAC→ReBAC→ABAC), Filter for lists, denial audit, break-glass + impersonation flows. AC: criteria #3, #4; permission matrix tests. D: E6, E7. T: authz matrix suite. R: **high.**
**E9. Audit framework** — partitioned append-only table, Writer in TenantDB, immutability grants, export job. AC: criterion #4; UPDATE attempt fails at DB. D: E4. R: medium.
**E10. Outbox/event framework** — outbox table, same-tx Writer, relay (SKIP LOCKED), dispatcher, inbox helper. AC: criterion #9; crash test. D: E4. R: high.
**E11. Background jobs** — River integration, Registry/Runner, retry/backoff/DLQ, job_runs mirror, schedulers, graceful drain. AC: criterion #10; kill-test. D: E10. R: medium.
**E12. Rule framework** — definitions/versions tables, registry, resolver w/ scope+temporal resolution, exclusion constraint, flags sugar, seed bundles. AC: criterion #8. D: E4 (E13 for approval flow — stub until then). R: medium.
**E13. Workflow framework** — definition schema+validation, runtime (start/decide/delegate/override), tasks+assignees, SLA sweeper, WorkflowSim. AC: criterion #7; generic approval flows tested. D: E8, E10, E11. R: **high — biggest single build.**
**E14. Document/file + comment/attachment framework** — tables, storage port + s3 adapter + fake, presign flows, scan hook, grants, retention sweep; comments/attachments services. AC: upload→scan→download w/ audit; grants honored. D: E8, E11. R: medium.
**E15. Notification framework** — templates+deliveries, dispatcher, channel ports, fakes, preferences, retries. AC: template resolution tenant→platform→locale; dead-letter path. D: E11. R: low.
**E16. Webhook & integration framework** — provider registry, inbound verify/replay/ingest, outbound sign/deliver/breaker, admin redeliver. AC: replayed event ignored; breaker opens/half-opens (fake time). D: E11. R: medium.
**E17. API/error/validation/pagination helpers** — httpx toolbox, ProblemError mapping, validator wrapper, cursor codec, allowlist builders, RouteMeta enforcement. AC: criteria #3, #11. D: E2. R: low (build early, everything uses it — start alongside E4).
**E18. Base model & DTO primitives** — kernel/model structs, response envelopes, ETag/If-Match helpers, IdemStore + WithIdempotency. AC: idempotent replay test; 412 on stale version. D: E17. R: low.
**E19. Module SDK & bootstrap** — Module/ModuleContext, registries, Validate, seed loader, Kernel/App composition root, sample `requests` module, contract suite. AC: criteria #1, #15, #16. D: E4–E18 interfaces (can start early against stubs). R: high.
**E20. Testkit** — everything in 08 §2. AC: criterion #13; used by all epics' tests (grow it continuously, finalize here). D: E4+. R: medium.
**E21. CLI/codegen/templates** — module generator, CRUD generator, seed validator, openapi merge. AC: criterion #6. D: E19. R: low.
**E22. Observability & ops** — otel+metrics wiring, health/readiness, dashboards-as-code starter, runbooks, backup/restore rehearsal doc. AC: log-field + trace tests; alert rules for outbox lag/DLQ. D: E11. R: low.
**E23. Performance benchmarks & security tests** — budget bench suite, race gate, OWASP checklist tests (BOLA probes, injection probes via allowlists, upload abuse). AC: criteria #17, #18. D: E19. R: medium.

## 4. Reference-domain boundary check ✅

Confirmed **absent from core** (and lint-denylisted): building, wing, flat, housing owner, society
member / associate member / nominal member, committee, chairman, secretary, treasurer, maintenance
bill, defaulter, AGM, society notice, parking allocation, water/STP/WTP, visitor gate entry,
conveyance, redevelopment, bye-law elections. Grep of this blueprint's kernel sections finds these
words only in "do not include" statements and the society-module illustrations.

A future `society` module provides all of them via the extension points, kernel untouched:
tables+resource types (`society.building/unit/notice/complaint/bill`), relationships
(`society.owner_of_unit`, `society.occupier_of_unit`, `society.member_of`), roles
(`society.tenant.chairman/secretary/treasurer`, term-limited assignments), workflows (membership
approval, notice approval, bill approval, complaint escalation — YAML), rule points
(`society.agm.notice_period_days`, `society.billing.frequency`, `society.defaulter.threshold_amount`,
`society.parking.eligibility`), document classes (notice, minutes, share certificate), notification
templates, jobs (billing generation), reports. Elections/AGM voting reuse the generic `vote` step type.

## 5. Final recommendation — first 10 things to create

Build the walking skeleton in this order; each lands with its tests:

1. `go.mod`, `Makefile`, `.golangci.yml`, `deployments/compose.yaml` — E1.
2. `/internal/kernel/config` + `/internal/kernel/logging` — E2.
3. `/migrations/000_bootstrap.sql` + `/internal/kernel/database` (pool, goose runner, **TxManager/TenantDB with SET LOCAL**) — E3/E4 core.
4. `/migrations/001_tenants_users.sql` + `/internal/kernel/tenant` (context, resolver middleware).
5. `/internal/testkit` (NewDB, CreateTenant, **AssertRLSIsolation**) — proves #3/#4 immediately.
6. `/internal/kernel/errors` + `/internal/kernel/httpx` (ProblemError, helpers, RouteMeta router).
7. `/internal/kernel/auth` (OIDC verify + test-token issuer) + `/internal/kernel/authz` skeleton (Actor, Evaluator with RBAC path).
8. `/internal/platform` (Module, ModuleContext, App/Kernel composition root, Validate).
9. `/internal/modules/requests` — the neutral sample module driving SDK ergonomics + contract suite.
10. `/cmd/api`, `/cmd/worker`, `/cmd/migrate` mains wired to the composition root.

From there, follow the backlog (outbox → jobs → rules → workflow → documents → notify → webhooks).
The framework is "done enough for product work" when the acceptance criteria in §2 are all green —
at that point the society module is a new folder, seed files, and business code, and the same is
true for a school, club, or vendor-management product.
