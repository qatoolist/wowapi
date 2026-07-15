---
id: W03-E01-S004-T001
type: task
title: Sequencing plan
status: todo
parent_story: W03-E01-S004
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on: []
acceptance_criteria:
  - AC-W03-E01-S004-01
artifacts:
  - ART-W03-E01-S004-001
evidence:
  - EV-W03-E01-S004-001
---

# W03-E01-S004-T001 — Sequencing plan

## Task Definition

### Task objective

Produce a sequencing plan document stating the repo-by-repo order for the wowsociety impersonation-
flow cutover: wowapi ships SEC-01 T1 (`identity_grant`) and T5 (privileged-session resolver) first;
wowsociety then adopts `grant_id` in its own `identity_impersonation_session` table and
`startImpersonation`/`stopImpersonation` code; only then does the coordinated cutover occur. **This
is a coordination-artifact task. It produces a planning document, not product code — no wowapi or
wowsociety source file is modified by this task.**

### Parent story

W03-E01-S004 — Cross-repo cutover plan for the wowsociety impersonation-flow breaking change.

### Owner

unassigned

### Status

todo

### Dependencies

None (can be drafted in parallel with T002/T003, though its content references the same subject
matter).

### Detailed work

1. Confirm the current state of S001/S002's `identity_grant` schema and resolver contract (at
   whatever point in their own execution this task is performed) as the concrete target the plan
   sequences against.
2. Draft the sequencing plan: state the three-phase order (wowapi ships T1+T5 → wowsociety adopts
   `grant_id` → coordinated cutover), name the specific wowsociety files/tests known to require
   rework (`whoami.go:39,51`, `impersonation.go`, `abac_test.go:52-94`,
   `whoami_impersonation_test.go:31-56`, per PLAN §5.2's own citation), and state explicitly what
   "coordinated cutover" means operationally for this specific change (feature flag, version-gate,
   or hard cutover date — determined during this task's own execution, not invented in advance; see
   `plan.md`'s "Unresolved questions").
3. Circulate the draft for review by at least a wowapi-side reviewer.
4. Record the review outcome.

### Expected files or components affected

None in source code — a new planning document is produced (path TBD at implementation time).

### Expected output

A reviewed sequencing plan document satisfying AC-W03-E01-S004-01.

### Required artifacts

ART-W03-E01-S004-001 (sequencing plan document).

### Required evidence

EV-W03-E01-S004-001 (review report).

### Related acceptance criteria

AC-W03-E01-S004-01.

### Completion criteria

The document exists, names concrete wowsociety files/tests (not vague references), states an
explicit repo-by-repo order, and has passed review with no open finding.

### Verification method

Document review against the checklist in "Detailed work," logged as a review report.

### Risks

Low direct risk (documentation task); the underlying coordination risk it addresses is RISK-W03-002
at epic/wave scope, which this task mitigates but does not eliminate.

### Rollback or recovery considerations

Not applicable — no code to roll back; a document revision is not a "rollback" in the mandate §8.6
sense.

## Implementation Record

### What was actually implemented

Produced the sequencing plan document: `sequencing-plan.md`.

### Components changed

None (documentation only).

### Files changed

- `impl/waves/wave-03-identity-and-session-security/epics/epic-001-server-side-session-state/stories/story-004-cross-repo-cutover-plan/sequencing-plan.md` (new)

### Interfaces introduced or changed

None.

### Configuration changes

None.

### Schema or migration changes

None.

### Security changes

None directly; the plan documents a security-critical cutover.

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

The exact wowsociety engineering owner and timeline remain TBD; the plan states what must be
coordinated rather than inventing specifics (per mandate §18).

### Follow-up items

- Assign wowsociety engineering owner.
- Review with wowsociety-side reviewer.

### Relationship to the approved plan

Matches `plan.md`: T001 produced a sequencing plan document satisfying AC-W03-E01-S004-01.

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W03-E01-S004-01 | Document review against the checklist | N/A (documentation review) | Plan reviewed and accepted, no open finding | review report | unassigned |

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
