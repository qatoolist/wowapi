---
id: W01-E02-DEPS
type: epic-dependencies
epic: W01-E02
wave: W01
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W01-E02 — Dependencies

## Upstream (epics/waves this epic depends on)

- **W00-E02-S003** (ADR-ification, a different wave's epic) — this epic's S002 story implements
  decision **D-08** (thin in-kernel `pgx.QueryTracer` over the existing `observability.Tracer` port,
  not `otelpgx`), which is ratified there, not here. At the time this epic's planning documents were
  authored, the ADR file itself (expected at
  `impl/waves/wave-00-baseline-and-verification/epics/epic-002-baseline-capture/stories/story-003-adr-ification/decisions/`)
  had not yet been created. This epic does not block on that file existing to be *planned*, but
  S002's implementation work is blocked on D-08 being ratified before it *begins* — see `wave.md`'s
  own entry-criteria note: "D-08 (pgx query tracing approach) is ratified before W01-E02-S002 can
  implement it as specified rather than as an open question."
- **W00 (baseline)** generally — per the wave-level entry criteria, W00's exit gate (re-verification
  of executed slices, ratified ADRs) gates this epic's start alongside the rest of W01.

## Internal (intra-epic, story/task-level)

- **W01-E02-S002 depends on W01-E02-S001's task T001, not the whole of S001.** S002's
  `pgx.QueryTracer` implementation calls the `TraceID()`/`SpanID()` (or equivalent span-context)
  accessors that S001/T001 adds to the `observability.Span` port and its otel adapter implementation.
  It does **not** depend on S001/T002 (the ctx-aware `slog.Handler` wrapper) — the query tracer
  consumes the `Tracer` port directly to attach a child span, it does not go through the logging
  pipeline at all. This is recorded precisely (task-level, not story-level) because collapsing it to
  "S002 depends on S001" would incorrectly imply S002 is blocked on the logging-wrapper work, which it
  is not; the two tasks (S001/T002 and S002/T001) are independently implementable and independently
  verifiable once S001/T001 exists.
- Consequently, `story.md` front matter for S002 records `depends_on: ["W01-E02-S001"]` (the mandate's
  §6 metadata schema is story-granular, not task-granular, for the `depends_on` field), and this
  distinction is spelled out in prose here and in S002's own `plan.md` so a reader does not
  over-interpret the story-level dependency as covering S001/T002 as well.

## Downstream (epics/stories that depend on this epic)

None recorded. Per `wave.md`'s downstream dependency table, none of W01's downstream consumers
(W03-E01, W05-E03, W06-E02) depend on W01-E02 specifically — they depend on W01-E01 and W01-E03
outputs. This epic's outputs (correlation attrs, pgx spans) are consumed operationally (by whoever
reads logs/traces) rather than architecturally by a later wave's implementation.

## Cross-cutting note

Per the wave's own PF-ARCH cross-cutting notes (reproduced in each story's `story.md`): this epic's
dependencies are "none — independent of AR-01/02." This epic's observability work does not block on,
and is not blocked by, the ApplicationModel work in W05.

## External dependencies

- OTel adapter (`adapters/tracing/otel`) and its OTLP exporter are already present in the module
  graph — this epic extends the existing adapter's `otelSpan` type, it does not introduce a new
  external service or SDK dependency.
- No new Go module dependency is introduced by either story (see each story's "Reuse tier" framing
  in `plan.md`): S001 uses stdlib `log/slog` plus the existing `observability`/`otel` packages; S002
  uses the `pgx.QueryTracer` interface already available from the `pgx` module already imported by
  `kernel/database`.

## Decision dependencies

- **D-08** — pgx query tracer design (thin in-kernel tracer, not `otelpgx`). Source: W00-E02-S003.
  Consumed by: W01-E02-S002. Blocking: yes, for S002's implementation start; no, for this epic's
  planning documents (which correctly cite D-08 by ID without re-deriving or presuming its exact
  final wording — see `epic.md` and S002's `plan.md`).
