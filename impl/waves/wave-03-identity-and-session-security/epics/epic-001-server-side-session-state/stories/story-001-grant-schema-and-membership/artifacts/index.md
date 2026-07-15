---
id: W03-E01-S001-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W03-E01-S001
status: complete
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W03-E01-S001 — Artifacts index

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Path | Status |
|---|---|---|---|---|---|---|---|---|
| ART-W03-E01-S001-001 | `identity_grant` migration (up/down) | migration | implementation | New table: status, tenant, actor, impersonated user, approver, reason, activation/expiry/revocation, opaque grant ID | SEC-01 | W03-E01-S001-T001 | `migrations/00039_identity_grant.sql` | produced |
| ART-W03-E01-S001-002 | `identity_grant` RLS policy | schema | implementation | RLS FORCE + `app_platform`-only write grants | SEC-01 | W03-E01-S001-T001 | `migrations/00039_identity_grant.sql` | produced |
| ART-W03-E01-S001-003 | `PrincipalStore.ActiveTenantAccess` implementation | source-code change | implementation | New method querying `user_tenant_access` | SEC-01 | W03-E01-S001-T002 | `adapters/auth/pgprincipal/pgprincipal.go` | produced |
| ART-W03-E01-S001-004 | `Verifier.Actor` unconditional membership call site | source-code change | implementation | Removes the `CapacityID != uuid.Nil` gate on the membership check | SEC-01 | W03-E01-S001-T002 | `kernel/auth/auth.go` | produced |
| ART-W03-E01-S001-005 | `user_tenant_access` data-audit report | pre-implementation baseline | pre-implementation | Confirms or characterizes gaps in "every valid session has a live row" | SEC-01 | W03-E01-S001-T002 | `evidence/index.md` | produced |
| ART-W03-E01-S001-006 | Zero/unknown-tenant rejection logic | source-code change | implementation | Pre-`WithTenantID` rejection path | SEC-01 | W03-E01-S001-T003 | `kernel/auth/auth.go` | produced |
