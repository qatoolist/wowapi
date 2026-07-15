---
id: W03-E01-S001-TASKS-INDEX
type: tasks-index
parent_story: W03-E01-S001
status: complete
created_at: 2026-07-13
updated_at: 2026-07-13
---

# W03-E01-S001 — Tasks index

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W03-E01-S001-T001](task-001-identity-grant-migration.md) | identity_grant migration (SEC-01 T1) | unassigned | complete | none | `identity_grant` exists in the schema with RLS FORCE, the unique partial index, and `app_platform`-only write grants | AC-W03-E01-S001-01 | complete | pass |
| [W03-E01-S001-T002](task-002-active-tenant-access-membership.md) | ActiveTenantAccess + unconditional membership check (SEC-01 T2) | unassigned | complete | none | `PrincipalStore.ActiveTenantAccess` implemented; `Verifier.Actor` calls it unconditionally; data-audit report produced | AC-W03-E01-S001-02 | complete | pass |
| [W03-E01-S001-T003](task-003-zero-unknown-tenant-rejection.md) | Zero/unknown-tenant rejection (SEC-01 T3) | unassigned | complete | W03-E01-S001-T002 | A zero or garbage-UUID tenant claim is rejected before any tenant transaction opens | AC-W03-E01-S001-03 | complete | pass |
| [W03-E01-S001-T004](task-004-independent-review.md) | Independent review | unassigned | complete | W03-E01-S001-T001, W03-E01-S001-T002, W03-E01-S001-T003 | A completed review report confirming the checklist above, recorded as evidence | AC-W03-E01-S001-01, AC-W03-E01-S001-02, AC-W03-E01-S001-03 | complete | pass |
