---
id: IMPL-W04-E04-S001
type: implementation-record
parent_story: W04-E04-S001
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Implementation record — W04-E04-S001

## What was actually implemented

- Widened `kernel/audit.chainHash` to cover every persisted field, including canonicalized
  `metadata` and `tx_id`, using a versioned hash input scheme (`hash_version=1` historical,
  `hash_version=2` widened).
- Added migration `00037_audit_hash_version.sql` with a `hash_version smallint NOT NULL DEFAULT 1`
  column and a W02-E01 manifest block.
- Implemented version-branched verification in `audit.Writer.Verify`: historical rows verify under
  the v1 scheme, new rows under v2, and unknown versions fail closed.
- Added per-field tamper tests (`TestIntegrationAuditChainDetectsPerFieldTampering`) mutating every
  declared field independently, including `metadata` and `tx_id`.
- Added version-branch verification tests (`TestIntegrationAuditHashVersionBranching` and
  `TestIntegrationAuditUnknownHashVersionFailsClosed`).

## Components changed

- `kernel/audit/audit.go` — `chainHash`, `Verify`, `Record`, `Query`, `Log`.
- `kernel/audit/audit_test.go` — per-field tamper and version-branch tests.
- `migrations/00037_audit_hash_version.sql` — new column migration with manifest.

## Files changed

- `kernel/audit/audit.go`
- `kernel/audit/audit_test.go`
- `migrations/00037_audit_hash_version.sql` (new)

## Interfaces introduced or changed

- `func chainHash(version int16, prev string, seq int64, id uuid.UUID, occurredAt time.Time, txID string, metadata []byte, fields ...string) string` — internal; version selects scheme.
- `func canonicalizeMetadata(m map[string]any) ([]byte, error)` — internal; deterministic JSON for hashing.
- `func recomputeRowHash(...)` — internal; version-branch dispatcher.
- `audit.Log` — added `Metadata map[string]any` and `HashVersion int16` fields.

## Configuration changes

None.

## Schema or migration changes

- `migrations/00037_audit_hash_version.sql` adds `hash_version smallint NOT NULL DEFAULT 1` to
  `audit_logs`. Manifest classification: online, no backfill, lock budget 2000 ms.

## Security changes

- Tamper-evidence now covers `metadata` and `tx_id`; any field-level mutation breaks verification.

## Observability changes

- `Verify` failure reasons now include the row's `hash_version` so operators know which scheme was
  in play.

## Tests added or modified

- `TestIntegrationAuditChainDetectsPerFieldTampering` — 17 subtests, one per persisted field.
- `TestIntegrationAuditHashVersionBranching` — v1 historical + v2 new rows in one chain.
- `TestIntegrationAuditUnknownHashVersionFailsClosed` — unrecognized `hash_version` fails.

## Implementation dates

2026-07-13

## Technical debt introduced

None.

## Known limitations

None.

## Follow-up items

- wowsociety staging audit re-verification drill (PROD-05) is tracked as a product-side coordination
  item, not framework-side scope.

## Relationship to the approved plan

Matches `plan.md`. D-04 enacted exactly: `hash_version=2` chosen for the new scheme, historical rows
verify under v1, new rows under v2.
