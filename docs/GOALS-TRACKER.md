# wowapi — Goals & Backlog Tracker

Companion to [SRS.md](SRS.md). Tracks every goal and work item by status: **Done · Deferred/Rescoped · Pending**.
Cross-checked against the tree on 2026-07-04 (50 commits; 251 Go files; 108 test files; 24 migrations; Go 1.26).
The authoritative gate (`make ci` in containers, `WOWAPI_REQUIRE_DB=1`) is green; the hosted GitHub CI last ran green
on `78fcc0a` (the tip before these docs). The SRS/tracker commits are local and pending the user's push, after which
CI runs on them.

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

All phases delivered with per-phase evidence bundles (`docs/implementation/evidence/phase-XX/`) and decision records
(`D-XXXX`). Kernel subsystems verified present under `kernel/`.

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
| CA-11 | P1 | Audit completeness (Context accessors, `tx_id`, `audit verify` CLI) | 🟡 Closed **except** scheduled **anchor-export** (⏸️ deferred, small/low-risk) |
| CA-12 | P1 | O2/O3/O5 ops finishers | ✅ Closed — O3 upgrade/deprecation policy ✅; **O2** schema-snapshot diff drill ✅ (B-4); **O5** scripted PITR + object-storage restore drills ✅ (B-5); production PITR provider-owned per D-0080 |
| CA-13 | P1 | Step-up gate test + API-key×step-up decision | ✅ Closed |
| CA-14 | P1 | Integration credential → `config.Secret` (structural redaction) | ✅ Closed (residual: per-tenant credential rotation runbook — `docs/operations/integration-credential-rotation.md` — ✅) |
| CA-15 | P1 | Unregistered notification channel fails loudly (no silent `sent`) | ✅ Closed |

---

## 5. Pending / deferred backlog

Everything below is **documented and low-risk**; none blocks v1 readiness. Ordered by rough priority.

| # | Item | Type | Where tracked | Notes |
|---|---|---|---|---|
| B-1 | **golangci-lint backlog (~160)** — 154 production `errcheck`, 3 `unused`, 2 `unparam`, 1 `unconvert` | Code hygiene | `docs/working/lint-backlog.md` | Gate blocks **new** code (`make lint-new`); backlog burned down package-by-package, never blanket-`//nolint`. |
| B-2 | ~~Perf budgets for new hot paths~~ | Perf gate | `bench-budgets.txt` | ✅ **Done** — benchmarks + budgets added for audit Record/chain, sequence Allocate, token bucket, CachingStore hit/miss, edge middleware; `make bench-budget` now enforces them (30 benches). |
| B-3 | **CA-11 anchor-export job** — scheduled audit-chain anchor for offline tail-truncation detection | Feature | CA-11 | Chain verify exists in-DB; anchor-export adds out-of-band tamper evidence. |
| B-4 | ~~**CA-12 O2** — schema-snapshot diffing in the reversibility drill~~ | Ops | CA-12 / D-0080 | ✅ **Done** — `scripts/migration_reversibility_drill.sh` (`make drill-reversibility`) diffs normalized `pg_dump --schema-only` snapshots after up→down→up; fails on asymmetric Down. |
| B-5 | ~~**CA-12 O5** — scripted PITR / object-storage restore legs~~ | Ops | CA-12 / D-0080 | ✅ **Done** — real scripted round-trips: `scripts/pitr_restore_drill.sh` (throwaway PG + WAL replay to target) + `scripts/object_storage_restore_drill.sh` (MinIO). Production PITR stays provider-owned (D-0080). |
| B-6 | **CA-10 read-replica router** (if ever needed in-kernel) | Feature | CA-10 | Rescoped to deployment; only revisit if a product needs kernel-level RO routing. |
| B-7 | **CA-6 reference-stack app-smoke** — nginx header smoke against a scaffolded running product | CI | CA-6 residual | Header posture already unit-tested (`kernel/httpx/edge_test.go`); needs a scaffolded product in CI. |
| B-8 | ~~CA-1 DLQ-depth gauge~~ | Metrics | CA-1 residual | ✅ **Done** — `dlq_depth{queue="jobs"\|"events"}` emitted on the leader-safe scheduler; alert `WowapiDLQDepthHigh` added. |
| B-9 | **CA-9 jobs/notify trace sub-paths** — extend async trace propagation beyond the outbox seam | Observability | CA-9 residual | Outbox is the primary fan-out; same tracer seam reused. |

---

## 6. Definition of Done (Goal 2, 25-point) — status
Container-first build ✅ · 24 real-DB test categories ✅ · per-phase evidence bundles ✅ · independent review per phase
✅ · boundary lint (no forbidden imports / no domain leakage) ✅ · deny-by-default authz proven ✅ · RLS isolation
proven ✅ · outbox atomicity + crash/retry proven ✅ · leader-safe scheduler exactly-once proven ✅ · secrets never
printed (verified) ✅ · generated module compiles + passes contract from an external repo ✅ · `make ci` green in
containers ✅ · hosted CI green on `78fcc0a` (these doc commits re-run it on push) ✅ · no open critical/high review
findings ✅.

**Outstanding for a clean `v1.0.0` tag:** burn down B-1 (lint backlog) to zero, and close the
low-risk ops finishers B-3…B-5. None are architectural.

---

## 7. Notes on retired working files
The AI-prompt / conversation files that seeded this project (`Goal.md`, `Goal 1.1.md`, `Goal 1.2.md`, `Goal 2.md`,
`goal-test.md`, `ROADMAP-wowapi.md`, `VERIFICATION-wowapi-hardening.md`) have been **retired from the repository** and
their durable content folded into [SRS.md](SRS.md) and this tracker. They are preserved **locally** (git-ignored) for
reference. `bench-budgets.txt` is **not** a prompt file — it is live perf-gate config (`make bench-budget`) and stays
tracked. `CHANGELOG.md` remains the release ledger.
