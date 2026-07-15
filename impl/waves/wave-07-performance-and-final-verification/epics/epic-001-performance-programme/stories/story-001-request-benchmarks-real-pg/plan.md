---
id: PLAN-W07-E01-S001
type: plan
parent_story: W07-E01-S001
status: implemented
created_at: 2026-07-12
updated_at: 2026-07-14
---

# Plan — W07-E01-S001

Per mandate §8.5. Confirmed facts, planned changes, and assumptions are distinguished explicitly below.

## Proposed architecture

A dedicated Linux amd64 GitHub Actions reference runner (provisional, per DEC-Q9's default) hosting a
DB-backed benchmark suite exercising 6 workload profiles against real Postgres, instrumented for
cost-breakdown attribution, published against a versioned `perf/reference-v1.json` artifact.

## Implementation strategy

1. Stand up the provisional Linux amd64 reference runner (T1).
2. Build `perf/reference-v1.json`'s skeleton, recording the full §14 field list.
3. Build DB-backed benchmarks for all 6 named workload profiles against real Postgres (T2), ensuring no
   RLS guard is weakened.
4. Build the cold/warm cache × 1/10/100-concurrent-tenant variant matrix (T3), with realistic seed data
   for the 100-tenant case.
5. Build cost-breakdown attribution instrumentation, reusing existing tracing/span infrastructure where
   practical (T4).
6. Publish results against `perf/reference-v1.json` (T5), explicitly framing absolute-SLO acceptance as
   conditional on DEC-Q9.

## Expected package or module changes

New: `perf/` (or equivalent) package hosting `reference-v1.json` and its fixtures; new DB-backed
benchmark files; new cost-breakdown instrumentation. New CI infrastructure: the reference runner itself.

## Expected file changes where determinable

- `perf/reference-v1.json` (new).
- New DB-backed benchmark files (exact location TBD).
- New cost-breakdown instrumentation (exact location TBD).
- New CI workflow configuration for the reference runner.

## Contracts and interfaces

`perf/reference-v1.json`'s own schema is the primary new contract — the field list PLAN T1's own
acceptance criterion names.

## Data structures

The `perf/reference-v1.json` schema itself.

## APIs

None affected.

## Configuration changes

None to application configuration; new CI/infrastructure configuration for the reference runner.

## Persistence changes

None to the framework's own schema. Benchmark fixtures may require a seeded dataset for the concurrency-
matrix's 100-tenant case, scoped to the benchmark's own throwaway database.

## Migration strategy

Not applicable.

## Concurrency implications

T3's own concurrency-matrix variants are themselves the concurrency-testing surface this story builds;
no new framework concurrency primitive is introduced.

## Error-handling strategy

Not applicable in the traditional sense — this is benchmark tooling, not a runtime request path.

## Security controls

The RLS-guard-not-weakened constraint (per `story.md` "Security considerations") is a required
implementation discipline throughout T2/T3.

## Observability changes

T4's cost-breakdown attribution reuses existing tracing/span infrastructure where practical.

## Testing strategy

- T1: `perf/reference-v1.json`'s own recorded-field-completeness check.
- T2: DB-backed benchmark execution against real Postgres, confirming no RLS guard weakened.
- T3: concurrency-matrix benchmark execution, minimum 6 combinations per profile.
- T4: cost-breakdown attribution output, confirming separate reporting per component.
- T5: the published comparison report itself, checked for explicit DEC-Q9 conditionality language.

## Regression strategy

Once published, `perf/reference-v1.json` becomes the ongoing baseline this epic's sibling stories (and
future performance work) compare against.

## Compatibility strategy

Not applicable.

## Rollout strategy

T1 lands first (shared prerequisite); T2-T4 may proceed in parallel with each other once T1's reference
environment exists; T5 publishes once T2-T4 are complete.

## Rollback strategy

If the reference runner proves unstable as a comparison baseline (e.g. excessive noise between runs),
diagnose and stabilize per systematic-debugging discipline; do not silently widen acceptance thresholds
to paper over noise without recording why.

## Implementation sequence

T1 → (T2, T3, T4 in parallel) → T5, matching PLAN PERF-02's own dependency chain (T1 for reference
comparison; T2 can be written standalone first; T3 depends on T2; T4 depends on T2; T5 depends on
T1-T4).

## Task breakdown

- **W07-E01-S001-T001** — Reference runner + `perf/reference-v1.json` skeleton (T1).
- **W07-E01-S001-T002** — DB-backed benchmarks, all 6 profiles (T2).
- **W07-E01-S001-T003** — Concurrency-matrix variants (T3).
- **W07-E01-S001-T004** — Cost-breakdown attribution (T4).
- **W07-E01-S001-T005** — Publication against `perf/reference-v1.json`, DEC-Q9-conditional (T5).
- **W07-E01-S001-T006** — Independent review.

## Expected artifacts

`perf/reference-v1.json` + fixtures; the DB-backed benchmark suite; the concurrency-matrix harness; the
cost-breakdown instrumentation; the published comparison report.

## Expected evidence

`perf/reference-v1.json`'s field-completeness confirmation; DB-backed benchmark run output; concurrency-
matrix run output; cost-breakdown attribution output; the published report.

## Unresolved questions

- The exact §14 full field list beyond PLAN T1's own named fields (if directive §14 specifies further
  fields not captured in this generation batch's own source reading) — to be confirmed at implementation
  time against the directive document directly.
- The exact absolute-SLO threshold values — explicitly deferred to DEC-Q9's own resolution, not invented
  here.

## Approval conditions

This plan is approved for implementation once the owner and reviewer are assigned.
