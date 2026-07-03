# Phase 7 — Proof Bundle

Scope (phase-plan row 7): rules engine (registry, versioned rows, temporal resolution, approval
gating) + workflow engine (closed step set, boot validation, authz-gated runtime), migrations
00008/00009. Date: 2026-07-03.

## 1. Decision evidence
D-0051 (migrations 00008/00009, custom rules+workflow engines behind interfaces), D-0052 (rule
resolution algorithm: org-ancestry → tenant → platform → code default, temporal `at`), D-0053
(workflow closed step set + boot validation), D-0054 (Phase 7 review fixes: temporal resolution
includes superseded, write-time schema validation, draft/activate privilege split, workflow
fail-closed gating + Override authz gate).

## 2. Discussion evidence
- Rule activation vs the module role: activation changes runtime behavior, so it must not run on the
  module-facing app_rt role. Resolved by splitting `Propose` (app_rt inserts a DRAFT that never
  resolves) from `Activate` (app_platform supersedes + activates), mirroring the Phase 5/6
  app_platform config-write posture (SEC-13).
- Unenforced workflow gating: the runtime does not yet tally votes, enforce `min_approvals > 1`, or
  exclude self-approval. Rather than silently mis-tally (an authorization decision on an unenforced
  control), the definition validator FAILS CLOSED — such definitions are rejected at boot. Per R7
  this is the acceptable posture for an unshipped control; the alternative (accept + mis-enforce)
  was rejected.
- Historical resolution: the review reproduced that filtering `status='active'` dropped superseded
  versions, so a point-in-time query inside a past window fell through to the code default. Fixed by
  resolving over `status IN ('active','superseded')` within the temporal window.

## 3. Critique/review evidence
`review-findings.md`: 8 reproduced findings — 5 high (ARCH-60 historical resolution, SEC-40
write-time schema validation, SEC-39 Override authz gate, SEC-36/37 vote + min_approvals fail-closed)
and 3 med (SEC-38 self_approval, ARCH-64 approval completeness, ARCH-62 created_by). All fixed with
regression tests (or fail-closed + enforced). Two parallel review agents (security + architecture);
tenant/platform RLS split and resolution precedence verified solid.

## 4. Implementation evidence
New: `kernel/rules/` (rules, resolver, store, schema), `kernel/workflow/` (definition, registry,
runtime), migrations `00008_rules.sql` / `00009_workflow.sql`. Changed: `kernel/kernel.go` (wire
Rules/RulesResolver/Workflows/WorkflowRuntime), `module/module.go` + `app/context.go` + `app/boot.go`
(accessors + boot gates), `testkit` (Platform pool already present from Phase 5/6).
Team: 1 implementation agent (workflow definition/runtime + tests) + lead (rules engine, migrations,
all review fixes, wiring); 2 review agents (security, architecture).

## 5. Verification evidence
`command-log.md`: rules integration (precedence, historical + superseded-window, write-time schema
validation, approval gating), workflow definition validation (orphan/dangling/unreachable/unknown +
fail-closed vote/min_approvals/self_approval + approval completeness), workflow runtime
(approve/reject/deny/optimistic-lock/same-tx outbox/Override authz), full `make ci` +
`make test-integration` host and `make ci-container`. The auth flaky (TestVerify_TamperedSignature —
trailing base64 padding-bit flip) was root-caused and fixed to flip a significant bit; 200× stable.

## 6. Acceptance evidence
`acceptance-map.md`: all 21 Phase 7 exit criteria mapped to named tests. Carried forward: vote
tallying + min_approvals>1 + self_approval exclusion are fail-closed pending implementation; rule
schema validation is a focused type/enum validator (full JSON Schema deferred). Graphify `extract`
blocked on LLM key (R11).
