---
id: PLAN-W04-E03-S002
type: plan
parent_story: W04-E03-S002
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Plan — W04-E03-S002

Per mandate §8.5. Confirmed facts, planned changes, and implementation assumptions are distinguished
explicitly below. This plan does not invent precise code changes where the repository does not yet
provide enough information — in particular, it does not invent the exact API shape of
`W04-E01-S001`'s shared lease/fencing primitive, `W04-E01-S002`'s finalize-fencing logic, or
`W04-E01-S003`'s chaos harness, since none of those is visible from this epic's own planning scope;
this plan records what must be determined by reading those stories' actual landed code, not by
guessing their shape in advance.

## Proposed architecture

`bulk_items` is extended with lease columns matching the shared primitive's own schema pattern (as
already applied to `jobs_queue` by `W04-E01-S001`). `Service.next`'s claim path — already gated by
`W04-E03-S001`'s stopgap enforcement — is replaced with a single atomic SQL statement
(`UPDATE ... FROM (SELECT ... FOR UPDATE SKIP LOCKED LIMIT $batch) RETURNING ...`) that both selects
and leases a bounded batch of claimable items in one round trip, removing the plain unlocked
`SELECT ... LIMIT 1` and the stopgap's advisory-lock/CAS wrapper alike (superseding it, per
RISK-W04-E03-001). `runItem`'s finalize path gains a fencing check reusing `W04-E01-S002`'s
finalize-fencing logic, rejecting a stale (fenced) worker's write while preserving the existing
completion CAS guard already in place. A new pause/resume/cancel control surface at the
operation level governs whether bounded batch claims are issued and how in-flight items behave
across a lifecycle transition.

## Implementation strategy

1. Re-read `kernel/bulk`'s current claim, finalize, and lifecycle-control code at this story's
   actual start commit (post-`W04-E03-S001` landing), and read `W04-E01-S001`'s landed shared
   lease/fencing primitive API, `W04-E01-S002`'s landed finalize-fencing logic, and
   `W04-E01-S003`'s landed shared chaos harness — confirming their actual shapes before designing
   this story's integration against them (resolving `story.md`'s current-state re-confirmation
   step).
2. **T2** — Add lease columns to `bulk_items` via the shared primitive's own migration pattern;
   write a migration test confirming the columns exist and match the primitive's schema contract.
3. **T3** — Implement the atomic leased-claim SQL statement
   (`UPDATE ... FROM (SELECT ... FOR UPDATE SKIP LOCKED LIMIT $batch) RETURNING ...`), bounded to a
   configured default batch size (exact value TBD, documented with rationale); remove the plain
   unlocked `SELECT ... LIMIT 1` and `W04-E03-S001`'s stopgap wrapper, explicitly recording the
   supersession (mitigating RISK-W04-E03-001); confirm `runItem`'s existing idempotent completion
   CAS guard is unchanged by inspection and by a targeted test; write an `EXPLAIN`-plan assertion
   proving the statement uses `SKIP LOCKED`; write a concurrent `N>1` claimer test proving no two
   claimers receive the same row.
4. **T4** — Add item idempotency keys (exact scheme TBD, documented with rationale); integrate
   `W04-E01-S002`'s finalize-fencing logic into `runItem`'s finalize path, rejecting a stale
   (fenced) worker's write — explicitly reusing that story's logic, not designing an independent
   scheme; implement a retry policy and a cancellation path; prove fencing rejection by reusing
   DATA-02's chaos pattern (via the shared harness).
5. **T5** — Implement operation-level pause/resume/cancel controls (exact API shape TBD, documented
   with rationale), exercised against bounded batch claims so an in-flight claim correctly respects
   a pause/cancel transition; write lifecycle integration tests covering pause-then-resume,
   pause-then-cancel, and cancel-mid-batch scenarios.
6. **T6** — Build the named chaos test `DATA-04/chaos/duplicate_worker_test.go`, reusing (not
   reimplementing) the shared chaos harness built in `W04-E01-S003` (cross-referenced by name, per
   this epic's `dependencies.md`); confirm ≥2 processors concurrently claim/retry/pause/resume/
   cancel the same operation without duplicate effects or stale finalization.
7. Document the lease schema, claim SQL, idempotency scheme, fencing behavior (with explicit
   cross-reference to `W04-E01-S002`), retry policy, cancellation path, and lifecycle-control API.

## Expected package or module changes

`kernel/bulk` — `Service`'s claim path (T3), finalize path (T4), and a new lifecycle-control surface
(T5). `bulk_items` — new lease columns (T2) and any additional idempotency-key/lifecycle-state
columns T4/T5 require (exact schema TBD).

## Expected file changes where determinable

- `kernel/bulk/bulk.go` — `Service.next`'s claim path replaced (T3); `runItem`'s finalize path
  extended with fencing (T4).
- A new migration for `bulk_items`' lease columns (T2), and conditionally further migrations for
  idempotency-key/lifecycle-state columns (T4/T5) — exact migration numbers TBD.
- A new lifecycle-control entry point/method on `Service` (T5) — exact file/method TBD.
- A new named chaos test file `DATA-04/chaos/duplicate_worker_test.go` (T6, exact path matching the
  source's own required path verbatim).

## Contracts and interfaces

`Service`'s claim method's return contract changes from "a single item or none" (via `SELECT ...
LIMIT 1`) to "a bounded batch of leased items or none" (via the new atomic claim statement) — this is
a compatibility-relevant interface change, tracked in "Compatibility considerations" below. A new
fencing-check integration point in the finalize path, consuming `W04-E01-S002`'s fencing contract. A
new pause/resume/cancel control API on `Service` or an operation-level control object (exact shape
TBD).

## Data structures

Lease columns on `bulk_items` (exact fields determined by the shared primitive's own schema —
expected `lease_token`, `lease_generation`, `lease_expires_at`, per `wave.md`'s own description of
the primitive's shape: "`lease_token`, monotonic `lease_generation`, `lease_expires_at`, optional
heartbeat"). An idempotency-key column (T4, exact type TBD). A lifecycle-state representation for
pause/resume/cancel (T5, exact shape TBD — a status column, a separate control table, or both).

## APIs

`Service`'s claim method's contract changes (see "Contracts and interfaces"). A new lifecycle-control
API (pause/resume/cancel) is added, operation-scoped.

## Configuration changes

The bounded-batch size for T3's leased claim (`LIMIT $batch`) — exact default and whether it is a
hardcoded constant or a configuration key is an implementation-time decision, to be documented with
rationale.

## Persistence changes

Additive lease columns on `bulk_items` (T2), via the shared primitive's own migration pattern.
Conditionally, additive idempotency-key and lifecycle-state columns (T4/T5) — exact schema TBD.

## Migration strategy

All schema changes in this story are additive (new nullable/default-safe columns), consistent with
this epic's own out-of-scope note that the DATA-09 online-migration protocol is not itself required
for this epic's own migrations unless a specific migration's risk profile warrants it — to be
confirmed at implementation time against the manifest schema built in `W02-E01-S001`, if that
schema's CI gate is already enforced at this story's landing time.

## Concurrency implications

This story is fundamentally about concurrency correctness: T3's atomic `SKIP LOCKED` claim must
provably prevent two concurrent claimers from receiving the same row; T4's fencing must provably
reject a stale worker's finalize write under concurrent worker failover; T5's lifecycle controls must
behave correctly when a pause/cancel races against an in-flight claim or finalize; T6's chaos test is
this story's ultimate concurrency-correctness proof, exercising all of the above together under
adversarial interleavings via the shared harness.

## Error-handling strategy

A claim attempt that finds no claimable batch must return cleanly (no error, empty result) — not
conflated with a fencing rejection or a lifecycle-control rejection, which must each surface a
distinguishable error/outcome. A fenced worker's finalize write must fail with an error identifying
it as a fencing rejection, not a generic write failure.

## Security controls

Finalize fencing (T4) is the primary security-adjacent control in this story — see `story.md`
"Security considerations." It is implemented by reuse of `W04-E01-S002`'s already-reviewed logic,
not by an independently-designed (and independently-riskier) new mechanism.

## Observability changes

Lease acquisition, claim batch size, fencing rejections, and pause/resume/cancel transitions are
each logged at minimum — see `story.md` "Observability considerations."

## Testing strategy

- T2: migration test confirming lease columns exist and match the shared primitive's schema
  contract.
- T3: `EXPLAIN`-plan assertion proving `SKIP LOCKED` is used; concurrent `N>1` claimer test proving
  no two claimers receive the same row; a targeted test confirming `runItem`'s existing completion
  CAS guard is unchanged.
- T4: fenced-finalize-rejection test reusing DATA-02's chaos pattern; idempotency-key, retry-policy,
  and cancellation tests.
- T5: lifecycle integration tests (pause-then-resume, pause-then-cancel, cancel-mid-batch).
- T6: the named chaos test `DATA-04/chaos/duplicate_worker_test.go`, reusing the shared harness from
  `W04-E01-S003`, exercising ≥2 processors concurrently claiming/retrying/pausing/resuming/
  cancelling the same operation.

## Regression strategy

T3's `EXPLAIN`-plan assertion and concurrent claimer test become the regression guard against a
future change silently reverting to an unlocked claim path. T6's named chaos test is the ultimate
regression guard for the entire rebuilt path, run as part of this wave's own quality gates (per
`wave.md` "Quality gates": "DATA-04's fail-first evidence is the named chaos test
`DATA-04/chaos/duplicate_worker_test.go`").

## Compatibility strategy

`W04-E03-S001`'s stopgap enforcement is explicitly superseded (not left running in parallel) once
T3 lands — see RISK-W04-E03-001 and its required mitigation (an explicit supersession step recorded
here, in this plan, and confirmed by the independent-review task). Any existing caller of the claim
path should observe the same external claim/no-work contract even though the underlying mechanism
and its batch-vs-single-item shape changes — if a caller genuinely depends on single-item claim
semantics, that dependency must be identified and addressed explicitly, not silently broken.

## Rollout strategy

Landed as its own reviewable unit following `W04-E03-S001`'s stopgap and `W04-E01-S001`'s primitive.
No further phased rollout beyond the T2→T3→T4→T5→T6 internal sequencing itself.

## Rollback strategy

If T3's atomic claim statement or T4's fencing check produces a regression under production-like
load, revert to `W04-E03-S001`'s stopgap enforcement as a fallback (it remains available in version
control even after superseding it in the mainline path) while the regression is diagnosed — this is
explicitly why RISK-W04-E03-001's mitigation requires a clean, well-documented supersession rather
than a destructive removal of the stopgap's code.

## Implementation sequence

As listed under "Implementation strategy" above (steps 1–7), directly mirroring the source's own
T2→T3→T4→T5→T6 dependency chain.

## Task breakdown

- **W04-E03-S002-T001** — Lease columns via the shared primitive (T2, step 2 above).
- **W04-E03-S002-T002** — Atomic leased claim, bounded batch, `EXPLAIN`-plan assertion (T3, step 3
  above).
- **W04-E03-S002-T003** — Item idempotency keys, finalize fencing (reusing `W04-E01-S002`), retry
  policy, cancellation (T4, step 4 above).
- **W04-E03-S002-T004** — Pause/resume/cancel lifecycle controls, bounded batch claims (T5, step 5
  above).
- **W04-E03-S002-T005** — Named multi-worker chaos test, reusing the shared harness from
  `W04-E01-S003` (T6, step 6 above).
- **W04-E03-S002-T006** — Independent review (per mandate §14, scoped to this story).

## Expected artifacts

The `bulk_items` lease-column migration; the atomic leased-claim SQL implementation; the item
idempotency-key/finalize-fencing/retry/cancellation code; the pause/resume/cancel lifecycle-control
API; the named chaos test `DATA-04/chaos/duplicate_worker_test.go`; documentation of all of the
above.

## Expected evidence

Lease-column migration test output; `EXPLAIN`-plan `SKIP LOCKED` assertion plus concurrent `N>1`
claimer test output; fenced-finalize-rejection test output; lifecycle integration-test output; the
named chaos test's output.

## Unresolved questions

- Exact bounded-batch default value and whether it is a hardcoded constant or a configuration key
  (T3).
- Exact item idempotency-key scheme (T4).
- Exact pause/resume/cancel API shape — operation-level flags, a state-machine column, or a separate
  control table (T5).
- Exact API shape of `W04-E01-S001`'s shared lease/fencing primitive, `W04-E01-S002`'s
  finalize-fencing logic, and `W04-E01-S003`'s shared chaos harness — not yet knowable from this
  epic's own planning scope; must be read directly from those stories' landed code before this
  story's T2/T4/T6 implementation work begins.
- Whether any existing `Service.next` caller depends on single-item (not batch) claim semantics —
  to be confirmed at implementation time; if so, the compatibility strategy above must be revisited.

## Approval conditions

This plan is approved for implementation once: (a) `W04-E01-S001`, `W04-E01-S002`, and
`W04-E01-S003` have landed (or are confirmed landing in lockstep) so their actual APIs/logic/harness
can be read and integrated against rather than guessed, (b) the unresolved questions above are
answered and documented, and (c) the owner and reviewer are assigned.
