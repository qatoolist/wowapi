---
id: W00-E01-S002-T003
type: task
title: Confirm SD-03 — #25 bench-budget recalibration reflected in bench-budgets.txt
status: done
parent_story: W00-E01-S002
owner: W00E01S002 (wave-00 verification worker)
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on: []
acceptance_criteria: [AC-W00-E01-S002-03]
artifacts: [ART-W00-E01-S002-003]
evidence: [EV-W00-E01-S002-03]
---

# W00-E01-S002-T003 — Confirm SD-03 — #25 bench-budget recalibration reflected in bench-budgets.txt

## Task Definition

### Task objective

Confirm, at this task's execution commit SHA, that `bench-budgets.txt`'s entry count and values reflect
the O(n²)+empty-map sweep-benchmark measurement fix and budget recalibration introduced by PR #25
(commit `0a31186`), rather than a stale pre-#25 baseline — closing out wave-level risk RISK-W00-003.

### Parent story

W00-E01-S002 — Verify performance and benchmark-budget-gate slices at current HEAD.

### Owner

W00E01S002 (wave-00 verification worker).

### Status

`done` (executed 2026-07-13; 43 entries confirmed post-#25; evidence registered).

### Dependencies

None. This task is independent of T001 and T002, though it is most meaningfully executed alongside or
after T001 since T001's `make bench-budget` run is the mechanism that would most directly surface a
budget-file problem in practice (see `plan.md` "Dependencies" for why it is nonetheless kept as its own
task rather than folded into T001: it is a data/budget-file-content confirmation, distinct from T001's
code-behavior confirmation).

### Detailed work

1. Confirm the working tree is at a known commit SHA before inspecting anything.
2. Locate `bench-budgets.txt`. At drafting time this was confirmed to exist at the repository root
   (sibling to `go.mod`/`Makefile`). **Reconfirm this path has not moved** at execution time before
   proceeding — do not assume the drafting-time location is still accurate without a quick check.
3. Count the file's budget-entry lines: non-comment (not starting with `#`), non-blank lines. At
   drafting time this count was 43, matching the "43 budgeted entries" figure in
   `requirement-inventory.md` and wave-level `risks.md` RISK-W00-003.
4. If `make bench-budget` (or the `benchbudget` tool run standalone) reports its own parsed
   budget-entry count in its output, prefer that tool-reported count as a stronger source of truth than
   a manual line count, and reconcile any discrepancy between the two counting methods before
   concluding.
5. Read `bench-budgets.txt`'s header comment block, which documents the recalibration provenance (e.g.
   "Budgets are set at ~10x the measured values on Apple M3 Max (2026-07-04)... Measured baseline
   (Apple M3 Max, Go 1.26, darwin/arm64)"). Confirm this header is consistent with a post-#25
   recalibration (i.e., it does not describe pre-fix, O(n²)-inflated, or empty-map-skewed measurements).
6. Spot-check at least 3 individual budget-entry values (e.g. the sweep-related benchmarks most
   directly affected by the #25 fix, such as any `BenchmarkSweep*` or `BenchmarkTokenBucket*` entries)
   against commit `0a31186`'s actual diff (`git show 0a31186 -- bench-budgets.txt` or equivalent) to
   confirm the currently-committed values match what #25 introduced, and have not silently drifted in a
   later, uncaptured commit.
7. Record the exact entry count, the spot-checked entries and their values, and the git evidence
   (commit `0a31186` diff excerpt) tying the current file content to the #25 fix.
8. Register evidence per `evidence-policy.md`'s required fields.

### Expected files or components affected

None expected to change. File inspected: `bench-budgets.txt` (repository root, path to be reconfirmed
at execution time).

### Expected output

A confirmed entry count (expected 43, or an explicitly escalated discrepancy if different) and a
spot-check note showing at least 3 entries' values are consistent with the post-#25 recalibration; one
registered evidence record proving AC-W00-E01-S002-03.

### Required artifacts

None new. This task consumes an existing repository file; see `artifacts/index.md`.

### Required evidence

One evidence record (planned ID `EV-W00-E01-S002-03`), per `evidence-policy.md`.

### Related acceptance criteria

AC-W00-E01-S002-03.

### Completion criteria

This task is complete when: the file's location has been reconfirmed; its entry count has been counted
and compared against the expected 43-entry figure; at least 3 entries have been spot-checked against
commit `0a31186`'s diff; the result (match or discrepancy) has been recorded; and an evidence record is
registered in `evidence/index.md`.

### Verification method

Direct file inspection and `git show`/`git diff` comparison against commit `0a31186`, as described in
"Detailed work" above. This is an inspection-based verification, not a test-execution-based one — no
test command produces a pass/fail signal for this task; the signal is the entry count and value
comparison itself.

### Risks

- `RISK-W00-003` (epic/wave-level) — this task exists specifically to close out this risk: bench-budget
  baseline captured against stale, pre-#25 values, which would cause every later wave's
  "improvement-over-baseline" performance claim to be measured against the wrong starting point.
- Counting-method ambiguity risk: "43 budgeted entries" could in principle be counted differently
  (e.g., including or excluding certain comment-adjacent lines); step 4 of "Detailed work" mitigates
  this by preferring the tool's own reported count where available.

### Rollback or recovery considerations

No code change is made by this task, so there is nothing to roll back. If the entry count or values do
NOT match the post-#25 state — i.e., if `bench-budgets.txt` still reflects a stale pre-#25 baseline, or
has drifted from what #25 introduced in some other uncaptured way — **this is itself the RISK-W00-003
scenario materializing.** This task must not be marked `done` in that case. The finding must be
escalated (recorded as a failed-evidence record per `evidence-policy.md`, and flagged to the epic/wave
risk owner), not silently accepted as an acceptable baseline. As with T001/T002, the recovery path is a
follow-up investigation task opened within this story — PERF-01's evidence basis (which SD-03 directly
concerns) targets W00-E01-S002 itself, so there is no separate future-wave story to redirect this
finding to.

## Implementation Record

Executed 2026-07-13 at commit `0a31186cada5c275a588c74081cf977adf346e61` (branch `main`) — which is
itself PR #25's merge commit, making the comparison against the #25 state exact by construction.

### What was actually implemented

1. Path reconfirmed: `git ls-files '*bench-budgets*'` returns only `bench-budgets.txt` at the
   repository root.
2. Manual entry count: **43** non-comment, non-blank lines — matches the expected figure.
3. Tool-reported count reconciled: T001's `make bench-budget` output contains exactly **43**
   `OK    Benchmark...` result lines (one per budgeted entry); the two counting methods agree.
4. Drift check: `git status --porcelain bench-budgets.txt` empty and
   `git diff 0a31186 -- bench-budgets.txt` empty — the file is byte-identical to the post-#25 state;
   no later uncaptured drift.
5. Header consistency: budgets "~10x the measured values on Apple M3 Max (2026-07-04)"; the sweep
   baselines carry "(remeasured 2026-07-12: full-map sweep per iteration)" — post-#25 provenance,
   not pre-fix O(n²)/empty-map-skewed measurements.
6. Spot-check (3 entries) vs `git show 0a31186 -- bench-budgets.txt`:
   `BenchmarkTokenBucketSweepAt10k 4500000 0` (was 250000), `BenchmarkTokenBucketSweepAt100k
   66000000 0` (was 2900000), `BenchmarkTokenBucketAllow 300 0` (unchanged context) — all match.
   Live corroboration: today's measured sweep values (451615 / 7636231 ns/op) sit at the same order
   as #25's honest baselines (433262 / 6605698 ns/op) and are inconsistent with the pre-#25
   empty-map artifacts (23492 / 280212 ns/op).

RISK-W00-003 (stale pre-#25 baseline) does **not** materialize.

### Components changed

None — inspection-only, as planned.

### Files changed

No production file changed. Written (this story dir only):
`artifacts/T003-bench-budgets-inspection-note.md`, `evidence/baselines/EV-W00-E01-S002-03.md`,
this record.

### Interfaces introduced or changed

None.

### Configuration changes

None — `bench-budgets.txt` read-only.

### Schema or migration changes

Not applicable.

### Security changes

None.

### Observability changes

None.

### Tests added or modified

None — inspection-based task.

### Commits

None produced by this task.

### Pull requests

None.

### Implementation dates

2026-07-13 (~12:16 +05:30).

### Technical debt introduced

None.

### Known limitations

None material. HEAD being the #25 commit itself makes the diff comparison trivially exact — a later
re-verification at a newer HEAD would need the full `git diff 0a31186..HEAD -- bench-budgets.txt`
check this task performed via `git diff 0a31186`.

### Follow-up items

None.

### Relationship to the approved plan

Matched `plan.md` exactly, including the preference for the tool-reported entry count where
available (step 4 of "Detailed work"): both counting methods were used and reconciled.

## Verification Record

### Planned verification procedure

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W00-E01-S002-03 | Inspect `bench-budgets.txt` (repository root; path reconfirmed at execution time); count non-comment, non-blank entry lines; spot-check at least 3 entries against `git show 0a31186 -- bench-budgets.txt` | Local development machine or CI runner; text inspection and `git` only, no test execution required | 43 entries (or an explicitly escalated discrepancy if different, per RISK-W00-003); spot-checked values consistent with the post-#25 O(n²)+empty-map-fix recalibration, not stale pre-#25 numbers | Entry-count and spot-check inspection note (plus `git show` excerpt) | unassigned |

### Actual result

43 entries (manual count and tool-reported count agree); file byte-identical to commit `0a31186`
(the #25 recalibration); 3-entry spot-check matches the #25 diff; header provenance post-#25.

### Pass or fail

Pass.

### Evidence identifier

EV-W00-E01-S002-03 (`evidence/baselines/EV-W00-E01-S002-03.md`).

### Execution date

2026-07-13.

### Commit or revision

`0a31186cada5c275a588c74081cf977adf346e61` (branch `main`).

### Environment

Local dev machine — macOS 26.5.2 (Darwin 25.5.0), arm64 Apple M3 Max; git 2.55.0. Inspection-based,
load-insensitive.

### Reviewer

Pending — conductor acceptance gate.

### Findings

None — no discrepancy; RISK-W00-003 closed out for this story.

### Retest status

Not required.

### Final conclusion

AC-W00-E01-S002-03 satisfied at `0a31186`. Task `done`.

## Deviations Record

No deviation from the planned inspection method.
