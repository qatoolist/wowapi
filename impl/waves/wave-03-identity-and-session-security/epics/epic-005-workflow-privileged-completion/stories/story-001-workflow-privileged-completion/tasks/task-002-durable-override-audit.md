---
id: W03-E05-S001-T002
type: task
title: Durable override audit (SEC-02 T5)
status: done
parent_story: W03-E05-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on:
  - W03-E05-S001-T001
acceptance_criteria:
  - AC-W03-E05-S001-02
artifacts:
  - ART-W03-E05-S001-002
evidence:
  - EV-W03-E05-S001-002
  - EV-W03-E05-S001-003
---

# W03-E05-S001-T002 — Durable override audit (SEC-02 T5)

## Task Definition

### Task objective

Persist actor, impersonator, grant ID (from W03-E01-S001), source/target states, reason, and
ratification outcome in a durable audit record, written in the same transaction as the state jump.
An audit-write failure must roll back the override.

### Parent story

W03-E05-S001 — Workflow privileged completion — ratification and durable override audit.

### Owner

unassigned

### Status

todo

### Dependencies

**W03-E05-S001-T001** — this task's audit record includes a ratification-outcome field that cannot
be recorded before T001's chosen path (reject or implement) exists. **Soft, external, for the
grant-ID field specifically: W03-E01-S001's `identity_grant` table must exist with a stable grant-ID
shape.** This gate is restated here per this story's own design goal that a task-level reader cannot
miss it by reading only this file.

### Detailed work

1. Confirm T001's chosen ratification path and its resulting "ratification outcome" value shape.
2. Confirm `authz.Actor.GrantID` exists and is populated from verified grants by W03-E01-S001.
3. Implement the durable audit record using the existing `kernel/audit.Writer` and `audit_logs`
   table: actor, impersonator, grant ID, source state, target state, reason, ratification outcome.
4. Add a required `*audit.Writer` parameter to `workflow.NewRuntime` and update all call sites.
5. Extend `Override`'s existing transaction to write this audit row in the same transaction as the
   state jump, before any instance/task mutation.
6. Implement the audit-write-failure-rolls-back-the-override behavior: if the audit write fails, the
   entire transaction (state jump included) rolls back.
7. Write the audit-present/complete test: confirm every override produces a complete audit row with
   all required fields populated.
8. Write the fault-injection test: use a test-only audit redactor that makes metadata
   canonicalization fail; confirm the override transaction rolls back entirely, leaving zero effect.

### Expected files or components affected

`kernel/workflow/runtime.go` (`Override`'s transaction + `NewRuntime` signature);
`kernel/workflow/override_audit_test.go`; all `NewRuntime` call sites in `kernel/kernel.go`,
`kernel/workflow/*_test.go`, and `testkit/workflowsim_cov_test.go`.

### Expected output

Every override produces a complete, transactional audit row; an injected audit-write failure rolls
back the override, leaving zero effect from the attempted override.

### Required artifacts

ART-W03-E05-S001-002 (the durable audit-record implementation).

### Required evidence

EV-W03-E05-S001-002 (audit-present/complete test output), EV-W03-E05-S001-003 (fault-injection test
output).

### Related acceptance criteria

AC-W03-E05-S001-02.

### Completion criteria

The audit-present/complete test proves every override produces a complete audit row; the
fault-injection test proves an injected audit-write failure genuinely rolls back the override, not
merely asserted in prose.

### Verification method

Direct test execution against a testkit DB with a fault-injection harness, logged output retained as
evidence.

### Risks

RISK-W03-E05-002 (this task's audit-write-failure-rollback behavior is safety-critical — an
implementation bug could either silently allow an unaudited override, worse than today's state, or
incorrectly roll back a legitimate override on a transient audit-write blip) — see epic-level
`risks.md`.

### Rollback or recovery considerations

If the fault-injection test reveals the rollback does not work as intended, this is a blocking
finding for the story's closure, not a deferred follow-up, per RISK-W03-E05-002's own contingency.

## Implementation Record

*Not yet implemented.*

### What was actually implemented

Extended `workflow.Runtime` with a required `*audit.Writer`. `Override` writes a complete
`audit.Entry` (`workflow.instance.override`) inside its tenant transaction before mutating instance
or task state. The entry records actor, impersonator, grant ID, source/target states, reason, and
`ratification_outcome: rejected_interim`. All `NewRuntime` call sites were updated to supply an audit
writer.

### Components changed

`kernel/workflow` runtime; `kernel` composition root; `testkit` coverage test.

### Files changed

- `kernel/workflow/runtime.go`
- `kernel/workflow/override_audit_test.go`
- `kernel/workflow/runtime_test.go`
- `kernel/workflow/runtime_extra_test.go`
- `kernel/workflow/internal_extra_test.go`
- `testkit/workflowsim_cov_test.go`
- `kernel/kernel.go` (already wired)

### Interfaces introduced or changed

`workflow.NewRuntime` adds a required `*audit.Writer` sixth parameter. `Override` signature is
unchanged.

### Configuration changes

*Not applicable.*

### Schema or migration changes

None. Reuses existing `audit_logs` / `audit_chain` schema.

### Security changes

A privileged override cannot commit without a durable audit row; audit-write failure rolls back the
override. This closes the unaudited-override gap.

### Observability changes

Override operations now produce a queryable `workflow.instance.override` audit row with full
attribution metadata.

### Tests added or modified

- `TestOverrideAuditRowPresent` — AC-02 audit completeness.
- `TestOverrideAuditFailureRollsBack` — AC-02 fault-injection/rollback.

### Commits

Working tree changes on HEAD `733ef3e930cbb3f89f5bbc53d8f562c60e426513`.

### Pull requests

Not created in this session.

### Implementation dates

2026-07-13.

### Technical debt introduced

*None anticipated.*

### Known limitations

None.

### Follow-up items

None.

### Relationship to the approved plan

Matches `../plan.md`. No deviations.

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W03-E05-S001-02 | Run the audit-present/complete test; run the fault-injection test with an injected audit-write failure | Local dev or CI, testkit DB, fault-injection harness | Every override produces a complete audit row; an injected audit-write failure rolls back the override, leaving zero effect | audit test report + fault-injection test report | unassigned |

### Actual result

`TestOverrideAuditRowPresent` and `TestOverrideAuditFailureRollsBack` both passed against a real
testkit database.

### Pass or fail

Pass.

### Evidence identifier

EV-W03-E05-S001-002, EV-W03-E05-S001-003.

### Execution date

2026-07-13.

### Commit or revision

HEAD `733ef3e930cbb3f89f5bbc53d8f562c60e426513` (working tree).

### Environment

Local dev; DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable;
WOWAPI_REQUIRE_DB=1.

### Reviewer

Pending independent review.

### Findings

None.

### Retest status

Not required.

### Final conclusion

AC-02 satisfied.

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
