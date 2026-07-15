---
id: W00-E01-S002-T002
type: task
title: Re-verify PERF-06 T1 — fail-closed benchbudget missing-benchmark gate
status: done
parent_story: W00-E01-S002
owner: W00E01S002 (wave-00 verification worker)
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on: []
acceptance_criteria: [AC-W00-E01-S002-02]
artifacts: [ART-W00-E01-S002-002]
evidence: [EV-W00-E01-S002-02]
---

# W00-E01-S002-T002 — Re-verify PERF-06 T1 — fail-closed benchbudget missing-benchmark gate

## Task Definition

### Task objective

Re-verify, at this task's execution commit SHA, that PERF-06 T1's fail-closed behavior in
`internal/tools/benchbudget/main.go` still holds: when a benchmark is budgeted in `bench-budgets.txt`
but absent from the actual bench output, the tool exits non-zero (a tracked build failure) rather than
warning and continuing — by re-running the relevant coverage test and registering mandate-§10 evidence.

### Parent story

W00-E01-S002 — Verify performance and benchmark-budget-gate slices at current HEAD.

### Owner

W00E01S002 (wave-00 verification worker).

### Status

`done` (executed 2026-07-13; test passed and fail-first gate check confirmed exit 1; evidence registered).

### Dependencies

None.

### Detailed work

1. Confirm the working tree is at a known commit SHA before running anything.
2. Open `internal/tools/benchbudget/coverage_test.go` and confirm the exact current name of the test
   function that exercises the missing-benchmark-fails-the-build path. At drafting time, the strongest
   candidate is `TestMainMissingBenchmarkFails`, whose doc comment reads: "re-exercises the exit-path
   harness below to confirm a budgeted-but-absent benchmark causes a real CI failure (exit 1), not just
   a warning. See PERF-06 T1." **Do not assume this name is still accurate without re-checking the file
   at execution time** — if the repository has changed since this task was drafted, use whatever the
   actual current test name is instead, per mandate §8.5's instruction not to invent precision the
   repository does not yet support.
3. Run the confirmed test, e.g. `go test ./internal/tools/benchbudget/... -run
   TestMainMissingBenchmarkFails -v` (adjust the `-run` pattern to match step 2's finding).
4. Confirm the test's internal assertions pass: specifically, that the subprocess it spawns (which
   simulates a budgeted-but-missing `BenchmarkGhost`) exits with code 1, and that its output contains
   both the missing benchmark's name and an explanatory message (e.g. "budgeted but not found in bench
   output") rather than a silent pass.
5. Separately, **describe but do not execute** the fail-first revert-proof manual check as part of this
   task's documented verification method (see "Verification method" below) — it belongs to actual task
   execution, not to this planning-document pass.
6. Confirm `internal/tools/benchbudget/main.go`'s missing-benchmark code path is structured as a
   tracked, appended violation (contributing to a non-zero exit) rather than a `WARN`-and-continue log
   line — by reading the relevant function and/or relying on the passing test as the behavioral proof.
7. Record the exact command run, its exit code, and relevant output excerpts (the subprocess test
   itself reports "12/12 pass" or similar per the story's context — confirm the actual current pass
   count rather than assuming the story-drafting-time figure still holds).
8. Register evidence per `evidence-policy.md`'s required fields.

### Expected files or components affected

None expected to change. Files read/exercised: `internal/tools/benchbudget/main.go`,
`internal/tools/benchbudget/coverage_test.go`, `internal/tools/benchbudget/main_test.go` (if relevant
to the same coverage run).

### Expected output

A pass/fail result for the confirmed benchbudget coverage test (subprocess exit-1 assertion), with
exit code and relevant output captured; one registered evidence record proving
AC-W00-E01-S002-02 (or a `failed`-status evidence record plus an escalated finding, if the test does
not pass).

### Required artifacts

None new. This task consumes existing repository files; see `artifacts/index.md`.

### Required evidence

One evidence record (planned ID `EV-W00-E01-S002-02`), per `evidence-policy.md`.

### Related acceptance criteria

AC-W00-E01-S002-02.

### Completion criteria

This task is complete when: the exact test name/`-run` pattern has been confirmed against the current
repository state; the test has been run at a recorded commit SHA; its exit code and relevant output are
captured; the PERF-06 T1 fail-closed behavioral claim has been confirmed true or a discrepancy has been
recorded as a finding; and an evidence record is registered in `evidence/index.md`.

### Verification method

Direct command execution and output inspection (see "Detailed work" steps 3-4). Additionally, this
task's verification method **describes** — as a documented procedure for whoever executes this task,
not as something performed during this planning-document creation — a manual fail-first revert-proof
check: temporarily remove one budgeted entry from a scratch/working copy of `bench-budgets.txt` (not
the tracked file, or restored immediately if the tracked file is used), run `make bench-budget`, confirm
it now exits non-zero, and then restore the original file exactly. This revert-proof check is a
stronger confirmation that the gate is genuinely fail-closed (not merely passing today because nothing
is currently missing) — but it is explicitly out of scope for the planning-document-creation pass that
produced this task file; it is scoped to this task's actual execution.

### Risks

- `RISK-W00-001` (epic/wave-level) — PERF-06 T1 could have regressed since the reviewed SHA.
- Test-name drift risk: if `TestMainMissingBenchmarkFails` has been renamed, split, or removed since
  drafting, step 2 of "Detailed work" is the designed mitigation — reconfirm before running.

### Rollback or recovery considerations

No code change is made by this task, so there is nothing to roll back. If PERF-06 T1 fails to
re-verify, the same story-internal pattern as task T001 applies: **this task must not be marked `done`,
and this story must not move to `accepted`.** Per `requirement-inventory.md`, PERF-06's `INV`
disposition targets W00-E01-S002 itself for its T1 portion (T3/T4 fuzz scope is the only part targeted
elsewhere, at W07-E02-S002, and that is out of this task's scope entirely). The recovery path is to open
a follow-up investigation task within this same story, not to redirect to a different future-wave
story.

## Implementation Record

Executed 2026-07-13 at commit `0a31186cada5c275a588c74081cf977adf346e61` (branch `main`).

### What was actually implemented

1. Reconfirmed the test name against `internal/tools/benchbudget/coverage_test.go` at this SHA:
   `TestMainMissingBenchmarkFails` exists at line 103, doc comment ending "See PERF-06 T1" — the
   plan's candidate name was accurate; no drift.
2. Ran `go test ./internal/tools/benchbudget/... -run TestMainMissingBenchmarkFails -v` — exit 0,
   `--- PASS: TestMainMissingBenchmarkFails (0.01s)`; the spawned subprocess (budgeted-but-absent
   `BenchmarkGhost`) exited 1 with the explanatory message, per the test's internal assertions.
3. Executed the fail-first revert-proof check WITHOUT touching the tracked `bench-budgets.txt`:
   wrote a scratch budgets file (`artifacts/T002-scratch-budgets-ghost.txt`) containing one real
   benchmark present in the piped output (`BenchmarkConfigDefaults`, generous budget) plus one ghost
   entry (`BenchmarkGhostMissingFromOutput`), then ran
   `go test -bench=BenchmarkConfigDefaults -benchmem -run=^$ ./kernel/config/... | go run
   ./internal/tools/benchbudget <scratch>` — the gate exited **1** printing
   `FAIL  BenchmarkGhostMissingFromOutput  budgeted but not found in bench output`, while the real
   in-budget benchmark printed `OK`. Fail-closed behavior confirmed live at this SHA.
4. Source confirmation: `internal/tools/benchbudget/main.go`'s package doc and violation handling
   treat a budgeted-but-absent benchmark as a tracked violation contributing to a non-zero exit
   ("the tool fails closed"), not a WARN-and-continue log line.

### Components changed

None — verification-only, as planned.

### Files changed

No production file changed; tracked `bench-budgets.txt` untouched (`git status --porcelain
bench-budgets.txt` empty after execution). Written (this story dir only):
`artifacts/T002-go-test-benchbudget-missingbenchmark.log`, `artifacts/T002-failfirst-ghost-check.log`,
`artifacts/T002-scratch-budgets-ghost.txt`, `evidence/tests/EV-W00-E01-S002-02.md`, this record.

### Interfaces introduced or changed

None.

### Configuration changes

None.

### Schema or migration changes

Not applicable.

### Security changes

None.

### Observability changes

None.

### Tests added or modified

None — existing test re-run only; the fail-first check used a scratch data file, not a new test.

### Commits

None produced by this task.

### Pull requests

None.

### Implementation dates

2026-07-13 (12:14:40–12:15:54 +05:30).

### Technical debt introduced

None.

### Known limitations

The fail-first check drove the gate tool directly over its documented stdin-pipe contract (the same
form `make bench-budget` uses) with a single-package bench stream, rather than re-running the full
multi-package `make bench-budget` against a mutated tracked file — see story `deviations.md` DEV-02
for why this is the safer, equivalent form.

### Follow-up items

None — no regression surfaced.

### Relationship to the approved plan

Matched `plan.md`: test name reconfirmed before running, exact `-run` pattern used, fail-first
revert-proof check executed in the execution phase as scoped. One method deviation (ghost-entry
scratch file instead of the task's literal "remove one budgeted entry" phrasing) recorded as DEV-02
in story-level `deviations.md`.

## Verification Record

### Planned verification procedure

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W00-E01-S002-02 | Confirm exact test name against `internal/tools/benchbudget/coverage_test.go` (candidate: `TestMainMissingBenchmarkFails`), then run `go test ./internal/tools/benchbudget/... -run <confirmed-name> -v`. Fail-first revert-proof check described, not executed, at this planning stage | Local development machine or CI runner, Go toolchain per `go.mod`, no external dependency | Test exits 0, confirming the subprocess it spawns exits 1 with an explanatory missing-benchmark message — i.e. PERF-06 T1's fail-closed gate is intact | Test-execution log | unassigned |

### Actual result

`go test ./internal/tools/benchbudget/... -run TestMainMissingBenchmarkFails -v` exit 0 (PASS);
fail-first ghost-entry gate check exit 1 with `budgeted but not found in bench output` naming the
missing benchmark. PERF-06 T1 fail-closed contract intact.

### Pass or fail

Pass.

### Evidence identifier

EV-W00-E01-S002-02 (`evidence/tests/EV-W00-E01-S002-02.md`).

### Execution date

2026-07-13.

### Commit or revision

`0a31186cada5c275a588c74081cf977adf346e61` (branch `main`).

### Environment

Local dev machine — macOS 26.5.2 (Darwin 25.5.0), arm64 Apple M3 Max; go1.26.5 darwin/arm64; GNU
Make 3.81. Concurrent sibling-worker load present (exit-code assertions load-insensitive).

### Reviewer

Pending — conductor acceptance gate.

### Findings

None — no regression; test name had not drifted.

### Retest status

Not required — first run passed.

### Final conclusion

AC-W00-E01-S002-02 satisfied at `0a31186`. Task `done`.

## Deviations Record

One method deviation: the fail-first revert-proof check was executed against a scratch budgets file
with a ghost entry (budgeted-but-absent benchmark) fed through the tool's documented stdin-pipe
contract, instead of the task's literal "remove one budgeted entry from a copy of
`bench-budgets.txt` and run `make bench-budget`" phrasing — removing an entry would relax the gate
rather than trigger its fail-closed path, and mutating the tracked file was prohibited during this
wave. Recorded as DEV-02 in story-level `deviations.md`.
