---
id: IMPL-W07-E01-S002
type: implementation-record
parent_story: W07-E01-S002
status: complete
created_at: 2026-07-12
updated_at: 2026-07-14
---

# Implementation record — W07-E01-S002

## What was actually implemented

- Audited the actual `rule_versions` migration before query design.
- Replaced the Go per-ancestor lookup loop with one indexed set-based statement preserving
  nearest-ancestor → tenant → platform → code-default precedence and historical-window behavior.
- Confirmed the current active-only exclusion index and added an online historical-resolution index.
- Added real-PostgreSQL parity, query-count, live-update, catalog, and EXPLAIN coverage.
- Generated four EXPLAIN fixtures and a DEC-Q9-honest relative comparison.

## Components changed

`kernel/rules`, `migrations`, `perf/results`, and this story's artifact/evidence/lifecycle records.

## Files changed

- `kernel/rules/resolver.go`
- `kernel/rules/resolver_perf_test.go`
- `migrations/00048_rule_versions_resolution_indexes.sql`
- `migrations/rules_resolution_indexes_test.go`
- `migrations/migrations_test.go`
- `perf/results/perf-03-*.json`
- W07-E01-S002 artifact, evidence, task, verification, and closure records

## Interfaces introduced or changed

None. `Resolver.Resolve` and every public contract remain unchanged.

## Configuration changes

None.

## Schema or migration changes

`00048_rule_versions_resolution_indexes.sql` adds
`rule_versions_history_resolution_idx` concurrently and removes the obsolete narrower
`rule_versions_lookup`; its Down restores the original lookup before dropping the new index.

## Security changes

None. The statement still runs on the caller's `TenantDB`; existing RLS remains the tenant boundary.

## Observability changes

Four committed real `EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON)` fixtures.

## Tests added or modified

- Legacy/set-based parity for six precedence/history cases.
- Real query counts at ancestry depths 3/10/50.
- Four index-plan/cardinality fixtures.
- Migration catalog/index-contract test.
- Existing live-update and full `kernel/rules` focused regression suite rerun with DB required.

## Commits

Working-tree implementation based on `733ef3e930cbb3f89f5bbc53d8f562c60e426513`; evidence records
state this explicitly and do not represent the base SHA as containing the changes.

## Pull requests

None.

## Implementation dates

2026-07-14.

## Technical debt introduced

None identified.

## Known limitations

DEC-Q9 is open. Local container timing observations are not absolute-SLO or accepted linux/amd64
reference-runner claims.

## Follow-up items

None within story scope. B13/schema unification remains intentionally out of scope.

## Relationship to the approved plan

Matched T0–T6. The plan intentionally left the exact SQL shape and T0 outcome unresolved; execution
confirmed active-only current indexing and selected indexed LATERAL set evaluation after that audit.
