---
id: W05-E03-S001-T004
type: task
title: Documentation/test/manifest export projections
status: todo
parent_story: W05-E03-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on:
  - W05-E03-S001-T001
  - W05-E03-S001-T002
  - W05-E03-S001-T003
acceptance_criteria:
  - AC-W05-E03-S001-03
artifacts:
  - ART-W05-E03-S001-004
evidence:
  - EV-W05-E03-S001-004
---

# W05-E03-S001-T004 — Documentation/test/manifest export projections

## Task Definition

### Task objective

Extend T002's golden-delta coverage to documentation/test/manifest export output, sharing fixtures
with AR-05 where practical.

### Parent story

W05-E03-S001 — Manifest schema and derived-projection tooling.

### Owner

unassigned

### Status

todo

### Dependencies

W05-E03-S001-T001, W05-E03-S001-T002, W05-E03-S001-T003 (the full preceding surface).

### Detailed work

1. Implement documentation-table and manifest-export projection derivation from the manifest.
2. Extend the golden-delta fixture(s) to cover doc-table/manifest-export output.
3. Write `AR-03/full_projection_golden_test.go`.
4. Note fixture-sharing opportunities with AR-05 (W06-E04 scope) for future coordination, without
   requiring AR-05's own stories to exist.
5. Document the extended projection coverage.

### Expected files or components affected

Documentation/manifest-export projection tooling (exact location TBD).

### Expected output

Golden-delta coverage extended to doc-table/manifest-export output, proven by the named test.

### Required artifacts

ART-W05-E03-S001-004.

### Required evidence

EV-W05-E03-S001-004.

### Related acceptance criteria

AC-W05-E03-S001-03.

### Completion criteria

`AR-03/full_projection_golden_test.go` passes, covering doc-table and manifest-export output.

### Verification method

Direct execution of `AR-03/full_projection_golden_test.go`.

### Risks

Low-medium, per PLAN T5's own risk column — "share fixtures with AR-05."

### Rollback or recovery considerations

If fixture-sharing with AR-05 proves impractical at this story's own scope, proceed with
independent fixtures and record the missed sharing opportunity as a note for AR-05's future
implementers, not a blocking issue for this task.

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
| AC-W05-E03-S001-03 | Run `AR-03/full_projection_golden_test.go` | Local dev or CI, Go toolchain | Golden-delta coverage extends to doc-table/manifest-export output | golden-delta test report | unassigned |

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
