---
id: W06-E02-S003-T001
type: task
title: OpenAPI semantic diff (blocked on W06-E02-S001)
status: blocked
parent_story: W06-E02-S003
owner: W06E02Impl
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on:
  - W06-E02-S001
acceptance_criteria:
  - AC-W06-E02-S003-01
artifacts:
  - ART-W06-E02-S003-001
evidence:
  - EV-W06-E02-S003-001
---

# W06-E02-S003-T001 — OpenAPI semantic diff (blocked on W06-E02-S001)

## Task Definition

### Task objective

Implement an OpenAPI semantic diff classifying breaking changes per DX-06's 3.1/2020-12 baseline. BLOCKED: cannot begin until W06-E02-S001 (DX-06) reaches accepted.

### Parent story

W06-E02-S003

### Owner

unassigned

### Status

todo

### Dependencies

W06-E02-S001 (DX-06) must be `accepted` — a lossy merge cannot be meaningfully diffed, per MATRIX CS-15's own framing. This task must not begin before that entry criterion is satisfied.

### Detailed work

1. Confirm W06-E02-S001 has reached `accepted`.
2. Build the semantic-diff mechanism for OpenAPI documents, classifying breaking changes per DX-06's
   3.1/2020-12 baseline.
3. Write a seeded breaking-OpenAPI fixture and confirm the gate fails it.

### Expected files or components affected

An extension within or alongside W06-E02-S001's semantic-diff CI job.

### Expected output

A gate failing a seeded breaking-OpenAPI fixture, classified per DX-06's baseline.

### Required artifacts

ART-W06-E02-S003-001 (OpenAPI semantic-diff gate).

### Required evidence

EV-W06-E02-S003-001 (seeded breaking-OpenAPI-fixture test report).

### Related acceptance criteria

AC-W06-E02-S003-01.

### Completion criteria

The seeded breaking-OpenAPI fixture fails the gate.

### Verification method

Direct execution of the gate against the fixture, once unblocked.

### Risks

The primary risk is beginning this task before its entry criterion is genuinely satisfied — mitigated by this task's own explicit blocked-status framing.

### Rollback or recovery considerations

If begun prematurely, halt and record a deviation; do not silently proceed against an unaccepted upstream dependency.

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
| AC-W06-E02-S003-01 | Run the gate against the seeded fixture, once W06-E02-S001 is accepted | CI | Breaking fixture fails the gate | CI gate test report | unassigned |

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
