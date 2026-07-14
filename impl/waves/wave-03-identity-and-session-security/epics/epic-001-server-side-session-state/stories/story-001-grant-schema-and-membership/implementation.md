---
id: IMPL-W03-E01-S001
type: implementation-record
parent_story: W03-E01-S001
status: complete
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Implementation record — W03-E01-S001

## What was actually implemented

- `identity_grant` migration (`migrations/00039_identity_grant.sql`) creating the server-side
  grant table with FORCE RLS, a partial unique index enforcing at most one active grant per actor,
  and `app_platform`-only write grants.
- `PrincipalStore.ActiveTenantAccess(ctx, userID, tenantID uuid.UUID) error` implemented in
  `adapters/auth/pgprincipal/pgprincipal.go` against `user_tenant_access`.
- `Verifier.Actor` in `kernel/auth/auth.go` updated to call `ActiveTenantAccess` unconditionally
  for every actor carrying a non-zero `TenantID`; the previous `CapacityID != uuid.Nil` gate on
  membership was removed. Capacity validation is still preserved when `CapacityID != uuid.Nil`.
- Zero-UUID tenant claim rejected before any `database.WithTenantID` call with a
  `KindValidation` error; non-existent tenant UUIDs are rejected by `ActiveTenantAccess` as
  `KindForbidden` before `ValidateCapacity` (the only `WithTenantID` consumer in `Actor`) runs.
- Adversarial membership test suite, migration up/down test, RLS catalog extension test,
  zero/unknown-tenant negative tests, and a concurrent-grant-activation test.
- `user_tenant_access` data-audit query executed against the local database; result recorded in
  `evidence/index.md`.

## Components changed

- `kernel/auth` — `PrincipalStore` interface and `Verifier.Actor` behavior.
- `adapters/auth/pgprincipal` — new `Store.ActiveTenantAccess` method.
- `migrations` — new `00039_identity_grant.sql` migration and integration/catalog tests.
- `testkit` — `identity_grant` registered in the RLS census exclusion list.

## Files changed

- `migrations/00039_identity_grant.sql` (new)
- `migrations/migrations_test.go`
- `migrations/identity_grant_test.go` (new)
- `adapters/auth/pgprincipal/pgprincipal.go`
- `adapters/auth/pgprincipal/pgprincipal_test.go`
- `kernel/auth/auth.go`
- `kernel/auth/auth_test.go`
- `testkit/rls_isolation_all_test.go`
- `kernel/notify/service.go` — removed two unused imports that broke the build due to concurrent
  in-flight work (not part of this story's scope, but required for `go test` to compile).
- `migrations/00038_jobs_lease_columns.sql` — added missing `+wowapi:manifest` block.
- `migrations/00040_notify_webhook_lease_columns.sql` — added missing `+wowapi:manifest` block.
- `migrations/00041_bulk_operation_processor_lock.sql` — added missing `+wowapi:manifest` block.

## Interfaces introduced or changed

- `kernel/auth.PrincipalStore` gained `ActiveTenantAccess(ctx context.Context, userID,
  tenantID uuid.UUID) error`. Existing implementations must add this method; the only framework
  implementation (`pgprincipal.Store`) was updated.

## Configuration changes

None.

## Schema or migration changes

- New table `identity_grant` with columns: `id` (UUID PK), `status`, `tenant_id`, `actor_id`,
  `impersonated_user_id`, `approver_id`, `reason`, `activated_at`, `expires_at`, `revoked_at`.
- Partial unique index `identity_grant_one_active_per_actor ON identity_grant (actor_id) WHERE
  status = 'active'`.
- RLS FORCE with tenant-isolation policy and `app_platform` bypass policy.
- `GRANT SELECT, INSERT, UPDATE ON identity_grant TO app_platform` only.

## Security changes

- Closes MATRIX CS-07's top-ranked risk: tenant-isolation bypass via stale/unchecked membership.
- Closes the adjacent zero/unknown-tenant pass-through gap.
- `identity_grant` is writable only by `app_platform` and protected by FORCE RLS.

## Observability changes

None beyond the existing error taxonomy.

## Tests added or modified

- `migrations/identity_grant_test.go` — migration up/down, RLS catalog, grants, partial unique
  index.
- `migrations/migrations_test.go` — expected-file list and static migration-content checks.
- `adapters/auth/pgprincipal/pgprincipal_test.go` — `ActiveTenantAccess` valid/revoked/suspended/
  absent/foreign-tenant cases; concurrent grant activation.
- `kernel/auth/auth_test.go` — zero-tenant and garbage-UUID tenant rejection; updated
  `fakePrincipalStore` to support unconditional membership check.

## Commits

Re-verified at HEAD `733ef3e` with uncommitted W03-E01-S001 changes.

## Pull requests

None yet — tracked in working tree.

## Implementation dates

2026-07-13.

## Technical debt introduced

None.

## Known limitations

- The `user_tenant_access` data audit against the local dev database showed zero gaps; production
  data must be audited before unconditional enforcement is considered safe in a live environment
  (RISK-W03-004).
- DEC-Q1 remains open (human-blocked); implementation proceeds against the safe default documented
  in `story.md`.

## Follow-up items

- Production `user_tenant_access` data audit before full rollout.
- S002/S003 will consume `identity_grant` for server-side capacity selection and resolver work.

## Relationship to the approved plan

Implementation matches `plan.md` with no deviations requiring a `deviations.md` entry. The
`identity_grant` column set exactly matches the story's specification; the migration was routed
through DATA-09 conventions (manifest block, bounded lock/statement timeouts) as a new-table
expand-phase migration.
