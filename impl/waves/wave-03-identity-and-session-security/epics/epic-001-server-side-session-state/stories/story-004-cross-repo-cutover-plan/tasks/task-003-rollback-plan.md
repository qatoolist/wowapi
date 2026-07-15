---
id: W03-E01-S004-T003
type: task
title: Rollback plan
status: todo
parent_story: W03-E01-S004
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on: []
acceptance_criteria:
  - AC-W03-E01-S004-03
artifacts:
  - ART-W03-E01-S004-003
evidence:
  - EV-W03-E01-S004-003
---

# W03-E01-S004-T003 — Rollback plan

## Task Definition

### Task objective

Produce a rollback plan document covering both failure directions of the cutover: (a) wowapi-side
grant-table enforcement causing a wowsociety regression, and (b) wowsociety-side `grant_id`
adoption itself being found broken. **This is a coordination-artifact task. It produces a planning
document, not product code.**

### Parent story

W03-E01-S004 — Cross-repo cutover plan for the wowsociety impersonation-flow breaking change.

### Owner

unassigned

### Status

todo

### Dependencies

None (can be drafted in parallel with T001/T002), though its content is most coherent once T001's
sequencing plan establishes what stage-by-stage rollback points exist.

### Detailed work

1. Enumerate the specific revert steps for failure direction (a): if wowapi's T2 unconditional
   enforcement or T5 resolver causes a wowsociety-side regression, what is reverted first
   (wowsociety trust behavior, or wowapi enforcement itself), and how is consistency confirmed
   afterward.
2. Enumerate the specific revert steps for failure direction (b): if wowsociety's own `grant_id`
   adoption is found broken (e.g. a migration issue in `identity_impersonation_session`), what is
   the wowsociety-side rollback, and does it require any coordinated wowapi-side action.
3. State how each rollback path confirms neither repo is left in an inconsistent state (e.g. a
   partially-cutover state where wowapi expects `grant_id` but wowsociety has reverted its adoption
   of it).
4. Circulate the draft for review by at least a wowapi-side reviewer.
5. Record the review outcome.

### Expected files or components affected

None in source code — a new planning document is produced (path TBD at implementation time).

### Expected output

A reviewed rollback plan document satisfying AC-W03-E01-S004-03.

### Required artifacts

ART-W03-E01-S004-003 (rollback plan document).

### Required evidence

EV-W03-E01-S004-003 (review report).

### Related acceptance criteria

AC-W03-E01-S004-03.

### Completion criteria

The document exists, covers both named failure directions with specific revert steps, and has
passed review with no open finding.

### Verification method

Document review against the checklist in "Detailed work," logged as a review report.

### Risks

Low direct risk (documentation task); the underlying coordination risk it addresses is RISK-W03-002
at epic/wave scope.

### Rollback or recovery considerations

This task's entire subject matter is rollback/recovery planning for another change — this task
itself has no code to roll back.

## Implementation Record

### What was actually implemented

Produced the rollback plan document: `rollback-plan.md`.

### Components changed

None (documentation only).

### Files changed

- `impl/waves/wave-03-identity-and-session-security/epics/epic-001-server-side-session-state/stories/story-004-cross-repo-cutover-plan/rollback-plan.md` (new)

### Interfaces introduced or changed

None.

### Configuration changes

None.

### Schema or migration changes

None.

### Security changes

None directly; the plan documents rollback of a security-critical cutover.

### Observability changes

None.

### Tests added or modified

None.

### Commits

Local working changes only.

### Pull requests

None.

### Implementation dates

2026-07-13.

### Technical debt introduced

None.

### Known limitations

Assumes wowsociety has a normal revert/deploy pipeline; no custom rollback automation is
specified.

### Follow-up items

- Assign wowsociety engineering owner.
- Review with wowsociety-side reviewer.

### Relationship to the approved plan

Matches `plan.md`: T003 produced a rollback plan document satisfying AC-W03-E01-S004-03.

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W03-E01-S004-03 | Document review against the checklist | N/A (documentation review) | Plan reviewed and accepted, no open finding | review report | unassigned |

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
