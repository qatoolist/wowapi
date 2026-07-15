---
id: W05-E02-S002-T001
type: task
title: Type-erased provider graph, zero-reflection proof
status: todo
parent_story: W05-E02-S002
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on: []
acceptance_criteria:
  - AC-W05-E02-S002-01
artifacts:
  - ART-W05-E02-S002-001
evidence:
  - EV-W05-E02-S002-001
---

# W05-E02-S002-T001 — Type-erased provider graph, zero-reflection proof

## Task Definition

### Task objective

Build a type-erased provider graph, type-erasing only at compile/boot time, with zero `reflect.*`
calls at `Resolve` time, proven by benchmark and static lint.

### Parent story

W05-E02-S002 — Zero-reflection provider graph, boot-time validation, and profile projection.

### Owner

unassigned

### Status

todo

### Dependencies

None within this story (S002's own first task); depends on W05-E02-S001 at story scope.

### Detailed work

1. Design the graph's internal representation for zero-reflection dispatch at `Resolve` time.
2. Implement the graph.
3. Write the benchmark (`AR-02/hotpath_no_reflection_bench.txt`'s producing benchmark).
4. Write a static lint check scanning for `reflect.*` calls on the `Resolve` code path.
5. Document the design.

### Expected files or components affected

A new provider-graph package (exact location TBD).

### Expected output

A working provider graph with zero hot-path reflection, proven by benchmark and lint.

### Required artifacts

ART-W05-E02-S002-001.

### Required evidence

EV-W05-E02-S002-001.

### Related acceptance criteria

AC-W05-E02-S002-01.

### Completion criteria

Benchmark and lint both confirm zero `reflect.*` calls at `Resolve` time.

### Verification method

Direct execution of the benchmark and lint check.

### Risks

Medium, per PLAN T3's own risk column — "naive implementations reflect per-call."

### Rollback or recovery considerations

If the benchmark/lint reveals hot-path reflection, redesign the dispatch mechanism before proceeding
to T002/T003.

## Implementation Record

*Not yet implemented.*

### What was actually implemented

*Not yet implemented.*

### Components changed

*Not yet implemented.*

### Files changed

*Not yet implemented.*

### Interfaces introduced or changed

*Not yet implemented.*

### Configuration changes

*Not yet implemented.*

### Schema or migration changes

*Not applicable.*

### Security changes

*Not applicable.*

### Observability changes

*Not applicable.*

### Tests added or modified

*Not yet implemented.*

### Commits

*Not yet implemented.*

### Pull requests

*Not yet implemented.*

### Implementation dates

*Not yet implemented.*

### Technical debt introduced

*None anticipated.*

### Known limitations

*Not yet implemented.*

### Follow-up items

*Not yet implemented.*

### Relationship to the approved plan

*Not yet implemented.*

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W05-E02-S002-01 | Run benchmark and lint | Local dev or CI, Go toolchain | Zero `reflect.*` calls at Resolve time | benchmark + lint report | unassigned |

### Actual result

*Not yet executed.*

### Pass or fail

*Not yet executed.*

### Evidence identifier

*Not yet executed.*

### Execution date

*Not yet executed.*

### Commit or revision

*Not yet executed.*

### Environment

*Not yet executed.*

### Reviewer

*Not yet executed.*

### Findings

*Not yet executed.*

### Retest status

*Not yet executed.*

### Final conclusion

*Not yet executed.*

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
