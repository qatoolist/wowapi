---
id: PLAN-W04-E01-S001
type: plan
parent_story: W04-E01-S001
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Plan — W04-E01-S001

Per mandate §8.5. Confirmed facts, planned changes, and implementation assumptions are distinguished
explicitly below; this plan does not invent precise code changes where the repository does not yet
provide enough information.

## Proposed architecture

A new, standalone kernel-level lease/fencing type — not embedded inside `kernel/jobs` — exposing an
opaque `lease_token`, a monotonically increasing `lease_generation`, `lease_expires_at`, and an
optional heartbeat extension point. The type's own comparison semantics (does a given
token/generation pair still represent the current lease holder) are the primitive's core contract;
callers (this epic's S002, and eventually W04-E02/W04-E03) supply their own persistence (their own
table's lease columns) and call into the primitive for comparison/generation-bump logic, rather than
the primitive owning persistence itself. This keeps the primitive a reusable building block rather
than a `jobs_queue`-specific mechanism.

## Implementation strategy

1. Re-read `kernel/jobs`'s claim/finalize/reclaim SQL and W02-E01-S002's interim checkpoint-lease
   implementation fresh at this story's actual start commit to confirm the current-state assessment
   still holds.
2. Design the primitive's exact type/package location and its token/generation comparison API,
   documenting trade-offs among candidate locations (new `kernel/lease` package vs. a subpackage
   elsewhere).
3. Implement the primitive: `lease_token` generation, monotonic `lease_generation` semantics,
   `lease_expires_at` handling, and the optional heartbeat extension point.
4. Write unit tests on token/generation comparison semantics per PLAN T1's own "Tests" column.
5. Validate the primitive's field set and semantics against DATA-03's (W04-E02) and DATA-04's
   (W04-E03) own stated needs — read their PLAN task rows and the wave-level `wave.md`'s framework-
   capabilities list for what each requires from a shared lease type — and record the review outcome
   before treating the design as locked, per RISK-W04-E01-001's mitigation.
6. Design and execute the interim-checkpoint-lease migration: read any existing
   W02-E01-S002-interim-lease checkpoint state, re-express it under this primitive's schema, and
   remove the interim lease code path.
7. Write a migration test proving no in-flight backfill checkpoint state is lost or duplicated
   across the cutover.
8. Document the primitive's contract and the completed migration.

## Expected package or module changes

A new lease/fencing primitive package (exact location TBD — see "Unresolved questions"). Removal of
W02-E01-S002's interim checkpoint-lease code path, replaced by calls into the new primitive.

## Expected file changes where determinable

- A new lease/fencing primitive package and its unit tests (exact file path TBD).
- W02-E01-S002's interim checkpoint-lease implementation file(s) — modified to migrate state, then
  the interim-specific code removed (exact file path TBD, not yet confirmed by file/line pending
  this story's own start-commit re-read).
- A new migration test proving no checkpoint-state loss/duplication across the cutover.

## Contracts and interfaces

The lease/fencing primitive's own type and comparison API: fields `lease_token`,
`lease_generation` (monotonic), `lease_expires_at`, and an optional heartbeat extension point;
methods/functions for claiming (issuing a fresh token + generation bump), comparing a supplied
token/generation against the current lease state, and expiry checking. Exact typing/signature to be
determined per the design step above and validated against DATA-03/DATA-04's needs before locking.

## Data structures

The lease/fencing primitive's own struct/type, per "Contracts and interfaces" above. No application
data model change from this story alone (S002 adds the `jobs_queue` lease columns that use this
primitive).

## APIs

None affected — this story is kernel-internal tooling, not a runtime API change.

## Configuration changes

None anticipated. `lease_expires_at`'s default duration/budget, if configurable, is recorded as an
implementation-time decision — not confirmed by the source beyond the primitive's own field list.

## Persistence changes

None from this story directly — the primitive itself does not own persistence (see "Proposed
architecture"). The interim-checkpoint-lease migration (step 6/7 above) does read and re-express
existing persisted checkpoint state, but does not introduce a new schema of its own beyond what the
migration mechanics require.

## Migration strategy

The interim-checkpoint-lease migration (see "Implementation strategy" steps 6-7) is this story's own
migration concern — reading state written under W02-E01-S002's interim lease format and
re-expressing it under this primitive's schema. Per RISK-W04-001's mitigation, this must be an
explicit migration step, not a big-bang cutover.

## Concurrency implications

The primitive's own token/generation comparison logic must be safe under concurrent claim attempts
(the scenario it exists to fence against) — this is the primitive's core correctness property, not
an incidental concern. The interim-lease migration must also correctly handle a backfill job that is
genuinely in-flight at cutover time (per RISK-W04-001's contingency: pause, migrate, resume, rather
than cutting over underneath a running job).

## Error-handling strategy

A comparison against a stale or mismatched token/generation must return an unambiguous
"fencing rejected" result, not a silent pass or an ambiguous error. The interim-lease migration must
fail loudly (not silently drop state) if it encounters checkpoint state it cannot confidently
re-express under the new schema.

## Security controls

Correct token/generation comparison semantics are themselves the security-relevant property this
primitive exists to provide (preventing a stale worker's writes from being accepted as current) —
not optional hardening.

## Observability changes

None separately mandated for this story specifically; S002's own fenced finalize/reclaim paths are
where observable fencing-rejection events matter operationally (see `story.md` "Observability
considerations").

## Testing strategy

- Unit tests on token/generation comparison semantics: a current token/generation pair compares as
  valid; a stale one (superseded generation, or expired) compares as rejected.
- Migration test: simulate an in-flight backfill checkpoint written under the interim lease format,
  execute the migration, and confirm the checkpoint state is correctly readable under the new
  primitive's schema with no loss or duplication.
- No separate integration or race test is mandated for this story specifically beyond the
  comparison-semantics unit tests and the migration test — S002's own fenced-application tests
  exercise the primitive under realistic concurrent-claim conditions.

## Regression strategy

Once the primitive is locked and consumed by S002 (and eventually W04-E02/W04-E03), any future
change to its comparison semantics is itself a breaking change to three consumers at once — this is
exactly why AC-W04-E01-S001-02 requires the cross-consumer field-set review before locking.

## Compatibility strategy

The interim-checkpoint-lease migration is this story's primary compatibility concern (see
"Migration strategy" and `story.md` "Compatibility considerations"). No other compatibility concern
is identified for the primitive's own initial introduction, since nothing outside W02-E01-S002
currently depends on any predecessor lease mechanism.

## Rollout strategy

Single story, landed as its own reviewable unit. The interim-lease migration should complete within
this story's own rollout — not left as a dangling follow-up — given RISK-W04-001's framing that an
incomplete migration risks reprocessing or skipping backfill rows.

## Rollback strategy

If the primitive's design is found materially deficient after S002 or the cross-consumer review
begins consuming it, revert to a design-review cycle rather than shipping a known-deficient shared
type across three epics. If the interim-lease migration surfaces checkpoint state it cannot safely
translate, halt the migration, do not remove the interim lease code path, and escalate per
RISK-W04-001's contingency (pause any in-flight backfill, complete the migration, then resume).

## Implementation sequence

As listed under "Implementation strategy" above (steps 1–8). Step 5 (cross-consumer review) must
occur before the primitive's design is treated as locked; step 6 (interim-lease migration) must
complete, with step 7's test passing, before the interim lease code path is removed.

## Task breakdown

- **W04-E01-S001-T001** — Shared primitive design, implementation, and cross-consumer field-set
  review (steps 2–5 above).
- **W04-E01-S001-T002** — Interim-checkpoint-lease migration (steps 6–7 above).
- **W04-E01-S001-T003** — Independent review (per mandate §14, scoped to this story).

## Expected artifacts

The shared lease/fencing primitive package; the interim-checkpoint-lease migration tooling/code;
documentation of the primitive's contract and the completed migration.

## Expected evidence

Unit-test output for token/generation comparison semantics; the cross-consumer field-set review
record; migration test output proving no checkpoint-state loss/duplication across the cutover.

## Unresolved questions

- Exact package location for the shared primitive (new `kernel/lease` package vs. elsewhere) — to be
  decided at implementation time per the design step above.
- Whether the optional heartbeat extension point is actually exercised by any of S002/S003/W04-E02/
  W04-E03's own needs, or left as a documented-but-unused extension point at this story's own
  landing — to be determined by the cross-consumer field-set review (step 5).
- Exact interim-checkpoint-lease migration mechanics (read-then-translate-then-remove vs. a
  dual-write transition window) — to be chosen and documented at implementation time, per
  RISK-W04-001's mitigation requiring an explicit migration step.
- Whether `lease_expires_at`'s default duration is a hardcoded constant or a configuration key — not
  confirmed by the source beyond the primitive's own field list.

## Approval conditions

This plan is approved for implementation once: (a) the unresolved questions above — most centrally,
the primitive's package location and the cross-consumer field-set review's outcome — are answered,
and (b) the owner and reviewer are assigned.
