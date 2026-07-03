# Phase 5 — Proof Bundle

Scope (phase-plan row 5): public module SDK (full Context for current capabilities), Kernel + App
composition root / boot lifecycle, seed loader, private neutral fixture module, public testkit +
contract suite, external scratch-consumer test. Date: 2026-07-03.

## 1. Decision evidence
D-0040 (Context accessor scope), D-0041 (Kernel+App boot wires the evaluator + gates registries),
D-0042 (declarative seed loader), D-0043 (scratch-consumer test). Post-review: D-0044 (seed
ownership covers role grants + granted_via; grant reconciliation), D-0045 (seeds run as
app_platform; app_tenant_id_or_null for hybrid-table RLS), D-0046 (diff-based RLS contract check).

## 2. Discussion evidence
- The boot ordering problem (evaluator needs the permission registry, which is populated during
  module Register): resolved with a shared registry pointer — the evaluator holds it, modules fill
  it during Register, and boot gates on its Err() before serving. The registry is complete before
  any request runs.
- Seed privilege: the review showed the contract validated seeds under superuser, never testing the
  SEC-13 boundary. Running as a real app_platform login surfaced an RLS interaction (the strict
  app_tenant_id() raises for a tenant-less platform connection writing NULL-template rows).
  Resolved with a forgiving app_tenant_id_or_null() scoped to the roles/policies hybrid tables,
  keeping the strict loud fail-closed for pure tenant tables.
- Contract integrity: the RLS check was name-prefix-based (evadable) and seed idempotency was
  no-error (not no-change). Both tightened — diff-based RLS over created tables, checksum-based
  seed idempotency.

## 3. Critique/review evidence
`review-findings.md`: 9 findings (2 high seed-ownership/privilege reproduced, several medium). All
fixed with regression tests (seed ownership units + the strengthened contract) or accepted with
tracking. The boot gates and import cleanliness were reviewer-verified/reproduced as sound.

## 4. Implementation evidence
New: `kernel/kernel.go`, `kernel/seeds/`, `app/boot.go`, `testkit/{contract,fixtures,authz_asserts,
consumer_test,contract_requests_test}.go`, `internal/testmodules/requests/`. Changed:
`module/module.go` + `app/context.go` (Phase 5 accessors + bootState), `testkit/db.go` (app_platform
pool), migrations 00001/00006 (app_tenant_id_or_null + hybrid-table policies), Makefile
(test-contract). Team: 1 implementation agent (neutral fixture module) + lead (seeds, kernel/app
boot, testkit contract + fixtures, all review fixes); 1 comprehensive review agent.

## 5. Verification evidence
`command-log.md`: seed loader units, boot/contract build, boundary lint, the module contract suite
green, the external scratch-consumer green (host + tools container), full `make ci` +
`make test-integration` (no Phase 2/4 regression from the migration change).

## 6. Acceptance evidence
`acceptance-map.md`: all 14 Phase 5 exit criteria mapped; the two headline criteria (external repo
imports wowapi; module registers everything without framework edits) proven by the scratch-consumer
and the neutral fixture. Carried forward: deploy-time seed runner as app_platform (Phase 10/CLI),
durable audit (Phase 6), later-phase Context accessors, moduleContext fail-fast (hardening).
Graphify `extract` blocked on LLM key (R11).
