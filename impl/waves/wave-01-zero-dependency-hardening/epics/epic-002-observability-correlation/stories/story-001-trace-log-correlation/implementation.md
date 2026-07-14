---
id: IMPL-W01-E02-S001
type: implementation-record
parent_story: W01-E02-S001
status: recorded
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Implementation record — W01-E02-S001

Implemented 2026-07-13 by W01Obs against HEAD `0a31186cada5c275a588c74081cf977adf346e61`
(conductor owns commits; this record describes the uncommitted working change).

## What was actually implemented

1. **T001 — port extension.** `Span` gained `TraceID() string` / `SpanID() string`. The port
   definition (Tracer, Span, NoOpTracer, context helpers) now lives in the new stdlib-only leaf
   package `kernel/tracing`, re-exported by alias from `kernel/observability` — see
   `deviations.md` DEV-W01-E02-S001-001 for why (pre-existing observability→httpx→authz→database
   import chain would otherwise cycle when S002's kernel/database consumes the port). `noopSpan`
   returns `""` for both; `otelSpan` returns `SpanContext().TraceID()/SpanID().String()` guarded by
   `HasTraceID()`/`HasSpanID()` (invalid context → `""`, per the documented contract).
2. **T002 — span retrieval decision (plan's "unresolved question 1"):** package-level
   `ContextWithSpan(ctx, Span)` / `SpanFromContext(ctx) (Span, bool)` in `kernel/tracing`
   (forwarded by `kernel/observability`). Real adapters store the port-level span in the context
   they return from `StartSpan` — `adapters/tracing/otel.StartSpan` now does this; `NoOpTracer`
   deliberately does not, keeping the disabled path allocation-free. This is the de facto contract
   future ctx-aware observability code should reuse (story's residual-risk note (b)).
3. **T002 — handler wrapper:** `observability.NewCorrelatingHandler(h slog.Handler) slog.Handler`
   (`kernel/observability/correlation.go`). `Handle` injects `trace_id`/`span_id` (on a cloned
   record) only when `SpanFromContext` finds a span whose `TraceID()` is non-empty; otherwise pure
   pass-through (keys genuinely absent, zero extra allocation). `Enabled`/`WithAttrs`/`WithGroup`
   delegate, re-wrapping so derived handlers stay correlating.
4. **T002 — wiring decision (plan's "unresolved question 2"): approach (a)** — `logging.New` wraps
   its constructed format handler, so correlation is a property of every process logger
   (api/worker/migrate), not just the access log. `AccessLog` needed no change: it already logs via
   `InfoContext(r.Context())` and runs inside `Trace(tr)` in the documented chain order; the
   end-to-end test proves the access-log line carries the exported span's IDs.

## Components changed

`kernel/tracing` (new), `kernel/observability`, `kernel/logging`, `adapters/tracing/otel`.

## Files changed

- `kernel/tracing/tracing.go` — new: port definition + noop + context helpers.
- `kernel/observability/tracing.go` — port defs replaced by aliases; `Trace` middleware unchanged.
- `kernel/observability/correlation.go` — new: correlating handler + forwarders.
- `kernel/logging/logging.go` — `New` wraps with `NewCorrelatingHandler`; imports observability.
- `adapters/tracing/otel/otel.go` — `otelSpan.TraceID()/SpanID()`; `StartSpan` stores span via
  `ContextWithSpan`.
- Tests: `kernel/logging/correlation_test.go` (new), `kernel/observability/correlation_test.go`
  (new: matrix, WithAttrs/WithGroup, middleware end-to-end, benchmark pair),
  `kernel/observability/tracing_test.go`, `kernel/outbox/outbox_test.go`,
  `kernel/jobs/trace_test.go`, `kernel/notify/trace_test.go` (test fakes widened with the two new
  methods — the compile-time-enforced consequence the story predicted).

## Interfaces introduced or changed

`tracing.Span` (= `observability.Span`) +2 methods; new `tracing.Tracer`/`Span` aliases;
new exported funcs `ContextWithSpan`, `SpanFromContext`, `NewCorrelatingHandler`.

## Configuration changes

None (as planned — correlation is structural, not config-gated).

## Schema or migration changes

None.

## Security changes

None new; `redactAttr` coexistence verified by `TestCorrelationCoexistsWithSecretRedaction`.

## Observability changes

This story is the observability change.

## Tests added or modified

See "Files changed". Fail-first discipline: positive-case test written and captured failing before
`logging.New` was wired (`evidence/tests/ev-001-fail-first-before.txt`).

## Commits

None by this worker — conductor owns commits. Working-tree diffstat (my files):
7 modified (+53/−48) plus 6 new files (see above).

## Pull requests

None.

## Implementation dates

2026-07-13 (single session).

## Technical debt introduced

None identified.

## Known limitations

- After `WithGroup`, injected `trace_id`/`span_id` land inside the open group — standard slog
  semantics for record-appended attrs (same as any wrapper, e.g. otelslog); presence is the
  contract, asserted in `TestCorrelatingHandlerSurvivesWithAttrsAndWithGroup`.
- An unsampled-but-valid span still yields correlation attrs (IDs exist even when not exported) —
  useful for joining logs to upstream sampled trace segments; noted, not a gap.

## Follow-up items

None blocking. wowsociety committed `main.go` backport remains optional per `wave.md`.

## Relationship to the approved plan

Matches `plan.md` except DEV-W01-E02-S001-001 (port moved to leaf package, aliases preserve every
consumer) — see `deviations.md`. Both deferred decisions resolved as the plan's stated preference
(package-level helpers; wiring at `logging.New`).
