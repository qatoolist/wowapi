---
id: W04-E03-S002-T001
type: task
title: Lease columns via the shared primitive
status: done
parent_story: W04-E03-S002
owner: W04BulkSafety
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on:
  - W04-E01-S001
  - W04-E03-S001
acceptance_criteria:
  - AC-W04-E03-S002-01
artifacts:
  - ART-W04-E03-S002-001
evidence:
  - EV-W04-E03-S002-001
---

# W04-E03-S002-T001 — Lease columns via the shared primitive

## Task Definition

### Task objective

Add lease columns to `bulk_items` by reusing DATA-02's shared lease/fencing primitive built in
`W04-E01-S001` — not a bespoke copy — matching the same schema pattern the primitive already applies
to `jobs_queue`.

### Parent story

W04-E03-S002 — Leased claims, finalize fencing, lifecycle controls, and the named multi-worker
chaos test.

### Owner

W04BulkSafety.

### Status

done.

### Dependencies

W04-E01-S001 (the shared lease/fencing primitive must exist and its schema pattern must be readable
before this task can integrate against it); W04-E03-S001 (this epic's stopgap, whose enforcement
this task's downstream sibling T002 will supersede).

### Detailed work

1. Read `W04-E01-S001`'s landed shared lease/fencing primitive — its migration pattern for
   `jobs_queue` (`lease_token`, monotonic `lease_generation`, `lease_expires_at`) — and confirm the
   exact column set and types to reuse for `bulk_items`.
2. Write an additive migration adding the same lease-column set to `bulk_items`, matching the
   primitive's schema contract exactly.
3. Write a migration test confirming the columns exist on `bulk_items` and match the primitive's
   schema contract.
4. Document the lease-column schema and its explicit relationship to the shared primitive.

### Expected files or components affected

Migration `00044_bulk_items_lease_and_lifecycle.sql`; `kernel/bulk` types and tests.

### Expected output

`bulk_items` carries lease columns matching the shared primitive's schema contract, proven by a
migration test.

### Required artifacts

ART-W04-E03-S002-001 (the lease-column migration).

### Required evidence

EV-W04-E03-S002-001 (migration-test report).

### Related acceptance criteria

AC-W04-E03-S002-01.

### Completion criteria

The migration test confirms `bulk_items` has the lease columns and they match the shared primitive's
schema contract exactly.

### Verification method

Direct execution of the migration test against a live PostgreSQL instance; comparison against
`W04-E01-S001`'s own `jobs_queue` lease columns to confirm contract parity.

### Risks

None.

### Rollback or recovery considerations

Additive-only migration; goose Down reverses the column additions.

## Implementation Record

### What was actually implemented

- Migration `00044_bulk_items_lease_and_lifecycle.sql` adds to `bulk_items`:
  - `lease_token text`
  - `lease_generation bigint NOT NULL DEFAULT 0`
  - `lease_expires_at timestamptz`
  - `idempotency_key uuid`
- Migration also adds `max_attempts int NOT NULL DEFAULT 3` to `bulk_operations`, extends status
  check constraints for pause/resume/cancel, and drops the superseded stopgap columns from
  migration `00041`.
- `TestIntegrationBulkLeaseColumnsExist` proves the columns are populated after a claim.

### Components changed

`migrations`, `kernel/bulk`.

### Files changed

- `migrations/00044_bulk_items_lease_and_lifecycle.sql`
- `kernel/bulk/bulk_test.go`

### Interfaces introduced or changed

`bulk.Item` now carries `Lease` and `IdempotencyKey`.

### Configuration changes

None.

### Schema or migration changes

Migration `00044` (additive + supersession of stopgap columns).

### Security changes

None.

### Observability changes

None.

### Tests added or modified

`TestIntegrationBulkLeaseColumnsExist`.

### Commits

Working tree changes.

### Implementation dates

2026-07-13.

### Technical debt introduced

None.

### Known limitations

None.

### Follow-up items

None.

### Relationship to the approved plan

Matches plan.

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W04-E03-S002-01 | Run the migration test against a live PostgreSQL instance | Local dev or CI, PostgreSQL instance | Lease columns exist on `bulk_items`, matching the shared primitive's schema contract | migration-test report | W04BulkSafety |

### Actual result

`TestIntegrationBulkLeaseColumnsExist` passes; lease columns present and populated.

### Pass or fail

Pass.

### Evidence identifier

EV-W04-E03-S002-001.

### Execution date

2026-07-13.

### Commit or revision

HEAD (working tree).

### Environment

Local PostgreSQL via `make up`; `DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable`.

### Reviewer

W04BulkSafety (review folded into task completion per tasks/index.md).

### Findings

None.

### Retest status

N/A.

### Final conclusion

Accepted.

## Deviations Record

None.
