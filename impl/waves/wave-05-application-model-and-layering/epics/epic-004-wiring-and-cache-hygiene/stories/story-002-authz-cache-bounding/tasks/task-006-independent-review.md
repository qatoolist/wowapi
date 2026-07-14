---
id: W05-E04-S002-T006
type: task
title: Independent review
status: todo
parent_story: W05-E04-S002
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on:
  - W05-E04-S002-T001
  - W05-E04-S002-T002
  - W05-E04-S002-T003
  - W05-E04-S002-T004
  - W05-E04-S002-T005
acceptance_criteria:
  - AC-W05-E04-S002-01
  - AC-W05-E04-S002-02
  - AC-W05-E04-S002-03
  - AC-W05-E04-S002-04
artifacts: []
evidence: []
---

# W05-E04-S002-T006 — Independent review

## Task Definition

### Task objective

Perform an independent review of this story's implementation per mandate §14, confirming: T004's
epoch-bump wiring is genuinely complete across every enumerated framework-side mutation path (not
merely claimed), given PLAN's own "Highest-risk task" framing for SEC-04 T4; D-06 is enacted as
ratified, not reinterpreted; the DATA-07 T4 cache-invalidation AC-closure relationship is correctly
recorded; no source requirement (SEC-04 T1-T6) was silently dropped or narrowed.

### Parent story

W05-E04-S002 — Bounded, epoch-invalidated authorization cache.

### Owner

unassigned

### Status

todo

### Dependencies

W05-E04-S002-T001 through T005 (review requires all preceding tasks implemented first).

### Detailed work

1. Independently re-run (or re-inspect the CI output of) T004's simulated cross-pod test.
2. Independently confirm the enumerated mutation-path list (role/permission assignment writes in
   `kernel/authz`, seeds, SEC-01's grant-table writes) is genuinely complete — cross-check against
   the actual framework codebase for any additional mutation path not enumerated, given RISK-W05-005's
   own "wiring completeness" concern.
3. Confirm each epoch bump is genuinely atomic (same transaction) with its triggering mutation, not
   merely eventually-consistent.
4. Confirm T001-T003's own tests (bounded cache, eviction metrics, singleflight) match PLAN's own
   acceptance criteria.
5. Confirm T005's decision-provenance and prod-config-gate tests are genuine, not weakened.
6. Confirm this story's `story.md` correctly records the DATA-07 T4 cache-invalidation AC-closure
   relationship by ID.
7. Confirm evidence in `evidence/index.md` identifies the tested commit SHA for every item.
8. Record findings; if any issue is found, it must be resolved or explicitly accepted before this
   story moves to `accepted`.

### Expected files or components affected

None (review-only task; no code change).

### Expected output

An independent-review record confirming or rejecting this story's readiness for `accepted` status.

### Required artifacts

None.

### Required evidence

None beyond this task's own review record.

### Related acceptance criteria

AC-W05-E04-S002-01, AC-W05-E04-S002-02, AC-W05-E04-S002-03, AC-W05-E04-S002-04.

### Completion criteria

The review record confirms all four acceptance criteria are proven with valid evidence, or lists
findings that must be resolved before this story can close.

### Verification method

Manual review against mandate §14's checklist, with independent re-inspection of the mutation-path
enumeration and re-execution of the cross-pod test as the specific "genuinely, not merely claimed"
checks for this story's Highest-risk-task content.

### Risks

RISK-W05-005 — mitigated by requiring the reviewer to independently cross-check the mutation-path
enumeration against the actual codebase, not merely trust T004's own self-reported list.

### Rollback or recovery considerations

Not applicable — a review task has no code to roll back; a failed review blocks story acceptance
until its findings are resolved.

## Implementation Record

*Not applicable — this is a review task, not an implementation task.*

### What was actually implemented

*Not applicable.*

### Components changed

*Not applicable.*

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

*Not applicable.*

### Commits

*Not applicable.*

### Pull requests

*Not applicable.*

### Implementation dates

*Not applicable.*

### Technical debt introduced

*Not applicable.*

### Known limitations

*Not applicable.*

### Follow-up items

*Not yet executed.*

### Relationship to the approved plan

*Not applicable.*

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W05-E04-S002-01 | Independent review against mandate §14 checklist | Documentation + code review | Confirmed: bounded cache and eviction metrics genuinely proven | review report | unassigned |
| AC-W05-E04-S002-02 | Independent re-inspection of mutation-path enumeration + re-execution of cross-pod test | Code review + codebase cross-check + test execution | Confirmed: mutation-path enumeration complete, epoch bumps atomic | review report | unassigned |
| AC-W05-E04-S002-03 | Independent review against mandate §14 checklist | Code review + test-output inspection | Confirmed: decision provenance and prod-config gate genuine | review report | unassigned |
| AC-W05-E04-S002-04 | Independent review of `story.md`'s Dependencies section | Documentation review | Confirmed: DATA-07 T4 cross-reference correctly recorded | review report | unassigned |

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
