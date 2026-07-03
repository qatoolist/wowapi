# Phase 7 — Review Findings

Two parallel critique agents (one security-focused on the rules value/scope/approval path, one
architecture-focused on the workflow definition/runtime) reviewed the rules + workflow slice against
the blueprint (09 patterns, rules resolution, workflow closed step-set) with live DB probes on
2026-07-03. They reproduced eight correctness/authorization gaps and confirmed the tenant/platform
RLS split is production-grade.

| ID | Sev | Finding | Resolution | Status |
|---|---|---|---|---|
| ARCH-60 | high | historical resolution was broken — `lookup` filtered `status = 'active'`, so a value that was active in the past then superseded resolved to the CODE DEFAULT for an `at` inside its old window — reproduced | resolver SQL now `status IN ('active','superseded')` + temporal `effective_from <= at` / `effective_to` window; `TestIntegrationRuleHistoricalSupersededWindow` (v1=5@−10d, v2=9@−2d; resolve@−5d == 5) | **fixed** |
| SEC-40 | high | a rule value that violated the point's `value_schema` was accepted at write and only failed (or silently mis-decoded) on read — reproduced | `kernel/rules/schema.go` `validateAgainstSchema` (type + enum) runs in `Propose` before INSERT; `TestIntegrationRuleSchemaValidationAtWrite` (string for an integer point → `KindValidation`) | **fixed** |
| SEC-39 | high | `workflow.Runtime.Override` had no authz gate — any caller could force an instance to any step | `Override(ctx, actor, id, to, reason)` evaluates `workflow.instance.override` on the instance resource in the request TenantDB; deny → `KindForbidden` | **fixed** |
| SEC-36 | high | vote steps were accepted but the runtime never tallied quorum/pass/window — a vote step would advance on a single ballot (authorization decision on an unenforced control) | definition `Validate` FAILS CLOSED: any `vote` step is rejected at boot until tallying is implemented; `TestValidateFailsClosedOnUnenforcedGating/vote_step` | **fixed (fail-closed)** |
| SEC-37 | high | `policy.min_approvals > 1` was accepted but the runtime advanced on the FIRST approval | `Validate` rejects `min_approvals > 1` at boot (fail-closed); regression sub-test | **fixed (fail-closed)** |
| SEC-38 | med | `policy.self_approval: false` was accepted but submitter-exclusion was not enforced — the requester could approve their own request | `Policy.SelfApproval` changed `bool`→`*bool` (unset ≠ explicit false); `Validate` rejects `self_approval:false` at boot (fail-closed); regression sub-test | **fixed (fail-closed)** |
| ARCH-64 | med | an approval step with only `on_approve` dead-ended a rejection at runtime | `Validate` requires BOTH `on_approve.next` and `on_reject.next` on every approval step; `TestValidateApprovalRequiresBothTransitions` | **fixed** |
| ARCH-62 | med | `rule_versions.created_by` was hardcoded nil — no audit link to the proposing actor | `Propose` reads `database.ActorIDFrom(ctx)` into `created_by` | **fixed** |
| SEC-13 (posture) | — | activation (which changes runtime behavior) must not run on the module-facing app_rt role | `Propose` inserts a DRAFT only (app_rt INSERT); `Activate` supersedes + activates as app_platform via a role-scoped `rule_versions_platform_all` policy; drafts never resolve — `TestIntegrationRuleApprovalGating` | **enforced** |

Reviewer-verified solid (positive): rule_versions RLS (tenant reads only its own + platform rows via
`app_tenant_id_or_null()`; app_rt cannot UPDATE existing versions); the one-active-per-scope EXCLUDE
gist constraint; resolution precedence org-ancestry → tenant → platform → code default; workflow
optimistic locking; same-tx outbox emission on transitions; definition immutability per (key,
version); domain-neutral boundary (`scripts/lint_boundaries.sh` — rules/workflow import kernel/* only).

Residual risk (honest):
- Vote tallying, `min_approvals > 1`, and `self_approval` exclusion are **not implemented** — they are
  rejected at boot rather than silently mis-enforced (per R7: fail-closed is the acceptable posture
  for an unshipped control). Implementing them is a future phase; definitions relying on them will
  not load until then.
- Rule schema validation is a FOCUSED validator (top-level `type` + `enum`), not full JSON Schema;
  nested constraints are not checked at write time. Sufficient for the shipped rule points; a full
  validator can replace it without call-site changes.
