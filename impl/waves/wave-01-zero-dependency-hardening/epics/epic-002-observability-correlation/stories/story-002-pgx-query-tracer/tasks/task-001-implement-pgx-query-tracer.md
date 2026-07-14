---
id: W01-E02-S002-T001
type: task
title: Implement pgx.QueryTracer over the Tracer port
status: done
parent_story: W01-E02-S002
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on:
  - W01-E02-S001-T001
acceptance_criteria:
  - AC-W01-E02-S002-01
  - AC-W01-E02-S002-02
artifacts: []
evidence: []
---

# W01-E02-S002-T001 — Implement pgx.QueryTracer over the Tracer port

## Task Definition

*Per mandate §8.6. This section defines the task before work begins.*

### Task objective

Implement a thin, hand-rolled `pgx.QueryTracer` in `kernel/database` that consumes the
`observability.Tracer` port to open a span per query (as a child of whatever parent span is present
in the query's context), attach statement-summary/rows-affected attrs, mark the span errored on
query failure, and end the span — wired into the pool via a new `Option`. This implements FBL-06 T3
per ratified decision D-08 (thin in-kernel tracer, not `otelpgx`).

A single task carries both the implementation and its wiring and test, per mandate §12's
anti-fragmentation guidance and the epic's sizing note (~50 LOC, one coherent deliverable, one
evidence artifact) — see the parent story's `plan.md` "Implementation strategy" for the full
rationale.

### Parent story

W01-E02-S002 — Pgx query tracer.

### Owner

Unassigned.

### Status

`done` (2026-07-13).

### Dependencies

- W01-E02-S001-T001 (the `Span` port extension) — recorded as a hard prerequisite per this
  programme's explicit task-dependency direction, though this task's core mechanism
  (`StartSpan`/`End`/`SetAttr`/`RecordError`) does not itself require the `TraceID()`/`SpanID()`
  accessors T001 adds; see `../../plan.md` "Confirmed facts" for this nuance.
- D-08 ratification confirmed (from W00-E02-S003) — a blocking pre-implementation confirmation step,
  not a task in this story's own tree (D-08's ADR is owned and authored by W00-E02-S003).

### Detailed work

1. Confirm D-08's ratified wording matches this story's planning assumption (thin in-kernel tracer,
   not `otelpgx`) before writing any code. If it differs, stop and record a deviation rather than
   proceeding against a stale assumption.
2. Implement the `pgx.QueryTracer` type: `TraceQueryStart(ctx, conn, data pgx.TraceQueryStartData)
   context.Context` calls `tr.StartSpan(ctx, <span name>)` and stashes the returned `Span` in the
   returned context (private context key); `TraceQueryEnd(ctx, conn, data pgx.TraceQueryEndData)`
   retrieves the stashed `Span`, sets a statement-summary attr and a rows-affected attr, calls
   `RecordError(data.Err)` if non-nil, and calls `End()`.
3. Confirm the statement-summary attr derivation does not leak literal parameter values from a
   parameterized query (see `../../story.md` "Security considerations").
4. Add the new `Option` (working name `WithQueryTracer`) attaching the tracer to
   `pc.ConnConfig.Tracer`, following the existing `Option` convention in `kernel/database/database.go`
   (`WithSetRole`, `WithConnRLSGuard`).
5. Write the fail-first trace-fixture test: run a query inside a context carrying a parent span
   (test tracer / in-memory span recorder), export the trace tree, confirm no query span exists
   before wiring; confirm the query span appears as a direct child of the parent span after wiring.
6. Write the statement-summary/rows-affected/error test and the literal-leakage negative test (see
   `../../story.md` acceptance criteria).
7. Confirm the existing `kernel/database` test suite (RLS-enforcement check, pool construction,
   existing `Option`s) passes unmodified.

### Expected files or components affected

- New file under `kernel/database/` for the `pgx.QueryTracer` implementation.
- `kernel/database/database.go` (new `Option`, or placed in the new file — decided at implementation
  time).

### Expected output

A wired, tested `pgx.QueryTracer` producing correctly-parented, correctly-attributed query spans,
opted into via an explicit `Option` with zero behavior change for callers that do not pass it.

### Required artifacts

ART-W01-E02-S002-001 (tracer implementation), ART-W01-E02-S002-002 (pool-config wiring diff) — see
`../../artifacts/index.md`.

### Required evidence

EV-W01-E02-S002-001 (trace-tree export test), EV-W01-E02-S002-002 (statement-summary/rows-affected/
error test), EV-W01-E02-S002-003 (literal-leakage negative test) — see `../../evidence/index.md`.

### Related acceptance criteria

AC-W01-E02-S002-01, AC-W01-E02-S002-02 (both proven at this task).

### Completion criteria

Both acceptance criteria pass with registered evidence; D-08's ratified wording is confirmed to
match (or a deviation is recorded); existing `kernel/database` regression suite passes unmodified;
independent review confirms no OTel SDK type leaked into `kernel/database`'s public surface
(RISK-W01-E02-003).

### Verification method

`go test ./kernel/database/...` against a real Postgres instance (per project preference for real
integration tests over mocks crossing the DB process boundary); trace-tree assertions against an
in-memory/test span exporter.

### Risks

RISK-W01-E02-001 (D-08 ratification timing — this task is the direct implementation risk-bearer) and
RISK-W01-E02-003 (OTel type leakage into `kernel/database`) are both most concretely realized at this
task; mitigations per `../../risks.md` (epic level) and step 1 above.

### Rollback or recovery considerations

Revert the commit, or stop passing the new `Option` at the composition root. No persistent state is
created (spans are ephemeral, exported asynchronously); no residual-data rollback concern.

## Implementation Record

Executed 2026-07-13 by W01Obs.

### What was actually implemented

Step 1 (blocking): D-08 ratified wording confirmed against ADR-W00-E02-S003-008 — matches the
planning assumption; no deviation. Steps 2–7 delivered as specified: `queryTracer` +
`WithQueryTracer` in `kernel/database/query_tracer.go`; private-context-key start→end correlation;
`db.<VERB>` span names from a closed verb set; `db.statement` = trimmed/512-byte-truncated SQL as
issued (never `Args`); `db.rows_affected` on success; `RecordError` on failure; nil/NoOpTracer
input leaves pool config untouched. Full narrative: `../implementation.md`.

### Components changed

`kernel/database`.

### Files changed

`kernel/database/query_tracer.go` (new), `kernel/database/query_tracer_test.go` (new).

### Interfaces introduced or changed

`database.WithQueryTracer(tr tracing.Tracer) Option` (type-identical to the planned
`observability.Tracer` signature via S001's alias — see `../deviations.md`).

### Configuration changes

None.

### Schema or migration changes

None.

### Security changes

Bound parameters never reach span attrs (by construction + literal-leakage test).

### Observability changes

This task is the change.

### Tests added or modified

Six real-Postgres integration tests (child-span tree with fail-first before/after pair,
attrs, error marking, literal leakage, parentless root span, sampling inheritance), reusing the
package's `guardTestDSN` skip convention.

### Commits

None — conductor owns commits.

### Pull requests

None.

### Implementation dates

2026-07-13.

### Technical debt introduced

None.

### Known limitations

Query hooks only (no batch/copy/prepare/connect tracing — out of FBL-06 T3 scope, inert absence).

### Follow-up items

Composition-root opt-in in the regenerated scaffold (additive caller work).

### Relationship to the approved plan

D-08 confirmed matching (the field this section was required to record). Architecture as planned;
port imported from `kernel/tracing` per the cross-story deviation (`../deviations.md`).

## Verification Record

*Per mandate §8.8. Table below is planned before execution; fields after it are filled after
execution.*

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W01-E02-S002-01 | Fail-first trace-fixture integration test: query span as child of parent span | Local / CI, real Postgres instance | Fails before wiring (no span), passes after (correct parent-child relationship) | integration-test report | framework architecture lead |
| AC-W01-E02-S002-02 | Statement-summary/rows-affected/error/literal-leakage tests | Local / CI, real Postgres instance | Attrs correct; error-marking correct; no literal leakage | integration-test report | framework architecture lead |

### Actual result

All six integration tests pass with `-race` against real Postgres; full package regression suite
passes unmodified. Per-AC detail in `../verification.md`.

### Pass or fail

Pass (both ACs).

### Evidence identifier

EV-W01-E02-S002-001, -002, -003.

### Execution date

2026-07-13, ~07:29–07:35 UTC.

### Commit or revision

`0a31186cada5c275a588c74081cf977adf346e61` (HEAD) + uncommitted working change.

### Environment

Local dev machine, macOS Darwin 25.5.0 arm64, go1.26.5; Postgres 16-alpine (compose stack).

### Reviewer

Pending — conductor gate (must re-confirm RISK-W01-E02-003: no OTel type in kernel/database).

### Findings

Import-cycle discovery, resolved via `kernel/tracing` leaf package (see deviations).

### Retest status

Final `-race` sweep green after the port relocation.

### Final conclusion

Task complete; both story ACs proven at this task.

## Deviations Record

*Per mandate §8.9. Initially state that deviations are not yet known. The approved plan must not be
silently altered to hide deviations.*

The pre-registered likely deviation (D-08 wording mismatch) did NOT materialize — ADR confirmed
matching before implementation. One reference deviation applies (port imported from
`kernel/tracing`; primary record in S001) — see `../deviations.md`.

### Deviation ID

Not applicable.

### Approved plan

Not applicable.

### Actual implementation

Not applicable.

### Reason

Not applicable.

### Impact

Not applicable.

### Risks

Not applicable.

### Approval

Not applicable.

### Compensating controls

Not applicable.

### Follow-up work

Not applicable.
