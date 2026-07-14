---
id: W03-E01-S001-EVIDENCE-INDEX
type: evidence-index
parent_story: W03-E01-S001
status: complete
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W03-E01-S001 — Evidence index

| Evidence ID | Type | Task | Acceptance criteria proven | Execution command | Commit SHA | Result | Status |
|---|---|---|---|---|---|---|---|
| EV-W03-E01-S001-001 | migration test report | W03-E01-S001-T001 | AC-W03-E01-S001-01 | `go test ./migrations/... -run 'TestIdentityGrantMigrationUpDown' -count=1` with `WOWAPI_TEST_DSN` set | 1626b11 | Pass: migration applies and reverts cleanly | accepted |
| EV-W03-E01-S001-002 | RLS catalog report | W03-E01-S001-T001 | AC-W03-E01-S001-01 | `go test ./migrations/... -run 'TestIdentityGrantRLSCatalog' -count=1` with `WOWAPI_TEST_DSN` set | 1626b11 | Pass: FORCE RLS, policies, partial unique index, and app_platform-only grants confirmed | accepted |
| EV-W03-E01-S001-003 | adversarial test report | W03-E01-S001-T002 | AC-W03-E01-S001-02 | `go test ./adapters/auth/pgprincipal/... -run 'TestActiveTenantAccess' -count=1` with `WOWAPI_TEST_DSN` set | 1626b11 | Pass: revoked/suspended/absent/foreign-tenant membership rejected | accepted |
| EV-W03-E01-S001-004 | baseline/audit report | W03-E01-S001-T002 | AC-W03-E01-S001-02 | SQL query against local `wowapi` database (see below) | 1626b11 | Clean: 0 active memberships, 0 distinct users, 0 distinct tenants, 0 capacity-without-membership gaps | accepted |
| EV-W03-E01-S001-005 | negative test report | W03-E01-S001-T003 | AC-W03-E01-S001-03 | `go test ./kernel/auth/... -run 'TestActor_ZeroTenantRejected|TestActor_GarbageTenantRejected' -count=1` | 1626b11 | Pass: zero-tenant → KindValidation; garbage-tenant → KindForbidden | accepted |
| EV-W03-E01-S001-006 | review report | W03-E01-S001-T004 | AC-W03-E01-S001-01, AC-W03-E01-S001-02, AC-W03-E01-S001-03 | Independent review checklist per mandate §14 | 1626b11 | No open issues found; review complete | accepted |

## user_tenant_access data-audit query

Executed 2026-07-13 against the local `wowapi` database:

```sql
WITH gap AS (
  SELECT DISTINCT u.id AS user_id
    FROM users u
    JOIN acting_capacities ac ON ac.user_id = u.id
   WHERE u.status = 'active'
     AND ac.status = 'active'
     AND ac.valid_to IS NULL
     AND NOT EXISTS (
       SELECT 1 FROM user_tenant_access uta
        WHERE uta.user_id = u.id
          AND uta.tenant_id = ac.tenant_id
          AND uta.status = 'active'
          AND uta.valid_to IS NULL
     )
)
SELECT
  (SELECT count(*) FROM user_tenant_access WHERE status = 'active' AND valid_to IS NULL) AS active_memberships,
  (SELECT count(DISTINCT user_id) FROM user_tenant_access WHERE status = 'active' AND valid_to IS NULL) AS distinct_users,
  (SELECT count(DISTINCT tenant_id) FROM user_tenant_access WHERE status = 'active' AND valid_to IS NULL) AS distinct_tenants,
  (SELECT count(*) FROM gap) AS users_with_capacity_but_no_membership;
```

Result:

```
active_memberships | distinct_users | distinct_tenants | users_with_capacity_but_no_membership
-------------------+----------------+------------------+--------------------------------------
                 0 |              0 |                0 |                                     0
```

Interpretation: against the local dev dataset there are zero active memberships and zero users
with an active capacity but no live `user_tenant_access` row. Production data must be audited
before unconditional enforcement is considered safe in a live environment (RISK-W03-004).

## Fail-first evidence

The new tests were designed to assert behavior the pre-change code did not implement:

- `TestActiveTenantAccess` (pgprincipal) and the updated `fakePrincipalStore` require live
  `user_tenant_access`; before the fix, `Verifier.Actor` skipped membership checks entirely for
  capacity-less actors.
- `TestActor_ZeroTenantRejected` and `TestActor_GarbageTenantRejected` assert pre-`WithTenantID`
  rejection; before the fix, zero/non-existent tenants reached actor construction or downstream
  capacity validation.

These tests fail if the unconditional-membership check or zero-tenant guard is removed.
