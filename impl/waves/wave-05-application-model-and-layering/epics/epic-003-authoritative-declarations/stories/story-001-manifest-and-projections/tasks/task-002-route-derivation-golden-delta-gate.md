---
id: W05-E03-S001-T002
type: task
title: Route derivation and golden-declaration-delta acceptance gate
status: todo
parent_story: W05-E03-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on:
  - W05-E03-S001-T001
acceptance_criteria:
  - AC-W05-E03-S001-02
artifacts:
  - ART-W05-E03-S001-002
evidence:
  - EV-W05-E03-S001-002
---

# W05-E03-S001-T002 — Route derivation and golden-declaration-delta acceptance gate

## Task Definition

### Task objective

Derive route registration/metadata from the manifest, and prove — via a golden-fixture delta test
that IS AR-03's own acceptance gate — that a manifest change deterministically produces the expected
full projection diff (route/permission/resource/schema/OpenAPI/lifecycle/profile/test/doc) with no
other hand-edited file.

### Parent story

W05-E03-S001 — Manifest schema and derived-projection tooling.

### Owner

unassigned

### Status

todo

### Dependencies

W05-E03-S001-T001 (the manifest schema this task derives from); depends on W05-E01 and W05-E02 at
story scope.

### Detailed work

1. Implement route registration/metadata derivation from the manifest schema.
2. Design the golden-fixture test scenario: a manifest change with a known-correct expected full
   projection diff.
3. Write `AR-03/golden_declaration_delta_test.go`: assert the actual diff matches the expected diff
   exactly, and that no other file was hand-edited.
4. Confirm the test is deterministic (re-run multiple times, confirm identical results).
5. Document the derivation mechanism and the golden-delta gate's role.

### Expected files or components affected

Route-derivation tooling (exact location TBD).

### Expected output

Route registration/metadata correctly derived from the manifest, proven by the golden-delta test —
PLAN's own framing: "this test IS the acceptance gate."

### Required artifacts

ART-W05-E03-S001-002.

### Required evidence

EV-W05-E03-S001-002.

### Related acceptance criteria

AC-W05-E03-S001-02.

### Completion criteria

The golden-delta test passes deterministically (confirmed via repeat runs), covering the full named
projection surface.

### Verification method

Direct execution of `AR-03/golden_declaration_delta_test.go`, repeated to confirm determinism.

### Risks

High, per PLAN T3's own risk column — "this test IS the acceptance gate." See RISK-W05-E03-001 in
epic-level `risks.md`.

### Rollback or recovery considerations

If the golden-delta test proves non-deterministic or does not cover the full named projection
surface, treat as a blocking defect in the derivation design itself — do not ship a weakened or
partial version of this test.

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

*Not yet implemented — the golden-delta test's own diagnostic output; recorded here once
implemented.*

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
| AC-W05-E03-S001-02 | Run `AR-03/golden_declaration_delta_test.go` (repeated for determinism) | Local dev or CI, Go toolchain | Golden-fixture manifest change produces the expected full projection diff, deterministically, no other hand-edited file | golden-delta test report | unassigned |

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
