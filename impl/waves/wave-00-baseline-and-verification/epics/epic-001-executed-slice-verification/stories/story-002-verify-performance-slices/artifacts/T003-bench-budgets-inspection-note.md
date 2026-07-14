# T003 inspection note — bench-budgets.txt vs #25 recalibration (SD-03)

- Date: 2026-07-13T12:16+05:30. Commit: 0a31186cada5c275a588c74081cf977adf346e61 (branch main).
- Path reconfirmed: `bench-budgets.txt` at repository root is the ONLY match for
  `git ls-files '*bench-budgets*'`.
- Entry count (non-comment, non-blank lines): **43** — matches the "43 budgeted entries" figure in
  requirement-inventory.md / RISK-W00-003.
- Tool-reported count reconciliation: the `make bench-budget` run (artifact
  T001-make-bench-budget.log) prints exactly **43** `OK    Benchmark...` lines — one per budgeted
  entry — agreeing with the manual line count. No discrepancy.
- Working tree vs #25: `git status --porcelain bench-budgets.txt` is empty and
  `git diff 0a31186 -- bench-budgets.txt` is empty. HEAD IS commit 0a31186 (PR #25), so the
  committed file is byte-identical to the post-#25 recalibrated state — no later uncaptured drift.
- Header consistency: header documents budgets "~10x the measured values on Apple M3 Max
  (2026-07-04)" and the sweep baselines carry the post-#25 annotation
  "(remeasured 2026-07-12: full-map sweep per iteration)" — consistent with the O(n^2)+empty-map
  fix, not the pre-fix skewed measurements.
- Spot-check (>= 3 entries) against `git show 0a31186 -- bench-budgets.txt`:
  | Entry | Current value | #25 diff | Match |
  |---|---|---|---|
  | BenchmarkTokenBucketSweepAt10k | 4500000 ns, 0 allocs | `+BenchmarkTokenBucketSweepAt10k 4500000 0` (was 250000) | yes |
  | BenchmarkTokenBucketSweepAt100k | 66000000 ns, 0 allocs | `+BenchmarkTokenBucketSweepAt100k 66000000 0` (was 2900000) | yes |
  | BenchmarkTokenBucketAllow | 300 ns, 0 allocs | unchanged by #25 (context line) | yes |
- Live corroboration: the 2026-07-13 bench run measured SweepAt10k 451615.0 ns/op and SweepAt100k
  7636231.0 ns/op (both 0 allocs) — the same order as #25's honest full-map measurements
  (433262 / 6605698 ns/op), i.e. ~10x under the recalibrated budgets, and wildly inconsistent with
  the pre-#25 empty-map artifacts (23492 / 280212 ns/op). The committed budgets are therefore the
  post-#25 honest baseline. RISK-W00-003 does not materialize.

Conclusion: SD-03 CONFIRMED — bench-budgets.txt reflects the #25 recalibration exactly.
