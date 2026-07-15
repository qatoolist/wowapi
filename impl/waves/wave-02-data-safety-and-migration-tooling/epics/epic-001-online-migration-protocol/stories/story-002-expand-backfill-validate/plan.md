---
id: PLAN-W02-E01-S002
type: plan
parent_story: W02-E01-S002
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
---

# Plan — W02-E01-S002

Per mandate §8.5. Confirmed facts, planned changes, and implementation assumptions are distinguished
explicitly below. This plan does not invent the batch size/rate/window values PLAN T4 itself
classifies as "human decision... per migration" — it builds the harness's configuration surface, not
a specific migration's configured values.

## Proposed architecture

Three tooling layers built on top of S001's manifest schema and lock-timeout mechanism: (1)
expand-phase tooling that issues schema-additive, non-blocking DDL; (2) a backfill-job harness with
its own interim checkpoint-lease primitive, deliberately scoped narrower than DATA-02 T1's eventual
shared primitive; (3) validation-phase tooling producing a machine-checked artifact. No existing
kernel package's public contract changes — this is new, additive tooling.

## Implementation strategy

1. Confirm T3's own risk note by testing whether the current tooling supports issuing DDL statements
   outside the wrapping transaction (a prerequisite for non-transactional `CREATE INDEX
   CONCURRENTLY`).
2. Implement expand-phase tooling: nullable/default-safe column helpers, new table/index/
   compatibility-view creation helpers, `NOT VALID` constraint helpers, non-transactional
   `CREATE INDEX CONCURRENTLY` issuance.
3. Write the old-reader-compatibility test: confirm an old-version application reader and a
   new-version reader both accept the expanded schema during the migration window.
4. Design the interim checkpoint-lease mechanism, explicitly scoped to checkpoint-token +
   resumability only (no fencing generations, no heartbeats) — document this scope boundary
   in the mechanism's own code comments and in this story's documentation, so a future reader
   (including W04-E01-S001's own implementer) does not mistake it for DATA-02 T1's full primitive.
5. Implement the backfill-job harness: resumable, tenant-scoped, keyset-paginated, checkpointed
   (using step 4's interim lease), with bounded batch/tx time and rate controls (configurable, not
   hardcoded to a specific migration's values).
6. Write the named interrupted/resumed backfill test: interrupt a backfill mid-run, resume it, and
   confirm no row is reprocessed and no row is skipped.
7. Implement validation-phase tooling: `VALIDATE CONSTRAINT` orchestration, reconciliation query
   execution, and machine-checked artifact capture (a defined artifact schema, not free-form
   prose).
8. Write an artifact-schema test confirming the validation-phase report conforms to its schema and
   correctly reports zero mismatches on clean data (and, if feasible, a non-zero-mismatch case to
   confirm the report correctly surfaces a mismatch rather than silently passing).

## Expected package or module changes

New packages for expand-phase tooling, the backfill-job harness (including the interim checkpoint-
lease mechanism), and validation-phase tooling (exact locations TBD, expected adjacent to S001's
manifest/lock-timeout code).

## Expected file changes where determinable

Not yet determinable by file/line — this is new tooling from zero, and the exact package structure
depends on decisions made during S001's own implementation (e.g. where the manifest-schema
validation tool lives, since this story's tooling is expected to be layered adjacent to it).

## Contracts and interfaces

The interim checkpoint-lease's own interface (checkpoint-token issuance, resumability check) —
deliberately scoped narrower than DATA-02 T1's eventual shared-primitive interface, so that
W04-E01-S001's later migration is a scope expansion, not a breaking interface replacement, if this
story's own interface design anticipates that expansion path. This anticipation is a design goal for
this story's plan, not a confirmed guarantee — to be validated at implementation time.

## Data structures

A checkpoint-state record (backfill job ID, last-processed keyset position, checkpoint token) for
the interim lease mechanism. A validation-phase report artifact schema (per-constraint/per-query
mismatch counts, pass/fail, timestamps).

## APIs

None affected — this is new, additive tooling, not a change to an existing runtime API.

## Configuration changes

The backfill harness's batch size, rate, and window controls are configurable per PLAN T4's own
framing ("human decision on batch size/rate/window per migration") — exact configuration mechanism
(config keys, command-line flags, or per-migration manifest fields consuming S001's schema) to be
determined at implementation time.

## Persistence changes

New tables or columns for the interim checkpoint-lease's state (exact schema TBD) and, if the
validation-phase artifact capture requires durable storage rather than a file-based artifact, a
table for that as well — to be determined at implementation time.

## Migration strategy

This story's own tooling has no data to migrate; it is the tooling other migrations (starting with
W02-E02's DATA-01 rollout) will use.

## Concurrency implications

The backfill harness must correctly handle concurrent access to its own checkpoint state (a resume
attempted twice concurrently, or a backfill job restarted while a previous instance is still
finishing) — this is exactly the "resumable... checkpointed" requirement PLAN T4 names, and is the
subject of the interrupted/resumed test itself.

## Error-handling strategy

A backfill interruption must leave the checkpoint state in a consistent, resumable position — no
partial-batch commit that would cause reprocessing or skipping on resume. Validation-phase tooling
must fail loudly (not silently pass) on a malformed or incomplete reconciliation query result.

## Security controls

None beyond what S001's lock-timeout mechanism already provides. No new access-control surface is
introduced by this story's tooling.

## Observability changes

The backfill harness's checkpoint state should be queryable for progress monitoring (implementation-
time addition supporting the interrupted/resumed test's own verifiability).

## Testing strategy

- Old-reader-compatibility test (T3's own named test).
- Interrupted/resumed backfill test (T4's own explicitly-required named test) — "no reprocessing or
  skipping."
- Artifact-schema test for validation-phase reports (T5's own named test).
- No separate race-detector test is named by the source beyond the interrupted/resumed scenario
  itself, which is inherently a concurrency/interruption test.

## Regression strategy

Once landed, this story's tooling becomes the required path for any future online migration —
W02-E02's DATA-01 rollout (this wave) is the first real consumer, and its own use of this tooling is
itself a regression-style proof that the tooling generalizes beyond its own test fixtures.

## Compatibility strategy

Expand-phase tooling's compatibility strategy is its entire purpose (see "Compatibility
considerations" in `story.md`). The interim checkpoint-lease's compatibility strategy toward its own
future replacement (W04-E01-S001) is addressed under "Contracts and interfaces" above.

## Rollout strategy

Single story, landed as its own reviewable unit. No production migration is executed as part of this
story's own rollout — W02-E02 is the first real consumer, scheduled as a separate story.

## Rollback strategy

Revert expand-phase, backfill, or validation-phase tooling independently if any is found to produce
incorrect results in its own test suite; the backfill harness's own interrupted/resumed test is the
primary rollback trigger — any regression in row-reprocessing/skipping behavior must halt further
rollout of that tooling immediately.

## Implementation sequence

As listed under "Implementation strategy" above (steps 1–8). Step 4 (interim-lease scope-boundary
design) must be settled before step 5 (harness implementation) begins, since the harness's
checkpoint mechanism depends on it.

## Task breakdown

- **W02-E01-S002-T001** — Expand-phase tooling (steps 1–3 above).
- **W02-E01-S002-T002** — Backfill-job harness and interim checkpoint-lease mechanism (steps 4–6
  above) — this is the story's highest-risk task, carrying RISK-W02-001.
- **W02-E01-S002-T003** — Validation-phase tooling (steps 7–8 above).
- **W02-E01-S002-T004** — Independent review (per mandate §14, scoped to this story, with specific
  attention to the interim-lease deviation being honestly recorded).

## Expected artifacts

Expand-phase tooling; the backfill-job harness and interim checkpoint-lease mechanism; validation-
phase tooling and its artifact schema; documentation for all three including the interim-lease
scope-boundary note.

## Expected evidence

Old-reader-compatibility test output; interrupted/resumed backfill test output; artifact-schema test
output for the validation-phase report.

## Unresolved questions

- Exact configuration mechanism for the backfill harness's batch size/rate/window controls (config
  keys vs. manifest fields vs. command-line flags).
- Exact persistence schema for the interim checkpoint-lease's state.
- Whether the validation-phase artifact is file-based or database-persisted.
- Whether the interim checkpoint-lease's interface can be designed to anticipate a clean expansion
  path to DATA-02 T1's full primitive, or whether W04-E01-S001 will necessarily require an interface
  break — to be assessed once the interim lease's actual interface is designed.

## Approval conditions

This plan is approved for implementation once: (a) W02-E01-S001 has landed (or is far enough along
that its manifest schema's shape is stable), (b) the interim checkpoint-lease's scope-boundary design
is documented and reviewed specifically for the "not DATA-02 T1's full primitive" distinction, and
(c) the owner and reviewer are assigned.
