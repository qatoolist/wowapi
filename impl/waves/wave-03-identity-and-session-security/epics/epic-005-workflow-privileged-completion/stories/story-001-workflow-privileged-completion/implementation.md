---
id: IMPL-W03-E05-S001
type: implementation-record
parent_story: W03-E05-S001
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Implementation record — W03-E05-S001

## What was actually implemented

- **T4 (ratification):** chose the "reject" path. Added a parsed `RatifyBy` string field to
  `workflow.Definition` and `workflow.Step`. `Definition.Validate` rejects any non-empty
  `ratify_by` declaration (definition-level or step-level) with a clear, fail-closed error
  documenting the interim Wave-0-compatible posture.
- **T5 (durable override audit):** extended `workflow.Runtime` with a required `*audit.Writer`,
  plumbed through `NewRuntime`. `Runtime.Override` now writes a complete `audit.Entry` inside the
  same tenant transaction as the state jump, before any instance/task mutation. The audit row
  records actor, impersonator, grant ID, source/target states, reason, and ratification outcome.
- **Rollback guarantee:** because the audit write occurs inside `txm.WithTenant`, any failure
  (including the injected failure in `TestOverrideAuditFailureRollsBack`) rolls back the entire
  transaction, leaving zero effect from the attempted override.

## Components changed

- `kernel/workflow` — definition model, validation, runtime constructor, `Override`, audit context
  helpers, new tests.
- `kernel` — composition root already supplied `auditWriter` to `NewRuntime`.
- `testkit` — workflowsim coverage test updated for new constructor.

## Files changed

- `kernel/workflow/definition.go` — `RatifyBy` fields and validation rejections.
- `kernel/workflow/runtime.go` — `audit.Writer` dependency, `Override` audit write,
  `withAuditActor`/`grantIDStr` helpers.
- `kernel/workflow/override_audit_test.go` — new tests.
- `kernel/workflow/runtime_test.go` — updated `buildRuntime`.
- `kernel/workflow/runtime_extra_test.go` — updated `buildRT` and standalone `NewRuntime` calls.
- `kernel/workflow/internal_extra_test.go` — updated nil-deps panic test.
- `testkit/workflowsim_cov_test.go` — updated `buildCovRuntime`.
- `kernel/kernel.go` — already wired `auditWriter` into `NewRuntime`.

## Interfaces introduced or changed

- `workflow.NewRuntime(txm, reg, ev, ob, idgen, aud)` now requires a non-nil `*audit.Writer` as the
  sixth argument.
- `Override(ctx, actor, instanceID, to, reason)` signature is unchanged.

## Configuration changes

None.

## Schema or migration changes

None. T5 reuses existing migrations `00017_audit_logs.sql`, `00018_audit_chain.sql`,
`00023_audit_tx_id.sql`, and `00037_audit_hash_version.sql`. T4 "reject" path requires no schema.

## Security changes

- Every privileged override is now durably audited in the same transaction as the state jump.
- A failure to durably audit rolls back the override, preventing an unaudited privileged state jump.
- `ratify_by`-declaring definitions are rejected at boot/validation time, so the runtime cannot be
  asked to enforce an unimplemented ratification control.

## Observability changes

Override operations now produce a queryable `workflow.instance.override` audit row with full
attribution metadata.

## Tests added or modified

- `TestRatifyByDefinitionRejected` — AC-01 boundary test (definition-level and step-level).
- `TestOverrideAuditRowPresent` — AC-02 audit completeness test.
- `TestOverrideAuditFailureRollsBack` — AC-02 fault-injection/rollback test.
- Existing `TestIntegrationOverrideAuthzGate`, `TestIntegrationOverrideFailsClosedWithoutPermission`,
  and `TestIntegrationWorkflowOverride` continue to pass as the T1–T3 regression guard.

## Commits

Working tree changes on HEAD `733ef3e930cbb3f89f5bbc53d8f562c60e426513`.

## Pull requests

Not created in this session.

## Implementation dates

2026-07-13.

## Technical debt introduced

None. The "reject" interim posture is the directive-sanctioned Wave-0 outcome, not debt.

## Known limitations

Real ratification state machine (override-then-ratify, pending-not-yet-effective, rejection reverts)
is intentionally out of scope for this story; deferred to a future story when scope and schedule
allow.

## Follow-up items

- Future story to implement the real ratification state machine, at which point `RatifyBy`
  validation will be relaxed and the runtime will interpret the field.

## Relationship to the approved plan

Matches `plan.md` after recording the "reject" T4 decision. No deviations.
