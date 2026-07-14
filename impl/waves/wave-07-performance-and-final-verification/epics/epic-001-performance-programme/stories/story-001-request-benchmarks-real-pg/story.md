---
id: W07-E01-S001
type: story
title: Request benchmarks against real PostgreSQL — reference environment + DB-backed benchmarks
status: accepted
wave: W07
epic: W07-E01
owner: W07-Phase-A-Execution.W07E01S001
reviewer: W05ReviewGateFinal
priority: P1
created_at: 2026-07-12
updated_at: 2026-07-14
source_requirements:
  - PERF-02
depends_on: []
blocks:
  - W07-E01-S002
  - W07-E01-S003
  - W07-E01-S004
acceptance_criteria:
  - AC-W07-E01-S001-01
  - AC-W07-E01-S001-02
  - AC-W07-E01-S001-03
  - AC-W07-E01-S001-04
  - AC-W07-E01-S001-05
artifacts:
  - ART-W07-E01-S001-001
  - ART-W07-E01-S001-002
  - ART-W07-E01-S001-003
  - ART-W07-E01-S001-004
  - ART-W07-E01-S001-005
evidence:
  - EV-W07-E01-S001-001
  - EV-W07-E01-S001-002
  - EV-W07-E01-S001-003
  - EV-W07-E01-S001-004
  - EV-W07-E01-S001-005
decisions: []
risks: []
---

# W07-E01-S001 — Request benchmarks against real PostgreSQL — reference environment + DB-backed benchmarks

## Story ID

W07-E01-S001

## Title

Request benchmarks against real PostgreSQL — reference environment + DB-backed benchmarks

## Objective

Stand up a dedicated Linux amd64 reference runner and a `perf/reference-v1.json` skeleton (T1 — the
shared prerequisite this epic's own three sibling stories also consume); build DB-backed benchmarks for
public/authenticated-read/authenticated-write/resource-authz/idempotent-write/async-enqueue profiles
against real Postgres, not fakes (T2); exercise cold/warm cache × 1/10/100 concurrent-tenant variants
(T3); attribute cost by pool wait / tx setup / authz query / handler query / serialization / middleware
separately (T4); and publish relative/container results against `perf/reference-v1.json` now, with
**absolute-SLO acceptance explicitly conditional on DEC-Q9** (T5).

## Value to the framework

PLAN's own PERF-02 evidence: "`BenchmarkDispatch`'s own doc comment states auth/authz are exercised via
fakes, no real DB. Every tenant transaction issues 2-4 statements (role bind, optional RLS-enforcement
`pg_roles` check, tenant bind, optional actor bind) before handler code runs." Without real-DB
benchmarks, the framework's own most performance-relevant path — a tenant-scoped request going through
its own RLS/tenant-tx machinery — has never been measured against the infrastructure it actually runs
against in production. MATRIX CS-16's own correction is this story's own governing framing: "PERF-02..05
were framed as wholly blocked on a reference env when only absolute-SLO gating is" — this story unblocks
the *majority* of PERF-02's own value (real measurement, cost attribution, relative comparison) today,
while explicitly not overclaiming the portion (absolute SLO thresholds) that genuinely still depends on
DEC-Q9.

## Problem statement

PLAN's own PERF-02 task table: "T1. Stand up dedicated Linux amd64 reference runner +
`perf/reference-v1.json` skeleton — **shared prerequisite across PERF-02/03/04/05** | — | Artifact
records CPU/runner digest, Go version, Postgres config, pool size, dataset cardinality, tenant
distribution, workload seed, warm-up/measurement durations (§14 full field list) | N/A (infra) |
`perf/reference-v1.json` + fixtures | **High — new CI infrastructure, no owner/timeline established
anywhere in the directive**." T2 through T4 build the actual DB-backed benchmark suite and its cost-
attribution instrumentation; T5's own acceptance criterion states plainly: "Full closure-contract text
satisfied," but is itself marked "**Blocked until T1 exists — this is the finding's actual closure
gate**." REVIEW §12's own correction (cited by MATRIX CS-16) is the resolution this story implements:
build T1 as a provisional, advisory-only Linux amd64 GH-runner-based artifact now, per DEC-Q9's own
provisional default, rather than waiting for a dedicated bare-metal runner decision that has "no owner/
timeline established anywhere in the directive."

## Source requirements

PERF-02 (T1–T5). DEC-Q9 (tracked at epic level, directly constraining this story's own T1/T5).

## Current-state assessment

Per PLAN's own evidence (to be re-confirmed at this story's own execution commit): no `perf/reference-
v1.json` exists. `BenchmarkDispatch`'s own doc comment confirms auth/authz are exercised via fakes, not
a real database. No cost-breakdown instrumentation exists distinguishing pool wait from tx setup from
authz query from handler query from serialization from middleware. No cold/warm cache × concurrent-
tenant benchmark matrix exists.

## Desired state

A `perf/reference-v1.json` skeleton exists, recording CPU/runner digest, Go version, Postgres config,
pool size, dataset cardinality, tenant distribution, workload seed, and warm-up/measurement durations
per the directive's §14 full field list, built against a provisional Linux amd64 GitHub Actions runner
per DEC-Q9's own provisional default (advisory-only initially). DB-backed benchmarks exist for all 6
named workload profiles, using real Postgres, recording p50/p95/p99, allocations, SQL count, bytes, pool
wait, tx duration, lock wait, and plan hash — without weakening any RLS guard to win the benchmark (an
explicit directive prohibition this story's own implementation must honor). Cold/warm cache × 1/10/100-
concurrent-tenant variants (6 minimum combinations per workload profile) are all benchmarked. Cost is
attributed by pool wait / tx setup / authz query / handler query / serialization / middleware
separately — no single aggregate number. Results are published against `perf/reference-v1.json` as
relative/container comparisons now; every absolute-SLO acceptance criterion this story's own AC-05
states is explicitly conditional on DEC-Q9's full resolution, not asserted unconditionally.

## Scope

- **T1** — Stand up a dedicated Linux amd64 reference runner (provisional, per DEC-Q9's default) and a
  `perf/reference-v1.json` skeleton, recording the full §14 field list.
- **T2** — DB-backed benchmarks for public/authenticated-read/authenticated-write/resource-authz/
  idempotent-write/async-enqueue profiles, against real Postgres, recording p50/p95/p99, allocations,
  SQL count, bytes, pool wait, tx duration, lock wait, plan hash — without weakening any RLS guard.
- **T3** — Cold/warm cache × 1/10/100 concurrent-tenant variants, minimum 6 combinations per workload
  profile, with realistic seed data for the 100-tenant case.
- **T4** — Cost-breakdown attribution: pool wait / tx setup / authz query / handler query /
  serialization / middleware, separately, via span-based or `EXPLAIN`-correlated instrumentation.
- **T5** — Publish results against `perf/reference-v1.json` — relative/container comparison now;
  absolute-SLO acceptance explicitly conditional on DEC-Q9.

## Out of scope

- **DEC-Q9's own full resolution** (a dedicated bare-metal runner decision) — a later SRE decision, not
  this story's own scope; T1 builds only the provisional default.
- **PERF-03, PERF-04, PERF-05's own benchmark content** — this epic's sibling stories' own scope; this
  story's own T1 output (`perf/reference-v1.json` and the reference-runner infrastructure) is a shared
  prerequisite those stories consume, not duplicated or re-derived by them.
- **CS-16's own 7-package hot-path benchmark expansion** — W07-E01-S004's own scope; this story's own T2
  benchmarks are complete-request benchmarks, a different scope than CS-16's own package-level
  micro-benchmarks (though both ultimately feed the same `bench-budgets.txt` gate).

## Assumptions

- DEC-Q9's provisional default (a Linux amd64 GitHub Actions runner) is confirmed from REVIEW §F row 9
  as the correct starting point for T1 — this story does not invent a different infrastructure choice.
- The exact §14 full field list for `perf/reference-v1.json` (beyond the named fields: CPU/runner
  digest, Go version, Postgres config, pool size, dataset cardinality, tenant distribution, workload
  seed, warm-up/measurement durations) is confirmed from PLAN's own T1 acceptance criterion as the
  authoritative field list; this story does not add or omit fields beyond what PLAN's own criterion
  names, unless directive §14 itself (not directly reviewed as part of this generation batch) specifies
  additional fields — if so, this is recorded as an implementation-time confirmation step, not invented
  here.
- The exact absolute-SLO threshold values this story's AC-05 defers to DEC-Q9 are, per this wave's own
  task-brief instruction, explicitly NOT invented here — "IF DEC-Q9 resolves to a dedicated bare-metal
  runner, THEN absolute SLO X applies; UNTIL then, relative/container comparison against
  `perf/reference-v1.json` is the acceptance bar."

## Dependencies

None within W07-E01 for this story's own entry (it is the epic's own shared-prerequisite story, per the
epic's own "Internal" dependency note). No dependency on any other wave beyond the transitive all-prior-
waves entry gate. Blocks W07-E01-S002, S003, S004's own publication tasks (each consumes this story's T1
output, `perf/reference-v1.json`, for their own before/after publication).

## Affected packages or components

New: `perf/reference-v1.json` and its supporting fixtures; DB-backed benchmark files (exact location
TBD, likely alongside existing `kernel/*` benchmark files or in a new `perf/` package); cost-breakdown
instrumentation (exact mechanism TBD — span-based or `EXPLAIN`-correlated). New CI infrastructure: the
dedicated Linux amd64 reference runner itself.

## Compatibility considerations

Not applicable — this is additive benchmarking infrastructure with no runtime API surface change.

## Security considerations

PLAN's own explicit prohibition: benchmarks must not weaken any RLS guard to win the benchmark — this is
a required implementation constraint, not optional discipline. Any benchmark that appears to require
weakening RLS to achieve a target number is itself a signal of a real performance problem in the RLS
path, not license to bypass it.

## Performance considerations

This story IS the performance-measurement infrastructure; there is no separate performance concern
beyond the benchmarks' own accuracy and the reference environment's own stability as a comparison
baseline.

## Observability considerations

The cost-breakdown attribution (T4) is itself an observability-adjacent capability — it should reuse the
framework's existing tracing/span infrastructure (per FBL-06's own OTel correlation work, W01-E02) where
practical, rather than building a parallel ad hoc instrumentation mechanism.

## Migration considerations

Not applicable.

## Documentation requirements

Document the reference-environment's own setup (so a future maintainer understands what "the reference
runner" means and how to reproduce results against it); document the `perf/reference-v1.json` schema;
document the cost-breakdown attribution methodology.

## Acceptance criteria

- **AC-W07-E01-S001-01**: `perf/reference-v1.json` exists and records the full §14 field list (CPU/runner
  digest, Go version, Postgres config, pool size, dataset cardinality, tenant distribution, workload
  seed, warm-up/measurement durations), built against the provisional Linux amd64 reference runner.
- **AC-W07-E01-S001-02**: DB-backed benchmarks exist for all 6 named workload profiles against real Postgres
  (not fakes), recording p50/p95/p99, allocations, SQL count, bytes, pool wait, tx duration, lock wait,
  and plan hash, with no RLS guard weakened to achieve a result.
- **AC-W07-E01-S001-03**: Cold/warm cache × 1/10/100-concurrent-tenant variants are benchmarked, minimum 6
  combinations per workload profile, with realistic seed data for the 100-tenant case.
- **AC-W07-E01-S001-04**: Cost is attributed separately by pool wait / tx setup / authz query / handler query /
  serialization / middleware — no single aggregate number is the only reported metric.
- **AC-W07-E01-S001-05**: Results are published against `perf/reference-v1.json` as relative/container
  comparisons. **IF DEC-Q9 resolves to a dedicated bare-metal runner, THEN this story's own absolute-SLO
  acceptance criteria (to be defined at that time) apply; UNTIL then, relative/container comparison
  against `perf/reference-v1.json` is the acceptance bar** — this AC is not satisfied by an
  unconditional absolute-latency claim.

## Required artifacts

- `perf/reference-v1.json` + fixtures (T1).
- DB-backed benchmark suite, all 6 workload profiles (T2).
- Concurrency-matrix benchmark harness (T3).
- Cost-breakdown instrumentation (T4).
- The published relative/container comparison report (T5).
See `artifacts/index.md`.

## Required evidence

- `perf/reference-v1.json`'s own recorded-field-completeness confirmation (T1).
- DB-backed benchmark run output, real Postgres, all 6 profiles (T2).
- Concurrency-matrix benchmark run output, minimum 6 combinations per profile (T3).
- Cost-breakdown attribution output, per-component (T4).
- The published comparison report itself (T5).
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`: `story.md`
and `plan.md` complete, all five acceptance criteria numbered and measurable, no dependency, owner/
reviewer assignment pending, DEC-Q9's own conditionality on AC-W07-E01-S001-05 recorded explicitly rather than
silently assumed resolved.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; all five acceptance
criteria verified with evidence in `evidence/index.md`; `closure.md` completed; independent review
passed per mandate §14, specifically confirming AC-W07-E01-S001-05's own conditional framing is genuinely
honored in this story's written acceptance record, not silently converted into an unconditional claim.

## Risks

None recorded at this story's own scope beyond the epic-level DEC-Q9 risk (RISK-W07-001) and the
epic-level "new CI infrastructure, no owner/timeline established" risk PLAN's own T1 risk note
identifies — both tracked at epic scope, not duplicated here as a separate story-level risk entry.

## Residual-risk expectations

Once T1's provisional-default framing and T5's explicit DEC-Q9 conditionality are both honored, residual
risk is expected to be low for this story's own relative/container-comparison scope.

## Plan

See `plan.md`.
