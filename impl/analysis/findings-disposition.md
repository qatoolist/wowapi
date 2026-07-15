---
id: ANALYSIS-FINDINGS-DISPOSITION
type: analysis
title: Findings disposition register - PLAN/REVIEW/MATRIX items with severity, current state, resolution, and verification requirement
status: complete
created_at: 2026-07-12
updated_at: 2026-07-12
derived: false
---

# Findings disposition

Per mandate S11.3. One row per item in `requirement-inventory.md` tables A (38 PLAN findings), B (15
REVIEW findings/decisions), and C (9 MATRIX verify-outcomes/constraints) - 62 rows, all accounted
for, none sampled or abbreviated. Fields: finding ID, source, severity, description, current-state,
proposed-resolution, disposition, implementation-location, verification-requirement, closure-status.

Derivation rules applied uniformly:
- Severity: Pri P0 -> Critical, P1 -> High, P2 -> Medium, P3 -> Low; QG-classed items keep their class label.
- Current-state: `planned` -> "not started"; `partial`/`INV` -> "partially executed - see PLAN S8/S9 for
  exact slice"; `blocked` -> "blocked pending human decision or upstream dependency"; MATRIX
  verify-outcomes -> "already verified" or "not-applicable - justified retention" per their own
  disposition.
- Verification-requirement: the 8 executed-but-unverified slices (AR-04, AR-05, AR-06, SEC-02,
  PERF-01, PERF-06, DATA-08 W0-slice, REL-04 T1-T4) all point to W00 re-verification; fully-planned
  items point to their future story's own verification.md; already-verified MATRIX rows note
  re-verification is not required absent code changes.
- Closure-status: `open` for planned/partial/blocked items; `closed-pending-programme-acceptance` for
  not-applicable/rejected/superseded/deferred items - nothing is closed outright since no work has
  executed under this programme yet.

## A. Plan findings (PLAN S5 - 38 rows)

| ID | Source | Severity | Description | Current-state | Proposed-resolution | Disposition | Implementation-location | Verification-requirement | Closure-status |
|---|---|---|---|---|---|---|---|---|---|
| AR-01 | PLAN S5 PF-ARCH | High P1 | Ownership-bound ApplicationModel T1-T11 | not started | W05-E01-S001..S004; D-02/D-03 resolve open questions | planned | W05-E01-S001..S004 | verification per story's own verification.md once implemented | open |
| AR-02 | PLAN S5 PF-ARCH | High P1 | Typed port keys + compiled provider graph T1-T7 | not started | W05-E02-S001..S003; depends AR-01 T1/T2 | planned | W05-E02-S001..S003 | verification per story's own verification.md once implemented | open |
| AR-03 | PLAN S5 PF-ARCH | High P1 | One authoritative declaration, derived projections T1-T5 | not started | W05-E03-S001..S002; T2 = DX-06 duplicate, single owner DX-06 (CONFLICT-01) | planned | W05-E03-S001..S002 | verification per story's own verification.md once implemented | open |
| AR-04 | PLAN S5 PF-ARCH | High P1 | Eliminate boot-time silent behaviour | partially executed - see PLAN S8/S9 for exact slice; T1 EXECUTED verified x2 | T2-T5 planned, dep AR-01; T5 waiver shared w/ SEC-06/DX-07 (CONFLICT-07) | partial | W05-E03-S002 | W00 baseline-and-verification wave - re-verify at pinned HEAD per requirement-inventory.md disposition INV/partial | open |
| AR-05 | PLAN S5 PF-ARCH | High P1, DOC/QG | Composition/doc drift removal | partially executed - see PLAN S8/S9 for exact slice; T1/T2 EXECUTED | T3 doc-example CI gate CS-22 spec, T4/T5 dep AR-03 | partial | W06-E04-S002 | W00 baseline-and-verification wave - re-verify at pinned HEAD per requirement-inventory.md disposition INV/partial | open |
| AR-06 | PLAN S5 PF-ARCH | High P1 | Remove hidden constructor bypasses | partially executed - see PLAN S8/S9 for exact slice; T1 EXECUTED | T2 lint + T3 audit planned | partial | W05-E04-S001 | W00 baseline-and-verification wave - re-verify at pinned HEAD per requirement-inventory.md disposition INV/partial | open |
| SEC-01 | PLAN S5 PF-SEC | Critical P0 | Server-side tenant/privileged session state T1-T7 | not started | W03-E01-S001..S004; D-01 ratified; DEC-Q1 safe default unblocks; BREAKING wowsociety | planned | W03-E01-S001..S004 | verification per story's own verification.md once implemented | open |
| SEC-02 | PLAN S5 PF-SEC | Critical P0 | Workflow privileged ops fail closed | partially executed - see PLAN S8/S9 for exact slice; T1-T3 EXECUTED verified x2 | T4 ratification design + T5 audit remain | partial | W03-E05-S001 | W00 baseline-and-verification wave - re-verify at pinned HEAD per requirement-inventory.md disposition INV/partial | open |
| SEC-03 | PLAN S5 PF-SEC | High P1 | Webhook replay bound to authenticated data T1-T4 | not started | W03-E03-S001; breaking Verifier interface | planned | W03-E03-S001 | verification per story's own verification.md once implemented | open |
| SEC-04 | PLAN S5 PF-SEC | High P1 | Bound authz staleness/memory | not started | W05-E04-S002; CS-17: LRU approved dep + epoch table D-06; P0 if cache prod-enabled | planned | W05-E04-S002 | verification per story's own verification.md once implemented | open |
| SEC-05 | PLAN S5 PF-SEC | High P1, QG | Versioned security verification profile | not started | W07-E02-S001; Wave-6-class per plan | planned | W07-E02-S001 | verification per story's own verification.md once implemented | open |
| SEC-06 | PLAN S5 PF-SEC | High P1 | Outbound-security escape-hatch governance | not started | W03-E02-S001; D-07 ratified; T5 waiver shared w/ AR-04/DX-07 (CONFLICT-07) | planned | W03-E02-S001 | verification per story's own verification.md once implemented | open |
| DATA-01 | PLAN S5 PF-DATA | Critical P0 | Composite tenant FKs T1-T8 | not started | W02-E02-S001..S002; dep DATA-09 T1-T5 for risky steps; wowsociety has own instance PROD-01 | planned | W02-E02-S001..S002 | verification per story's own verification.md once implemented | open |
| DATA-02 | PLAN S5 PF-DATA | Critical P0 | Lease generations/fencing + idempotency T1-T7 | not started | W04-E01-S001..S003; T1 shared primitive is keystone | planned | W04-E01-S001..S003 | verification per story's own verification.md once implemented | open |
| DATA-03 | PLAN S5 PF-DATA | Critical P0 | Remote I/O outside DB transactions T1-T8 | not started | W04-E02-S001..S002; scope refined by MATRIX CS-11: external effects only; T7 = DATA-08 W0-T2 duplicate done (CONFLICT-03) | planned | W04-E02-S001..S002 | verification per story's own verification.md once implemented | open |
| DATA-04 | PLAN S5 PF-DATA | High P1 | Bulk multi-worker safety T1-T6 | not started | W04-E03-S001; T1 stopgap can land early in-wave | planned | W04-E03-S001 | verification per story's own verification.md once implemented | open |
| DATA-05 | PLAN S5 PF-DATA | High P1 | Version allocation races + blob GC T1-T5 | not started | W02-E03-S001 | planned | W02-E03-S001 | verification per story's own verification.md once implemented | open |
| DATA-06 | PLAN S5 PF-DATA | High P1 | Resource-mirror aggregate write contract T1-T4 | not started | W02-E04-S001; T2 shared fix w/ DATA-07 T3, one owner (CONFLICT-06) | planned | W02-E04-S001 | verification per story's own verification.md once implemented | open |
| DATA-07 | PLAN S5 PF-DATA | High P1 | Relationship semantics + actor attribution T1-T4 | blocked pending human decision or upstream dependency; HARD dep SEC-01 | W03-E04-S001; secondary dep SEC-04 | blocked-then-planned | W03-E04-S001 | verification per story's own verification.md once implemented | open |
| DATA-08 | PLAN S5 PF-DATA | Critical/High P0/P1 | Compliance evidence complete/durable | partially executed - see PLAN S8/S9 for exact slice; W0 slice EXECUTED verified x2 | W6-T1 hash widening D-04 + T2-T5 planned | partial | W04-E04-S001..S002 | W00 baseline-and-verification wave - re-verify at pinned HEAD per requirement-inventory.md disposition INV/partial | open |
| DATA-09 | PLAN S5 PF-DATA | Critical P0 | Online expand/backfill/validate/contract protocol T1-T9 | not started | W02-E01-S001..S003; precedes DATA-01 T4/T5 + DATA-08 W6-T1; T9 CI drills | planned | W02-E01-S001..S003 | verification per story's own verification.md once implemented | open |
| DX-01 | PLAN S5 PF-DX | Critical P0 | Source-built CLI path validity incl. T5 scaffold harness | not started | W01-E04-S001; T5 harness = shared primitive for DX-02/DX-04 | planned | W01-E04-S001 | verification per story's own verification.md once implemented | open |
| DX-02 | PLAN S5 PF-DX | Critical P0 | Generator emits in-set verb + boots PF-2 | not started | W01-E04-S001; one-token fix + generator-boots test CS-14, re-verified at HEAD | planned | W01-E04-S001 | verification per story's own verification.md once implemented | open |
| DX-03 | PLAN S5 PF-DX | High P1, ARCH/FUT | Module DSL design | not started | W06-E01-S001; design-investigation story only Wave-4-class per plan | deferred | W06-E01-S001 | verification per story's own verification.md once implemented | closed-pending-programme-acceptance |
| DX-04 | PLAN S5 PF-DX | High P1 | Golden consumer + upgrade matrix | implemented, verified, and accepted 2026-07-14 | W06-E01-S002; installed CLI, all 8 subsystems across 2 modules, real infrastructure, tagged v1.1.0→local candidate replay, Wave-4 required gate | accepted | W06-E01-S002 | EV-W06-E01-S002-003/004/005/007/010/012/014 | closed |
| DX-05 | PLAN S5 PF-DX | High P1, DOC | CLI/docs/version identity singular | partially executed - see PLAN S8/S9 for exact slice; T1/T2 EXECUTED | S6-vs-S9 status inconsistency = T-DOC-01 (CONFLICT-04) | partial | W01-E04-S002 | W00 baseline-and-verification wave - re-verify at pinned HEAD per requirement-inventory.md disposition INV/partial | open |
| DX-06 | PLAN S5 PF-DX | High P1 | OpenAPI merge complete-or-loud T1-T3 | not started | W06-E02-S001; single owner of AR-03 T2 scope (CONFLICT-01); validator dep decision at impl | planned | W06-E02-S001 | verification per story's own verification.md once implemented | open |
| DX-07 | PLAN S5 PF-DX | High P1 | Truthful readiness/config diagnostics T1-T4 | not started | W04-E04-S003; T4 dep AR-04 T5 waiver mechanism (CONFLICT-07) | planned | W04-E04-S003 | verification per story's own verification.md once implemented | open |
| PERF-01 | PLAN S5 PF-PERF | Critical P0 | Token-bucket sweep fix | already verified; EXECUTED + #25 recalibrated sweep budgets to honest full-map measurements | verify at current HEAD against post-#25 budgets (CONFLICT-05) | INV | W00-E01-S002 | W00 baseline-and-verification wave - re-verify at pinned HEAD per requirement-inventory.md disposition INV/partial | open |
| PERF-02 | PLAN S5 PF-PERF | High P1 | Complete-request benchmarks vs real PG | implemented, verified, and story-accepted 2026-07-14 | Six real-PG profiles; 36-cell matrix; six-part attribution; relative/container publication; absolute SLO conditional on DEC-Q9 | accepted | W07-E01-S001 | EV-W07-E01-S001-001..005 plus clean independent review | closed |
| PERF-03 | PLAN S5 PF-PERF | High P1 | Rules resolution bounded SQL | implemented, verified, and story-accepted 2026-07-14 | Set-based precedence parity; current/history index access; SQL count 8/8/8 at depths 3/10/50; live updates visible | accepted | W07-E01-S002 | EV-W07-E01-S002-001..007 plus clean independent review | closed |
| PERF-04 | PLAN S5 PF-PERF | High P1 | Sweeper/worker N+1 + unbounded materialization | implemented, verified, and story-accepted 2026-07-14 | Bounded materialization, fixed/batched queries, W04 lease/fencing reuse, chaos/ordering proof, relative publication | accepted | W07-E01-S003 | EV-W07-E01-S003-001..008 plus clean independent review | closed |
| PERF-05 | PLAN S5 PF-PERF | Medium P2 | Explicit object checksum behaviour | implemented, verified, and story-accepted 2026-07-14 | Required checksums; labeled bounded repair; metrics; interrupt/resume no-duplicate backfill; conditional publication | accepted | W07-E01-S004 | EV-W07-E01-S004-001..007 plus clean independent review | closed |
| PERF-06 | PLAN S5 PF-PERF | High P1, QG | Fail-closed performance gates | partially executed - see PLAN S8/S9 for exact slice; T1 EXECUTED | T3/T4 fuzz scope = REL-04 T8 single-owner (CONFLICT-02), W07-E02-S002 | INV | W00-E01-S002 | W00 baseline-and-verification wave - re-verify at pinned HEAD per requirement-inventory.md disposition INV/partial | open |
| REL-01 | PLAN S5 PF-REL | Critical P0 | Release gated on exact published commit T1-T9 | not started | W06-E03-S001..S002; ~85% buildable now; final activation = DEC-Q10 admin | planned | W06-E03-S001..S002 | verification per story's own verification.md once implemented | open |
| REL-02 | PLAN S5 PF-REL | Critical/High P0/P1, QG | Security checks blocking-or-replaced | not started | W06-E03-S003; Trivy soft-fail + visibility-guard dormancy cited at exact lines MATRIX CS-23 | planned | W06-E03-S003 | verification per story's own verification.md once implemented | open |
| REL-03 | PLAN S5 PF-REL | High P1, QG | Compatibility gates split a/b | not started | W06-E02-S002..S003; a=T1,T2,T4,T6,T8,T9 now; b=T3 DX-06, T5 AR-03/DX-03, T7 DX-04 | planned | W06-E02-S002..S003 | verification per story's own verification.md once implemented | open |
| REL-04 | PLAN S5 PF-REL | High P1, QG | Truthful integration coverage | partially executed - see PLAN S8/S9 for exact slice; T1-T4 EXECUTED verified x2 | T5-T8 planned; T8 owns fuzz, shared w/ PERF-06 T3/T4 (CONFLICT-02) | partial | W07-E02-S002 | W00 baseline-and-verification wave - re-verify at pinned HEAD per requirement-inventory.md disposition INV/partial | open |

A-table row count: 38 (27 planned, 8 partial/INV, 1 deferred, 1 blocked-then-planned, 1 accepted) -
matches `requirement-inventory.md`'s own stated totals.

## B. Review findings and decisions (REVIEW SO/SU - 15 rows)

| ID | Source | Severity | Description | Current-state | Proposed-resolution | Disposition | Implementation-location | Verification-requirement | Closure-status |
|---|---|---|---|---|---|---|---|---|---|
| FBL-01 | REVIEW SJ/SO/ST | High P1 | Kernel re-home 9 pkgs into foundation/ | not started | W05-E05-S001..S002; CS-01 mechanics; dep AR-01/02; wowsociety mfa migration is product-coordination PROD-02 | planned | W05-E05-S001..S002 | verification per story's own verification.md once implemented | open |
| FBL-02 | REVIEW SO/ST | Critical P0-prod | Production seed-sync path PF-9 | not started | W02-E05-S001; CS-21 acceptance bar fixed; design detail = story investigation task | planned | W02-E05-S001 | verification per story's own verification.md once implemented | open |
| FBL-03 | REVIEW SO | Medium P2, DOC | Reconcile wowsociety upstream register | not started | W01-E04-S002; mark PF-2/PF-6/RFF-001 etc closed when their fixes land | planned | W01-E04-S002 | verification per story's own verification.md once implemented | open |
| FBL-04 | REVIEW SO/SK | High P1 | Adopt cenkalti/backoff for duplicated retry | not started | W04-E02-S003; approved dep; parity + fault-injection tests | planned | W04-E02-S003 | verification per story's own verification.md once implemented | open |
| FBL-05 | REVIEW SO/SH, MATRIX CS-23 | High P1, QG | Enable zero-cost leak linters sqlclosecheck etc | not started | W01-E01-S001; counts measured at HEAD CS-23; noctx 2 prod fixes | planned | W01-E01-S001 | verification per story's own verification.md once implemented | open |
| FBL-06 | REVIEW SO/SU D-08, MATRIX CS-05 | High P1 | OTel trace/log correlation + pgx tracer D-08 | not started | W01-E02-S001..S002; CS-05 T1-T3 | planned | W01-E02-S001..S002 | verification per story's own verification.md once implemented | open |
| FBL-07 | REVIEW SO/SI | High/Medium P1/P2, QG | Utilisation closure: gosec triage, go mod verify, license signal, nightly fuzz, hook DB-skip | partially executed - see PLAN S8/S9 for exact slice; nightly CI schedule EXISTS since #24 SD-02 | fuzz portion still seed-replay only; remaining sub-items W01-E01-S002..S003 | partial | W01-E01-S002..S003 | verification per story's own verification.md once implemented | open |
| FBL-08 | REVIEW SH/SI, MATRIX CS-08 | High P1 | Central validation enforcement RouteMeta seam | not started | W01-E03-S002; CS-08 T1-T3; compat: profile-flag first | planned | W01-E03-S002 | verification per story's own verification.md once implemented | open |
| FBL-09 | REVIEW SH, MATRIX CS-09 | High P1 | HTTP server timeouts + CSRF body bound | not started | W01-E03-S001; CS-09; template-delivery model; wowsociety backport = PROD-03 | planned | W01-E03-S001 | verification per story's own verification.md once implemented | open |
| D-01..D-09 | REVIEW SU | n/a ARCH | Nine ratified architecture decisions | not started | W00-E02-S003; ADR-ification story; enacted inside their target stories; DEC register tracks | planned | W00-E02-S003 | verification per story's own verification.md once implemented | open |
| T-DOC-01 | requirement-inventory.md, derived from PLAN S6 vs S9 | Low P3, DOC | Fix plan S6-vs-S9 DX-05 status inconsistency | not started | W01-E04-S002 (CONFLICT-04) | planned | W01-E04-S002 | verification per story's own verification.md once implemented | open |
| T-TEST-01 | REVIEW SH, MATRIX S3 adjudication | Medium P2, VERIF | Diagnose intermittent e2e full-suite failure, re-scoped | not started | W01-E04-S003; reproduce-first; original shared-DB diagnosis withdrawn per MATRIX S3 testkit template-clone isolation exists | planned | W01-E04-S003 | verification per story's own verification.md once implemented | open |
| DEC-Q1 | REVIEW SF Q1 | Critical P0, ARCH | IdP grant_id claim contract | blocked pending human decision or upstream dependency | safe default unblocks build per REVIEW SF Q1; tracked at W03-E01, framework owns grant record, only claim-shape wiring gated | blocked-human | W03-E01 tracked | verification per story's own verification.md once implemented | open |
| DEC-Q9 | REVIEW SF Q9 | High P1, OPS | Reference-perf-env ownership | blocked pending human decision or upstream dependency | provisional default set GH runner + reference json; tracked at W07-E01, only absolute-SLO gating waits | blocked-human | W07-E01 tracked | verification per story's own verification.md once implemented | open |
| DEC-Q10 | REVIEW SF Q10/SG | Critical P0, OPS | Repo-admin activation: branch/tag/env protection | blocked pending human decision or upstream dependency | merge-queue rulesets unavailable on user-owned repo, session fact; tracked at W06-E03, only final activation gated | blocked-human | W06-E03 tracked | verification per story's own verification.md once implemented | open |

B-table row count: 15 (12 planned/partial, 3 human-blocked) - matches `requirement-inventory.md`'s own
stated totals.

## C. Matrix verify-outcomes and constraints (9 rows - no new work, recorded so nothing silently disappears)

| ID | Source | Severity | Description | Current-state | Proposed-resolution | Disposition | Implementation-location | Verification-requirement | Closure-status |
|---|---|---|---|---|---|---|---|---|---|
| CS-03 | MATRIX S2 Configuration | n/a VERIF | Config fail-closed + fingerprint | already verified | re-earned with citations in MATRIX; W00 pins evidence pointer | INV-then-verified | W00 evidence pointer pin | re-verification not required unless code changes - evidence pointer pinned in W00 | open, pending W00 pin |
| CS-19 | MATRIX S2 i18n | n/a VERIF | i18n freeze + key-echo fallback | already verified | same treatment as CS-03; W00 pins evidence pointer | INV-then-verified | W00 evidence pointer pin | re-verification not required unless code changes - evidence pointer pinned in W00 | open, pending W00 pin |
| CS-24 | MATRIX S2 SSRF, S1 dedup | n/a VERIF | SSRF dial-time guard | already verified; verified strength | gosec G704 annotation task inside FBL-07 | INV-then-verified | W00 evidence pointer pin + FBL-07 gosec annotation | re-verification not required unless code changes - evidence pointer pinned in W00 | open, pending W00 pin and FBL-07 annotation |
| CS-10 | MATRIX S2 pgx rows contract | n/a CONSTR | pgx rows contract | not started; constraint recorded, not yet mechanically enforced | decided: keep raw pgx.Rows; FBL-05 enforces mechanically; pool lifetime config keys = task in W01-E01-S001 | planned | W01-E01-S001 via FBL-05 | verification per story's own verification.md once implemented | open |
| CS-25 | MATRIX S2 Secrets, REVIEW SU D-09 | n/a DOC/OPS | Secrets rotation contract D-09 | not started | restart-based rotation documented; file-provider = deferred DEF-01 | planned | documentation task, target per D-09/W00-E02-S003 ADR | verification per story's own verification.md once implemented | open |
| K-RETAIN | REVIEW SK | n/a ARCH | SK retained customs: config, i18n, scheduler, rate-limiter, TOTP | not-applicable - justified retention | no work required; retentions justified in REVIEW SK/S29 answer 14 | not-applicable | n/a | re-verification not required unless code changes - evidence pointer pinned in W00 | closed-pending-programme-acceptance |
| K-P2 | REVIEW SK | Low FUT | gobreaker breaker + jwx JWKS evaluations | not started, deferred | DEF-02/DEF-03 with reopen triggers per REVIEW SK | deferred | tracking/deferred-items-register.md DEF-02/DEF-03 | verification per story's own verification.md once implemented | closed-pending-programme-acceptance |
| M-REJ | REVIEW SM | n/a REJ | Rejected deps: viper/envconfig, kernel message bus, custom crypto | not-applicable - justified retention, rejected not adopted | rationale in REVIEW SM; no further work | rejected | n/a | re-verification not required unless code changes - evidence pointer pinned in W00 | closed-pending-programme-acceptance |
| B11/B12/B13 | framework-backlog-p2-decisions.md, decisions.md D-0090 | Low FUT | Parked P2 backlog: radix router, schema unification, hot overlays | not started, deferred | DEF-04..06; reopen triggers in framework-backlog-p2-decisions.md, re-verified independently 2026-07-11 per D-0090 | deferred | tracking/deferred-items-register.md DEF-04/DEF-05/DEF-06 | re-verification not required unless code changes - evidence pointer pinned in W00 | closed-pending-programme-acceptance |

C-table row count: 9 (3 verified-INV, 2 planned, 4 deferred/rejected/not-applicable) - matches
`requirement-inventory.md`'s own stated totals.

Combined row count across A+B+C: 38 + 15 + 9 = 62, matching the required minimum. No row from
`requirement-inventory.md` was sampled, abbreviated, or dropped from this register.
