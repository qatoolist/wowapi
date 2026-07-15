---
id: PLAN-W02-E01-S001
type: plan
parent_story: W02-E01-S001
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
---

# Plan — W02-E01-S001

Per mandate §8.5. Confirmed facts, planned changes, and implementation assumptions are distinguished
explicitly below; this plan does not invent precise code changes where the repository does not yet
provide enough information — PLAN DATA-09 T1 itself is explicit that "per-migration classification
is human judgment every time," so this plan defines the schema and the enforcement mechanism, not a
predetermined classification for every future migration.

## Proposed architecture

A new migration-manifest validation layer, sitting alongside the existing migration-registration
mechanism (`Makefile`'s `migrate` target, `check_migrations.sh`), that adds a machine-checkable
declaration of a migration's risk profile. A lock-timeout enforcement wrapper around DDL execution
that aborts and retries within a bounded ceiling when a statement exceeds the 2-second budget. No
change to the existing migration-apply mechanism's own forward-apply semantics is required by this
story — the manifest and lock-budget layers are additive checks and a runtime wrapper, not a
replacement for how migrations are applied.

## Implementation strategy

1. Re-read `Makefile`'s `migrate` target and `check_migrations.sh` fresh at this story's actual
   start commit to confirm the current-state assessment (no manifest, no lock-budget concept) still
   holds.
2. Design the manifest schema's exact format: draft options (inline migration-file header comment,
   sibling YAML/JSON manifest file per migration, or a manifest registry file) with trade-offs, and
   select one, documenting the rationale.
3. Submit the schema design for external review per PLAN T1's own risk note, before treating it as
   locked.
4. Implement the manifest-schema CI validation: a tool that reads every migration's manifest entry
   and fails the build on a missing required field.
5. Write a negative fixture test: a migration with a manifest entry missing a required field, and
   confirm CI fails against it.
6. Implement the lock-timeout enforcement mechanism: wrap DDL execution for online-classified
   migrations with a 2-second lock-timeout budget, abort-and-retry on timeout, and a bounded retry
   ceiling (exact bound to be set at implementation time, documented in the mechanism itself).
7. Write a test against a deliberately concurrently-locked table, confirming clean abort (no partial
   DDL) and bounded retry behavior.
8. Document both the manifest schema and the lock-budget mechanism.

## Expected package or module changes

A new manifest-schema validation tool/library and a lock-timeout enforcement mechanism (exact
package location TBD — see "Unresolved questions"). Extensions to `Makefile` and/or
`check_migrations.sh` to invoke the new validation as part of the existing migration-check flow, or
a new standalone CI step if that integration is cleaner — to be determined at implementation time.

## Expected file changes where determinable

- A new manifest-schema definition and validator (exact file path TBD).
- A new lock-timeout enforcement mechanism (exact file path TBD, expected near the existing
  migration-execution code, location not yet confirmed by file/line).
- `Makefile` and/or `check_migrations.sh` — extended to invoke the new manifest validation (or
  superseded by it, depending on the schema-format decision in step 2 above).
- New negative fixture migration(s) for the CI validation test.

## Contracts and interfaces

A manifest-entry data contract (fields: online/maintenance classification, rows/bytes estimate,
lock/statement timeout, N/N-1 compatibility flag, backfill owner, validation query, rollback/
forward-fix plan) — exact typing/serialization format to be determined per the schema-format
decision above.

## Data structures

The manifest-entry struct/schema itself, per "Contracts and interfaces" above. No application data
model change.

## APIs

None affected — this story is tooling-internal, not a runtime API change.

## Configuration changes

None anticipated beyond the lock-timeout budget and retry-ceiling values, which may be hardcoded
constants or configuration keys — to be determined at implementation time (PLAN T2's own framing,
"human-set retry ceiling," suggests a deliberately-chosen fixed value rather than a runtime-tunable
config key, but this is not confirmed by the source).

## Persistence changes

None. This story adds tooling that inspects migrations; it does not itself add or change any
database table.

## Migration strategy

Not applicable in the schema/data-migration sense — this story is itself part of the migration-
tooling layer, not a consumer of it.

## Concurrency implications

The lock-timeout enforcement mechanism directly concerns concurrency: it protects concurrent
traffic against a DDL statement holding a lock beyond the 2-second budget. Its own abort-and-retry
loop must itself be safe under concurrent migration attempts (unlikely in practice, since migrations
are typically applied by a single deploy process, but the retry ceiling's bound should not assume
single-threaded execution without confirming that assumption at implementation time).

## Error-handling strategy

A DDL statement exceeding the lock-timeout budget must abort cleanly — no partial DDL applied. A
migration failing manifest validation must fail the build with a clear, field-specific error
message (not a generic validation failure), so a migration author knows exactly which field is
missing.

## Security controls

The bounded retry ceiling is itself the required security control (PLAN T2's own risk note: "Bound
total retries — unbounded retry is a deploy-time DoS"). This is not optional hardening; it is a
required acceptance-criterion-adjacent control.

## Observability changes

A lock-timeout abort/retry event should be logged (implementation-time addition, not separately
mandated by the source beyond the retry-ceiling requirement itself — see `story.md` "Observability
considerations").

## Testing strategy

- Fail-first / positive-negative pair: a migration with a complete manifest entry validates; a
  migration missing a required field fails CI (negative fixture test).
- Lock-timeout test: against a deliberately concurrently-locked table, confirm clean abort (no
  partial DDL) and bounded retry behavior within the ceiling.
- No integration or race test is separately required beyond the concurrently-locked-table test
  itself, which is inherently a concurrency scenario.

## Regression strategy

The manifest-schema CI validation, once wired into the existing migration-check flow (or its
successor), becomes the regression guard: any future migration lacking a manifest entry fails CI
going forward.

## Compatibility strategy

To be resolved in this story's own implementation: whether the manifest-schema CI gate is enforced
immediately against all future migrations, or phased in with a transition period for any migration
already in flight at this story's landing. No source guidance exists specifically for DATA-09 T1 on
this point; the decision and its rationale must be recorded, not silently defaulted.

## Rollout strategy

Single story, landed as its own reviewable unit — no phased rollout beyond the compatibility-
strategy question above (which concerns migration-author-facing enforcement timing, not this
story's own code rollout).

## Rollback strategy

Revert the manifest-schema CI gate and lock-timeout wrapper if either produces false positives that
block a legitimate migration; the manifest schema itself, once migrations have been authored against
it, is harder to revert without a compatibility plan — this is exactly why AC-W02-E01-S001-02
requires external review before the schema is locked.

## Implementation sequence

As listed under "Implementation strategy" above (steps 1–8). Step 3 (external review) must occur
before step 4 (CI validation implementation) locks in the schema as enforced.

## Task breakdown

- **W02-E01-S001-T001** — Manifest schema design, external review, and CI validation (steps 2–5
  above).
- **W02-E01-S001-T002** — Lock-timeout enforcement mechanism (steps 6–7 above).
- **W02-E01-S001-T003** — Independent review (per mandate §14, scoped to this story).

## Expected artifacts

The manifest schema definition; the manifest-schema CI validator; the lock-timeout enforcement
mechanism; manifest-schema and lock-budget documentation.

## Expected evidence

Manifest-schema positive/negative fixture test output; lock-timeout abort/retry test output against
a concurrently-locked table; the external-review record for the manifest schema.

## Unresolved questions

- Exact manifest storage format (inline header comment, sibling file, or registry) — to be decided
  at implementation time per the design step above, with external review before locking.
- Exact retry-ceiling bound (number of retries, backoff schedule) for the lock-timeout mechanism —
  "human-set retry ceiling" per PLAN T2, value to be chosen and documented at implementation time.
- Whether the lock-timeout budget/retry ceiling are hardcoded constants or configuration keys.
- Whether the manifest-schema CI gate is enforced immediately or phased in with a transition period.
- Exact package location for the new validation tool/lock-timeout mechanism.

## Approval conditions

This plan is approved for implementation once: (a) the unresolved questions above — most centrally,
the manifest storage format — are answered following the external-review step required by
AC-W02-E01-S001-02, and (b) the owner and reviewer are assigned.
