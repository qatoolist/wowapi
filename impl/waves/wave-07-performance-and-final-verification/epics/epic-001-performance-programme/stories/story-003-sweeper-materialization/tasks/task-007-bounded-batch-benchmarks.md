---
id: W07-E01-S003-T007
type: task
title: Bounded-batch benchmarks and budget entries
status: complete
parent_story: W07-E01-S003
owner: W07-Scoping-Dispatch.W07E01S003
created_at: 2026-07-12
updated_at: 2026-07-14
depends_on:
  - W07-E01-S003-T001
  - W07-E01-S003-T002
  - W07-E01-S003-T003
acceptance_criteria:
  - AC-W07-E01-S003-07
artifacts:
  - ART-W07-E01-S003-007
evidence:
  - EV-W07-E01-S003-007
---

# W07-E01-S003-T007 — Bounded-batch benchmarks and budget entries

## Task Definition

### Task objective

Build bounded-batch benchmarks at due-row cardinality tiers, landing budget entries in the same PR, per PERF-06's own fail-closed policy.

### Parent story

W07-E01-S003

### Owner

W07-Scoping-Dispatch.W07E01S003

### Status

complete

### Dependencies

W07-E01-S003-T001, W07-E01-S003-T002, W07-E01-S003-T003 (benchmarks measure the bounded-batch mechanisms these tasks build).

### Detailed work

1. Build benchmarks at due-row cardinality tiers, parallel structure to PERF-01's own tiers.
2. Land budget entries in bench-budgets.txt in the same PR, per PERF-06's own fail-closed policy
   (T1, already EXECUTED at W00-E01-S002).

### Expected files or components affected

New benchmark files and bench-budgets.txt entries.

### Expected output

Bounded-batch benchmarks with same-PR budget entries.

### Required artifacts

ART-W07-E01-S003-007 (bounded-batch benchmarks + budget entries).

### Required evidence

EV-W07-E01-S003-007 (benchmark output + budget-entry confirmation).

### Related acceptance criteria

AC-W07-E01-S003-07

### Completion criteria

Benchmarks exist at cardinality tiers with same-PR budget entries.

### Verification method

Direct execution of the benchmark suite; inspection of bench-budgets.txt for the new entries.

### Risks

Medium — same orphan-benchmark risk as PERF-01 T6, per PLAN T7's own risk note (budget entries not landing in the same PR would leave the new benchmark unenforced).

### Rollback or recovery considerations

If a budget entry is found missing after this task's own PR merges, treat as a defect and add it immediately, not as a follow-up.

## Implementation Record

### What was actually implemented

`BenchmarkSweepSLABatch` seeds real PostgreSQL with 10, 1,000, and 100,000 due tasks and benchmarks
one bounded invocation. Same-change ceilings for all three sub-benchmarks were added to
`bench-budgets.txt` and enforced through the existing fail-closed budget parser.

### Components changed

Workflow performance test and repository benchmark budgets.

### Files changed

`kernel/workflow/sweeper_perf_test.go`, `bench-budgets.txt`,
`perf/results/perf-04-sweeper-before.txt`, `perf/results/perf-04-sweeper-after.txt`.

### Interfaces introduced or changed

None.

### Configuration changes

Adds three benchmark budget rows.

### Schema or migration changes

None.

### Security changes

Benchmark uses the real tenant transaction/RLS path.

### Observability changes

None.

### Tests added or modified

Three real-Postgres sub-benchmarks and focused budget-gate execution.

### Commits

Working tree based on entry SHA `733ef3e`.

### Pull requests

None.

### Implementation dates

2026-07-13.

### Technical debt introduced

None.

### Known limitations

Local darwin/arm64 numbers are relative evidence only; absolute ceilings remain conditional on DEC-Q9.

### Follow-up items

Re-run in the final DEC-Q9 reference environment when that decision is accepted.

### Relationship to the approved plan

Matches plan T7 and keeps absolute claims conditional.

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W07-E01-S003-07 | Three-tier benchmark piped to focused benchbudget subset | PostgreSQL 16.9, Apple M3 Max | Bounded allocations and all same-change ceilings pass | benchmark report | W07-Scoping-Dispatch.W07E01S003ReviewR |

### Actual result

PASS: 135,926,667/107,030,167/2,341,632,250 ns/op and 4,206/9,816/9,791 allocs/op
at 10/1k/100k respectively; all below registered ceilings.

### Pass or fail

PASS for same-host relative/budget evidence; absolute SLO not assessed.

### Evidence identifier

EV-W07-E01-S003-007.

### Execution date

2026-07-13.

### Commit or revision

Working tree based on entry SHA `733ef3e`.

### Environment

darwin/arm64, Apple M3 Max, Go 1.26.5, PostgreSQL 16.9 Docker service.

### Reviewer

W07-Scoping-Dispatch.W07E01S003ReviewR.

### Findings

No executor finding; DEC-Q9 is an explicit programme-level residual dependency.

### Retest status

Focused benchmark budget gate passed.

### Final conclusion

Same-host AC-07 benchmark/budget evidence is complete; absolute SLO acceptance remains conditional.
## Deviations Record

*No deviations recorded yet.*

### Deviation ID

*Not applicable.*

### Approved plan

*Not applicable.*

### Actual implementation

*Not applicable.*

### Reason

*Not applicable.*

### Impact

*Not applicable.*

### Risks

*Not applicable.*

### Approval

*Not applicable.*

### Compensating controls

*Not applicable.*

### Follow-up work

*Not applicable.*
