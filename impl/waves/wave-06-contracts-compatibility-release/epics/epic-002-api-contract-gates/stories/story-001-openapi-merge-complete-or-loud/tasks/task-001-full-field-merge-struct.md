---
id: W06-E02-S001-T001
type: task
title: Full-field merge struct and per-field policy
status: done
parent_story: W06-E02-S001
owner: W06E02Impl
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on: []
acceptance_criteria:
  - AC-W06-E02-S001-01
artifacts:
  - ART-W06-E02-S001-001
  - ART-W06-E02-S001-002
evidence:
  - EV-W06-E02-S001-001
---

# W06-E02-S001-T001 — Full-field merge struct and per-field policy

## Task Definition

### Task objective

Expand the OpenAPI merge struct to cover every 3.1 top-level field and every components.* field, with an explicit per-field merge policy, proven by a fixture-driven test suite.

### Parent story

W06-E02-S001

### Owner

unassigned

### Status

todo

### Dependencies

None.

### Detailed work

1. Enumerate every OpenAPI 3.1 top-level field and every components.* sub-field against the 3.1.1
   specification.
2. Design an explicit per-field merge policy (union, identical-required, or reject-on-conflict) for
   each field, with documented rationale.
3. Implement the expanded merge struct and per-field policy logic in internal/cli/openapi_cmd.go.
4. Write the fixture-driven test suite: one fragment per field type.

### Expected files or components affected

internal/cli/openapi_cmd.go (expanded merge struct); new fixture test files.

### Expected output

An expanded merge struct where every field either merges correctly or is explicitly rejected, proven by the fixture suite.

### Required artifacts

ART-W06-E02-S001-001 (expanded merge struct with per-field policy), ART-W06-E02-S001-002 (fixture-driven per-field test suite).

### Required evidence

EV-W06-E02-S001-001 (per-field-type fixture test output).

### Related acceptance criteria

AC-W06-E02-S001-01.

### Completion criteria

The fixture suite passes for every OpenAPI 3.1 top-level/components.* field type.

### Verification method

Direct execution of the fixture-driven test suite.

### Risks

None beyond the general design-under-specification risk (RISK-W06-E02-002 at epic scope, elaborated for the validator decision, not this task specifically).

### Rollback or recovery considerations

If a field's policy proves wrong after landing, revise that field's policy and re-run the fixture suite; do not silently widen the policy without recording why.

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
| AC-W06-E02-S001-01 | Run the fixture-driven per-field test suite | Local dev or CI, Go toolchain | Every field merges correctly per policy or is explicitly rejected | fixture test report | unassigned |

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
