---
id: IMPL-W01-E02-S002
type: implementation-record
parent_story: W01-E02-S002
status: recorded
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Implementation record — W01-E02-S002

Implemented 2026-07-13 by W01Obs against HEAD `0a31186cada5c275a588c74081cf977adf346e61`
(conductor owns commits; this record describes the uncommitted working change).

## D-08 ratification confirmation (blocking step 1 of the plan)

Confirmed BEFORE writing code: the ratified ADR exists at
`impl/waves/wave-00-baseline-and-verification/epics/epic-002-baseline-capture/stories/story-003-adr-ification/decisions/adr-008-pgx-query-tracer-not-otelpgx.md`
(status: ratified), and its wording — "thin in-kernel `pgx.QueryTracer` implementation (~50 LOC)
over the existing observability `Tracer` port — NOT `otelpgx`" — matches this story's planning
assumption exactly. No deviation on this axis.

## What was actually implemented

`kernel/database/query_tracer.go` (new file, implementation ~70 LOC + docs):

- `queryTracer{tr}` implements `pgx.QueryTracer`. `TraceQueryStart` opens a span via
  `tr.StartSpan(ctx, "db."+sqlVerb(SQL))` — the name comes from a CLOSED set of leading SQL verbs
  (fallback `db.query`), bounding span-name cardinality by construction — sets `db.statement`
  (whitespace-trimmed SQL as issued, truncated to 512 bytes on a rune boundary; bound `Args` are
  NEVER touched), and stashes the Span in the returned context under a private key.
  `TraceQueryEnd` retrieves it, calls `RecordError(data.Err)` on failure or sets
  `db.rows_affected` on success, and ends the span.
- `WithQueryTracer(tr tracing.Tracer) Option` follows the `WithSetRole`/`WithConnRLSGuard`
  convention: sets `pc.ConnConfig.Tracer`. Nil/`NoOpTracer` input leaves the config untouched,
  preserving the documented zero-cost disabled path (no per-query context allocation when tracing
  is off).
- Sampling: none here — query spans inherit the parent decision via the otel adapter's ParentBased
  sampler, proven by `TestIntegrationQueryTracerInheritsParentSamplingDecision` (unsampled parent →
  zero exported spans).

Resolved plan "unresolved questions": (a) D-08 wording — confirmed, above; (b) span naming /
statement summary — closed verb set for the name, trimmed+truncated parameterized SQL for the
attr (raw-SQL-in-name rejected for cardinality; hashing rejected as operator-hostile);
(c) start→end correlation — private context key (`querySpanKey`), as the plan's working assumption.

## Components changed

`kernel/database` only (production code). The port it consumes is `tracing.Tracer` — type-identical
to `observability.Tracer` via S001's alias (see S001 DEV-W01-E02-S001-001: importing
`kernel/observability` directly from `kernel/database` is an import cycle through httpx/authz;
`kernel/tracing` is the leaf the port now lives in).

## Files changed

- `kernel/database/query_tracer.go` — new (tracer + Option).
- `kernel/database/query_tracer_test.go` — new (six integration tests, real Postgres).

## Interfaces introduced or changed

Exported: `database.WithQueryTracer(tr tracing.Tracer) Option`. No OTel type appears anywhere in
`kernel/database` (verified: `grep 'go.opentelemetry' kernel/database/*.go` excluding tests →
empty; RISK-W01-E02-003 closed).

## Configuration changes

None — programmatic opt-in at the composition root, per plan.

## Schema or migration changes

None.

## Security changes

Statement-summary attr cannot carry bound parameter values by construction (Args never read);
behaviorally guarded by the literal-leakage test (EV-W01-E02-S002-003). Residual (pre-declared):
string-concatenated SQL would surface its literals in `db.statement` — flagged in the Option's doc
comment; no wowapi call path builds SQL that way.

## Observability changes

This story is the observability change.

## Tests added or modified

`kernel/database/query_tracer_test.go`, all against real Postgres (postgres:16-alpine compose
stack; `guardTestDSN` skip-convention reused): child-span trace tree (fail-first pair captured),
statement/rows-affected attrs, error marking, literal leakage, parentless root span, sampling
inheritance. Fail-first: full test file written first with fixtures on plain `NewPool`; captured
failing ("no query span parented under … 1 spans exported"); the only fixture delta afterwards is
adding `database.WithQueryTracer(tr)` — exactly the wiring under test.

## Commits

None — conductor owns commits.

## Pull requests

None.

## Implementation dates

2026-07-13.

## Technical debt introduced

None.

## Known limitations

- Only query tracing (`pgx.QueryTracer`) is implemented — batch/copy/prepare/connect hooks
  (`TraceBatch*` etc.) are out of FBL-06 T3's scope; pgx invokes them only if the tracer implements
  the corresponding optional interfaces, so their absence is inert.
- A query with no parent span produces a root span (or nothing, when unsampled) — intended,
  per the story's residual-risk note (b).

## Follow-up items

Composition-root opt-in (`WithQueryTracer(k.Tracer)` where `NewPool` is called in the regenerated
scaffold) is additive caller work per the story's compatibility posture ("no existing NewPool call
site is required to change"); wowsociety backport optional per `wave.md`.

## Relationship to the approved plan

Matches the plan's architecture verbatim (private-context-key correlation, `WithQueryTracer`
convention, no config gating). One inherited cross-story deviation: the port import is
`kernel/tracing` instead of `kernel/observability` (S001's DEV-W01-E02-S001-001, forced by the
pre-existing import graph; type-identical by alias). Recorded in this story's `deviations.md` as a
reference deviation.
