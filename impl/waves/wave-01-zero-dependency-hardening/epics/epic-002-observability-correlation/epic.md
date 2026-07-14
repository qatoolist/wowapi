---
id: W01-E02
type: epic
title: Observability correlation
status: planned
wave: W01
owner: unassigned
reviewer: unassigned
priority: P1
created_at: 2026-07-12
updated_at: 2026-07-12
source_requirements:
  - FBL-06
depends_on:
  - W00-E02-S003
stories:
  - W01-E02-S001
  - W01-E02-S002
decisions:
  - D-08
risks: []
---

# W01-E02 — Observability correlation

## Epic objective

Wire the two most valuable, currently-missing joins in wowapi's existing OpenTelemetry pipeline:
log-to-trace correlation (a log record emitted inside a traced request carries `trace_id`/`span_id`)
and database-to-trace correlation (pgx query spans appear as children of the request span). Both
joins are additive to infrastructure that already exists end-to-end (adapter, OTLP exporter, HTTP
and worker spans) — this epic does not build new tracing infrastructure, it connects two consumers
(the logger, the DB pool) to the tracing port that already exists.

## Problem being solved

FBL-06 ("OTel trace/log correlation + pgx tracer") identifies that wowapi ships a complete tracing
pipeline that is functionally unused for its two highest-value diagnostic joins. Closure-depth
MATRIX CS-05 classifies this as a **utilisation gap**, not a missing-capability gap: the mechanism
(`kernel/observability.Tracer`/`Span` port, the `adapters/tracing/otel` binding, the OTLP exporter,
`Trace(tr)` HTTP middleware, worker/relay tracer wiring in `app/worker.go`) is present and working,
but:

- `kernel/logging.New` (`kernel/logging/logging.go:85-104`) builds a plain `slog.Logger` with no
  context awareness. No handler pulls `trace.SpanContextFromContext(ctx)` (confirmed: zero repo-wide
  hits for that call). `AccessLog` (`kernel/observability/middleware.go:46-62`) logs `request_id` but
  never `trace_id`/`span_id`. A log record and the trace it was emitted inside cannot be joined.
- The `Span` port itself (`kernel/observability/tracing.go:17-39`) has no `TraceID()`/`SpanID()`
  accessor at all — there is no vendor-neutral way for a consumer (the logging handler, or anything
  else) to read the active trace/span identifiers out of the port. `adapters/tracing/otel/otel.go`'s
  `otelSpan` (lines 99-111) wraps the real OTel span but never exposes its `SpanContext()` through the
  port surface.
- `kernel/database.NewPool` (`kernel/database/database.go:128-148`) configures only `MaxConns` on the
  `pgxpool.Config` — no `pgx.QueryTracer` is attached. Every database call inside a traced request is
  therefore invisible in the trace tree: DB time cannot be distinguished from application time when
  reading a trace.

Both gaps are cheap to close because the hard part (a working, sampled, propagating tracer with a
real exporter) already exists; only the missing accessor and two small integration points are absent.

## Scope

- Extending the `observability.Span` port with vendor-neutral trace/span ID accessors and
  implementing them in the otel adapter (S001, task T001).
- A context-aware `slog.Handler` wrapper that injects `trace_id`/`span_id` attrs when a recording
  span is present in the request context, wired into the logger construction path used by
  `AccessLog` and other handler loggers (S001, task T002).
- A thin, hand-rolled `pgx.QueryTracer` implementation in `kernel/database`, consuming the same
  `Tracer` port extended by S001/T001, attached via the pool's existing `Option` mechanism
  (`kernel/database/database.go:48`) (S002, task T001).

## Out of scope

- Introducing an OTel log-export bridge (`otelslog` or equivalent) to ship logs to the OTLP
  collector as OTel log records. The mandate's brief for this epic explicitly rejects this: the need
  here is correlation attributes on existing slog records, not a new OTLP log-export pipeline, and
  adding the bridge would introduce a new dependency where none is needed. If this need resurfaces,
  it is a distinct future story, not part of this epic.
- Any change to sampling policy, exporter configuration, or the OTel SDK wiring in
  `adapters/tracing/otel` beyond exposing `SpanContext()` through the adapter's own `otelSpan` type.
- `otelpgx` (the third-party OTel-pgx bridge). Decision D-08 (ratified in W00-E02-S003) selects a
  hand-rolled thin tracer over the existing port instead; see "Required decisions" below.
- Any change to `app/worker.go`'s existing relay/runner tracer wiring — that wiring already works and
  is out of this epic's scope.
- wowsociety backport of these changes to its committed `main.go`; both stories are additive to the
  regenerated scaffold only (see each story's compatibility considerations).

## Source requirements

- FBL-06 — "OTel trace/log correlation + pgx tracer (D-08)", classification IMPL, priority P1,
  disposition `planned`, target `W01-E02-S001..S002` per `impl/analysis/requirement-inventory.md`
  row FBL-06. MATRIX CS-05 gives the closure-detail spec (T1-T3) this epic's two stories implement.

## Architectural context

This epic closes the "Utilisation" gap MATRIX CS-05 identifies: wowapi already has a complete OTel
pipeline — the `observability.Tracer`/`Span` port (`kernel/observability/tracing.go`), the OTel
adapter binding with OTLP-over-HTTP export and configurable ratio sampling
(`adapters/tracing/otel/otel.go`), the `Trace(tr)` HTTP middleware that opens a server span per
request and tags it with route/method/status/request_id, and existing propagation across the
worker/relay async boundary (`app/worker.go:79,86` via `outbox.WithRelayTracer`/
`jobs.WithRunnerTracer`). What is missing is not tracing infrastructure — it is two consumers wired
to that infrastructure's most valuable outputs: the structured logger (so operators can pivot from a
log line to its trace) and the database pool (so operators can see DB time inside a trace). This
epic's architectural discipline mirrors the rest of the adapters layer: the port
(`kernel/observability.Tracer`/`Span`) stays vendor-neutral — no OTel types leak into
`kernel/observability`'s interfaces or into `kernel/database`'s consumer of that port — only the
adapter implementation in `adapters/tracing/otel` touches OTel SDK types directly. This is the same
port-discipline boundary decision D-08 exists to protect for S002 specifically (rejecting `otelpgx`
because it would bind OTel vendor types into `kernel/database`).

## Included stories

- **W01-E02-S001 — trace-log-correlation**: extends the `Span` port with `TraceID()`/`SpanID()`
  accessors (task T001) and adds a ctx-aware `slog.Handler` wrapper plus `AccessLog` wiring so log
  records inside an active span carry `trace_id`/`span_id`, with a specific negative-case test proving
  the attrs are genuinely absent (not empty-string) outside a span (task T002). Implements FBL-06
  T1+T2.
- **W01-E02-S002 — pgx-query-tracer**: implements a thin `pgx.QueryTracer` in `kernel/database` over
  the `Tracer` port extended by S001/T001, attached to the pool config, producing a per-query span
  attached as a child of the parent span (task T001). Implements FBL-06 T3, per ratified decision
  D-08. Depends on S001's T001 (the port extension) but not on S001's T002 (the logging wrapper) — see
  `dependencies.md`.

## Dependencies

- **S002 → S001 (task T001 only)**: S002's `pgx.QueryTracer` calls the same `Tracer`/`Span` port S001
  extends with `TraceID()`/`SpanID()`. This is an internal, intra-epic dependency recorded precisely
  at task granularity in `dependencies.md` — S002 does not depend on S001's task T002 (the slog
  handler wrapper), since the query tracer consumes the `Tracer` port directly, not the logging
  pipeline.
- **This epic → W00-E02-S003 (decision D-08)**: S002 implements a decision (thin in-kernel tracer,
  not `otelpgx`) that is ratified upstream in W00-E02-S003, a sibling wave's epic/story. This epic
  does not re-derive that decision; it cites D-08 as an upstream input. At the time this epic's
  planning documents were authored, the ADR file for D-08 (expected at
  `impl/waves/wave-00-baseline-and-verification/epics/epic-002-baseline-capture/stories/story-003-adr-ification/decisions/`)
  had not yet been created — see each story's `plan.md` "Unresolved questions" section for how this is
  tracked as an assumption to confirm, not silently presumed unchanged.
- **Cross-cutting, per the wave's own PF-ARCH notes**: this epic has no dependency on, and is not a
  blocker for, AR-01/AR-02 (the ApplicationModel work in W05) — it is independent observability
  plumbing, not application-model-shaped work.

## Risks

No epic-specific risks beyond the two recorded in `risks.md` (S002's dependency on an
upstream-ratified decision that may not exist yet when this epic starts; and the S001 negative-case
test being a specific, easy-to-get-wrong assertion — "absent" must mean absent, not empty-string).

## Required decisions

- **D-08** — pgx query tracer approach: thin, hand-rolled `pgx.QueryTracer` over the existing
  `observability.Tracer` port, not the third-party `otelpgx` bridge. Ratified in W00-E02-S003 (a
  sibling wave's epic/story, not owned by this epic). This epic's S002 **implements** D-08; it does
  not own or author the ADR. No `decisions/` directory exists under either of this epic's stories —
  the ADR file itself belongs to W00-E02-S003.

## Epic acceptance criteria

- **AC-W01-E02-01** — A log record emitted via a handler-logger inside an HTTP request with an active
  recording span carries `trace_id` and `span_id` string attrs whose values match the span's actual
  OTel trace/span IDs. A log record emitted with no active span in its context carries neither
  attribute (verified as genuinely absent from the record's attribute set, not present with an empty
  string value). Traces to W01-E02-S001, AC-W01-E02-S001-*.
- **AC-W01-E02-02** — The no-op tracer code path (the default when no tracing adapter is wired) shows
  no allocation regression versus the pre-epic baseline, proven by a benchmark comparing the
  ctx-aware handler wrapper against the current plain handler with `NoOpTracer`/no active span.
  Traces to W01-E02-S001, AC-W01-E02-S001-*.
- **AC-W01-E02-03** — A pgx query executed inside a traced request produces a span that appears as a
  direct child of that request's span in an exported trace tree (verified via an in-memory span
  exporter test fixture, not merely unit-level mock assertions). Traces to W01-E02-S002,
  AC-W01-E02-S002-*.
- **AC-W01-E02-04** — The `observability.Span` port interface itself contains no OpenTelemetry SDK
  types in its method signatures; only `adapters/tracing/otel`'s implementation of that interface
  references OTel SDK types. Verified by code review against both S001/T001's port change and
  S002/T001's tracer implementation. Traces to both stories.

## Closure conditions

- Both stories (W01-E02-S001, W01-E02-S002) reach status `accepted` per `governance/definition-of-done.md`.
- All four epic acceptance criteria above are verified with registered evidence (not merely
  implemented) per mandate §2.5.
- Decision D-08 is confirmed ratified with the wording this epic's S002 plan assumed, or a
  deviation is recorded in S002's `deviations.md` if the ratified wording differs.
- No unresolved regression: the epic's `risks.md` entries are either closed or explicitly accepted
  with rationale.
- This epic's `closure-report.md` is completed and accepted by the acceptance authority.
