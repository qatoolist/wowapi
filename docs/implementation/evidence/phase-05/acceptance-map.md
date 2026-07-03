# Phase 5 — Acceptance Map

Phase 5 exit criteria (Goal 2 Phase 5 + phase-plan row 5 + blueprint 06/08 §2/11) → proof.

| # | Criterion | Proof |
|---|---|---|
| 1 | Public `module` SDK (full Context for current capabilities) | `module/module.go` (Routes/Permissions/Resources/Authz/Tx/IDGen/Migrations/Seeds/OpenAPI/Health/Port/ProvidePort/Config/Validator/Logger); D-0040 scope |
| 2 | Kernel + App composition root; boot lifecycle | `kernel/kernel.go` (New builds the evaluator over a shared registry); `app/boot.go` (Register in dep order → validate → seed); D-0041 |
| 3 | **Permission-registry boot gate** (closes Phase 4 deferral) | `app/boot.go` rejects a route whose permission isn't declared; reproduced by the review; module.Context.Authz() now non-nil at boot |
| 4 | Seed loader (declarative YAML → idempotent catalog sync) | `kernel/seeds`; `seeds_test.go` (merge, ownership, strict unknown-key, foreign-grant/granted_via rejection); D-0042/D-0044 |
| 5 | Seeds run under real platform privilege (SEC-13) | contract syncs via the `app_platform` pool; `app_tenant_id_or_null()` for hybrid-table RLS (D-0045) |
| 6 | Module registries (permissions, resource types, roles/rel-types via seeds) | `authz.Registry`, `resource.Registry`, seed-driven roles/rel-types; fed into the boot registry |
| 7 | Private neutral fixture module | `internal/testmodules/requests` (domain-neutral, boundary-lint clean): routes+perms+seeds+migration+port+health via public Context only |
| 8 | Public `testkit` + contract suite | `testkit/contract.go` `RunModuleContract`; fixtures (`CreateTenant/User/Capacity/Org/Role/GrantRole/CreateResource`), `AssertAllowed/Denied` |
| 9 | **Module contract tests pass** | `TestIntegrationRequestsModuleContract`: boots on empty namespace, migrate+seed idempotent (effect-checksum), RLS forced on created tables (diff-based, ARCH-48), rejects invalid config key |
| 10 | **External scratch product repo imports wowapi** (Definition-of-Done 3–4) | `TestIntegrationScratchConsumer`: tmpdir module, `replace` to working tree, public packages only, passes `RunModuleContract` — zero framework edits |
| 11 | Product module registers everything without framework edits | demonstrated by both the neutral fixture and the scratch `widgets` module |
| 12 | No package cycles; import law incl. testkit-as-composer | `scripts/lint_boundaries.sh` OK; depguard; Go blocks external `internal/` imports (verified) |
| 13 | Container-first verification | host `make ci` + `make test-integration`; `docker compose run tools make test-contract` green (external-consumer flow works without host Go) |
| 14 | Evidence bundle + review | this directory; review-findings.md (2 high seed-ownership/privilege issues fixed with regression tests) |

Carried forward: a deploy-time seed RUNNER as app_platform (Phase 10/CLI); durable authz-denial
audit (Phase 6); moduleContext fail-fast construction (hardening). Later-phase Context accessors
(Events/Jobs/Rules/Workflows/Documents/Notify/Webhooks) arrive with their kernel packages (D-0040).
Graphify `extract` blocked on LLM key (R11).
