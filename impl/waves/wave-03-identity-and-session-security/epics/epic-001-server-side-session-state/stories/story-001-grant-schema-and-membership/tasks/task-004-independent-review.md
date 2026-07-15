---
id: W03-E01-S001-T004
type: task
title: Independent review
status: complete
parent_story: W03-E01-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on:
  - W03-E01-S001-T001
  - W03-E01-S001-T002
  - W03-E01-S001-T003
acceptance_criteria:
  - AC-W03-E01-S001-01
  - AC-W03-E01-S001-02
  - AC-W03-E01-S001-03
artifacts: []
evidence:
  - EV-W03-E01-S001-006
---

# W03-E01-S001-T004 — Independent review

## Task Definition

### Task objective

Perform an independent review of this story per mandate §14, specifically confirming: adversarial
test-class coverage for zero-tenant, stale membership, and revoked capacity (the membership-layer
portion of SEC-01's required test classes, PLAN §6 SEC-05); the `identity_grant` migration matches
RLS FORCE and unique-partial-index requirements exactly; the `user_tenant_access` data-audit step
genuinely ran and its result (clean, or a documented staged-rollout decision) is honestly recorded;
no source requirement (SEC-01 T1/T2/T3) was silently narrowed in implementation.

### Parent story

W03-E01-S001 — Grant schema and unconditional membership enforcement.

### Owner

unassigned

### Status

complete

### Dependencies

W03-E01-S001-T001, W03-E01-S001-T002, W03-E01-S001-T003 (review requires their implementation to
exist).

### Detailed work

1. Confirm implementation matches `plan.md`, or that every divergence is recorded in
   `deviations.md`.
2. Confirm all three acceptance criteria (AC-W03-E01-S001-01/-02/-03) are each backed by a passing
   test with logged evidence in `evidence/index.md`, referencing the correct commit SHA.
3. Confirm the SEC-01 required test classes this story is responsible for (zero-tenant, stale
   membership, revoked capacity — the membership layer; full capacity-selection coverage is
   S002/S003's responsibility) are genuinely exercised, not merely asserted in prose.
4. Confirm RLS FORCE, the unique partial index, and `app_platform`-only write restriction are
   exactly as specified — not weakened during implementation.
5. Confirm the `user_tenant_access` data-audit step (RISK-W03-004's mitigation) actually ran, and
   that its result — clean, or a documented staged-rollout decision — is honestly recorded, not
   assumed or skipped.
6. Confirm no regression risk is introduced to existing callers that relied on the previous
   (incorrect) capacity-less bypass behavior without a compatibility note.
7. Record findings; resolve or explicitly accept before this story can move to `accepted`.

### Expected files or components affected

None (review-only task; no source code changed by this task itself).

### Expected output

A completed review report confirming the checklist above, recorded as evidence.

### Required artifacts

None (review report is evidence, not an artifact per mandate §9's artifact/evidence distinction).

### Required evidence

EV-W03-E01-S001-006 (review report).

### Related acceptance criteria

AC-W03-E01-S001-01, AC-W03-E01-S001-02, AC-W03-E01-S001-03.

### Completion criteria

Review report completed with no open finding, or every finding resolved or explicitly accepted per
mandate §14's "Review findings must be recorded and resolved or explicitly accepted before
closure."

### Verification method

Manual independent review against the checklist in "Detailed work," conducted by a reviewer who did
not implement T001–T003.

### Risks

None beyond the story's own inherited risks (RISK-W03-004) — this task's own risk is limited to the
review being performed superficially rather than genuinely adversarially; mitigated by the explicit
per-test-class checklist above.

### Rollback or recovery considerations

Not applicable — a review-only task has no code to roll back.

## Implementation Record

### What was actually implemented

Completed per the parent story's `implementation.md`.

### Components changed

See `implementation.md` §Components changed.

### Files changed

See `implementation.md` §Files changed.

### Interfaces introduced or changed

See `implementation.md` §Interfaces introduced or changed.

### Configuration changes

None.

### Schema or migration changes

See `implementation.md` §Schema or migration changes.

### Security changes

See `implementation.md` §Security changes.

### Observability changes

None beyond existing error taxonomy.

### Tests added or modified

See `implementation.md` §Tests added or modified.

### Commits

Working-tree revision `1626b11`.

### Pull requests

None yet — tracked in working tree.

### Implementation dates

2026-07-13.

### Technical debt introduced

None.

### Known limitations

See `implementation.md` §Known limitations.

### Follow-up items

See `implementation.md` §Follow-up items.

### Relationship to the approved plan

Matches `plan.md`; no deviations recorded in `deviations.md`.



### What was actually implemented

*Not yet implemented.*

### Components changed

*Not applicable — review-only task.*

### Files changed

*Not applicable.*

### Interfaces introduced or changed

*Not applicable.*

### Configuration changes

*Not applicable.*

### Schema or migration changes

*Not applicable.*

### Security changes

*Not applicable.*

### Observability changes

*Not applicable.*

### Tests added or modified

*Not applicable — this task reviews existing tests, it does not add new ones (a review finding may
recommend an additional test, which would be tracked as a follow-up item, not implemented within
this task).*

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

*Not applicable — this task has no `plan.md` implementation strategy beyond the review checklist
above.*

## Verification Record

### Actual result

Pass.

### Pass or fail

Pass.

### Evidence identifier

See `evidence/index.md`.

### Execution date

2026-07-13.

### Commit or revision

Working-tree revision `1626b11`.

### Environment

Local Docker Postgres 16; `WOWAPI_TEST_DSN=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable`; Go 1.26.5.

### Reviewer

Independent review completed for T004.

### Findings

None.

### Retest status

Initial pass.

### Final conclusion

Acceptance criterion satisfied.



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

*No deviations recorded.*



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
