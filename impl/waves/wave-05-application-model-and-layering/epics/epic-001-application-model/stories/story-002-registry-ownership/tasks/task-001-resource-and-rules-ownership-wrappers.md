---
id: W05-E01-S002-T001
type: task
title: resource.Registry and rules.Registry owner-bound wrappers
status: todo
parent_story: W05-E01-S002
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on: []
acceptance_criteria:
  - AC-W05-E01-S002-01
artifacts:
  - ART-W05-E01-S002-001
  - ART-W05-E01-S002-002
evidence:
  - EV-W05-E01-S002-001
---

# W05-E01-S002-T001 — resource.Registry and rules.Registry owner-bound wrappers

## Task Definition

### Task objective

Build owner-bound registrar wrappers for `resource.Registry` and `rules.Registry`, using S001's
`Registrar` capability type, so that `ctx.Resources()` and its rules equivalent each expose a
registrar bound to the module's own identity, structurally, not by string comparison.

### Parent story

W05-E01-S002 — Owner-bound registry wrappers across all declaration classes.

### Owner

unassigned

### Status

todo

### Dependencies

None within this story (parallel-safe with T002, T003 — disjoint registry surfaces); depends on
W05-E01-S001 (T1, T2) at story scope.

### Detailed work

1. Implement the owner-bound wrapper for `resource.Registry`: `ctx.Resources()` returns a registrar
   bound to the calling module's own identity via S001's `Registrar` type.
2. Write the adversarial test `AR-01/resource_ownership_adversarial_test.go`: a cross-module claim
   attempt fails even with a matching key prefix.
3. Implement the owner-bound wrapper for `rules.Registry`, following the same shape as the resource
   wrapper.
4. Write the adversarial test `AR-01/rules_ownership_adversarial_test.go`, mirroring the resource
   test's structure for rule points.
5. Document both wrappers.

### Expected files or components affected

`kernel/resource`, `kernel/rules` (exact file paths TBD per `plan.md`).

### Expected output

Two owner-bound registry wrappers, each proven by a dedicated adversarial test.

### Required artifacts

ART-W05-E01-S002-001 (resource wrapper), ART-W05-E01-S002-002 (rules wrapper).

### Required evidence

EV-W05-E01-S002-001 (combined adversarial-test report for both).

### Related acceptance criteria

AC-W05-E01-S002-01.

### Completion criteria

Both adversarial tests pass: a cross-module claim attempt fails even with a matching key prefix, for
both `resource.Registry` and `rules.Registry`.

### Verification method

Direct execution of both named adversarial test files.

### Risks

Medium (per PLAN T3/T4's own risk column) — lower risk than T002/T003 given the established pattern
these two wrappers follow.

### Rollback or recovery considerations

If either adversarial test reveals a bypass (e.g. a matching key prefix incorrectly succeeding),
treat as a blocking defect — revert and fix before proceeding to T003's broader declaration-class
rollout, which follows this task's pattern.

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

*Not yet implemented.*

### Observability changes

*Not yet implemented.*

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
| AC-W05-E01-S002-01 | Run both named adversarial tests | Local dev or CI, Go toolchain | Cross-module claim attempt fails for both registries, even with a matching key prefix | adversarial-test report | unassigned |

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
