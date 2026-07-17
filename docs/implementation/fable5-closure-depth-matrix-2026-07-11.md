# Fable 5 — Closure-Depth Matrix (companion to the final architecture review)

- **Purpose:** convert every §H capability-area row (30) and §I mandatory-capability row (20) of
  `fable5-final-architecture-review-2026-07-11.md` from *classification + gate pointer* into a
  **closure specification** an implementation agent can execute without repeating the source
  investigation. This document exists because post-review questioning demonstrated the review
  systematically stopped at "Class B — see gate X" depth.
- **Status:** COMPLETE — all 25 closure specifications populated from live-source evidence (two
  read-only inventory passes + Fable 5 personal verification and adjudication); passed the
  independent review gate in two iterations (iteration 1: FAIL, 9 findings, all fixed; iteration 2:
  PASS, zero new findings).
- **Scope discipline:** documentation/planning artifact only. No production code is changed by
  this document. Every "fix" below is a *specified future task*, traceable to a backlog ID.

## 0. Spec template (every CS row below must fill all fields)

| Field | Meaning |
|---|---|
| Current evidence | file:symbol[:line] of what exists today, verified against live source |
| Defect / gap | what is specifically missing, unsafe, incomplete, or suboptimal |
| Consequence | runtime / security / correctness / maintenance impact if unaddressed |
| Why insufficient | why current state fails the production-readiness bar (§15.A criteria named) |
| Target state | concrete end state, not an aspiration |
| Fix | the actual implementation, named types/functions/config — not a section pointer |
| Reuse tier | stdlib / already-present dep / fuller config of existing tool / mature new lib / justified custom |
| Affected surface | packages, public contracts, config keys, docs |
| Fail-first verification | the test or command that fails before the fix and passes after |
| Acceptance criteria | objective, binary |
| Closure evidence | what artifact proves it (test output, lint run, CI log) |
| Dependencies / sequence | what must land first |
| Priority / risk | P0–P3 + risk class |
| wowsociety impact | breaking / additive / none, with the concrete rework if any |
| Backlog ID | final task ID(s) |

**Template-conformance notes (per the independent review gate):** (1) *Target state* is merged
into *Fix* wherever stating both would duplicate one sentence — the field is satisfied, not
skipped. (2) For **verify-outcome rows with no task** (CS-03 config, CS-04 errors, CS-19 i18n,
CS-24 SSRF), the defect/consequence/fix/fail-first fields are **n/a by verdict** — the spec's
purpose there is the re-earned evidence itself, and the citations in the spec body are the
closure record. (3) *Closure evidence* for every task-bearing spec is centralised in the §2.1
register below (one canonical artifact per CS) rather than repeated per spec.

## 1. Dedup mapping — 50 rows → consolidated closure specs

Every §H row (H1–H30) and §I row (I1–I20) maps to exactly one CS; overlaps are merged only where
the underlying defect is identical. Traceability is total: no row is dropped.

| CS | Title | §H rows | §I rows | Anchor tasks |
|---|---|---|---|---|
| CS-01 | Kernel layering & module structure | H1, H20, H27(part) | I1 | FBL-01 |
| CS-02 | Registration model, DI, extensibility, lifecycle manifest | H2, H4, H19 | I2, I15 | AR-01/02/03 |
| CS-03 | Configuration (verify-Ready) | H3 | I3 | — |
| CS-04 | Structured errors (verify-Ready + error-wrapping hygiene) | H5 | I4 | new: lint utilisation |
| CS-05 | Logging ↔ trace correlation & observability | H6, H7 | I5, I6(part) | new FBL-06 |
| CS-06 | Health/readiness migration-currency | H29(part) | I6(part) | DX-07 |
| CS-07 | Identity & session security | H8 | I7 | SEC-01 |
| CS-08 | Validation enforcement path (verify-Ready) | H9 | I8 | — |
| CS-09 | HTTP transport hygiene (timeouts, body limits) | H10, H11 | — | new (candidate) |
| CS-10 | Data access & pgx resource contract | H12, H13 | I10 | new FBL-05 |
| CS-11 | Jobs, outbox, lease/fencing, drain | H14, H24 | I9, I14 | DATA-02/03 |
| CS-12 | Resilience primitives (retry, breaker, limiter) | H15 | I11 | FBL-04, SEC-04(part) |
| CS-13 | Test infrastructure (e2e isolation, fuzz/race utilisation) | H16 | I12 | T-TEST-01 |
| CS-14 | Generator correctness | H17 | — | DX-02 |
| CS-15 | API contract & compatibility gates | H18, H21 | I13, I18 | DX-06, REL-03 |
| CS-16 | Performance verification programme | H22 | — | PERF-02..05 (§12 constrained) |
| CS-17 | Authz cache bounding & invalidation | H23 | — | SEC-04 |
| CS-18 | Tenant FK integrity | H25 | — | DATA-01 |
| CS-19 | i18n runtime behaviour (verify-retain) | H26 | — | — |
| CS-20 | Audit hash-chain completeness | H28 | I19 | DATA-08 W6 |
| CS-21 | Deployment readiness & seed-sync | H29 | I17 | FBL-02 |
| CS-22 | Documentation gates | H30 | I20 | new: doc-example gate spec |
| CS-23 | Static-analysis & CI-gate utilisation | — | I16 (**downgrade candidate**) | new FBL-05/07 |
| CS-24 | Outbound-HTTP SSRF guard depth | H8(part), H15(part) | — | verify; new if gap |
| CS-25 | Secrets lifecycle (rotation, provider surface) | H3(part) | I3(part) | verify; new if gap |

Rows H1–H30 / I1–I20 are all accounted for above (several §I "Ready" rows appear as
*verify-Ready* specs: a Ready verdict must be re-earned with evidence, not presumed).

## 2. Closure specifications

Two anchor classes. **Plan-anchored** specs summarise diagnosis + fix inline and cite the
executable task table in `premier-framework-implementation-plan.md` §5 (which carries the full
per-task acceptance/tests/evidence rows — that table is the implementation contract, verified to
contain file:symbol-level evidence, not a pointer). **Direct** specs are fully self-contained here
because no plan finding covers them.

---

### CS-01 — Kernel layering & module structure *(direct; FBL-01)*
- **Current evidence:** `go list ./kernel/...` = 39 sub-packages (40 incl. root; personally verified). Nine are app-foundation/adapter concerns wearing kernel paths: `webhook, notify, document, artifact, attachment, comment, bulk, integration, mfa`. `kernel/storage` is a correct port and stays. wowsociety imports exactly one re-home candidate: `kernel/mfa` (5 files, `internal/modules/identity/`) — grep-verified; the other 8 = 0 imports.
- **Defect:** delivery engines with network I/O, document services, and feature subsystems live at kernel import paths; the kernel cannot honour "small and stable" while owning them.
- **Consequence:** v1 stabilisation would freeze the wrong public surface; every kernel-version bump drags nine unrelated subsystems' churn with it.
- **Why insufficient:** fails §15.A "stable abstraction" + WOW-Review §18's four-level architecture at the definitional level.
- **Target state:** kernel = `lifecycle, config, errors, model, secrets, module, context, database(tx contract), storage(port), logging, httpx-core, authz-core, audit-core` tier; the nine move to `foundation/<pkg>` (new top-level, app-foundation layer) with any true adapter halves split into `adapters/`.
- **Fix (mechanics, not just intent):** (1) create `foundation/` tree; (2) `git mv` each package, update import paths repo-wide (mechanical, 8 of 9 are zero-consumer outside wowapi); (3) `kernel/mfa` → `foundation/mfa` with a **deprecated forwarding shim** left at `kernel/mfa` (type aliases + var forwarding) for one minor version so wowsociety migrates on its own schedule, then remove; (4) extend `depguard` (`.golangci.yml` kernel rule) to deny `kernel → foundation` imports and add a `foundation` rule denying `foundation → app`; (5) extend `scripts/lint_boundaries.sh` allowlist so a *new* kernel package addition fails CI without an explicit allowlist edit (review-forcing).
- **Reuse tier:** fuller configuration of existing tools (depguard + boundaries script already exist — this is a utilisation win, no new tooling).
- **Fail-first:** the extended depguard rule fails **today** against the nine packages (that failing run is the diagnosis artifact); passes after re-home.
- **Acceptance:** `go list ./kernel/... | wc -l` ≤ target-list count; depguard + boundaries lint green; wowsociety identity suite green on `foundation/mfa` (or on the shim during the grace window).
- **Dependencies:** AR-01/02 first (re-homing mid-registration-rework causes double churn). **Priority:** P1, pre-v1-stabilisation hard gate. **Risk:** import-path churn only; behaviour-preserving moves.
- **wowsociety:** `kernel/mfa` only — scoped auth-critical migration of 5 identity files + full identity/authz suite re-run; other 8 zero-impact. **IDs:** FBL-01, D-02/D-03 (sequencing context).

### CS-16 — Performance verification programme *(plan-anchored + utilisation finding)*
- **Current evidence:** benchbudget gate fails closed (PERF-06 verified at HEAD); `bench-budgets.txt` = 43 budgeted entries; **exactly 8 of 55 non-cmd packages have any `Benchmark*`** (kernel/{audit,authz,config,filtering,httpx,pagination,policy,sequence} — matches `BENCH_PKGS`, `Makefile:206-214`), leaving hot-path candidates **kernel/database, jobs, outbox, workflow, auth, mfa, httpclient** and all adapters unbenched; no reference environment exists.
- **Defect:** the budget-gate *mechanism* is mature but its *coverage* is 8/55 — a perf regression in the transaction manager, job claim loop, or outbox relay is invisible to the only performance gate the repo has; and PERF-02..05 were framed as wholly blocked on a reference env when only absolute-SLO gating is.
- **Consequence:** "premier framework" performance claims are unverifiable precisely on the packages a consumer exercises most.
- **Why insufficient:** fails §15.A "operational evidence"; also the Utilisation principle — gate infrastructure present, majority of surface unenrolled.
- **Target state / Fix:** (1) provisional reference env per §F Q9 (Linux amd64 GH runner + committed `perf/reference-schema1.json`, advisory first); (2) add benchmarks + budget entries for the 7 named hot-path packages (claim/finalize loop, tenant-tx open/commit, relay dispatch batch, token verify, TOTP derive, guarded dial); (3) execute plan PERF-02 (complete-request vs real PostgreSQL), PERF-03 (rules resolution as bounded SQL), PERF-04 (sweeper N+1/unbounded materialisation), PERF-05 (checksum behaviour) as **relative/container** comparisons now — each plan task table already specifies its measurements; only absolute-SLO thresholds wait on env ownership.
- **Reuse tier:** fuller use of existing tooling (benchbudget + BENCH_PKGS + testkit DB).
- **Fail-first:** `make bench-budget` after adding a hot-path package with no budget entry fails (PERF-06's own enforcement); each PERF item's before/after relative run.
- **Acceptance:** BENCH_PKGS covers the 7 named packages; budgets exit-0; PERF-02..05 relative evidence recorded.
- **Priority:** P1. **wowsociety:** none direct. **IDs:** PERF-02..05, Q9-provisional, FBL-07 (bench-coverage item).

### CS-02 — Registration model, DI, extensibility, lifecycle manifest *(plan-anchored)*
- **Current evidence:** mutable mega-`Context` (39-method interface, `module/module.go`); `authz.Registry.Register(p Permission)` has **no owner parameter at all** (plan AR-01 T5 — the widest ownership gap); registries return backing maps/slices from `Specs()`/`Points()`; hand-maintained `kernel/lifecycle` manifest; zero `port.Key`-style typed ports.
- **Defect:** any module can claim any declaration identity; post-boot mutation silently succeeds; snapshot reads are aliasing internal state; provider graph is by-hand and unvalidated.
- **Consequence:** the entire ownership story authz/audit assume is unenforced — a hostile or buggy module can shadow another module's permissions/rules; no compile-time detection of missing/duplicate providers.
- **Why insufficient:** fails §15.A "stable abstraction", "correct implementation", "negative-path coverage" — there are no adversarial ownership tests because there is no ownership boundary to test.
- **Target state / Fix:** plan AR-01 T1–T11 + AR-02 T1–T7 exactly as specified (owner-bound `Registrar` capability minted from `Manifest.ID` with unexported seal; `collect→validate→seal→expose`; `port.Key[T]` + `Define/Provide/Require/Resolve`; graph validation of duplicates/cycles/missing/scope at boot; deterministic model hash; legacy adapter for compatibility). Fable 5 pre-decisions D-02 (one Registrar type + typed keys) and D-03 (post-seal error-not-panic in prod) resolve the plan's two open design questions in §5.1 note (5)/(6).
- **Reuse tier:** justified custom (this IS the framework's core abstraction); reuses `kernel/lifecycle` scope-rank logic per AR-02 T4 rather than duplicating.
- **Fail-first:** adversarial cross-module claim fixtures (plan AR-01 T3–T6) fail today by *succeeding* — the claim goes through; after fix they error at the registrar boundary.
- **Acceptance / evidence:** plan §5.1 per-task criteria verbatim; model-hash determinism; `-race` clean.
- **Dependencies:** AR-01 T1/T2 before everything; AR-03 after; blocks FBL-01 phase-2 and AR-04 T2–T5.
- **Priority:** P1 (core). **wowsociety:** not breaking under legacy adapter; drop dead `s.rulesReg` field (`policy/pack.go:334-338`). **IDs:** AR-01, AR-02, AR-03, D-02, D-03.

### CS-07 — Identity & session security *(plan-anchored)*
- **Current evidence:** `kernel/auth/auth.go:181-208` `Verifier.Actor` copies `TenantID`/`ImpersonatorUserID`/`BreakGlass` straight from JWT claims; membership checked only when `CapacityID != uuid.Nil`; `user_tenant_access` table exists (`00002_core_identity.sql:54-83`) but **no Go code queries it**; no grant table exists.
- **Defect:** server trusts client-presented session/impersonation state; a validly-signed token with stale/forged tenant or impersonation claims is honoured.
- **Consequence:** tenant-isolation bypass via stale membership; unauditable impersonation. This is the top-ranked security risk (§A).
- **Why insufficient:** fails §15.A "secure defaults" and "negative-path coverage" — no revoked-membership/forged-grant negatives exist because the checks don't.
- **Target state / Fix:** plan SEC-01 T1–T7 (grant table w/ RLS FORCE + one-active-grant partial index; unconditional `ActiveTenantAccess` in `Verifier.Actor`; zero-tenant rejection pre-`WithTenantID`; server-side capacity selection; grant-ID resolver replacing claim copy; auth_time/acr/amr freshness; credential schemes). Safe default per review §F Q1: framework owns the grant record keyed by grant-ID; IdP claim shape is tuning, not a blocker.
- **Reuse tier:** stdlib + existing deps (pgx, jwt/v5 `WithValidMethods` already correct); no new dependency.
- **Fail-first:** adversarial membership/grant negatives (plan SEC-01 required test classes: token substitution, zero-tenant, stale membership, revoked capacity, expired step-up, rotation, JWKS failure) — all currently *pass wrongly* or are untestable.
- **Dependencies:** none upstream; DATA-07 hard-depends on this. **Priority:** P0.
- **wowsociety:** BREAKING for impersonation (plan §5.2: `whoami.go:39,51`, `impersonation.go`, test fixtures build `authz.Actor{}` literals) — two-repo coordinated cutover, framework owns grant validity (D-01). **IDs:** SEC-01, D-01.

### CS-11 — Jobs, outbox, lease/fencing, drain *(plan-anchored)*
- **Current evidence:** jobs claim SQL returns no lease token/generation; finalize matches only `id`; `ReclaimStalled` blind-resets (plan DATA-02 evidence — confirmed race: reclaimed worker A's late finalize overwrites B's outcome). `notify/service.go:464-467` (the `SendPending` doc-comment) self-documents "Real production deployments should move the network call outside the tx to avoid holding locks during I/O"; webhook delivery + secret resolution run inside `plat.WithTenant(...)` (DATA-03). *(Citation corrected by the review gate from :446-449 — the comment sits ~20 lines below the plan doc's original anchor.)* Bulk: migration 00016 claims SKIP LOCKED, `bulk.go:123-144` actually does a plain unlocked SELECT (DATA-04).
- **Defect:** duplicate external effects on lease expiry; remote I/O holds DB transactions open; multi-worker bulk is documented-safe but actually unsafe.
- **Consequence:** duplicate side-effects (double notifications/webhook posts), pool exhaustion under provider latency, silent lost updates.
- **Why insufficient:** fails "correct implementation" + "operational evidence" — the at-least-once story has no fencing, so it is at-least-once-with-overwrites.
- **Target state / Fix:** plan DATA-02 T1–T7, DATA-03 T1–T8, DATA-04 T1–T6 — the single shared lease/fencing primitive (DATA-02 T1) first, then three-stage claim→effect-outside-tx→fenced-finalize for notify/webhook, SKIP-LOCKED bounded batch for bulk, chaos tests at every named boundary.
- **Reuse tier:** justified custom on stdlib+pgx (a DB-backed fencing primitive is framework-specific); explicitly **not** a new queue dependency.
- **Fail-first:** the named chaos tests (DATA-02 T7 / DATA-03 T8 / DATA-04 T6) — constructible today and failing (duplicate effect observed) before the fix.
- **Dependencies:** DATA-02 T1 is the keystone; DATA-09's harness for the riskiest rollouts. **Priority:** P0 (DATA-02/03), P1 (DATA-04).
- **Drain (I9) — evidence now in, gap narrower than assumed:** a real bounded drain exists — `app/worker.go:108-141` self-draining goroutines racing a `ShutdownDrain` budget (default 30s), `errDrainTimeout` on exceed, leaked work recovered via `ReclaimStalled`; HTTP shutdown via `RunHooks` (`app/run.go:43-81`) with 30s stop timeout. §I9's "drain proof" conditional is therefore satisfied by existing code + tests; the *remaining* I9 gap is exactly DATA-02's fencing (a drained-past worker's late finalize), not a missing drain.
- **Evidence refinement:** the current at-least-once posture is *explicitly documented* as an accepted idempotent-worker tradeoff (`kernel/jobs/runner.go:437-438,108-113`); DATA-03's exposure is scoped to **external side effects only** — DB effects are already exactly-once via the outbox inbox-dedup (`kernel/outbox/relay.go:191,205-219`), with the re-dispatch window at the outer-tx commit (`relay.go:157`). The fix contract stands; the honest framing is "make the documented assumption enforceable" not "fix an unacknowledged race."
- **wowsociety:** none today (zero jobs/notify/webhook/bulk imports — plan-verified). **IDs:** DATA-02, DATA-03, DATA-04.

### CS-14 — Generator correctness *(plan-anchored; PF-2 re-verified at HEAD by Fable 5)*
- **Current evidence (two-sided, personally verified):** `internal/cli/templates/crud/resource.go.tmpl:54` emits `RouteMeta{Permission: "{{.PermPrefix}}.delete"}`; `kernel/authz/registry.go:15-19`'s closed verb set = {create, read, list, update, deactivate, restore, approve, reject, assign, export, admin, ingest, activate} — **no `delete`**; `registry.go:88-90` rejects it at `Register`. Every `gen crud` output is dead-on-arrival at boot, exactly as wowsociety PF-2 documented at v1.0.0 — still true at HEAD. Also: `templates/module/module.go.tmpl:44-46` auto-wires migrations/seeds/OpenAPI; only line 48's routes/permissions/health/ports TODO remains (plan's corrected scope).
- **Fix:** emit `deactivate` (matches the template's own soft-delete TODO at `resource.go.tmpl:146` — the semantics already agree, only the string is wrong) + the PF-2-suggested **generator-output-boots CI test** (DX-01 T5's isolated-scaffold harness is the vehicle: generate → boot → assert no closed-set rejection). Do **not** widen the kernel verb set — the closed-set discipline is correct.
- **Fail-first:** the generator-output-boots test fails today with the closed-verb-set rejection.
- **Priority:** P0 (Wave-0 slice — one template token + one harness test). **wowsociety:** none to existing modules (hand-repaired already); closes PF-2 in its upstream register (FBL-03). **IDs:** DX-01 T5, DX-02, FBL-03.

### CS-15 — API contract & compatibility gates *(plan-anchored)*
- **Current evidence:** `internal/cli/openapi_cmd.go:139-144` `mergeFragment` unmarshals only `paths` + `components.schemas` into an anonymous struct — all other OpenAPI 3.1 top-level fields silently dropped (duplicate paths/schemas *do* fail loudly, `:148-158` — the loud-on-collision half already exists); zero `apidiff`/`gorelease` hits in Makefile/CI/docs; zero `/vN` path versioning (plan REL-03; evidence pass item 10).
- **Defect/Consequence:** a module declaring `security` on a fragment ships an API with that requirement silently absent from the published contract — a *security-adjacent* documentation lie; breaking API changes are detectable only by consumers breaking.
- **Fix:** DX-06 T1–T3 (full-field merge w/ per-field policy, 3.1.1/2020-12 validation, semantic diff) with **single ownership shared with AR-03 T2** (identical contract — plan flags this; Fable 5 assigns ownership to DX-06, AR-03 T2 becomes a cross-reference); REL-03 split into REL-03a/REL-03b **per the plan's own recommendation** (`premier-framework-implementation-plan.md:694`): **REL-03a (buildable now) = T1, T2, T4, T6, T8, T9** (Go API diff via `golang.org/x/exp/apidiff`/`gorelease`, compile matrix, config compat, migration drill, arch smoke, SBOM-verify fold-in); **REL-03b (blocked) = T3 (on DX-06), T5 (on DX-03/AR-03), T7 (on DX-04)**.
- **Reuse tier:** mature existing tooling — `apidiff`/`gorelease` are the standard Go answers; an OpenAPI 3.1 validator dependency needed for DX-06 T2 (evaluate `pb33f/libopenapi` — decision at implementation, security-review licence).
- **Fail-first:** fixture fragment with a `security` block → merged output today lacks it (provable now); seeded breaking-API fixture fails the new diff gate.
- **Priority:** P1 — but **I13 is one of only two not-ready mandatory capabilities**, so it heads the P1 queue after the P0 chain. **wowsociety:** audit module OpenAPI fragments for silently-dropped fields once T1 ships. **IDs:** DX-06 (owner), AR-03-T2 (xref), REL-03a/b.

### CS-18 — Tenant FK integrity *(plan-anchored)*
- **Current evidence:** 8 tenant-scoped child tables FK only the parent `id`, never `(tenant_id, id)` (plan DATA-01, confirmed exactly); RLS proves the child's tenant, nothing proves parent-child agreement.
- **Fix:** plan DATA-01 T1–T8 (CONCURRENTLY unique parent indexes → catalog scanner as permanent CI gate → mismatch audit → NOT VALID composite FKs → VALIDATE → negative tests under both roles). T6 (CI gate) first if sequencing allows — cheapest, most durable.
- **Fail-first:** platform-role seeded cross-tenant parent/child insert succeeds today; fails after.
- **Priority:** P0. **wowsociety:** real independent instance `policy_override.rule_version_id` (`00002_override.sql:16`) — needs wowapi's `UNIQUE(tenant_id,id)` on `rule_versions` first, then follow DATA-09 protocol. **IDs:** DATA-01, DATA-09 (T1–T5 precede the risky steps).

### CS-20 — Audit hash-chain completeness *(plan-anchored)*
- **Current evidence:** `kernel/audit/audit.go:130-179` — `chainHash` covers 15 length-prefixed fields (prev_hash…reason) but excludes `metadata` (**documented** at `:155-159`: jsonb round-trip reformatting makes the stored form unreproducible) and `tx_id` (inserted via `pg_current_xact_id()` at `:140`, never hashed); Verify (`:195-248`) recomputes with the identical field list, plus `Anchor`/`CheckAnchor` (`:253-311`) for tail-truncation. The exclusions are deliberate and internally consistent — which does not make them sufficient: tamper on `metadata`/`tx_id` remains chain-invisible, and the documented jsonb rationale is precisely why the fix must hash a **canonicalized pre-serialization form**, not the stored jsonb. No `hash_version` column exists.
- **Defect/Consequence:** an attacker (or bug) can alter audit `metadata`/`tx_id` on a row without breaking the chain — the tamper-evidence guarantee is partial, which for compliance evidence is close to none.
- **Fix:** DATA-08 W6-T1 with D-04 ratified (add `hash_version smallint NOT NULL DEFAULT 1` in the same migration; canonicalise metadata pre-serialisation, never hash stored jsonb; verification branches by version) + W6-T2..T5 per plan.
- **Fail-first:** tamper test mutating `metadata` on a chained row — passes verification today (that's the defect), fails after.
- **Priority:** P0/P1 (W6-T1 heads it). **wowsociety:** BREAKING-adjacent — live audit rows exist (`identity/service.go`, `impersonation.go` grant/revoke writes); historical rows must verify under v1 branch; staging verification pass required before `FRAMEWORK_VERSION` bump. **IDs:** DATA-08 W6, D-04.

### CS-21 — Deployment readiness & seed-sync *(plan-anchored + FBL-02)*
- **Current evidence:** readiness template registers only `"db"`+`"seeds"` checks — no migration-currency check despite the health contract's own doc claiming it (plan DX-07 evidence, incl. `CapacityMode` advisory-default and `HTTPMaxInFlight=0` backpressure-off defaults); **no production seed-sync path at all** (wowsociety PF-9: prod boots with deny-everything catalogs — prod-blocking, never in the original 38).
- **Fix:** DX-07 T1–T4 (migration-currency in readiness; seed/rule/model-hash checks; `config doctor` via `go env GOMOD`; prod-profile capacity/backpressure enforcement behind AR-04's waiver mechanism — built once) + **FBL-02**: a `wowapi seed sync --env prod` path (idempotent, RLS-respecting, versioned catalog manifests, dry-run + audit) — design detail to be ratified in Phase 5, but the acceptance bar is fixed now: *a prod-profile boot on an empty catalog DB reaches readiness only after seed-sync has run, and the readiness payload reports the seed/catalog hash*.
- **Fail-first:** boot prod-profile against stale-migrated DB → readiness returns 200 today (defect), 503 after; prod boot with empty catalogs → currently silently deny-everything, after: named readiness failure.
- **Priority:** P0-prod (FBL-02), P1 (DX-07). **wowsociety:** its generated `cmd/api/main.go:240-243` has the identical readiness gap — backport after T1; PF-9 is *its* finding, closing it closes the register entry (FBL-03). **IDs:** FBL-02, DX-07, AR-04-T5 (waiver, xref), FBL-03.
- **Evidence refinement (source pass):** framework readiness mechanism itself is correct and fail-closed — `kernel/httpx/health.go:52-79` runs each check with a 3s timeout, 503 on any failure, reports `config_fingerprint`; `app/health.go:9-14` documents DB/migration checks as a *comment-only contract* supplied via `extra`. The defect is thus precisely located: contract-by-comment at the seam + template omission at the product end.

### CS-05 — Logging ↔ trace correlation & observability *(direct; FBL-06)*
- **Current evidence (both passes agree, zero contradictions):** `kernel/logging/logging.go:85-104` builds a plain slog logger, no ctx awareness. `kernel/observability/tracing.go:66-86` `Trace(tr)` middleware creates a real span, tags `http.request_id`. `kernel/observability/middleware.go:46-62` `AccessLog` logs `request_id` — never trace/span IDs. The `Tracer`/`Span` **port itself has no `TraceID()`/`SpanID()` accessor** (`tracing.go:17-39`); `adapters/tracing/otel/otel.go:99-111` wraps the otel span but never exposes `SpanContext` through the port. Repo-wide: zero `SpanContextFromContext` hits; `otelslog` bridge absent from the module graph. **pgx query tracing absent**: `kernel/database/database.go:128-148` sets only `MaxConns`, no `pgx.QueryTracer` — DB time is invisible in every trace despite a complete OTel pipeline (adapter + OTLP exporter + HTTP/worker spans via `app/worker.go:79,86`). Framework repo never assembles `Trace` itself — wired only via the product scaffold template (`templates/init/cmd_api_main.go.tmpl:273`), default `NoOpTracer`.
- **Defect:** two halves of one observability story coexist in one package without joining: a log line cannot be correlated to its trace; a trace cannot see its DB spans.
- **Consequence:** production incident triage requires manual request-id cross-referencing; slow-query attribution inside a request trace is impossible.
- **Why insufficient:** fails §15.A "operational evidence" — the SLO/debugging workflow the observability layer exists for doesn't function end-to-end.
- **Fix (FBL-06, three tasks):** **T1** — extend the `Span` port with `TraceID() string` / `SpanID() string` (empty for no-op; otel adapter returns `SpanContext().TraceID().String()`), keeping the port vendor-neutral. **T2** — a ctx-aware `slog.Handler` wrapper in `kernel/observability` that injects `trace_id`/`span_id` attrs when a recording span is in ctx, wired where `AccessLog` and handler loggers are assembled; `AccessLog` adds the same two fields. **T3** — a thin `pgx.QueryTracer` implementation in `kernel/database` consuming the existing observability `Tracer` port (~50 LOC: span per query, statement summary + rows-affected attrs), attached in `chainAfterConnect`-adjacent pool config; sampling inherits the parent span. **Decision D-08:** hand-rolled thin tracer via the existing port, **not** `otelpgx` — a third-party bridge would bind OTel vendor types into `kernel/database`, breaking the port discipline the adapters layer gets right.
- **Reuse tier:** stdlib + already-present deps + fuller use of existing tool (the port). `otelslog` bridge explicitly rejected for now: correlation needs only attrs, not OTLP log export.
- **Fail-first:** test asserting a log record emitted inside a traced request carries `trace_id` — fails today (attr absent); trace-fixture test asserting a DB span child — fails today.
- **Acceptance:** correlation attrs present under active span, absent (not empty-string noise) without; pgx spans appear in the trace tree; no-op tracer path allocation-neutral (bench).
- **Dependencies:** none — independent of AR-01/02. **Priority:** P1. **wowsociety:** additive; regenerated scaffold gains it, existing main.go backports optionally. **IDs:** FBL-06, D-08.

### CS-08 — Validation enforcement path *(direct; NEW finding FBL-08 — §H9/§I8 "Ready" was an overclaim)*
- **Current evidence:** `kernel/validation/validation.go:53-74` registers only `TagNameFunc`; zero custom validators repo-wide. `kernel/httpx/decode.go:52-67` `BindAndValidate[T]` is an **opt-in helper**; `kernel/httpx/router.go` has zero binding/validation references — nothing enforces that a mutating handler validates its DTO. A handler that skips `BindAndValidate` gets zero validation with no framework safety net.
- **Defect:** validation is present-but-discretionary — exactly the "relies on package presence rather than effective behaviour" failure mode this pass exists to catch. The review graded it A-/Ready by library choice, not by enforcement.
- **Consequence:** one forgotten helper call = an unvalidated write endpoint; undetectable by any current gate.
- **Fix (FBL-08):** enforcement at the `RouteMeta` seam (the boot-validated metadata that already exists): **T1** — add `RouteMeta.Request` (a DTO prototype or `Validate bool` + type token) for mutating verbs; boot-time check fails any POST/PUT/PATCH route whose meta declares no request contract (waiver field for genuinely body-less mutations, consistent with AR-04 T5's waiver mechanism). **T2** — a `BindAndValidate`-calling generic handler adaptor so declaring the type *is* wiring the validation (no dual bookkeeping). **T3** — crud/scaffold templates updated to the adaptor.
- **Reuse tier:** fuller use of existing tools (RouteMeta boot validation + validator/v10 already present); no new dep; modest justified custom (the adaptor).
- **Fail-first:** fixture route registering POST with no request contract boots today; fails at boot after T1.
- **Acceptance:** boot rejects undeclared mutating routes; adversarial test posts an invalid DTO to a generated route and gets 400 with field errors.
- **Dependencies:** coordinates with AR-03 (RouteMeta is a projection input) — build T1 compatibly, don't wait. **Priority:** P1 (security-adjacent). **wowsociety:** additive at first (boot check behind a profile flag for one version), then enforced; audit its handlers for missing `BindAndValidate` before flipping. **IDs:** FBL-08.

### CS-09 — HTTP transport hygiene *(direct; NEW finding FBL-09)*
- **Current evidence:** no `http.Server{}` literal in wowapi — construction lives in the **product scaffold template**; the generated `wowsociety/cmd/api/main.go:308-312` sets only `ReadHeaderTimeout`. `ReadTimeout`/`WriteTimeout`/`IdleTimeout` unset → Go's infinite defaults. Mitigations present: `BodyLimit` via `http.MaxBytesReader` (`kernel/httpx/edge.go:157-166`), per-request `http.TimeoutHandler`, full middleware chain (RequestID→Recover→Locale→SecureHeaders→CORS→Trace→metrics→AccessLog→RateLimit→BodyLimit→Timeout). No compression (acceptable; reverse-proxy concern).
- **Defect:** slow-write/idle-connection exhaustion (Slowloris-response-side) unmitigated; `http.TimeoutHandler` bounds handler time, not connection read/write time.
- **Fix (FBL-09):** scaffold template sets all four timeouts from `config` (new `HTTP.ReadTimeout` etc. keys with safe defaults: read 30s, write 60s, idle 120s, header 10s — final values a config decision, defaults fail-safe); `config.Validate` rejects zero values in prod profile (same pattern as the SSRF-disable prod rejection at `kernel/config/config.go:261-263`). Also fold gosec's G120 (`kernel/httpx/csrf.go:118` unbounded `r.FormValue`) here: CSRF middleware must apply `MaxBytesReader` defensively since its chain position is app-controlled.
- **Reuse tier:** stdlib + existing config machinery.
- **Fail-first:** template-render test asserting all four timeouts present; prod-profile config with zero timeout fails `Validate`.
- **Priority:** P1. **wowsociety:** backport four lines to its committed main.go (template fix doesn't retro-apply — same delivery model as DX-07 T1). **IDs:** FBL-09.

### CS-10 — Data access & pgx resource contract *(direct; FBL-05 — diagnosis refined by evidence)*
- **Current evidence:** all 26 production `.Query(` sites across 15 files close rows and check `rows.Err()` — **zero violations today** (evidence pass, corroborated by `sqlclosecheck`/`rowserrcheck` actually run: 0 hits each). `txmanager.go:165,181` return raw `pgx.Rows`/`pgx.Row` (caller-owned close) — idiomatic `database/sql`-shaped contract. Pool config: only `MaxConns` (default 16, range 2–200, `kernel/config/config.go:99,211-212`); `MaxConnLifetime`/`MaxConnIdleTime` left at pgx defaults. No `QueryTracer` (→ CS-05 T3).
- **Refined verdict:** the "pgx-leak caveat" was directionally right about the *contract* but wrong about the *state* — current code is exemplary. FBL-05 is therefore a **regression guard**, not a corrective fix: enable `sqlclosecheck`+`rowserrcheck`+`bodyclose`+`noctx` while they're at zero/near-zero cost, so the contract is machine-enforced before drift, not after.
- **Fix:** (1) FBL-05 linter enablement (fixing noctx's 2 prod hits: `internal/cli/config_delegate.go:34`, `lint_cmd.go:129` — pass ctx to `exec.CommandContext`); (2) expose `MaxConnLifetime`/`MaxConnIdleTime` as config keys with pgx-default defaults (long-lived-connection credential-rotation and LB-rebalance hygiene); (3) keep raw `pgx.Rows` public contract — **decided, closed** (wrapper types rejected: reinventing `database/sql`'s contract for no benefit).
- **Fail-first:** for (1) a fixture with an unclosed-rows diff fails the enabled linter; for (2) config round-trip test.
- **Priority:** P1 (cheap, preventive). **wowsociety:** none (lint config is wowapi-internal; new config keys optional). **IDs:** FBL-05.

### CS-12 — Resilience primitives *(plan-anchored + §K)*
- **Current evidence:** breaker: `kernel/webhook/breaker.go` per-endpoint registry, injectable clock. Rate limiter: `kernel/httpx/ratelimit.go:205-225` in-memory map+mutex, single-node **by documented design** (`:26-27` frames Redis as a future adapter) — PERF-01's fix verified at HEAD. Retry: hand-rolled twice; `cenkalti/backoff/v5` + `sethvargo/go-retry` both indirect in `go.mod:27,53`, unused.
- **Fix:** FBL-04 (adopt `cenkalti/backoff/v5`, retry-schedule parity + fault-injection tests); breaker→`sony/gobreaker` stays P2-evaluate per §K; limiter retained as-designed.
- **Priority:** P1 (FBL-04). **IDs:** FBL-04, §K rows.

### CS-13 — Test infrastructure *(direct; T-TEST-01 re-scoped — original diagnosis corrected)*
- **Current evidence:** testkit provides **real per-test DB isolation**: `testkit/db.go:83-144,313` clones a per-test database via `CREATE DATABASE ... TEMPLATE` from a content-hashed migrated template, dropped in `t.Cleanup`. Fake clock (`testkit/fakes/clock.go:15-35`), RLS/authz asserts, module-contract runner, `WorkflowSim` DSL. Absent: fault-injection helpers, general HTTP harness. `internal/e2e` skips are toolchain/offline-related.
- **Correction:** the review's T-TEST-01 diagnosis ("shared-DB concurrency flake") is **unsubstantiated** — the isolation mechanism the diagnosis assumed missing exists. The observed fact stands (one full-suite `internal/e2e` failure that passed 4/4 isolated), the cause attribution does not.
- **Re-scoped T-TEST-01:** (1) reproduce under `-count` + parallel full-suite; (2) determine whether `internal/e2e` actually uses `testkit.NewDB` cloning or its own DB wiring; (3) fix what the reproduction shows — do not pre-commit to a mechanism. Fail-first = the reproduction run itself.
- **Additional gaps (fold into FBL-07):** hosted fuzzing never runs — CI replays seed corpus only (`ci.yml:98-101`, `-run '^Fuzz'`, no `-fuzz=`), `make test-fuzz` exists un-wired; = plan REL-04 T8/PERF-06 with exact citation now. Pre-push hook lets DB tests silently self-skip locally (no `WOWAPI_REQUIRE_DB`) and omits `-race` — hooks are a strict CI subset, acceptable, but the silent DB skip contradicts the repo's own skip-hygiene stance; one-line fix.
- **Priority:** P2. **IDs:** T-TEST-01 (re-scoped), REL-04-T8 (xref), FBL-07 (hook item).

### CS-17 — Authz cache bounding & invalidation *(direct + plan SEC-04)*
- **Current evidence:** `kernel/authz/caching.go:29-36` plain `map[string]cachedAssignments`+mutex, unbounded, no LRU; key `tenantID+"|c:/u:/s:"+id` (:57-66); TTL default **1s** (:44-53). `Invalidate`/`InvalidateTenant`/`InvalidateAll` exist (:93-121) but have exactly **one** production caller repo-wide (`kernel/seeds/seeds.go:278`, `InvalidateAll`); `kernel/kernel.go:118-121` documents grant/revoke invalidation as a *product-owned obligation*.
- **Defect:** unbounded growth under tenant×principal cardinality; correctness-by-convention invalidation — the framework performs the mutation (grant/revoke paths) but delegates the cache consequence to the product's memory.
- **Fix:** plan SEC-04 + review D-06, now concretised: **T1** replace map with `hashicorp/golang-lru/v2` (approved §L) sized by config; **T2** per-tenant epoch column (D-06) checked on read — framework-side mutation paths bump the epoch **in the same tx**, making invalidation structural, not conventional; `Invalidate*` methods stay for product-triggered cases. TTL floor stays as backstop.
- **Fail-first:** grant-revoke-then-check test currently serves the stale allow within TTL with no invalidation call — after T2 it observes the epoch bump immediately.
- **Dependencies:** T1 (LRU swap) is independent, land any time; T2's epoch bumps must be added to the framework mutation paths that exist **today** (role/permission assignment writes in `kernel/authz`, seeds) and extended to SEC-01's grant table **when it lands** — T2 does not wait on SEC-01, but SEC-01's new mutation paths must adopt the epoch bump as part of their own acceptance (cross-CS sequencing note added at the gate's request).
- **Priority:** P1 (P0 if cache enabled in prod). **wowsociety:** removes an undocumented obligation — strictly safer. **IDs:** SEC-04, D-06.

### CS-19 — i18n runtime behaviour *(verify-retain — CONFIRMED)*
- **Evidence:** layered load (`kernel/i18n/embed.go:11-28`, `catalog.go:29-41`); real freeze seal (`catalog.go:98-103`, post-freeze `Add` no-op :82-92); missing key falls back exact→default-locale→**echoes the key**, documented never-erroring (:108-134). Retained per §K; verdict earned, not presumed. **No task.**

### CS-22 — Documentation gates *(direct + plan-anchored; the §I20 "named without a how" row, now specified)*
- **Current evidence:** documentation drift has already happened twice at reviewer-visible severity: `README.md:148-153`/blueprint 11 described phantom `RunAPI`/`RunWorker`/`RunMigrate` APIs, and blueprint 06 listed five `Context` methods that don't exist (both fixed by AR-05 T1/T2 at `345e4ce`, verified §D). Nothing prevents recurrence: zero `//go:generate` directives repo-wide, no generated-code-currency check, no doc-example compile gate in any workflow or Makefile (toolchain inventory, all zero-hit-verified).
- **Defect:** normative Go examples in `docs/blueprint/*.md` and `README.md` are prose — they can silently rot against the live API, and did.
- **Consequence:** a consumer following the documented API writes code that doesn't compile; reviewer trust in all docs drops to zero after the first phantom API.
- **Why insufficient:** fails §15.A "docs" + "examples" criteria — examples exist but nothing proves they work (the "artifact-doesn't-actually-work" class).
- **Target state / Fix (AR-05 T3, mechanics):** a small extractor tool (`internal/tools/docexamples`) that scans the normative doc set for fenced ` ```go ` blocks tagged normative (an HTML comment marker above the fence, e.g. `<!-- doc-example: compile -->`, so illustrative pseudo-code opts out explicitly), writes each into a generated throwaway package, and `go build`s them; wired as a CI step in the `unit` job and a `make docs-check` target. Adversarial fixture: a deliberately staled example (calling a removed symbol) must fail the gate.
- **Reuse tier:** justified small custom tool (~150 LOC) on stdlib (`go/parser` not even needed — build failure is the check); no new dependency.
- **Fail-first:** run the extractor against a fixture doc referencing a phantom API (e.g. the pre-AR-05 `RunAPI` text, resurrectable from git history) — gate fails; current corrected docs pass.
- **Acceptance:** every tagged normative example compiles in CI; the staled-example fixture fails; `make docs-check` exists and CI calls it.
- **Dependencies:** none for T3 (AR-05 T4 generated-docs waits on AR-03). **Priority:** P2. **wowsociety:** none (wowapi docs only); pattern reusable there later. **IDs:** AR-05 T3 (this spec), T4/T5 (follow AR-03).

### CS-23 — Static-analysis & CI-gate utilisation *(direct; FBL-05/07 — the new axis, full inventory)*
- **Toolchain evidence (all personally spot-verified or worker-run with counts):** golangci-lint **v2.11.4** pinned (`Makefile:16`, `ci.yml:36`); config enables standard + 4. All 25 queried analyzers ship unenabled. Actual runs: **zero-hit set** (enable free): `sqlclosecheck, rowserrcheck, bodyclose, wastedassign, makezero, musttag, testifylint`. **Near-zero:** `noctx` 2 prod hits (CLI exec without ctx), `copyloopvar` 1 prod (dead pre-1.22 idiom, `app/maintenance.go:148`), `gocritic` `exitAfterDefer` (`internal/tools/migrate/main.go:49`). **Adjudicated by Fable 5:** `nilerr`'s `kernel/policy/policy.go:166` is deliberate fail-closed (unparseable runtime value → condition false; malformed *policy* errors at :161) — annotate, don't "fix"; `exhaustive`'s workflow hits (`definition.go:313`, `runtime.go:170`) are covered by fail-closed `default:` arms — annotate; `errorlint`'s `httpx/middleware.go:54` compares a recovered panic value to `http.ErrAbortHandler`, which net/http documents as a panicked sentinel — `==` is defensible, `errors.Is` harmless to adopt. **gosec (38):** triage list = G704 JWKS fetch taint (`kernel/auth/jwks.go:204,210` — governed by SEC-06's trusted-issuer config; annotate with that justification), G120 unbounded form parse (`csrf.go:118` → fixed in FBL-09), G115 int-overflow set (audit/database/jobs/mfa/pagination — review each conversion, most are bounded by prior validation; annotate or bound), G304 buildinfo file read (tool-only). `forcetypeassert`: `jwks.go:112` + `config/bind.go:150` — add checked assertions.
- **CI evidence:** all 7 workflows blocking where active; zero `continue-on-error`. Dormant-by-visibility: CodeQL (`codeql.yml:44`), Scorecard (`scorecard.yml:39`), dependency-review + its `license-check: true` (`security-scan.yml:80,93`) — **license scanning currently fully inert** (Trivy's `license` scanner also not enabled, `:71`). Trivy soft-fails by design (`:75`) = plan REL-02's exact scope, now with line citations. Absent: `go mod verify` (zero hits anywhere), ci.yml schedule, hosted mutation fuzzing (CS-13). Makefile↔CI: `test-fuzz`/drills/`goreleaser-check` never CI-called; named `test-security`/`test-contract` subsets execute inside the full gate but are never independently reported (acceptable; note only). Dependabot weekly with 7-day cooldown — fine.
- **Dependency utilisation:** direct-dep single-file imports are all deliberate documented chokepoints — **no gap**. `golang.org/x/sync` unused; hand-rolled errgroup-shaped drain at `app/worker.go:115-138` **retained deliberately** (bounded-drain semantics exceed errgroup; documented ARCH-57); `kernel/jobs/runner.go:357-367` fan-in could adopt errgroup but aggregates no errors — noted, not adopted. `testify` transitive-only, unused — consistent hand-rolled asserts are a house style, no action. Earlier "hashicorp libs transitive" observation was contamination from golangci-lint's own module closure — corrected; `golang-lru/v2` remains a **new** approved dep (§L stands).
- **Fix:** **FBL-05** (zero-cost set + noctx/copyloopvar fixes — one PR); **FBL-07** (judged set: `gosec` with the triage list above + inline `#nosec` justifications, `errorlint`, `exhaustive` w/ annotations, `forcetypeassert`, `usestdlibvars`; plus `go mod verify` step in ci.yml, Trivy license-scanner enablement or `go-licenses` as the interim private-repo license signal, scheduled nightly ci.yml run incl. `-fuzz` per REL-04 T8, pre-push DB-skip fix). `wrapcheck`/`revive` **rejected** (≈50 hits each, noise-dominant without heavy tuning; staticcheck+errorlint cover the real classes).
- **Fail-first:** each enablement run is its own failing-state artifact.
- **Priority:** FBL-05 P1 (one sitting), FBL-07 P1/P2 staged. **wowsociety:** none directly; recommend mirroring the final linter set. **IDs:** FBL-05, FBL-07, REL-02 (xref), REL-04-T8 (xref).

### CS-24 — Outbound-HTTP SSRF guard *(verify — hypothesis REFUTED, strength confirmed)*
- **Evidence:** guard is **dial-time**, not parse-time: `kernel/httpclient/client.go:84-87,177-209` installs a custom `DialContext` that resolves, checks the resolved IPs, and dials the verified IP directly — closing DNS-rebinding TOCTOU; each redirect hop re-enters the dialer; `Proxy=nil` closes env-proxy bypass (`:70-83`); blocked classes include IPv6-embedded-v4 unwrapping (`guard.go:60-73,97-124`); the disable flag is rejected in prod (`config.go:261-263`). **No task** — this is a §15.A-grade implementation and is recorded as such. gosec's G704 on the JWKS path is annotation work (CS-23), not a gap.

### CS-25 — Secrets lifecycle *(direct; decided posture)*
- **Evidence:** single provider `adapters/secrets/envprovider` (`secretref://env/VAR`); resolution once at boot (`kernel/secrets/secrets.go:47`); no rotation; `config.Secret` redaction is comprehensive (String/GoString/Format/JSON/Text/LogValue all covered; `Reveal()` only escape; type-level `secretref://` enforcement).
- **Fable 5 decision (D-09):** boot-time-once + restart-based rotation is an **acceptable, explicitly documented v1 contract** — most orchestrators roll pods on secret change; hot-reload plumbing through every consumer (pgx pool, JWKS client, S3 creds) is real complexity with modest v1 payoff. Tasks: **T1** document the contract + rotation runbook (restart-triggering); **T2 (P2)** provider seam already exists — a file-based provider (K8s mounted-secret pattern) is the next increment when needed, *not* a vault client in the kernel.
- **Priority:** P2. **IDs:** D-09, FBL-03-adjacent doc task.

### CS-03 / CS-04 / CS-06 — verify-Ready outcomes
- **CS-03 Config: CONFIRMED Ready.** Fail-closed behaviours are real returned errors (`load.go:132-139`, `config.go:254-264` via `errors.Join`); fingerprint SHA-256 over canonical JSON with Secret-redacted marshal (`fingerprint.go:18,29-35`). Documented limitation (rotation without ref-change doesn't change the fingerprint) is inherent to the (correct) redacted-hash design. No task.
- **CS-04 Errors: Ready with one hygiene task.** kerr structure solid; `errorlint`/`nilerr` adjudications in CS-23 found no real defect — enablement + annotations only.
- **CS-06 Health:** folded into CS-21 (same seam).

## 2.1 Closure-evidence register (one canonical proving artifact per task-bearing CS)

Evidence root convention follows the plan's: `docs/implementation/evidence/premier/<ID>/`.

| CS | Closure evidence (the artifact that proves it) |
|---|---|
| CS-01 | Extended depguard+boundaries lint output (fail-before/pass-after pair) + wowsociety identity-suite log on new mfa path → `evidence/premier/FBL-01/` |
| CS-02 | Plan §5.1's per-task evidence paths verbatim (`AR-01/lifecycle_test_output.txt` … `AR-02/legacy_port_adapter_compat_test_output.txt`) |
| CS-03 | n/a — verify-Ready; the citations in the CS body are the record |
| CS-04 | Lint-enablement run log (errorlint/nilerr annotations reviewed) → `evidence/premier/FBL-07/` |
| CS-05 | Correlation test output (trace_id attr present/absent matrix) + exported trace tree showing pgx child spans → `evidence/premier/FBL-06/` |
| CS-06 | folded → CS-21 |
| CS-07 | Plan §5.2 SEC-01 evidence paths (`SEC-01/membership-tests.md` etc.) |
| CS-08 | Boot-rejection test output (undeclared mutating route) + adversarial invalid-DTO 400 test → `evidence/premier/FBL-08/` |
| CS-09 | Template-render assertion + prod-profile zero-timeout `Validate` rejection test → `evidence/premier/FBL-09/` |
| CS-10 | sqlclosecheck/rowserrcheck/bodyclose/noctx enablement logs (0 issues at HEAD) + pool-config round-trip test → `evidence/premier/FBL-05/` |
| CS-11 | Plan chaos-test artifacts (`DATA-02/chaos/duplicate_worker_lease_expiry_test.go` output, `DATA-03/chaos/`, `DATA-04/chaos/`) |
| CS-12 | Retry-schedule parity + fault-injection test outputs → `evidence/premier/FBL-04/` |
| CS-13 | The reproduction-run artifact (`-count`+parallel full suite) and resulting diagnosis note → `evidence/premier/T-TEST-01/` |
| CS-14 | Generator-output-boots test log (fail at HEAD, pass after) → plan DX-02 evidence path |
| CS-15 | DX-06 per-field merge fixtures + REL-03a CI job logs (apidiff seeded-breaking fixture) |
| CS-16 | `bench-budgets.txt` diff (7 new hot-path packages) + PERF-02..05 relative-run outputs → `evidence/premier/PF-PERF/` |
| CS-17 | LRU-bound eviction test + epoch-bump-observed invalidation test → `evidence/premier/PF-SEC/SEC-04/` |
| CS-18 | Plan DATA-01 evidence paths (`DATA-01/cross-tenant-fk-negative/` etc.) |
| CS-19 | n/a — verify-retain; citations in the CS body are the record |
| CS-20 | Per-field tamper-matrix output (every field independently breaks verification) → `DATA-08/wave6/audit-hash/` |
| CS-21 | Stale-migration readiness-503 test + prod empty-catalog boot log (named failure → seeded pass) → `evidence/premier/FBL-02/`, `DX-07/` |
| CS-22 | `make docs-check` CI log + staled-example fixture failure → `AR-05/doc_compile_ci_gate_test_output.txt` |
| CS-23 | Per-linter enablement run logs, `go mod verify` CI step output, nightly-fuzz job log → `evidence/premier/FBL-05/`, `FBL-07/` |
| CS-24 | n/a — verified strength; citations in the CS body are the record |
| CS-25 | The published rotation-contract doc + restart-rotation runbook (D-09) → `evidence/premier/D-09/` |

## 3. Fable 5 adjudication log for this pass (evidence workers vs. personal verification)
- **OVERTURNED a worker refutation — DX-02/PF-2 is REAL at HEAD:** the evidence worker declared "no verb input exists → NO-GAP"; personally verified `internal/cli/templates/crud/resource.go.tmpl:54` emits `{{.PermPrefix}}.delete` while `kernel/authz/registry.go:15-19`'s closed verb set contains no `delete` — generated modules still fail boot exactly as wowsociety PF-2 (`docs/upstream/06-pf-2-...md`) reported at v1.0.0. The worker searched for a CLI parameter; the defect is the emitted permission. CS-14/DX-02 stands, now with the precise two-sided citation and the fix pinned: emit `deactivate` (matches the template's soft-delete TODO at `resource.go.tmpl:146`) + a generator-output-boots CI test.
- **ACCEPTED a worker correction — T-TEST-01 diagnosis was wrong:** testkit's template-clone isolation exists (`testkit/db.go:83-144`); "shared-DB concurrency" was asserted without this check. Re-scoped in CS-13.
- **ACCEPTED — validation-enforcement gap (FBL-08):** new material finding; §H9/§I8 downgraded.
- **REJECTED two "potential bug" flags after personal reads:** `policy.go:166` (deliberate fail-closed) and the `exhaustive` workflow hits (fail-closed defaults) — see CS-23.
- **ACCEPTED with correction — module-graph contamination:** the "hashicorp transitive" observation came from golangci-lint's own closure; verified via `go mod graph` framing. cenkalti/sethvargo remain genuinely in `go.mod` (:27,:53).
