---
id: PLAN-W04-E03-S001
type: plan
parent_story: W04-E03-S001
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Plan — W04-E03-S001

Per mandate §8.5. Confirmed facts, planned changes, and implementation assumptions are distinguished
explicitly below; this plan does not invent precise code changes where the repository does not yet
provide enough information beyond what the source itself cites (migration `00016`'s header and
`kernel/bulk/bulk.go:123-144`'s `Service.next`).

## Proposed architecture

No new architectural component. A targeted correction to an existing migration file's header
comment, plus a single-processor enforcement check added at the `Service` API boundary in
`kernel/bulk`, ahead of (or wrapping) the existing `Service.next` claim path. The enforcement
mechanism is additive to the existing unlocked `SELECT ... LIMIT 1` — this story does not replace
that query with the atomic `SKIP LOCKED` claim (that is `W04-E03-S002-T003`'s scope); it adds a gate
in front of it that rejects a second concurrent caller.

## Implementation strategy

1. Re-read migration `00016`'s header comment and `kernel/bulk/bulk.go:123-144`'s `Service.next`
   fresh at this story's actual start commit, confirming the false claim and the absence of
   enforcement still hold (resolving `story.md`'s current-state re-confirmation step).
2. Decide the enforcement mechanism: PostgreSQL advisory lock keyed on `bulkID` (`pg_advisory_lock`/
   `pg_try_advisory_lock` family, session- or transaction-scoped), or a CAS check against a
   processing-owner column added to the `bulk_operations` (or equivalent) table. Document the
   trade-off and the chosen mechanism's rationale.
3. Correct migration `00016`'s header comment to remove the false "safe across replicas" claim,
   replacing it with an accurate statement of the current, stopgap-enforced property.
4. Implement the chosen enforcement mechanism at the `Service` API boundary, such that a second
   caller attempting to process the same `bulkID` while a first is active is rejected with a clear,
   distinguishable error (not a generic failure indistinguishable from an unrelated error).
5. Write a concurrency test: 2 processors attempt to claim/process the same `bulkID`
   simultaneously; confirm exactly one succeeds and the second is rejected, not silently racing.
6. Add observability (at minimum, a log line) for a rejected second-processor attempt.
7. Document the corrected claim and the enforcement mechanism.

## Expected package or module changes

`kernel/bulk` — the `Service` type gains the enforcement check at its API boundary (exact method(s)
affected TBD, expected to be `Service.next` or its caller, per the source's own line citation). If
the CAS mechanism is chosen, a new column may be added to the relevant table (exact table TBD,
expected `bulk_operations` or equivalent given `bulkID`-scoped locking).

## Expected file changes where determinable

- Migration `00016`'s header comment (exact file path to be confirmed at step 1 above; expected
  under the repository's migrations directory).
- `kernel/bulk/bulk.go` (the file explicitly cited by the source for `Service.next`, lines 123-144)
  — enforcement check added at or near this function.
- A new concurrency test file under `kernel/bulk` (exact path TBD), or under `DATA-04/stopgap/` per
  the source's own evidence-path convention for this task.
- If the CAS mechanism is chosen: a new small, additive migration for the processing-owner column
  (exact migration number TBD, to follow the repository's existing migration-numbering convention).

## Contracts and interfaces

`Service`'s public claim/processing-entry API gains a rejection error path (an error type or sentinel
distinguishing "rejected: another processor is active" from other failure modes) — exact shape TBD
at implementation time.

## Data structures

If the CAS mechanism is chosen: a processing-owner column (type TBD — likely a nullable identifier
or a boolean/timestamp claim marker) on the relevant table. If the advisory-lock mechanism is chosen:
no new data structure, only a runtime lock acquisition keyed on `bulkID`.

## APIs

`Service`'s claim/processing-entry method gains a new rejection outcome (distinguishable error) when
a second concurrent caller is detected. No other API surface change.

## Configuration changes

None anticipated.

## Persistence changes

None if the advisory-lock mechanism is chosen (advisory locks are session/transaction-scoped
PostgreSQL primitives, not a schema change). If the CAS mechanism is chosen: an additive column on
the relevant table, via a small, separate migration.

## Migration strategy

The only migration-adjacent action in this story is (a) correcting migration `00016`'s header
comment, which is a comment-only edit, not a schema change, and (b), conditionally, adding a new
CAS-supporting column via a small additive migration if that mechanism is chosen over the advisory
lock. Neither is a data migration in the DATA-09 online-migration-protocol sense — both are small,
low-risk, additive changes.

## Concurrency implications

This story's entire purpose is a concurrency-correctness fix: rejecting a second concurrent
processor rather than allowing it to race against the first's unlocked `SELECT ... LIMIT 1`. The
enforcement mechanism itself must be safe under the exact concurrency scenario it targets — the
concurrency test (step 5) is this story's primary evidence, not a secondary check.

## Error-handling strategy

A rejected second-processor attempt must fail with a clear, distinguishable error — not a generic
database error or an ambiguous failure that could be mistaken for an unrelated problem. The first
(accepted) processor's own behavior must be unaffected by the presence of the new enforcement check.

## Security controls

None beyond the concurrency-correctness control itself (see "Concurrency implications").

## Observability changes

A rejected second-processor attempt is logged with the `bulkID` and the rejecting mechanism
(implementation-time addition, not separately mandated by the source beyond the rejection behavior
itself).

## Testing strategy

- The named concurrency test: 2 processors on the same `bulkID`, confirming exactly one succeeds and
  the second is cleanly rejected — this is the source's own required test ("Concurrency test: 2
  processors on the same `bulkID`").
- No additional integration or race-detector test is separately required by the source beyond this
  concurrency test itself, though running it under Go's race detector is a reasonable
  implementation-time addition given the concurrency-correctness nature of the fix.

## Regression strategy

The concurrency test itself becomes the regression guard — any future change to `Service.next` or
its enforcement wrapper that reintroduces the race would fail this test.

## Compatibility strategy

A caller already respecting the (previously documentation-only) "single processor per operation"
contract observes no behavioral change. Only a second, genuinely concurrent caller against the same
`bulkID` is newly rejected. No compatibility flag or phased rollout is required — this is a pure
correctness fix with no legitimate caller depending on the old racy behavior.

## Rollout strategy

Single story, landed as its own reviewable unit — no phased rollout. This story is explicitly
designed to ship independently and fast, ahead of `W04-E03-S002`'s full rewrite.

## Rollback strategy

Revert the enforcement mechanism if it produces false-positive rejections against a legitimate
single-processor caller under normal (non-concurrent) operation; the migration comment correction is
low-risk and does not require a rollback plan beyond a standard doc revert.

## Implementation sequence

As listed under "Implementation strategy" above (steps 1–7). Step 2 (mechanism decision) must occur
before step 4 (implementation) locks in the chosen approach.

## Task breakdown

- **W04-E03-S001-T001** — Correct false migration comment; implement single-processor enforcement
  (advisory lock or CAS) at the `Service` API boundary; concurrency test (steps 1–7 above).

No separate independent-review task is added for this story — see `tasks/index.md`'s "Grouping
rationale" for the explicit reasoning.

## Expected artifacts

The corrected migration `00016` header comment; the single-processor enforcement mechanism (code);
documentation of the corrected claim and the mechanism.

## Expected evidence

2-processor concurrency test output (`DATA-04/stopgap/`).

## Unresolved questions

- Exact enforcement mechanism: PostgreSQL advisory lock vs. CAS against a processing-owner column —
  to be decided at implementation time, documented with rationale.
- Exact migration file path for `00016` and exact line range within `kernel/bulk/bulk.go` beyond the
  source's own citation (123-144) — to be confirmed by direct inspection at this story's start
  commit.
- If CAS is chosen: exact table and column name for the processing-owner marker.

## Approval conditions

This plan is approved for implementation once: (a) the mechanism-choice question above is answered
and documented, and (b) the owner and reviewer are assigned.
