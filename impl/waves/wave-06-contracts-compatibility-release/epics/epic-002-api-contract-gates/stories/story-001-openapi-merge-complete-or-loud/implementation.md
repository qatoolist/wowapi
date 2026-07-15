---
id: IMPL-W06-E02-S001
type: implementation-record
parent_story: W06-E02-S001
status: verified
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Implementation record — W06-E02-S001

Complete-or-loud OpenAPI 3.1.1 merge, structural validation, and semantic compatibility
are implemented and focused-test verified. Independent review remains pending.

## What was actually implemented

- Typed merge policies cover all OpenAPI 3.1 top-level and `components.*` fields: identical-required singletons, keyed unions, stable-deduplicated servers/security, and conflict rejection.
- The merged document is forced to OpenAPI 3.1.1 with JSON Schema 2020-12 dialect, then validated through pinned `libopenapi-validator`.
- `wowapi openapi diff` uses `libopenapi.CompareDocuments` and fails each change classified breaking under the repository's v1 policy.

## Components changed

`internal/cli` OpenAPI command/merge/diff; root module dependencies; compatibility CI workflow and caller.

## Files changed

See `artifacts/index.md`; production changes are in `internal/cli/openapi_cmd.go`, `openapi_merge.go`, and `openapi_diff.go`.

## Interfaces introduced or changed

CLI adds `wowapi openapi diff --baseline FILE --current FILE`; merge output is now strictly OpenAPI 3.1.1.

## Configuration changes

No runtime configuration changes.

## Schema or migration changes

No database migration changes; generated OpenAPI JSON Schema uses explicit draft 2020-12 dialect.

## Security changes

Pinned validator dependencies were approved after the dated licence/advisory review in `evidence/security/validator-dependency-review.md`.

## Observability changes

None.

## Tests added or modified

Fixture-driven tests cover every required field policy, malformed structural objects, additive changes, required request additions, response removal, and security weakening.

## Commits

No commit created; shared dirty workspace based on `733ef3e`.

## Pull requests

None.

## Implementation dates

2026-07-13.

## Technical debt introduced

None identified.

## Known limitations

Hosted CI evidence remains pending; independent review passed with no open issues.

## Follow-up items

Register hosted CI evidence and acceptance authority disposition.

## Relationship to the approved plan

Implementation matched the approved plan's full-field, structural-validation, and semantic-diff architecture; validator choice is now resolved by the recorded decision.
