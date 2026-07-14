---
id: W05-E05-S001-T004
type: task
title: Boundaries-lint allowlist extension
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
  - ART-W05-E05-S001-004
evidence:
  - EV-W05-E05-S001-004
---

# W05-E05-S001-T004 — Boundaries-lint allowlist extension

## Task Definition

### Task objective

Extend `scripts/lint_boundaries.sh`'s allowlist so a new kernel package addition fails CI without an
explicit allowlist edit — a review-forcing mechanism per MATRIX CS-01's own step 5.

### Parent story

W05-E05-S001 — Foundation tree, package moves, and mfa forwarding shim.

### Owner

unassigned

### Status

todo

### Dependencies

W05-E05-S001-T001, W05-E05-S001-T002 (the final post-move kernel package set this allowlist is built
against).

### Detailed work

1. Determine the final, post-move set of retained `kernel/` packages.
2. Extend `scripts/lint_boundaries.sh`'s allowlist to this exact set.
3. Write an adversarial fixture: adding a new, un-allowlisted kernel package, confirming CI fails
   without an explicit allowlist edit.
4. Document the allowlist and its review-forcing rationale.

### Expected files or components affected

`scripts/lint_boundaries.sh`.

### Expected output

CI fails on a new un-allowlisted kernel package addition.

### Required artifacts

ART-W05-E05-S001-004.

### Required evidence

EV-W05-E05-S001-004.

### Related acceptance criteria

AC-W05-E05-S001-03.

### Completion criteria

The adversarial fixture confirms CI fails on the un-allowlisted addition.

### Verification method

Direct execution of `scripts/lint_boundaries.sh` against the adversarial fixture.

### Risks

Low-medium — reuse of existing tooling, per MATRIX CS-01's own "Reuse tier" framing.

### Rollback or recovery considerations

If the allowlist is found to be too permissive (does not actually block a new addition) or too
restrictive (blocks a legitimate existing package), fix before proceeding.

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

*Not yet implemented — the extended allowlist; recorded here once implemented.*

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
| AC-W05-E05-S001-03 | Run `scripts/lint_boundaries.sh` against the adversarial fixture | Local dev or CI | CI fails on a new un-allowlisted kernel package | adversarial-lint report | unassigned |

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
