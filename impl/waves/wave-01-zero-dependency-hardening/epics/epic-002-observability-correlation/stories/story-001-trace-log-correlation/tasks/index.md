---
id: W01-E02-S001-TASKS-INDEX
type: tasks-index
parent_story: W01-E02-S001
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W01-E02-S001 — Tasks index

Per mandate §16.4.

| Task | Title | Owner | Status | Dependencies | Output | Related ACs | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W01-E02-S001-T001](task-001-extend-span-port-traceid-spanid.md) | extend-span-port-traceid-spanid | W01Obs | done | none | `observability.Span` extended with `TraceID()`/`SpanID()` (port now defined in leaf `kernel/tracing`, aliased); both implementations updated | AC-W01-E02-S001-01, -02 | done (folded into T002 evidence) | verified — pending conductor review |
| [W01-E02-S001-T002](task-002-ctx-aware-slog-handler-wiring.md) | ctx-aware-slog-handler-wiring | W01Obs | done | W01-E02-S001-T001 | `NewCorrelatingHandler` wired into `logging.New`; fail-first + negative-case tests; allocation benchmark (0 allocs/op) | AC-W01-E02-S001-01, -02, -03 | done (EV-W01-E02-S001-001/-002/-003) | verified — pending conductor review |
