---
id: W04-E03-S001-T001
type: task
title: Correct false migration comment; enforce single-processor via advisory lock/CAS; concurrency test
status: done
parent_story: W04-E03-S001
owner: W04BulkSafety
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on: []
acceptance_criteria:
  - AC-W04-E03-S001-01
  - AC-W04-E03-S001-02
artifacts:
  - ART-W04-E03-S001-001
  - ART-W04-E03-S001-002
  - ART-W04-E03-S001-003
evidence:
  - EV-W04-E03-S001-001
  - EV-W04-E03-S001-002
---

# W04-E03-S001-T001 — Correct false migration comment; enforce single-processor via advisory lock/CAS; concurrency test

## Task Definition

### Task objective

Correct migration `00016`'s header comment to remove the false "safe across replicas" claim, and
implement single-processor enforcement (advisory lock or CAS) at the `Service` API boundary in
`kernel/bulk` so a second concurrent processor attempting to process the same `bulkID` is rejected
rather than silently racing against the first, proven by a 2-processor concurrency test.

### Parent story

W04-E03-S001 — Bulk multi-worker stopgap — correct false safety claim, enforce single-processor.

### Owner

W04BulkSafety.

### Status

done.

### Dependencies

None.

### Detailed work

1. Re-read migration `00016`'s header comment and `kernel/bulk/bulk.go:123-144`'s `Service.next` at
   this task's actual start commit, confirming the false "safe across replicas" claim and the
   absence of any mechanical single-processor enforcement still hold (resolving `story.md`'s
   current-state re-confirmation step).
2. Decide the enforcement mechanism — PostgreSQL advisory lock keyed on `bulkID`, or a CAS check
   against a processing-owner column — and document the rationale for the choice (resolves
   `plan.md`'s "Unresolved questions" item on mechanism selection).
3. Correct migration `00016`'s header comment: remove the false cross-replica-safety claim; state
   the actual, now-enforced single-processor property.
4. Implement the chosen enforcement mechanism at the `Service` API boundary, with a clear,
   distinguishable rejection error when a second concurrent caller targets an already-active
   `bulkID`.
5. Write the named concurrency test: 2 processors attempt to claim/process the same `bulkID`
   simultaneously; confirm exactly one succeeds and the second is cleanly rejected, not silently
   racing (`DATA-04/stopgap/`, per the source's own evidence-path convention).
6. Add observability (at minimum, a log line) for a rejected second-processor attempt, including the
   `bulkID` and the rejecting mechanism.
7. Document the corrected claim and the chosen enforcement mechanism.

### Expected files or components affected

Migration `00016`'s header comment; `kernel/bulk/bulk.go` and the `Service` API boundary generally;
a new concurrency test in `kernel/bulk/bulk_test.go`; a new additive migration for the CAS
processing-owner columns.

### Expected output

A corrected migration `00016` header comment; a working single-processor enforcement mechanism at
the `Service` API boundary; a passing 2-processor concurrency test proving the second processor is
rejected, not racing.

### Required artifacts

ART-W04-E03-S001-001 (corrected migration comment), ART-W04-E03-S001-002 (enforcement mechanism),
ART-W04-E03-S001-003 (documentation).

### Required evidence

EV-W04-E03-S001-001 (documentation-diff record), EV-W04-E03-S001-002 (concurrency-test report).

### Related acceptance criteria

AC-W04-E03-S001-01, AC-W04-E03-S001-02.

### Completion criteria

Migration `00016`'s header comment no longer claims cross-replica safety and instead states the
actual enforced property; a second processor attempting to process the same `bulkID` as an active
first processor is rejected, not racing — proven by the named 2-processor concurrency test passing.

### Verification method

Direct inspection of the corrected migration comment; direct execution of the 2-processor
concurrency test against a live PostgreSQL instance, logged output retained as evidence.

### Risks

RISK-W04-E03-001 (epic-level `risks.md`) — this task's stopgap mechanism must be cleanly
supersedable by `W04-E03-S002`'s T2 lease-column mechanism; this task's own scope should avoid
building anything that would complicate that clean handoff (e.g. avoid coupling the enforcement
mechanism so tightly into `Service.next`'s internals that S002's rewrite cannot cleanly remove it).

### Rollback or recovery considerations

Revert the enforcement mechanism if it produces false-positive rejections against a legitimate
single-processor caller under normal, non-concurrent operation; the migration comment correction is
low-risk and reversible via a standard documentation revert if needed.

## Implementation Record

### What was actually implemented

- Migration `00016` header comment corrected to remove the false "FOR UPDATE SKIP LOCKED — safe
  across replicas" claim and state the actual single-processor property.
- Added migration `00041_bulk_operation_processor_lock.sql` with additive CAS columns
  `processor_id` and `processor_started_at` on `bulk_operations`.
- Added `ErrConcurrentProcessor` sentinel (`KindConflict`) in `kernel/bulk/bulk.go`.
- Added `Service.acquireProcessor` / `Service.releaseProcessor` helpers implementing a CAS guard
  with a 5-minute timeout.
- Wrapped `Service.Process` with the CAS acquisition and a deferred release; rejected attempts are
  logged via the service logger.
- Added `WithLogger` option to `Service`; wired the kernel logger in `kernel/kernel.go`.
- Added `TestIntegrationBulkConcurrentProcessorRejected` in `kernel/bulk/bulk_test.go`.

### Components changed

`kernel/bulk`, `migrations`, `kernel` (logger wiring).

### Files changed

- `migrations/00016_bulk_operations.sql`
- `migrations/00041_bulk_operation_processor_lock.sql`
- `kernel/bulk/bulk.go`
- `kernel/bulk/bulk_test.go`
- `kernel/kernel.go`

### Interfaces introduced or changed

- New exported error: `bulk.ErrConcurrentProcessor`.
- New method: `(*Service).WithLogger(*slog.Logger) *Service`.
- `Service.Process` now rejects concurrent callers against the same `bulkID`.

### Configuration changes

None.

### Schema or migration changes

Migration `00041` adds `processor_id uuid` and `processor_started_at timestamptz` to
`bulk_operations`, with `UPDATE` grant to `app_rt`.

### Security changes

None beyond the concurrency-correctness control itself.

### Observability changes

Rejected concurrent processors are logged at INFO with `bulk_id`. Lock-release failures are logged
at ERROR.

### Tests added or modified

- `TestIntegrationBulkConcurrentProcessorRejected` (new) — 2 processors, same `bulkID`, second
  rejected with `KindConflict`, first completes all 3 items.
- Existing bulk integration tests re-run and pass.

### Commits

Working tree changes; no git commits per session constraints.

### Implementation dates

2026-07-13.

### Technical debt introduced

Deliberate, tracked supersession by `W04-E03-S002`'s T2 lease-column mechanism (RISK-W04-E03-001).
The stopgap CAS columns will be removed/superseded when the leased-claim rewrite lands.

### Known limitations

The CAS timeout is hardcoded at 5 minutes. This is acceptable for a stopgap; the leased-claim
rewrite will remove this limitation entirely.

### Follow-up items

- `W04-E03-S002` must explicitly supersede this stopgap (remove `processor_id` / `processor_started_at`
  reliance and the CAS guard) when the lease-column mechanism lands.

### Relationship to the approved plan

Matches plan. Mechanism chosen: CAS (not advisory lock), documented in artifacts index because
`Process` spans multiple `TxManager` transactions and a session-scoped advisory lock cannot be held
across them without a dedicated connection abstraction.

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W04-E03-S001-01 | Inspect migration `00016`'s corrected header comment | Documentation / source inspection | False "safe across replicas" claim removed; actual single-processor-enforced property stated | documentation-diff record | W04BulkSafety |
| AC-W04-E03-S001-02 | Run the 2-processor concurrency test against the same `bulkID` | Local dev or CI, PostgreSQL instance | Exactly one processor succeeds; the second is cleanly rejected, not racing | concurrency-test report | W04BulkSafety |

### Actual result

- AC-W04-E03-S001-01: migration `00016` header corrected in place.
- AC-W04-E03-S001-02: `TestIntegrationBulkConcurrentProcessorRejected` passes.

### Pass or fail

Pass.

### Evidence identifier

EV-W04-E03-S001-001, EV-W04-E03-S001-002.

### Execution date

2026-07-13.

### Commit or revision

HEAD (working tree).

### Environment

Local PostgreSQL via `make up`; `DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable`.

### Reviewer

W04BulkSafety (review folded into T001 per tasks/index.md grouping rationale).

### Findings

None.

### Retest status

N/A.

### Final conclusion

Accepted.

## Deviations Record

No deviations.
