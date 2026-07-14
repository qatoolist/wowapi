---
id: PLAN-W01-E02-S002
type: plan
parent_story: W01-E02-S002
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
derived: false
---

# Plan — W01-E02-S002 (pgx-query-tracer)

Per mandate §8.5: "Do not invent precise code changes where the repository does not yet provide
enough information. Clearly distinguish confirmed facts, planned changes, and implementation
assumptions." This plan is organized to keep that distinction explicit throughout.

## Confirmed facts (verified against live source at planning time, 2026-07-12)

- `kernel/database/database.go:128-148` (`NewPool`) — sets `pc.MaxConns` and applies caller-supplied
  `Option`s via `for _, o := range opts { o(pc) }`. No `Option` currently sets any tracer.
- `kernel/database/database.go:48` — `type Option func(*pgxpool.Config)`. Confirmed pattern already
  in use for `WithSetRole` (`database.go:57`) and `WithConnRLSGuard` (`database.go:72`) — a new
  `Option` for the query tracer follows this exact established convention.
- `kernel/database/database.go:113-123` — `chainAfterConnect` exists for composing `AfterConnect`
  hooks; likely not needed for a `QueryTracer` (which attaches to `pc.ConnConfig.Tracer`, not
  `AfterConnect`), but confirmed present as a related pattern in the same file.
- No `pgx.QueryTracer` implementation exists anywhere in the repository (repo-wide search
  confirmed). No `otelpgx` dependency in `go.mod`.
- `observability.Tracer`/`Span` port (`kernel/observability/tracing.go:17-39`) already provides
  `StartSpan(ctx, name) (context.Context, Span)`, `Span.End()`, `Span.SetAttr(key, value string)`,
  `Span.RecordError(err error)` — all methods this story's tracer needs are already present on the
  port *before* W01-E02-S001-T001 lands; T001 adds `TraceID()`/`SpanID()`, which this story's core
  mechanism does not strictly require (see `story.md` "Definition of ready" for why the dependency is
  still recorded as a hard prerequisite per this programme's explicit direction).
- `app/worker.go:79,86` confirms the pattern of passing `k.Tracer` (the process's configured
  `observability.Tracer`) into a component's constructor via a `With*Tracer` functional option
  (`outbox.WithRelayTracer(k.Tracer)`, `jobs.WithRunnerTracer(k.Tracer)`) — this story's
  `WithQueryTracer`-style `Option` follows the same established naming/wiring convention.
- The `pgx` module (already a direct dependency, since `kernel/database` already imports
  `github.com/jackc/pgx/v5/pgxpool`) defines the `pgx.QueryTracer` interface
  (`TraceQueryStart(ctx, conn, data TraceQueryStartData) context.Context` /
  `TraceQueryEnd(ctx, conn, data TraceQueryEndData)`) — no new module dependency is needed to
  implement it; the interface is already available from the existing `pgx` import.
- D-08's characterization in `requirement-inventory.md` row FBL-06 and `wave.md`'s "Assumptions"
  section: "D-08 (pgx query tracing approach) is ratified by W00-E02-S003 before W01-E02-S002
  begins; if not yet ratified, W01-E02-S002 documents this as a blocking gap rather than
  re-deciding it" — and the epic mandate's own framing: "hand-rolled thin tracer via the existing
  port, NOT otelpgx — a third-party bridge would bind OTel vendor types into kernel/database,
  breaking the port discipline the adapters layer gets right." At planning time, the ADR file itself
  (expected under W00-E02-S003's `decisions/` directory) did not yet exist — confirmed by directory
  listing at planning time.

## Planned changes

### Proposed architecture

A new type implementing `pgx.QueryTracer`, constructed with an `observability.Tracer`, living in
`kernel/database`. On `TraceQueryStart`, it calls `tr.StartSpan(ctx, <query span name>)` and stashes
the returned `Span` in the returned `context.Context` (via a private context key, or via pgx's own
`TraceQueryStartData`/`TraceQueryEndData` correlation mechanism — pgx guarantees `TraceQueryEnd`
receives the `context.Context` returned by `TraceQueryStart`, so a private context key is the
straightforward correlation mechanism). On `TraceQueryEnd`, it retrieves the stashed `Span`, sets
`SetAttr` for a statement summary and rows-affected count, calls `RecordError` if
`TraceQueryEndData.Err != nil`, and calls `End()`.

Attached to the pool via a new `Option`, e.g.:

```go
func WithQueryTracer(tr observability.Tracer) Option {
    return func(pc *pgxpool.Config) {
        pc.ConnConfig.Tracer = newQueryTracer(tr)
    }
}
```

(Exact naming confirmed at implementation time; `WithQueryTracer` is the working name consistent
with the `With*Tracer` convention already established in `app/worker.go`.)

### Implementation strategy

Single task (T001) — per mandate §12's anti-fragmentation guidance and the epic's own sizing note
("~50 LOC... a single task is appropriate here... this is one coherent deliverable with one evidence
artifact"). Splitting the tracer implementation from its wiring, or from its test, would produce
tasks too small to independently track or verify — the tracer has no meaning without being wired,
and the wiring has no meaning without the tracer, and both are proven by the same trace-tree-export
test. This mirrors mandate §12's own guidance: "avoid excessive fragmentation into trivial tasks that
provide no tracking value."

### Expected package or module changes

- `kernel/database` only.

### Expected file changes where determinable

- New file under `kernel/database/` for the `pgx.QueryTracer` implementation (exact name TBD —
  e.g. `tracing.go` or `query_tracer.go` — not yet created, so not asserted as fact).
- `kernel/database/database.go` — new `Option` function (or placed in the new file instead; either
  is consistent with existing conventions in the package).

### Contracts and interfaces

Implements `pgx.QueryTracer` (an interface pgx already defines, not one this story defines). Consumes
`observability.Tracer`/`observability.Span` (existing ports, unmodified by this story beyond the
T001 accessor addition this story does not directly require for its core mechanism).

### Data structures

A struct holding the `observability.Tracer` reference; no exported fields anticipated.

### APIs

No public HTTP/RPC API changes. Internal kernel-package addition only.

### Configuration changes

None. The tracer is opted into programmatically via the new `Option`, not via `config.DB` fields —
consistent with how `WithSetRole`/`WithConnRLSGuard` are also programmatic opt-ins rather than config
fields.

### Persistence changes

None.

### Migration strategy

Not applicable.

### Concurrency implications

`pgx.QueryTracer` methods are called concurrently across connections in the pool — the
implementation must be safe for concurrent use. Since state is correlated via `context.Context`
(a new context per query, not shared mutable state on the tracer struct itself), this is expected to
be safe by construction, but is explicitly verified (not merely assumed) via the trace-fixture test
exercising concurrent queries if practical, or at minimum via code review confirming no shared
mutable state exists on the tracer struct outside the injected `Tracer` reference (which is itself
expected to be safe for concurrent use, per the existing `observability.Tracer` port's usage
elsewhere in the codebase, e.g. shared across HTTP handlers already).

### Error-handling strategy

A failed query (`TraceQueryEndData.Err != nil`) results in `Span.RecordError(err)` — mirroring the
existing pattern in `kernel/observability/tracing.go`'s `Trace` middleware, which does not currently
call `RecordError` itself but establishes `RecordError` as the port's error-marking mechanism. No new
error type or error-handling contract is introduced; this story's tracer never returns an error
itself (pgx's `TraceQueryStart`/`TraceQueryEnd` signatures do not return errors).

### Security controls

The statement-summary attr must not include literal parameter values. `pgx.TraceQueryStartData.SQL`
holds the query text as issued (with `$1`/`$2`-style placeholders for parameterized queries, not
literal values, in the standard parameterized-query path) — this story's implementation must confirm
(via a test using a parameterized query with a sensitive-looking literal argument, e.g. a fake
password string passed as a bound parameter) that the attr never contains the literal value. If any
call site in the codebase issues raw string-concatenated SQL (a pattern this story does not introduce
but must not silently trust), the statement-summary attr could leak that literal — this is flagged as
a known residual risk to check during implementation, not resolved in advance, since confirming "no
call site does this" is a broader audit than this story's scope.

### Observability changes

This entire story is the observability change.

### Testing strategy

- Fail-first fixture test: run a pgx query inside a `context.Context` carrying a parent span
  produced by an in-memory/test tracer setup (e.g. the otel adapter wired with an in-memory span
  recorder, or a lightweight test double satisfying `observability.Tracer` directly — decided at
  implementation time), assert the exported trace tree contains a query span whose parent-span ID
  matches the request span's ID. Confirm this test fails before the `Option`/tracer exists (no DB
  span appears) and passes after.
- Statement-summary/rows-affected/error test: assert the query span carries the expected attrs; run
  a deliberately failing query (e.g. malformed SQL against a real or fixture Postgres instance,
  consistent with the project's stated preference for real integration tests over mocks where a
  local DB is available) and assert `RecordError` was invoked (span marked errored).
- Literal-leakage test: parameterized query with a sensitive-looking bound parameter; assert the
  statement-summary attr does not contain that literal value.
- Regression: existing `kernel/database` test suite (RLS-enforcement check, pool construction,
  existing `Option`s) must continue to pass unmodified.

### Regression strategy

Run the full `kernel/database` package test suite before and after the change against a real
Postgres instance (per this repository's stated preference for real integration tests over mocks
crossing a process boundary — a DB is exactly that kind of boundary).

### Compatibility strategy

Purely additive `Option` — no existing `NewPool` call site changes behavior unless it explicitly
passes the new option. No compatibility risk to existing callers.

### Rollout strategy

No feature flag needed — opt-in via explicit `Option` at the composition root. A deployment that does
not pass `WithQueryTracer` sees no behavior change at all.

### Rollback strategy

Revert the commit, or simply stop passing the `Option` at the composition root. No persistent state
is created (spans are ephemeral, exported asynchronously by the existing OTel batch exporter) — no
residual-data rollback concern.

### Implementation sequence

1. Confirm D-08's ratified wording matches this plan's assumption (see "Unresolved questions") —
   blocking step before any code is written.
2. Confirm W01-E02-S001-T001 has landed (per "Definition of ready").
3. Implement the `pgx.QueryTracer` type and the `Option`.
4. Write the fail-first trace-tree-export test; confirm it fails pre-wiring, passes post-wiring.
5. Write the statement-summary/rows-affected/error test.
6. Write the literal-leakage test.
7. Confirm regression suite passes.

### Task breakdown

- **W01-E02-S002-T001** — Implement `pgx.QueryTracer` over the `Tracer` port; wire via new `Option`
  into pool config; trace-fixture child-span test (covers both the implementation and its proof, per
  the single-task rationale above).

### Expected artifacts

- `pgx.QueryTracer` implementation source.
- `Option` wiring diff.

### Expected evidence

- Trace-tree export test output (child span under parent, before/after fail-first comparison).
- Statement-summary/rows-affected/error-marking test output.
- Literal-leakage negative test output.

## Unresolved questions

- **D-08's exact ratified wording** — not confirmed at planning time (ADR file did not exist yet).
  This plan proceeds on the working assumption stated in "Confirmed facts" above, sourced from
  `requirement-inventory.md` and `wave.md` (both of which predate and characterize D-08 consistently
  with this plan's approach) — but this is explicitly an assumption to reconfirm against the actual
  ratified ADR before implementation begins, not a confirmed fact. If W00-E02-S003 has completed and
  produced a different design by the time this story starts, a deviation is recorded rather than this
  plan being silently rewritten to match.
- **Query span naming / statement-summary derivation** — whether to use the raw (parameterized)
  `TraceQueryStartData.SQL` truncated to a bounded length, a normalized/hashed form, or a
  caller-supplied label, is not yet decided. Resolved during implementation; recorded in
  `implementation.md`.
- **Context-key correlation mechanism** between `TraceQueryStart` and `TraceQueryEnd` — private
  context key is the working assumption (see "Proposed architecture") but the exact key type/name is
  an implementation detail, not committed here.

## Approval conditions

This plan is considered approved and ready for implementation once: (a) D-08's ratified wording is
confirmed to match this plan's working assumption (or a deviation is pre-recorded if it does not),
(b) W01-E02-S001-T001 has landed, and (c) the story's `story.md` status moves from `planned` to
`ready` per `governance/definition-of-ready.md`. No code has been written against this plan as of
this document's creation.
