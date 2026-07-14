---
id: W01-E02-STORIES-INDEX
type: stories-index
epic: W01-E02
wave: W01
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W01-E02 — Stories index

Per mandate §16.4 pattern, extended to story scope (this is the epic's index over its own stories;
each story's own `tasks/index.md` provides the task-level view).

| Story | Title | Owner | Status | Dependencies | Output | Related ACs | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W01-E02-S001](story-001-trace-log-correlation/story.md) | trace-log-correlation | W01Obs | verified | none (independent) | `Span` port extended with `TraceID()`/`SpanID()` (port relocated to leaf `kernel/tracing`, observability aliases — DEV-W01-E02-S001-001); `NewCorrelatingHandler` wired into `logging.New` | AC-W01-E02-S001-01, -02, -03 | done | verified — pending conductor acceptance |
| [W01-E02-S002](story-002-pgx-query-tracer/story.md) | pgx-query-tracer | W01Obs | verified | W01-E02-S001 (task T001, landed) | `queryTracer` + `WithQueryTracer` in `kernel/database/query_tracer.go`, per-query child spans proven against real Postgres | AC-W01-E02-S002-01, -02 | done | verified — pending conductor acceptance |
