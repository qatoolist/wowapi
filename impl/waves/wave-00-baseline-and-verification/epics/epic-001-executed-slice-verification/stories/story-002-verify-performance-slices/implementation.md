---
id: IMPL-W00-E01-S002
type: implementation-record
parent_story: W00-E01-S002
status: recorded
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Implementation record — W00-E01-S002

*This record aggregates the implementation reality of the story across all of its tasks. For this
story, "implementation" means running the re-verification commands named in `plan.md` and
registering evidence — no framework code changed.*

Executed 2026-07-13 at commit `0a31186cada5c275a588c74081cf977adf346e61` (branch `main`) — which is
itself PR #25's merge commit (the SD-03 fix commit).

## What was actually implemented

- **T001 (PERF-01)** — `go test ./kernel/httpx/... -race` exit 0 (no data races);
  `make bench-budget` exit 0 (all 43 budgeted benchmarks OK). Source confirmation at the same SHA:
  sweep recomputes refill during the sweep pass (`ratelimit.go` `sweep()`), hard cap enforced with
  deterministic at-cap rejection (`WithHardCap`), backward-compatible 2-arg `NewTokenBucket`
  preserved; `export_test.go` intact (`SweepForTest`, `SeedBucketsForTest`).
- **T002 (PERF-06 T1)** — test name reconfirmed (`TestMainMissingBenchmarkFails`,
  `coverage_test.go:103`); `go test ./internal/tools/benchbudget/... -run
  TestMainMissingBenchmarkFails -v` exit 0 (PASS). Fail-first revert-proof check executed against a
  scratch budgets file with a ghost entry (tracked `bench-budgets.txt` untouched): gate exited 1
  printing `FAIL  BenchmarkGhostMissingFromOutput  budgeted but not found in bench output`.
- **T003 (SD-03)** — `bench-budgets.txt` (root, sole match) has 43 non-comment, non-blank entries;
  reconciled with the tool-reported count (43 `OK` lines in the bench-budget run); file
  byte-identical to commit `0a31186`; 3-entry spot-check matches the #25 diff; live sweep
  measurements corroborate the post-#25 honest baseline. RISK-W00-003 does not materialize.

## Components changed

None — verification-only.

## Files changed

No production file changed (`git status` clean outside `impl/` except pre-existing untracked files
not owned by this story). Written, all inside this story directory: `artifacts/T001-*.log`,
`artifacts/T002-*.log`, `artifacts/T002-scratch-budgets-ghost.txt`,
`artifacts/T003-bench-budgets-inspection-note.md`, `evidence/tests/EV-W00-E01-S002-0{1,2}.md`,
`evidence/baselines/EV-W00-E01-S002-03.md`, plus updated index/record/task files.

## Interfaces introduced or changed

None.

## Configuration changes

None — `bench-budgets.txt` was read-only input.

## Schema or migration changes

Not applicable.

## Security changes

None.

## Observability changes

None.

## Tests added or modified

None — existing tests re-run only (`kernel/httpx` package tests, benchbudget coverage test).

## Commits

None produced by this story's execution (governance documents to be committed by the conductor's
roll-up).

## Pull requests

None.

## Implementation dates

2026-07-13 (12:11–12:16 +05:30).

## Technical debt introduced

None.

## Known limitations

Benchmark timings were captured with concurrent sibling-worker load; budgets carry ~10x headroom and
all 43 passed, so exit-code results are robust, but the ns/op figures should not be reused as a
quiet-machine baseline (that is W00-E02-S001's deliverable). See `deviations.md` DEV-01.

## Follow-up items

None — all three re-verifications passed; no regression, no investigation task needed.

## Relationship to the approved plan

Execution followed `plan.md`'s implementation sequence 1–7 exactly. Two recorded deviations, neither
affecting any AC's validity: DEV-01 (concurrent machine load during the benchmark run) and DEV-02
(fail-first check via ghost-entry scratch file rather than the task's literal "remove one budgeted
entry" phrasing). See `deviations.md`.
