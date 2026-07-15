---
id: W01-E02-ACCEPTANCE
type: epic-acceptance
epic: W01-E02
wave: W01
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W01-E02 — Epic-level acceptance

Epic-level acceptance criteria are restated from `epic.md` here for standalone reference, plus each
story's contributing acceptance criteria. Story-level detail lives in each story's own `story.md`.

## AC-W01-E02-01 — Trace/log correlation present when active, absent when not

A log record emitted via a handler-logger inside an HTTP request with an active recording span
carries `trace_id` and `span_id` string attrs whose values match the span's actual OTel trace/span
IDs. A log record emitted with no active span in its context carries neither attribute — verified as
genuinely absent from the record's attribute set (key not present), not present with an empty-string
value. Traces to W01-E02-S001 (AC-W01-E02-S001-01, AC-W01-E02-S001-02).

## AC-W01-E02-02 — No-op tracer path is allocation-neutral

The no-op tracer code path (the default when no tracing adapter is wired) shows no allocation
regression versus the pre-epic baseline, proven by a benchmark comparing the ctx-aware handler
wrapper against the current plain handler with `NoOpTracer`/no active span. Traces to W01-E02-S001
(AC-W01-E02-S001-03).

## AC-W01-E02-03 — pgx spans appear as trace-tree children

A pgx query executed inside a traced request produces a span that appears as a direct child of that
request's span in an exported trace tree, verified via an in-memory span exporter test fixture (not
merely unit-level mock assertions). Traces to W01-E02-S002 (AC-W01-E02-S002-01).

## AC-W01-E02-04 — Port stays vendor-neutral

The `observability.Span` port interface itself contains no OpenTelemetry SDK types in its method
signatures; only `adapters/tracing/otel`'s implementation of that interface references OTel SDK
types. Verified by code review against both S001/T001's port change and S002/T001's tracer
implementation. Traces to both stories.

## Acceptance authority

Framework architecture lead (role-based, per `wave.md`'s split: this epic falls under the
"ARCH-adjacent linter/observability/HTTP work" bucket assigned to the framework architecture lead,
not the developer-experience lead).

## Acceptance status

| AC | Status | Evidence | Notes |
|---|---|---|---|
| AC-W01-E02-01 | not started | — | — |
| AC-W01-E02-02 | not started | — | — |
| AC-W01-E02-03 | not started | — | — |
| AC-W01-E02-04 | not started | — | — |

## Acceptance record — 2026-07-13

Satisfied 2026-07-13. All acceptance criteria for W01-E02 are met; independent review passed
(W01ReviewGate); accepted by conductor.
