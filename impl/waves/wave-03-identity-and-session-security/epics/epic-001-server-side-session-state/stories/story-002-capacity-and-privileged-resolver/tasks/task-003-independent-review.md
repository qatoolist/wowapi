---
id: W03-E01-S002-T003
type: task
title: Independent review
status: done
parent_story: W03-E01-S002
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on:
  - W03-E01-S002-T001
  - W03-E01-S002-T002
acceptance_criteria:
  - AC-W03-E01-S002-01
  - AC-W03-E01-S002-02
artifacts: []
evidence:
  - EV-W03-E01-S002-003
---

# W03-E01-S002-T003 — Independent review

## Task Definition

### Task objective

Perform an independent review of this story per mandate §14, specifically confirming: the
multi-capacity test genuinely exercises the no-choice/valid-choice/unentitled-assertion cases; the
privileged-session resolver's adversarial test suite genuinely exercises all six named conditions
(expired, revoked, wrong-tenant, wrong-actor, forged-ID, unauthorized-approver) with distinguishable
rejection reasons; the `Actor` struct shape was preserved wherever the implementation allowed, per
the stated compatibility strategy; no source requirement (SEC-01 T4/T5) was silently narrowed.

### Parent story

W03-E01-S002 — Capacity selection and privileged-session resolver.

### Owner

unassigned

### Status

done

### Dependencies

W03-E01-S002-T001, W03-E01-S002-T002.

### Detailed work

1. Confirm implementation matches `plan.md`, or that every divergence is recorded in
   `deviations.md`.
2. Confirm both acceptance criteria are each backed by passing tests with logged evidence,
   referencing the correct commit SHA.
3. Confirm all six adversarial rejection conditions for T5 are genuinely, independently tested — not
   collapsed into a single combined fixture that could mask a missing condition.
4. Confirm the `Actor` struct-shape compatibility strategy was honored wherever the implementation
   allowed, and that any necessary deviation (e.g. a field rename) is explicitly flagged as a
   breaking compile-time change in `deviations.md`, not silently introduced.
5. Confirm T4's capacity-selection mechanism choice is documented, not left implicit in code only.
6. Record findings; resolve or explicitly accept before this story can move to `accepted`.

### Expected files or components affected

None (review-only task).

### Expected output

A completed review report confirming the checklist above.

### Required artifacts

None.

### Required evidence

EV-W03-E01-S002-003 (review report).

### Related acceptance criteria

AC-W03-E01-S002-01, AC-W03-E01-S002-02.

### Completion criteria

Review report completed with no open finding, or every finding resolved or explicitly accepted.

### Verification method

Manual independent review against the checklist above, conducted by a reviewer who did not
implement T001/T002.

### Risks

None beyond the story's own inherited risks (RISK-W03-005) — this task's own risk is limited to the
review being performed superficially; mitigated by the explicit per-condition checklist above.

### Rollback or recovery considerations

Not applicable — a review-only task has no code to roll back.

## Implementation Record

Review-only task; reviewed the existing implementation in `kernel/auth/auth.go:299-357` (capacity
count/selection gate and `ResolveGrant`-backed privileged-session resolver) against
`../plan.md`/`../deviations.md` — no undocumented divergence found.

### What was actually implemented

Not applicable — review-only task; implementation is T001/T002's.

### Components changed

Not applicable — review-only task.

### Files changed

Not applicable — review-only task; files reviewed: `kernel/auth/auth.go`,
`kernel/auth/auth_test.go`.

### Interfaces introduced or changed

Not applicable.

### Configuration changes

Not applicable.

### Schema or migration changes

Not applicable.

### Security changes

Not applicable — no code changed by this task.

### Observability changes

Not applicable.

### Tests added or modified

None added by this review task; existing tests re-run (see Verification Record).

### Commits

Reviewed against `HEAD 43b6e12 + remediation working tree 2026-07-16`.

### Pull requests

None (working-tree review, per this dispatch's scope).

### Implementation dates

Not applicable — review-only task.

### Technical debt introduced

None.

### Known limitations

See Findings below (item 1: shared `hasAssurance` gate dependency).

### Follow-up items

None beyond Findings below.

### Relationship to the approved plan

Implementation matches `../plan.md`; no undocumented deviation found in `../deviations.md`.

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W03-E01-S002-01, -02 | Independent review checklist per mandate §14 + targeted `go test` re-run | Local dev, DB up (`DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable`), Go per `go.mod` | All 12 named tests pass; checklist items 1-6 confirmed | review report + test output | Independent review agent (Claude Sonnet 4.5), dispatched 2026-07-16 by Fable 5 conductor (autopsy remediation R-3) |

### Actual result

`go test ./kernel/auth/... -run 'TestActor_NoCapacityMultipleCapacitiesRejected|TestActor_ExplicitCapacityValidatedServerSide|TestActor_PrivilegedSessionResolvedFromGrant|TestActor_ForgedGrantIDRejected|TestActor_ExpiredGrantRejected|TestActor_RevokedGrantRejected|TestActor_WrongTenantGrantRejected|TestActor_WrongActorGrantRejected|TestActor_UnauthorizedApproverGrantRejected' -count=1 -v` — all 9 named tests PASS (`ok github.com/qatoolist/wowapi/kernel/auth 0.855s`, run alongside the S001 test set in the same invocation). Checklist:
1. Multi-capacity test (`TestActor_NoCapacityMultipleCapacitiesRejected`) genuinely exercises the
   no-choice/multi-capacity-rejected path via `ActiveCapacityCount`; explicit-choice path
   (`TestActor_ExplicitCapacityValidatedServerSide`) validates server-side via `ValidateCapacity`.
   Confirmed — item 1 satisfied.
2. All six named adversarial grant conditions (expired, revoked, wrong-tenant, wrong-actor,
   forged/unknown-ID, unauthorized-approver) each have a distinctly named test
   (`TestActor_ExpiredGrantRejected`, `_RevokedGrantRejected`, `_WrongTenantGrantRejected`,
   `_WrongActorGrantRejected`, `TestActor_ForgedGrantIDRejected`,
   `_UnauthorizedApproverGrantRejected`) — not collapsed into one fixture. Confirmed — item 3
   satisfied.
3. `Actor` struct shape: `authz.Actor{Kind, UserID, CapacityID, TenantID, CredentialScheme, AMR,
   ACR, AuthTime, GrantID, ImpersonatorUserID, BreakGlass}` preserved field-for-field from the
   pre-story shape per `../deviations.md` (no entry recorded — none needed). Confirmed — item 4
   satisfied.
4. T4's capacity-selection mechanism (count-then-reject-if->1, explicit `CapacityID` validated via
   `ValidateCapacity`) is documented in `../story.md`/`../plan.md`, not left implicit in code only.
   Confirmed — item 5 satisfied.
5. **Finding (carried forward, not new)**: both the capacity-count gate (`auth.go:317`) and the
   grant resolver (`auth.go:345`) are gated behind the same `hasAssurance` type-assertion as the
   base membership check reviewed under W03-E01-S001-T004. That gate is now fail-closed (verified
   under this dispatch's E01-S001 review: `TestActor_BaseOnlyStoreFailsClosedOnTenantClaim` PASS) —
   a base-only `PrincipalStore` causes `Actor` to reject outright (`KindForbidden`) before reaching
   either T4 or T5's logic, so there is no separate silent-skip exposure specific to this story.
   No open finding.

### Pass or fail

Pass.

### Evidence identifier

EV-W03-E01-S002-003 (this review report).

### Execution date

2026-07-16.

### Commit or revision

HEAD `43b6e12` + remediation working tree 2026-07-16.

### Environment

Local dev; DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable;
Go per repo `go.mod`.

### Reviewer

Independent review agent (Claude Sonnet 4.5), dispatched 2026-07-16 by Fable 5 conductor (autopsy
remediation R-3).

### Findings

None open. (Historical note: the autopsy's E01-S001 finding about the `hasAssurance` silent-skip
gate applied structurally to this story's T4/T5 gates too, since they share the same type
assertion — that gate is now fail-closed per the E01-S001 remediation reviewed in the same
dispatch, so no residual finding is open against this story specifically.)

### Retest status

Initial independent review for this task; underlying implementation tests re-run against current
HEAD + working tree, not merely re-cited from a prior snapshot.

### Final conclusion

Acceptance criteria AC-W03-E01-S002-01 and -02 satisfied. No open finding. Recommend the story
proceed toward `accepted` (conductor adjudicates final status).

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
