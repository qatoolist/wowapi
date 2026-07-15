---
id: DEV-W00-E01-S002
type: deviation-record
parent_story: W00-E01-S002
status: recorded
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Deviations record — W00-E01-S002

Executed 2026-07-13 at commit `0a31186cada5c275a588c74081cf977adf346e61`. Per mandate §8.9 and §2.6,
this file records divergence between the approved `plan.md` and what actually occurred; the approved
plan was not altered. Two deviations recorded; neither affects the validity of any acceptance
criterion's result.

## DEV-01 — Concurrent machine load during the benchmark-sensitive capture

- **Planned**: `plan.md` did not explicitly require a quiet machine, but benchmark timing captures
  implicitly assume representative conditions.
- **Actual**: the machine was shared with sibling W00 verification workers running test suites
  concurrently during the `make bench-budget` run (12:12–12:13 +05:30). Coordinated over IRC with
  W00E02S001 (the wave-level baseline-capture worker), which deliberately deferred its own
  timing-sensitive baseline run.
- **Impact**: none on AC-W00-E01-S002-01's validity — the AC is an exit-code/no-violation gate, and
  budgets carry ~10x headroom; all 43 entries passed with wide margins (e.g. SweepAt10k 451615 ns/op
  vs 4500000 budget). The recorded ns/op figures must not be reused as a quiet-machine performance
  baseline; that capture belongs to W00-E02-S001.
- **Disposition**: noted in every affected evidence record's environment field; accepted.

## DEV-02 — Fail-first revert-proof check method (T002)

- **Planned** (task T002 "Verification method", story `plan.md` "Testing strategy"): "temporarily
  remove one budgeted entry from a scratch/working copy of `bench-budgets.txt` ... run
  `make bench-budget`, confirm it now exits non-zero, then restore."
- **Actual**: executed as a ghost-entry check against a scratch budgets file
  (`artifacts/T002-scratch-budgets-ghost.txt`) containing one real benchmark present in the piped
  bench output plus one budgeted-but-absent benchmark, driven through the gate tool's documented
  stdin-pipe contract (`go test -bench=... | go run ./internal/tools/benchbudget <scratch>`) — the
  same pipe form `make bench-budget` itself uses. The tracked `bench-budgets.txt` was never
  modified (`git status --porcelain bench-budgets.txt` empty throughout).
- **Why**: (a) literally *removing* an entry from the budgets file relaxes the gate rather than
  triggering its fail-closed path — the failure mode under verification is "budgeted benchmark
  absent from bench output", which the ghost entry produces directly; (b) `make bench-budget` reads
  the hard-coded tracked path `bench-budgets.txt`, so exercising it with a mutated budget set would
  have required editing the tracked production file, prohibited for this wave's workers.
- **Impact**: the executed check is a strictly more faithful probe of PERF-06 T1's fail-closed
  contract, with observed gate exit 1 and the missing benchmark named in the report. No impact on
  AC-W00-E01-S002-02's validity.
- **Disposition**: recorded here and in task T002's records; accepted.

No other deviation occurred: commands, order, evidence fields, and the tool-count reconciliation all
matched `plan.md`.
