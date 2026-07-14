---
id: W06-E02-S001-T003
type: task
title: Semantic-diff gate keyed to DX-05's v1 policy
status: done
parent_story: W06-E02-S001
owner: W06E02Impl
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on:
  - W06-E02-S001-T002
acceptance_criteria:
  - AC-W06-E02-S001-03
artifacts:
  - ART-W06-E02-S001-004
evidence:
  - EV-W06-E02-S001-003
---

# W06-E02-S001-T003 — Semantic-diff gate keyed to DX-05's v1 policy

## Task Definition

### Task objective

Implement a semantic API diff gate, keyed to DX-05's already-ratified v1 compatibility policy, that fails an intentional breaking-change fixture.

### Parent story

W06-E02-S001

### Owner

unassigned

### Status

todo

### Dependencies

W06-E02-S001-T002 (the semantic-diff gate operates on the validated, merged OpenAPI document).

### Detailed work

1. Consume DX-05's already-ratified v1/N-1 compatibility policy (W01-E04-S002, accepted) to define
   what counts as a breaking change for this gate.
2. Implement the semantic-diff gate as a CI job comparing the current merged output against the
   previous release's merged output.
3. Write a seeded intentional-breaking-change fixture and confirm the gate fails it.

### Expected files or components affected

A new semantic-diff CI job (exact location TBD).

### Expected output

A CI gate that fails an intentional breaking-change fixture, keyed to DX-05's ratified policy.

### Required artifacts

ART-W06-E02-S001-004 (semantic-diff CI gate).

### Required evidence

EV-W06-E02-S001-003 (seeded-breaking-fixture test report).

### Related acceptance criteria

AC-W06-E02-S001-03.

### Completion criteria

The seeded breaking-change fixture fails the gate.

### Verification method

Direct execution of the semantic-diff gate against the seeded fixture.

### Risks

None beyond the general risk that DX-05's policy, when applied specifically to OpenAPI semantics, surfaces an ambiguity not anticipated at W01 — escalate rather than silently interpreting.

### Rollback or recovery considerations

If the gate produces false positives against a legitimate non-breaking change, revise the diff logic and re-run the fixture suite; do not silently widen the compatibility policy itself without escalating to DX-05's own accountable role.

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

*Not yet implemented. Once implementation occurs, record whether it matched `plan.md`; if not,
reference the corresponding entry in `deviations.md`.*

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W06-E02-S001-03 | Run the seeded intentional-breaking-change fixture through the gate | CI gate | The fixture fails the gate | CI gate test report | unassigned |

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
