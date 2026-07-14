---
id: W03-E05-S001-T001
type: task
title: Ratification decision + implementation (SEC-02 T4)
status: done
parent_story: W03-E05-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on: []
acceptance_criteria:
  - AC-W03-E05-S001-01
artifacts:
  - ART-W03-E05-S001-001
evidence:
  - EV-W03-E05-S001-001
---

# W03-E05-S001-T001 — Ratification decision + implementation (SEC-02 T4)

## Task Definition

### Task objective

Make and record the reject-vs-implement decision for ratification, then implement the chosen path:
either a real definition field and state transition (override-then-ratify happy path;
pending-not-yet-effective; rejection reverts), or an explicit rejection of `ratify_by`-declaring
definitions as an interim, Wave-0-compatible posture.

### Parent story

W03-E05-S001 — Workflow privileged completion — ratification and durable override audit.

### Owner

unassigned

### Status

todo

### Dependencies

None.

### Detailed work

1. Read `kernel/workflow/runtime.go` at this task's actual start commit, confirming ratification is
   still "a bare `TODO` comment with zero implementation," per PLAN's own citation.
2. **Make the reject-vs-implement decision.** Evaluate the actual scope the "implement" path would
   require against RISK-W03-E05-001's own bounding requirement (exactly the three named states, no
   broader ratification framework). Record the decision and its rationale in `../story.md`/
   `../plan.md`.
3. **Decision: reject.** Add a parsed `RatifyBy` field to `workflow.Definition` and `workflow.Step`
   and reject any non-empty value at `Definition.Validate` time with a clear, fail-closed error.
4. Write the rejection-boundary test covering both definition-level and step-level `ratify_by`
   declarations.

### Expected files or components affected

`kernel/workflow/definition.go` (model + validation); `kernel/workflow/override_audit_test.go`
(rejection-boundary test). No migration needed for the "reject" path.

### Expected output

Ratification is either implemented as a real definition field and state transition, or
`ratify_by`-declaring definitions are explicitly rejected — the current `TODO` no longer exists
either way.

### Required artifacts

ART-W03-E05-S001-001 (the ratification implementation, whichever path was chosen, plus the
design-decision record).

### Required evidence

EV-W03-E05-S001-001 (ratification test output — path-dependent).

### Related acceptance criteria

AC-W03-E05-S001-01.

### Completion criteria

The chosen path's own test(s) pass; the decision and its rationale are recorded in `../story.md`/
`../plan.md`, not left implicit.

### Verification method

Direct test execution against the chosen path's own test suite, logged output retained as evidence.

### Risks

RISK-W03-E05-001 (the "implement" path is genuinely greenfield design work and risks expanding beyond
a bounded task) — see epic-level `risks.md`. If the state-machine design is found growing beyond the
three named states during implementation, this task splits further work into a follow-up story rather
than silently expanding.

### Rollback or recovery considerations

If "implement" is chosen and later found to have introduced an unintended bypass in one of the three
states, that state's transition logic is revertible independently of the other two, since each is a
distinct, bounded case per the design's own scoping.

## Implementation Record

*Not yet implemented.*

### What was actually implemented

Added parsed `RatifyBy` string fields to `workflow.Definition` and `workflow.Step`. Extended
`Definition.Validate` to reject any non-empty `ratify_by` value at definition level or step level
with a clear error documenting the interim Wave-0-compatible posture. Recorded the "reject" decision
and rationale in `../story.md` and `../plan.md`.

### Components changed

`kernel/workflow` definition model and validation.

### Files changed

- `kernel/workflow/definition.go`
- `kernel/workflow/override_audit_test.go`
- `../story.md`, `../plan.md`

### Interfaces introduced or changed

Additive `RatifyBy string` field on `Definition` and `Step`. No other signature changes.

### Configuration changes

*Not applicable.*

### Schema or migration changes

None. The "reject" path requires no persisted ratification state.

### Security changes

Definitions that rely on unimplemented ratification gating are rejected at boot/validation time,
preventing a fail-open reliance on a missing control.

### Observability changes

*Not applicable.*

### Tests added or modified

- `TestRatifyByDefinitionRejected` in `kernel/workflow/override_audit_test.go` (definition-level and
  step-level subtests).

### Commits

Working tree changes on HEAD `733ef3e930cbb3f89f5bbc53d8f562c60e426513`.

### Pull requests

Not created in this session.

### Implementation dates

2026-07-13.

### Technical debt introduced

*None anticipated. If "reject" is chosen, the interim posture itself is not technical debt — it is
the directive's own sanctioned Wave-0-compatible outcome.*

### Known limitations

Real ratification state machine not implemented; deferred to a future story.

### Follow-up items

Future story to implement the real ratification state machine (override-then-ratify,
pending-not-yet-effective, rejection reverts) and relax `Validate` accordingly.

### Relationship to the approved plan

Matches `../plan.md` after recording the "reject" decision. No deviations.

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W03-E05-S001-01 | Run the rejection-boundary test | Local dev or CI, testkit DB | `ratify_by`-declaring definitions rejected at validation time with a clear error; decision recorded | ratification rejection-boundary test report | unassigned |

### Actual result

`TestRatifyByDefinitionRejected` passed for both definition-level and step-level `ratify_by`
declarations.

### Pass or fail

Pass.

### Evidence identifier

EV-W03-E05-S001-001.

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

AC-01 satisfied.

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
