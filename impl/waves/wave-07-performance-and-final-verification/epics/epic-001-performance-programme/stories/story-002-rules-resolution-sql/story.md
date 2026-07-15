---
id: W07-E01-S002
type: story
title: Rules resolution collapsed to bounded SQL — set-based query, index verification, parity proof
status: accepted
wave: W07
epic: W07-E01
owner: W07-Scoping-Dispatch.W07E01S002
reviewer: W05ReviewGateFinal
priority: P1
created_at: 2026-07-12
updated_at: 2026-07-14
source_requirements:
  - PERF-03
depends_on:
  - W07-E01-S001
blocks: []
acceptance_criteria:
  - AC-W07-E01-S002-01
  - AC-W07-E01-S002-02
  - AC-W07-E01-S002-03
  - AC-W07-E01-S002-04
  - AC-W07-E01-S002-05
  - AC-W07-E01-S002-06
artifacts:
  - ART-W07-E01-S002-001
  - ART-W07-E01-S002-002
  - ART-W07-E01-S002-003
  - ART-W07-E01-S002-004
  - ART-W07-E01-S002-005
  - ART-W07-E01-S002-006
evidence:
  - EV-W07-E01-S002-001
  - EV-W07-E01-S002-002
  - EV-W07-E01-S002-003
  - EV-W07-E01-S002-004
  - EV-W07-E01-S002-005
  - EV-W07-E01-S002-006
  - EV-W07-E01-S002-007
decisions: []
risks: []
---

# W07-E01-S002 — Rules resolution collapsed to bounded SQL — set-based query, index verification, parity proof

## Story ID

W07-E01-S002

## Title

Rules resolution collapsed to bounded SQL — set-based query, index verification, parity proof

## Objective

Verify the current `rule_versions` index definitions before designing the new query (T0, a gap-fill
task PLAN itself flags as needing re-confirmation); design and implement one set-based SQL query
replacing the per-org-ancestor sequential-query loop, preserving exact precedence semantics (T1);
confirm indexes match both current and historical predicates (T2); produce `EXPLAIN (ANALYZE, BUFFERS)`
fixtures at representative depth/history cardinality (T3); prove result-parity and SQL-count-constant-
with-depth (T4); preserve live per-request rule updates with no stale-read regression (T5); and publish
before/after evidence against `perf/reference-v1.json` (T6).

## Value to the framework

PLAN's own PERF-03 evidence: "one sequential SQL query per org ancestor, worst case `len(ancestors) + 2`
round trips." This story converts a query cost that scales linearly with organizational depth into one
that is bounded regardless of ancestry depth — directly relevant to any tenant whose organizational
hierarchy grows over time, since today's cost model means a deeper hierarchy silently gets slower rules
resolution with no code change on the tenant's own part. The T0 gap-fill task exists because PLAN's own
evidence flags a genuine, unresolved uncertainty about the *current* state this story must design
against: "**One item flagged as genuinely unverified in this pass** (agent-transport failure truncated
the sub-check): the directive's claim that current `rule_versions` indexing favors active-only lookup
should be re-confirmed against migrations before implementation" — this story does not silently assume
that claim is true, it re-verifies it first.

## Problem statement

PLAN's own PERF-03 evidence, confirmed exactly: "one sequential SQL query per org ancestor, worst case
`len(ancestors) + 2` round trips." PLAN's own T0 task row: "**(Gap-fill)** Verify current `rule_versions`
index definitions before designing the new query | — | Confirm/refute the directive's indexing claim |
`grep \"CREATE INDEX\" migrations/*rules*.sql` | `PERF-03/index-audit.json` | Low but must precede T2."
T1's own acceptance criterion: "Single SQL statement replaces the `for` loop, preserving exact
precedence semantics." T5's own explicit non-regression constraint, quoted directly from PLAN: "Preserve
live per-request rule updates — explicit non-regression constraint ('B13 is not needed for rules')."

## Source requirements

PERF-03 (T0–T6).

## Current-state assessment

Per PLAN's own evidence (to be re-confirmed at this story's own execution commit, and specifically
re-verified by this story's own T0 task, since PLAN's own agent-transport failure left the indexing
claim genuinely unverified): rules resolution issues one sequential SQL query per organizational
ancestor, worst case `len(ancestors) + 2` round trips, following a nearest-ancestor-first → tenant →
platform → code-default precedence order.

## Desired state

`rule_versions`'s current index definitions are confirmed or refuted against the directive's own claim
(T0), before any new query design proceeds. A single set-based SQL statement replaces the `for` loop,
preserving the exact nearest-ancestor-first → tenant → platform → code-default precedence order. Indexes
are confirmed or added matching both current and historical query predicates, shown via `EXPLAIN
(ANALYZE, BUFFERS)` to use index access, not a sequential scan, at both shallow and deep org-ancestry
cardinalities and both low and high historical-version counts. Result parity holds against the old
per-ancestor-loop implementation, and SQL-count-per-request stays constant across 3/10/50-level org
ancestries (not growing with depth). Live per-request rule updates continue to be visible with no
stale-read regression. Before/after evidence is published against `perf/reference-v1.json`.

## Scope

- **T0** — Verify current `rule_versions` index definitions before designing the new query.
- **T1** — Design one set-based query over ancestry + tenant + platform fallback, preserving exact
  precedence semantics.
- **T2** — Add/confirm indexes matching both current and historical predicates.
- **T3** — `EXPLAIN (ANALYZE, BUFFERS)` fixtures at representative depth/history cardinality (shallow
  and deep org ancestries, low and high historical-version counts).
- **T4** — Result-parity + SQL-count-constant-with-depth tests (3/10/50-level ancestries).
- **T5** — Preserve live per-request rule updates — explicit non-regression constraint.
- **T6** — Publish before/after evidence against `perf/reference-v1.json`.

## Out of scope

- **B13** (schema unification / hot-overlay work) — PLAN's own T5 framing explicitly excludes it: "B13
  is not needed for rules." Confirmed a parked P2 backlog item per `requirement-inventory.md` table C
  (DEF-06), not this story's own scope.
- **PERF-02's own reference-environment build** — W07-E01-S001's own scope; this story's T6 consumes
  that environment's `perf/reference-v1.json`, it does not rebuild it.
- **Org-scoped policy for wowsociety** — PLAN's own wowsociety-impact note confirms wowsociety's own
  policy module explicitly excludes org scope ("societies are single-org tenants in E0"), so this
  story's fix is not currently load-bearing for wowsociety, though it becomes so if wowsociety later adds
  org-scoped policy — a forward dependency, not this story's own concern.

## Assumptions

- T0's own re-verification outcome (confirm or refute the directive's indexing claim) is genuinely
  unknown at this planning stage — PLAN's own evidence explicitly flags this as unverified, not a
  confirmed fact this plan can pre-state.
- The exact new query's SQL shape (recursive CTE, window function, or another set-based approach) is not
  specified by any source document beyond "one set-based query" — this story's own T1 design work
  determines the exact approach.

## Dependencies

Depends on W07-E01-S001 (specifically, T6's own consumption of `perf/reference-v1.json`, built by
S001's own T1) — not a hard blocker on this story's other tasks (T0-T5 may proceed independently of
S001), but T6 specifically cannot publish before S001's T1 exists. No dependency on any other wave
beyond the transitive all-prior-waves entry gate.

## Affected packages or components

The rules-resolution query path (exact package, likely within `internal/modules/policy/` or
`kernel/policy/`, per PLAN's own wowsociety-impact note citing `internal/modules/policy/
rulepoints.go:162-168` as the consumer-side reference); new or confirmed database indexes on
`rule_versions`.

## Compatibility considerations

T1's own precedence-semantics-preservation requirement is the primary compatibility constraint: the
new query must produce identical resolution results to the old per-ancestor loop for every existing
precedence scenario — this is a strict behavioral-preservation requirement, not merely a performance
optimization with incidental behavior changes.

## Security considerations

Not directly applicable — this is a query-performance optimization with no new authorization surface;
the existing tenant/org-scoping semantics the query already enforces must be preserved exactly (per T1's
own precedence-preservation requirement).

## Performance considerations

This story IS the performance optimization; see "Objective" and "Desired state" above.

## Observability considerations

Not separately mandated beyond the `EXPLAIN` fixture evidence (T3) itself, which is itself an
observability artifact into the query planner's own behavior.

## Migration considerations

T2's own index additions (if any) are additive schema changes — new indexes, not a data migration. If
any new index requires `CREATE INDEX CONCURRENTLY` on a live table, this story's own implementation
should consider DATA-09's own online-migration protocol (W02-E01) as the appropriate mechanism, though
this is not itself a data migration in the schema-transformation sense DATA-09 was built to protect
against.

## Documentation requirements

Document the new set-based query's own precedence-preservation logic, so a future maintainer
understands why the query is shaped the way it is, not merely that it replaced a loop.

## Acceptance criteria

- **AC-W07-E01-S002-01**: T0's index-definition audit confirms or refutes the directive's own indexing claim
  against actual migrations, before T1's query design proceeds.
- **AC-W07-E01-S002-02**: A single SQL statement replaces the `for` loop, preserving the exact nearest-ancestor-
  first → tenant → platform → code-default precedence order, proven by result-parity unit tests.
- **AC-W07-E01-S002-03**: Indexes matching both current and historical predicates show index access, not a
  sequential scan, via `EXPLAIN (ANALYZE, BUFFERS)`.
- **AC-W07-E01-S002-04**: `EXPLAIN (ANALYZE, BUFFERS)` fixtures are committed for shallow and deep org
  ancestries, and low and high historical-version counts.
- **AC-W07-E01-S002-05**: Result-parity holds against the old implementation; SQL-count-per-request stays
  constant across 3/10/50-level org ancestries.
- **AC-W07-E01-S002-06**: Live per-request rule updates continue passing existing rule-update-visibility tests
  with no stale-read regression; before/after evidence is published against `perf/reference-v1.json`.

## Required artifacts

- The index-definition audit report (T0).
- The set-based rules-resolution query (T1).
- Confirmed/added indexes (T2).
- `EXPLAIN` fixture files (T3).
- Result-parity and SQL-count test suite (T4).
- The live-update-regression test confirmation (T5).
- The published before/after comparison (T6).
See `artifacts/index.md`.

## Required evidence

- Index-audit output (T0).
- Result-parity unit test output (T1).
- `EXPLAIN (ANALYZE, BUFFERS)` fixture output, per cardinality tier (T2, T3).
- Parity-and-SQL-count-constant test output (T4).
- Live-update-regression test output (T5).
- The published comparison report (T6).
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`: `story.md`
and `plan.md` complete, all six acceptance criteria numbered and measurable, dependency on
W07-E01-S001's T6-consumption recorded, owner/reviewer assigned, and T0's own genuinely-uncertain
outcome recorded as an unresolved pre-implementation question rather than pre-assumed.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; all six acceptance
criteria verified with evidence in `evidence/index.md`; `closure.md` completed; independent review
passed per mandate §14, specifically confirming T0's own audit was genuinely performed before T1/T2's
design work began, not skipped or retroactively assumed correct.

## Risks

None recorded at this story's own scope beyond the general risk that T0's audit outcome (if it refutes
the directive's own indexing claim) could require T1/T2's own design to change materially from what
this plan currently anticipates — this is exactly why T0 is sequenced first.

## Residual-risk expectations

Once T0's audit is genuinely performed first and all six acceptance criteria are verified, residual
risk is expected to be low.

## Plan

See `plan.md`.
