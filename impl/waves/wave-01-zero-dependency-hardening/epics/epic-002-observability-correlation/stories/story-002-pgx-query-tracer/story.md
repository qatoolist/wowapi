---
id: W01-E02-S002
type: story
title: Pgx query tracer
status: accepted
wave: W01
epic: W01-E02
owner: W01Obs
reviewer: unassigned
priority: P1
created_at: 2026-07-12
updated_at: 2026-07-13
source_requirements:
  - FBL-06
  - D-08
depends_on:
  - W01-E02-S001
blocks: []
acceptance_criteria:
  - AC-W01-E02-S002-01
  - AC-W01-E02-S002-02
artifacts:
  - ART-W01-E02-S002-001
  - ART-W01-E02-S002-002
evidence:
  - EV-W01-E02-S002-001
  - EV-W01-E02-S002-002
  - EV-W01-E02-S002-003
decisions:
  - D-08
---

# W01-E02-S002 — Pgx query tracer

## Story ID

W01-E02-S002

## Title

Pgx query tracer

## Objective

Make database time visible inside a request's trace by attaching a thin, hand-rolled
`pgx.QueryTracer` implementation to the runtime connection pool — one span per query, attached as a
child of whatever parent span (typically the HTTP request span) is active in the query's context.

## Value to the framework

wowapi already has a complete, working distributed-tracing pipeline (HTTP request spans, worker/relay
span propagation across the async boundary) — but database time, which is very often the dominant
cost in a request, is completely invisible in that trace today. An operator diagnosing a slow request
currently cannot tell, from the trace alone, whether the time was spent in application logic or in
Postgres. This closes that gap using only infrastructure the framework already owns (the tracing
port), with no new external dependency.

## Problem statement

`kernel/database.NewPool` (`kernel/database/database.go:128-148`) builds a `*pgxpool.Pool` configured
with `MaxConns` and whatever `Option`s (`kernel/database/database.go:48`, e.g. `WithSetRole`,
`WithConnRLSGuard`) are passed in — but no `pgx.QueryTracer` is ever attached to `pgxpool.Config`.
`pgx`'s `pgxpool.Config` has a `ConnConfig.Tracer` field (a `pgx.QueryTracer`) that, when set, is
invoked around every query with `TraceQueryStart`/`TraceQueryEnd` hooks — this mechanism exists and is
simply unused. Confirmed: no `pgx.QueryTracer` implementation exists anywhere in the repository, and
`otelpgx` (the natural off-the-shelf choice) is not in `go.mod`.

## Source requirements

- FBL-06, task T3 (pgx query tracer), per `impl/analysis/requirement-inventory.md` row FBL-06 and
  MATRIX CS-05's closure-detail spec.
- D-08 — this story **implements** the ratified decision that the tracer must be a thin, hand-rolled
  implementation over the existing `observability.Tracer` port, not the third-party `otelpgx` bridge.
  D-08 is ratified in W00-E02-S003 (a sibling wave's epic/story); this story does not own, author, or
  re-derive that decision — it cites it as an upstream input. See "Assumptions" and `plan.md`
  "Unresolved questions" for how this story tracks the ADR file's non-existence at the time this
  story's planning documents were authored.

## Current-state assessment

Confirmed by direct inspection of the current source tree at planning time:

- `kernel/database/database.go:128-148` (`NewPool`) — `pc.MaxConns = int32(cfg.MaxConns)`; the `for
  _, o := range opts { o(pc) }` loop applies caller-supplied `Option`s (`type Option func(*pgxpool.Config)`,
  `database.go:48`) but no `Option` sets `pc.ConnConfig.Tracer`.
- `chainAfterConnect` (`database.go:113-123`) exists as a helper for composing multiple
  `AfterConnect` hooks — an established pattern this story's `Option` could reuse if the tracer needs
  connection-level setup, though a `QueryTracer` is more likely attached directly to
  `pc.ConnConfig.Tracer` with no `AfterConnect` involvement (confirmed/refined during implementation).
- A complete OTel pipeline exists end-to-end: `adapters/tracing/otel` (adapter + OTLP exporter),
  `kernel/observability.Trace(tr)` (HTTP request spans), and `app/worker.go:79,86` (relay/runner
  tracer propagation across the async boundary via `outbox.WithRelayTracer(k.Tracer)` /
  `jobs.WithRunnerTracer(k.Tracer)`) — so a parent span is reliably present in the `context.Context`
  passed to a query in the common cases (HTTP request handler, worker job).
- No `pgx.QueryTracer` implementation exists in the repository. No `otelpgx` dependency in `go.mod`.

## Desired state

A thin `pgx.QueryTracer` implementation (target ~50 LOC per the epic's sizing note) lives in
`kernel/database`, consuming the `observability.Tracer` port (the same port W01-E02-S001/T001
extends with `TraceID()`/`SpanID()`, though this story's tracer primarily calls `StartSpan`/`End`/
`SetAttr`/`RecordError`, not the ID accessors directly — see `plan.md`). It implements pgx's
`TraceQueryStart(ctx, conn, data) context.Context` / `TraceQueryEnd(ctx, conn, data)` hooks: on
start, it calls `tr.StartSpan(ctx, ...)` (naming the span after the query, e.g. a bounded-cardinality
label, not the raw SQL text which could be high-cardinality or contain literal values) and stores the
resulting `Span` for the matching `TraceQueryEnd` call; on end, it attaches a statement-summary attr
and a rows-affected attr, records an error via `RecordError` if the query failed, and calls `End()`.
Sampling inherits the parent span (no independent sampling decision at the DB layer — a query span is
sampled if and only if its parent request span was sampled, which is the existing OTel adapter's
ratio-sampler behavior, unmodified by this story). The tracer is attached via a new `Option`
(`WithQueryTracer(tr observability.Tracer) Option` or equivalent name, decided at implementation
time) passed to `NewPool` at composition-root time.

## Scope

- The `pgx.QueryTracer` implementation itself (`TraceQueryStart`/`TraceQueryEnd`).
- The `Option` (or equivalent) wiring it into `pgxpool.Config` inside `kernel/database`.
- The trace-fixture test proving a DB span appears as a child of a parent span in an exported trace
  tree.

## Out of scope

- `otelpgx` — explicitly rejected by D-08; see "Problem statement" and `epic.md` "Out of scope."
- Any change to `pgxpool.Config`'s connection-lifecycle settings (`MaxConns`, idle timeouts, etc.) —
  unrelated to this story.
- Any change to the RLS-enforcement check (`database.go:100-111`) or `WithSetRole`/
  `WithConnRLSGuard` — unrelated existing security controls this story must not disturb.
- Sampling-policy changes — inherits parent span sampling unmodified, per "Desired state" above.
- wowsociety's committed `main.go` — additive to the regenerated scaffold only, per `epic.md` scope
  framing (mirrors S001's compatibility posture).

## Assumptions

- **D-08's exact ratified wording is assumed, not confirmed, at the time this story's plan was
  authored.** At planning time, the ADR file expected at
  `impl/waves/wave-00-baseline-and-verification/epics/epic-002-baseline-capture/stories/story-003-adr-ification/decisions/`
  did not yet exist (confirmed: the `decisions/` directory itself does not exist under that story
  path). This story's plan proceeds on the `requirement-inventory.md`/`wave.md` characterization of
  D-08 ("thin in-kernel pgx.QueryTracer via the existing port, NOT otelpgx — a third-party bridge
  would bind OTel vendor types into kernel/database, breaking the port discipline the adapters layer
  gets right") as the working assumption. **This must be confirmed against the actual ratified ADR
  before implementation begins**; if the ratified wording differs, that is recorded as a deviation
  (see `deviations.md`), not silently reconciled.
- The query-naming/labeling strategy (what string identifies a query span — e.g. derived from
  `pgx.TraceQueryStartData.SQL` truncated/hashed to bound cardinality, versus a caller-supplied label)
  is not yet decided; see `plan.md` "Unresolved questions."

## Dependencies

- **Depends on W01-E02-S001 — specifically task T001 only (the `Span` port extension), not task
  T002 (the logging wrapper).** This story's `pgx.QueryTracer` consumes the `observability.Tracer`/
  `Span` port directly to open and close spans; it does not go through the logging pipeline at all.
  The story-level `depends_on: ["W01-E02-S001"]` front-matter field is necessarily story-granular
  (per the mandate's own §6 schema), so this precise task-level scoping is recorded here in prose and
  in `../../dependencies.md` (epic level) to prevent misreading this as a dependency on S001's
  logging-wrapper work.
- **Depends on decision D-08** (ratified in W00-E02-S003) — see "Assumptions" above.
- No dependency on AR-01/AR-02/SEC-01/DATA-09 — this story is leaf work within W01, per the wave's
  own dependency-aware sequencing rationale.

## Affected packages or components

- `kernel/database` — new file for the `pgx.QueryTracer` implementation; `database.go` gains a new
  `Option` (exact file placement — new file vs. addition to `database.go` — decided at
  implementation time).

## Compatibility considerations

Purely additive: a new `Option`, opted into at the composition root (wherever `NewPool` is currently
called), with no default-on behavior change for any caller that does not pass the new option. No
existing `NewPool` call site is required to change. wowsociety impact is additive only (regenerated
scaffold gains it; existing deployment is unaffected unless it explicitly opts in).

## Security considerations

Query-span attrs must not leak sensitive data: the "statement summary" attr must be bounded and must
not include literal parameter values (pgx queries are typically parameterized, so this is largely
already true of `TraceQueryStartData.SQL`, which contains placeholders rather than literal values in
the common case — but this must be explicitly verified during implementation, not assumed, since a
query built via string concatenation rather than parameterization could leak literals into a span
attr). This mirrors the same defense-in-depth discipline `kernel/logging`'s `redactAttr` already
applies to logs — this story's tracer is a second surface with the same sensitivity class and must
not become an unredacted leak channel for what the logging layer already protects.

## Performance considerations

Per-query span creation adds overhead proportional to the tracing adapter's own span-creation cost
(inherited from the existing `Tracer` port — this story adds no new overhead mechanism beyond calling
`StartSpan`/`End` once per query). When no tracing adapter is wired (`NoOpTracer`), the `Option`
simply is not applied (or is applied with `NoOpTracer`, which is already a documented zero-cost path
per `observability.Tracer`'s own doc comment) — no new no-op-path benchmark is required for this
story beyond confirming the existing `NoOpTracer` zero-cost guarantee is not violated by how the
`Option` is wired (a straightforward code-review-level check, not a new formal benchmark acceptance
criterion, since S001 already establishes and proves the no-op-path discipline for the `Tracer`/
`Span` port itself).

## Observability considerations

This entire story is the observability change.

## Migration considerations

None. No schema, data, or config migration is involved.

## Documentation requirements

- Doc comment on the new `pgx.QueryTracer` implementation describing its span-per-query behavior,
  the parent-span-inheritance sampling model, and the security constraint on statement-summary attrs.
- Doc comment on the new `Option` describing how to opt in at the composition root.
- No `docs/blueprint/` changes anticipated as required by this story alone; `implementation.md` will
  record if any were needed.

## Acceptance criteria

- **AC-W01-E02-S002-01** — A pgx query executed inside a `context.Context` carrying an active parent
  span (e.g. an HTTP request span from `Trace(tr)`) produces a child span in the exported trace tree,
  verified via an in-memory span exporter test fixture (not a mock-only assertion) — the query span's
  parent-span ID matches the request span's own ID.
- **AC-W01-E02-S002-02** — The query span carries a statement-summary attr and a rows-affected attr;
  a failed query (e.g. a syntax error or constraint violation in the test fixture) results in the
  span being marked errored via `RecordError`, and the statement-summary attr contains no literal
  parameter values from a parameterized query.

## Required artifacts

- Source of the `pgx.QueryTracer` implementation.
- Diff showing the new `Option` and its wiring into `pgxpool.Config`.
- See `artifacts/index.md` for the registered index (status: not yet produced).

## Required evidence

- Trace-tree export test output showing the pgx child span under its parent.
- Statement-summary/rows-affected/error-marking test output.
- See `evidence/index.md` for the registered index (status: not yet produced).

## Definition of ready

Per `governance/definition-of-ready.md` — this story is ready once: W01-E02-S001-T001 has landed
(the `Span` port extension this story's tracer will build alongside — note: this story's tracer uses
`StartSpan`/`End`/`SetAttr`/`RecordError`, all of which already exist on the port before T001; T001's
`TraceID()`/`SpanID()` accessors are not strictly required by this story's core mechanism, but the
dependency is recorded at the story level per the mandate's task instructions and epic
`dependencies.md`'s framing — treated as a hard prerequisite by this programme's explicit direction,
not merely a nice-to-have), and D-08 is confirmed ratified with wording matching this story's
"Assumptions" section (or a deviation is recorded if not).

## Definition of done

Per `governance/definition-of-done.md` — this story is done once both acceptance criteria are
verified with registered evidence, the query-naming/statement-summary approach is confirmed not to
leak literal parameter values, D-08's ratified wording is confirmed to match (or a deviation is
recorded), and independent review (mandate §14) has passed, specifically checking that no OTel SDK
type leaked into `kernel/database`'s public surface (RISK-W01-E02-003).

## Risks

See `../../risks.md` (epic-level) RISK-W01-E02-001 (D-08 ratification timing) and
RISK-W01-E02-003 (OTel type leakage into `kernel/database`) — both most specific to this story.

## Residual-risk expectations

Once accepted, residual risk is limited to: (a) the query-naming/statement-summary strategy chosen
during implementation becomes the de facto pattern for any future per-query observability work — a
follow-up note in `implementation.md`, not a resolved concern in advance; (b) sampling remains
entirely parent-span-driven, so a query executed with no parent span in context (a background job not
wrapped in `jobs.WithRunnerTracer`, for instance) produces no span at all — this is intended
behavior, not a gap, but is worth recording explicitly so a future reader does not mistake "no span"
for "broken instrumentation" in that specific case.
