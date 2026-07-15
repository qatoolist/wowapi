---
id: W05-E03-S001-T003
type: task
title: Duplicate-identity/omitted-projection lint rule
status: todo
parent_story: W05-E03-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on:
  - W05-E03-S001-T001
  - W05-E03-S001-T002
acceptance_criteria:
  - AC-W05-E03-S001-03
artifacts:
  - ART-W05-E03-S001-003
evidence:
  - EV-W05-E03-S001-003
---

# W05-E03-S001-T003 — Duplicate-identity/omitted-projection lint rule

## Task Definition

### Task objective

Implement a lint rule that fails on hand-maintained duplicate identity or an omitted projection,
proven by adversarial fixtures for both failure modes.

### Parent story

W05-E03-S001 — Manifest schema and derived-projection tooling.

### Owner

unassigned

### Status

todo

### Dependencies

W05-E03-S001-T001, W05-E03-S001-T002 (the manifest schema and derivation tooling this lint rule
checks against).

### Detailed work

1. Design the lint rule: detect hand-maintained duplicate identity (a declaration hand-duplicated
   outside the manifest) and omitted projection (a manifest entry with no corresponding derived
   projection).
2. Implement the lint rule.
3. Write `AR-03/duplicate_omission_lint_test.go`: one adversarial fixture for duplicate identity, one
   for omitted projection.
4. Document the lint rule.

### Expected files or components affected

A new lint-rule package (exact tooling TBD).

### Expected output

A lint rule that fails both adversarial fixtures.

### Required artifacts

ART-W05-E03-S001-003.

### Required evidence

EV-W05-E03-S001-003.

### Related acceptance criteria

AC-W05-E03-S001-03.

### Completion criteria

Both adversarial fixtures (duplicate identity, omitted projection) fail lint.

### Verification method

Direct execution of `AR-03/duplicate_omission_lint_test.go`.

### Risks

Medium, per PLAN T4's own risk column.

### Rollback or recovery considerations

If either fixture is found to pass lint incorrectly, fix before proceeding to T004.

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
| AC-W05-E03-S001-03 | Run `AR-03/duplicate_omission_lint_test.go` | Local dev or CI, Go toolchain | Both adversarial fixtures fail lint | adversarial-lint report | unassigned |

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
