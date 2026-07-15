---
id: W05-E03-S001-T001
type: task
title: Manifest schema definition
status: todo
parent_story: W05-E03-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on: []
acceptance_criteria:
  - AC-W05-E03-S001-01
artifacts:
  - ART-W05-E03-S001-001
evidence:
  - EV-W05-E03-S001-001
---

# W05-E03-S001-T001 — Manifest schema definition

## Task Definition

### Task objective

Define the manifest schema, scoped to identity + projection inputs (not DX-03's full
typed-operation DSL), with fields traceable 1:1 to existing scattered declarations, and prove it
round-trips against at least one existing internal fixture module.

### Parent story

W05-E03-S001 — Manifest schema and derived-projection tooling.

### Owner

unassigned

### Status

todo

### Dependencies

None within this story (S001's own first task); depends on W05-E01 (AR-01 T1) at story scope.

### Detailed work

1. Audit the current scattered-declaration surface across the framework.
2. Design the manifest schema, explicitly bounding scope to identity + projection inputs.
3. Write `AR-03/manifest_schema_fixture_test.go`: round-trip against ≥1 existing internal fixture
   module.
4. Document the schema and its 1:1 traceability to existing declarations.

### Expected files or components affected

A new manifest-schema package (exact location TBD).

### Expected output

A manifest schema proven to round-trip against an existing fixture module.

### Required artifacts

ART-W05-E03-S001-001.

### Required evidence

EV-W05-E03-S001-001.

### Related acceptance criteria

AC-W05-E03-S001-01.

### Completion criteria

The round-trip test passes; the schema introduces no new parallel metadata system ahead of the
model.

### Verification method

Direct execution of `AR-03/manifest_schema_fixture_test.go`.

### Risks

Medium, per PLAN T1's own risk column — "scope-creep risk into DX-03 territory."

### Rollback or recovery considerations

If the schema design is found to creep into DX-03's excluded full-DSL territory, revise before
proceeding to T002, which depends on this schema's shape being stable.

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

*Not applicable — no database schema/migration; this task defines a manifest schema, not a database
change.*

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
| AC-W05-E03-S001-01 | Run `AR-03/manifest_schema_fixture_test.go` | Local dev or CI, Go toolchain | Schema round-trips against ≥1 existing fixture module | unit-test report | unassigned |

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
