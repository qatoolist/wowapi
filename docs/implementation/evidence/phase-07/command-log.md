# Phase 7 — Command Log

Exact commands with exit status; summarized output where long. Date: 2026-07-03.

| # | Command | Exit | Output summary |
|---|---|---|---|
| 1 | `make migrate` (00008/00009 on fresh schema) | 0 | 9 migrations applied: rule_definitions/rule_versions, workflow_definitions/instances/tasks/assignees + RLS/exclusion/grants |
| 2 | `DATABASE_URL=… go test ./kernel/rules/` | 0 | resolution precedence (default→platform→tenant), historical `at`, approval gating (draft doesn't resolve; activate as app_platform); registry key validation |
| 3 | `DATABASE_URL=… go test ./kernel/workflow/` (agent) | 0 | definition validation (orphan/dangling/unreachable/unknown-action), linear approval Start→approve→auto→terminal, reject→rejected, non-assignee denied, optimistic-lock conflict, same-tx outbox events, WorkflowSim, SLA sweep idempotency |
| 4 | `go build ./...` (kernel/rules + kernel/workflow wired into kernel/app/module Context) | 0 | Rules/RulesResolver/Workflows/WorkflowRuntime accessors; boot gates Rules().Err()/Workflows().Err() |
| 5 | `sh scripts/lint_boundaries.sh` | 0 | OK — workflow imports kernel/* only; domain-neutral |
| 6 | `unset DATABASE_URL; make ci` | 0 | vet, boundary lint, unit, race, build green |
| 7 | `make test-integration` (all packages) | 0 | rules/workflow/authz/outbox/jobs/relationship/resource/testkit integration green |
| 8 | `docker compose run --rm tools ... go test -run Integration ./kernel/rules/ ./kernel/workflow/` | 0 | rules + workflow integration green inside the tools container (container-first) |
| 9 | (review pass) `go test ./kernel/rules/` after ARCH-60/SEC-40/ARCH-62 fixes | 0 | + `TestIntegrationRuleHistoricalSupersededWindow` (v1=5@−10d, v2=9@−2d; resolve@−5d==5), `TestIntegrationRuleSchemaValidationAtWrite` (string→integer point → KindValidation) |
| 10 | (review pass) `go test ./kernel/workflow/` after SEC-36/37/38/ARCH-64/SEC-39 fixes | 0 | + `TestValidateFailsClosedOnUnenforcedGating` (vote/min_approvals>1/self_approval:false rejected at boot), `TestValidateApprovalRequiresBothTransitions`, Override authz gate |
| 11 | `make ci` (host, post-fix) | 0 | vet, boundary lint, unit, race, build green |
| 12 | `make ci-container` | 0 (2nd run) | **1st run FAILED**: `TestVerify_TamperedSignature` (kernel/auth) — trailing base64url sig char carries only discarded padding bits, so the flip left the signature valid (host passed by luck of the random subject UUID). Root-caused, fixed to flip the first char; `-count=200` stable; re-run green |
| 13 | `make test-integration` (host, post-fix) | 0 | rules/workflow/authz/outbox/jobs/relationship/resource/testkit integration green |
