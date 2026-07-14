---
id: W05-E05-S001-T003
type: task
title: Depguard extension
status: todo
parent_story: W05-E05-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on:
  - W05-E05-S001-T001
  - W05-E05-S001-T002
acceptance_criteria:
  - AC-W05-E05-S001-03
artifacts:
  - ART-W05-E05-S001-003
evidence:
  - EV-W05-E05-S001-003
---

# W05-E05-S001-T003 — Depguard extension

## Task Definition

### Task objective

Extend `depguard`'s `.golangci.yml` kernel rule to deny `kernel → foundation` imports, and add a
`foundation` rule denying `foundation → app` imports.

### Parent story

W05-E05-S001 — Foundation tree, package moves, and mfa forwarding shim.

### Owner

unassigned

### Status

todo

### Dependencies

W05-E05-S001-T001, W05-E05-S001-T002 (the `foundation/` tree these rules police must exist first).

### Detailed work

1. Extend the existing `depguard` kernel rule in `.golangci.yml` to deny `kernel → foundation`
   imports.
2. Add a new `foundation` depguard rule denying `foundation → app` imports.
3. Write an adversarial fixture: an attempted `kernel → foundation` import and an attempted
   `foundation → app` import, both expected to be denied.
4. Confirm the fixture that fails today against the (pre-move) nine packages now passes after the
   re-home — per MATRIX CS-01's own fail-first framing.
5. Document the extended rules.

### Expected files or components affected

`.golangci.yml`.

### Expected output

Both new denial rules active and proven by the adversarial fixture.

### Required artifacts

ART-W05-E05-S001-003.

### Required evidence

EV-W05-E05-S001-003.

### Related acceptance criteria

AC-W05-E05-S001-03.

### Completion criteria

The adversarial fixture confirms both denial rules trigger correctly.

### Verification method

Direct execution of the lint against the adversarial fixture.

### Risks

Low-medium — this is fuller configuration of existing tooling, per MATRIX CS-01's own "Reuse tier"
framing.

### Rollback or recovery considerations

If either denial rule is found to be bypassable or to produce false positives against legitimate
imports, fix before proceeding.

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

*Not yet implemented — the extended depguard rules; recorded here once implemented.*

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
| AC-W05-E05-S001-03 | Run the depguard adversarial fixture | Local dev or CI, Go toolchain (lint) | Both kernel→foundation and foundation→app imports denied | adversarial-lint report | unassigned |

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
