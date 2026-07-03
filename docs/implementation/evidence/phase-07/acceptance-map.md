# Phase 7 — Acceptance Map

Phase 7 exit criteria (Goal 2 Phase 7 + phase-plan row 7 + blueprint 09 rules/workflow) → proof.

| # | Criterion | Proof |
|---|---|---|
| 1 | Rule registry (typed points, module-owned keys) | `kernel/rules/rules.go` `Registry`/`Point`; `TestRegistryValidatesKeys` (malformed + foreign-module keys rejected) |
| 2 | Immutable versioned rule rows | migration `00008_rules.sql` `rule_versions` (append-only; status lifecycle draft→active→superseded) |
| 3 | One-active-per-scope invariant | `EXCLUDE USING gist` on (key, scope, tenant) WHERE status='active' |
| 4 | Resolution precedence org-ancestry → tenant → platform → default | `kernel/rules/resolver.go`; `TestIntegrationRuleResolutionPrecedence` (default→platform→tenant override) |
| 5 | **Historical resolution at a point in time** | `TestIntegrationRuleHistoricalResolution` + `TestIntegrationRuleHistoricalSupersededWindow` (ARCH-60: superseded value still resolves inside its old window) |
| 6 | **Value validated against schema at WRITE** | `kernel/rules/schema.go`; `TestIntegrationRuleSchemaValidationAtWrite` (SEC-40) |
| 7 | Approval gating (draft doesn't resolve; activation is platform-gated) | `Propose` drafts as app_rt; `Activate` as app_platform; `TestIntegrationRuleApprovalGating` (SEC-13) |
| 8 | Proposing-actor audit link | `rule_versions.created_by` from `ActorIDFrom(ctx)` (ARCH-62) |
| 9 | Workflow definition (closed step set, versioned/immutable) | `kernel/workflow/definition.go` `validStepTypes`; `TestParseDefinitionValidLinear`, `TestRegistryDuplicateDefinition` |
| 10 | Definition validation at boot (orphans/dangling/unreachable/unknown refs) | `Validate`; `TestValidateOrphanStep`/`DanglingTransition`/`NoTerminalReachable`/`UnknownAutoAction`/`UnknownResolver` |
| 11 | **Fail-closed on unenforced gating** | `TestValidateFailsClosedOnUnenforcedGating` (vote / min_approvals>1 / self_approval:false rejected — SEC-36/37/38) |
| 12 | Approval step completeness | `TestValidateApprovalRequiresBothTransitions` (both on_approve+on_reject required — ARCH-64) |
| 13 | Runtime: approve→auto→terminal, reject→rejected, non-assignee denied | `kernel/workflow/runtime_test.go` (agent-authored happy/deny paths) |
| 14 | Optimistic locking on transitions | runtime version-CAS; conflict test |
| 15 | Same-tx outbox emission on transition | runtime writes events in the business tx |
| 16 | **Override is authz-gated** | `Runtime.Override` evaluates `workflow.instance.override` (SEC-39) |
| 17 | Tenant isolation (rules + workflow rows) | `rule_versions`/`workflow_*` RLS (tenant + platform hybrid via `app_tenant_id_or_null()`) |
| 18 | Module.Context accessors wired | Rules/RulesResolver/Workflows/WorkflowRuntime on `module.Context`; boot gates `Rules().Err()`/`Workflows().Err()` |
| 19 | Domain-neutral boundary | `scripts/lint_boundaries.sh` OK — rules/workflow import kernel/* only |
| 20 | Container-first verification | host `make ci` + `make test-integration`; `make ci-container` green |
| 21 | Evidence bundle + review | this directory; review-findings.md (8 reproduced findings fixed, incl. 5 high) |

Carried forward: vote tallying + `min_approvals>1` + `self_approval` exclusion are fail-closed
(rejected at boot) pending implementation in a later phase; rule schema validation is a focused
`type`/`enum` validator (full JSON Schema deferred). Later-phase Context accessors (Documents → 8,
Notify/Webhooks → 9) arrive with their packages. Graphify `extract` blocked on LLM key (R11).
