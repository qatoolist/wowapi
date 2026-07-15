---
id: PLAN-W03-E05-S001
type: plan
parent_story: W03-E05-S001
status: ready
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Plan — W03-E05-S001

Per mandate §8.5. Confirmed facts, planned changes, and implementation assumptions are distinguished
explicitly below. This plan does not pre-judge T4's reject-vs-implement choice — both paths are
described, with the actual choice to be made and recorded at implementation time against the real
scope pressure the "implement" path would create, per RISK-W03-E05-001.

## Proposed architecture

No new package. Both T4 and T5 land within the existing `kernel/workflow` package. T4 either adds a
ratification state field/transition to the existing workflow-definition and runtime-state model
("implement" path), or adds a rejection check at definition-registration or override time for
`ratify_by`-declaring definitions ("reject" path). T5 adds a durable audit-record write to
`Override`'s existing transaction, sourcing the grant-ID field from W03-E01-S001's `identity_grant`
table.

## Implementation strategy

1. Confirm, at this story's actual start commit, that T1-T3 remain intact (re-run or re-review their
   existing test coverage) and that `Override`'s current implementation matches PLAN's own citation.
2. **Make and record the T4 reject-vs-implement decision.** Evaluate the actual scope of the
   "implement" path against RISK-W03-E05-001's own framing (bounded to exactly the three named states)
   before committing to it; if the state-machine design threatens to expand beyond those three
   states, choose "reject" instead, or split further work into a follow-up story rather than silently
   expanding this one.
3. **Decision: reject.** Add a parsed `ratify_by` field to `workflow.Definition` and `workflow.Step`
   and reject any non-empty value at `Validate` time with a clear, fail-closed error documenting the
   interim Wave-0-compatible posture. This is the chosen T4 path; it requires no schema change and
   keeps the scope bounded to a single story.
4. Confirm W03-E01-S001's `identity_grant` table exists with a stable grant-ID shape before
   implementing T5's grant-ID field specifically. If not yet accepted, T5's grant-ID field is blocked
   on that specific input — T4 is not blocked by this.
5. Implement T5: extend `Override`'s existing transaction to write a durable audit row (actor,
   impersonator, grant ID, source/target states, reason, ratification outcome) in the same
   transaction as the state jump. Add an `*audit.Writer` parameter to `workflow.NewRuntime` and
   update all call sites.
6. Implement the audit-write-failure-rolls-back-the-override behavior: if the audit write fails, the
   entire transaction (including the state jump) rolls back, leaving zero effect from the attempted
   override. Prove this with a test-only audit redactor that makes canonicalization fail.
7. Write the test suite: T4's rejection-boundary test; T5's audit-present/complete test and
   fault-injection test proving rollback-on-audit-failure.
8. Re-run or re-review T1-T3's existing test coverage as part of this story's own review, confirming
   no regression.

## Expected package or module changes

`kernel/workflow` (`runtime.go`: `Override`, plus wherever the ratification state machine or
interim-reject check, and the audit write, are implemented — exact files TBD at implementation time).

## Expected file changes where determinable

- `kernel/workflow/runtime.go` — `Override`'s transaction extended for T5's audit write; the T4
  interim-reject logic implemented in `Definition.Validate`; `NewRuntime` signature extended with an
  `*audit.Writer` parameter.
- `kernel/workflow/definition.go` — additive `RatifyBy` field on `Definition` and `Step`; validation
  rejects non-empty values.
- `kernel/workflow/override_audit_test.go` — new tests for the rejection boundary, the audit row
  completeness, and the audit-write-failure rollback.
- `kernel/kernel.go`, `kernel/workflow/*_test.go`, `testkit/workflowsim_cov_test.go` — updated
  `NewRuntime` call sites to supply an audit writer.
- No new migration: T5 reuses the existing `audit_logs` / `audit_chain` tables (migrations 00017,
  00018, 00023, 00037). T4 "reject" path requires no ratification schema.

## Contracts and interfaces

`workflow.NewRuntime` adds a required `*audit.Writer` parameter; all callers must supply one.
`Override`'s external signature is unchanged. `Definition` and `Step` gain an additive `RatifyBy`
string field.

## Data structures

T5's audit record is written via `kernel/audit.Entry` and persisted in the existing `audit_logs`
table. Fields used: `action` = `workflow.instance.override`; `entity_type` = `workflow_instance`;
`entity_id` = instance id; `old_value` = source state; `new_value` = target state; `reason` = override
reason; `actor_kind` = actor kind; `impersonator_id` = impersonator user id; `metadata` carries
`grant_id`, `source_state`, `target_state`, and `ratification_outcome` (`rejected_interim`).

T4's `RatifyBy` field is a string on `Definition` and `Step`, parsed but rejected at validation time.

## APIs

No public HTTP API surface is added or changed by this story — this is an internal `kernel/workflow`
transaction and state-model change.

## Configuration changes

None anticipated.

## Persistence changes

None. T5 reuses the existing `audit_logs` / `audit_chain` schema. T4's "reject" path requires no new
schema.

## Migration strategy

Routed through W02-E01's DATA-09 protocol per this wave's entry criteria. T5's audit-table change is
additive (new table or new columns). T4's ratification-field change, if "implement" is chosen, is
also additive to the existing workflow-definition schema.

## Concurrency implications

T5's audit-write-in-the-same-transaction-as-the-state-jump requirement is itself the primary
concurrency/atomicity guarantee this story establishes — no separate concurrency mechanism beyond the
existing transactional boundary `Override` already uses is anticipated.

## Error-handling strategy

**T5's audit-write failure must roll back the entire override transaction** — this is the story's
most safety-critical error-handling requirement, verbatim from PLAN's own T5 acceptance criterion:
"audit failure rolls back the override." This must be proven by a fault-injection test, not merely
asserted in prose. T4's rejection path (if chosen) must fail closed with a clear, distinguishable
error at the appropriate boundary (definition-registration or override time).

## Security controls

T5's audit-write-failure-rollback behavior is the central security control this story adds: a
privileged override that cannot be durably audited must not be allowed to take effect. T4's chosen
path (whichever it is) must not introduce a new fail-open surface — see `story.md` "Security
considerations."

## Observability changes

T5's durable audit record is itself the primary observability deliverable of this story.

## Testing strategy

- T4 "reject" path: a test that a `ratify_by`-declaring definition is rejected at validation time,
  covering both definition-level and step-level declarations.
- T5: an audit-present/complete test confirming every override produces a complete audit row; a
  fault-injection test proving an injected audit-write failure rolls back the override, leaving zero
  effect from the attempted override — this must be genuinely adversarial (inject a real failure into
  the audit-write path), not a happy-path test relabeled, per RISK-W03-E05-002. The chosen fault
  injection is a test-only audit redactor that makes metadata canonicalization fail.
- Regression: re-run or re-review T1-T3's existing test coverage, confirming no regression to their
  already-verified fail-closed behavior.

## Regression strategy

T5's audit-present/complete test and fault-injection test, run in CI, become the permanent regression
guard against a future change silently reintroducing an unaudited or non-atomic override path. T1-T3's
existing test suite continues to serve as the regression guard for the fail-closed permission check.

## Compatibility strategy

Per PLAN's own wowsociety-impact note, this story carries zero wowsociety compatibility risk — "Not
affected... No required changes, no sequencing constraint." Re-confirmed fresh at this story's own
execution time, not merely trusted from the cited snapshot, though the citation covers every relevant
symbol directly.

## Rollout strategy

T4 and T5 land together in this single story, since T5's audit record includes a ratification-outcome
field that depends on T4's chosen path existing first (a ratification outcome cannot be recorded
before ratification itself — whichever form it takes — is implemented).

## Rollback strategy

T4's chosen path (whichever it is) and T5's audit-write extension are both revertible independently
of each other's underlying schema, though T5's audit record's ratification-outcome field is
functionally dependent on T4 being in place first. If T5's audit-write-failure-rollback behavior is
found to incorrectly roll back a legitimate override on a transient audit-write blip post-rollout,
this is treated as a blocking finding requiring a fix, not a deferred follow-up, per RISK-W03-E05-002.

## Implementation sequence

Steps 1-8 under "Implementation strategy" above. Step 2 (the T4 reject-vs-implement decision) must be
made and recorded before step 3 (implementation) begins. Step 4 (confirming W03-E01-S001's
`identity_grant` table) must be confirmed before step 5's grant-ID-field-specific work, though T4's
own implementation (steps 2-3) is not gated on it.

## Task breakdown

- **W03-E05-S001-T001** — T4 reject-vs-implement decision + implementation (SEC-02 T4).
- **W03-E05-S001-T002** — T5 durable override audit, including the grant-ID field and the
  audit-write-failure-rollback behavior (SEC-02 T5).
- **W03-E05-S001-T003** — Independent review (mandate §14), with specific confirmation the
  fault-injection test is genuinely adversarial and T1-T3 remain intact.

## Expected artifacts

The T4 ratification implementation (state machine) or interim-reject implementation, plus the
design-decision record; the T5 durable audit-record implementation (schema + transactional write +
rollback-on-failure logic).

## Expected evidence

T4's three named-case test output (or the rejection-boundary test output); T5's audit-present/complete
test output and fault-injection test output; the T1-T3 regression-confirmation output.

## Resolved questions

- **T4 reject-vs-implement choice:** "reject" was chosen at implementation time because the real
  ratification state machine would require persisted instance state and a ratification task
  lifecycle that exceeds this story's bounded Wave-0 scope (RISK-W03-E05-001).
- **Rejection boundary:** definition validation time (`Definition.Validate`), so no instance can be
  started from a `ratify_by`-declaring definition.
- **Audit-table schema:** no new table; T5 reuses the existing `audit_logs` / `audit_chain` tables
  via `kernel/audit.Writer`.
- **`Override`'s external signature:** unchanged. `workflow.NewRuntime` gains a required
  `*audit.Writer` parameter so `Override` can write the audit row inside its transaction.

## Approval conditions

This plan is approved for implementation once: (a) the T4 reject-vs-implement decision is made and
recorded; (b) W03-E01-S001's `identity_grant` table's acceptance status is confirmed (blocking T5's
grant-ID field specifically, not T4); and (c) the owner and reviewer are assigned.
