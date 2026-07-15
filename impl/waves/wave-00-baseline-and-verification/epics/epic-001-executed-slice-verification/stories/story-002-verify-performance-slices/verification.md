---
id: VER-W00-E01-S002
type: verification-record
parent_story: W00-E01-S002
status: recorded
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Verification record — W00-E01-S002

## Planned verification procedure

Per mandate §8.8. One row per acceptance criterion for this story. This table describes the PLANNED
PROCEDURE only — no verification has been executed yet, and no result below should be read as a claim
that any command has passed or failed.

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W00-E01-S002-01 | Re-run `go test ./kernel/httpx/... -race` (confirms the 10k-eviction and race-condition assertions in `ratelimit_test.go` pass) and `make bench-budget` (confirms the sweep benchmarks in `bench_test.go` satisfy `bench-budgets.txt`'s budgets, tool exits 0) | Local development machine or CI runner, standard Go toolchain per `go.mod`, no external database or S3 dependency required | Both commands exit 0; `go test -race` reports no data races; `make bench-budget` reports no budget violations. Confirms PERF-01: the sweep recomputes refill during sweep (not only on `Allow`), a hard cap on tracked entries is enforced, and the backward-compatible 2-arg constructor is preserved | Test-execution log (`go test` output) + bench-budget tool output log | unassigned |
| AC-W00-E01-S002-02 | Re-run the benchbudget package's subprocess-based missing-benchmark coverage test — candidate `go test ./internal/tools/benchbudget/... -run TestMainMissingBenchmarkFails -v` (exact `-run` pattern must be reconfirmed against `internal/tools/benchbudget/coverage_test.go`'s actual test function name at execution time, per mandate §8.5's "do not invent precise... where the repository does not yet provide enough information" — see `plan.md` "Unresolved questions"). Additionally, the task's verification method DESCRIBES (does not execute during this planning-document pass) a manual fail-first revert-proof check: temporarily remove one entry from a scratch copy of `bench-budgets.txt`, confirm `make bench-budget` now exits non-zero, then restore the original file unmodified | Local development machine or CI runner, standard Go toolchain, no external dependency required | The subprocess test exits 0 (i.e., the exit-1 assertion it makes internally passes), confirming PERF-06 T1: a budgeted-but-missing benchmark causes `internal/tools/benchbudget/main.go` to exit non-zero with an explanatory message, not a silent WARN-and-continue | Test-execution log (`go test` output) | unassigned |
| AC-W00-E01-S002-03 | Inspect `bench-budgets.txt` (repository root; path to be reconfirmed at execution time per `plan.md` "Unresolved questions") — count non-comment, non-blank budget-entry lines, and spot-check at least 3 entries' values against the #25 O(n²)+empty-map-fix recalibration description in the file's own header comment and commit `0a31186`'s change description | Local development machine or CI runner; text inspection only, no test execution required | Entry count is 43 (per `requirement-inventory.md` and wave-level `risks.md` RISK-W00-003) — or, if the actual count differs, that discrepancy is itself escalated as a RISK-W00-003 materialization rather than silently accepted. Spot-checked values are consistent with post-#25 state (remeasured baselines, not stale pre-#25 numbers) | Entry-count and spot-check inspection note | unassigned |

## Post-execution record

Executed 2026-07-13. Every result below was actually observed; raw outputs are stored under this
story's `artifacts/` and referenced by the evidence records.

### Per-AC results

| Acceptance criterion | Actual result | Pass/fail | Evidence ID |
|---|---|---|---|
| AC-W00-E01-S002-01 | `go test ./kernel/httpx/... -race` exit 0 (`ok github.com/qatoolist/wowapi/kernel/httpx 3.977s`, no data races); `make bench-budget` exit 0, all 43 budgeted benchmarks `OK` (SweepAt10k 451615 ns/op / budget 4500000; SweepAt100k 7636231 ns/op / budget 66000000; both 0 allocs). PERF-01 claims confirmed in source: sweep-time refill recompute (`ratelimit.go` `sweep()`), hard cap (`WithHardCap` + at-cap rejection), 2-arg `NewTokenBucket` preserved | **Pass** | EV-W00-E01-S002-01 |
| AC-W00-E01-S002-02 | Test name reconfirmed as `TestMainMissingBenchmarkFails` (`coverage_test.go:103`); `go test ./internal/tools/benchbudget/... -run TestMainMissingBenchmarkFails -v` exit 0, `--- PASS (0.01s)` — subprocess exits 1 with explanatory message. Fail-first revert-proof check (ghost entry in scratch budgets file; tracked file untouched): gate exit 1, `FAIL BenchmarkGhostMissingFromOutput budgeted but not found in bench output` | **Pass** | EV-W00-E01-S002-02 |
| AC-W00-E01-S002-03 | 43 non-comment, non-blank entries (manual count) = 43 tool-reported `OK` lines; `bench-budgets.txt` byte-identical to commit `0a31186` (empty diff, clean tree); 3-entry spot-check matches the #25 diff (SweepAt10k 4500000, SweepAt100k 66000000, TokenBucketAllow 300); header carries post-#25 "remeasured 2026-07-12" provenance | **Pass** | EV-W00-E01-S002-03 |

### Actual result

All three acceptance criteria satisfied; no regression found in PERF-01, PERF-06 T1, or the SD-03
budget baseline.

### Pass or fail

Pass — 3/3 ACs.

### Evidence identifier

EV-W00-E01-S002-01 (`evidence/tests/`), EV-W00-E01-S002-02 (`evidence/tests/`),
EV-W00-E01-S002-03 (`evidence/baselines/`). Indexed in `evidence/index.md`.

### Execution date

2026-07-13 (12:11–12:16 +05:30).

### Commit or revision

`0a31186cada5c275a588c74081cf977adf346e61` (branch `main`) — the story's closing commit SHA, itself
PR #25's merge commit. Pinned per `evidence-policy.md`; not a moving-target reference.

### Environment

Local development machine — macOS 26.5.2 (Darwin 25.5.0), arm64 (Apple Silicon M3 Max); go1.26.5
darwin/arm64; GNU Make 3.81; git 2.55.0. No external DB/S3 dependency was needed, as planned.
Machine shared with sibling W00 verification workers running tests concurrently — concurrent load
present during the benchmark-timing-sensitive capture (see `deviations.md` DEV-01).

### Reviewer

Pending — conductor acceptance gate (per mandate §7 the story is not self-marked `accepted`).

### Findings

None. The one open plan question (exact `-run` pattern) resolved to the plan's candidate name
unchanged; the 43-entry count reconciled across both counting methods.

### Retest status

Not required — all first runs passed; no flaky-race symptom observed.

### Final conclusion

All three ACs proven at `0a31186cada5c275a588c74081cf977adf346e61`. Story moved to
`ready-for-review` awaiting the acceptance authority's independent confirmation.
