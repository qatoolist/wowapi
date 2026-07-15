---
id: IMPL-W03-E04-S001
type: implementation-record
parent_story: W03-E04-S001
status: recorded
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Implementation record — W03-E04-S001

## What was actually implemented

- DATA-07 T1: `Checker.Has` now resolves an actor's active capacity to its
  optional party via `acting_capacities.party_id`, enabling party-subject
  relationship edges to grant ReBAC access.
- DATA-07 T2: `Checker.Has` evaluates the full schema-enumerated subject-kind
  matrix (`capacity`, `party`, `resource`) with an explicit fail-closed default
  for any unenumerated kind.
- DATA-07 T3/T4 (non-cache-invalidation): `Relate` now requires a bound actor
  in ctx (ownership fail-closed), sources `created_by`/`updated_by` from
  `database.ActorIDFrom` (DATA-06 T2 attribution), upserts active edges while
  bumping `version`, and writes a durable `audit_logs` row in the same
  transaction.
- Cache-invalidation portion of AC-W03-E04-S001-03 remains deferred-linked to
  W05-E04-S002 per plan; no code was fabricated against a non-existent epoch
  table.

## Components changed

- `kernel/relationship` — `Checker.Has` evaluation logic and `Relate` mutation
  governance.

## Files changed

- `kernel/relationship/relationship.go`
- `kernel/relationship/relationship_test.go`
- `kernel/relationship/relationship_relate_test.go`

## Interfaces introduced or changed

- No public interface changes. `Checker.Has` signature is unchanged; only its
  internal evaluation logic was extended. `Relate` signature is unchanged;
  behavior now requires an actor in ctx and performs upsert/audit.

## Configuration changes

None.

## Schema or migration changes

None. Existing `relationships`, `acting_capacities`, `relationship_types`, and
`audit_logs` tables are used as-is.

## Security changes

- Party-subject ReBAC edges now grant access.
- Unenumerated `subject_kind` fails closed with `KindForbidden`.
- Relationship-edge mutation requires a bound actor and writes an audit row.

## Observability changes

- `Relate` writes an `audit_logs` row with action `relationship.relate`.

## Tests added or modified

- `TestIntegrationRelationshipHasPartySubject` (T1)
- `TestIntegrationRelationshipSubjectKindMatrix` (T2)
- `TestUnitResolveSubjectUnsupportedKind` (T2 fail-closed)
- `TestUnitResolveSubjectCapacityNil` (T2 system actor)
- `TestIntegrationRelateRequiresActor` (T4 ownership)
- `TestIntegrationRelateAttributesAndVersions` (T4 attribution + versioning)
- `TestIntegrationRelateWritesAudit` (T4 audit)
- Existing `TestIntegrationRelateAsPlatformThenHas`,
  `TestIntegrationRelateAsAppRtDenied`, and
  `TestIntegrationRelateTenantIsolation` updated to bind an actor in ctx.

## Commits

To be committed as part of the W03 implementation branch.

## Pull requests

To be opened as part of Wave 03 delivery.

## Implementation dates

- 2026-07-13

## Technical debt introduced

None.

## Known limitations

- Cache-invalidation trigger for relationship-edge mutation is deferred to
  W05-E04-S002 (SEC-04 epoch table). Recorded explicitly in `story.md` and
  `closure.md`.

## Follow-up items

- Independent review (W03-E04-S001-T004).
- Commit and integrate with W03 branch.

## Relationship to the approved plan

Implementation matches `plan.md`. W03-E01 was confirmed `accepted` before
substantive work began. DATA-06 T2's `database.ActorIDFrom` mechanism was
consumed as required; no independent attribution mechanism was reimplemented.
