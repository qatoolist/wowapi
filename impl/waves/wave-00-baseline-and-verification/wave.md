---
id: W00
type: wave
title: Baseline and verification
status: accepted
owner: unassigned
reviewer: unassigned
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-13
included_epics:
  - W00-E01
  - W00-E02
depends_on: []
blocks:
  - W01
  - W02
  - W03
  - W04
  - W05
  - W06
  - W07
source_requirements:
  - AR-04
  - AR-05
  - AR-06
  - SEC-02
  - PERF-01
  - PERF-06
  - DATA-08
  - REL-04
  - D-01
  - D-02
  - D-03
  - D-04
  - D-05
  - D-06
  - D-07
  - D-08
  - D-09
  - SD-01
  - SD-02
  - SD-03
  - SD-04
---

# W00 — Baseline and verification

## Objective

Re-verify, at the current repository HEAD, every finding-slice that the plan/review/matrix
documents claim was already executed — and capture the quantitative baselines (coverage, lint,
benchmark budgets, CI wall-clock, dependency/toolchain inventory) that every subsequent wave's
acceptance criteria will be measured against. Wave 00 implements nothing new; it re-pins existing
claims and establishes the "before" state for waves 01-07.

## Rationale

The mandate (§15) explicitly asks whether a dedicated Wave 00 is necessary for baseline capture,
repository health assessment, current coverage/static-analysis measurement, dependency inventory,
architectural decision inventory, and generation of implementation prerequisites — and directs that
Wave 00 not be assumed to contain feature implementation unless required to make later work safe and
measurable. Here it is required: `requirement-inventory.md` §A/§C records 8 PLAN findings with
disposition `implemented-needs-verification` or `partial` (executed portions only) plus 3 MATRIX
verify-outcome rows already independently re-earned by the closure-depth pass (CS-03, CS-19, CS-24).
None of these have evidence registered against the current `impl/` traceability structure — the
review/matrix documents cite `git log` commit SHAs and prose test descriptions, not
`impl/tracking/evidence-register.md`-conformant evidence records. Re-running the named tests now
and registering the evidence closes that gap before any wave that depends on these slices being true
(W01's AR-04/AR-06 remainder work, W03's SEC-02 ratification work, W07's REL-04/PERF-06 fuzz work)
can safely build on them. Session-delta facts SD-01..SD-04 (`requirement-inventory.md` §E — CI
parallelization, bench recalibration, doc archival) postdate the MATRIX/REVIEW documents and must be
folded into the baseline, not treated as already-known.

## Framework capabilities delivered

None — Wave 00 delivers no new framework capability. It delivers a **verified baseline**: a set of
evidence records proving 8 finding-slices are real at HEAD, and a set of baseline artifacts (coverage
%, lint hit-counts, bench-budget state, CI timing, dependency inventory) that later waves' acceptance
criteria reference as "regression from this point" or "improvement over this point."

## Included epics

- **W00-E01 — executed-slice-verification**: re-run the named tests for SEC-02, PERF-01, PERF-06,
  DATA-08 (W0 slice), AR-04 (T1), AR-05 (T1/T2), AR-06 (T1), REL-04 (T1-T4) at current HEAD; register
  evidence.
- **W00-E02 — baseline-capture**: capture coverage/lint/bench/CI-wall-clock baselines; inventory
  dependencies and pinned tool versions against REVIEW §L's approved-dependency register; write
  ADR files for D-01..D-09 into the decision register.

## Entry criteria

- Repository builds and `make ci` passes at the wave's starting commit (informational — not a Wave 00
  deliverable to prove, but a precondition for any test in E01 to be meaningful).
- `impl/analysis/requirement-inventory.md` exists and is the canonical allocation (satisfied —
  present at `impl/analysis/requirement-inventory.md`).

## Exit criteria

- All 8 executed finding-slices (SEC-02, PERF-01, PERF-06, DATA-08 W0, AR-04 T1, AR-05 T1/T2, AR-06
  T1, REL-04 T1-T4) have a `verification.md` with actual-result rows and registered evidence IDs,
  re-run at the wave's closing commit SHA — not merely cited from the review/matrix documents.
- CS-03 (config fail-closed + fingerprint), CS-19 (i18n freeze + key-echo), CS-24 (SSRF dial-time
  guard) verify-outcome claims are re-pinned with an evidence pointer at the same commit.
- Coverage %, full-tree lint state (including the 25-analyzer hit counts named in MATRIX CS-23),
  bench-budgets state post-#25 recalibration, and CI wall-clock per leg are captured as evidence
  artifacts under W00-E02-S001.
- go.mod direct/indirect dependency inventory and pinned tool versions are captured and cross-checked
  against REVIEW §L's approved register (all 10 original + backoff/golang-lru/gobreaker already
  approved — confirm no drift).
- ADR files exist for D-01 through D-09, registered in the decision register, each stating
  recommendation, safe default, and owner per REVIEW §U.
- Independent review confirms no fabricated evidence (dates/commits/tool-output actually produced,
  not asserted).

## Dependencies

None — Wave 00 is the first wave and has no upstream wave dependency. Individual stories depend only
on the ability to check out and run the current repository at HEAD.

## Assumptions

- The 8 finding-slices described as EXECUTED in PLAN §8/§9 and REVIEW §D are still present and
  unmodified at the wave's working HEAD; if any has regressed, that is itself a finding this wave's
  verification work surfaces (not an assumption the wave can proceed under).
- The commit `main @ 0a31186` cited in `impl/index.md` as the planning-time HEAD is a starting
  reference point, not necessarily the exact commit this wave executes against — the wave's own
  verification records must cite whatever commit SHA the work actually runs against.
- `docker compose` / Postgres / MinIO test infrastructure referenced by the plan's evidence commands
  (`make ci-container`, S3-gated tests) is available in the execution environment.

## Risks

See `risks.md` for the full register. Headline risk: a claimed-executed slice fails to re-verify at
current HEAD (regression since the review/matrix documents were written), which would block later
waves that build on it (e.g., W03's SEC-02 ratification depends on SEC-02 T1-T3 being genuinely
fail-closed today, not just at the reviewed SHA).

## Quality gates

- Every re-run test command and its output is captured as evidence (mandate §10 — evidence must
  identify commit SHA, execution command, environment, tool versions, date/time, result).
- No evidence record may claim "pass" without the actual command having been executed in this pass —
  citing the review/matrix documents' prior claim is not sufficient re-verification.
- Failed evidence (if any slice fails to re-verify) is preserved, not deleted, and marked per mandate
  §10 (failed / superseded / retested / resolved / accepted exception).

## Required artifacts

- Baseline coverage report, lint report, bench-budget snapshot, CI timing log (W00-E02-S001).
- Dependency/toolchain inventory document (W00-E02-S002).
- Nine ADR files, D-01 through D-09 (W00-E02-S003).

## Required evidence

- Re-run test output for each of the 8 executed slices' named test files/commands (W00-E01-S001/S002/S003).
- Evidence pointers for CS-03/CS-19/CS-24 verify-outcome re-confirmation.

## Expected implementation outcome

No production code changes are expected from Wave 00. The expected outcome is a fully evidenced,
current-HEAD-pinned baseline that every later wave's "before" state and every re-verification claim
can cite, plus nine ratified ADRs that resolve the design questions later waves would otherwise have
to re-litigate (D-02 Registrar design blocks W05's AR-01 T2; D-01 grant authority blocks W03's SEC-01
T1; D-06 epoch-table design blocks W05's SEC-04 T4; and so on).

## Acceptance authority

Framework architecture lead (role-based owner per mandate discipline — no named human DRI assigned
yet, per `impl/index.md`'s scope-discipline note inherited from PLAN §1 footnote 7).

## Closure conditions

All exit criteria above satisfied; `closure-report.md` completed with reviewer conclusion and
acceptance date; both epics' `closure-report.md` accepted; no evidence gap flagged as unresolved.
