---
id: W04-E03-ACCEPTANCE
type: epic-acceptance
epic: W04-E03
wave: W04
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W04-E03 — Epic-level acceptance

Reproduces and elaborates `epic.md`'s "Epic acceptance criteria" section as a standalone,
independently-referenceable acceptance record, consistent with the wave-level `../../acceptance.md`
pattern (the DATA-04 line item there maps onto this epic).

## AC-W04-E03-01 — False documentation claim corrected; second processor rejected

Migration `00016`'s header comment claiming the bulk processing path is "safe across replicas" via
`FOR UPDATE SKIP LOCKED` is removed or corrected to reflect the code's actual behavior at the point
of the fix. A second concurrent processor attempting to claim/process the same `bulkID` is
mechanically rejected (via advisory lock or CAS at the `Service` API boundary), not silently
racing, proven by a 2-processor concurrency test. Traces to W04-E03-S001.

## AC-W04-E03-02 — Lease columns and atomic leased claim proven

`bulk_items` gains lease columns via the shared primitive built in `W04-E01-S001`, proven by a
migration test. The atomic leased-claim SQL statement
(`UPDATE ... FROM (SELECT ... FOR UPDATE SKIP LOCKED LIMIT $batch) RETURNING ...`) provably uses
`SKIP LOCKED` via an `EXPLAIN`-plan assertion, is bounded to a configured batch size, and is
exercised by a concurrent `N>1` claimer test proving no two claimers receive the same row.
`runItem`'s existing idempotent completion CAS guard is confirmed unchanged. Traces to
W04-E03-S002 (T2, T3).

## AC-W04-E03-03 — Fencing, idempotency, retry, cancellation, and lifecycle controls proven

A fenced (stale) worker's finalize write to a `bulk_items` row is rejected, proven by reusing
DATA-02's chaos pattern (via the shared harness in `W04-E01-S003`) rather than a bespoke test.
Item idempotency keys, retry policy, and cancellation behave correctly under test. Pause/resume/
cancel operation-level controls, exercised against bounded batch claims, behave correctly mid-run —
proven by lifecycle integration tests. Traces to W04-E03-S002 (T4, T5).

## AC-W04-E03-04 — Named multi-worker chaos test passes

The named chaos test `DATA-04/chaos/duplicate_worker_test.go` passes: ≥2 processors concurrently
claim/retry/pause/resume/cancel the same operation without duplicate effects or stale finalization,
matching the Wave-3 exit gate wording verbatim. The test is built by reusing the shared chaos
harness from `W04-E01-S003` (DATA-02 T7) — the task record and its evidence must show the harness
was consumed, not reimplemented. Traces to W04-E03-S002 (T6).

## AC-W04-E03-05 — Independent review passed

Both stories (S001, S002) have passed independent review per mandate §14. S001's review confirms
the false migration-comment claim was genuinely corrected (not merely annotated) and that the
stopgap mechanism genuinely rejects a second concurrent processor under test, not merely in
documentation. S002's review specifically confirms genuine reuse (not reimplementation) of
`W04-E01-S002`'s finalize-fencing logic (T4) and `W04-E01-S003`'s shared chaos harness (T6), and
confirms `runItem`'s pre-existing completion CAS guard was not silently weakened by T3's rewrite
(RISK-W04-E03-002).

## Acceptance authority

Data/reliability lead, per `../../wave.md`'s wave-level acceptance authority (PLAN §5.3's
accountable role for PF-DATA).
