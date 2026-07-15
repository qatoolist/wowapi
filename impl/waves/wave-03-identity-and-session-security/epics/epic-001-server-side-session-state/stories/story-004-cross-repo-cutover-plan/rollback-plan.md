# Rollback plan — wowsociety impersonation-flow cutover (SEC-01 T1/T5)

## Scope

This document covers rollback of the two-repo coordinated cutover of wowsociety's impersonation
flow onto the wowapi framework's `identity_grant` table and privileged-session resolver. It is a
coordination artifact only.

## Failure directions

The cutover can fail in two distinct directions, each with its own rollback path.

### Direction (a) — wowapi-side enforcement causes a wowsociety regression

**Symptoms:**

- wowsociety impersonation success rate drops after the cutover.
- `whoami` returns 401/403 for previously-working impersonated sessions.
- `GrantRejection` rates spike (wrong-tenant, wrong-actor, expired, revoked, unauthorized-approver,
  not-found).
- wowsociety error logs point to framework grant resolution.

**Root-cause possibilities:**

- S001/S002 bug in the framework resolver.
- DEC-Q1 claim-shape mismatch (e.g. `grant_id` format differs from what the resolver expects).
- wowsociety is issuing tokens with `grant_id` values that do not correspond to valid framework
  grants.

**Rollback steps:**

1. **Stop the cutover traffic.** If a feature flag is in use, disable `use_framework_grant_id` in
   wowsociety. If a hard cutover was performed, proceed to step 2.
2. **Revert wowsociety `whoami.go`** to trust the legacy JWT claim fields
   (`impersonator_user_id`, `break_glass`) directly, as it did before the cutover. This restores
   the pre-cutover behavior on the wowsociety side.
3. **Keep wowapi code deployed.** Do not revert wowapi S001/S002. The framework grant table and
   resolver can remain in place; wowsociety simply stops relying on them.
4. **Confirm consistency:**
   - wowsociety `identity_impersonation_session` table is still the authority for its own
     impersonation UX.
   - No wowsociety code remains trying to read `actor.ImpersonatorUserID` from a framework-verified
     source.
5. **Investigate root cause** in a non-production environment before attempting another cutover.
6. **Re-attempt cutover** only after the root cause is fixed and staging validation is re-run.

**Consistency guarantee:** wowsociety returns to its pre-cutover trust model; wowapi's new
resolver is unused by wowsociety but does not interfere because wowsociety tokens no longer carry
`grant_id`.

### Direction (b) — wowsociety-side `grant_id` adoption is broken

**Symptoms:**

- wowsociety cannot start or stop impersonation after Phase 2 changes.
- `identity_impersonation_session` rows reference `grant_id` values that fail framework validation.
- wowsociety schema migration for `grant_id` fails or causes query errors.
- Tests fail in wowsociety CI after the Phase 2 merge.

**Root-cause possibilities:**

- Schema migration error in `identity_impersonation_session.grant_id`.
- `startImpersonation` mints grants with wrong tenant/actor/approver values.
- Token issuance does not include `grant_id`.
- Test fixtures were not fully rewritten.

**Rollback steps:**

1. **Revert the wowservice code change.** Roll back the PR/commit that added:
   - `grant_id` column and index in `identity_impersonation_session`.
   - `startImpersonation` / `stopImpersonation` framework grant integration.
   - `whoami.go` trust change.
   - rewritten test fixtures.
2. **Revert the schema migration.** Run the down-migration to drop `grant_id` from
   `identity_impersonation_session` (or leave it nullable and ignored if rollback must be fast).
3. **Coordinate wowapi-side action:**
   - No wowapi revert is needed if the problem is purely wowsociety-side.
   - If wowsociety tokens were already issuing `grant_id` claims, ensure the revert stops issuing
     them so that wowapi's resolver falls back to legacy claim handling.
4. **Confirm consistency:**
   - wowsociety tokens carry no `grant_id` claim.
   - wowsociety `whoami.go` trusts the legacy claim fields.
   - wowapi `Verifier.Actor` ignores `GrantID` because it is absent and uses legacy claims.
5. **Investigate and fix** in wowsociety staging before re-attempting Phase 2.

**Consistency guarantee:** both repos return to the pre-cutover contract: wowsociety owns
impersonation authority via its own table and JWT claims; wowapi framework grant resolver is
present but not invoked.

## Shared rollback considerations

### Avoiding split-brain

The dangerous intermediate state is: wowapi expects `grant_id` but wowsociety has reverted to
legacy claims. To avoid this:

- Always roll back wowsociety's token-issuance path first (stop emitting `grant_id`).
- Then roll back wowsociety's consumer path (`whoami.go`).
- wowapi can stay as-is.

### Data cleanup

- Do not delete `identity_grant` rows during rollback unless required for data retention policy.
  Unused grant rows are harmless.
- If Direction (b) leaves partially-migrated `identity_impersonation_session` rows with
  `grant_id` set, decide whether to null them out or retain them for forensics. Document the
  decision.

### Rollback time targets

| Direction | Target rollback decision | Target full revert |
|---|---|---|
| (a) | 5 minutes after alert | 15 minutes |
| (b) | 15 minutes after CI/test failure | 1 hour |

These are targets, not SLAs; the actual speed depends on wowsociety's deployment pipeline.

## Rollback verification checklist

- [ ] wowsociety tokens no longer carry `grant_id` (if reverting to legacy model).
- [ ] wowsociety `whoami.go` trusts legacy claims.
- [ ] wowsociety impersonation start/stop works end-to-end.
- [ ] wowsociety `abac_test.go`, `whoami_impersonation_test.go`, and `rls_test.go` pass.
- [ ] wowapi `identity_grant` resolver remains healthy (no errors from unused path).
- [ ] No cross-repo inconsistency: wowapi does not expect `grant_id` from wowsociety when
      wowsociety is in legacy mode.

## Communication plan

1. Alert the on-call wowsociety engineer and the wowapi security owner.
2. Post in the shared incident channel with:
   - failure direction (a) or (b)
   - rollback step in progress
   - ETA for full revert
   - whether wowapi action is required
3. Update the incident timeline as rollback completes.

## Out of scope

- Automated rollback tooling (this plan assumes manual revert via wowsociety's normal deployment
  pipeline).
- Long-term remediation after rollback (tracked as follow-up work).

## Review status

- wowapi-side reviewer: self-review as part of W03-E01-S004 acceptance.
- wowsociety-side reviewer: to be assigned by wowsociety engineering owner.

## Evidence

Review record: EV-W03-E01-S004-003.

## Acceptance

This rollback plan satisfies AC-W03-E01-S004-03.

## Review checklist

- [ ] Covers failure direction (a): wowapi-side enforcement causing wowsociety regression.
- [ ] Covers failure direction (b): wowsociety-side `grant_id` adoption broken.
- [ ] States how to confirm neither repo is left inconsistent.
- [ ] Includes rollback verification checklist.
- [ ] Reviewed by at least a wowapi-side reviewer.

## Sign-off

| Role | Name | Date | Signature/Approval |
|---|---|---|---|
| wowapi reviewer | self-review | 2026-07-13 | — |
| wowsociety reviewer | TBD | — | — |

## Version

1.0 — 2026-07-13.

## Status

Draft → reviewed → accepted as part of W03-E01-S004 closure.

## EOF
