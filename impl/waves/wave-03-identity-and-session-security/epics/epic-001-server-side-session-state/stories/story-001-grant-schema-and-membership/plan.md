---
id: PLAN-W03-E01-S001
type: plan
parent_story: W03-E01-S001
status: ready
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Plan — W03-E01-S001

Per mandate §8.5. Confirmed facts, planned changes, and implementation assumptions are distinguished
explicitly below.

## Proposed architecture

A new `identity_grant` table becomes the framework's server-side source of truth for break-glass
and impersonation activation state, keyed on an opaque grant ID. `PrincipalStore` gains a new
method (`ActiveTenantAccess`) queried unconditionally by `Verifier.Actor`. No new package is
introduced; the change lands within the existing `kernel/auth` and principal-store packages plus a
new migration.

## Implementation strategy

1. Confirm the exact current state of `Verifier.Actor` (`kernel/auth/auth.go:181-208`) and
   `PrincipalStore`'s existing methods at this story's actual start commit — re-read, don't trust
   the cited line numbers blindly (they may have drifted since PLAN was written).
2. Author the `identity_grant` migration: columns for status, tenant, actor, impersonated user,
   approver, reason, activation timestamp, expiry timestamp, revocation timestamp, and an opaque
   grant ID (UUID). Apply RLS FORCE per blueprint 03 §1. Add a unique partial index enforcing at
   most one active grant per actor (partial on `status = 'active'` or equivalent, exact predicate to
   be finalized against the chosen status enum at implementation time). Restrict INSERT/UPDATE to
   the `app_platform` role.
3. Route the migration through W02-E01's DATA-09 protocol — confirm which phases apply to a
   wholly-new table (no backfill target since it starts empty; expand-phase tooling and CI-drill
   registration likely still apply).
4. Run the `user_tenant_access` data-audit step: confirm every currently-valid session (by whatever
   sampling/full-scan method is practical against the actual data volume) has a live row, or
   characterize the gap.
5. Implement `PrincipalStore.ActiveTenantAccess(ctx, userID, tenantID) error` against
   `user_tenant_access`.
6. Change `Verifier.Actor`'s call site to invoke `ActiveTenantAccess` unconditionally for every
   actor carrying a `TenantID`, removing the `CapacityID != uuid.Nil` gate on this specific check.
7. Add zero/unknown-tenant rejection ahead of any `WithTenantID` call.
8. Write the adversarial test suite: revoked/absent/foreign-tenant membership rejected with a
   validly signed token (T2's acceptance criterion); zero/garbage-UUID tenant claim rejected (T3's
   acceptance criterion).

## Expected package or module changes

`kernel/auth` (`Verifier.Actor`), the principal-store package (new `ActiveTenantAccess` method), a
new migration package/directory for `identity_grant`.

## Expected file changes where determinable

- `kernel/auth/auth.go:181-208` — `Verifier.Actor`'s membership-check call site.
- The principal-store implementation file (PLAN cites `pgprincipal.Store` — exact current file path
  to be confirmed at implementation time).
- A new migration file (numbering to follow the existing `NNNNN_description.sql` convention, e.g.
  following on from whatever the highest-numbered migration is at this story's start commit).

## Contracts and interfaces

`PrincipalStore` gains one new method: `ActiveTenantAccess(ctx context.Context, userID, tenantID
uuid.UUID) error`. This is additive to the interface — existing callers are unaffected unless the
interface is defined as a closed set requiring every implementer to add the method (to be confirmed
at implementation time whether `PrincipalStore` is an interface with multiple implementers or a
concrete type).

## Data structures

`identity_grant` table: `id` (opaque grant ID, UUID PK), `status`, `tenant_id`, `actor_id`,
`impersonated_user_id` (nullable), `approver_id` (nullable), `reason`, `activated_at`,
`expires_at`, `revoked_at` (nullable). Exact column types/nullability to be finalized against the
existing schema's conventions (e.g. how `00002_core_identity.sql` types its own UUID/timestamp
columns) at implementation time.

## APIs

No public HTTP API surface is added by this story specifically (T1-T3 are schema and internal
verification-path changes); the grant-table's own admin/management API, if any, is out of this
story's scope (not named in PLAN's T1-T3 rows).

## Configuration changes

None anticipated.

## Persistence changes

New `identity_grant` table (see "Data structures"). RLS FORCE policy. Unique partial index.

## Migration strategy

Route through DATA-09 (W02-E01). Since `identity_grant` is a new, empty table, the riskiest DATA-09
phases (backfill, dual-version soak) may not apply in their full form — this is an implementation-
time determination, not assumed here. The migration itself should be reviewable in a single
expand-phase step (new table + RLS + index), consistent with DATA-09's "expand-phase tooling:
nullable/default-safe columns, new tables/indexes... don't block traffic" framing.

## Concurrency implications

The unique partial index (one active grant per actor) must hold under concurrent grant-activation
attempts — a race where two concurrent requests both attempt to activate a grant for the same actor
must result in exactly one succeeding, the other rejected by the constraint, not both silently
succeeding.

## Error-handling strategy

`ActiveTenantAccess` returns a distinguishable error type for "no live membership row found" versus
a genuine database/connectivity error, so `Verifier.Actor`'s caller can fail closed (reject the
actor) on the former without conflating it with an infrastructure fault. Zero/unknown-tenant
rejection similarly returns a distinguishable, testable error.

## Security controls

RLS FORCE (mandatory per blueprint 03 §1, not row-security-optional); unique partial index
preventing concurrent multi-grant activation for one actor; `app_platform`-only write restriction
preventing any other role (including `app_rt`) from writing grant rows directly.

## Observability changes

Not mandated by this story's acceptance criteria; a startup/runtime log or metric for membership-
check outcomes is a reasonable implementation-time addition, not required scope.

## Testing strategy

- Fail-first: confirm today's actual behavior (capacity-less actor bypasses membership check) with
  a test that currently passes wrongly (per MATRIX CS-07's framing) before the fix, then fails after
  the fix is reverted / passes correctly after the fix lands.
- Migration up/down test; RLS catalog extension test.
- Adversarial: revoked/absent/foreign-tenant membership rejected with a validly signed token (T2).
- Negative: zero/garbage-UUID tenant claim rejected pre-`WithTenantID` (T3).
- Concurrency test on the unique partial index: N concurrent grant-activation attempts for the same
  actor → exactly one succeeds.
- These map to the SEC-01-required test classes (PLAN §6 SEC-05, mandatory) this story is
  responsible for: zero-tenant, stale membership, revoked capacity (partial — full capacity
  coverage is S002's T4/T5 scope, this story covers the membership-layer portion).

## Regression strategy

The adversarial test suite itself, run in CI, is the regression guard — any future change that
reintroduces a capacity-less membership bypass or a zero-tenant pass-through would fail these tests.

## Compatibility strategy

Preserve the `Actor` struct shape wherever possible (per PLAN's own stated preference — "if the
resolver preserves the `authz.Actor` struct shape... wowsociety compiles unchanged and gets a strict
behavioral improvement"). This story's own T1/T2/T3 changes do not rename or remove any `Actor`
field; the resolver work that might (S002's T5) is out of this story's scope.

## Rollout strategy

Single migration + code change; the data-audit step (see "Implementation strategy" step 4) informs
whether unconditional enforcement can roll out immediately or needs staged enforcement — see
"Rollback strategy" and RISK-W03-004.

## Rollback strategy

The migration's down-path drops `identity_grant` cleanly (no other table references it yet within
this story's scope). The `Verifier.Actor` call-site change can be reverted independently if the
data-audit step or post-rollout monitoring reveals unconditional enforcement is rejecting a
significant volume of currently-valid sessions — in which case, per RISK-W03-004's contingency,
enforcement is staged behind a profile flag rather than reverted outright, and rolled forward once
the underlying data gap is closed.

## Implementation sequence

As listed under "Implementation strategy" above (steps 1-8). The data-audit step (4) must occur
before step 6 (flipping the call site to unconditional) per RISK-W03-004's mitigation.

## Task breakdown

- **W03-E01-S001-T001** — `identity_grant` migration (RLS, unique partial index, `app_platform`-only
  writes) — SEC-01 T1.
- **W03-E01-S001-T002** — `PrincipalStore.ActiveTenantAccess` + unconditional `Verifier.Actor` call
  site, including the `user_tenant_access` data-audit step — SEC-01 T2.
- **W03-E01-S001-T003** — zero/unknown-tenant rejection pre-`WithTenantID` — SEC-01 T3.
- **W03-E01-S001-T004** — independent review (mandate §14), scoped to the SEC-01 test classes this
  story covers: zero-tenant, stale membership, revoked capacity (membership layer).

## Expected artifacts

`identity_grant` migration files; RLS policy; `PrincipalStore.ActiveTenantAccess` implementation;
`Verifier.Actor` call-site change; zero/unknown-tenant rejection logic; the data-audit report.

## Expected evidence

Migration up/down test log; RLS catalog extension test log; adversarial membership test log;
zero/unknown-tenant negative test log; concurrency test log for the unique partial index; the
`user_tenant_access` data-audit report itself.

## Unresolved questions

- Exact current file/line for `PrincipalStore`'s implementation (PLAN cites `pgprincipal.Store` —
  to be confirmed at implementation time, since the package may have been renamed or moved since
  PLAN was written).
- Exact `identity_grant` column types/nullability and the unique partial index's exact predicate —
  to be finalized against the existing schema's conventions at implementation time, not invented
  here.
- Whether `PrincipalStore` is an interface with multiple implementers (requiring every implementer
  to add `ActiveTenantAccess`) or a single concrete type — affects the blast radius of the interface
  change; to be confirmed at implementation time.
- The exact sampling/full-scan method for the `user_tenant_access` data-audit step, and what
  threshold of gap (if any) is acceptable before staging enforcement behind a flag rather than
  enabling it unconditionally — this is a judgment call for implementation time, informed by
  whatever data volume is found, not pre-specified here per mandate §18.
- Which DATA-09 phases materially apply to a wholly-new, empty table — to be determined against
  W02-E01's actual manifest schema once it exists.

## Approval conditions

This plan is approved for implementation once: (a) the unresolved questions above are answered by a
first re-read of `kernel/auth/auth.go` and the principal-store package at story start, (b) the
`user_tenant_access` data-audit step has run and its result is known, and (c) the owner and reviewer
are assigned.
