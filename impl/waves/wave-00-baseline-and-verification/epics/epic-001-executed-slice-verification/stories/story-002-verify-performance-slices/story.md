---
id: W00-E01-S002
type: story
title: Verify performance and benchmark-budget-gate slices at current HEAD
status: accepted
wave: W00
epic: W00-E01
owner: W00E01S002 (wave-00 verification worker)
reviewer: unassigned
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-13
source_requirements: [PERF-01, PERF-06, SD-03]
depends_on: []
blocks: []
acceptance_criteria: [AC-W00-E01-S002-01, AC-W00-E01-S002-02, AC-W00-E01-S002-03]
artifacts: [ART-W00-E01-S002-001, ART-W00-E01-S002-002, ART-W00-E01-S002-003]
evidence: [EV-W00-E01-S002-01, EV-W00-E01-S002-02, EV-W00-E01-S002-03]
decisions: []
risks: []
---

# W00-E01-S002 — Verify performance and benchmark-budget-gate slices at current HEAD

## Story ID

W00-E01-S002.

## Title

Verify performance and benchmark-budget-gate slices at current HEAD.

## Objective

Re-verify, at the story's closing commit SHA, two finding-slices claimed EXECUTED by
`premier-framework-implementation-plan.md` (PLAN) / `fable5-final-architecture-review-2026-07-11.md`
(REVIEW) — PERF-01 (token-bucket sweep fix in `kernel/httpx/ratelimit.go`) and PERF-06 T1 (fail-closed
missing-benchmark gate in `internal/tools/benchbudget/main.go`) — and confirm a session-delta fact,
SD-03, that the #25 sweep-bench recalibration is correctly reflected in `bench-budgets.txt`. This
story produces mandate-§10-conformant evidence for all three; it implements nothing new.

## Value to the framework

The framework's own performance-regression gate (`make bench-budget`) and its rate-limiter's
sweep-eviction correctness are load-bearing for every later wave that touches `kernel/httpx` or relies
on CI catching a performance regression before merge. If PERF-01's sweep fix or PERF-06's fail-closed
gate has silently regressed since the reviewed SHA, every subsequent wave inherits a false sense of
safety — this story is a generic-framework confidence check, not a downstream-product concern, since
both the token-bucket limiter and the benchmark-budget tool are kernel/tooling infrastructure shared by
any product built on wowapi.

## Problem statement

`impl/analysis/requirement-inventory.md` §A records PERF-01 and PERF-06 (T1 portion) at `INV`
(implemented-needs-verification) disposition — each described in PLAN/REVIEW as already implemented and
tested, but the evidence backing that claim lives only as prose test descriptions and `git log`
commit-SHA citations that predate `impl/`'s own evidence register (mandate §10). Separately,
requirement-inventory.md §E records session-delta fact SD-03: commit `0a31186` (PR #25) fixed an
O(n²)+empty-map measurement defect in the sweep benchmarks and recalibrated `bench-budgets.txt`'s
budget values accordingly. Because SD-03 changed PERF-01's evidence basis *after* the PLAN/REVIEW
documents were written, this story must confirm the re-verification runs against the POST-#25 state,
not silently accept a stale pre-#25 baseline (wave-level risk `RISK-W00-003` names exactly this
failure mode).

## Source requirements

- PERF-01 — Token-bucket sweep fix. `requirement-inventory.md`: `INV | W00-E01-S002 | EXECUTED + #25
  recalibrated sweep budgets to honest full-map measurements — verify at current HEAD`.
- PERF-06 — Fail-closed performance gates (T1 only). `requirement-inventory.md`: `INV | W00-E01-S002 |
  T1 EXECUTED; T3/T4 fuzz scope = REL-04 T8 single-owner (W07-E02-S002)`.
- SD-03 — Sweep-bench O(n²)+empty-map fix; budgets recalibrated (#25). `requirement-inventory.md` §E:
  `Sweep-bench O(n²)+empty-map fix; budgets recalibrated (#25) — PERF-01 evidence basis changed —
  W00-E01-S002 verifies against NEW budgets`.

## Current-state assessment

Confirmed (by reading source at the time this story was drafted, not yet by re-running tests):

- `kernel/httpx/ratelimit.go`, `kernel/httpx/ratelimit_test.go`, `kernel/httpx/bench_test.go`, and
  `kernel/httpx/export_test.go` exist in the repository and are the files PLAN/REVIEW name for PERF-01.
- `internal/tools/benchbudget/main.go` and `internal/tools/benchbudget/coverage_test.go` exist; the
  latter contains `TestMainMissingBenchmarkFails` (see "Assumptions" below on why this is treated as
  the PERF-06 T1 test, not asserted with full certainty), whose doc comment states: "confirm a
  budgeted-but-absent benchmark causes a real CI failure (exit 1), not just a warning. See PERF-06 T1."
- `bench-budgets.txt` exists at the repository root and currently contains 43 non-comment, non-blank
  budget-entry lines, and its header comment states budgets were remeasured "Apple M3 Max (2026-07-04)"
  — consistent with, but not yet confirmed as definitively caused by, the #25 recalibration.
- Commit `0a31186` ("perf(bench): fix O(n²) setup + empty-map measurement in sweep benchmarks (#25)")
  exists in the repository's commit history, confirming SD-03's underlying fix commit is real and
  merged.

Not yet confirmed (requires actually running the commands, which this planning-stage story does not
do): whether `go test ./kernel/httpx/... -race` and `make bench-budget` currently pass; whether
`TestMainMissingBenchmarkFails` and the other benchbudget coverage tests currently pass; whether the 43
entries and their values are the exact set the #25 recalibration produced (as opposed to, e.g., a
subsequent uncaptured drift). All of these are exactly what this story's tasks re-verify.

## Desired state

All three source items have registered, mandate-§10-conformant evidence at the story's closing commit
SHA: PERF-01's sweep-recompute-on-eviction and hard-cap behavior confirmed intact via passing race and
benchmark-budget runs; PERF-06 T1's fail-closed missing-benchmark behavior confirmed intact via a
passing subprocess exit-1 assertion; and `bench-budgets.txt`'s entry count/values confirmed to reflect
the post-#25 recalibrated state, not a stale pre-#25 baseline. If any of the three fails to re-verify,
the story does not close — a follow-up investigation task is opened within this story (see "Rollback or
recovery considerations" in each task) rather than the story being silently marked accepted.

## Scope

- Re-running `go test ./kernel/httpx/... -race` and `make bench-budget` to re-verify PERF-01
  (task T001).
- Re-running the benchbudget package's coverage test(s) — specifically the subprocess-based exit-1
  assertion for a missing budgeted benchmark — to re-verify PERF-06 T1 (task T002).
- Inspecting `bench-budgets.txt`'s entry count and spot-checking its values against the #25
  O(n²)+empty-map fix description, to confirm SD-03 (task T003).
- Registering mandate-§10 evidence records for all three re-runs (command, commit SHA, environment,
  tool versions, date/time, result, reviewer).

## Out of scope

- PERF-06 T3/T4 (fuzz-testing scope) — `requirement-inventory.md` targets these at REL-04 T8, owned by
  W07-E02-S002, not this story.
- Any code change, fix, or remediation to `kernel/httpx/ratelimit.go`, `internal/tools/benchbudget/`,
  or `bench-budgets.txt` — this story is verification-only (epic-level scope statement,
  `epic.md` "Out of scope"). If a regression is found, the story stays open rather than being closed
  with a silent fix.
- SEC-02, AR-04, AR-06 (covered by sibling story W00-E01-S001) and DATA-08/REL-04 (covered by sibling
  story W00-E01-S003) — disjoint packages and disjoint test commands, tracked separately per
  `epic.md` "Architectural context."
- Quantitative baseline capture of the bench-budget snapshot as a wave-level artifact — that is
  W00-E02-S001's scope; this story only confirms PERF-01/PERF-06/SD-03's finding-slice correctness,
  not the wave-level baseline-capture deliverable.

## Assumptions

- `TestMainMissingBenchmarkFails` in `internal/tools/benchbudget/coverage_test.go` is the test PLAN/
  REVIEW intended by "PERF-06 T1 EXECUTED" — its doc comment explicitly says "See PERF-06 T1," which is
  strong but not certain confirmation there is no other, more specifically-named, test function this
  refers to. Task T002 states the exact `-run` pattern must be confirmed against the file's actual test
  function name at execution time (mandate §8.5: "do not invent precise... where the repository does
  not yet provide enough information").
- `bench-budgets.txt` at the repository root (confirmed to exist there at drafting time) is the file
  SD-03 and RISK-W00-003 refer to; no other `bench-budgets*.txt` file was found elsewhere in the
  repository at drafting time, but task T003 states this must be reconfirmed at execution time in case
  of drift.
- The 43 non-comment/non-blank lines counted in `bench-budgets.txt` at drafting time is the figure
  RISK-W00-003 and `requirement-inventory.md` mean by "43 budgeted entries" — this counting method
  (grep -v '^#' minus blank lines) is an assumption about how "entries" should be counted; task T003
  must reconfirm this against `make bench-budget`'s own parsed-entry count if that tool reports one, as
  a stronger source of truth than a line count.
- A working Go toolchain and the repository's existing Makefile targets (`make bench-budget`) are
  available in whatever environment executes this story's tasks; no external database or S3 dependency
  is assumed for any of this story's three tasks, unlike sibling story S003.

## Dependencies

- No dependency on any other epic or wave (epic-level: "W00-E01 is entry-point work").
- No dependency on sibling stories S001 or S003 — disjoint packages/files/commands per `epic.md`
  "Architectural context" and `dependencies.md`; S002 can execute in any order relative to them or in
  parallel.
- Internal to this story: T001 (PERF-01), T002 (PERF-06 T1), and T003 (SD-03 budget confirmation) are
  independent of each other in principle (different commands, different files) but T003 is most
  meaningfully executed alongside or after T001, since T001's `make bench-budget` run is itself the
  mechanism that would surface a budget-file problem; T003 is kept as its own task rather than folded
  into T001 because it is a data (budget-file-content) confirmation distinct from T001's code-behavior
  confirmation, even though both concern the same underlying #25 fix.

## Affected packages or components

No production code is expected to change (this is a re-verification story). The packages/files this
story's tasks read and run tests against:

- `kernel/httpx/ratelimit.go` — token-bucket rate limiter, sweep/eviction logic.
- `kernel/httpx/ratelimit_test.go` — 10k-eviction and race-condition tests.
- `kernel/httpx/bench_test.go` — sweep benchmark definitions.
- `kernel/httpx/export_test.go` — internal-state test-only exports supporting the above.
- `internal/tools/benchbudget/main.go` — benchmark-budget-gate tool; missing-benchmark-path behavior.
- `internal/tools/benchbudget/coverage_test.go` — subprocess-based coverage tests for `main.go`,
  including the missing-benchmark exit-1 assertion.
- `bench-budgets.txt` (repository root) — the budget-entry data file itself.

## Compatibility considerations

Not applicable in the implementation sense (no code changes). In the verification sense: task T001's
work statement explicitly notes the sweep fix preserved the rate limiter's backward-compatible 2-arg
constructor; re-verification confirms that compatibility surface is still intact, not just the sweep
behavior in isolation.

## Security considerations

Not directly applicable — the token-bucket rate limiter has security-adjacent value (it is part of the
framework's abuse/DoS-mitigation surface), so a regression in PERF-01's sweep correctness could
indirectly degrade that protection (e.g., stale entries never evicted, unbounded memory growth). This
story's re-verification is itself the security-relevant check; no new security control is introduced or
assessed beyond confirming the existing one still functions.

## Performance considerations

This is the load-bearing consideration for this story — the story exists specifically to protect
performance-gate integrity. Two distinct performance concerns are in scope:

- **PERF-01 runtime behavior**: the token-bucket sweep must recompute refill during sweep (not merely
  on `Allow`), and a hard cap on tracked entries must be enforced, to avoid unbounded memory growth
  under sustained load. Task T001 re-verifies this via the 10k-eviction race test and the sweep
  benchmarks.
- **PERF-06 gate integrity**: the benchmark-budget tool (`make bench-budget`) must fail the build (exit
  non-zero) when a budgeted benchmark is missing from the bench output, rather than silently
  WARN-and-continue — a silent-pass gate provides no actual regression protection. Task T002
  re-verifies this via the tool's own subprocess exit-1 test.
- **SD-03 budget-baseline correctness**: `bench-budgets.txt`'s 43 entries must reflect the post-#25
  O(n²)+empty-map-fix recalibration, not stale pre-#25 numbers, or every later wave's "improvement over
  baseline" performance claim would be measured against the wrong starting point (this is exactly
  `RISK-W00-003`). Task T003 re-verifies this by inspection.

## Observability considerations

Not applicable — no logging, metrics, or tracing changes are in scope for this verification-only
story. `make bench-budget`'s own console/CI output is the observability surface being re-verified (its
failure-reporting text), not a framework observability concern.

## Migration considerations

Not applicable — no data, schema, or configuration migration is involved in re-running existing tests
and inspecting an existing budget file.

## Documentation requirements

None beyond this story's own governance documents (`story.md`, `plan.md`, `implementation.md`,
`verification.md`, `deviations.md`, `closure.md`, task files, artifact/evidence indexes). No
user-facing or `docs/` documentation is expected to require updates as a result of this
verification-only story, since no behavior changes.

## Acceptance criteria

- **AC-W00-E01-S002-01**: `go test ./kernel/httpx/... -race` exits 0 AND `make bench-budget` exits 0 at
  the story's closing commit SHA, confirming PERF-01 (sweep recomputes refill during sweep, hard cap
  enforced, backward-compatible 2-arg constructor preserved); evidence registered in
  `evidence/index.md`.
- **AC-W00-E01-S002-02**: The benchbudget package's subprocess-based missing-benchmark test (see task
  T002 for exact `-run` pattern, confirmed at execution time) exits 0 (i.e., the underlying exit-1
  assertion it makes passes), confirming PERF-06 T1's missing-benchmark path is a tracked failure, not
  WARN+continue; evidence registered in `evidence/index.md`.
- **AC-W00-E01-S002-03**: `bench-budgets.txt` contains 43 non-comment, non-blank budget-entry lines
  (or the actual current count if different, in which case the discrepancy itself is escalated per
  RISK-W00-003 rather than silently accepted), and a spot-check of at least 3 entries' values is
  consistent with the #25 O(n²)+empty-map-fix recalibration description; evidence registered in
  `evidence/index.md`.

## Required artifacts

- None new — this story consumes existing repository artifacts (`kernel/httpx/ratelimit.go`,
  `internal/tools/benchbudget/main.go`, `bench-budgets.txt`) rather than producing new ones. See
  `artifacts/index.md` for the story's declared expected artifact entries (tool-output logs from the
  verification commands), registered at "not yet produced" status.

## Required evidence

- Test-execution log for `go test ./kernel/httpx/... -race` (AC-01).
- `make bench-budget` tool-output log (AC-01).
- Test-execution log for the benchbudget coverage test's missing-benchmark exit-1 assertion (AC-02).
- Entry-count and spot-check inspection note for `bench-budgets.txt` (AC-03).

See `evidence/index.md` for the full evidence-record entries, registered at "pending" status.

## Definition of ready

Per `governance/definition-of-ready.md` Story DoR: this story is specific (two performance/quality-gate
finding-slices plus one session-delta confirmation, not an aspirational theme); bounded (scope/out-of-
scope both stated above); implementable (the exact re-run commands are known, with the one open
question — PERF-06 T1's exact `-run` pattern — explicitly flagged as "must be confirmed at execution
time" rather than silently assumed); independently reviewable and verifiable (does not depend on S001
or S003 completing first); traceable to source requirements (front matter cites PERF-01, PERF-06,
SD-03); has measurable, numbered acceptance criteria; has dependencies and assumptions recorded above;
has a `plan.md` with task breakdown and unresolved questions; states expected artifact/evidence types;
and addresses compatibility/security/performance/observability/migration considerations above (several
explicitly marked not applicable with a one-line reason, per DoR). This story satisfies Story DoR and
may move `planned` → `ready` when an owner is assigned.

## Definition of done

Per `governance/definition-of-done.md` (referenced; not reproduced here): this story will satisfy DoD
when all three acceptance criteria have post-execution `verification.md` records with actual result,
pass/fail, evidence ID, execution date, commit SHA, environment, and reviewer; `implementation.md`
records what was actually run; `deviations.md` records any deviation from this plan (or confirms none
occurred); and `closure.md` is completed with reviewer conclusion. Per mandate §7, this story must not
be marked `accepted` solely because all three tasks are marked `done` — the acceptance authority must
independently confirm the evidence proves the acceptance criteria.

## Risks

- `RISK-W00-001` (epic/wave-level) — a finding-slice claimed EXECUTED fails to re-verify at current
  HEAD (regression since the reviewed SHA). Directly applicable to both PERF-01 (T001) and PERF-06 T1
  (T002).
- `RISK-W00-003` (epic/wave-level) — bench-budget baseline captured against stale, pre-#25 values if
  SD-03 is not correctly confirmed. Directly applicable to T003; this is the specific risk T003 exists
  to close out.
- Story-specific risk (not yet registered with a stable ID; to be added to the risk register if this
  story surfaces it as material): the exact `-run` pattern for PERF-06 T1's benchbudget coverage test
  could differ from `TestMainMissingBenchmarkFails` if a repository change renamed or restructured the
  test between drafting and execution — mitigated by task T002's instruction to reconfirm the test name
  against the file at execution time before running it.

## Residual-risk expectations

Even after this story is accepted, some residual risk remains: re-verification proves the slices are
intact at one pinned commit SHA, not permanently — a later commit could regress PERF-01's sweep
behavior or PERF-06's fail-closed gate without this story's evidence catching it (evidence is
revision-pinned per `evidence-policy.md`, not a standing guarantee). This residual risk is expected and
accepted as inherent to point-in-time verification; it is not a defect in this story's scope. Ongoing
protection against future regression is provided by the tests themselves running in CI on every
subsequent change, not by this story.

## Plan

See sibling file `plan.md` for the full §8.5 proposed-approach content (proposed architecture,
implementation strategy, task breakdown, unresolved questions, and approval conditions), per this
story template's guidance that the plan skeleton is instantiated as its own file rather than left
embedded here.
