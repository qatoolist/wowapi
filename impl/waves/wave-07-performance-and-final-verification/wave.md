---
id: W07
type: wave
title: Performance and final verification
status: in-progress
owner: W07-Phase-A-Execution
reviewer: unassigned
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-14
included_epics:
  - W07-E01
  - W07-E02
  - W07-E03
  - W07-E04
depends_on:
  - W00
  - W01
  - W02
  - W03
  - W04
  - W05
  - W06
blocks: []
source_requirements:
  - PERF-02
  - PERF-03
  - PERF-04
  - PERF-05
  - PERF-06
  - CS-16
  - SEC-05
  - REL-04
  - DEC-Q9
---

# W07 — Performance and final verification

## Objective

Execute the PERF-02..05 performance-verification programme as relative/container benchmarking now, with
absolute-SLO acceptance explicitly conditional on DEC-Q9's reference-performance-environment decision;
expand hot-path benchmark coverage from 8/55 to include the 7 named unbenched packages per MATRIX CS-16;
establish the versioned security verification profile (SEC-05) as a closure gate over SEC-01/03/04/06's
own substantially-complete state; complete REL-04's remaining truthful-coverage work (fail-not-skip E2E,
skip manifest, race-integration schedule, real coverage-guided fuzz — also owning the identical PERF-06
T3/T4 fuzz scope per CONFLICT-02); verify the framework-side PROD-01..05 coordination artifacts exist
without implementing any wowsociety-repo change; and re-run the REVIEW §30-style final approval gate
across the whole 8-wave programme, producing both a closure report and an explicit, human-facing
production-readiness claim-upgrade decision package. This is the programme's final wave.

## Rationale

`impl/index.md`'s wave map assigns W07 "PERF-02..05 relative programme (+DEC-Q9), SEC-05 profile, REL-04
remainder (real fuzz), product-alignment verification, programme closure gate," depending on "all
prior." `requirement-inventory.md` confirms every PERF-02..05 row targets W07-E01, SEC-05 targets
W07-E02-S001, REL-04 (T5-T8) targets W07-E02-S002, and the two closure epics (E03, E04) exist nowhere
else in the programme — this is genuinely the last wave, structurally incapable of running earlier,
because SEC-05's own PLAN dependency is explicit ("SEC-01–04 substantially complete") and the closure
gate's own purpose (re-running REVIEW §30's gate across the whole programme) is definitionally only
meaningful once every other wave has executed. REVIEW §12's own framing — cited by MATRIX CS-16 —
resolves what would otherwise look like a hard block on the entire performance programme: "Relative/
container now (REVIEW §12); absolute SLO gated on DEC-Q9" is the load-bearing acceptance-criteria
pattern this wave's entire E01 epic is built around, so that PERF-02..05's real, valuable relative-
comparison work is not held hostage to a reference-performance-environment decision (DEC-Q9) that
remains a genuine, unresolved human/infrastructure decision at this wave's own planning time.

## Framework capabilities delivered

- Complete-request benchmarks against real PostgreSQL (not fakes) across public/authenticated-read/
  authenticated-write/resource-authz/idempotent-write/async-enqueue profiles, cost-attributed by pool
  wait / tx setup / authz query / handler query / serialization / middleware, published against
  `perf/reference-v1.json` as relative/container comparisons now, with absolute-SLO thresholds
  explicitly conditional on DEC-Q9.
- Rules resolution collapsed from a per-org-ancestor sequential-query loop into bounded, index-verified
  SQL, with result-parity and SQL-count-constant-with-depth proof.
- Sweeper/worker N+1 and unbounded materialization removed from `SweepSLA`, webhook retry, and outbox
  dispatch, the last of these consuming W04's own DATA-02/DATA-03 lease primitives rather than
  re-deriving fencing logic.
- Explicit, required, audited object-checksum behavior for framework uploads, with the full-hash
  fallback bounded to a labeled repair path and a resumable async backfill for legacy objects.
- Benchmark coverage expanded from 8/55 non-cmd packages to include the 7 named hot-path packages
  (`kernel/database`, `jobs`, `outbox`, `workflow`, `auth`, `mfa`, `httpclient`) with bench-budget
  entries for each, per MATRIX CS-16's exact target list.
- A version-pinned security-verification control map linking every applicable ASVS 5.0.0/OWASP API
  Security Top 10 2023/NIST 800-63-4 control to an executable test or an approved waiver, backed by an
  external assessment.
- Fail-not-skip E2E prerequisites, a machine-checked skip manifest, a race-integration test schedule,
  and real time-bounded coverage-guided fuzzing on PR and scheduled runs — closing REL-04's remaining
  truthfulness gaps and, by single ownership, PERF-06's own T3/T4 fuzz scope.
- A framework-side verification that every PROD-01..05 wowsociety-coordination artifact exists and is
  documented, with no wowsociety-repository code change performed by this wave.
- A programme-wide re-run of the REVIEW §30-style final approval gate, a traceability-matrix
  completeness check, and a disposition audit across every `requirement-inventory.md` row — producing
  both a closure report and a separate, explicit production-readiness claim-upgrade decision package for
  the human authority.

## Included epics

- **W07-E01 — performance-programme**: PERF-02..05's relative/container benchmarking programme plus
  MATRIX CS-16's 7-package bench-coverage expansion; DEC-Q9 tracked at epic level.
- **W07-E02 — verification-hardening**: SEC-05's versioned security control map (a closure gate over
  SEC-01-04) and REL-04's remaining coverage-truthfulness work (owning PERF-06 T3/T4's fuzz scope).
- **W07-E03 — product-alignment-verification**: framework-side verification that the PROD-01..05
  wowsociety-coordination artifacts exist, with no wowsociety code change.
- **W07-E04 — programme-closure**: the final verification gate re-run and the closure/claim-upgrade
  decision package.

## Entry criteria

- All seven prior waves' (W00–W06) exit gates satisfied, per `impl/index.md`'s wave map: "Depends on |
  all prior." This is the strictest entry gate of any wave in the programme — no partial entry is
  possible, since W07-E02-S001 (SEC-05) hard-depends on SEC-01/03/04/06 (W03, W05) being substantially
  complete, and W07-E04's own closure-gate purpose requires every other wave to have already executed.

## Exit criteria

- PERF-02..05's relative/container comparison evidence is published against `perf/reference-v1.json`
  for all four findings' own task tables; every absolute-SLO acceptance criterion is explicitly recorded
  as conditional on DEC-Q9, not silently asserted unconditionally.
- BENCH_PKGS covers the 7 MATRIX CS-16-named hot-path packages with passing bench-budget entries.
- SEC-05's control map leaves zero open Critical/High findings per the external assessment, or each open
  finding has an approved waiver.
- REL-04 T5-T8 are complete: fail-not-skip E2E prerequisites; a machine-checked skip manifest; a
  race-integration test schedule; real time-bounded coverage-guided fuzzing on PR + scheduled runs
  (owning PERF-06 T3/T4's identical scope, per CONFLICT-02).
- Every PROD-01..05 row has a documented framework-side coordination artifact confirming existence, per
  W07-E03-S001's own framework-side-only acceptance bar.
- The REVIEW §30-style final approval gate has been re-run across the whole programme; the traceability
  matrix shows every `requirement-inventory.md` row with a disposition and no silent drop; the
  disposition audit confirms every item genuinely reached its recorded disposition, not merely claimed
  to.
- A closure report and a separate, explicit production-readiness claim-upgrade decision package exist
  for the human authority — this wave does not itself declare the framework production-ready.

## Dependencies

Depends on all seven prior waves (W00–W06), per `impl/index.md`'s wave map. See `dependencies.md` for
the full upstream detail, including the specific SEC-05-on-SEC-01/03/04/06 and W07-E01-on-DEC-Q9
dependencies, and the cross-wave PERF-04-T5-on-W04-DATA-02/03 lease-primitive dependency (already
satisfied by W04's own closure, consumed not re-derived here).

## Assumptions

- DEC-Q9 (reference-performance-environment ownership) is confirmed `blocked (human)` in
  `requirement-inventory.md` §B, with REVIEW §F row 9's own provisional default already recorded:
  "a Linux amd64 GitHub Actions runner + committed `perf/reference-v1.json` baseline, advisory-only
  initially; a dedicated bare-metal runner is a *later* SRE decision, not a blocker." This wave's own
  E01 stories proceed against the provisional default (relative/container benchmarking), not blocked on
  DEC-Q9's full resolution, per REVIEW §12's own explicit unblocking framing: "No (relative benchmarking
  proceeds now)."
- PERF-06's own T1/T2 are confirmed already `EXECUTED` at W00-E01-S002, verified there — this wave's own
  W07-E02-S002 picks up only PERF-06's remaining T3/T4 fuzz scope, owned by REL-04 T8, per CONFLICT-02's
  resolution; this wave does not re-implement PERF-06 T1/T2.
- W07-E03's own PROD-01..05 verification is confirmed framework-side-only per mandate §2.3's framework/
  product boundary — this wave does not implement any wowsociety-repository change under any
  circumstance, even where a PROD-0N item's underlying coordination need (e.g. PROD-04's SEC-01
  impersonation cutover) would benefit from a wowsociety-side code change; that change, if it happens, is
  wowsociety's own repository's concern, tracked but not performed here.

## Risks

See `risks.md`. Headline risks: DEC-Q9 remaining unresolved past this wave's own closure, leaving every
PERF-02..05 absolute-SLO acceptance criterion permanently conditional rather than eventually resolved;
SEC-05's external assessment surfacing an open Critical/High finding with no immediate remediation path;
the final verification gate (W07-E04-S001) discovering an unresolved gap in an earlier wave's own
closure that this wave cannot itself fix without reopening that wave's own scope.

## Quality gates

- Every PERF-02..05 relative-comparison claim is proven against `perf/reference-v1.json`, per each
  finding's own task table's own measurement columns — not asserted from code review alone.
- The 7-package bench-coverage expansion (CS-16) is proven by `make bench-budget` passing with the new
  entries present, and by the specific benchmark targets MATRIX CS-16 names (claim/finalize loop,
  tenant-tx open/commit, relay dispatch batch, token verify, TOTP derive, guarded dial) each having a
  corresponding benchmark.
- SEC-05's control map is proven by the external assessment's own report, not by an internal
  self-assessment alone.
- REL-04 T5-T8's fail-not-skip, skip-manifest, race-integration, and real-fuzz claims are each proven by
  their own named fail-first test, per PLAN REL-04's own "Tests" column for each task.
- The final verification gate (W07-E04-S001) is proven by an actual re-run of the REVIEW §30-style gate
  against current HEAD, not a restatement of REVIEW's own original 2026-07-11 conclusions.

## Required artifacts

- PERF-02: DB-backed benchmark suite; cost-breakdown instrumentation; the `perf/reference-v1.json`
  publication.
- PERF-03: the set-based rules-resolution query; index-verification audit; `EXPLAIN` fixtures.
- PERF-04: bounded-batch sweeper/webhook/outbox code; the leased-state-machine outbox rework (consuming
  W04's lease primitives).
- PERF-05: the required-checksum-on-upload enforcement; the bounded repair path; the resumable backfill.
- CS-16: 7 new benchmark files + bench-budget entries.
- SEC-05: the version-pinned control map; the external assessment report.
- REL-04: the fail-not-skip E2E wiring; the skip manifest; the race-integration test schedule; the
  real-fuzz CI wiring (PR + scheduled).
- W07-E03: the PROD-01..05 coordination-artifact-existence record.
- W07-E04: the final verification-gate re-run report; the traceability-matrix completeness check; the
  programme closure report; the production-readiness claim-upgrade decision package.

## Required evidence

- PERF-02..05: relative/container before/after comparison evidence against `perf/reference-v1.json`,
  per finding.
- CS-16: bench-budget-passing evidence for all 7 new packages.
- SEC-05: the external assessment's own report, zero open Critical/High or approved-waiver evidence.
- REL-04: named fail-first test evidence per T5-T8.
- W07-E03: the PROD-01..05 artifact-existence confirmation record.
- W07-E04: the final gate's own pass/fail record per capability area; the traceability completeness
  check's own output; the disposition audit's own output.

## Expected implementation outcome

A framework whose performance characteristics are measured against real infrastructure and published
for relative comparison today, with an honest, explicit trigger for when absolute SLOs will apply; whose
hot-path benchmark coverage matches where a consumer actually spends time, not merely where benchmarks
happen to already exist; whose security posture has been independently assessed against a version-pinned
control map, not merely self-certified; whose integration-test coverage claims are truthful (fail, not
silently skip); and whose overall production-readiness status has been honestly re-assessed at the
programme's actual final state, with the decision to upgrade any production-readiness claim left
explicitly to a human authority, not asserted by this wave itself.

## Acceptance authority

Performance/SRE lead for W07-E01, per PLAN §5.5's own "Accountable role: performance/SRE lead" for
PF-PERF; product-security lead for W07-E02, per PLAN §5.2's own "Accountable role: product-security
lead" for PF-SEC (SEC-05); a cross-functional authority (release/security-engineering lead +
product-security lead + performance/SRE lead + the programme's own designated human authority) for
W07-E03/E04, since programme closure and the claim-upgrade decision span every prior wave's own
accountable role.

## Closure conditions

All exit criteria satisfied; all four epics' `closure-report.md` accepted; `waves/index.md`'s W07 row
updated to reflect `accepted` status; the programme's own `impl/index.md` "Programme acceptance"
criteria are satisfied ("All waves closed per their `closure-report.md`; requirement-traceability matrix
shows every `planned` item accepted/deferred-with-approval; the REVIEW §30-style final gate re-run
passes; no unexplained deviation; production-readiness claim upgrade is a separate, explicit decision");
DEC-Q9's own resolution state (resolved or still-provisional) is recorded honestly, not silently
presented as settled if it remains open.
