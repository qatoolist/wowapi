---
id: PLAN-W04-E01-S002
type: plan
parent_story: W04-E01-S002
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Plan — W04-E01-S002

Per mandate §8.5. Confirmed facts, planned changes, and implementation assumptions are distinguished
explicitly below; this plan does not invent precise code changes where the repository does not yet
provide enough information.

## Proposed architecture

`jobs_queue` gains lease columns whose semantics are backed by W04-E01-S001's shared lease/fencing
primitive — this story is a consumer, not a redesign, of that primitive. Claim SQL is extended to
assign a fresh lease token and incremented generation atomically with the claim itself. The
`complete`/`fail` finalize code paths are extended to accept the caller's lease context and compare
it against the row's current lease state before applying the finalize, rejecting on mismatch.
`ReclaimStalled` is extended to bump `lease_generation` on every row it resets, using the same
primitive-provided generation-bump semantics S001 already implemented and unit-tested.

## Implementation strategy

1. Re-read `kernel/jobs`'s claim SQL, `claimedJob` struct, finalize code paths, and
   `ReclaimStalled` at this story's actual start commit to confirm the current-state assessment
   still holds.
2. Design the `jobs_queue` lease-column migration, mirroring S001's primitive schema; confirm
   reuse of the existing timeout-floor logic per T2's own risk note rather than introducing a second
   timeout source.
3. Implement the migration and extend claim SQL to assign a fresh lease token + `generation+1`;
   extend `claimedJob` to carry the lease context.
4. Write a migration + unit test proving claim assignment (AC-W04-E01-S002-01).
5. Extend the `complete`/`fail` finalize paths to compare lease token/generation and reject a
   mismatch; determine and document the rejection-surfacing mechanism (returned error, logged event,
   or both).
6. Extend `ReclaimStalled` to bump `lease_generation` on every reclaimed row.
7. Write a test proving: a stale finalize (simulating a since-reclaimed lease epoch) affects zero
   rows and is observably rejected (AC-W04-E01-S002-02); the same test additionally asserts the
   reclaimed row's `lease_generation` delta (AC-W04-E01-S002-03).
8. Confirm the fencing does not regress the existing at-least-once recovery path for a legitimate
   (non-superseded) worker's finalize — write a positive-case test alongside the negative one.
9. Document the lease-column schema and fencing behavior.

## Expected package or module changes

`kernel/jobs` — schema migration, claim SQL, `claimedJob` struct, finalize code paths,
`ReclaimStalled`. No new package; this story extends an existing one.

## Expected file changes where determinable

- A new migration file adding lease columns to `jobs_queue` (exact file path TBD, expected in the
  existing migrations directory).
- `kernel/jobs`'s claim SQL and `claimedJob` struct (exact file path TBD, not yet confirmed by
  file/line pending this story's own start-commit re-read).
- `kernel/jobs`'s finalize (`complete`/`fail`) code paths (exact file path TBD).
- `kernel/jobs`'s `ReclaimStalled` implementation (exact file path TBD).
- New tests for claim assignment, stale-finalize rejection, and reclaim generation-delta.

## Contracts and interfaces

`claimedJob`'s struct gains lease-context fields (token, generation) sourced from S001's primitive.
The finalize function signatures gain a lease-context parameter (or equivalent) used for the
fencing comparison. Exact signature shape to be determined at implementation time, consuming S001's
already-locked primitive API.

## Data structures

`jobs_queue`'s schema gains lease columns mirroring S001's primitive fields. `claimedJob`'s struct
gains corresponding in-memory fields.

## APIs

None affected externally — `kernel/jobs`'s internal claim/finalize/reclaim surface changes; no
public-facing API beyond the framework's own internal job-processing contract, whose signature
implications for worker code are S003's own T5 scope (deferred to that story per `story.md` "Out of
scope").

## Configuration changes

None anticipated beyond whatever S001's primitive itself introduces (e.g. `lease_expires_at`'s
default duration, if configurable) — this story does not introduce new configuration of its own.

## Persistence changes

Yes — the `jobs_queue` lease-column migration is this story's primary persistence change, per T2's
own required artifact path `DATA-09/jobs-lease-migration/` (source, verbatim path in the T-row is
`DATA-02/jobs-lease-migration/`).

## Migration strategy

Whether this migration explicitly routes through W02-E01's online-migration protocol (DATA-09) is
not mandated by the source for this specific migration — recorded as an implementation-time
decision, to be made consciously and documented, not defaulted silently. Compatibility with any
in-flight job claimed under the pre-fencing schema at deploy time must be considered (see `story.md`
"Compatibility considerations").

## Concurrency implications

The claim SQL's token/generation assignment must be atomic with the claim itself (no window where a
row is claimed but not yet lease-tagged). The finalize comparison must be safe under the exact
concurrent-claim/reclaim race this epic exists to fence against — this is the central correctness
property this story delivers, not an incidental concern.

## Error-handling strategy

A stale finalize attempt must be rejected observably (not silently ignored, not silently succeeding)
— T3's own acceptance criterion wording. A legitimate (non-superseded) finalize must succeed exactly
as it does today, with no regression to the existing at-least-once recovery path (T3's own risk
note).

## Security controls

The fenced finalize path is itself the security-relevant control (see `story.md` "Security
considerations") — not optional hardening.

## Observability changes

A stale-finalize rejection event should be observable (at minimum, logged or metric-recorded)
distinctly from a successful finalize, per T3's acceptance criterion ("observably rejected").

## Testing strategy

- Migration + unit test: claim SQL assigns a fresh token + `generation+1`; `claimedJob` carries the
  lease context.
- Stale-finalize test: a finalize attempt using a superseded token/generation affects zero rows and
  is observably rejected.
- Reclaim generation-delta test: `ReclaimStalled` bumps `lease_generation`; the same test as the
  stale-finalize test additionally asserts this delta, per T4's own "Same test as T3" instruction.
- Positive-case (non-regression) test: a legitimate, non-superseded finalize still succeeds exactly
  as before fencing was introduced.

## Regression strategy

The stale-finalize/reclaim-generation test pair, once passing, becomes the regression guard for this
story's fencing behavior going forward. The positive-case test guards against a fencing
implementation that is "too strict" and rejects legitimate finalizes.

## Compatibility strategy

To be resolved in this story's own implementation: how the lease-column migration handles any job
claimed under the pre-fencing schema at deploy time (see "Migration strategy" above). No source
guidance exists specifically for this point beyond T2's timeout-floor-reuse risk note.

## Rollout strategy

Single story, landed as its own reviewable unit, consuming S001's already-locked primitive.

## Rollback strategy

Revert the lease-column migration and fencing code if the finalize comparison produces false-
positive rejections against legitimate, non-superseded finalizes under real load — escalate for
redesign rather than silently loosening the fencing comparison without recording why.

## Implementation sequence

As listed under "Implementation strategy" above (steps 1–9). Step 3 (migration + claim SQL) must
land before step 5 (finalize fencing), which must land before step 6 (reclaim generation-bump),
matching PLAN DATA-02's own T2→T3, T2→T4 dependency structure.

## Task breakdown

- **W04-E01-S002-T001** — Lease-column migration and fenced claim SQL (steps 2–4 above).
- **W04-E01-S002-T002** — Fenced finalize paths (steps 5, 7 [stale-finalize half], 8 above).
- **W04-E01-S002-T003** — Fenced reclaim with generation bump (steps 6, 7 [generation-delta half]
  above).
- **W04-E01-S002-T004** — Independent review (per mandate §14, scoped to this story).

## Expected artifacts

The `jobs_queue` lease-column migration; the fenced claim/finalize/reclaim code; documentation of
the lease-column schema and fencing behavior.

## Expected evidence

Migration + unit-test output for claim assignment; stale-finalize-rejection test output; reclaim
generation-delta test output.

## Unresolved questions

- The exact existing timeout-floor logic to reuse for the lease-column migration (T2's own risk
  note) — to be confirmed by this story's own start-commit re-read, not invented here.
- Whether the stale-finalize rejection surfaces as a returned error, a logged event, or both — to be
  chosen and documented at implementation time.
- Whether this migration explicitly routes through W02-E01's online-migration protocol (DATA-09) —
  not mandated by the source for this specific migration; a conscious implementation-time decision.
- The exact compatibility handling for a job claimed under the pre-fencing schema at deploy time.

## Approval conditions

This plan is approved for implementation once: (a) the unresolved questions above — most centrally,
the migration-protocol routing decision — are answered, and (b) the owner and reviewer are assigned,
and (c) W04-E01-S001 has reached at least `implemented` status (this story's own dependency).
