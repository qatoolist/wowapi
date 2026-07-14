---
id: W07-E01-S002-T006
type: task
title: Live-update-visibility regression confirmation
status: done
parent_story: W07-E01-S002
owner: W07-Scoping-Dispatch.W07E01S002
created_at: 2026-07-12
updated_at: 2026-07-14
depends_on:
  - W07-E01-S002-T002
acceptance_criteria:
  - AC-W07-E01-S002-06
artifacts: []
evidence:
  - EV-W07-E01-S002-006
---

# W07-E01-S002-T006 — Live-update-visibility regression confirmation

## Task Definition

### Task objective

Preserve live per-request rule updates — explicit non-regression constraint ('B13 is not needed for rules').

### Parent story

W07-E01-S002

### Owner

unassigned

### Status

todo

### Dependencies

W07-E01-S002-T002 (regression is checked against T002's own new query).

### Detailed work

1. Re-run existing rule-update-visibility tests against the new query.
2. Confirm no stale-read regression.
3. Confirm B13 (schema unification/hot-overlay) is genuinely not needed, per PLAN's own explicit
   framing.

### Expected files or components affected

No new files — this task re-runs existing tests.

### Expected output

No stale-read regression; existing rule-update-visibility tests continue passing.

### Required artifacts

None beyond the regression evidence itself.

### Required evidence

EV-W07-E01-S002-006 (regression test output).

### Related acceptance criteria

AC-W07-E01-S002-06.

### Completion criteria

Existing rule-update-visibility tests continue passing against the new query.

### Verification method

Direct re-execution of existing rule-update-visibility tests.

### Risks

Low, per PLAN T5's own risk classification.

### Rollback or recovery considerations

If a regression is found, treat as a correctness defect in T002's own query design and escalate back to that task.

## Implementation Record

### What was actually implemented

No new cache or overlay was introduced. The resolver remains stateless and queries the caller's
`TenantDB` on every request, so an activation committed between requests is observed by the next
request snapshot.

### Tests added or modified

No additional live-update test was needed. The existing resolution, approval-gating, feature-rollout,
and org-scope tests already perform an update followed by another `Resolve`; all were rerun against the
new statement.

### Implementation dates

2026-07-14.

### Known limitations

None at this task's scope. B13 remains explicitly out of scope and unnecessary for live visibility.

### Relationship to the approved plan

Matched T5.

## Verification Record

| Acceptance criterion | Verification method | Required environment | Actual result | Evidence | Reviewer |
|---|---|---|---|---|---|
| AC-W07-E01-S002-06 | Full focused `kernel/rules` package with required DB | PostgreSQL 16.14 container | PASS — live updates visible; no skipped DB tests | EV-W07-E01-S002-006 | pending story independent review |

### Final conclusion

Passed on 2026-07-14; B13 was not required.

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
