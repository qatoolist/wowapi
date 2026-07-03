# Implementation Phase Plan

Active goal: [Goal 2.md](../../Goal%202.md). Blueprint: [docs/blueprint/](../blueprint/README.md).
Phases follow Goal 2's dependency-aware ordering; each phase ends with a proof bundle under
`docs/implementation/evidence/phase-XX/` and a coherent commit.

| Phase | Scope (blueprint refs) | Depends on | Key acceptance criteria (10-delivery §2 / 12 §11) | Status |
|---|---|---|---|---|
| 0 | Preflight blueprint fixes; planning artifacts; go.mod; public package layout; `kernel/config` core types (Env, Secret, Framework skeleton); `module` + `app` skeletons; `cmd/wowapi version`; Makefile; Dockerfile + compose; boundary lint script; first tests (00 §1, 04 §1, 11, 12) | — | #20 (partial), #24; `make help` works; package graph rules encoded | **done** (`8eda353`) |
| 1 | Full `kernel/config` loader (layers, strict decode, unsafe-prod checks, fingerprint), `kernel/secrets`, `kernel/logging`, app process views, startup/shutdown skeleton, CLI `config validate/print/schema` framework-side (12) | 0 | #25–#28 (loader half) | **done** (evidence/phase-01) |
| 2 | `kernel/database` (pool, TxManager/TenantDB, SET LOCAL), `kernel/model`, kernel migrations 000–001, migration runner, tenant/user/access tables, testkit DB helpers + `AssertRLSIsolation` (03, 05 §2) | 1 | #2, #5, #22 | **done** (evidence/phase-02) |
| 3 | `kernel/errors`, `kernel/httpx` (middleware, RouteMeta), `kernel/validation`, `kernel/pagination`, `kernel/filtering`, idempotency helpers + migration 00003 (04 §4–5, 05 §1–2) | 1, 2 | #3, #11 | **done** (evidence/phase-03) |
| 4 | `kernel/auth` (OIDC), actor/capacity model, `kernel/authz` + `kernel/policy`, `kernel/relationship`, `kernel/resource`, denial audit, break-glass/impersonation (01 §3, 03 migrations 002–004) | 2, 3 | #4 (denials), authz matrix | **done** (evidence/phase-04) |
| 5 | Public `module` SDK (full Context), registries, seed loader, `internal/testmodules/requests` fixture, public `testkit`, contract suite, scratch-consumer test (06, 08 §2, 11) | 3, 4 | #1, #13, #15, #16, #19–#21 | pending |
| 6 | `kernel/outbox` + dispatcher + inbox, `kernel/jobs` (River), worker process, retries/DLQ (07 §3, §7; 03 migration 009) | 2, 5 | #9, #10 | pending |
| 7 | `kernel/rules` (versions, resolution, approval) + `kernel/workflow` (runtime, tasks, SLA sweeper, WorkflowSim) (02; 03 migrations 005–006) | 4, 6 | #7, #8 | pending |
| 8 | `kernel/document` + storage adapter/fake, comments, attachments, retention (07 §4; 03 migration 007) | 4, 6 | doc-flow tests | pending |
| 9 | `kernel/notify`, `kernel/webhook`, `kernel/integration` + circuit breaker (07 §5–6; 03 migrations 008, 010) | 6 | webhook replay/retry tests | pending |
| 10 | `cmd/wowapi` full: init/new-module/gen crud/migrate create/seed validate/openapi merge/lint boundaries/config tooling/deploy render; golden tests (08 §3, 11 §5, 12 §8) | 5 | #6, #12, #23, #28 | pending |
| 11 | Observability wiring, benchmarks + budgets, race gate, security test suite, config fingerprint drift (07 §1–2, §9) | 3–10 | #17, #18, #26–#27 | pending |
| 12 | E2E acceptance: scratch product repo, api/worker/migrate smoke, acceptance map complete, release notes, final Graphify + review pass | all | all remaining | pending |

Rules in force throughout: no deep features before layout/config/container/test harness are stable
(Goal 2 preflight rule 5); every deviation from the blueprint → entry in `decisions.md` before code;
every phase → review pass (architecture, security, boundaries, tests) recorded in its proof bundle.
