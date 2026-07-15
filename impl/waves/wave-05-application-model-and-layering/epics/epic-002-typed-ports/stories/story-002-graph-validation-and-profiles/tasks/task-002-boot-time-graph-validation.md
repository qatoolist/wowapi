---
id: W05-E02-S002-T002
type: task
title: Boot-time graph validation
status: todo
parent_story: W05-E02-S002
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on:
  - W05-E02-S002-T001
acceptance_criteria:
  - AC-W05-E02-S002-02
artifacts:
  - ART-W05-E02-S002-002
evidence:
  - EV-W05-E02-S002-002
---

# W05-E02-S002-T002 — Boot-time graph validation

## Task Definition

### Task objective

Implement boot-time graph validation rejecting duplicate providers, missing requirements, undeclared
edges, cycles, and invalid scope/lifetime edges, reusing `kernel/lifecycle`'s existing scope-rank
ordering rather than duplicating it, with error messages naming both owners involved.

### Parent story

W05-E02-S002 — Zero-reflection provider graph, boot-time validation, and profile projection.

### Owner

unassigned

### Status

todo

### Dependencies

W05-E02-S002-T001 (the provider graph this task validates).

### Detailed work

1. Study `kernel/lifecycle`'s existing scope-rank ordering logic to identify the reusable component.
2. Implement validation for each of the five failure classes: duplicate providers, missing
   requirements, undeclared edges, cycles, invalid scope/lifetime edges.
3. Ensure each validation error names both owners involved.
4. Write `AR-02/boot_graph_validation_test.go`: one adversarial fixture per failure class.
5. Document the validation rules.

### Expected files or components affected

The provider-graph package (from T001); `kernel/lifecycle` (read/reused, not modified beyond what
reuse requires).

### Expected output

Boot-time validation rejecting all five failure classes, proven by the named test.

### Required artifacts

ART-W05-E02-S002-002.

### Required evidence

EV-W05-E02-S002-002.

### Related acceptance criteria

AC-W05-E02-S002-02.

### Completion criteria

All five adversarial fixtures are rejected with error messages naming both owners.

### Verification method

Direct execution of `AR-02/boot_graph_validation_test.go`.

### Risks

Medium, per PLAN T4's own risk column — "absorb/replace existing lifecycle scope logic, don't
duplicate it."

### Rollback or recovery considerations

If reuse of `kernel/lifecycle`'s logic proves impractical, escalate before duplicating it — a
duplicated, divergent ordering scheme is a correctness risk this task exists to avoid.

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

*Not yet implemented — the named-both-owners error requirement; recorded here once implemented.*

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
| AC-W05-E02-S002-02 | Run `AR-02/boot_graph_validation_test.go` | Local dev or CI, Go toolchain | All 5 failure classes rejected, errors name both owners | adversarial-test report | unassigned |

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
