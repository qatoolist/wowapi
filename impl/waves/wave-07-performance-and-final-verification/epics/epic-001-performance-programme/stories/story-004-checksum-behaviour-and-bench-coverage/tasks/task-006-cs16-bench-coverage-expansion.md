---
id: W07-E01-S004-T006
type: task
title: CS-16 bench-coverage expansion (7 packages)
status: done
parent_story: W07-E01-S004
owner: W07-Scoping-Dispatch.W07E01S004
created_at: 2026-07-12
updated_at: 2026-07-16
depends_on: []
acceptance_criteria:
  - AC-W07-E01-S004-06
  - AC-W07-E01-S004-07
artifacts:
  - ART-W07-E01-S004-006
evidence:
  - EV-W07-E01-S004-006
  - EV-W07-E01-S004-007
---

# W07-E01-S004-T006 — CS-16 bench-coverage expansion (7 packages)

## Task Definition

### Task objective

Add benchmarks and bench-budget entries for kernel/database (tenant-tx open/commit), jobs (claim/finalize loop), outbox (relay dispatch batch), workflow, auth (token verify), mfa (TOTP derive), and httpclient (guarded dial), per MATRIX CS-16's exact target list.

### Parent story

W07-E01-S004

### Owner

W07-Scoping-Dispatch.W07E01S004

### Status

complete

### Dependencies

None — independent of T001-T005 (PERF-05's own scope targets storage, disjoint from these 7 kernel packages).

### Detailed work

1. Add a benchmark to kernel/database targeting tenant-tx open/commit.
2. Add a benchmark to jobs targeting the claim/finalize loop.
3. Add a benchmark to outbox targeting relay dispatch batch.
4. Add a benchmark to workflow targeting its own most request-relevant hot path (exact target TBD).
5. Add a benchmark to auth targeting token verify.
6. Add a benchmark to mfa targeting TOTP derive.
7. Add a benchmark to httpclient targeting guarded dial.
8. Add bench-budgets.txt entries for all 7, in the same PR as their benchmarks, per PERF-06's own
   fail-closed policy.
9. Extend BENCH_PKGS in the Makefile.

### Expected files or components affected

7 new *_bench_test.go files; bench-budgets.txt (7 new entries); Makefile (BENCH_PKGS extended).

### Expected output

BENCH_PKGS covers all 7 named packages, each with a passing bench-budget entry.

### Required artifacts

ART-W07-E01-S004-006 (7 new benchmark files + bench-budget entries).

### Required evidence

EV-W07-E01-S004-006 (benchmark report per package), EV-W07-E01-S004-007 (make bench-budget passing output).

### Related acceptance criteria

AC-W07-E01-S004-06, AC-W07-E01-S004-07.

### Completion criteria

All 7 packages have a benchmark targeting their own named hot path; make bench-budget exits 0.

### Verification method

Direct execution of each new benchmark and of make bench-budget.

### Risks

Low-medium — orphan-benchmark risk if budget entries don't land in the same PR, per PERF-06's own fail-closed policy (already EXECUTED, so this risk is now mechanically enforced, not merely advisory).

### Rollback or recovery considerations

If a budget entry is found missing after this task's own PR merges, it will now fail CI immediately (per PERF-06's own already-executed fail-closed fix) — add the missing entry to restore green.

## Implementation Record

Added one meaningful benchmark in each CS-16 package: tenant transaction
open/commit, job claim/finalize, ten-event outbox relay dispatch, workflow
definition validation, RS256 token verification, TOTP derivation, and guarded
dial classification. Extended `BENCH_PKGS` with exactly those seven packages and
added same-change time/allocation budgets for all seven names.

Files: seven package-local `*bench_test.go` files, `Makefile`, and
`bench-budgets.txt`. Implemented 2026-07-14, working tree based on `733ef3e`;
no production interface/schema/configuration change, PR, debt, coverage
expansion beyond CS-16, or plan deviation.
## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W07-E01-S004-06 | exact seven-name benchmark run with `-benchtime=10x -benchmem` | local PostgreSQL where needed | all seven real hot paths execute and report | benchmark report | independent story reviewer |
| AC-W07-E01-S004-07 | `DATABASE_URL=<local> WOWAPI_REQUIRE_DB=1 make bench-budget` | local PostgreSQL | budget gate exits 0 with all seven entries | budget report | independent story reviewer |

**PASS**, 2026-07-14, working tree based on `733ef3e`.
EV-W07-E01-S004-006 and -007 record all measurements and the passing combined
budget gate. Independent review: correct, confidence 1, no findings.
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
