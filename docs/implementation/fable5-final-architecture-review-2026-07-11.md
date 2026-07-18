# Fable 5 — Final Architecture Review & Production-Readiness Programme

- **Authority:** Fable 5 (senior architecture reviewer, delivery lead, final quality gatekeeper).
- **Mandate:** `WOW-Review.md` (1748 lines; mandate discharged — archived 2026-07-11 to the `wowapi2` documentation archive, `archive/prompts-and-mandates/WOW-Review.md`). This document satisfies its deliverables §28.A–V and the 22 required answers §29.
- **Reviewed revision:** `HEAD = 345e4ce` (three review commits: `e6bed50` artifacts, `f71f308` first attempt, `345e4ce` second attempt).
- **Source of truth:** the original directive `docs/implementation/architecture-directive-2026-07-11.md` + evidence bundle, and the *actual repository state* — not the lower-cost model's narrative.
- **Repository safety:** this review made **no commits, no history rewrites, no force-push, no repo-settings changes**. It is an uncommitted working-tree artifact. Every "remove/revert" recommendation below is expressed as a *future corrective task*, never performed here (per WOW-Review §3).
- **Method & cost model:** two lower-cost Sonnet agents ran non-overlapping mechanical work (claim reproduction + test execution; dependency/reuse/wowsociety inventory), max two concurrent, no premium parallelism (WOW-Review §6). Fable 5 personally performed all architecture judgement, layer/boundary decisions, blocker adjudication, dependency approval, question resolution, and final sign-off. Claims were reproduced by **two independent passes that agree**; where a fact rests only on agent report and was not independently re-run by Fable 5, it is labelled **[agent-reported]**.

---

## A. Executive architecture assessment

**Current condition.** `wowapi` is a **strong, unusually complete pre-production framework foundation** with a genuinely differentiated security/tenancy core (forced RLS, fail-closed tenant transactions, deny-by-default authz, SSRF-guarded outbound HTTP, signed cursors, hash-chained audit). The three review commits are **real, verified, non-cosmetic work**: 8 minimal-scope finding-slices landed with reproducible before/after tests, and the whole tree builds/vets/tests green at HEAD. This is materially better than most frameworks at the same stage.

**It is not yet production-ready as a premier framework.** The verdict is **NOT READY (conditionally advanceable)**. The reasons are architectural, not cosmetic:

1. **Kernel bloat is the dominant architectural defect.** `kernel/` contains **39 sub-packages** — including `kernel/webhook`, `kernel/notify`, `kernel/document`, `kernel/artifact`, `kernel/attachment`, `kernel/comment`, `kernel/bulk`, `kernel/integration`, `kernel/mfa`, `kernel/storage`. A kernel that must "remain small and stable" (WOW-Review §18) cannot own a webhook delivery engine, an object-document service, and a comment subsystem. These ~10 are **application-foundation or adapter concerns wearing a kernel import path**. This is the single largest gap between the current shape and the mandated four-level architecture, and it is *not* on the original 38-finding list — the directive's AR-01..06 treat the registration surface but never name the layering violation directly.
2. **The mutable-registration / no-application-model defect (AR-01/AR-02) is still fully open** and is the correctly-identified #1 architectural risk. The 8 landed fixes are all Wave-0 correctness patches; none touch the foundational compile-seal-immutable-model work the framework actually needs.
3. **Two P0 security/identity findings (SEC-01 server-side session resolution, DATA-08 W6 full audit integrity) are open and carry real `wowsociety` breaking impact.**
4. **Public contracts leak vendor types** (pgx, jwt, — see §17/§J). Deliberate in places, but unmanaged.

**Highest risks (ranked):** (1) kernel bloat locks in the wrong layering before v1 stabilises the surface; (2) AR-01/02 mutable registration invalidates the ownership model the whole security story assumes; (3) SEC-01 trusts JWT-carried tenant/impersonation state; (4) no production seed-sync path (`wowsociety` PF-9, prod-blocking, **not in the 38 findings** — a genuine miss); (5) DATA-01 tenant-FK integrity gap reproduced in `wowsociety`'s own schema.

**Main preliminary-review defects:** (a) the plan never surfaced the kernel-layering violation as a finding; (b) it never surfaced the `wowsociety` `docs/upstream/` register (13 real framework findings, 2 prod-blocking) as inputs; (c) a **traceability defect** — §6 marks DX-05 `PLANNED` while §9 reports DX-05 T1/T2 executed; (d) it accepted "reference-performance-environment has no owner" as a hard blocker for *all* PERF work when most PERF verification can proceed with relative/container comparison.

**Is `wowapi` production-ready?** No. It is a strong foundation that is **~6–9 focused work-waves** from a defensible premier-framework claim, with a clear, sequenceable path. It should **not** carry a new "production-grade" or "premier" claim until the P0 set and the layering correction land. Internal/pilot development may continue.

---

## B. Three-commit and repository-state inventory

**Working tree at review time:** clean except `M WOW-Review.md` (the mandate itself, edited) and this new review file. No stray staged changes. Untracked/gitignored top-level dirs `bin/`, `visuals/`, `reviewer content/` are all correctly `.gitignore`d (`.gitignore:35,56,79`) — not part of the repository, not a defect.

| Commit | Message | Files | Purpose (verified) | Disposition |
|---|---|---|---|---|
| `e6bed50` | Review artifacts | 3 (directive 1103 L, command-log 161 L, evidence.json 198 L) | The original architecture-review inputs, committed for traceability. Content matches what the plan cites. | **Retain.** Legitimate durable evidence. |
| `f71f308` | Review work first attempt | 19 files (+3360/−70) | `WOW-Review.md` + the plan doc (853 L) + 4 Wave-0 code fixes (SEC-02, PERF-01, PERF-06, DATA-08 slice) + tests + `bench-budgets.txt`. | **Retain with correction.** Code fixes verified real (§D). `WOW-Review.md` landing in this commit is unusual (a mandate doc committed as "review work") but harmless. |
| `345e4ce` | Review work second attempt (HEAD) | 14 files (+304/−90) | AR-04/AR-05/AR-06 + REL-04 code/doc fixes + plan-doc §9/§10 update. | **Retain with correction.** Code fixes verified real (§D). Introduces the §6-vs-§9 DX-05 traceability inconsistency (§E). |

**Cumulative diff `e6bed50^..345e4ce`:** 34 files, +5116/−150. No files were *lost* or *reverted* between attempts; `345e4ce` is purely additive over `f71f308` except for the plan-doc §8 re-ordering (verified: §-heading order is identical and correct in both committed versions — the section-ordering bug I introduced mid-session in the *prior* goal was fixed before commit; the committed docs are clean). The `testkit/workflowsim_cov_test.go` one-line change in `f71f308` is the SEC-02 collateral fix (the missed call site caught by that pass's own review gate) — **confirmed present and correct at HEAD**.

**Relationship between attempts:** first attempt = 4 Wave-0 slices + plan; second = 3 more slices (AR-04/05/06) + REL-04 + plan update. No superseded or contradicted work between them. Clean incremental history.

---

## C. Changed-file disposition — full per-file table (all 34 cumulative files)

Per WOW-Review §8. Status key: **R** retain-verified-correct, **RC** retain-with-correction, **RM** retain-consider-re-home. No file warrants removal. Origin = originating finding/purpose. Verification = how confirmed at HEAD.

| # | Path | Commit | Status | Purpose / origin | Quality / issues | Verification | Task |
|---|---|---|---|---|---|---|---|
| 1 | docs/implementation/architecture-directive-2026-07-11.md | e6bed50 | R | Original directive (38 findings) | Authoritative input, matches plan citations | Read; content = source of truth | — |
| 2 | .../command-log.md | e6bed50 | R | Review command log | Consistent w/ reproduced results | Cross-checked E-08/E-09 vs live S3 | — |
| 3 | .../evidence.json | e6bed50 | R | Review evidence | Well-formed | Parsed | — |
| 4 | WOW-Review.md | f71f308 | RM | This review's governing mandate | Legitimate; unusual location (repo root) | Read in full | T-DOC-02 |
| 5 | docs/implementation/premier-framework-implementation-plan.md | f71f308,345e4ce | RC | The 38-finding plan | **DX-05 §6/§9 status inconsistency** | §E spot-check confirmed | **T-DOC-01** |
| 6 | bench-budgets.txt | f71f308 | R | PERF-01 new bench budgets | 2 new entries, exit-0 | `make bench-budget` = 43 OK | — |
| 7 | internal/tools/benchbudget/main.go | f71f308 | R | PERF-06: missing-bench fails CI | Correct violations-append | Reproduced ×2, revert-proof | — |
| 8 | internal/tools/benchbudget/coverage_test.go | f71f308 | R | PERF-06 test | Subprocess exit-1 assert | 12/12 pass | — |
| 9 | kernel/httpx/ratelimit.go | f71f308 | R | PERF-01 sweep+hardcap | Backward-compat preserved | Reproduced ×2 -race | — |
| 10 | kernel/httpx/ratelimit_test.go | f71f308 | R | PERF-01 tests | 10k eviction + race | pass | — |
| 11 | kernel/httpx/bench_test.go | f71f308 | R | PERF-01 sweep benches | budgeted | exit-0 | — |
| 12 | kernel/httpx/export_test.go | f71f308 | R | PERF-01 test shim | Test-only export | compiles | — |
| 13 | kernel/kernel.go | f71f308,345e4ce | R | DATA-08 wiring (f7) + AR-06 closure (34) | Non-overlapping edits, correct | Reproduced ×2 | — |
| 14 | kernel/attachment/attachment.go | f71f308 | R | DATA-08: outbox err propagated | `_ =` discard gone | Fault-injection rollback test | — |
| 15 | kernel/attachment/coverage_test.go | f71f308 | R | DATA-08 test | Real rollback proof | DB test pass | — |
| 16 | kernel/notify/service.go | f71f308 | R | DATA-08: legal-delivery audit | events_outbox in-tx, correct choice | pos+neg test | — |
| 17 | kernel/notify/notify_test.go | f71f308 | R | DATA-08 test | conditional gate proven | pass | — |
| 18 | kernel/workflow/runtime.go | f71f308 | R | SEC-02: mandatory evaluator | nil-guard+unconditional check | Reproduced ×2 -race | — |
| 19 | kernel/workflow/runtime_extra_test.go | f71f308 | R | SEC-02 adversarial | fail-closed proven | pass | — |
| 20 | kernel/workflow/runtime_lifecycle_test.go | f71f308 | R | SEC-02 fixture fix | consistent | pass | — |
| 21 | kernel/workflow/runtime_test.go | f71f308 | R | SEC-02 fixture fix | consistent | pass | — |
| 22 | testkit/workflowsim_cov_test.go | f71f308 | R | SEC-02 collateral call-site fix | The caught missed site | pass | — |
| 23 | .github/workflows/ci.yml | 345e4ce | R | REL-04: S3 env in gate | inherits via ci-container | S3 execute-not-skip | — |
| 24 | Makefile | 345e4ce | R | REL-04: WOWAPI_REQUIRE_S3+endpoint | correct defaults | 20 S3 tests pass | — |
| 25 | deployments/compose.yaml | 345e4ce | R | REL-04: minio healthcheck + canonical var | correct | `docker compose config` verified | — |
| 26 | docs/user-guide/build-deploy.md | 345e4ce | R | REL-04 stale S3_ENDPOINT ref fix | caught by that batch's gate | read | — |
| 27 | app/boot.go | 345e4ce | R | AR-04: unknown-namespace reject | deterministic named err | Reproduced ×2, full suite green | — |
| 28 | app/boot_extra_test.go | 345e4ce | R | AR-04 neg fixture + pos repair | correct | pass | — |
| 29 | README.md | 345e4ce | R | AR-05/DX-05 doc drift | verified vs source | grep confirms no RunAPI etc. | — |
| 30 | docs/blueprint/06-module-sdk.md | 345e4ce | R | AR-05: Context interface | matches 39-method live iface | diffed | — |
| 31 | docs/blueprint/11-...consumption.md | 345e4ce | R | AR-05/DX-05: composition + CLI | phantom APIs removed | read | — |
| 32 | docs/operations/upgrade-and-deprecation-policy.md | 345e4ce | R | DX-05: v1 policy | matches CHANGELOG/tags | read | — |
| 33 | kernel/authz/caching_internal_test.go | 345e4ce | R | AR-06 sentinel test | proves instance routing | -race pass | — |
| 34 | kernel/kernel_rules_test.go | 345e4ce | R | AR-06 cache-on integration | correct | DB test pass | — |

**Summary: 30 R, 3 RC (all the plan doc, one row), 1 RM (WOW-Review.md). Zero removals.** This review file itself is a 35th artifact (RM — consider `docs/implementation/` placement, done: it lives there).

---

## D. Implementation & claim-verification review (the 8 "executed" findings)

Reproduced by two independent agent passes **that agree on every item**, cross-checked against my own direct spot-checks (build/vet, section order, S3 test, audit-metadata type). Verdict per finding — **all 8 are genuinely, correctly implemented at the minimal scope claimed; none is fully closed per the directive's §13.2 bar, which the plan itself states honestly.**

| Finding | Claim | Verification verdict | Evidence |
|---|---|---|---|
| **SEC-02** | Workflow `Override` fails closed; nil evaluator rejected | **VERIFIED.** `NewRuntime` panics on nil `ev`; `Override` check unconditional. Collateral `testkit` call-site fix present. `go test ./kernel/workflow/... -race` green. | Reproduced ×2; the missed sibling call site was caught by that pass's own review gate — evidence the review discipline works. |
| **PERF-01** | Token-bucket sweep recomputes refill; hard cap added; backward-compatible | **VERIFIED.** Sweep no longer compares stale token count; `WithHardCap` added; 2-arg constructor unchanged. 10k-key eviction + race tests pass. | Reproduced ×2. New benches in `bench-budgets.txt`, exit 0. |
| **PERF-06** | Missing budgeted benchmark now fails CI | **VERIFIED.** `violations` append replaces WARN+continue; subprocess exit-1 test. `make bench-budget` exit 0 / 43 OK. | Reproduced ×2 incl. revert-proof. |
| **DATA-08 (W0 slice)** | Attachment outbox error propagated (rollback); legal-delivery audit via `events_outbox` | **VERIFIED.** `_ =` discard gone; legal-delivery write in-tx; migration 00011 grant confirmed. Fault-injection rollback test passes. **W6 (audit-hash widening) correctly NOT done.** | Reproduced ×2. Design choice (events_outbox vs audit_logs) is sound — matches migration intent. |
| **AR-04 (T1)** | Boot rejects unknown module config namespaces | **VERIFIED.** Deterministic named error; negative fixture; positive test repaired. Full `go test ./...` green (both passes). T2–T5 correctly deferred (depend on AR-01). | Reproduced ×2. `wowsociety` no-op confirmed (no `modules:` key). |
| **AR-05 (T1/T2)** | README/blueprint composition-root + Context-interface drift fixed | **VERIFIED.** `RunAPI/RunWorker/RunMigrate` confirmed absent; `Context` doc matches live 39-method interface. | Reproduced ×2 against source. |
| **AR-06 (T1)** | `orgAncestry` closure uses composed `authzStore` | **VERIFIED.** Sentinel-injection test proves instance routing. Cache-enabled integration test added. | Reproduced ×2, `-race`. |
| **REL-04 (T1–T4)** | S3 test wiring by default; TOTP determinism audited | **VERIFIED.** `Makefile ci-container` sets `WOWAPI_REQUIRE_S3`+`S3_TEST_ENDPOINT`; 20 S3 tests execute-not-skip on compose defaults; TOTP deterministic across TZ (3 `time.Now()` sites are pre-timestamp error paths — genuinely unreachable, audit outcome correct). | Reproduced ×2. |

**CI/race/benchmark/S3/TOTP claims:** all **VERIFIED**. The only cross-run divergence was a single `internal/e2e` failure in one full-suite run that passed in the other and in 4/4 isolated runs — an intermittent full-suite-only failure, not a HEAD defect. *(Correction, closure-depth pass: this section originally attributed it to "a shared-DB concurrency flake" — that cause was asserted without checking testkit, which already provides per-test template-clone DB isolation (`testkit/db.go:83-144`). The observed fact stands; the cause is undiagnosed. See the re-scoped **T-TEST-01** in §O/CS-13: reproduce first, then diagnose.)*

**Independent-review claims:** the plan reports each batch passed an independent gate; the first batch's gate caught one real Critical (the `testkit` call site) and fixed it. I confirm that fix is at HEAD. The review discipline is real, not theatre.

---

## E. Preliminary-plan quality review

**What was correct and strong:** the 38-finding decomposition is faithful to the directive; the wowsociety-impact analysis is genuinely grounded (it found the real `policy_override` DATA-01 instance and the live-audit-row DATA-08 breaking impact by reading wowsociety source, not guessing); the honesty discipline is exemplary (explicit "PLANNED only", "not closed per §13.2", disclosed procedural shortcuts). The 8 executed slices were correctly chosen as the smallest genuinely-independent, non-blocked pieces.

**What was weak / wrong / missing (Fable 5 corrections):**

| Plan aspect | Assessment | Correction |
|---|---|---|
| Kernel-layering violation | **Entirely missed.** 39 kernel sub-packages (40 incl. the `kernel` root) incl. webhook/notify/document/comment/bulk — never flagged. | New finding **FBL-01** (P1). §J maps the re-homing. |
| `wowsociety/docs/upstream/` register | **Missed as input.** 13 real framework findings (2 prod-blocking: PF-9 no prod seed-sync, RFF-001 prod storage — RFF-001 since resolved) never folded into the backlog. | New findings **FBL-02** (PF-9, P0-prod), **FBL-03** (upstream-register reconciliation). |
| DX-05 status | **Traceability defect.** §6 = PLANNED, §9 = EXECUTED T1/T2. | **T-DOC-01**: reconcile; DX-05 T1/T2 *did* ship — the matrix is wrong, not the prose. |
| Reference-perf-env blocker | **Over-stated.** Treated as blocking all PERF-02/03/04/05. | §12 constrains it: relative/container benchmarking proceeds now; only absolute-SLO gating waits on the env. |
| Retry/backoff duplication | **Missed reuse opportunity.** `cenkalti/backoff/v5` + `sethvargo/go-retry` are already in the module graph, unused, while retry is hand-rolled twice. | Reuse register §K → **FBL-04**. |
| "10 questions require human" | **Over-transferred.** Several are answerable now by architectural judgement. | §F reduces the genuine human set to **3**. |

**Disposition:** accept the plan as a strong first pass; **strengthen** it with FBL-01..04 + T-DOC/T-TEST tasks; **reject nothing** outright; the plan is retained as the backlog spine with the corrections in §O layered on.

---

## F. Resolution of the 10 unresolved questions — reduced to 3 genuine human decisions

The plan's §7 listed 7 "genuinely undecided" + 3 "human/organisational" items. Fable 5 adjudication: **7 of 10 are resolvable now** by architectural judgement, evidence, or testing; **3 remain genuine human decisions** (and even those have safe defaults so implementation is not blocked).

| # | Question (restated) | Class | Fable 5 decision + safe default | Blocks impl? |
|---|---|---|---|---|
| 1 | SEC-01: will the IdP mint an opaque `grant_id`; who approves break-glass? | **Genuine human (product/security-lead)** | *Safe default that unblocks:* build the server-side `identity_grant` table + resolver **now**, keyed on grant-ID, and have `Verifier.Actor` consult it. If the IdP cannot emit `grant_id`, the framework still owns the grant record and looks it up by session — the JWT only carries a stable subject. Implementation proceeds against the safe default; the human decision only tunes claim shape. | **No** |
| 2 | wowsociety `identity_impersonation_session` vs framework grant table authority | **Fable 5 decision (framework boundary)** — resolved | **Framework owns grant validity/expiry/revocation; wowsociety keeps its table for product UX/audit only.** This is the correct dependency direction (WOW-Review §1). Recorded as decision **D-01**. | No |
| 3 | AR-01 `Registrar` type: one shared vs per-subsystem | **Fable 5 decision (public contract)** — resolved | **One generic owner-bound `Registrar` capability type**, with per-subsystem *typed keys* (`Key[T]`) rather than per-subsystem registrar types. Capability confusion is prevented by the key's phantom type + owner binding, not by multiplying registrar types. Decision **D-02**. | No |
| 4 | AR-01/AR-04 post-seal mutation: error vs panic in prod | **Fable 5 decision (concurrency/lifecycle)** — resolved | **Error in production builds; panic only under an explicit `dev`/test build tag.** A framework must not convert a benign retained-handle into a prod crash. Decision **D-03**. | No |
| 5 | DATA-08 W6 audit-hash `hash_version` discriminator design | **Answerable by technical analysis** — resolved | **Add a `hash_version smallint NOT NULL DEFAULT 1` column in the same migration; verification branches on it.** Historical rows verify under v1; new rows under v2 (metadata + tx_id included). Standard append-only-log versioning. Decision **D-04**. | No |
| 6 | REL-01 GoReleaser split-mode (`--skip=publish` vs hand-rolled) | **Answerable by evidence/testing** — resolved | **Use GoReleaser `release --skip=publish` for build-candidate + a separate `goreleaser publish` step.** Supported in current GoReleaser; no hand-rolled pipeline needed. Decision **D-05** (verify against pinned GoReleaser version at implementation time). | No |
| 7 | SEC-04 cross-pod cache invalidation transport (LISTEN/NOTIFY vs epoch poll) | **Fable 5 decision (concurrency)** — resolved | **Per-tenant epoch integer in a small `authz_epoch` table, polled on the existing authz read path; Postgres `LISTEN/NOTIFY` as an optional latency optimisation, not a correctness dependency.** Avoids a new message bus in the kernel. Decision **D-06**. | No — but this whole finding is P1, not on the critical path. |
| 8 | SEC-06 JWKS-client governance model | **Fable 5 decision (security)** — resolved | **Require trusted-issuer/egress config to be a declared, fingerprinted `config` field; reject a custom JWKS `*http.Client` in `prod` profile unless the trusted-issuer allowlist is set.** Decision **D-07**. | No |
| 9 | Reference-performance-environment ownership | **Infrastructure decision w/ provisional default** | *Provisional:* a **Linux amd64 GitHub Actions runner + committed `perf/reference-schema1.json`** baseline, advisory-only initially; a dedicated bare-metal runner is a *later* SRE decision, not a blocker. See §12. | **No** (relative benchmarking proceeds now) |
| 10 | GitHub org-admin actions (branch/tag/env protection) for REL-01/REL-02 | **Genuine repo-administration** | Only the *final activation* needs admin. All workflow YAML, gate manifest, and verification script are authorable + testable now against a scratch repo. See §G. | **No** for implementation; **Yes** for rollout enforcement only. |

**Net: the genuine human-decision set is 3** — Q1 (IdP claim contract, with a safe default that unblocks build), Q9 (perf-env ownership, provisional default set), Q10 (repo-admin rollout activation). None blocks *implementation*; each blocks only a specific *final* activation or tuning.

## G. Blocker-resolution plan — implementation vs validation vs enforcement vs rollout

**GitHub repo-administration "blocker" (REL-01/REL-02):** decompose per WOW-Review §11 — most is not blocked.

| Layer | Blocked? | Who | Interim default |
|---|---|---|---|
| Workflow authoring (`required-gates.yml`, gate manifest, `verify_release.sh` + golden-failure tests) | **No** — author + unit-test now | any-tier agent | build against a scratch/throwaway repo |
| Local validation (exact-SHA gate logic, tamper fixtures) | **No** | agent | dry-run locally |
| Blocking-scanner rollout (Trivy `exit-code:1`, waiver schema) | **No** — code now | agent | report-only baseline first, then flip |
| Branch protection on `main` | **Enforcement only** | repo admin (human) | CI is advisory until set |
| Protected `release` GitHub Environment | **Rollout only** | repo admin (human) | publish job runs unprotected in scratch until set |
| Tag protection ruleset | **Rollout only** | repo admin (human) | — |

**Verdict:** REL-01/REL-02 are **~85% implementable and fully testable now**; only the last-mile *activation* of branch/env/tag protection is genuinely human. Do not classify the whole workstream blocked.

**Reference-performance-environment "blocker" (PERF-02/03/04/05):** see §12 — constrained; only absolute-SLO gating waits.

## H. Complete capability matrix — 30 areas (condensed; classification per §15 A–H)

| # | Area | Class | Location | Note |
|---|---|---|---|---|
| 1 | Application structure | **B** (incomplete) | app/, module/, kernel/ | Kernel-bloat layering violation (FBL-01) |
| 2 | DI / IoC | **C** (partial) | manual constructor wiring | No compiled provider graph (AR-02 open); manual wiring is defensible but unenforced |
| 3 | Configuration | **A-** | kernel/config | Strong (fail-closed, provenance); ~4,300 LOC custom — reuse review (K) says *retain*, justified |
| 4 | Lifecycle | **B** | app/boot.go, kernel/lifecycle | Hand-maintained manifest (AR-02); graceful shutdown present |
| 5 | Error handling | **A-** | kernel/errors (kerr) | Structured, typed Kinds; solid |
| 6 | Logging | **A-** *(corrected)* | kernel/logging (slog) | stdlib slog + redaction — correct library choice, no reinvention. **But "no reinvention" ≠ industry-best as-shipped:** zero trace/span correlation (no ctx-aware handler) — a log line cannot be joined to its request trace (FBL-06/CS-05). Redaction suffix-list is defense-in-depth only, correctly documented as such. |
| 7 | Observability | **B+** | kernel/observability + adapters | Ports clean, OTel/Prom in adapters; SLO kit absent |
| 8 | Security foundations | **B** | kernel/auth,authz,mfa,httpx | Strong core; **SEC-01 P0 open** (JWT-trusted session state) |
| 9 | Validation | **B** *(downgraded)* | kernel/validation | validator/v10 wrapper is correct reuse, **but enforcement is opt-in per handler** — `BindAndValidate` is a helper the router never requires; a forgotten call = an unvalidated write endpoint (FBL-08/CS-08) |
| 10 | Transport abstraction | **B** | kernel/httpx | HTTP-only; pgx leak in DBTX is the transport-boundary concern |
| 11 | Routing/middleware | **A-** | kernel/httpx (ServeMux) | stdlib 1.22 mux — correct; RouteMeta boot-validated |
| 12 | Data-access | **B** | kernel/database (pgx) | **pgx types leak into public `DBTX`/`Option`** (J) — deliberate but unmanaged |
| 13 | Transaction/consistency | **A-** | kernel/database txmanager | Fail-closed tenant tx, SET LOCAL — a genuine strength |
| 14 | Background processing | **B** | kernel/jobs,outbox | Custom DB-backed runner; **DATA-02/03 lease/fencing open** |
| 15 | Resilience | **C** | kernel/webhook/breaker, ratelimit | Custom breaker + rate-limiter; PERF-01 fixed; reuse review (K) flags candidates |
| 16 | Testing support | **A-** | testkit | Strong real-Postgres testkit; **e2e isolation flake (T-TEST-01)** |
| 17 | Developer tooling | **B** | internal/cli | `gen crud` emits invalid verb (DX-02 open, wowsociety PF-2) |
| 18 | API contract mgmt | **C** | internal/cli/openapi | Merge silently drops fields (DX-06 open) |
| 19 | Extensibility | **C** | module registration | **AR-01/02 mutable registration open** — the core gap |
| 20 | Modularity | **B** | module/, kernel/ | Ownership model not enforced (AR-01) |
| 21 | Compatibility/upgrade | **B** | docs + go.mod | v1 policy now corrected (DX-05); no automated API-diff gate (REL-03) |
| 22 | Performance controls | **B** | benchbudget, bench_test | PERF-06 fixed; no reference env (constrained, §12) |
| 23 | Caching | **C** | kernel/authz/caching | Unbounded map (SEC-04 open); reuse candidate golang-lru (K) |
| 24 | Events/messaging | **B** | kernel/outbox,integration | Transactional outbox present; relay tx-boundary issue (DATA-03) |
| 25 | Multi-tenancy | **A** | kernel/database, authz | Forced RLS — the flagship strength; **DATA-01 FK gap** (P0) is the one hole |
| 26 | i18n / time | **B** | kernel/i18n, model.Clock | Custom i18n ~1,500 LOC (reuse review: retain w/ caveat); no plurals by design |
| 27 | File/object storage | **A-** (F: adapter) | kernel/storage + adapters/s3 | Correct port+adapter (minio) — but `kernel/storage` port is fine, `kernel/document`/`artifact`/`attachment` are the bloat (FBL-01) |
| 28 | Auditability/compliance | **B** | kernel/audit, retention | Hash chain excludes metadata+tx_id (**DATA-08 W6 open, P0/P1**) |
| 29 | Deployment/ops | **B** | deployments/, cmd | Reference stack + smoke; readiness omits migration-currency (DX-07) |
| 30 | Documentation/governance | **B** | docs/ | Drift corrected (AR-05); no doc-example CI gate yet |

**No area is class-A-complete across the board; none is class-E-missing-entirely.** The framework is broadly **B (implemented-but-incomplete)** with A-grade tenancy/config/logging and C-grade extensibility/contract/caching.

## I. Mandatory-capability readiness — 20 capabilities

Ready = A. Conditionally ready = B/C with a clear gap. Not ready = D/E.

| # | Mandatory capability | Verdict | Gate |
|---|---|---|---|
| 1 | Clear architecture/module structure | **Conditionally ready** | FBL-01 kernel re-home |
| 2 | DI / decoupling | **Conditionally ready** | AR-02 typed provider graph |
| 3 | Typed config + validation | **Ready** | — |
| 4 | Structured errors | **Ready** | — |
| 5 | Structured logging | **Ready** | — |
| 6 | Metrics/tracing/health | **Conditionally ready** | health migration-currency (DX-07) |
| 7 | AuthN/AuthZ integration | **Conditionally ready** | **SEC-01 (P0)** |
| 8 | Input validation + secure defaults | **Conditionally ready** *(downgraded — CS-08)* | FBL-08: validation library present but enforcement discretionary; RouteMeta-seam boot check + binding adaptor specified |
| 9 | Graceful startup/shutdown | **Conditionally ready** | Narrowed: a real bounded drain exists (`app/worker.go:108-141`, 30s budget + `ReclaimStalled` recovery); the residual gap is only DATA-02's fencing of a drained-past worker's late finalize (CS-11) |
| 10 | DB + transaction integration | **Ready** (with pgx-leak caveat) | FBL-05/CS-10: keep the raw `pgx.Rows` contract (idiomatic, `database/sql`-shaped), enforce close/err-check mechanically via `sqlclosecheck`+`rowserrcheck` — a decided fix, not an open caveat |
| 11 | Timeouts/cancellation/resilience | **Conditionally ready** | breaker/limiter reuse (K); DATA-03 |
| 12 | Unit/integration/functional testing | **Ready** (with e2e flake) | T-TEST-01 |
| 13 | API contract + versioning | **Not ready** | DX-06 merge + REL-03 diff gate |
| 14 | Background jobs/messaging | **Conditionally ready** | **DATA-02/03 (P0)** |
| 15 | Extension/plugin mechanisms | **Not ready** | **AR-01/02 (P1, core)** |
| 16 | Static analysis/linting/tooling | **Conditionally ready** *(downgraded — see CS-23)* | FBL-05/FBL-07: lint stack runs at default tier (`.golangci.yml`: standard set + 4 extras); `sqlclosecheck`, `rowserrcheck`, `bodyclose`, `noctx`, `gosec`, `errorlint` etc. ship with the installed golangci-lint but are not enabled |
| 17 | Deployment readiness | **Conditionally ready** | seed-sync prod path (FBL-02/PF-9, P0) |
| 18 | Upgrade/compat/deprecation | **Conditionally ready** | REL-03 automated gates |
| 19 | Auditability | **Conditionally ready** | **DATA-08 W6 (P0/P1)** |
| 20 | Complete documentation | **Conditionally ready** | doc-example CI gate |

**Ready: 5. Conditionally ready: 13. Not ready: 2 (API-contract, extension-mechanism).** Zero are irredeemably missing. *(Rows 8 and 16 downgraded from Ready by the closure-depth pass — row 8 because validation is present-but-discretionary (a library's existence is not enforcement), row 16 because the lint stack runs at default tier while the same binary ships resource-leak/security analyzers unenabled. Both downgrades are the two-principles test — Reuse AND Utilisation — doing its job.)*

**Closure depth:** every §H and §I row now maps to a closure specification (diagnosis + implementable fix + fail-first verification + acceptance evidence) in `fable5-closure-depth-matrix-2026-07-11.md` — the §H/§I tables above remain the survey lens; the matrix is the execution lens. A Verdict/Class in these tables is no longer the terminal answer for any row.

## J. Four-level architecture map + the layering correction (FBL-01)

Target (WOW-Review §18) vs current placement:

- **Kernel (must stay small/stable):** *should* be `lifecycle, config, errors, model, secrets, module, context, database(tx-contract only)`. **Currently polluted** by ~10 packages that are not kernel concerns.
- **Application foundation:** validation, authz-policy integration, transaction boundaries, domain execution — partly in kernel, acceptable.
- **Infrastructure adapters:** `adapters/*` is clean (auth/pgprincipal, storage/s3, metrics/prometheus, tracing/otel, secrets/envprovider). **This layer is done correctly.**
- **Operational foundation:** testkit, logging, observability, benchbudget, deployment — present, reasonable.

**FBL-01 re-home plan (P1, a v2-module-path change — deliberate, phased):**

| Package | Current | Correct layer | Rationale |
|---|---|---|---|
| `kernel/webhook`, `kernel/notify` | kernel | **App-foundation service + adapter split** | Delivery engines w/ network I/O — not kernel |
| `kernel/document`, `kernel/artifact`, `kernel/attachment` | kernel | **App-foundation** (over `kernel/storage` port) | Domain-ish document services |
| `kernel/comment`, `kernel/bulk` | kernel | **App-foundation / optional module** | Feature subsystems |
| `kernel/integration` | kernel | **App-foundation** | Integration orchestration |
| `kernel/mfa` | kernel | **App-foundation security** (crypto primitives may stay low) | TOTP service, not kernel core |
| `kernel/storage` (port) | kernel | **Retain** — this is the correct adapter boundary | Port stays; impls in adapters/ |

This is the largest single architectural correction and **must precede v1 stabilisation** or the wrong surface locks in. It is a deliberate breaking change with a `wowsociety` migration (§P).

## K. Reuse-opportunity register (build-vs-reuse; §19 mandate)

Applying reuse-before-build with senior judgement — **not** automatic replacement (§19.6).

**Second principle added by the closure-depth pass — Utilisation:** §19 as originally applied
only asked *"did we avoid writing unnecessary custom code?"* It never asked *"are we fully using
the mature capabilities already integrated?"* That second test found real gaps the first one
structurally cannot: the lint binary ships resource-leak/security analyzers that the config never
enables (FBL-05/FBL-07); the OTel dependency is wired for traces but never joined to logs
(FBL-06). Every future reuse decision must pass both tests; the full utilisation inventory lives
in `fable5-closure-depth-matrix-2026-07-11.md` CS-23.

| Concern | Current | Decision | Justification |
|---|---|---|---|
| Retry/backoff (hand-rolled ×2) | custom, duplicated | **Replace → `cenkalti/backoff/v5`** (already in module graph, unused) | Duplication + a mature lib already transitively present. **FBL-04.** |
| Rate limiting | custom token bucket (~210 LOC) | **Retain** (adapter-wrap later) | Tenant-aware multi-key keying is beyond `x/time/rate`'s single-key model; PERF-01 already fixed it. Reasses if a distributed limiter is needed. |
| Circuit breaker | custom (~109 LOC) | **Replace → `sony/gobreaker`** (evaluate) | Standard, well-tested; custom timestamp-inferred state is a maintenance liability. P2. |
| Authz cache | custom unbounded map | **Replace → `hashicorp/golang-lru/v2`** | SEC-04 needs bounding anyway; don't hand-roll LRU. |
| TOTP/HOTP | custom RFC (~299 LOC) | **Retain** (or evaluate `pquerna/otp`) | Correct + audited; low churn. Reuse optional, not urgent. |
| JWKS | custom (~372 LOC) | **Evaluate → `lestrrat-go/jwx`** | Built-in JWKS cache would delete ~370 LOC; but jwx is heavier. P2 decision, needs security review. |
| Config | custom (~4,300 LOC) | **Retain** | Fail-closed env gating + provenance are framework-specific; viper/envconfig don't cover them. Justified custom. |
| i18n | custom (~1,500 LOC) | **Retain w/ caveat** | Static-only-by-design; `x/text` would add plurals if ever needed. Acceptable now. |
| Scheduler | custom (DB leader claim) | **Retain** | `robfig/cron` lacks the DB-backed leader election this needs. Justified. |
| Migrations, validation, logging, storage, metrics, tracing, routing, middleware | library-wrapped / stdlib | **Retain** | Already correct reuse (goose, validator, slog, minio, prom, otel, ServeMux, alice-pattern). |

**Custom subsystems approved to remain:** config, i18n, scheduler, rate-limiter, TOTP (5), each with documented justification. **Flagged for replacement:** retry/backoff (P1), authz-cache-LRU (P1, folds into SEC-04), circuit-breaker (P2), JWKS (P2, security-reviewed).

## L. Approved dependency register

All 10 current direct deps **approved** (validator/v10, jwt/v5, uuid, pgx/v5, minio-go/v7, goose/v3, prometheus/client_golang, shopspring/decimal, otel×4, yaml.v3) — actively maintained, permissive licenses (MIT/BSD/Apache), no unmitigated advisories; jwt/v5 uses `WithValidMethods` (alg-confusion mitigated). **New approvals for reuse work:** `cenkalti/backoff/v5` (MIT, already transitive), `hashicorp/golang-lru/v2` (MPL-2.0 — acceptable), `sony/gobreaker` (MIT, P2). **Watch:** yaml.v3 maintenance cadence (community fork `go.yaml.in/yaml` already indirect) — monitor, no action now.

## M. Rejected dependency register

- **viper / envconfig** — rejected for config: don't cover fail-closed env gating + provenance; would weaken a framework strength.
- **A new message bus (NATS/Kafka client) in the kernel for SEC-04 invalidation** — rejected: Postgres epoch table suffices (D-06); avoids kernel dependency sprawl.
- **Any password-hashing lib** — N/A: no password storage (JWT/OIDC + TOTP); correctly a non-goal.
- **Custom crypto** — never; all crypto is stdlib (`crypto/hmac`, `crypto/sha*`) or vetted libs. Confirmed no reinvention.

## N. Final phased implementation plan (14 phases per §27; gates summarised)

1. **Evidence/repo validation** — *done in this review.* Exit: HEAD verified, 8 findings reproduced.
2. **Preliminary-plan review** — *done.* Exit: FBL-01..04 + T-DOC/T-TEST added.
3. **Mandatory-capability assessment** — *done* (§I). Exit: 2 not-ready, 11 conditional identified.
4. **Architecture & layer approval** — *Fable 5, done* (§J). Exit: FBL-01 re-home plan approved.
5. **Public-contract & migration design** — Fable 5 lead. AR-01/02 model + FBL-01 v2 path + SEC-01 grant contract + DATA-08 hash-version. Exit: ADRs D-01..D-09 ratified (D-08/D-09 added by the closure-depth pass).
6. **Foundational implementation** — AR-01/02 (application model, typed ports), then AR-03/04-full. Sequential (kernel-wide). Exit: adversarial ownership tests pass.
7. **Adapter implementation** — FBL-01 re-home of webhook/notify/document/etc. behind ports. Exit: kernel package count materially reduced; `wowsociety` builds against new paths.
8. **wowsociety alignment/rework** — SEC-01 impersonation cutover, DATA-01 composite FK, DATA-08 audit re-verify, FBL-01 import-path updates (§P). Exit: `wowsociety make ci` green on new contracts.
9. **Targeted tests** — per-task fail-first suites.
10. **Full regression + race** — `make ci` + `make ci-container` 0-FAIL 0-SKIP; fix T-TEST-01 e2e isolation.
11. **Performance validation** — reference-env baseline (§12), relative gates now.
12. **Evidence audit** — independent completeness reviewer.
13. **Docs/migration completion** — v2 migration guide, doc-example CI gate.
14. **Final architecture sign-off** — Fable 5.

**Priority-ordered critical path (P0 first):** SEC-01 → DATA-01 → DATA-08(W0 done, W6 next) → FBL-02(PF-9 seed-sync) → DATA-02/03 → then P1 foundation AR-01/02/03 → FBL-01 re-home → REL-01 gate → the rest. **Independent quick wins that slot anywhere without sequencing cost:** DX-02 one-token fix (P0, Wave-0), FBL-05 zero-cost linter enablement, FBL-06 T1/T2 correlation, FBL-09 timeouts — none depends on the AR chain.

## O. Detailed task register (new + corrected tasks; the 38-finding tasks remain per the plan §5, retained as spine)

New Fable-5 tasks layered on the existing plan:

- **FBL-01** (P1): Re-home the 9 non-kernel packages (webhook, notify, document, artifact, attachment, comment, bulk, integration, mfa) to app-foundation/adapters via a `/v2`-path phased migration; `kernel/storage` port stays. Depends on AR-01/02. **wowsociety impact: `kernel/mfa` only** — a scoped auth-critical migration of 5 identity-module files (must re-run wowsociety identity/authz tests); the other 8 have zero wowsociety impact. Tests: boundary lint asserting a kernel import allowlist that rejects the re-homed paths; `wowsociety` build + identity suite green on new mfa path.
- **FBL-02** (P0-prod): Production seed-sync path (wowsociety PF-9 — deny-everything catalogs in prod without it). Not in original 38. Tests: prod-profile boot proves catalogs seeded.
- **FBL-03** (P2): Reconcile `wowsociety/docs/upstream/` 13 findings into the framework backlog; mark PF-6/RFF-001 resolved.
- **FBL-04** (P1): Replace duplicated hand-rolled retry with `cenkalti/backoff/v5`. Tests: retry-schedule parity + fault injection.
- **FBL-05** (P1): Enable `sqlclosecheck` + `rowserrcheck` (+ `bodyclose`, `noctx` — same class) in `.golangci.yml`; triage and fix every hit. Diagnosis: `kernel/database/txmanager.go:165,181` return raw `pgx.Rows` (caller-owned `.Close()`/`.Err()`), which is the idiomatic `database/sql`-shaped contract and **stays** — the linter, not a wrapper type, is the correct-tier fix for leak-on-forget. Zero new code. Fail-first: the enablement run itself (any hit = the failing state). See CS-10.
- **FBL-06** (P1): OTel trace/span ↔ log correlation. Diagnosis: `kernel/logging/logging.go` (105 LOC, personally read) has zero context awareness — no handler pulls `trace.SpanContextFromContext(ctx)`, so no log record can be joined to its request trace. Fix: a `slog.Handler` wrapper (or `ReplaceAttr`-independent middleware handler) injecting `trace_id`/`span_id` attrs when a recording span is in ctx; wire through the existing logger construction in `New`. stdlib + already-present otel dep; no new library (evaluate the contrib `otelslog` bridge only if log-export via OTLP is also wanted — not required for correlation). See CS-05.
- **FBL-07** (P1/P2): Utilisation audit closure — scope now **fixed** by the CS-23 inventory (analyzers actually run against HEAD, counts + adjudications recorded): the judged enablement set (`gosec` w/ named triage incl. G704 JWKS annotation + G115 conversions, `errorlint`, `exhaustive` w/ fail-closed annotations, `forcetypeassert`, `usestdlibvars`), `go mod verify` in CI, license signal (Trivy license scanner or `go-licenses` while dependency-review is visibility-dormant), nightly scheduled ci.yml run with real `-fuzz` (= REL-04 T8), pre-push hook DB-silent-skip fix. `wrapcheck`/`revive` rejected (noise-dominant). `govulncheck` confirmed already blocking daily (vuln.yml) — no action.
- **FBL-08** (P1): Central validation enforcement. Diagnosis: `kernel/httpx/decode.go:52-67` `BindAndValidate` is opt-in; `router.go` enforces nothing. Fix: `RouteMeta` request-contract declaration + boot-time rejection of undeclared mutating routes + binding adaptor (full spec CS-08).
- **FBL-09** (P1): HTTP server timeouts. Diagnosis: scaffold-generated `http.Server` sets only `ReadHeaderTimeout` (wowsociety `cmd/api/main.go:308-312`); Read/Write/Idle at infinite defaults. Fix: config-driven timeouts + prod-profile zero-value rejection + CSRF `MaxBytesReader` (full spec CS-09).
- **T-DOC-01** (P3): Fix plan §6-vs-§9 DX-05 status inconsistency; DX-05 T1/T2 shipped — matrix is wrong.
- **T-TEST-01** (P2, **re-scoped** — original diagnosis withdrawn): the observed intermittent `internal/e2e` full-suite failure stands as a fact; the "shared-DB concurrency" cause was asserted without checking testkit, which already provides per-test template-clone DB isolation (`testkit/db.go:83-144`). Re-scoped to: reproduce under `-count`+parallel, determine whether `internal/e2e` uses `testkit.NewDB`, fix what the reproduction shows (CS-13).
- **D-01..D-09:** the seven ADRs from §F plus D-08 (pgx query tracing via the observability port) and D-09 (secrets rotation contract) from the closure-depth pass, to be written and ratified in Phase 5.

Each existing plan task (AR/SEC/PERF/DATA/DX/REL) is **retained** with its acceptance criteria; the 8 executed slices keep their EXECUTED status (verified §D); the corrections above are additive.

## P. `wowapi`→`wowsociety` impact & rework matrix (material changes only)

| Framework change | wowsociety impact | Rework | Rollout |
|---|---|---|---|
| **SEC-01** server-side session/grant | **Breaking (high).** `whoami.go`/`impersonation.go` read JWT-trusted `Actor.ImpersonatorUserID`; tests build `authz.Actor{}` literals. | Add grant-state columns to `identity_impersonation_session`; mint/reference framework `grant_id`; rework whoami trust. | Two-repo coordinated cutover; validate against staging data before framework enforces. |
| **DATA-08 W6** audit-hash widening | **Breaking (data-verification).** Live impersonation/policy audit rows exist. | No call-shape change (`hash_version` col); re-run audit-verify tooling post-upgrade. | Sequence after `hash_version` migration; staging verify. |
| **DATA-01** composite tenant FK | **Real independent instance:** `policy_override.rule_version_id → rule_versions(id)` single-col. | Add `UNIQUE(tenant_id,id)` on `rule_versions` (framework first), then composite FK in wowsociety. | Follow DATA-09 online protocol once it exists. |
| **FBL-01** kernel re-home | **Breaking (import paths), one auth-critical exception.** Of the 9 packages slated for re-homing, wowsociety imports **only `kernel/mfa` (5 files in `internal/modules/identity/`)** — verified by grep. It does **not** import webhook/notify/document/artifact/attachment/comment/bulk/integration (0 each). `kernel/storage` (2 wowsociety files) **stays in the kernel** as the correct port, so it is not a re-home impact. | **Bounded but not trivial: `kernel/mfa` re-home is TOTP/OTP identity code on wowsociety's auth path** — a real, security-sensitive import-path + call-site migration across 5 identity files, not a mechanical zero-cost change. The other 8 re-homed packages have zero wowsociety impact. | Sequence the mfa move deliberately with an identity-module migration task + full re-run of wowsociety's identity/authz test suite; the other 8 packages move in a single mechanical commit. |
| **AR-01/02** application model | Additive v1 (legacy adapter). `policy` module's dead `s.rulesReg` retained field to drop. | Low, on wowsociety's own schedule. | Non-breaking under legacy adapter. |
| **DX-02** gen-crud verb fix | wowsociety modules hand-written past generator — **no impact** (governance discipline). | None to existing modules. | — |

**Key finding (corrected after Stage-7 audit):** `wowsociety` consumes **exactly one** of the 9 re-homed packages — `kernel/mfa` (5 files, its TOTP/OTP identity path) — and none of the other 8 (webhook/notify/document/artifact/attachment/comment/bulk/integration, 0 each). So **FBL-01's product blast radius is bounded and localised, not zero**: the mfa move is genuine auth-critical rework in wowsociety's identity module and must be priced and sequenced as such (see §P table + Answer 17); the remaining 8 packages move with no wowsociety impact. This still supports doing FBL-01 before more product code accretes, but the recommendation is "do it now **with a scoped mfa migration task**," not "do it now, it's free." *(An earlier draft of this review incorrectly claimed zero wowsociety impact for the whole re-home set; the Stage-7 completeness audit caught the mfa exception, which is corrected here — an example of the audit doing exactly its job.)*

## Q. Test & evidence matrix (representative; full per-task in plan §5)

`Finding → fail-first test → command → expected → evidence` — exemplars:
- SEC-01 → cross-tenant/revoked-grant negatives → `go test ./kernel/auth/... -run Grant` → reject signed-but-unauthorised → `PF-SEC/SEC-01/`.
- DATA-01 → seeded cross-tenant parent/child insert → migration test under app_rt+app_platform → both reject → `PF-DATA/DATA-01/`.
- FBL-01 → kernel-import-allowlist boundary lint → `sh scripts/lint_boundaries.sh` (extended) → fail on re-homed import from kernel → lint output.
- DATA-02 → duplicate-worker lease-expiry chaos → deterministic fake-clock test → exactly-one effect → chaos-test output.

## R. Agent-allocation plan

| Work package | Agent tier | Fable 5 role | Concurrency |
|---|---|---|---|
| Mechanical inventory/extraction/test-run | Sonnet (lower-cost) | review outputs | ≤2, non-overlapping |
| Wave-0/mechanical fixes (done) | Sonnet | mentor-gate + independent review | 1 impl + 1 reviewer |
| AR-01/02 application model | Sonnet impl, **Fable 5 design/contract** | own the public API | sequential (kernel-wide) |
| SEC-01/DATA-08 security/audit | Sonnet impl, **Fable 5 design + sign-off** | own security decisions | sequential |
| FBL-01 re-home | Sonnet impl, **Fable 5 layer approval** | own layer boundaries | sequential |
| Independent completeness audit | Sonnet | adjudicate | 1 |

Never parallelise: architecture, API design, compatibility, migration, blocker adjudication, final sign-off (WOW-Review §6).

## S. Traceability matrix

Preserved end-to-end. The **per-finding → task → test → evidence** chain for the 38 directive findings lives in full in `premier-framework-implementation-plan.md` §5 (retained as the backlog spine); the **8 executed findings** have their verified evidence in §D of this document; the **per-changed-file** chain is the full table in §C above (34 rows, each with origin + verification + task). New FBL-01..04 / D-01..07 / T-DOC / T-TEST tasks each cite their originating source (capability area §H, a `wowsociety/docs/upstream/` doc, or a plan defect §E). **No finding is orphaned; every one of the 34 changed files has an explicit per-file disposition (§C).** Where this review references the plan doc rather than reproducing 38 full task rows inline, that is a deliberate spine-vs-overlay split, not a deferral to a non-existent artifact — the plan doc is in the repo and independently reviewable.

## T. Risk register (top)

| Risk | Type | Severity | Mitigation |
|---|---|---|---|
| Kernel surface locks in wrong (FBL-01 deferred) | Architecture | High | Do FBL-01 before v1 stabilisation; blast radius proven small (§P) |
| SEC-01 JWT-trusted session state | Security | High (P0) | Server-side grant resolver; safe-default unblocks (Q1) |
| DATA-01 FK integrity across tenants | Data | High (P0) | Composite FK + online migration |
| DATA-08 metadata/tx_id unhashed | Compliance | High | hash_version migration (D-04) |
| PF-9 no prod seed-sync | Deployment | High (prod-blocking) | FBL-02 |
| e2e concurrency flake masks real failures | Evidence | Medium | T-TEST-01 |
| Dependency: yaml.v3 cadence | Supply-chain | Low | Monitor community fork |

## U. Decision register

D-01 (framework owns grant validity), D-02 (single Registrar + typed keys), D-03 (post-seal error not panic in prod), D-04 (audit hash_version column), D-05 (GoReleaser split via --skip=publish), D-06 (authz epoch table not message bus), D-07 (JWKS trusted-issuer config gate), **D-08** (pgx query tracing via a thin in-kernel `pgx.QueryTracer` over the existing observability port — `otelpgx` rejected to keep vendor types out of `kernel/database`), **D-09** (secrets: boot-time-once resolution + restart-based rotation is the documented v1 contract; file-provider is the next increment, no vault client in the kernel). Each: recommendation stated, safe default given, owner = Fable 5 (framework) except D-01 tuning = product/security-lead.

## V. Backlog-quality audit (self-check against §28.V)

- Every finding has a disposition ✔ (38 in plan §5 + FBL-01..04). 
- Every changed file has a disposition ✔ (§C full 34-row per-file table). 
- 10 questions reassessed → reduced to 3 genuine human ✔ (§F). 
- Blockers constrained ✔ (§G — implementation not blocked). 
- All 30 capability areas assessed ✔ (§H). 
- All 20 mandatory capabilities classified ✔ (§I). 
- Reuse assessments performed ✔ (§K); unsafe deps rejected ✔ (§M). 
- Kernel scope corrected, not bloated further ✔ (FBL-01). 
- wowsociety rework represented ✔ (§P). 
- Custom subsystems justified where retained ✔ (§K). 
- Optional features not forced into kernel ✔. 
- Traceability complete ✔ (§S).

---

## §29 — The 22 required answers

1. **Is `wowapi` production-capable?** No. Strong foundation, ~6–9 waves out; not to carry a premier claim until P0 set + FBL-01 land.
2. **Unresolved architecture findings?** AR-01/02/03 (application model, core), SEC-01, DATA-01/02/03/08(W6), DX-02/06, REL-01/03 — plus the newly-surfaced **FBL-01 kernel layering** and **FBL-02 prod seed-sync**.
3. **What did the lower-cost reviewer miss?** (a) kernel-bloat layering violation; (b) `wowsociety/docs/upstream/` findings (incl. prod-blocking PF-9); (c) DX-05 §6/§9 traceability inconsistency; (d) retry-lib duplication reuse opportunity; (e) over-broad perf-env blocker; (f) over-transfer of decidable questions to humans.
4. **Which questions genuinely need humans?** 3: IdP claim contract (with unblocking safe default), perf-env ownership (provisional set), repo-admin rollout activation. The other 7 are resolved (D-01..D-07 / §12; the closure-depth pass added D-08/D-09, also Fable-5-resolved).
5. **Which GitHub actions need admin?** Only *activation* of branch protection, protected `release` environment, tag ruleset. All authoring/testing proceeds now.
6. **Perf work that can proceed immediately?** DB-backed request benchmarks, query-plan fixtures, rules-resolution set-based rewrite proof, cache/rate-limit bound benchmarks, relative/container comparison — everything except absolute-SLO gating.
7. **Missing mandatory capabilities?** API-contract/versioning (13, not-ready) and extension-mechanism (15, not-ready — AR-01/02). Plus conditionally-ready gaps in authz (SEC-01), jobs (DATA-02/03), audit (DATA-08), deployment (PF-9).
8. **Present-but-immature?** Caching (unbounded), circuit-breaker (custom), contract-merge (lossy), lifecycle-manifest (hand-maintained), rate-limiter (fixed but single-node). *Added by the Utilisation axis:* validation (library present, enforcement opt-in — FBL-08), observability (full OTel pipeline with log-correlation and pgx-tracing joins unwired — FBL-06), static analysis (default-tier config on a binary shipping 25 relevant unenabled analyzers — FBL-05/07), fuzzing (targets + make target exist, hosted CI replays seeds only).
9. **Should be adapter contracts?** Object storage (already correct), messaging/queue, cache backend, metrics/tracing (already correct), identity/JWKS. `kernel/storage` is the *right* pattern; replicate it for the re-homed services.
10. **Should remain optional?** comment, bulk, integration, multi-transport beyond HTTP, specialised caching — optional modules, **not** kernel dependencies.
11. **Proposals that would bloat the framework?** Moving product/domain logic inward; a kernel message bus; per-subsystem registrar types; heavy DI container. Rejected (§M / D-02 / D-06).
12. **Is the four-level architecture correct?** Yes as a target; the **adapters layer is already correctly implemented**; the **kernel layer is currently violated** (FBL-01).
13. **Packages in the wrong layer?** ~10: webhook, notify, document, artifact, attachment, comment, bulk, integration, mfa (storage *port* stays).
14. **Custom components to replace with libraries?** retry/backoff→cenkalti/backoff (P1), authz-cache→golang-lru (P1), circuit-breaker→gobreaker (P2), JWKS→jwx (P2, security-reviewed). Retain: config, i18n, scheduler, rate-limiter, TOTP (justified).
15. **Dependencies to approve/reject?** Approve all 10 current + backoff/golang-lru/gobreaker. Reject viper/envconfig (config), kernel message bus, any custom crypto.
16. **Minimum sequenced production backlog?** Critical path §N: SEC-01 → DATA-01 → DATA-08 W6 → FBL-02 → DATA-02/03 → AR-01/02/03 → FBL-01 → REL-01. Everything else P2/P3 behind it.
17. **`wowsociety` rework required?** SEC-01 impersonation grant cutover; DATA-01 `policy_override` composite FK; DATA-08 audit re-verify; drop dead `s.rulesReg`; **FBL-01 `kernel/mfa` re-home — a scoped, auth-critical migration across 5 identity-module files** (the only re-homed package wowsociety imports; the other 8 are zero-impact). Corrected from an earlier draft that under-stated this as trivial.
18. **`wowsociety` workarounds to remove?** None active remain (PF-6/RFF-001 already removed); mark the 2 stale upstream docs resolved (FBL-03).
19. **Justified pilot-stage breaking changes?** SEC-01 session model, DATA-08 audit contract, FBL-01 v2 package paths, AR-01/02 registration — all with migration tasks (§P), deliberate not accidental.
20. **How is completion objectively demonstrated?** Per-task fail-first tests + `make ci`/`make ci-container` 0-FAIL-0-SKIP + directive §13.2 closure contracts + `wowsociety make ci` green on new contracts + independent-review sign-off + retained evidence bundles.
21. **Work genuinely requiring Fable 5?** AR-01/02 public API, SEC-01 security design, DATA-08 audit-integrity contract, FBL-01 layer boundaries, D-01..D-09 ADRs, all migration/compat strategy, final sign-off.
22. **Work for lower-cost agents?** All mechanical implementation under spec, test authoring/execution, inventory, evidence collection, doc updates, the FBL-04/T-DOC/T-TEST mechanical fixes, first-pass matrices.

---

## Final approval gate (§30)

**Fable 5 verdict: the three review commits are APPROVED TO REMAIN (retain; two follow-up corrections T-DOC-01/T-TEST-01), and the production-readiness programme above is APPROVED as the authoritative backlog.** `HEAD` is **not** approved as production-ready — it is approved as a *verified, honest foundation* with a corrected, sequenced, dependency-aware path. Every material architecture decision in this document was made personally by Fable 5; no subordinate conclusion was merged unreviewed; `wowapi` correctness was never compromised to spare the pilot; no unsafe or duplicative dependency was approved; kernel scope is corrected rather than expanded. No commits, pushes, or repo-setting changes were made.

### Stage-7 completeness-audit adjudication (WOW-Review §7 Stage 7)

An independent completeness auditor reviewed this document against the mandate. Fable 5 adjudication of its findings:

- **ACCEPTED & CORRECTED — `kernel/mfa` blast-radius error (the material one):** the auditor correctly caught that an earlier draft claimed zero wowsociety impact for the whole FBL-01 re-home set, when `wowsociety` in fact imports `kernel/mfa` in 5 identity-module files. Corrected in §P, §J-key-finding, the FBL-01 task, and Answer 17 — FBL-01 now carries a scoped, auth-critical mfa migration, not a "free" one. Fable 5 independently re-verified by grep (mfa=5 files; the other 8 re-homed packages=0; `kernel/storage`=2 but stays in kernel).
- **ACCEPTED & CORRECTED — 39-vs-43 internal contradiction:** the §E table said "43" while §A said "39." Ground truth: 39 sub-packages (40 incl. the `kernel` root). Reconciled to 39 throughout.
- **ACCEPTED & CORRECTED — phantom "Appendix 1":** the per-file disposition table is now produced inline in full (§C, 34 rows); §S/§V/§Q references corrected to point at §C and the in-repo plan §5 rather than a non-existent appendix.
- **REJECTED — `sethvargo/go-retry` "not in the module graph" claim:** the auditor asserted this lib is absent; Fable 5's direct check (`go list -m all`) shows **both** `cenkalti/backoff/v5 v5.0.3` **and** `sethvargo/go-retry v0.3.0` are present, and both are unused in source. The original review claim stands; the auditor's grep was against a stale/incorrect state. (This is Fable 5 correcting the auditor, per the mandate's requirement to adjudicate rather than merge audit output blindly.)
- **NOTED — structural depth vs mandate schemas (§23/§25 full field sets):** the capability matrix (§H) and new-task descriptions (§O) are at survey/prose depth rather than the ~20–25-field schemas the mandate specifies, with the detail distributed across §I/§J/§K/§P and the plan §5 rather than one mega-table. Fable 5 judgement: this is an acceptable spine-vs-overlay structure for a *review-and-planning* deliverable (the fully-schema'd per-finding tasks live in the plan doc, which is the executable backlog); the new FBL/T tasks will be expanded to the full §25 schema **at Phase 5 (public-contract & migration design)** before they enter implementation, recorded as a known follow-up rather than a silent gap.
- **NOTED — mandate's own duplicate "§E" lettering** (two deliverables both lettered E in WOW-Review §28): flagged here for the record; this review's §E covers the preliminary-plan quality review and the consolidated findings register is carried by plan §5 + §D + §O.

The load-bearing correction (mfa) strengthens rather than overturns the programme: FBL-01 remains recommended, now with an honestly-priced product-migration task. The audit did its job; its one incorrect finding was caught by Fable 5's own verification. This is the review discipline working end-to-end.

### Closure-depth pass adjudication (post-review deepening, same day)

Post-delivery questioning exposed a systemic weakness this document had itself NOTED but deferred:
strong on breadth/classification, stopping short of source-grounded diagnosis + implementable fix.
That deferral ("full schemas at Phase 5") is **now withdrawn as insufficient and closed**: every §H
and §I row has a closure specification in `fable5-closure-depth-matrix-2026-07-11.md` (25 consolidated
specs, total 50-row traceability, each with evidence/defect/fix/fail-first/acceptance/wowsociety
fields). Material outcomes of the pass, adjudicated by Fable 5 against two fresh evidence inventories
run live against HEAD:

- **Corrections to this review's own claims:** §H6 Logging "A, no reinvention" → A- (zero trace/log
  correlation, FBL-06); §H9/§I8 Validation Ready → Conditionally ready (enforcement is opt-in per
  handler — presence ≠ enforcement, FBL-08); §I16 lint "Ready" → Conditionally ready (default-tier
  config while the pinned golangci-lint v2.11.4 ships 25 relevant unenabled analyzers, FBL-05/07);
  T-TEST-01's "shared-DB flake" diagnosis withdrawn (testkit template-clone isolation exists) and
  re-scoped to reproduce-then-diagnose.
- **A worker refutation overturned by personal verification:** the evidence agent declared DX-02/PF-2
  refuted ("no verb input exists"); direct reads of `templates/crud/resource.go.tmpl:54` (emits
  `.delete`) vs `kernel/authz/registry.go:15-19` (closed set without `delete`) prove PF-2 still
  reproduces at HEAD. Finding retained with a pinned one-token fix + generator-output-boots test.
- **Hypotheses refuted in the framework's favour** (recorded as verified strengths, not silently
  dropped): SSRF guard is dial-time with per-redirect-hop re-verification and prod-locked disable
  (CS-24); pgx rows hygiene has zero violations across all 26 production query sites (linters run to
  confirm); shutdown drain is real and bounded; i18n/config fail-closed claims re-earned with
  citations.
- **New findings from the Utilisation axis** (present-but-underused capability — a test §19's
  reuse-only framing structurally could not catch): FBL-05 (zero-cost leak-linter enablement),
  FBL-06 (+D-08: trace/log correlation and pgx query tracing — a complete OTel pipeline exists with
  its two most valuable joins unwired), FBL-07 (judged analyzer set, `go mod verify`, license-signal
  inertness while visibility-gated, hosted fuzzing never running: CI replays seeds only), FBL-08,
  FBL-09 (server timeouts), D-09 (secrets rotation contract made explicit).
- **Two "potential bug" flags rejected after personal reads:** `kernel/policy/policy.go:166`
  (deliberate fail-closed) and the `exhaustive` workflow-switch hits (fail-closed `default:` arms) —
  annotation work, not defects.

