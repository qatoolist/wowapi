# Phase 4 — Acceptance Map

Phase 4 exit criteria (Goal 2 Phase 4 + phase-plan row 4 + blueprint 01 §3) → proof.

| # | Criterion | Proof |
|---|---|---|
| 1 | OIDC verifier + principal/actor model | `kernel/auth/`; 11 tests incl. alg-confusion rejection; `authz.Actor` (user/system/webhook + impersonation/break-glass) |
| 2 | Capacities, roles/permissions/assignments | migration 00004 (acting_capacities), 00006 (roles/permissions/role_permissions/actor_assignments); `kernel/authz` |
| 3 | **Deny by default** (acceptance #4) | `TestDenyByDefault`, `TestFilterNoGrantDeniesAll`, `TestUnregisteredPermissionIsError`; every Evaluate error path returns `Decision{}` (deny) |
| 4 | **Authz matrix** | `TestAuthzMatrix` (table-driven: org member/approver/org admin/tenant admin/vendor × create/approve/read/admin × org/resource/tenant targets) + RBAC/ReBAC/ABAC unit tests |
| 5 | Policy evaluator (ABAC, deny-first) | `kernel/policy`; `TestABACDenyOverridesRBAC`, operator suite; deny-first + fail-closed on unresolved attribute (`TestABACDenyUnresolvedAttributeFailsClosed`) |
| 6 | Relationship framework (ReBAC) | `kernel/relationship`; `TestReBACRelationshipGrant`, `TestIntegrationRelationshipHas` (live edge/expiry/tenant isolation) |
| 7 | Resource registry + mirror | `kernel/resource` (Registry + Registrar); `TestIntegrationRegistrarUpsertBumpsVersion` |
| 8 | **Sensitive denials audited** | `TestSensitiveDenialAudited`, deny-audit in `TestABACDenyOverridesRBAC` + fail-closed audit; AuditSink port (durable audit_logs writer Phase 6) |
| 9 | Deny-by-default proven at DB layer (no self-grant) | `TestIntegrationNoSelfGrantViaAssignments` (RBAC), `TestIntegrationNoSelfGrantViaRelationships` (ReBAC) — live permission-denied |
| 10 | Scope integrity enforced | `TestIntegrationScopeCheckConstraints` (DB CHECKs) + `covers()` guards (SEC-26/29) |
| 11 | RLS on all tenant tables (00004–00006) | ENABLE+FORCE + tenant_id policy on every tenant table; platform-template read admission on roles/policies; verified by security review |
| 12 | Migrations idempotent + reversible | `make migrate` ×2 (0 on rerun); `migrations_test.go` (markers, ordering); goose Down drops tables |
| 13 | Evaluator runs in the request tx (no extra conns) | ARCH-36 fix: Store/Checker/Evaluator take TenantDB; `store_pg_test` drives via WithTenantRO |
| 14 | No package cycles; import law | `scripts/lint_boundaries.sh` OK + depguard (kernel→kernel only) |
| 15 | Container-first verification | host `make ci` + `make test-integration`; prior phases' container gates unaffected |
| 16 | Evidence bundle + reviews | this directory; review-findings.md (security + architecture, 3 reproduced highs fixed) |

Carried forward to Phase 5 (documented in D-0039 / review-findings): app boot wiring of the
evaluator into `module.Context.Authz()` and the boot-time permission-registry gate; ReBAC list
visibility + ABAC deny in `Filter`; PrincipalStore DB adapter; per-request memoization/30s snapshot
cache. Phase 8: `resources.org_id` transition authorization (SEC-31). Graphify `extract` still
blocked (R11).
