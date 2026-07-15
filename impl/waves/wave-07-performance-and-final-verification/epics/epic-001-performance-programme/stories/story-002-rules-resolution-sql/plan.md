---
id: PLAN-W07-E01-S002
type: plan
parent_story: W07-E01-S002
status: implemented
created_at: 2026-07-12
updated_at: 2026-07-14
---

# Plan — W07-E01-S002

Per mandate §8.5. T1's own exact query design cannot be finalized until T0's own audit outcome is known
— this is stated explicitly rather than invented. Confirmed facts, planned changes, and assumptions are
distinguished explicitly below.

## Proposed architecture

A single set-based SQL query (recursive CTE or equivalent, exact shape TBD by T0's own audit outcome and
T1's own design work) replacing the current per-org-ancestor sequential-query loop, preserving exact
precedence semantics, backed by indexes confirmed or added to match both current and historical query
predicates.

## Implementation strategy

1. Run the index-definition audit (T0): `grep "CREATE INDEX" migrations/*rules*.sql` and confirm/refute
   the directive's own indexing claim.
2. Design the set-based query (T1), informed by T0's own outcome, preserving exact
   nearest-ancestor-first → tenant → platform → code-default precedence.
3. Add or confirm indexes matching both current and historical predicates (T2).
4. Produce `EXPLAIN (ANALYZE, BUFFERS)` fixtures at shallow/deep ancestry and low/high history
   cardinality (T3).
5. Write result-parity and SQL-count-constant-with-depth tests (T4).
6. Confirm live per-request rule-update visibility continues passing existing tests (T5).
7. Publish before/after evidence against `perf/reference-v1.json` (T6).

## Expected package or module changes

The rules-resolution query path (exact package TBD, likely `internal/modules/policy/` or
`kernel/policy/`); new or confirmed database indexes on `rule_versions`.

## Expected file changes where determinable

- The rules-resolution query implementation file (exact path TBD).
- New migration(s) for any confirmed-needed index (T2), following DATA-09's own online-migration
  protocol (W02-E01) if the index addition requires `CREATE INDEX CONCURRENTLY` on a live table.
- New `EXPLAIN` fixture files (T3).
- New result-parity and SQL-count test files (T4).

## Contracts and interfaces

None new — the query's own external contract (precedence semantics, return shape) is preserved exactly,
per T1's own requirement.

## Data structures

None new at the schema level beyond any confirmed-needed index (T2).

## APIs

None affected.

## Configuration changes

None.

## Persistence changes

Possible new index(es) on `rule_versions`, per T2's own outcome — additive, not a data-transforming
migration.

## Migration strategy

If T2 requires a new index on a live table, follow DATA-09's own online-migration protocol
(`CREATE INDEX CONCURRENTLY`) per this programme's own established pattern (W02-E01).

## Concurrency implications

None beyond what the existing rules-resolution path already handles.

## Error-handling strategy

The new query must fail the same way the old loop did for any genuinely invalid input — no silent
behavior change beyond the intended performance improvement.

## Security controls

None new — existing tenant/org-scoping semantics are preserved exactly.

## Observability changes

None beyond the `EXPLAIN` fixture evidence itself.

## Testing strategy

- T0: index-definition audit via grep against actual migrations.
- T1: result-parity unit tests against the old per-ancestor-loop implementation.
- T2/T3: `EXPLAIN (ANALYZE, BUFFERS)` fixtures at shallow/deep and low/high cardinality.
- T4: parametrized parity + SQL-count-constant tests across 3/10/50-level ancestries.
- T5: existing rule-update-visibility tests, re-run to confirm no regression.
- T6: before/after comparison against `perf/reference-v1.json`.

## Regression strategy

T4's own SQL-count-constant-with-depth test becomes the ongoing regression guard against a future change
reintroducing the per-ancestor query-count scaling this story eliminates.

## Compatibility strategy

T1's own exact-precedence-preservation requirement is this story's entire compatibility strategy.

## Rollout strategy

T0 → T1 → T2/T3 in parallel → T4 → T5 → T6, matching PLAN PERF-03's own dependency chain.

## Rollback strategy

If the new query is found to diverge from the old loop's precedence semantics after landing, revert to
the old loop and re-diagnose — a precedence-semantics divergence is a correctness defect, not an
acceptable performance-vs-correctness tradeoff.

## Implementation sequence

As listed under "Implementation strategy" above (steps 1–7), matching PLAN PERF-03's own T0→T1→T2→T3→
T4→T5→T6 dependency chain exactly (T0 must precede T2 per PLAN's own explicit note).

## Task breakdown

- **W07-E01-S002-T001** — Index-definition audit (T0).
- **W07-E01-S002-T002** — Set-based query design (T1).
- **W07-E01-S002-T003** — Index confirmation/addition (T2).
- **W07-E01-S002-T004** — `EXPLAIN` fixtures (T3).
- **W07-E01-S002-T005** — Parity + SQL-count tests (T4).
- **W07-E01-S002-T006** — Live-update-visibility regression confirmation (T5).
- **W07-E01-S002-T007** — Publication against `perf/reference-v1.json` (T6).

## Expected artifacts

The index-definition audit report; the set-based query; confirmed/added indexes; `EXPLAIN` fixture
files; the result-parity/SQL-count test suite; the live-update-regression confirmation; the published
before/after comparison.

## Expected evidence

Index-audit output; result-parity unit test output; `EXPLAIN` fixture output; parity-and-SQL-count test
output; live-update-regression test output; the published comparison report.

## Unresolved questions

- T0's own audit outcome (confirm or refute the directive's indexing claim) — genuinely unknown until
  the audit is performed.
- The exact new query's SQL shape — to be determined by T1's own design work, informed by T0's outcome.

## Approval conditions

This plan is approved for implementation once the owner and reviewer are assigned; T1's own design work
must not begin until T0's own audit outcome is known, per PLAN's own explicit "must precede T2" note
(and, by this story's own logical extension, T1's design work too, since T1 informs T2).
