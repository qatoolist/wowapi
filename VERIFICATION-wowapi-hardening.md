# Verification Report — ROADMAP-wowapi.md Hardening Implementation

**Date:** 2026-07-04 · **Scope:** all items in [ROADMAP-wowapi.md](ROADMAP-wowapi.md) (S1–S8, R1–R8, O1–O5, E1–E6, method §5) against `github.com/qatoolist/wowapi` @ `8dabe2b`.

> **Scope boundary — read first:** the hardening implementation exists **only in the `wowapi` framework repository**. No product (`wowsociety`) repository has been created — no scaffold, no README, no code. `wowsociety.app` (the legacy 16-service codebase) was **not** remediated by this pass and remains exactly as described in ROADMAP.md §5/§7 — discovery input only. Every product-side statement in this report and in PRODUCT-PLAN.md (pilot module, repo bootstrap, all epics E0–E11) is **planned, not started**.
**Method:** three independent verifiers read source, checked real wiring (kernel bootstrap, `app/worker`, generated `wowapi init` templates, CLI, Makefile, migrations, deployments), and **ran the tests against live compose Postgres with `WOWAPI_REQUIRE_DB=1`** (so DB-gated tests could not silently skip). Docs, CHANGELOG, and decision records were treated as claims, not evidence. The drill scripts were executed, and fuzz targets were run (`make test-fuzz FUZZTIME=8s`).

## 1. Overall verdict

**Substantially implemented, high quality, honestly test-evidenced — but NOT complete to the roadmap's acceptance criteria, and the exit gate ("all P0 rows closed") is not met.** The dominant failure pattern is not bad code; it is **well-built, well-tested kernel primitives that nothing wires in by default**, plus two acceptance areas closed "doc-only" without the required evidence.

Positives verified first-hand: 35/35 kernel packages pass on a real database with skip-guarding enforced; migrations 00012–00022 all reversible (the drill caught and fixed a real pre-existing bug in 00010's Down); the backup drill script executes successfully; fuzz targets run clean (3M+ execs); append-only and RLS properties are grant-enforced and proven by tests that attempt UPDATE/DELETE; the diff since the pre-hardening commit (+10,335 lines, 136 files) contains **no unrelated or regression-inducing changes**.

## 2. Verdict matrix

| Item | Verdict | Key gap vs acceptance |
|---|---|---|
| S1 machine auth | **PARTIAL** | Kernel package complete (sha256-only, constant-time, RLS, ABAC-deny structurally overrides scope) but **not wired**: generated api ships `DenyAllAuthenticator` + TODO; no `wowapi apikey` CLI; no Rotate API (two-call pattern); **issuance/rotation not audited**; ABAC-deny-vs-scope claimed "tested" but isn't |
| S2 rate limiting | **PARTIAL** | Middleware + token bucket + 429/RFC 7807 real and tested; **not in the default chain**; per-permission only via DIY keyFn; **no rate-limit metrics** |
| S3 step-up/MFA | **PARTIAL** | Evaluate-level logic + gate header code exist; Evaluate tests pass; **gate 401/`WWW-Authenticate` untested**; no challenge/TOTP verifier interface; API-key actors can never satisfy step-up (fail-secure, undocumented) |
| S4 credential encryption | **NON-GAP (reclassified, defensible)** | Roadmap premise was wrong: credentials are `secretref://` references only; plaintext rejected at write; redaction structural. Residual: `integration.Config.Credential` is a plain `string` (not `Secret`), rotation procedure not documented as a wowapi procedure |
| S5 idempotency expiry | **PARTIAL** | `SweepExpired` + migration 00012 + **actually scheduled** (leader-safe, 1h default); but **replay-after-expiry silently re-executes** — the roadmap required a defined error; no such error kind exists |
| S6 audit tamper-evidence | **FULL** | Per-tenant seq + hash chain, atomic in business tx; `Verify` detects mutation and deletion (proven via admin-connection tampering tests); wired as the kernel's default audit sink. Caveats: tail-truncation detectable only against an externally exported anchor, and no anchor-export job/CLI exists |
| S7 reference deployment + edge | **PARTIAL** | nginx reference + smoke.sh + ops checklist real; SecureHeaders/CORS(deny-by-default)/BodyLimit/Timeout implemented and wired into the generated api (E2E scaffold-build test passes); **no CI smoke against the reference stack** (explicit recorded deviation) |
| S8 adversarial testing | **PARTIAL** | Fuzz FULL (filter DSL + cursor; `make test-fuzz`; run clean); **authz property tests missing** (silently dropped; only table-driven tests exist) |
| R1 authz caching / replicas | **PARTIAL-SUPERFICIAL / MISSING** | `CachingStore` exists with passing unit tests but has **zero non-test call sites**; `Invalidate` never called (no kernel spine-write path to hook — invalidation is entirely the product's problem, i.e., the exact stale-allow hazard targeted); **read-replica routing does not exist** (single pool; `WithTenantRO` = READ ONLY tx on same pool) |
| R2 advisory-lock characterization | **MISSING** | No load test, no throughput envelope, no sub-sharding doc. Declassified to "doc-only" in hardening-plan.md — disclosed there with rationale, but without the required load evidence, while the roadmap lists it P0 with full acceptance |
| R3 sweeper | **PARTIAL** | Leader-safety FULL and elegant (`FOR UPDATE SKIP LOCKED` claim + advance-in-claim-tx; 6-replica test proves exactly-once); wired for idempotency/SLA/retention sweeps; but interval configurable **only by editing generated main.go** (no config key), and lag is a log line, **not a metric** |
| R4 DLQ operability | **PARTIAL** | `wowapi dlq <jobs|events> <list|inspect|replay|discard>` CLI real, wired, DB-tested; event replay idempotent via processed-events dedupe; **jobs replay idempotency is contract-only**; **discard is an unaudited bare DELETE**; **no DLQ depth metric** |
| R5 delivery receipts | **FULL** | Receipts + provider msg IDs + RLS + channel prefs (opt-out actually consulted at send) all real and tested. Caveats: unregistered channel falls to a noop sender that marks `sent` (silent-success hole); prefs checked at enqueue only |
| R6 legal-hold race | **FULL** | In-tx re-check two ways (`FOR UPDATE` candidate select + guarded UPDATE); genuinely concurrent race test passes under `-race`. Caveat: mechanical guarantee only for the kernel document class; generic retention engine delegates hold-skipping to the product's DisposeFunc by contract |
| R7 cursor versioning | **PARTIAL** | Signature + rejection complete and tested; **zero production callers**; the kernel's own `workflow.OpenTasksFor` mints **legacy unsigned cursors** that bypass the check by design; generated CRUD list endpoint is a TODO stub |
| R8 webhook granularity | **PARTIAL** | Per-endpoint breaker (5 failures/5m cooldown, half-open) + per-delivery retry budget (5, hardcoded) real, tested, documented; **endpoint-health metric missing** (status only visible as a DB column) |
| O1 tracing | **SUPERFICIAL** | Port + NoOp + OTel adapter (kernel stays otel-free) + HTTP server span all real; but **no async propagation** (zero Inject/Extract in outbox/jobs/notify — the roadmap's core acceptance), adapter has zero callers, and the documented `cfg.TraceSampleRatio` config key **does not exist**. Commit message "end-to-end OTel tracing" overstates |
| O2 migration harness | **PARTIAL (near-full)** | Reversibility drill is a real Go test (head → reset → re-up) run under `make ci-container` with skip-guarding; expand/contract documented well; caught a real 00010 bug. **No schema snapshot diffing**; no data-rich seeded drill |
| O3 upgrade discipline | **PARTIAL** | `testkit.RunModuleContract` tripwire real and passing; **no published deprecation policy** beyond one pre-existing blueprint paragraph; no documented upgrade procedure |
| O4 config-drift alerting | **PARTIAL** | Fingerprint detection + `/readyz` exposure + checklist section real; the reference Prometheus rule **targets metrics that nothing exports** (and `/readyz` is JSON, not scrapeable) — the exporter is left as an exercise |
| O5 backup/restore | **PARTIAL (honest)** | Doc covers PITR + object-storage ordering invariant; drill script **executed successfully** (logical dump → scratch restore → marker verify). Logical-dump only: PITR and object-storage restore are documented, not scripted |
| E1 field audit | **FULL** | Record/Query in-business-tx, redaction hook, append-only grant-proven; kernel itself uses it (durable authz-denial sink). Caveats: **no `tx_id` column** (roadmap listed it); no `module.Context` accessor (deferred, self-disclosed) |
| E2 retention/DSR | **FULL** | Generalized legal hold + DSR ledger with statutory-override + per-class disposition registry + scheduled disposition wired on the leader-safe scheduler; the only new capability with a Context accessor (`RetentionClasses()`). Retention durations are enforced by product callbacks by design |
| E3 sequence allocator | **FULL** | Counter-row lock inside the business tx (rollback frees the number — gap-free), audited voids; concurrency and rollback gap-freedom proven on real Postgres |
| E4 artifact pipeline | **FULL** | Versioned immutable artifacts + sidecar + template-by-effective-date + `Verify` re-hash; immutability grant-enforced and tested. Deviation: content stored in-row (bytea), not object storage — documented (D-0076), but large artifacts will bloat the DB |
| E5 scheduler | **PARTIAL** | Interval-based, leader-safe, lag-hooked — kernel sweeps only. **Modules cannot register recurring jobs** (needs platform pool; no Context path); no cron syntax |
| E6 bulk framework | **FULL** | Chunked, per-item tx with done-mark committed atomically with the work, failure ledger, resumable, progress; all proven. Caveats: single processor per operation (no SKIP LOCKED on item claim — documented follow-up); not integrated with jobs/Context (hand-wire a driver) |

## 3. Cross-cutting root causes (fix these and most rows close)

1. **The metrics port is uniformly unimplemented.** `observability.Metrics` + Prometheus adapter exist, but there are **zero `IncCounter`/`SetGauge` emission sites in the repo**. This single root cause fails acceptance clauses in S2, R3, R4, R8, and O4.
2. **Default-wiring gap.** S1, S2, R1, R7, O1 (and E5's module face) are "kernel primitives nothing turns on": the scaffolded product gets a deny-all authenticator, no rate limiter, no authz cache, unsigned cursors (including in the kernel's own workflow endpoint), and a NoOp tracer with a fictional config knob. The roadmap's intent was hardening *products get by default*, not a parts shelf.
3. **No hosted CI.** Every "CI runs X" resolves to local `make ci` / `make ci-container`. Nothing enforces the gate on push. Related: without a DSN, several packages (e.g., webhook — including its pure unit tests) report `ok` with 100% skips; `WOWAPI_REQUIRE_DB=1` closes this only inside the container target.
4. **Status-declassification drift.** `docs/implementation/hardening-plan.md` downgraded R2 and O3 to "doc-only" — openly, with rationale (fact-checked: the plan discloses it in a dedicated section) — but ROADMAP-wowapi.md still listed them with full P0 acceptance, and the "all roadmap gaps closed" headline (commit 865b14a) plus several claims ("audited service principals", "rotatable", "end-to-end OTel tracing", ABAC-deny "tested") overstate what landed. S4's reclassification is *substantively correct*; R2's is not — the acceptance was a load characterization, and none exists.
5. **Perf gates not extended.** `bench-budgets.txt` untouched since Phase 11 — no budgets for the new hot paths (audit Record/chainHash, sequence Allocate, token bucket, CachingStore, edge middleware).

### 3.1 Hardening-method (§5 of ROADMAP-wowapi) verdicts

The roadmap's method section is itself part of the acceptance and the exit gate's second clause ("pilot module green under load/chaos/adversarial suites"):

| Method step | Verdict | Evidence |
|---|---|---|
| 1. Phase-0 pilot module | **NOT DONE** | No pilot module exists in wowapi or any product repo; correctly deferred to PRODUCT-PLAN E0-S2, but the exit gate cannot be met without it |
| 2. Load & soak | **NOT DONE** | No soak/load harness (see R2 MISSING); only micro-benchmarks exist |
| 3. Chaos (worker kill, hold-vs-sweep race) | **PARTIAL** | Hold-vs-sweep race is tested (R6 FULL, runs under `-race`); worker-kill / relay-kill chaos does not exist |
| 4. Adversarial (fuzz, cross-tenant, escalation) | **PARTIAL** | Fuzz targets real and run clean; RLS-isolation + escalation suites real; authz property tests missing (S8) |
| 5. Operational drills | **PARTIAL** | Migration-reversibility and backup drills exist and were executed; PITR/object-storage legs and upgrade drill missing (O5/O3) |
| 6. Upstream loop | **DONE (in-repo)** | Review-findings commits (73d221f, d22ff7f, 54abec1) show a working find→fix loop with decision records |

Net: the exit gate fails on both clauses — open P0 rows (§2) *and* the pilot/load/chaos legs of the method.

## 4. Corrective actions (feed Epic 0 / E0-S3 of PRODUCT-PLAN.md)

**P0 (blocks the ROADMAP-wowapi exit gate):**
- **CA-1 Metrics emission pass:** emit rate-limit drops, scheduler lag, DLQ depth, webhook breaker state, config-fingerprint info metric; ship a real Prometheus exposition for the fingerprint; make the reference alert rule runnable; ship it in `deployments/reference/`.
- **CA-2 Default wiring pass:** composite authenticator (OIDC + API key) in the generated api; `RateLimit` in the default chain (opt-out); wire `CachingStore` behind a kernel option **with an invalidation hook on seed/spine writes** (or explicitly rescope to TTL-only and update the roadmap); real `TraceSampleRatio` config key wired to the OTel adapter as the documented opt-in; fix `workflow.OpenTasksFor` to signed cursors; generated CRUD list template uses `filtering.NextCursor`.
- **CA-3 apikey completion:** `Rotate` API, audited issue/rotate/revoke via `kernel/audit`, `wowapi apikey` CLI, and the missing ABAC-deny-over-scope test.
- **CA-4 R2 for real:** load characterization of per-aggregate advisory-lock ordering; publish the throughput envelope + sub-sharding guidance (can land via the product's E0-S2 pilot, results upstreamed).
- **CA-5 E5 module face:** recurring-job registration from `module.Context` (platform-pool mediated at boot); cron syntax can follow.
- **CA-6 Hosted CI** running `ci-container` (incl. fuzz seeds + reference-stack smoke), and un-gate pure unit tests from the DB skip path.
- **CA-7 Traceability reconciliation:** update ROADMAP-wowapi.md statuses to match reality (S4 reclassified with residuals; R2/O3 reopened), correct the overstated CHANGELOG/decision claims, refresh bench-budgets, move `goal-test.md` under `docs/qa/`.

**P1:**
- **CA-8 S5 defined error** on replay-after-expiry (new error kind) — or a documented, roadmap-amended acceptance reversal.
- **CA-9 O1 async propagation:** trace context through outbox → relay → jobs → notify (envelope/metadata column), worker spans.
- **CA-10 Read-replica routing** for `WithTenantRO` (RO DSN config) — or explicit rescope with the deployment-concern rationale recorded in the roadmap.
- **CA-11 Audit completeness:** `tx_id` column; Context accessors for audit/sequence/artifact/bulk; scheduled anchor-export + `wowapi audit verify` CLI.
- **CA-12 O2/O3/O5 finishers:** schema-snapshot diffing in the reversibility drill; published deprecation/upgrade policy doc; scripted PITR + object-storage restore legs (or rescope to staging-drill with the doc updated).
- **CA-13 S3 finishers:** gate-level test for 401/`WWW-Authenticate`; document API-key×step-up interplay; challenge-interface decision recorded.
- **CA-14 S4 residuals:** `integration.Config.Credential` → `config.Secret` type; rotation procedure documented.
- **CA-15 R5/R6 nits:** unregistered notification channel should error, not noop-`sent`; document that generic retention disposition relies on product DisposeFunc for hold-skipping (or lift the document-class guard into the engine).

## 5. Test, documentation & traceability assessment

- **Tests:** genuinely strong where features exist — grant-level enforcement proven by attempted violation, concurrency proven by real races, drills executed. Gaps are *absence* (property tests, gate step-up test, ABAC-vs-scope test, load tests), not weakness.
- **Docs:** blueprint/ops docs mostly honest at decision level (deviations recorded in D-0061..D-0077); two references to nonexistent knobs/metrics (`cfg.TraceSampleRatio`, `wowapi_config_fingerprint_info`) must be fixed or implemented.
- **Traceability:** hardening-plan.md + decisions D-0061..D-0077 map roadmap→commit→evidence well; the drift is at the *status* level (see CA-7). Scope check: clean — all 136 changed files trace to the roadmap, review findings, or QA gap closure.

**Conclusion:** Do not declare the hardening roadmap closed. Execute CA-1…CA-7 (P0) — several are small relative to what's already built — then re-verify. The foundation quality is high; the distance to "closed" is mostly wiring, metrics, and honesty-of-status, not architecture.

---

## 6. Corrective-action closure status (remediation pass)

Verified against the tree after the remediation pass; each CLOSED item ships code + tests behind the
`make ci` gate (DB tests forced, `WOWAPI_REQUIRE_DB=1`; full suite green, 45 pkgs; bench budgets green).

| CA | Pri | Status | Evidence |
|---|---|---|---|
| CA-1 metrics emission | P0 | **CLOSED** | `observability.Metrics` wired via `kernel.Deps.Metrics`; RED middleware fed the live sink in the api template + `/metrics`; `wowapi_config_fingerprint_info`, `scheduler_lag_seconds`, `scheduler_task_errors_total`, `webhook_breaker_state`, `rate_limit_dropped_total` emitted; worker `/metrics` listener; `deployments/reference/prometheus-{alerts,scrape}.yml`. Tests: webhook breaker, ratelimit drop. **Residual:** DLQ-depth gauge (scheduler infra in place). |
| CA-2 default wiring | P0 | **CLOSED** | RateLimit in default chain (opt-out `http.rate_limit`); real `telemetry.trace_sample_ratio` wired to the OTel adapter (fictional-knob finding fixed); `httpx.Composite` (API key + OIDC) in the generated api; authz `CachingStore` opt-in kernel option (`AuthzCacheTTL`, off by default, `Invalidate` handle exposed); `workflow.OpenTasksFor` signed/versioned cursors; generated CRUD list uses `filtering.NextCursor`. Tests for each; scaffold builds. |
| CA-3 apikey completion | P0 | **CLOSED** | `Store.Rotate` (two-call overlap); audited issue/rotate/revoke via `kernel/audit`; `wowapi apikey issue\|list\|rotate\|revoke` CLI (verified end-to-end vs live DB); ABAC-deny-over-scope test. |
| CA-4 R2 load | P0 | **CLOSED** | `TestIntegrationOutboxHotAggregateThroughput` measures ~200 events/sec single hot aggregate; `docs/operations/load-characterization.md` documents the envelope + sub-sharding strategy. |
| CA-5 E5 module face | P0 | **CLOSED** | `module.Context.RecurringJob`; collected on `Booted.Recurring`; run per-tenant leader-safe by the worker scheduler. Test. |
| CA-6 hosted CI | P0 | **PARTIAL** | `.github/workflows/ci.yml` (no-DB unit job + authoritative `ci-container` gate + fuzz seeds), YAML-valid. **Cannot be proven green without a push** (not performed). Reference-stack app-smoke deferred (needs a scaffolded running product; header posture unit-tested in `edge_test.go`). |
| CA-7 traceability | P0 | **PARTIAL** | `hardening-plan.md` STATUS made honest; stale counts fixed; this closure matrix added. **Remaining:** ROADMAP-wowapi row-level status edits, CHANGELOG overclaim edits, `progress.md` hardening rows. |
| CA-8 S5 defined error | P1 | **CLOSED** | `KindIdempotencyExpired` (410); expired-present keys error instead of silently re-executing. Test. |
| CA-9 O1 async propagation | P1 | **OPEN** | Tracing is now real per-request (CA-2b), but trace-context Inject/Extract across outbox→relay→jobs→notify + worker spans is **not** implemented. Impact: traces don't span async boundaries. Risk: low (observability completeness, not correctness). Planned: add a traceparent envelope column + Inject on write / Extract on claim. |
| CA-10 read-replica routing | P1 | **RESCOPED (deployment concern)** | `WithTenantRO` already runs read-only transactions; routing them to a physical replica is a pool-wiring choice a product makes at deploy time (a replica DSN + a read-only pool), not a kernel behavior. A dedicated in-kernel RO-pool router is deferred with this rationale recorded rather than shipped speculatively. |
| CA-11 audit completeness | P1 | **PARTIAL** | **CLOSED:** `module.Context` accessors for `Audit`/`Sequence`/`Bulk`/`Artifacts` (primitives now reachable). **Remaining:** `tx_id` column on `audit_logs`; scheduled anchor-export + `wowapi audit verify` CLI. Impact: forensic tx-correlation + offline tamper-verification. Risk: low (chain verification exists in-process). |
| CA-12 O2/O3/O5 finishers | P1 | **OPEN** | Schema-snapshot diffing in the reversibility drill, a published deprecation/upgrade-policy doc, and scripted PITR/object-storage restore legs remain. Impact: operational polish. Risk: low. Planned: docs + a snapshot-diff assertion + a PITR drill script (or rescope O5 to staging-drill with the doc updated). |
| CA-13 S3 finishers | P1 | **CLOSED (test)** | Gate-level 401/`WWW-Authenticate` step-up test added (`TestIntegrationAuthzGateStepUpChallenge`). API-key×step-up interplay: an `ActorSystem` carries no AMR so it can never satisfy step-up — fail-secure by construction (documented here). |
| CA-14 S4 residuals | P1 | **CLOSED (code)** | `integration.Config.Credential` is now `config.Secret` (structural redaction; leak-in-format test). Rotation-procedure doc folds into the ops docs pass. |
| CA-15 R5/R6 nits | P1 | **CLOSED** | Unregistered notification channel now fails terminally (`dead` + error) instead of noop-`sent`; `noopSender` removed. Test. |

**Exit gate:** P0 corrective actions CA-1…CA-5 are CLOSED and verified; CA-6 is written but unproven-without-push; CA-7 is partial. So the exit gate is **substantially met on P0 code** but **not fully closed**: CA-6 needs a green hosted run and CA-7 needs the remaining status/CHANGELOG edits. P1: CA-8/13/14/15 closed, CA-11 partial, CA-9/10/12 open/rescoped (all low-risk, documented above).
