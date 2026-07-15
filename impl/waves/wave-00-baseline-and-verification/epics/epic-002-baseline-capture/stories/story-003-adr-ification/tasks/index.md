---
id: W00-E02-S003-TASKS-INDEX
type: tasks-index
parent_story: W00-E02-S003
status: complete
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W00-E02-S003 — Tasks index

Per mandate §16.4. Three tasks, grouped by logical cluster per the task-grouping decision recorded
in `../plan.md` "Task breakdown — and the task-grouping decision (mandate §12)" — nine near-identical
one-ADR tasks were considered and rejected as excessive fragmentation; a single nine-ADR task was
considered and rejected as mixing unrelated subject areas.

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W00-E02-S003-T001](task-001-application-model-session-authority-decisions.md) | Application-model / session-authority decisions (D-01, D-02, D-03) | W00-E02-S003 execution worker (agent) | done | none | `decisions/adr-001-...md`, `adr-002-...md`, `adr-003-...md` | AC-W00-E02-S003-01, AC-W00-E02-S003-03 | complete (2026-07-12 authoring + 2026-07-13 verification/corrections) | pass (2026-07-13) |
| [W00-E02-S003-T002](task-002-data-release-security-decisions.md) | Data / release / security decisions (D-04, D-05, D-06, D-07) | W00-E02-S003 execution worker (agent) | done | none | `decisions/adr-004-...md`, `adr-005-...md`, `adr-006-...md`, `adr-007-...md` | AC-W00-E02-S003-01, AC-W00-E02-S003-03 | complete (2026-07-12 authoring + 2026-07-13 verification/corrections) | pass (2026-07-13) |
| [W00-E02-S003-T003](task-003-observability-secrets-decisions.md) | Observability / secrets decisions (D-08, D-09) | W00-E02-S003 execution worker (agent) | done | none | `decisions/adr-008-...md`, `adr-009-...md` | AC-W00-E02-S003-01, AC-W00-E02-S003-03 | complete (2026-07-12 authoring + 2026-07-13 verification/corrections) | pass (2026-07-13) |

`decisions/index.md` (registering all nine ADRs, AC-W00-E02-S003-02) is assembled after T001-T003
produce their ADR files; it is not owned by any single task since it depends on all three task
outputs — its assembly is tracked as part of this story's own closure, not as a fourth task, per
mandate §12's fragmentation-avoidance guidance (a pure aggregation step with no independent design
content does not need its own task record).
