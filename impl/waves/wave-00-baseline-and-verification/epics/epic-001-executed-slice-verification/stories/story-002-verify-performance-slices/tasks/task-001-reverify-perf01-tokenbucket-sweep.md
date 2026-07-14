---
id: W00-E01-S002-T001
type: task
title: Re-verify PERF-01 — token-bucket sweep fix
status: done
parent_story: W00-E01-S002
owner: W00E01S002 (wave-00 verification worker)
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on: []
acceptance_criteria: [AC-W00-E01-S002-01]
artifacts: [ART-W00-E01-S002-001]
evidence: [EV-W00-E01-S002-01]
---

# W00-E01-S002-T001 — Re-verify PERF-01 — token-bucket sweep fix

## Task Definition

### Task objective

Re-verify, at this task's execution commit SHA, that PERF-01's token-bucket sweep fix in
`kernel/httpx/ratelimit.go` still behaves as claimed EXECUTED by PLAN/REVIEW: the sweep recomputes
refill during sweep (not only on `Allow`), a hard cap on tracked entries is enforced, and the
backward-compatible 2-arg constructor is preserved — by re-running the named test files/commands and
registering mandate-§10 evidence.

### Parent story

W00-E01-S002 — Verify performance and benchmark-budget-gate slices at current HEAD.

### Owner

W00E01S002 (wave-00 verification worker).

### Status

`done` (executed 2026-07-13; both commands passed; evidence registered).

### Dependencies

None. This task does not depend on any other task in this story or any other story.

### Detailed work

1. Confirm the working tree is at a known commit SHA before running anything.
2. Run `go test ./kernel/httpx/... -race`. This exercises `kernel/httpx/ratelimit_test.go` (10k-entry
   eviction test and the race-condition test covering concurrent `Allow`/sweep access) and any other
   test in the package.
3. Run `make bench-budget`. This exercises `kernel/httpx/bench_test.go`'s sweep benchmarks against the
   budgeted values in `bench-budgets.txt`, using `internal/tools/benchbudget/main.go` as the enforcing
   tool.
4. Confirm, by reading `kernel/httpx/ratelimit.go` (or by the test assertions passing, whichever is the
   more direct signal), that: (a) the sweep path recomputes token refill during the sweep pass itself,
   not only when `Allow` is called on a given key; (b) a hard cap on the number of tracked
   rate-limiter entries is enforced, preventing unbounded growth; (c) the rate limiter's constructor
   still accepts the original 2-argument call signature (backward compatibility), in addition to any
   newer signature introduced by the fix.
5. Confirm `kernel/httpx/export_test.go` still exists and supports whatever internal-state
   test-only exports the above tests rely on (e.g., to inspect sweep internals for assertions).
6. Record the exact commands run, their exit codes, and relevant output excerpts.
7. Register evidence per `evidence-policy.md`'s required fields (evidence ID, type, story/task, AC
   proven, execution command, commit SHA, branch/tag, environment, tool versions, date/time, result,
   file/URI, reviewer).

### Expected files or components affected

None expected to change. Files read/exercised: `kernel/httpx/ratelimit.go`,
`kernel/httpx/ratelimit_test.go`, `kernel/httpx/bench_test.go`, `kernel/httpx/export_test.go`,
`bench-budgets.txt`, `internal/tools/benchbudget/main.go` (invoked via `make bench-budget`).

### Expected output

A pass/fail result for `go test ./kernel/httpx/... -race` and for `make bench-budget`, each with exit
code and relevant output captured; one registered evidence record proving AC-W00-E01-S002-01 (or a
`failed`-status evidence record plus an escalated finding, if either command does not pass).

### Required artifacts

None new. This task consumes existing repository files; it does not itself produce a new artifact
beyond the evidence record and its underlying log/output capture (see `artifacts/index.md`).

### Required evidence

One evidence record (planned ID `EV-W00-E01-S002-01`) covering both commands' results, per
`evidence-policy.md`.

### Related acceptance criteria

AC-W00-E01-S002-01.

### Completion criteria

This task is complete when: both commands have been run at a recorded commit SHA; their exit codes and
relevant output are captured; the PERF-01 behavioral claims (sweep-time refill recompute, hard cap,
backward-compatible constructor) have been confirmed true or a discrepancy has been recorded as a
finding; and an evidence record is registered in `evidence/index.md` referencing the exact commit SHA,
commands, and result.

### Verification method

Direct command execution and output inspection, as described in "Detailed work" above. No separate
verification step beyond running the commands themselves — this task's "verification" and
"implementation" (in this re-verification story's sense) are the same activity: running the commands
and confirming their result matches the expected PERF-01 behavior. See story-level `verification.md`
for the acceptance-criterion-level planned verification procedure.

### Risks

- `RISK-W00-001` (epic/wave-level) — PERF-01 could have regressed since the reviewed SHA `345e4ce`;
  this task's entire purpose is to surface that if true, not to assume it away.
- Flaky-race risk: `-race` runs can occasionally surface environment-specific timing issues not
  reproducible locally; if a failure occurs, note whether it reproduces on a second run before treating
  it as a confirmed regression (see wave-level `risks.md` RISK-W00-001 mitigation note on CI conditions
  not reproduced locally).

### Rollback or recovery considerations

This task makes no code change, so there is nothing to "roll back" in the usual sense. If PERF-01 fails
to re-verify (a regression is found), the correct response is procedural, not a code rollback within
this task: **this task must not be marked `done`, and this story must not move to `accepted`.** Per
`requirement-inventory.md`, PERF-01's target for its `INV` disposition is W00-E01-S002 itself — unlike
sibling story S001's findings (SEC-02/AR-04/AR-06), whose *unexecuted remainder* work targets a
distinct future-wave story, PERF-01 has no separate future-wave story to redirect a regression fix to.
The rollback/recovery path is therefore: open a new follow-up task within this same story
(`W00-E01-S002-T00N`, next available number) to investigate and, once a fix is designed, hand that fix
to whatever story is judged appropriate to implement it (a new story, since this one is
verification-only per its own scope) — but the investigation and finding-recording happens here, in
this story, not by silently deferring to an unrelated future wave.

## Implementation Record

Executed 2026-07-13 at commit `0a31186cada5c275a588c74081cf977adf346e61` (branch `main`).

### What was actually implemented

Ran `go test ./kernel/httpx/... -race` (exit 0, `ok github.com/qatoolist/wowapi/kernel/httpx
3.977s`, no data races) and `make bench-budget` (exit 0, all 43 budgeted benchmarks `OK`, including
`BenchmarkTokenBucketSweepAt10k` 451615.0 ns/op / 0 allocs and `BenchmarkTokenBucketSweepAt100k`
7636231.0 ns/op / 0 allocs — well within the post-#25 budgets 4500000/66000000). Confirmed in source
at the same SHA: (a) `sweep()` recomputes refill during the sweep pass itself using the same refill
formula `Allow` uses, projected from each bucket's last-touch time (`kernel/httpx/ratelimit.go`
~304-324); (b) hard cap enforced deterministically — at-cap new-key admission forces a synchronous
sweep and is rejected if still at cap (`WithHardCap`, ~275-285); (c) backward-compatible 2-arg
constructor `NewTokenBucket(ratePerSec, burst)` preserved (~228), options carried by the separate
`NewTokenBucketWithOptions`. `kernel/httpx/export_test.go` exists with `SweepForTest` and
`SeedBucketsForTest` supporting the tests/benchmarks.

### Components changed

None — verification-only, as planned.

### Files changed

No production file changed. Written (this story dir only): `artifacts/T001-go-test-kernel-httpx-race.log`,
`artifacts/T001-make-bench-budget.log`, `evidence/tests/EV-W00-E01-S002-01.md`, this record.

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

None — existing tests re-run only.

### Commits

None produced by this task (governance documents committed by the conductor's roll-up).

### Pull requests

None.

### Implementation dates

2026-07-13 (12:11:56–12:13:43 +05:30).

### Technical debt introduced

None.

### Known limitations

Benchmark timings were captured with concurrent sibling-worker load on the machine; budgets have
~10x headroom and all 43 passed, so the exit-0 result is robust, but the ns/op figures in the log
should not be reused as a quiet-machine baseline (that capture is W00-E02-S001's scope).

### Follow-up items

None — no regression surfaced.

### Relationship to the approved plan

Execution matched `plan.md` exactly (same commands, same order, environment prerequisites confirmed
first). One environmental caveat (concurrent load) recorded in story-level `deviations.md` DEV-01.

## Verification Record

### Planned verification procedure

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W00-E01-S002-01 | Run `go test ./kernel/httpx/... -race`, then `make bench-budget`; inspect exit codes and output | Local development machine or CI runner, Go toolchain per `go.mod`, no external DB/S3 dependency | Both commands exit 0; race detector reports no data races; bench-budget tool reports no budget violations for the sweep benchmarks | Test-execution log + bench-budget tool output log | unassigned |

### Actual result

`go test ./kernel/httpx/... -race` exit 0, no races; `make bench-budget` exit 0, 43/43 budgets OK.
PERF-01 behavioral claims (sweep-time refill recompute, hard cap, 2-arg constructor) confirmed in
source at the pinned SHA.

### Pass or fail

Pass.

### Evidence identifier

EV-W00-E01-S002-01 (`evidence/tests/EV-W00-E01-S002-01.md`).

### Execution date

2026-07-13.

### Commit or revision

`0a31186cada5c275a588c74081cf977adf346e61` (branch `main`).

### Environment

Local dev machine — macOS 26.5.2 (Darwin 25.5.0), arm64 Apple M3 Max; go1.26.5 darwin/arm64; GNU
Make 3.81. Concurrent load present (sibling W00 workers running tests during capture).

### Reviewer

Pending — conductor acceptance gate.

### Findings

None — no regression, no discrepancy.

### Retest status

Not required — first run passed; no flaky-race symptom observed.

### Final conclusion

AC-W00-E01-S002-01 satisfied at `0a31186`. Task `done`.

## Deviations Record

No deviation from the planned commands or method. Environmental caveat (concurrent machine load
during the benchmark run) recorded as DEV-01 in story-level `deviations.md`.
