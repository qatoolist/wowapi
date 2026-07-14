---
id: PLAN-W02-E04-S001
type: plan
parent_story: W02-E04-S001
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
---

# Plan — W02-E04-S001

Per mandate §8.5. Confirmed facts, planned changes, and implementation assumptions are distinguished
explicitly below; this plan does not invent precise code changes where the repository does not yet
provide enough information.

## Proposed architecture

A new typed aggregate repository/unit-of-work helper in `kernel/resource`, wrapping the existing
business-row write, the registrar's mirror `Upsert`, an audit-row write, and an outbox-entry write
in a single database transaction. The existing low-level `Upsert` API remains available (per PLAN's
own wowsociety-compatibility note), so this is an additive, preferred-path change, not a breaking
replacement. No new package is introduced; the helper is added to the existing `kernel/resource`
package (or a clearly-scoped sub-package, to be determined at implementation time).

## Implementation strategy

1. Re-read `kernel/resource`'s current package documentation, `registrar_pg.go:38-58`, and the
   reference handler's current implementation at this story's actual start commit, confirming the
   current-state assessment still holds (manual two-statement pattern, `uuid.Nil` actor).
2. Design the typed aggregate repository/unit-of-work helper's interface: what it accepts (the
   business-row write operation, module-supplied), what it produces (mirror upsert, audit write,
   outbox write, all framework-owned), and how it enforces atomicity (a single transaction spanning
   all four writes).
3. Implement the helper (T1).
4. Write the fault-injection test suite: inject a failure at each of the 4 stages (business write,
   mirror upsert, audit write, outbox write) independently, confirming full rollback at every stage.
5. Implement the actor-attribution fix inside the same helper (T2): source `created_by` from
   context; reject a missing actor for a user-initiated write; confirm a system-actor path is
   unaffected.
6. Write the actor-attribution test (with/without actor, system vs. user path).
7. Migrate the reference handler onto the new helper (T3), confirming existing reference tests still
   pass.
8. Update `kernel/resource` documentation to describe the implemented contract (T4).

## Expected package or module changes

`kernel/resource` (new helper, updated documentation); `kernel/resource/registrar_pg.go` (actor-
attribution fix); the reference handler's package (exact location TBD, migrated onto the new
helper).

## Expected file changes where determinable

- `kernel/resource/registrar_pg.go` — actor-attribution fix (exact line range to be re-confirmed at
  implementation time; PLAN cites `:38-58` as of the source document's writing).
- A new file (or addition to an existing file) implementing the typed aggregate repository/unit-of-
  work helper — exact path TBD.
- `kernel/resource`'s package documentation file — updated per T4.
- The reference handler's source file — exact path TBD, migrated to call the new helper.

## Contracts and interfaces

A new helper interface/type (exact shape TBD) that accepts a module's business-row write operation
and internally performs the mirror upsert, audit write, and outbox write as part of the same
transaction. The existing low-level `Upsert` API on the registrar remains unchanged and available.

## Data structures

No new persistent data structure — the mirror, audit, and outbox tables already exist and are
already written to (just not atomically, and with a placeholder actor). This story changes how they
are written, not their schema.

## APIs

The new helper is an additive API within `kernel/resource`; no existing public API is removed. The
low-level `Upsert` API's signature is unchanged.

## Configuration changes

None anticipated.

## Persistence changes

None — no schema change. The transactional scope of existing writes changes (four writes now share
one transaction instead of at least two separate ones), which is a behavior change but not a schema
change.

## Migration strategy

Not applicable — no schema or data migration (see `story.md` "Migration considerations").

## Concurrency implications

Bundling four writes into one transaction increases the transaction's duration compared to today's
two-independent-statements pattern; this is an intentional atomicity/consistency trade-off. No new
concurrency primitive (lock, lease, etc.) is introduced by this story.

## Error-handling strategy

A fault at any of the four stages (business write, mirror upsert, audit write, outbox write) must
roll back the entire transaction — no partial write. A missing actor on a user-initiated write must
fail fast with a clear error, not silently substitute a placeholder.

## Security controls

The actor-attribution fix (T2) is itself the required security/accountability control for this
story — see `story.md` "Security considerations."

## Observability changes

None mandated beyond clear fault-injection-test attribution of which stage failed (implementation-
time testing convenience, not a runtime observability requirement).

## Testing strategy

- Fault-injection test suite: 4 independent fault points (business write, mirror upsert, audit
  write, outbox write), each confirming full transaction rollback.
- Actor-attribution test: with actor present (user-initiated, succeeds with real `created_by`);
  without actor present on a user-initiated write (fails fast); system-actor path (succeeds,
  unaffected).
- Reference-handler regression test: existing reference tests continue to pass after migration onto
  the new helper.
- Manual documentation review (T4's own verification method, per PLAN's own "Tests" column for T4:
  "Manual review").

## Regression strategy

The reference-handler migration (T3) itself is the primary regression check — existing reference
tests passing unmodified after the migration confirms the new helper preserves the reference
handler's observable behavior.

## Compatibility strategy

The existing low-level `Upsert` API remains available (per PLAN's own wowsociety-compatibility
note) — no existing consumer (including wowsociety's `committeeseat.go`, which is out of this
story's scope) is broken by this story landing.

## Rollout strategy

Single story, landed as its own reviewable unit. The reference handler's migration (T3) is the
proof-of-pattern; other module consumers migrate on their own schedule (not mandated by this story).

## Rollback strategy

Revert the helper and the reference-handler migration if fault-injection testing or the actor-
attribution test surfaces a defect that cannot be resolved within this story's bounded scope; the
low-level `Upsert` API remaining available means a revert does not strand any consumer.

## Implementation sequence

As listed under "Implementation strategy" above (steps 1–8), matching PLAN DATA-06's own Depends-on
column: T1 first (steps 2–4), T2 second (steps 5–6, depends on T1), T3 third (step 7, depends on
T1+T2), T4 in parallel with or after T1 (step 8, depends on T1 only).

## Task breakdown

- **W02-E04-S001-T001** — Typed aggregate repository/unit-of-work helper (steps 2–4 above).
- **W02-E04-S001-T002** — Actor-attribution fix (steps 5–6 above); single owner of the shared
  `registrar_pg.go` fix surface with DATA-07 T3.
- **W02-E04-S001-T003** — Reference-handler migration (step 7 above).
- **W02-E04-S001-T004** — `kernel/resource` documentation update (step 8 above).
- **W02-E04-S001-T005** — Independent review (per mandate §14, scoped to this story).

## Expected artifacts

The typed aggregate repository/unit-of-work helper; the `registrar_pg.go` actor-attribution fix; the
migrated reference handler; updated `kernel/resource` documentation.

## Expected evidence

Fault-injection test output (4 stages); actor-attribution test output; reference-handler regression
test output; documentation-review record.

## Unresolved questions

- Exact package location for the new helper (within `kernel/resource` itself, or a new sub-package)
  — to be decided at implementation time.
- Exact current line range of `registrar_pg.go`'s nil-actor placeholder (PLAN cites `:38-58` at the
  time PLAN was written) — to be re-confirmed at implementation time.
- Exact identity (fully-qualified path) of "the reference handler" PLAN refers to — to be confirmed
  at implementation time; if ambiguous, the choice and rationale are recorded in T3's task record.
- Exact fault-injection mechanism (a test-only hook, an interface substitution, or another approach)
  for simulating a failure at each of the 4 stages independently — to be determined at
  implementation time.

## Approval conditions

This plan is approved for implementation once: (a) the unresolved questions above are answered by a
first re-read of `kernel/resource`, `registrar_pg.go`, and the reference handler at story start, and
(b) the owner and reviewer are assigned.
