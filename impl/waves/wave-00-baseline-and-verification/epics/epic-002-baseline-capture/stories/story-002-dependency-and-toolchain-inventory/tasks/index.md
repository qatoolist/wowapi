---
id: W00-E02-S002-TASKS-INDEX
type: task-index
parent_story: W00-E02-S002
status: complete
created_at: 2026-07-12
updated_at: 2026-07-13
derived: false
---

# Task index — W00-E02-S002

Per mandate §16.4. Roll-up of this story's tasks. This index is a derived view of each task file's
own front matter (`naming-conventions.md` Adaptation 1: each task is one flat file containing all
four §8.6–§8.9 sections) — if this table and a task file's own front matter ever disagree, the task
file wins.

| Task ID | Title | Owner | Status | Dependencies | Output | Related acceptance criteria | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| W00-E02-S002-T001 | go.mod inventory and approved-register cross-check | W00E02S002 (wave-00 worker) | done | none | `artifacts/post-implementation/dependency-inventory.md` + raw `go list`/`go mod graph`/`go mod why` evidence (`evidence/logs/`, `evidence/reviews/`) | AC-W00-E02-S002-01, AC-W00-E02-S002-02 | complete (2026-07-13) | pass (EV-W00-E02-S002-001, -002) |
| W00-E02-S002-T002 | Pinned tool-version inventory | W00E02S002 (wave-00 worker) | done | none | `artifacts/post-implementation/tool-version-inventory.md` + version-check command evidence (`evidence/logs/tool-versions.txt`) | AC-W00-E02-S002-03 | complete (2026-07-13) | pass (EV-W00-E02-S002-003) |

Both tasks are independent of each other (neither's output is a required input to the other) and
may be worked in parallel, per `plan.md` "Implementation strategy."
