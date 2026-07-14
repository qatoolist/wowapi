---
id: W03-E01-S003-T003
type: task
title: Independent review
status: todo
parent_story: W03-E01-S003
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on:
  - W03-E01-S003-T001
  - W03-E01-S003-T002
acceptance_criteria:
  - AC-W03-E01-S003-01
  - AC-W03-E01-S003-02
artifacts: []
evidence:
  - EV-W03-E01-S003-003
---

# W03-E01-S003-T003 — Independent review

## Task Definition

### Task objective

Perform an independent review of this story per mandate §14, specifically confirming: the
"expired step-up" required test class (PLAN §6 SEC-05) is genuinely exercised; the credential-scheme
distinction mechanism correctly rejects a mismatched scheme without over- or under-restricting
legitimate combinations; and — specifically for this story — that the DX-03 cross-cut coordination
note (`plan.md`'s "Unresolved questions") was recorded explicitly rather than silently resolved or
silently ignored.

### Parent story

W03-E01-S003 — Assurance freshness and credential-scheme distinction.

### Owner

unassigned

### Status

todo

### Dependencies

W03-E01-S003-T001, W03-E01-S003-T002.

### Detailed work

1. Confirm implementation matches `plan.md`, or that every divergence is recorded in
   `deviations.md`.
2. Confirm both acceptance criteria are each backed by a passing test with logged evidence in
   `evidence/index.md`, referencing the correct commit SHA.
3. Confirm the "expired step-up" required test class is genuinely exercised (stale `auth_time` +
   valid `amr` → step-up fails), not merely asserted in prose.
4. Confirm the credential-scheme distinction test genuinely proves a `CredentialUser`-scoped
   permission rejects a valid API-key actor, and spot-check that a correctly-scoped user-credential
   request is not incorrectly rejected by the same mechanism (a false-positive-rejection check).
5. **Confirm the DX-03 cross-cut coordination note was recorded, not silently resolved.** Verify
   that this story's `CredentialScheme` mechanism is documented as a candidate for reconciliation
   with DX-03 (W06-E01-S001), per `plan.md`'s "Unresolved questions," and that no part of this
   story's implementation or documentation claims to have made DX-03's eventual design decision.
6. Record findings; resolve or explicitly accept before this story can move to `accepted`.

### Expected files or components affected

None (review-only task).

### Expected output

A completed review report confirming the checklist above, recorded as evidence.

### Required artifacts

None (review report is evidence, not an artifact per mandate §9's artifact/evidence distinction).

### Required evidence

EV-W03-E01-S003-003 (review report).

### Related acceptance criteria

AC-W03-E01-S003-01, AC-W03-E01-S003-02.

### Completion criteria

Review report completed with no open finding, or every finding resolved or explicitly accepted per
mandate §14.

### Verification method

Manual independent review against the checklist in "Detailed work," conducted by a reviewer who did
not implement T001/T002.

### Risks

None beyond the story's own inherited scope — this task's own risk is limited to the DX-03
cross-cut note being overlooked during review; mitigated by making it an explicit checklist item
(step 5 above).

### Rollback or recovery considerations

Not applicable — a review-only task has no code to roll back.

## Implementation Record

### What was actually implemented

Independent review performed against the checklist in this task's "Detailed work":

1. Implementation matches `plan.md`; no deviations required.
2. Both acceptance criteria are backed by tests in `kernel/authz/assurance_freshness_test.go` and
   `kernel/authz/credential_scheme_test.go`.
3. The "expired step-up" required test class is exercised: stale `auth_time` + valid `amr` →
   `step_up_freshness_required`.
4. The credential-scheme test proves a `CredentialUser`-scoped permission rejects a valid API-key
   actor, and the positive-path tests confirm correctly-scoped actors are not falsely rejected.
5. The DX-03 cross-cut coordination note is recorded in `plan.md`, `story.md`, T002's
   "Technical debt introduced", and this task's review findings.

### Components changed

None (review-only).

### Files changed

None (review-only).

### Interfaces introduced or changed

None.

### Configuration changes

None.

### Schema or migration changes

None.

### Security changes

None directly.

### Observability changes

None.

### Tests added or modified

None (reviewed existing tests).

### Commits

None.

### Pull requests

None.

### Implementation dates

2026-07-13.

### Technical debt introduced

None.

### Known limitations

DB-backed tests are skipped without `DATABASE_URL`. Review relied on package builds and passing
non-DB tests.

### Follow-up items

- Re-run DB-backed tests with `DATABASE_URL` set.

### Relationship to the approved plan

Review confirms implementation matches `plan.md` with no deviations.

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W03-E01-S003-01, -02 | Independent review checklist per mandate §14 | N/A (documentation/code review) | No open finding, or every finding resolved/accepted | review report | unassigned |

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
