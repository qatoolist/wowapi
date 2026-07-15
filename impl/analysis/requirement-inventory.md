---
id: ANALYSIS-REQ-INV
type: analysis
title: Requirement inventory — canonical item register with dispositions and targets
status: complete
created_at: 2026-07-12
updated_at: 2026-07-12
derived: false
---

# Requirement inventory (canonical)

Every material actionable item from the primary sources, with stable ID (existing IDs retained
per mandate §1.1/§5), classification (§1.3), priority, disposition (§1.4), and target
wave/epic/story. This file is the **canonical allocation**; the traceability matrices in
`../tracking/` are derived views.

Sources: `premier-framework-implementation-plan.md` (PLAN §5; 38 findings),
`fable5-final-architecture-review-2026-07-11.md` (REVIEW; FBL/D/T items),
`fable5-closure-depth-matrix-2026-07-11.md` (MATRIX; CS-01..CS-25 closure specs = the dedup
layer over 30 §H + 20 §I capability rows), plus the post-review session delta (git history
`4eca9f4..0a31186`: CI parallelization #23/#24, sweep-bench fix + budget recalibration #25).

Classification key: IMPL (implementation requirement), ARCH (architecture decision), CONSTR
(design constraint), VERIF (verification requirement), DOC (documentation requirement), OPS
(operational requirement), QG (quality gate), RISK, TD (technical debt), FUT (future
enhancement), REJ (rejected), SUP (superseded), INFO.

Disposition key: `planned`, `implemented-needs-verification` (INV), `partial`, `blocked`,
`deferred`, `rejected`, `superseded`, `duplicate`, `not-applicable`.

## A. Plan findings (PLAN §5 — 38 findings; per-task detail lives in the plan's own tables, which remain the task-level source of record)

| ID | Title | Class | Pri | Disposition | Target | Notes |
|---|---|---|---|---|---|---|
| AR-01 | Ownership-bound ApplicationModel (T1–T11) | IMPL | P1 | planned | W05-E01-S001..S004 | Core; D-02/D-03 resolve its open questions |
| AR-02 | Typed port keys + compiled provider graph (T1–T7) | IMPL | P1 | planned | W05-E02-S001..S003 | Depends AR-01 T1/T2 |
| AR-03 | One authoritative declaration, derived projections (T1–T5) | IMPL | P1 | planned | W05-E03-S001..S002 | T2 = DX-06 duplicate → single owner DX-06 (see duplicate-analysis) |
| AR-04 | Eliminate boot-time silent behaviour | IMPL | P1 | partial | W05-E03-S002 | T1 EXECUTED (verified ×2); T2–T5 planned, dep AR-01; T5 waiver shared w/ SEC-06/DX-07 *(target corrected 2026-07-12: was S003, canonical grouping is E03-S002 — see change-log)* |
| AR-05 | Composition/doc drift removal | DOC/QG | P1 | partial | W06-E04-S002 | T1/T2 EXECUTED; T3 doc-example CI gate (CS-22 spec), T4/T5 dep AR-03 |
| AR-06 | Remove hidden constructor bypasses | IMPL | P1 | partial | W05-E04-S001 | T1 EXECUTED; T2 lint + T3 audit planned |
| SEC-01 | Server-side tenant/privileged session state (T1–T7) | IMPL | P0 | planned | W03-E01-S001..S004 | D-01 ratified; DEC-Q1 safe default unblocks; BREAKING wowsociety |
| SEC-02 | Workflow privileged ops fail closed | IMPL | P0 | partial | W03-E05-S001 | T1–T3 EXECUTED (verified ×2); T4 ratification design + T5 audit remain |
| SEC-03 | Webhook replay bound to authenticated data (T1–T4) | IMPL | P1 | planned | W03-E03-S001 | Breaking Verifier interface |
| SEC-04 | Bound authz staleness/memory | IMPL | P1 | planned | W05-E04-S002 | CS-17: LRU (approved dep) + epoch table (D-06); P0 if cache prod-enabled |
| SEC-05 | Versioned security verification profile | QG | P1 | planned | W07-E02-S001 | Wave-6-class per plan |
| SEC-06 | Outbound-security escape-hatch governance | IMPL | P1 | planned | W03-E02-S001 | D-07 ratified |
| DATA-01 | Composite tenant FKs (T1–T8) | IMPL | P0 | planned | W02-E02-S001..S002 | Dep DATA-09 T1–T5 for risky steps; wowsociety has own instance (product-level, PROD-01) |
| DATA-02 | Lease generations/fencing + idempotency (T1–T7) | IMPL | P0 | planned | W04-E01-S001..S003 | T1 shared primitive is keystone |
| DATA-03 | Remote I/O outside DB transactions (T1–T8) | IMPL | P0 | planned | W04-E02-S001..S002 | Scope refined by MATRIX CS-11: external effects only; T7 = DATA-08 W0-T2 duplicate (done) |
| DATA-04 | Bulk multi-worker safety (T1–T6) | IMPL | P1 | planned | W04-E03-S001 | T1 stopgap can land early in-wave |
| DATA-05 | Version allocation races + blob GC (T1–T5) | IMPL | P1 | planned | W02-E03-S001 | |
| DATA-06 | Resource-mirror aggregate write contract (T1–T4) | IMPL | P1 | planned | W02-E04-S001 | T2 shared fix w/ DATA-07 T3 (one owner) |
| DATA-07 | Relationship semantics + actor attribution (T1–T4) | IMPL | P1 | blocked→planned | W03-E04-S001 | HARD dep SEC-01; secondary SEC-04 |
| DATA-08 | Compliance evidence complete/durable | IMPL | P0/P1 | partial | W04-E04-S001..S002 | W0 slice EXECUTED (verified ×2); W6-T1 hash widening (D-04) + T2–T5 planned |
| DATA-09 | Online expand/backfill/validate/contract protocol (T1–T9) | IMPL | P0 | planned | W02-E01-S001..S003 | Precedes DATA-01 T4/T5 + DATA-08 W6-T1; T9 CI drills |
| DX-01 | Source-built CLI path validity (incl. T5 scaffold harness) | IMPL | P0 | planned | W01-E04-S001 | T5 harness = shared primitive for DX-02/DX-04 |
| DX-02 | Generator emits in-set verb + boots (PF-2) | IMPL | P0 | planned | W01-E04-S001 | One-token fix + generator-boots test (CS-14, re-verified at HEAD) |
| DX-03 | Module DSL design | ARCH/FUT | P1 | deferred | W06-E01-S001 | Design-investigation story only (Wave-4-class per plan) |
| DX-04 | Golden consumer + upgrade matrix | IMPL | P1 | accepted | W06-E01-S002 | AC-01..AC-05 pass: versioned install; 8 subsystems/2 modules; real infrastructure; tagged v1.1.0→local candidate replay; Wave-4 required gate; final independent PASS EV-W06-E01-S002-014 |
| DX-05 | CLI/docs/version identity singular | DOC | P1 | partial | W01-E04-S002 | T1/T2 EXECUTED; §6-vs-§9 status inconsistency = T-DOC-01 |
| DX-06 | OpenAPI merge complete-or-loud (T1–T3) | IMPL | P1 | planned | W06-E02-S001 | Single owner of AR-03 T2 scope; validator dep decision at impl |
| DX-07 | Truthful readiness/config diagnostics (T1–T4) | IMPL | P1 | planned | W04-E04-S003 | T4 dep AR-04 T5 waiver mechanism |
| PERF-01 | Token-bucket sweep fix | IMPL | P0 | INV | W00-E01-S002 | EXECUTED + #25 recalibrated sweep budgets to honest full-map measurements — verify at current HEAD |
| PERF-02 | Complete-request benchmarks vs real PG | IMPL | P1 | accepted | W07-E01-S001 | Six real-PG profiles, 36 cold/warm × concurrency cells, and six-part attribution published; absolute SLO gated on open DEC-Q9 |
| PERF-03 | Rules resolution bounded SQL | IMPL | P1 | accepted | W07-E01-S002 | Parity, current/history index access, constant 8/8/8 SQL count at depths 3/10/50, and live-update visibility accepted |
| PERF-04 | Sweeper/worker N+1 + unbounded materialization | IMPL | P1 | accepted | W07-E01-S003 | 100-row bound, fixed/batched queries, W04 lease/fencing consumption, and relative evidence accepted; absolute SLO conditional on DEC-Q9 |
| PERF-05 | Explicit object checksum behaviour | IMPL | P2 | accepted | W07-E01-S004 | Required checksums, labeled bounded repair, metrics, resumable no-duplicate backfill, and conditional publication accepted |
| PERF-06 | Fail-closed performance gates | QG | P1 | INV | W00-E01-S002 | T1 EXECUTED; T3/T4 fuzz scope = REL-04 T8 single-owner (W07-E02-S002) |
| REL-01 | Release gated on exact published commit (T1–T9) | IMPL | P0 | planned | W06-E03-S001..S002 | ~85% buildable now; final activation = DEC-Q10 (admin) |
| REL-02 | Security checks blocking-or-replaced | QG | P0/P1 | planned | W06-E03-S003 | Trivy soft-fail + visibility-guard dormancy cited at exact lines (MATRIX CS-23) |
| REL-03 | Compatibility gates (split a/b) | QG | P1 | planned | W06-E02-S002..S003 | a=T1,T2,T4,T6,T8,T9 now; b=T3(DX-06),T5(AR-03/DX-03),T7(DX-04) |
| REL-04 | Truthful integration coverage | QG | P1 | partial | W07-E02-S002 | T1–T4 EXECUTED (verified ×2); T5–T8 planned (T8 owns fuzz, shared w/ PERF-06 T3/T4) |

## B. Review findings + decisions (REVIEW §O/§U; MATRIX specs give the closure detail)

| ID | Title | Class | Pri | Disposition | Target | Notes |
|---|---|---|---|---|---|---|
| FBL-01 | Kernel re-home (9 pkgs → foundation/) | IMPL | P1 | planned | W05-E05-S001..S002 | CS-01 mechanics; dep AR-01/02; wowsociety mfa migration story is product-coordination (PROD-02) |
| FBL-02 | Production seed-sync path (PF-9) | IMPL | P0-prod | planned | W02-E05-S001 | CS-21 acceptance bar fixed; design detail = story investigation task |
| FBL-03 | Reconcile wowsociety upstream register | DOC | P2 | planned | W01-E04-S002 | Mark PF-2/PF-6/RFF-001 etc. as closed when their fixes land |
| FBL-04 | Adopt cenkalti/backoff for duplicated retry | IMPL | P1 | planned | W04-E02-S003 | Approved dep; parity + fault-injection tests |
| FBL-05 | Enable zero-cost leak linters (sqlclosecheck etc.) | QG | P1 | planned | W01-E01-S001 | Counts measured at HEAD (CS-23); noctx 2 prod fixes |
| FBL-06 | OTel trace/log correlation + pgx tracer (D-08) | IMPL | P1 | planned | W01-E02-S001..S002 | CS-05 T1–T3 |
| FBL-07 | Utilisation closure (gosec triage, go mod verify, license signal, nightly fuzz, hook DB-skip) | QG | P1/P2 | partial | W01-E01-S002..S003 | Nightly ci schedule EXISTS since #24 (fuzz portion still seed-replay only) |
| FBL-08 | Central validation enforcement (RouteMeta seam) | IMPL | P1 | planned | W01-E03-S002 | CS-08 T1–T3; compat: profile-flag first |
| FBL-09 | HTTP server timeouts + CSRF body bound | IMPL | P1 | planned | W01-E03-S001 | CS-09; template-delivery model (wowsociety backport = PROD-03) |
| D-01..D-09 | Nine ratified architecture decisions | ARCH | — | planned | W00-E02-S003 | ADR-ification story; enacted inside their target stories; DEC register tracks |
| T-DOC-01 | Fix plan §6-vs-§9 DX-05 inconsistency | DOC | P3 | planned | W01-E04-S002 | |
| T-TEST-01 | Diagnose intermittent e2e full-suite failure (re-scoped) | VERIF | P2 | planned | W01-E04-S003 | Reproduce-first; original shared-DB diagnosis withdrawn |
| DEC-Q1 | IdP grant_id claim contract | ARCH | P0 | blocked (human) | W03-E01 (tracked) | Safe default unblocks build (REVIEW §F Q1) |
| DEC-Q9 | Reference-perf-env ownership | OPS | P1 | blocked (human) | W07-E01 (tracked) | Provisional default set (GH runner + reference json) |
| DEC-Q10 | Repo-admin activation (branch/tag/env protection) | OPS | P0 | blocked (human) | W06-E03 (tracked) | Merge-queue rulesets unavailable on user-owned repo (session fact) |

## C. Matrix verify-outcomes and constraints (no new work; recorded so nothing silently disappears)

| ID | Title | Class | Disposition | Notes |
|---|---|---|---|---|
| CS-03 | Config fail-closed + fingerprint | VERIF | INV→verified | Re-earned with citations in MATRIX; W00 pins evidence pointer |
| CS-19 | i18n freeze + key-echo fallback | VERIF | INV→verified | Same |
| CS-24 | SSRF dial-time guard | VERIF | INV→verified | Verified strength; gosec G704 annotation task inside FBL-07 |
| CS-10 | pgx rows contract | CONSTR | planned | Decided: keep raw pgx.Rows; FBL-05 enforces mechanically; pool lifetime config keys = task in W01-E01-S001 |
| CS-25 | Secrets rotation contract (D-09) | DOC/OPS | planned | Restart-based rotation documented; file-provider = deferred (DEF-01) |
| K-RETAIN | §K retained customs (config, i18n, scheduler, rate-limiter, TOTP) | ARCH | not-applicable | Justified retentions; no work |
| K-P2 | gobreaker (breaker) + jwx (JWKS) evaluations | FUT | deferred | DEF-02/DEF-03 with reopen triggers per §K |
| M-REJ | Rejected deps (viper/envconfig, kernel message bus, custom crypto) | REJ | rejected | Rationale in REVIEW §M |
| B11/B12/B13 | Parked P2 backlog (radix router, schema unification, hot overlays) | FUT | deferred | DEF-04..06; reopen triggers in framework-backlog-p2-decisions.md (D-0090) |

## D. Product-level items (framework boundary per mandate §2.3 — recorded, excluded from framework implementation)

| ID | Title | Rationale | Enabling framework capability |
|---|---|---|---|
| PROD-01 | wowsociety `policy_override` composite FK | Product schema fix | DATA-01 T1 (parent unique index) + DATA-09 protocol |
| PROD-02 | wowsociety `kernel/mfa` import migration (5 identity files) | Product code migration | FBL-01 re-home ships deprecated forwarding shim |
| PROD-03 | wowsociety readiness/timeout backports to committed main.go | Product hand-edit | DX-07 T1 + FBL-09 fix the templates |
| PROD-04 | SEC-01 impersonation cutover (whoami/impersonation/tests) | Product auth flow rework | SEC-01 T1/T5 grant contract + coordinated rollout plan |
| PROD-05 | DATA-08 W6 staging audit re-verification before version bump | Product compliance drill | hash_version branch verification (D-04) |

## E. Session delta (git history 4eca9f4..0a31186 — facts newer than the primary sources)

| ID | Fact | Effect on inventory |
|---|---|---|
| SD-01 | CI gate parallelized (3 legs), toolbox image GHA-cached, docs-only skip (#23) | Quality-gate baseline changed; W00 baseline captures new wall-clocks |
| SD-02 | Bench path-scoped on PRs; nightly schedule; merge_group support (#24) | Partially advances FBL-07 (nightly exists); REL-04 T8 fuzz remains |
| SD-03 | Sweep-bench O(n²)+empty-map fix; budgets recalibrated (#25) | PERF-01 evidence basis changed — W00-E01-S002 verifies against NEW budgets |
| SD-04 | Doc archival to wowapi2 (#22) | Historical docs out of repo; archive index is the provenance record |

**Totals:** 38 plan findings (8 partial/INV, 27 planned, 1 deferred-design, 1 blocked→planned,
1 accepted) · 15 review/decision items (12 planned/partial, 3 human-blocked) · 9
matrix-outcome/constraint rows (3 verified-INV, 2 planned, 4 deferred/rejected/n-a) · 5 product-level ·
4 session facts.
No source item dropped; duplicates consolidated with single owners recorded above and in
`duplicate-analysis.md`.
