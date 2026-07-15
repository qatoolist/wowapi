---
id: W01-E02-S002-TASKS-INDEX
type: tasks-index
parent_story: W01-E02-S002
status: recorded
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W01-E02-S002 — Tasks index

Per mandate §16.4. A single task is used for this story per mandate §12's anti-fragmentation
guidance — see `plan.md` "Implementation strategy" for the rationale (one coherent ~50-LOC
deliverable with one evidence artifact: the tracer implementation has no meaning without its wiring,
and both are proven by the same trace-tree-export test).

| Task | Title | Owner | Status | Dependencies | Output | Related ACs | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W01-E02-S002-T001](task-001-implement-pgx-query-tracer.md) | implement-pgx-query-tracer | W01Obs | done | W01-E02-S001-T001 (landed); D-08 ratification confirmed against ADR-W00-E02-S003-008 | `queryTracer` + `WithQueryTracer` in `kernel/database/query_tracer.go`; real-Postgres trace-tree tests | AC-W01-E02-S002-01, -02 | done (EV-W01-E02-S002-001/-002/-003) | verified — pending conductor review |
