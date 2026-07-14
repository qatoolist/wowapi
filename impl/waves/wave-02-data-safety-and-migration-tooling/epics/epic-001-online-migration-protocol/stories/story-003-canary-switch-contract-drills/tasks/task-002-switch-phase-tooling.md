---
id: W02-E01-S003-T002
type: task
title: Switch-phase tooling
status: todo
parent_story: W02-E01-S003
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on:
  - W02-E01-S003-T001
acceptance_criteria:
  - AC-W02-E01-S003-02
artifacts:
  - ART-W02-E01-S003-002
  - ART-W02-E01-S003-006
evidence:
  - EV-W02-E01-S003-002
---

# W02-E01-S003-T002 — Switch-phase tooling

## Task Definition

### Task objective

Implement switch-phase tooling — an observable compatibility flag and dual-schema-version consumer
support — and prove, via the named switch-rollback test, the explicitly-required property:
application rollback after switch, with no destructive `Down` (PLAN DATA-09 T7's bolded acceptance
criterion — "The core safety property this protocol exists to guarantee" per T7's own risk column).

### Parent story

W02-E01-S003 — Canary, switch, and contract-phase tooling with the full CI drill pipeline.

### Owner

unassigned

### Status

todo

### Dependencies

W02-E01-S003-T001 (PLAN T7's "Depends-on" column names T6 — switch follows canary).

### Detailed work

1. Implement the observable compatibility flag: an operator/process-readable indication of which
   schema-version posture the application is running in (exact read interface per `plan.md`
   "Contracts and interfaces").
2. Implement dual-schema-version consumer support: N-1 and N application versions coexisting
   against the same database during the switch window.
3. Ensure no destructive `Down` exists in any tooling-managed migration path — rollback is achieved
   by application-version rollback against the still-compatible expanded schema, never by schema
   reversal.
4. Write the named switch-rollback test (`DATA-09/switch-rollback/`): perform a switch, roll the
   application back, assert correct behavior throughout and that no destructive `Down` was executed.
5. Document the compatibility-flag semantics and the human-decision boundary: "the decision to flip
   in production is human" (PLAN T7's own classification column) — the tooling provides mechanics,
   not the decision.

### Expected files or components affected

New switch-tooling package (location TBD per `plan.md`); possibly persistence for compatibility-flag
state (per `plan.md` "Unresolved questions").

### Expected output

Switch tooling with an observable compatibility flag, dual-version support, and a passing named
switch-rollback test.

### Required artifacts

ART-W02-E01-S003-002 (switch tooling), ART-W02-E01-S003-006 (documentation, shared).

### Required evidence

EV-W02-E01-S003-002 (named switch-rollback test output).

### Related acceptance criteria

AC-W02-E01-S003-02.

### Completion criteria

The named switch-rollback test passes — application rollback after switch works, no destructive
`Down`, flag observable — evidenced against a named commit SHA.

### Verification method

Direct execution of the named switch-rollback test; compatibility-flag observability inspection.

### Risks

PLAN T7's own risk column identifies this as "the core safety property this protocol exists to
guarantee" — a weakened rollback test (e.g. one asserting only that the rollback completes, not that
behavior is correct afterward) would defeat the protocol's purpose; the independent review (T006)
specifically checks the test's assertions.

### Rollback or recovery considerations

This task builds the rollback mechanism itself. For the task's own code: plain revert; no data
impact, no production migration executed.

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

*Not yet implemented — possible compatibility-flag state persistence; recorded here once
implemented.*

### Security changes

*Not applicable.*

### Observability changes

*Not yet implemented — the observable compatibility flag is this task's own observability
deliverable.*

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
| AC-W02-E01-S003-02 | Named switch-rollback test | Local dev or CI, PostgreSQL, dual application versions | Rollback after switch correct; no destructive `Down`; flag observable | integration-test report | unassigned |

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
