---
id: W06-E02-S001-T002
type: task
title: Validator-dependency decision and structural validation
status: done
parent_story: W06-E02-S001
owner: W06E02Impl
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on:
  - W06-E02-S001-T001
acceptance_criteria:
  - AC-W06-E02-S001-02
  - AC-W06-E02-S001-04
artifacts:
  - ART-W06-E02-S001-003
evidence:
  - EV-W06-E02-S001-002
  - EV-W06-E02-S001-004
---

# W06-E02-S001-T002 — Validator-dependency decision and structural validation

## Task Definition

### Task objective

Evaluate and select an OpenAPI 3.1 validator dependency (candidate: pb33f/libopenapi) with a security/licence review, then wire it in to validate the merged document against 3.1.1/2020-12.

### Parent story

W06-E02-S001

### Owner

unassigned

### Status

todo

### Dependencies

W06-E02-S001-T001 (structural validation operates on the expanded merge struct's output).

### Detailed work

1. Evaluate the pb33f/libopenapi candidate (per MATRIX CS-15) and any alternatives, including a
   security and licence review.
2. Select the validator and record the decision with its review outcome.
3. Wire the selected validator into the merge command to validate the final merged document against
   OpenAPI 3.1.1 / JSON Schema 2020-12.
4. Write a malformed-merged-output negative fixture test confirming the command fails on invalid
   output.

### Expected files or components affected

internal/cli/openapi_cmd.go (validation wiring); go.mod (new dependency); a decision record.

### Expected output

A validator dependency selected with a documented security/licence review, wired in to validate the merged document.

### Required artifacts

ART-W06-E02-S001-003 (structural validator wiring and decision record).

### Required evidence

EV-W06-E02-S001-002 (structural-validation test report), EV-W06-E02-S001-004 (validator security/licence review record).

### Related acceptance criteria

AC-W06-E02-S001-02, AC-W06-E02-S001-04.

### Completion criteria

The selected validator's security/licence review is recorded and predates its use; valid output passes, malformed output fails.

### Verification method

Direct execution of the structural-validation test plus inspection of the review record.

### Risks

RISK-W06-E02-001 (validator decision made without adequate review if rushed) — see epic-level `risks.md`.

### Rollback or recovery considerations

If the selected validator fails review after being wired in, revert to evaluation and select an alternative; do not silently keep a rejected dependency.

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
| AC-W06-E02-S001-02 | Run structural validation against valid and malformed merged-output fixtures | Local dev or CI, Go toolchain | Valid output passes; malformed output fails | structural-validation test report | unassigned |
| AC-W06-E02-S001-04 | Inspect the validator decision record for a security/licence review outcome | Documentation review | A dated review record exists and predates the dependency's use | review report | unassigned |

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
