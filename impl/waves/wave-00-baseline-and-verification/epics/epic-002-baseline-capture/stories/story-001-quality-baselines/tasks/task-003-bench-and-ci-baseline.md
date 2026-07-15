---
id: W00-E02-S001-T003
type: task
title: Bench-budget and CI wall-clock baseline
status: done
parent_story: W00-E02-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on: []
acceptance_criteria:
  - AC-W00-E02-S001-03
  - AC-W00-E02-S001-04
artifacts: []
evidence: []
---

# W00-E02-S001-T003 — Bench-budget and CI wall-clock baseline

## Task Definition

*Per mandate §8.6. Defines the task before work begins.*

### Task objective

Run `make bench-budget` to confirm the post-#25-recalibration benchmark-budget state (expected 43
budgeted entries), and separately record the current `.github/workflows/ci.yml` shape and per-leg
wall-clock timing, reflecting the SD-01 (3-leg parallelization, #23) and SD-02 (bench path-scoping,
#24) session-delta facts. Grouped as one task per `../plan.md`'s stated rationale: both are
"inspect current CI/bench configuration and record its current numbers," sharing owner and risk
profile, with no dependency between them.

### Parent story

W00-E02-S001 — Quality baselines.

### Owner

Unassigned.

### Status

`todo` (per `impl/governance/status-model.md` §7.3).

### Dependencies

None. This task can run independently of T001 and T002.

### Detailed work

**Bench-budget baseline:**

1. Confirm test infrastructure is reachable (Postgres, per `TEST_DSN`/`DATABASE_URL`).
2. By direct inspection of `bench-budgets.txt`, count its budgeted entries and confirm the count
   equals the expected post-#25 value (43, per `impl/waves/wave-00-baseline-and-verification/
   risks.md` RISK-W00-003 and MATRIX CS-16). If the count differs, record the actual count and flag
   this as drift — do not silently assume 43 without confirming.
3. Run `make bench-budget` (`DATABASE_URL=... WOWAPI_REQUIRE_DB=1 go test -bench=. -benchmem
   -run=^$ $(BENCH_PKGS) | go run ./internal/tools/benchbudget bench-budgets.txt`).
4. Capture the pass/fail result per budgeted benchmark. A budget violation is a valid, recordable
   result — record it, do not treat it as a task-execution failure.
5. Register the result as an evidence record (type: benchmark results).

**CI wall-clock baseline:**

6. Read the current `.github/workflows/ci.yml` in full and record its job/leg structure: `changes`
   (diff classification), `workflow-lint` (actionlint), `unit` (no-DB: fmt/vet/lint/tidy/
   boundaries/test-unit/build), `gate` (matrix `[test, race]`, container, DB+S3 required, gated on
   `needs.changes.outputs.code == 'true'`), `gate-bench` (path-scoped on PRs via
   `needs.changes.outputs.bench`, unconditional on main push/nightly per the `schedule: cron: "17 3
   * * *"` trigger), `reference-smoke`, `coverage` (profile + floor).
7. Obtain per-leg wall-clock timing. Preferred source: the most recent accessible GitHub Actions run
   history for this workflow (if the execution environment has access, e.g. via `gh run list` /
   `gh run view` against the repository). Fallback, if run history is not accessible: a fresh local
   timed run of the container-equivalent targets (`make ci-container-test`, `make ci-container-race`,
   `make ci-container-bench`), explicitly labeled as a local approximation (differs from hosted-
   runner timing due to GHA cache warm/cold state and runner hardware).
8. Explicitly state in the evidence record which data source was used (hosted run history vs. local
   approximation) — this must not be left ambiguous.
9. Explicitly note the SD-01 (3-leg parallelization, toolbox image GHA-cached, docs-only skip, #23)
   and SD-02 (bench path-scoping on PRs, nightly schedule, `merge_group` support, #24) session-delta
   facts this baseline reflects, distinguishing the current post-#23/#24 shape from any pre-#23/#24
   serial-pipeline assumption.
10. Register the result as an evidence record (type: CI execution record).
11. Register both results' artifacts (bench-budget run output; CI-timing observations) per
    `impl/governance/artifact-policy.md`, and add entries to `../artifacts/index.md` and
    `../evidence/index.md`.

### Expected files or components affected

None in the source tree. This task is read/run-only against `bench-budgets.txt`,
`internal/tools/benchbudget`, and `.github/workflows/ci.yml`.

### Expected output

A bench-budget-baseline evidence record (confirming entry count and pass/fail per benchmark) and a
CI-wall-clock evidence record (job/leg structure plus per-leg timing, with data-source and
session-delta facts explicitly noted).

### Required artifacts

Bench-budget snapshot (raw `make bench-budget` output, `bench-budgets.txt` entry-count
confirmation); CI timing log (per-leg wall-clock observations).

### Required evidence

Bench-budget-baseline evidence record (type: benchmark results); CI-wall-clock evidence record
(type: CI execution record).

### Related acceptance criteria

AC-W00-E02-S001-03, AC-W00-E02-S001-04.

### Completion criteria

Both evidence records exist in `../evidence/index.md`: the bench-budget record states the confirmed
entry count and per-benchmark result; the CI-timing record states the job/leg structure, per-leg
timing (or explicitly-labeled approximation), the data source used, and the SD-01/SD-02 facts
reflected.

### Verification method

Per `../verification.md`'s AC-03 and AC-04 rows: re-run (or review the run of) `make bench-budget`
and confirm the entry count and pass/fail results match the evidence record; review the recorded
`ci.yml` job/leg structure against the actual file content at the cited commit SHA, and confirm the
timing data source is stated.

### Risks

- RISK-W00-003 (bench-budget baseline captured against stale pre-#25 budgets) — mitigated by
  explicit entry-count confirmation (step 2) before treating the capture as authoritative.
- RISK-W00-005 (CI/coverage baseline misdescribing current CI shape) — mitigated by reading the
  current `.github/workflows/ci.yml` directly at execution time (step 6), not from memory or the
  prior review documents.
- If GitHub Actions run history is inaccessible and the local-approximation fallback is used, the
  recorded wall-clock figures may not precisely match hosted-runner reality — mitigated by requiring
  explicit labeling of the data source (step 8), so downstream consumers of this baseline know its
  precision limits.

### Rollback or recovery considerations

Not applicable — this task performs no write to the repository or any persistent system.

## Implementation Record

*Per mandate §8.7. Not yet executed — no implementation claims are pre-populated.*

### What was actually implemented

Both captures executed: (a) `bench-budgets.txt` entry count confirmed at 43 (post-#25, no drift) then `make bench-budget` run — 43/43 OK, exit 0; timed window serialized with the one bench-active sibling per conductor instruction. (b) `.github/workflows/ci.yml` read directly at HEAD; per-leg wall-clock captured from hosted GitHub Actions run 29229288699 whose headSha equals the execution commit (preferred data source; local fallback not used).

### Components changed

None.

### Files changed

None in the committed source tree. Evidence/artifact files written under this story directory only.

### Interfaces introduced or changed

None.

### Configuration changes

None.

### Schema or migration changes

None.

### Security changes

None.

### Observability changes

None.

### Tests added or modified

None.

### Commits

None made by this task (verification-only; the conductor owns commits). Executed against `0a31186cada5c275a588c74081cf977adf346e61`.

### Pull requests

None.

### Implementation dates

2026-07-13.

### Technical debt introduced

None.

### Known limitations

Bench ns/op figures are a local snapshot with possible residual background load (serialized only against the bench-active sibling); the authoritative signal is the budget-gate pass. CI timing reflects one nightly run; artifact includes 8 recent runs for context.

### Follow-up items

None.

### Relationship to the approved plan

Followed `../plan.md` step 3 exactly; the "hosted run history vs local approximation" unresolved question resolved in favor of hosted history (stated in the evidence record).

## Verification Record

*Per mandate §8.8. Table below is planned before execution; fields after it are filled after
execution.*

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W00-E02-S001-03 | Run `make bench-budget`; confirm `bench-budgets.txt` entry count (expected 43) and per-benchmark pass/fail. | Local dev or CI container with Postgres reachable (`WOWAPI_REQUIRE_DB=1`). | Entry count confirmed; every budgeted benchmark reports a result (pass or specific violation). | Benchmark results | unassigned |
| AC-W00-E02-S001-04 | Read `.github/workflows/ci.yml`; record job/leg structure and per-leg wall-clock (hosted run history preferred, local-approximation fallback explicitly labeled). | GitHub Actions run-history access (preferred) or local Docker environment (fallback). | Job/leg structure recorded; timing recorded with data source stated; SD-01/SD-02 facts explicitly noted. | CI execution record | unassigned |

### Actual result

Bench: 43 entries confirmed; 43/43 benchmarks within budget, 0 violations, exit 0. CI: total 3m24s; legs — changes 6s, workflow-lint 25s, unit 2m14s, gate-test 1m39s, gate-race 3m04s, gate-bench 3m12s, reference-smoke 1m00s, coverage 2m54s; SD-01/SD-02 shape confirmed by direct file read; SD-03 visible in run history (pre-#25 run 14m23s vs post-#25 ≤3m50s).

### Pass or fail

**PASS (both captures).**

### Evidence identifier

EV-W00-E02-S001-003 (bench), EV-W00-E02-S001-004 (CI).

### Execution date

2026-07-13.

### Commit or revision

`0a31186cada5c275a588c74081cf977adf346e61` (main; also the observed CI run’s headSha).

### Environment

Local dev workstation, macOS (Darwin 25.5.0) arm64, go1.26.5; real Postgres 16 via compose; concurrent sibling load present. CI timing from hosted GitHub Actions runners (unaffected by local load).

### Reviewer

Unassigned (conductor review gate pending).

### Findings

No drift: entry count exactly 43 as expected (RISK-W00-003 closed for this capture); ci.yml shape matches SD-01/SD-02 exactly (RISK-W00-005 closed for this capture).

### Retest status

Not required — first capture, no failed run to retest.

### Final conclusion

AC-W00-E02-S001-03 and AC-W00-E02-S001-04 satisfied; both baselines registered.

## Deviations Record

*Per mandate §8.9. No deviations recorded yet.*

### Deviation ID

*Assign a stable deviation ID (`DEV-W00-E02-S001-T003-NNN`) if a deviation occurs.*

### Approved plan

*State what `../plan.md` said.*

### Actual implementation

*State what was actually implemented.*

### Reason

*State the reason for the deviation.*

### Impact

*State the impact of the deviation.*

### Risks

*State risks introduced by the deviation.*

### Approval

*State who approved the deviation and when.*

### Compensating controls

*State any compensating controls put in place.*

### Follow-up work

*State any follow-up work arising from the deviation.*
