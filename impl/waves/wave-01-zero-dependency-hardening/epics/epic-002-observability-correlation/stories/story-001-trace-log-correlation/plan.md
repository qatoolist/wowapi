---
id: PLAN-W01-E02-S001
type: plan
parent_story: W01-E02-S001
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
derived: false
---

# Plan — W01-E02-S001 (trace-log-correlation)

Per mandate §8.5: "Do not invent precise code changes where the repository does not yet provide
enough information. Clearly distinguish confirmed facts, planned changes, and implementation
assumptions." This plan is organized to keep that distinction explicit throughout.

## Confirmed facts (verified against live source at planning time, 2026-07-12)

- `kernel/observability/tracing.go:17-39` — `Span` interface has exactly `End()`,
  `SetAttr(key, value string)`, `RecordError(err error)`. No trace/span ID accessor exists.
- `kernel/observability/tracing.go:41-58` — `NoOpTracer`/`noopSpan` implementation: every method is a
  true no-op; `StartSpan` returns `ctx` unchanged.
- `adapters/tracing/otel/otel.go:99-111` — `otelSpan{span oteltrace.Span}` implements `End`, `SetAttr`,
  `RecordError` only; `span.SpanContext()` is never called.
- `kernel/logging/logging.go:85-104` — `New(w, cfg)` builds `slog.NewJSONHandler`/
  `slog.NewTextHandler` directly with only `HandlerOptions{Level, ReplaceAttr: redactAttr}`; no
  wrapping handler exists.
- `kernel/observability/middleware.go:46-62` — `AccessLog(logger)` calls
  `logger.InfoContext(r.Context(), "request", "request_id", ..., "method", ..., "route", ...,
  "status", ..., "dur_ms", ..., "bytes", ...)` — no `trace_id`/`span_id`.
- `kernel/observability/tracing.go:66-86` — `Trace(tr)` middleware: `ctx, span :=
  tr.StartSpan(ctx, "HTTP "+r.Method)`; `rr := r.WithContext(ctx)`; the span-carrying context reaches
  the handler chain (and therefore `AccessLog`, which runs on `rr` per the documented middleware
  ordering "after RequestID (so request id is available) and Recover").
- Repo-wide: zero hits for `SpanContextFromContext` or any OTel log-bridge dependency in `go.mod`.
- `go.mod` already imports `go.opentelemetry.io/otel/trace` (transitively, via
  `adapters/tracing/otel`'s direct dependency on `go.opentelemetry.io/otel/trace` per
  `otel.go:17`'s `oteltrace "go.opentelemetry.io/otel/trace"` import) — so the OTel `trace` package's
  `TraceID`/`SpanID` stringer types are already in the module graph; no new dependency is needed for
  the adapter-side implementation.

## Planned changes

### Proposed architecture

Two additive changes to existing types, plus one new type:

1. Widen `observability.Span` (the port) with two new methods: `TraceID() string`, `SpanID() string`.
   `noopSpan` returns `""` for both. `otelSpan` returns
   `s.span.SpanContext().TraceID().String()` / `s.span.SpanContext().SpanID().String()`.
2. A new `slog.Handler` wrapper type in `kernel/observability` (working name: `CorrelatingHandler` —
   final name determined at implementation time) that: on `Handle(ctx, record)`, checks whether a
   recording span is retrievable from `ctx`; if so and its `TraceID()`/`SpanID()` are non-empty,
   clones the record and appends `trace_id`/`span_id` attrs before delegating to the wrapped handler;
   otherwise delegates unchanged.
3. Wiring: either (a) `logging.New` wraps its constructed handler with the new
   `CorrelatingHandler` before returning `slog.New(h)` — this makes correlation apply to every log
   call in the process, not just `AccessLog` — or (b) `AccessLog` itself is updated to read
   `trace_id`/`span_id` off the span in `r.Context()` and add them as explicit fields, mirroring how
   it already reads `request_id`. Both approaches satisfy the acceptance criteria; approach (a) is
   architecturally preferable (correlation becomes a property of every log call, not just the access
   log line) and is the current planning preference, but the final choice depends on how the "active
   span" is retrieved from `ctx` (see "Unresolved questions" below) and is confirmed during
   implementation, not pre-committed here.

### Implementation strategy

Task-sequenced: T001 lands the port extension first (a small, low-risk, mechanically verifiable
change — extend interface, implement on both types, compile-check no other `Span` implementer
exists). T002 lands the handler wrapper and its wiring, which depends on T001's accessors existing.
This ordering also unblocks W01-E02-S002 as early as possible within this story's own execution,
since S002 only needs T001, not T002.

### Expected package or module changes

- `kernel/observability` — `tracing.go` (Span interface, noopSpan), new file for the handler wrapper.
- `adapters/tracing/otel` — `otel.go` (otelSpan).
- `kernel/logging` — `logging.go`, if wiring approach (a) above is confirmed during implementation.
- `kernel/observability/middleware.go` — if wiring approach (b) is confirmed instead.

### Expected file changes where determinable

- `kernel/observability/tracing.go` — add two methods to the `Span` interface; add two methods to
  `noopSpan`.
- `adapters/tracing/otel/otel.go` — add two methods to `otelSpan`.
- A new file under `kernel/observability/` for the handler wrapper (exact name TBD at implementation
  time — not asserted as fact here).
- Possibly `kernel/logging/logging.go` and/or `kernel/observability/middleware.go`, per the wiring
  decision above.

### Contracts and interfaces

`observability.Span` gains:

```go
// TraceID returns the active span's trace ID as its canonical string form, or ""
// for a no-op span or a span with no valid trace context.
TraceID() string
// SpanID returns the active span's span ID as its canonical string form, or ""
// for a no-op span or a span with no valid trace context.
SpanID() string
```

### Data structures

None new beyond the handler wrapper struct itself (wraps an inner `slog.Handler`; no exported
fields anticipated).

### APIs

No public HTTP/RPC API changes. This is an internal kernel-port and logging-infrastructure change.

### Configuration changes

None anticipated. Correlation is not expected to be config-gated — it should be a structural property
of "was a span active," not a separate on/off flag, since the no-op tracer already provides the
"off" behavior for free (empty string → no attrs).

### Persistence changes

None.

### Migration strategy

Not applicable — no data or schema migration involved.

### Concurrency implications

None beyond what already exists: `slog.Handler.Handle` must remain safe for concurrent use (as all
`slog.Handler` implementations must be, per the stdlib contract) — the wrapper must not introduce
shared mutable state without synchronization. The wrapper is expected to be stateless (pure
delegation plus attribute injection per call), so no new concurrency risk is anticipated.

### Error-handling strategy

`TraceID()`/`SpanID()` return plain strings, never errors — matching the existing `Span` interface's
error-free style (`SetAttr`, `End` are also error-free). The handler wrapper's `Handle` method
follows `slog.Handler`'s existing error-return contract; no new error paths are introduced beyond
what wrapping already implies (delegate errors pass through unchanged).

### Security controls

None new. `redactAttr` in `kernel/logging/logging.go` must continue to apply correctly regardless of
where the new handler wrapper sits in the chain — this is a testing concern (verify redaction still
works with the wrapper present), not a new control.

### Observability changes

This entire story is the observability change.

### Testing strategy

- Unit test: construct a `context.Context` with a real (non-no-op) span active (e.g. via an in-memory
  OTel exporter/tracer test fixture, or the otel adapter's own `Tracer` with a `sdktrace.NewSpanRecorder`-style
  test provider), emit a log record through the wrapped handler, assert `trace_id`/`span_id` attrs
  are present and match the span's actual IDs. This is the fail-first test: written first, confirmed
  to fail against the current (unmodified) handler, then made to pass by the fix.
- Negative-case unit test: emit a log record with `context.Background()` (no span) or with
  `NoOpTracer`'s no-op span active, assert the record's attribute set does NOT contain a `trace_id`
  or `span_id` key at all (not merely that the value is `""`) — this is the specific test RISK-W01-E02-002
  calls out as easy to get wrong.
- Benchmark: `go test -bench` comparing allocations/op for log emission through the new wrapper
  (no-op tracer path) versus the pre-existing plain handler, using `testing.B.ReportAllocs()`.
- Regression: existing tests for `kernel/logging` (redaction, level parsing, format selection) and
  `kernel/observability` (existing `Trace`/`AccessLog` tests) must continue to pass unmodified except
  where wiring changes require an updated call site.

### Regression strategy

Run the full `kernel/logging` and `kernel/observability` package test suites before and after the
change; any newly-failing test is investigated as a regression, not silently adjusted to pass.

### Compatibility strategy

The `Span` interface widening is additive (new methods, no removed/changed methods) — source-breaking
only for a hypothetical external `Span` implementer outside `noopSpan`/`otelSpan`. Confirmed at
planning time: no other package in this repository implements `observability.Span` (only the two
implementations cited above exist). This is reconfirmed by a compile check during implementation
(the whole module must build) rather than left as an assumption.

### Rollout strategy

No rollout mechanism needed beyond normal code deployment — this is not behind a feature flag, since
the behavior is purely additive/observational (extra log attrs) with no risk to existing request
handling logic.

### Rollback strategy

Revert the commit(s). No data/state is created that would need cleanup — correlation attrs are
ephemeral (attached only to in-flight log records), so rollback has no residual-state concern.

### Implementation sequence

1. T001: extend `observability.Span` port + both implementations; compile-check the whole module.
2. T002: implement the ctx-aware handler wrapper; write the fail-first positive-case test (confirm
   it fails pre-wiring); wire it into the logger construction/`AccessLog` path (decision made at this
   point, per "Unresolved questions" below); confirm the positive-case test now passes.
3. T002 continued: write and confirm the negative-case test.
4. T002 continued: write and confirm the allocation-neutrality benchmark.

### Task breakdown

- **W01-E02-S001-T001** — Extend `Span` port with `TraceID()`/`SpanID()`; implement on `noopSpan` and
  `otelSpan`.
- **W01-E02-S001-T002** — Ctx-aware `slog.Handler` wrapper; wire into `AccessLog`/logger construction;
  fail-first positive-case test; negative-case "genuinely absent" test; allocation-neutrality
  benchmark.

### Expected artifacts

- Extended `Span` port interface diff (both port file and both implementation files).
- New handler-wrapper source file.
- Wiring diff (whichever of `logging.go`/`middleware.go` is chosen).

### Expected evidence

- Fail-first test transcript (fails before fix, passes after).
- Negative-case test transcript.
- Benchmark output (before/after allocation comparison).

## Assumptions (flagged for confirmation during implementation, not presumed)

- **How "the active span" is retrieved from `ctx` inside the new handler wrapper** is not yet
  determined. Two candidate approaches: (a) the wowapi `Tracer`/`Span` port could gain a companion
  `SpanFromContext(ctx) (Span, bool)`-style helper on the `Tracer` interface or as a package-level
  function in `kernel/observability`, mirroring the OTel SDK's own `trace.SpanFromContext`; or (b) the
  handler wrapper could be constructed with a reference to the specific `Tracer` in use (so it can
  call a method on that tracer to inspect ctx), rather than being globally reusable across tracer
  instances. This design choice is deferred to implementation time — the story only commits to the
  *outcome* (attrs present when a real span is active, absent when not), not the retrieval mechanism.
  This is exactly the kind of "what must be determined during the story" case mandate §18 calls for
  recording explicitly rather than inventing.
- **Wiring point** (`logging.New` vs. `AccessLog`) — see "Proposed architecture" above; the final
  choice is an implementation-time decision, tentatively leaning toward wrapping at `logging.New` for
  broader coverage, but not committed.
- No assumption is made about D-08 or W00-E02-S003 for this story — S001 has no dependency on that
  decision (only S002 does); this is noted here only to avoid ambiguity given the epic's shared
  framing.

## Unresolved questions

- Exact mechanism for span retrieval from `ctx` (see above) — resolved during T002's implementation,
  recorded in `implementation.md` once decided.
- Final wiring point (`logging.New` vs `AccessLog` vs both) — resolved during T002.
- Exact handler-wrapper type name — cosmetic, resolved during implementation.

## Approval conditions

This plan is considered approved and ready for implementation once: (a) the story's `story.md` status
moves from `planned` to `ready` per `governance/definition-of-ready.md`, and (b) an owner is assigned
(currently `unassigned`). No code has been written against this plan as of this document's creation.
