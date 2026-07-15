# Staging-validation plan — wowsociety impersonation-flow cutover (SEC-01 T1/T5)

## Scope

This document describes how S001's T2 unconditional-membership enforcement and S002's T5
privileged-session resolver are validated against wowsociety staging data before either is made
unconditional/enforced in wowsociety's production path. It is a coordination artifact only.

## Validation objective

Prove that:

1. Every wowsociety staging actor that currently carries a tenant claim still has a live
   `user_tenant_access` row under S001 T2's unconditional check.
2. Every wowsociety staging impersonation session that should be valid resolves to a live
   `identity_grant` row under S002 T5's resolver.
3. Existing wowsociety tests continue to pass after the cutover.

## Environment

- **Source data:** wowsociety staging database, specifically:
  - `user_tenant_access` rows for active users.
  - `identity_impersonation_session` rows (active + recently expired/revoked).
  - `identity_grant` rows minted during Phase 2 of the sequencing plan.
- **Access control:** validate against staging data only. Production-derived data may be used if
  anonymized and access is restricted to the engineers executing the validation, documented in the
  deployment runbook, and approved by wowsociety's security/data owner. No ad-hoc production
  access.
- **Tooling:** use wowsociety's existing migration/validation harness; if none exists, run SQL
  probes and the test suites listed below.

## Validation steps

### V1. Membership data audit (S001 T2)

Run a query that would fail under the new unconditional `ActiveTenantAccess` check:

```sql
SELECT u.id, u.email, t.id AS tenant_id
FROM users u
JOIN tenants t ON ...
LEFT JOIN user_tenant_access uta
  ON uta.user_id = u.id
 AND uta.tenant_id = t.id
 AND uta.status = 'active'
 AND uta.valid_to IS NULL
WHERE uta.id IS NULL
  AND <u,t is a staging actor with a tenant claim>;
```

**Pass criterion:** zero rows. Any non-zero result is a go/no-go blocker until the membership gap
is remediated.

### V2. Grant resolver data audit (S002 T5)

For every active `identity_impersonation_session` row in staging that should still be valid:

1. Confirm a matching `identity_grant` row exists with:
   - `status = 'active'`
   - `expires_at > now()`
   - `revoked_at IS NULL`
   - `tenant_id` matches the session's tenant
   - `actor_user_id` matches the impersonated user
2. Confirm the grant's `impersonator_user_id` and `break_glass` values match what wowsociety
   expects for that session.

**Pass criterion:** 100% of active staging impersonation sessions have a valid, matching grant row.

### V3. Adversarial grant negatives

Insert adversarial grant rows in staging (or use the wowapi adversarial test fixtures) and confirm
the resolver rejects them:

- expired grant
- revoked grant
- wrong-tenant grant
- wrong-actor grant
- unauthorized-approver grant
- forged/unknown grant ID

**Pass criterion:** each case returns the expected `GrantRejection` code and does not populate
`Actor.ImpersonatorUserID`.

### V4. Test-suite regression

Re-run the wowsociety test suites PLAN cites as good regression coverage:

```bash
# wowsociety repo
go test ./internal/modules/identity/... -run 'TestABAC|TestWhoami|TestImpersonation'
go test ./internal/... -run 'RLS|rls'  # rls_test.go
```

**Pass criterion:** all tests pass with no new failures attributable to the cutover.

### V5. End-to-end impersonation smoke test

In wowsociety staging:

1. Start impersonation as a support user.
2. Verify `whoami` returns the impersonator + impersonated identities.
3. Verify the JWT contains `grant_id`.
4. Stop impersonation.
5. Verify a replay of the old JWT (after stop) is rejected.

**Pass criterion:** full happy path and revocation path work.

## Go/no-go criteria

| Gate | Criterion | Blocker? |
|---|---|---|
| G1 | V1 membership audit returns zero gaps. | Yes |
| G2 | V2 grant audit returns 100% match. | Yes |
| G3 | V3 adversarial negatives all rejected with correct reason. | Yes |
| G4 | V4 test suites pass with no new failures. | Yes |
| G5 | V5 end-to-end smoke test passes. | Yes |
| G6 | Observability dashboard shows grant lookups at expected rate. | No (advisory) |

If any blocker gate fails, abort the cutover and return to Phase 2 rework.

## Observability recommendations

During the validation window and cutover, wowsociety should monitor:

- Rate of `identity_grant` lookups (framework metric or query log).
- Impersonation start/stop success rate.
- `whoami` error rate, especially 401/403 spikes.
- Distribution of `GrantRejection` reasons if logged.

A temporary dashboard or structured log query should be prepared before the cutover.

## Security considerations

- Staging data may contain synthetic PII; handle it under wowsociety's data-handling policy.
- Do not validate against production impersonation data without explicit access controls and
  approval.
- Grant-authority migration is security-critical; require a coordinated go/no-go review with at
  least one wowsociety security or identity owner.

## Out of scope

- Fixing data gaps found during validation (tracked as remediation work, not validation).
- Production cutover execution (see `sequencing-plan.md`).
- Long-term observability (recommendations only; implementation is wowsociety's).

## Review status

- wowapi-side reviewer: self-review as part of W03-E01-S004 acceptance.
- wowsociety-side reviewer: to be assigned by wowsociety engineering owner.

## References

- `docs/implementation/premier-framework-implementation-plan.md` §5.2 SEC-01 wowsociety-impact prose.
- `impl/waves/wave-03-identity-and-session-security/epics/epic-001-server-side-session-state/stories/story-001-grant-schema-and-membership/`
- `impl/waves/wave-03-identity-and-session-security/epics/epic-001-server-side-session-state/stories/story-002-capacity-and-privileged-resolver/`
- `docs/tracking/decision-register.md` D-01.
- `impl/waves/wave-02-data-safety-and-migration-tooling/epics/epic-001-online-migration-protocol/` (DATA-09 online migration discipline).

## WIP / unresolved

- Exact wowsociety staging DB connection and anonymization procedure.
- Whether wowsociety has feature-flag infrastructure for phased cutover.
- Final decision on DEC-Q1 claim shape (safe default assumed).

These items must be determined before the plan is executed but are intentionally not invented here
per mandate §18.

## Evidence

Review record: EV-W03-E01-S004-002.

## Rollback trigger

If staging validation fails, follow `rollback-plan.md` failure direction (b): wowsociety-side
`grant_id` adoption is broken.

## Test suites to re-run

- `internal/modules/identity/abac_test.go`
- `internal/modules/identity/whoami_impersonation_test.go`
- `internal/.../rls_test.go` (exact path TBD by wowsociety owner)

## Acceptance

This staging-validation plan satisfies AC-W03-E01-S004-02.

## Warnings

- A green staging validation does not guarantee production safety if production data differs
  materially from staging. The production `user_tenant_access` audit (S001 deferred work) remains a
  pre-condition for production cutover.
- Do not skip V3 adversarial negatives because V2 matches 100% — the resolver's rejections are the
  security control, not the happy path.

## Review checklist

- [ ] Names concrete wowsociety test suites.
- [ ] States explicit go/no-go criteria.
- [ ] Covers both membership (T2) and resolver (T5) validation.
- [ ] Includes access-control note for production-derived data.
- [ ] Recommends observability for the cutover window.
- [ ] Reviewed by at least a wowapi-side reviewer.

## Sign-off

| Role | Name | Date | Signature/Approval |
|---|---|---|---|
| wowapi reviewer | self-review | 2026-07-13 | — |
| wowsociety reviewer | TBD | — | — |

## Notes

This plan is a living document. If DEC-Q1 resolves differently than the safe default, revise V2
and V3 accordingly before execution.

## See also

- `sequencing-plan.md`
- `rollback-plan.md`
- `story.md`
- `plan.md`

## Document owner

W03-E01-S004 task owner (unassigned at planning time; executed by this session).

## Version

1.0 — 2026-07-13.

## Status

Draft → reviewed → accepted as part of W03-E01-S004 closure.

## EOF
