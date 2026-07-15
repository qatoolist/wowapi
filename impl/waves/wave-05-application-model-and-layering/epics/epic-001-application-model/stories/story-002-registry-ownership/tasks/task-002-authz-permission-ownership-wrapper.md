---
id: W05-E01-S002-T002
type: task
title: authz.Registry permission-registration owner-bound wrapper
status: todo
parent_story: W05-E01-S002
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on: []
acceptance_criteria:
  - AC-W05-E01-S002-02
artifacts:
  - ART-W05-E01-S002-003
evidence:
  - EV-W05-E01-S002-002
---

# W05-E01-S002-T002 — authz.Registry permission-registration owner-bound wrapper

## Task Definition

### Task objective

Close the framework's widest zero-ownership-check registration gap: change
`authz.Registry.Register(p Permission)`, which today has no owner parameter at all, into an
owner-bound API that derives the module prefix from S001's `Registrar` capability type.

### Parent story

W05-E01-S002 — Owner-bound registry wrappers across all declaration classes.

### Owner

unassigned

### Status

todo

### Dependencies

None within this story (parallel-safe with T001, T003 — disjoint registry surface); depends on
W05-E01-S001 (T1, T2) at story scope.

### Detailed work

1. Design the new `authz.Registry` permission-registration API: derives the module prefix from the
   bound `Registrar` rather than accepting an arbitrary (or, today, entirely absent) owner
   parameter.
2. Implement the new API, replacing or wrapping the existing `Register(p Permission)` entry point.
3. Write the adversarial test `AR-01/authz_ownership_adversarial_test.go`: a cross-module permission
   claim is rejected at the registrar boundary.
4. Confirm this task's implementation is structured so S004's legacy adapter (built afterward) can
   route existing callers of the old `Register(p Permission)` signature through the new API without
   requiring every existing caller to be rewritten immediately — do not implement a breaking change
   with no compatibility path under consideration.
5. Document the new API and the security gap it closes.

### Expected files or components affected

`kernel/authz` (exact file paths TBD per `plan.md`).

### Expected output

An owner-bound `authz.Registry` permission-registration API, proven by the named adversarial test,
with the old signature's compatibility path considered (not yet implemented — that is S004's scope)
but not foreclosed by this task's own design.

### Required artifacts

ART-W05-E01-S002-003 (authz owner-bound wrapper).

### Required evidence

EV-W05-E01-S002-002 (adversarial-test report).

### Related acceptance criteria

AC-W05-E01-S002-02.

### Completion criteria

A cross-module permission claim is rejected at the registrar boundary — proven by the named
adversarial test passing.

### Verification method

Direct execution of `AR-01/authz_ownership_adversarial_test.go`.

### Risks

High (per PLAN T5's own risk column: "only registry with zero existing ownership check"). This is
the framework's actual security boundary at its widest point of exposure — see RISK-W05-001 in
epic-level `risks.md`.

### Rollback or recovery considerations

If the adversarial test reveals the new API still permits a cross-module claim (e.g. an incorrectly
derived module prefix), treat as a blocking security defect — do not ship this task until genuinely
closed.

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

*Not yet implemented — this task's entire purpose is a security change; recorded here once
implemented.*

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
| AC-W05-E01-S002-02 | Run `AR-01/authz_ownership_adversarial_test.go` | Local dev or CI, Go toolchain | Cross-module permission claim rejected at the registrar boundary | adversarial-test report | unassigned |

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
