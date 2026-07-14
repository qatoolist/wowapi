---
id: W01-E02-S001-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W01-E02-S001
status: produced
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W01-E02-S001 — Artifacts index

Per mandate §9.2. All artifacts are produced as working-tree source at HEAD
`0a31186cada5c275a588c74081cf977adf346e61` (conductor owns commits); repository paths below are the
canonical locations, so no artifact copies are duplicated under `artifacts/`.

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Status |
|---|---|---|---|---|---|---|---|
| ART-W01-E02-S001-001 | Extended Span port interface diff | interface | implementation | `Span` port gains `TraceID()`/`SpanID()`. Port definition extracted to leaf package `kernel/tracing/tracing.go` (see `../deviations.md` DEV-W01-E02-S001-001); `kernel/observability/tracing.go` re-exports by alias; `noopSpan` impls in `kernel/tracing/tracing.go`; `otelSpan` impls in `adapters/tracing/otel/otel.go` | FBL-06 T1 | W01-E02-S001-T001 | produced |
| ART-W01-E02-S001-002 | Ctx-aware slog.Handler wrapper source | source-code package | implementation | `kernel/observability/correlation.go` — `NewCorrelatingHandler` + `ContextWithSpan`/`SpanFromContext` forwarders (canonical helpers in `kernel/tracing`) | FBL-06 T2 | W01-E02-S001-T002 | produced |
| ART-W01-E02-S001-003 | AccessLog/logger-construction wiring diff | source-code package | implementation | `kernel/logging/logging.go` `New` wraps its format handler with `observability.NewCorrelatingHandler` (approach (a): every process logger correlates; `AccessLog` inherits via `InfoContext(r.Context())` inside `Trace(tr)` — proven by `TestAccessLogInsideTraceMiddlewareCarriesExportedSpanIDs`); adapter-side `StartSpan` stores the port span via `observability.ContextWithSpan` in `adapters/tracing/otel/otel.go` | FBL-06 T2 | W01-E02-S001-T002 | produced |
| ART-W01-E02-S001-004 | Updated Span/handler-wrapper doc comments | design document | post-implementation | Contract docs on `kernel/tracing.Span`/`Tracer`, alias notes in `kernel/observability/tracing.go`, wrapper contract on `NewCorrelatingHandler`, wiring note in `logging.New` | FBL-06 T1/T2 | W01-E02-S001-T001/T002 | produced |

## Notes

Tests added (supporting artifacts): `kernel/logging/correlation_test.go`,
`kernel/observability/correlation_test.go` (matrix, middleware end-to-end, benchmark pair).
