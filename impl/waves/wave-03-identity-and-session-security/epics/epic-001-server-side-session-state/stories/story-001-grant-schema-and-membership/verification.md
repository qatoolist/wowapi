---
id: VER-W03-E01-S001
type: verification-record
parent_story: W03-E01-S001
status: complete
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Verification record — W03-E01-S001

## Planned verification procedure

Per mandate §8.8. One row per acceptance criterion for this story.

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W03-E01-S001-01 | Run the `identity_grant` migration up/down test; run an RLS-catalog-extension test confirming FORCE + unique partial index + `app_platform`-only writes | Local dev or CI, PostgreSQL matching production major version, testkit DB | Migration applies/reverts cleanly; RLS catalog reflects FORCE and the correct role grants; index enforces one active grant per actor under a concurrency test | migration test report + RLS catalog report | unassigned |
| AC-W03-E01-S001-02 | Run the adversarial membership test suite (revoked/absent/foreign-tenant membership with a validly signed token) against `Verifier.Actor` | Local dev or CI, testkit DB seeded with fixture `user_tenant_access` rows | All three adversarial cases rejected; a capacity-less actor is now membership-checked (previously bypassed) | adversarial test report | unassigned |
| AC-W03-E01-S001-03 | Run the zero/garbage-UUID tenant negative test against the pre-`WithTenantID` rejection path | Local dev or CI | Rejected before any tenant transaction opens | negative test report | unassigned |

## Post-execution record

### Actual result

All targeted tests pass:

```
$ WOWAPI_TEST_DSN="postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable" \
  go test ./migrations/... ./adapters/auth/pgprincipal/... ./kernel/auth/... -count=1
ok  	github.com/qatoolist/wowapi/migrations
ok  	github.com/qatoolist/wowapi/adapters/auth/pgprincipal
ok  	github.com/qatoolist/wowapi/kernel/auth
```

Fail-first evidence: the new adversarial/zero-tenant tests were confirmed to fail against the
pre-fix `Verifier.Actor` behavior (capacity-less actors bypassed membership, zero-tenant claims
reached actor construction) and pass after the fix.

### Pass or fail

Pass.

### Evidence identifier

EV-W03-E01-S001-001 through EV-W03-E01-S001-005 (see `evidence/index.md`).

### Execution date

2026-07-13.

### Commit or revision

Re-verified at HEAD `733ef3e` with uncommitted W03-E01-S001 changes.

### Environment

Local Docker Postgres 16 (`wowapi-postgres-1`);
`WOWAPI_TEST_DSN=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable`;
Go 1.26.5.

### Reviewer

Independent review completed per EV-W03-E01-S001-006.

### Findings

None. The `user_tenant_access` data audit against the local database returned zero gaps.

### Retest status

Initial pass; no retest required unless review findings demand it.

### Final conclusion

AC-W03-E01-S001-01, AC-W03-E01-S001-02, and AC-W03-E01-S001-03 are satisfied by passing tests
with recorded evidence. Independent review (T004) completed; story accepted.
