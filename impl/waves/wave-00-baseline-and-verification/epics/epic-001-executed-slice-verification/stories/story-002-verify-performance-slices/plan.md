---
id: PLAN-W00-E01-S002
type: plan
parent_story: W00-E01-S002
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
---

# Plan — W00-E01-S002

Per mandate §8.5: "Do not invent precise code changes where the repository does not yet provide enough
information. Clearly distinguish confirmed facts, planned changes, and implementation assumptions."
This plan describes a **re-verification** approach — there is no new code to design or implement.

## Proposed architecture

Not applicable. This story introduces no new architecture, contract, or interface. **No code changes
are expected as a result of this story.** The "architecture" being validated already exists in
`kernel/httpx/ratelimit.go` (token-bucket sweep) and `internal/tools/benchbudget/main.go`
(benchmark-budget gate); this plan's only concern is how to re-prove that existing architecture still
behaves as PLAN/REVIEW claim.

## Implementation strategy

Re-run the exact named test files/commands PLAN/REVIEW cite for PERF-01 and PERF-06 T1, at this
story's own closing commit SHA, and inspect `bench-budgets.txt` for SD-03. Each of the three tasks
below is a confirmed-fact-gathering activity, not an implementation activity. If any re-run fails, the
failure itself is preserved as evidence (mandate §10 failed-evidence-preservation rule) and escalated
as a finding — it is not silently retried until green, and it is not "fixed" within this story (fixes
are out of scope; see `story.md` "Out of scope").

## Expected package or module changes

None. No package or module is expected to change as part of this story.

## Expected file changes where determinable

None. No file is expected to change. The files this story's tasks touch are read-only inputs
(`kernel/httpx/ratelimit.go` and its tests, `internal/tools/benchbudget/main.go` and its tests,
`bench-budgets.txt`) plus new governance/evidence documents this story itself creates under its own
`impl/` directory tree (task/verification/evidence records) — those are documentation artifacts of the
verification process, not framework code changes.

## Contracts and interfaces

None new or changed. Existing contracts under test: the token-bucket rate limiter's public
constructor(s) (backward-compatible 2-arg form, per PERF-01's confirmed-implemented behavior) and the
benchbudget tool's CLI exit-code contract (0 = all budgets satisfied, non-zero = at least one
violation, including missing-benchmark).

## Data structures

None new or changed.

## APIs

None new or changed.

## Configuration changes

None. `bench-budgets.txt` is treated as a read-only data input for this story's task T003; it is
inspected, not modified.

## Persistence changes

None applicable — no database, storage, or persistence layer is involved in this story.

## Migration strategy

Not applicable.

## Concurrency implications

None new. Re-verification of PERF-01 exercises the existing race-detector coverage
(`go test ... -race`) for the token-bucket sweep's concurrent-access paths, but does not add new
concurrency surface.

## Error-handling strategy

Not applicable to new code (none is written). The error-handling behavior *under verification* is
PERF-06 T1's fail-closed contract: `internal/tools/benchbudget/main.go` must exit non-zero when a
budgeted benchmark is missing, rather than warning and continuing. Task T002 re-confirms this contract
holds.

## Security controls

None new. See `story.md` "Security considerations."

## Observability changes

None.

## Testing strategy

This story's entire content *is* its testing strategy — re-running existing tests, not writing new
ones (mandate §13: "Do not create tests merely to increase numerical coverage... avoid duplicating
existing test coverage unless the new test closes an identified behavioural gap." No behavioral gap is
identified here; existing coverage is re-exercised, not extended):

- **T001 (PERF-01)**: `go test ./kernel/httpx/... -race` — exercises `ratelimit_test.go`'s 10k-eviction
  and race-condition assertions. `make bench-budget` — exercises `bench_test.go`'s sweep benchmarks
  against `bench-budgets.txt`'s budgeted entries.
- **T002 (PERF-06 T1)**: the benchbudget package's subprocess-based coverage test exercising the
  missing-benchmark-fails-the-build path (candidate: `TestMainMissingBenchmarkFails` in
  `internal/tools/benchbudget/coverage_test.go` — see "Unresolved questions" below for why this is a
  candidate, not a confirmed final answer). This story additionally *describes*, but explicitly does
  NOT execute during this planning-document pass, a manual fail-first revert-proof check: temporarily
  remove a budget entry from a scratch copy of `bench-budgets.txt`, confirm `make bench-budget` now
  fails non-zero, then restore the original file. That manual check belongs to T002's actual execution
  phase, not to this planning artifact.
- **T003 (SD-03)**: inspection, not test execution — count `bench-budgets.txt`'s non-comment/non-blank
  entry lines and spot-check values against the #25 fix's description (O(n²) setup cost removed,
  empty-map measurement corrected).

## Regression strategy

Not applicable in the introduce-new-regression sense (no code changes). In the detect-existing
regression sense: this entire story is a regression-detection exercise for PERF-01/PERF-06/SD-03. Any
regression found is preserved as `failed` evidence and escalated, per `story.md` "Rollback or recovery
considerations" pattern described in each task below.

## Compatibility strategy

Re-verification confirms the existing backward-compatible 2-arg constructor for the rate limiter
remains intact; no compatibility change is introduced.

## Rollout strategy

Not applicable — no code is rolled out. The "output" of this story is evidence records and governance
document state changes (status transitions), which take effect immediately upon the story's acceptance.

## Rollback strategy

Not applicable to code (none changes). If verification reveals a regression, the *rollback* concept
does not apply in the usual sense of reverting a change — see "Unresolved questions" and each task's
own "Rollback or recovery considerations" for how a regression finding is handled procedurally
(the story stays open; a follow-up investigation task is opened within this story, since PERF-01/PERF-06
target W00-E01-S002 itself per their `INV` disposition in `requirement-inventory.md`, unlike findings
such as SEC-02/AR-04/AR-06 in sibling story S001 whose *unexecuted remainder* targets a distinct
future-wave story).

## Implementation sequence

1. Confirm environment prerequisites (Go toolchain available; no external DB/S3 dependency needed for
   this story, unlike sibling S003).
2. Execute T001: `go test ./kernel/httpx/... -race`, then `make bench-budget`. Record result.
3. Execute T002: confirm the exact benchbudget coverage test name/`-run` pattern against
   `internal/tools/benchbudget/coverage_test.go`'s current content, run it, record result. Separately
   *describe* (do not execute in this planning pass) the fail-first revert-proof manual check as part
   of the task's documented verification method.
4. Execute T003: inspect `bench-budgets.txt`, count entries, spot-check values, record result.
5. Register evidence for all three (mandate §10 fields) in `evidence/index.md`.
6. Update `verification.md` post-execution record, `implementation.md`, and `deviations.md` (or confirm
   no deviation).
7. Complete `closure.md` once all three ACs are proven or a regression is escalated.

## Task breakdown

- **W00-E01-S002-T001** — Re-verify PERF-01: token-bucket sweep fix. Commands: `go test
  ./kernel/httpx/... -race`, `make bench-budget`. Related AC: AC-W00-E01-S002-01.
- **W00-E01-S002-T002** — Re-verify PERF-06 T1: fail-closed benchbudget missing-benchmark gate.
  Command: the benchbudget coverage test's subprocess exit-1 assertion (exact `-run` pattern to be
  confirmed at execution time). Related AC: AC-W00-E01-S002-02.
- **W00-E01-S002-T003** — Confirm SD-03: #25 bench-budget recalibration reflected in
  `bench-budgets.txt`. Activity: inspection, not test execution. Related AC: AC-W00-E01-S002-03.

## Expected artifacts

None new beyond this story's own governance/tracking documents. See `artifacts/index.md` for the
declared expected artifact entries (verification tool-output logs), registered "not yet produced."

## Expected evidence

- `EV-W00-E01-S002-01` (planned ID) — `go test ./kernel/httpx/... -race` + `make bench-budget` output,
  proving AC-01.
- `EV-W00-E01-S002-02` (planned ID) — benchbudget coverage-test output, proving AC-02.
- `EV-W00-E01-S002-03` (planned ID) — `bench-budgets.txt` entry-count/spot-check inspection note,
  proving AC-03.

These evidence IDs are planned identifiers only — no evidence has been produced yet (see
`evidence/index.md`, status "pending").

## Unresolved questions

- **Exact `-run` pattern for PERF-06 T1's re-verification test.** `TestMainMissingBenchmarkFails` in
  `internal/tools/benchbudget/coverage_test.go` is the strongest candidate — its doc comment reads "See
  PERF-06 T1" verbatim — but this plan does not treat that as fully confirmed. Task T002 must
  re-inspect the file at execution time and use whatever the actual current test function name is,
  rather than assuming this plan's candidate name remains accurate if the file has changed since
  drafting.
- **Exact path convention for `bench-budgets.txt`.** Confirmed at drafting time to live at the
  repository root (`bench-budgets.txt`, sibling to `go.mod`/`Makefile`). This plan treats the path as
  confirmed, not merely assumed, because it was directly located during drafting — but task T003 should
  reconfirm the path has not moved by the time it executes, in case of intervening repository
  reorganization.
- **Counting method for "43 budgeted entries."** RISK-W00-003 and `requirement-inventory.md` state "43
  budgeted entries" without specifying the exact counting method. This plan's assumption (non-comment,
  non-blank lines) produced 43 at drafting time, consistent with the cited figure — but if
  `make bench-budget` itself reports a parsed-entry count in its output, task T003 should prefer that
  tool-reported count as the more authoritative source over a manual line count.
- **Whether any regression exists.** Not knowable until T001/T002/T003 actually execute. This plan
  takes no position on whether PERF-01, PERF-06 T1, or the bench-budget baseline currently pass —
  stating otherwise would violate mandate §18 ("Do not claim that tests passed unless they were
  actually executed").

## Approval conditions

This plan is considered approved and ready for implementation (i.e., ready for the story to move
`planned` → `ready`) when: an owner is assigned; the Definition of Ready checklist in `story.md`
"Definition of ready" is satisfied; and the acceptance authority has no open objection to the task
breakdown or the three unresolved questions above being resolved during task execution rather than
before it (consistent with mandate §8.5's guidance not to invent precision the repository does not yet
support).
