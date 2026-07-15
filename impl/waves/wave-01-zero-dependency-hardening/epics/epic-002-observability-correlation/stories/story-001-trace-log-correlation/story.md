---
id: W01-E02-S001
type: story
title: Trace/log correlation
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
depends_on: []
blocks:
  - W01-E02-S002
acceptance_criteria:
  - AC-W01-E02-S001-01
  - AC-W01-E02-S001-02
  - AC-W01-E02-S001-03
artifacts:
  - ART-W01-E02-S001-001
  - ART-W01-E02-S001-002
  - ART-W01-E02-S001-003
  - ART-W01-E02-S001-004
evidence:
  - EV-W01-E02-S001-001
  - EV-W01-E02-S001-002
  - EV-W01-E02-S001-003
decisions: []
risks: []
---

# W01-E02-S001 — Trace/log correlation

## Story ID

W01-E02-S001

## Title

Trace/log correlation

## Objective

Make it possible to join a log record to the distributed trace it was emitted inside, by (1) giving
the `observability.Span` port a vendor-neutral way to expose the active trace/span IDs and (2)
wiring a context-aware log handler that attaches `trace_id`/`span_id` attrs to every log record
emitted while a recording span is active in the request context.

## Value to the framework

Every wowapi process already emits structured logs and, when a tracing adapter is wired, already
opens a real distributed trace per request. Today these two signals are unjoinable: an operator
reading a log line has no way to pivot to the trace that produced it, and an operator reading a trace
has no way to pull the log lines emitted during it. This is pure operational leverage on
infrastructure the framework already owns — no new subsystem, just the missing connective tissue
between two existing ones. It is generic to any product built on wowapi, not specific to wowsociety.

## Problem statement

`kernel/logging.New` (`kernel/logging/logging.go:85-104`) builds a plain `*slog.Logger` with a
`ReplaceAttr` hook for secret redaction, but no handler in the chain is context-aware: nothing pulls
`trace.SpanContextFromContext(ctx)` (or an equivalent) to attach trace/span identifiers to a record.
`AccessLog` (`kernel/observability/middleware.go:46-62`) calls `logger.InfoContext(r.Context(), ...)`
— so a `context.Context` carrying the request's span *is* available at every logging call site that
uses `InfoContext`/`ErrorContext`/etc. — but nothing in the handler chain reads it for trace/span IDs;
`AccessLog` only extracts `request_id` via `httpx.RequestIDFrom`.

The blocking gap is one layer down: the `observability.Span` port
(`kernel/observability/tracing.go:17-39`) exposes `End()`, `SetAttr(key, value string)`, and
`RecordError(err error)` — there is no `TraceID()`/`SpanID()` accessor at all. The otel adapter's
`otelSpan` (`adapters/tracing/otel/otel.go:99-111`) wraps a real `oteltrace.Span` internally but never
exposes its `SpanContext()` through the port surface, so even if a log handler wanted to read the
active span's IDs, there is no vendor-neutral way to ask for them.

## Source requirements

- FBL-06, tasks T1 (port extension) and T2 (log correlation wiring), per
  `impl/analysis/requirement-inventory.md` row FBL-06 and MATRIX CS-05's closure-detail spec.

## Current-state assessment

Confirmed by direct inspection of the current source tree at planning time:

- `kernel/observability/tracing.go:17-39` — the `Span` interface has exactly three methods: `End()`,
  `SetAttr(key, value string)`, `RecordError(err error)`. No trace/span ID accessor exists.
- `adapters/tracing/otel/otel.go:99-111` — `otelSpan` wraps `span oteltrace.Span` but its exported
  methods (`End`, `SetAttr`, `RecordError`) do not touch `span.SpanContext()`.
- `kernel/logging/logging.go:85-104` — `New` builds `slog.NewJSONHandler`/`slog.NewTextHandler`
  directly with only a `ReplaceAttr` option; no wrapping handler, no context awareness beyond what
  `slog.Logger.InfoContext` already passes through unused.
- `kernel/observability/middleware.go:46-62` — `AccessLog` uses `logger.InfoContext(r.Context(), ...)`
  and attaches `request_id`, `method`, `route`, `status`, `dur_ms`, `bytes` — never `trace_id`/
  `span_id`.
- Repo-wide search confirms zero call sites of `trace.SpanContextFromContext` (or any OTel
  `SpanContextFromContext` equivalent) — the correlation join genuinely does not exist anywhere in
  the codebase today, not merely in these two files.
- No `otelslog` (or similar OTel-log-bridge) dependency exists in `go.mod` — and per this story's
  explicit scope boundary, none will be added; the fix is attribute injection into existing `slog`
  records, not a new OTLP log-export pipeline.

## Desired state

- `observability.Span` gains `TraceID() string` and `SpanID() string` methods. `NoOpTracer`'s
  `noopSpan` returns `""` for both (so no-op remains zero-cost and callers never need a nil check).
  The otel adapter's `otelSpan` returns `span.SpanContext().TraceID().String()` and
  `.SpanID().String()` respectively.
- A new ctx-aware `slog.Handler` wrapper lives in `kernel/observability` (not `kernel/logging`, to
  keep the OTel-adjacent concern colocated with the tracing port it depends on — see "Affected
  packages" below for the exact placement rationale). It reads the active `observability.Span` out of
  context (via whatever context-carrying mechanism `StartSpan`/`Trace` already use to store it —
  confirmed at implementation time, not invented here) and, when a recording span is present, injects
  `trace_id`/`span_id` attrs into the record before delegating to the wrapped handler. When no span
  is present, the wrapper is a pass-through with no injected attrs.
- `AccessLog` and the process's handler-logger construction path are wired to use this wrapper (or
  `AccessLog` itself is updated to add the same two attrs directly — the exact placement is an
  implementation-sequence decision recorded in `plan.md`, not pre-decided here per mandate §8.5's
  "do not invent precise code changes" instruction).

## Scope

- Extending the `observability.Span` port interface with `TraceID()`/`SpanID()` and both
  implementations (`noopSpan`, `otelSpan`).
- A ctx-aware `slog.Handler` wrapper in `kernel/observability`.
- Wiring that wrapper (or equivalent attribute injection) into `AccessLog` and the handler-logger
  construction path used by processes (api/worker/migrate) so it actually takes effect at runtime,
  not merely exists as an unused type.
- The fail-first test proving correlation attrs are present when a span is active.
- The negative-case test proving correlation attrs are genuinely absent (key not present) when no
  span is active.
- The allocation-neutrality benchmark for the no-op tracer path.

## Out of scope

- Any OTel log-export bridge (`otelslog` or equivalent) — explicitly rejected; see `epic.md`
  "Out of scope."
- Changes to `redactAttr`/secret redaction logic in `kernel/logging/logging.go` beyond what is
  needed to coexist with the new handler wrapper (the wrapper must not interfere with existing
  redaction — this is a compatibility constraint, not new scope).
- Any change to the OTel SDK wiring, exporter configuration, or sampling policy in
  `adapters/tracing/otel` beyond exposing `SpanContext()`.
- wowsociety's committed `main.go` — this story affects the regenerated scaffold and library code
  only; backporting to wowsociety's hand-edited entrypoint is optional per `wave.md`'s framing of
  FBL-06 as "additive; existing main.go backports optionally (not required)."

## Assumptions

- The mechanism by which `Trace(tr)` middleware's active span becomes reachable from a
  `context.Context` at an arbitrary downstream logging call site (i.e., how the wrapper will locate
  "the active span" to read `TraceID()`/`SpanID()` from) is confirmed to already exist implicitly —
  `StartSpan` returns a `context.Context` carrying the span, and the otel adapter's
  `t.tracer.Start(ctx, name)` stores the span in ctx via the OTel SDK's own context-key mechanism.
  The exact code path the new handler wrapper uses to retrieve "the currently active span" (whether
  via a wowapi-level context key set by `StartSpan`, or by asking the otel adapter to expose a
  `SpanFromContext`-equivalent helper) is an implementation decision to be made during the story, not
  presumed here — see `plan.md` "Unresolved questions."
- No production code path currently relies on `observability.Span` being a strictly 3-method
  interface (e.g. no external mock or type assertion outside the adapters/kernel packages depends on
  the interface's exact method set) — this is a low-risk assumption for an interface addition
  (widening, not narrowing) but is stated explicitly since a fail-first check should confirm no
  compile-time break in existing consumers.

## Dependencies

- None upstream — this story is independent of AR-01/AR-02 and does not require any other W01 story
  to land first. `W01-E02-S002` depends on this story's task T001 (the port extension); see this
  story's own `dependencies.md`... (note: per mandate structure, a story does not have its own
  standalone `dependencies.md` — cross-story dependency detail for the epic lives in the epic-level
  `dependencies.md`, referenced here).

## Affected packages or components

- `kernel/observability/tracing.go` — `Span` interface, `noopSpan` implementation.
- `adapters/tracing/otel/otel.go` — `otelSpan` implementation.
- `kernel/observability/` — new file for the ctx-aware `slog.Handler` wrapper (exact filename
  determined at implementation time; likely `kernel/observability/logging.go` or
  `kernel/observability/correlation.go` — not yet created, so not asserted here as fact).
- `kernel/observability/middleware.go` — `AccessLog`, if the wiring point is there rather than at
  logger construction.
- `kernel/logging/logging.go` — `New`, if the wrapper is applied at construction time rather than at
  the `AccessLog` middleware layer (implementation-sequence decision, see `plan.md`).

## Compatibility considerations

Widening the `observability.Span` interface with two new methods is a breaking change for any
external implementer of that interface outside this repository's own two implementations
(`noopSpan`, `otelSpan`). Since `observability.Span` is a kernel port intended to be implemented only
by adapters (the pattern `adapters/metrics/prometheus` and `adapters/tracing/otel` establish), this
is treated as a low-compatibility-risk addition, but the story's plan must confirm no other adapter
package implements `Span` before treating it as safe. wowsociety impact is additive only (per
`epic.md` scope) — the regenerated scaffold gains correlation "for free"; the existing committed
`main.go` is not required to change.

## Security considerations

`trace_id`/`span_id` are non-sensitive, low-cardinality identifiers already exposed on HTTP spans via
the `Trace(tr)` middleware (e.g., propagated as `traceparent` headers) — attaching them to log
records introduces no new sensitive-data exposure. The existing `redactAttr` defense-in-depth
mechanism in `kernel/logging/logging.go` must continue to function unaffected by the new handler
wrapper; this is verified, not merely assumed, in this story's test plan.

## Performance considerations

The no-op tracer path (the default when no adapter is wired) must remain allocation-neutral — this is
a specific, benchmarked acceptance criterion (AC-W01-E02-S001-03), not a general aspiration. The
active-span path's cost (calling `TraceID()`/`SpanID()` and injecting two string attrs per log
record) is expected to be small but is not separately budgeted in this story; if it proves material
during implementation, that finding is recorded rather than silently absorbed.

## Observability considerations

This story's entire content is an observability change: it is the mechanism, not a caller of it.

## Migration considerations

None. No schema, data, or config migration is involved — this is a pure code-level and
attribute-level change.

## Documentation requirements

- Update the doc comments on `observability.Span` (`kernel/observability/tracing.go`) to describe
  the new `TraceID()`/`SpanID()` methods and their no-op-vs-real-adapter contract.
- Document the new handler wrapper's behavior (present-when-active, absent-when-not) in its own doc
  comment, since this is the load-bearing contract this story's acceptance criteria test.
- No changes to `docs/blueprint/` are anticipated to be required by this story alone, but the story's
  `implementation.md` must record whether any were needed once implementation occurs.

## Acceptance criteria

- **AC-W01-E02-S001-01** — A log record emitted (via `InfoContext` or equivalent) inside a
  `context.Context` carrying an active recording span (produced by `Trace(tr)` or a direct
  `StartSpan` call with a real, non-no-op tracer) carries a `trace_id` attr whose value equals
  `span.TraceID()` and a `span_id` attr whose value equals `span.SpanID()`.
- **AC-W01-E02-S001-02** — A log record emitted inside a `context.Context` with no active span (or
  with the `NoOpTracer`'s no-op span) carries neither a `trace_id` nor a `span_id` attribute key —
  verified by asserting the keys are absent from the record's attribute set, not merely that their
  values are empty strings. This is a specific negative-case test, not an incidental side effect of
  the positive-case test.
- **AC-W01-E02-S001-03** — A benchmark comparing log emission through the new ctx-aware handler
  wrapper (with `NoOpTracer`/no active span) against the pre-existing plain handler shows no
  statistically meaningful allocation increase (0 additional allocations per operation, or a
  documented, justified exception if truly unavoidable).

## Required artifacts

- Diff/description of the extended `Span` port interface (both port and both implementations).
- Source of the new ctx-aware `slog.Handler` wrapper.
- Wiring diff showing `AccessLog`/logger-construction integration.
- See `artifacts/index.md` for the registered index (status: not yet produced).

## Required evidence

- Fail-first test run showing the positive-case correlation test failing before the fix and passing
  after.
- Negative-case test run output (attrs genuinely absent).
- Allocation-neutrality benchmark output (before/after comparison).
- See `evidence/index.md` for the registered index (status: not yet produced).

## Definition of ready

Per `governance/definition-of-ready.md` — this story is ready once: the port's current shape is
confirmed (done, see "Current-state assessment"), the wiring point (handler-construction vs.
`AccessLog`) is decided during `plan.md` refinement, and no blocking dependency remains open (none
identified — this story has no upstream story dependency).

## Definition of done

Per `governance/definition-of-done.md` — this story is done once all three acceptance criteria are
verified with registered evidence, the port extension is confirmed non-breaking for existing
adapters, documentation is updated, and independent review (mandate §14) has passed, specifically
checking the negative-case test is a genuine absence check (RISK-W01-E02-002).

## Risks

See `../../risks.md` (epic-level) RISK-W01-E02-002 (negative-case test weakness) — the risk most
specific to this story.

## Residual-risk expectations

Once accepted, the residual risk is limited to: (a) any future adapter package implementing `Span`
independently of `noopSpan`/`otelSpan` would need to add the two new methods — a compile-time-enforced
constraint, not a silent gap; (b) the exact mechanism for "locating the active span from ctx" chosen
during implementation becomes a de facto contract other future ctx-aware observability code should
reuse rather than reinvent — this is recorded as a follow-up note in `implementation.md` once known,
not resolved in advance.
