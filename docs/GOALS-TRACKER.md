# wowapi — Goals & Backlog Tracker

Companion to [SRS.md](SRS.md). Tracks every goal and work item by status: **Done · Deferred/Rescoped · Pending**.
Cross-checked against the tree on 2026-07-05 (Go 1.26; 329 Go files; 184 test files; 28 migrations).
The authoritative gate (`make ci` in containers, `WOWAPI_REQUIRE_DB=1`) is green on the current tree. The
D-0083→D-0085 review-follow-up commits (through `a1ee245`) are pushed and hosted GitHub CI is **green on
`a1ee245`** (all 5 workflows: ci, govulncheck, codeql, scorecard, security-scan). Note: `make ci` runs vet +
boundary lint (not the full `golangci-lint`); the B-1 lint backlog is **closed** (2026-07-05, D-0087) — full
`make lint` now reports 0, and `make lint-new` keeps changed code clean.

**Legend:** ✅ Done · 🟡 Partial · ⏸️ Deferred/Rescoped (documented, low-risk) · ⬜ Pending/backlog.

---

## 1. Top-level goals

| Goal | Source | What | Status |
|---|---|---|---|
| **Goal 1.0** | `Goal.md` (40 sections + patterns) | The framework itself — kernel, module SDK, all subsystems | ✅ Done (built via Phases 0–12) |
| **Goal 1.1** | `Goal 1.1.md` → `blueprint/11` | Framework-as-dependency: public vs internal surface, product-repo consumption, combined migrations, installable CLI, boundary rules | ✅ Done |
| **Goal 1.2** | `Goal 1.2.md` → `blueprint/12` | Config & deployment robustness: 5 typed config layers, precedence, secrets-by-reference, prod safety, per-process views, CLI config tooling | ✅ Done |
| **Goal 2** | `Goal 2.md` | 13-phase container-first build plan (Phase 0–12) with 25-point Definition of Done | ✅ Done (all phases) |
| **Hardening tranche** | `ROADMAP-wowapi.md`, `VERIFICATION-*.md` | Compliance/robustness pass (S/R/E/O items) + 3rd-party corrective actions CA-1…CA-15 | 🟡 Mostly closed (see §3–4) |
| **Docs & cleanup** | this session's `/goal` | README + user guide; git identity/signing cleanup; quality-gate culture; SRS + this tracker; retire prompt files from repo | ✅ Done |

---

## 2. Build phases (Goal 2 — Phase 0–12)

All phases delivered with per-phase evidence bundles (archived 2026-07-11 to `wowapi2/archive/evidence/phase-XX/`;
redirect map at `docs/implementation/evidence/README.md`) and decision records (`D-XXXX`). Kernel subsystems
verified present under `kernel/`.

| Phase | Deliverable | Status | Key code |
|---|---|---|---|
| 0 | Bootstrap: repo scaffold, Makefile, Dockerfile, compose, lint + boundary-lint, CI plan | ✅ | `Makefile`, `deployments/compose.yaml`, `.golangci.yml` |
| 1 | Config (typed layers, secret refs + redaction) + logging + app skeleton + per-process views | ✅ | `kernel/config`, `kernel/secrets`, `kernel/logging`, `app` |
| 2 | Database, migrations, tenant foundation (pgx pool, TxManager, RLS helpers) | ✅ | `kernel/database`, `migrations/00001–00006` |
| 3 | HTTP, errors, validation, pagination, idempotency, RouteMeta gate | ✅ | `kernel/httpx`, `kernel/errors`, `kernel/validation`, `kernel/pagination`, `kernel/filtering` |
| 4 | Identity, actor/capacity, RBAC/ReBAC/ABAC authz, policy, relationship, resource, record-level filter | ✅ | `kernel/auth`, `kernel/authz`, `kernel/policy`, `kernel/relationship`, `kernel/resource` |
| 5 | Module SDK (`module.Context`), seeds, testkit, neutral fixture, contract suite | ✅ | `module`, `kernel/seeds`, `testkit`, `internal/testmodules` |
| 6 | Transactional outbox, events, jobs (River), retry/DLQ, worker | ✅ | `kernel/outbox`, `kernel/jobs`, `migrations/00007,00013` |
| 7 | Rule engine + workflow engine (versioned, approval-gated, historical resolve, SLA sweeper) | ✅ | `kernel/rules`, `kernel/workflow`, `migrations/00008,00009` |
| 8 | Documents, files, comments, attachments, storage adapter, scan hook, grants, retention/redaction | ✅ | `kernel/document`, `kernel/attachment`, `kernel/comment`, `kernel/storage`, `migrations/00010` |
| 9 | Notifications, webhooks (in/out), integrations, breaker | ✅ | `kernel/notify`, `kernel/webhook`, `kernel/integration`, `migrations/00011,00022` |
| 10 | CLI, codegen, OpenAPI merge (`init/new-module/gen/migrate/seed/openapi/lint/config/deploy`) | ✅ | `cmd/wowapi`, `internal/cli` |
| 11 | Observability, performance, security hardening (metrics/tracing/health, redaction, fingerprint, benchmarks) | ✅ | `kernel/observability`, `adapters/{metrics,tracing}`, `bench-budgets.txt` |
| 12 | E2E acceptance & release readiness (external scratch-repo, api/worker/migrate smoke) | ✅ | acceptance tests; `make ci` green in containers |

---

## 3. Hardening tranche (ROADMAP `S/R/E/O`)

Delivered compliance/robustness primitives (all shipped with migrations + tests behind the gate):

| Item | Capability | Status | Code / migration |
|---|---|---|---|
| S1 | Machine auth / scoped **API keys** (issue/rotate/revoke/expire; sha256-only) | ✅ | `kernel/apikey`, `00019` |
| S3 | **Step-up / MFA** (`Permission.StepUp`, `Actor.AMR`, 401 + `WWW-Authenticate`) | ✅ | `kernel/authz`, gate test |
| S6 | **Audit hash-chain** (per-tenant seq + `row_hash`, `audit.Verify`) | ✅ | `kernel/audit`, `00017,00018` |
| E2 | **Retention / DSR / legal hold** (generalized hold, disposition registry, DSR ledger) | ✅ | `kernel/retention`, `00020` |
| E3 | **Gap-free sequences** (in-tx counter lock, audited voids) | ✅ | `kernel/sequence`, `00015` |
| E4 | **Artifact pipeline** (immutable, content-hashed, versioned) | ✅ | `kernel/artifact`, `00021` |
| R3 | **Leader-safe scheduler** (SKIP-LOCKED claim, exactly-once across replicas) | ✅ | `kernel/jobs`, `00014` |
| R6 | Legal-hold-vs-sweep race fixed (in-tx re-check) | ✅ | `kernel/retention` |

---

## 4. Third-party corrective actions (CA-1…CA-15)

From `VERIFICATION-wowapi-hardening.md` §6 (closure matrix), reconciled with post-push CI state.

| CA | Pri | Item | Status |
|---|---|---|---|
| CA-1 | P0 | Metrics emission (RED, scheduler lag, breaker, rate-limit, config fingerprint, DLQ depth) | ✅ Closed — DLQ-depth gauge `dlq_depth{queue}` now emitted leader-safe (B-8 done) |
| CA-2 | P0 | Default wiring (RateLimit in chain, real trace-sample-ratio, composite authn — API key + config-gated OIDC/JWT, signed cursors) | ✅ Closed — CA-2(a): `kernel/auth` OIDC/JWT `Authenticator` (JWKS RS256/ES256) shipped + wired; CA-2(b): authz cache `InvalidateAll` wired into `seeds.Sync` (D-0079) |
| CA-3 | P0 | API-key completion (Rotate, audited issue/rotate/revoke, CLI, ABAC-scope test) | ✅ Closed |
| CA-4 | P0 | Outbox hot-aggregate load characterization (~200 ev/s) | ✅ Closed |
| CA-5 | P0 | Recurring-job module face (`Context.RecurringJob`, leader-safe) | ✅ Closed |
| CA-6 | P0 | Hosted CI (unit + authoritative container gate + fuzz seeds) | ✅ Closed — **green on GitHub** for `78fcc0a` (run 28712692871); was "cannot prove without push". These doc commits are unpushed and re-run CI on push. |
| CA-7 | P0 | Traceability / status honesty (this matrix, ROADMAP pointer, CHANGELOG) | ✅ Closed |
| CA-8 | P1 | Idempotency-expired defined error (410 `KindIdempotencyExpired`) | ✅ Closed |
| CA-9 | P1 | Async trace propagation (`events_outbox.trace_context`, child spans) | ✅ Closed (residual: jobs/notify reuse same tracer seam) |
| CA-10 | P1 | Read-replica routing | ⏸️ **Rescoped** — deploy-time pool wiring, not kernel behavior (`WithTenantRO` exists); rationale recorded |
| CA-11 | P1 | Audit completeness (Context accessors, `tx_id`, `audit verify` CLI, anchor-export) | ✅ Closed — scheduled **anchor-export** now ships (`audit.ExportAnchors` on the leader-safe scheduler → append-only `audit_anchors`, migration 00027; `audit.CheckAnchor` detects tail-truncation `Verify` misses; B-3 done) |
| CA-12 | P1 | O2/O3/O5 ops finishers | ✅ Closed — O3 upgrade/deprecation policy ✅; **O2** schema-snapshot diff drill ✅ (B-4); **O5** scripted PITR + object-storage restore drills ✅ (B-5); production PITR provider-owned per D-0080 |
| CA-13 | P1 | Step-up gate test + API-key×step-up decision | ✅ Closed |
| CA-14 | P1 | Integration credential → `config.Secret` (structural redaction) | ✅ Closed (residual: per-tenant credential rotation runbook — `docs/operations/integration-credential-rotation.md` — ✅) |
| CA-15 | P1 | Unregistered notification channel fails loudly (no silent `sent`) | ✅ Closed |

---

## 5. Pending / deferred backlog

Everything below is **documented and low-risk**; none blocks v1 readiness. Ordered by rough priority.

| # | Item | Type | Where tracked | Notes |
|---|---|---|---|---|
| B-1 | ~~**golangci-lint backlog (~160)** — 154 `errcheck`, 3 `unused`, 2 `unparam`, 1 `unconvert`~~ | Code hygiene | `docs/working/lint-backlog.md` | ✅ **Closed 2026-07-05 (D-0087)** — `make lint` = 0. 149 CLI `fmt.Fprint*` terminal writes via one scoped `.golangci.yml` exclusion; 11 real code fixes (behavior-preserving, tests green). `make lint-new` keeps it closed. |
| B-2 | ~~Perf budgets for new hot paths~~ | Perf gate | `bench-budgets.txt` | ✅ **Done** — benchmarks + budgets added for audit Record/chain, sequence Allocate, token bucket, CachingStore hit/miss, edge middleware; `make bench-budget` now enforces them (30 benches). |
| B-3 | ~~CA-11 anchor-export job~~ | Feature | CA-11 | ✅ **Done** — `audit.ExportAnchors` on the leader-safe scheduler writes append-only anchors (migration 00027); `audit.CheckAnchor` detects offline tail-truncation. |
| B-4 | ~~**CA-12 O2** — schema-snapshot diffing in the reversibility drill~~ | Ops | CA-12 / D-0080 | ✅ **Done** — `scripts/migration_reversibility_drill.sh` (`make drill-reversibility`) diffs normalized `pg_dump --schema-only` snapshots after up→down→up; fails on asymmetric Down. |
| B-5 | ~~**CA-12 O5** — scripted PITR / object-storage restore legs~~ | Ops | CA-12 / D-0080 | ✅ **Done** — real scripted round-trips: `scripts/pitr_restore_drill.sh` (throwaway PG + WAL replay to target) + `scripts/object_storage_restore_drill.sh` (MinIO). Production PITR stays provider-owned (D-0080). |
| B-6 | ~~CA-10 read-replica router~~ | Feature | CA-10 | ⏸️ **Rescoped** to deployment — RO routing is a deploy-time pool-wiring choice (`WithTenantRO` exists); revisit only if a product needs kernel-level RO routing. |
| B-7 | ~~CA-6 reference-stack app-smoke~~ | CI | CA-6 residual | ✅ **Done (D-0089)** — `make smoke-reference` (CI job `reference-smoke`) scaffolds a product, runs it behind the reference nginx over TLS (`deployments/reference/smoke-compose.yaml`), and asserts the security-header posture is delivered THROUGH the proxy (`deployments/reference/smoke.sh`): the app's headers forwarded, TLS terminated, and HSTS owned authoritatively at the edge (nginx `proxy_hide_header`+`add_header`). This exercises the proxy/TLS wiring; the in-process posture is separately unit-tested in `kernel/httpx/edge_test.go`. |
| B-8 | ~~CA-1 DLQ-depth gauge~~ | Metrics | CA-1 residual | ✅ **Done** — `dlq_depth{queue="jobs"\|"events"}` emitted on the leader-safe scheduler; alert `WowapiDLQDepthHigh` added. |
| B-9 | ~~CA-9 jobs/notify trace sub-paths~~ | Observability | CA-9 residual | ✅ **Done (D-0088)** — jobs (`jobs.WithTracer`, `kernel/jobs/trace_test.go`) and notify (`notify.WithTracer`, `kernel/notify/trace_test.go`) capture the current request's W3C traceparent into the job/delivery envelope and continue it when the async runner/sender executes. |

---

## 6. Definition of Done (Goal 2, 25-point) — status
Container-first build ✅ · 24 real-DB test categories ✅ · per-phase evidence bundles ✅ · independent review per phase
✅ · boundary lint (no forbidden imports / no domain leakage) ✅ · deny-by-default authz proven ✅ · RLS isolation
proven ✅ · outbox atomicity + crash/retry proven ✅ · leader-safe scheduler exactly-once proven ✅ · secrets never
printed (verified) ✅ · generated module compiles + passes contract from an external repo ✅ · `make ci` green in
containers ✅ · hosted CI green on `329cc0e` (all 5 workflows) ✅ · no open critical/high review
findings ✅.

**Outstanding for a clean `v1.0.0` tag:** none — every tracked backlog item is closed or rescoped. **B-7**
(CA-6 reference-stack app-smoke) is now closed (D-0089): a CI job runs a scaffolded product behind the reference
nginx over TLS and smoke-tests the security headers through the proxy. B-1 (lint) is closed (D-0087, `make lint`
= 0); the ops finishers B-2…B-5 and B-8 are closed and B-9 is closed (D-0088, jobs/notify trace propagation
shipped + tested); B-6 is rescoped to deployment (see the backlog table). The generated-scaffold correctness
gaps (D-0083) and review follow-ups (D-0084/D-0085) are fixed. Pre-tag hardening is **done (D-0089)**:
golangci-lint is pinned (`GOLANGCI_VERSION`) and full-tree `make lint` is now
the enforced CI gate. None are architectural.

> **Status reconciliation (2026-07-11 architecture review).** The paragraph above describes the
> *v1.0.0-tag backlog*, which is indeed closed. It does **not** mean no architectural work remains: the
> Fable 5 architecture-review programme
> ([directive](implementation/architecture-directive-2026-07-11.md) →
> [final review](implementation/fable5-final-architecture-review-2026-07-11.md) →
> [closure-depth matrix](implementation/fable5-closure-depth-matrix-2026-07-11.md) →
> [implementation plan](implementation/premier-framework-implementation-plan.md)) opened 38+ findings,
> several P0/architectural (SEC-01, DATA-01/02/03/08, AR-01/02, FBL-01 kernel re-home). That programme
> is the authoritative outstanding-work register; this tracker's per-goal tables remain the record of
> *completed* goals.

> **Status reconciliation (2026-07-16 implementation autopsy).** The impl/ directory contains the authoritative
> execution ledger for the Waves 00–07 programme mandated in 2026-07-05. An independent third-party audit of
> that programme (Fable 5, `impl/reports/implementation-autopsy-report-2026-07-16.md`) found that the programme
> is **NOT complete** and the framework is **NOT production-ready**, despite commit `e8cda6b` claiming full
> finalization. Summary: **25 of 75 stories (33%) are fully verified; ~17 are implemented but incomplete or
> unreviewed; four waves have contradictory status ledgers; at least three false completion claims exist on
> security stories; one confirmed code defect contradicts an accepted acceptance criterion (webhook delivery
> in open transaction); the quality-floor gate was silently weakened with no record.**
>
> Wave verdicts (per `implementation-autopsy-report-2026-07-16.md` §14):
> - **W00, W01:** Accepted-with-reservations (4/6 and 7/10 verified respectively; review-gate evidence missing)
> - **W02:** Closure REJECTED (code substantially real but review gate falsely claimed; statuses contradictory)
> - **W03:** Open—implemented-unaccepted (substantive code but zero stories validly accepted; contract + tamper gaps)
> - **W04:** Acceptance REJECTED (E02 contains false acceptances and confirmed code defect; closure report is a template)
> - **W05:** Not executed as a wave (8 stories missing; FBL-01 done off-ledger; AR-01/02 built-not-wired; SEC-04 missing)
> - **W06:** Partially verified, wave open (E03 release gating + E04-S001 verified; 4 story claims unsupported)
> - **W07:** In progress (E01 perf work real but evidence mis-pinned; E02/E03 legitimately blocked; closure gate not run)
>
> This tracker's tables in §1–5 remain the record of *framework Goal 1.0/1.1/1.2 and Goal 2 Phases 0–12*,
> all completed before the Waves 00–07 programme began. The impl/ programme is a *subsequent* governance/
> verification layer on top of that code base. Consult `implementation-autopsy-report-2026-07-16.md` for the
> complete findings, remediation plan, and the programme's path to genuine closure.

---

## 7. Notes on retired working files
The AI-prompt / conversation files that seeded this project (`Goal.md`, `Goal 1.1.md`, `Goal 1.2.md`, `Goal 2.md`,
`goal-test.md`, `ROADMAP-wowapi.md`, `VERIFICATION-wowapi-hardening.md`, `WOW-Review.md`) have been **retired from
the repository** and their durable content folded into [SRS.md](SRS.md) and this tracker. As of 2026-07-11 they are
preserved in the **`wowapi2` documentation archive** (`archive/prompts-and-mandates/`, `archive/plans/`,
`archive/reviews/` — see its `ARCHIVE-INDEX.md` for the full path/provenance map), no longer as loose git-ignored
local files. `bench-budgets.txt` is **not** a prompt file — it is live perf-gate config (`make bench-budget`) and
stays tracked. `CHANGELOG.md` remains the release ledger.
