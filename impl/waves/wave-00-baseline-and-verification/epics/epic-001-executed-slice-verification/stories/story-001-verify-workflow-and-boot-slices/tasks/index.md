---
id: W00-E01-S001-TASKS-INDEX
type: task-index
parent_story: W00-E01-S001
status: draft
created_at: 2026-07-12
updated_at: 2026-07-13
derived: false
---

# Task index — W00-E01-S001

Per mandate §16.4. Derived roll-up view of this story's three tasks — canonical status lives in
each task file's front matter (`status-model.md` "Canonical source of truth").

| Task ID | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| W00-E01-S001-T001 | Re-verify SEC-02 workflow fail-closed behavior | worker W00E01S001 | done | none | `evidence/tests/sec02-workflow-race.log`; evidence `EV-W00-E01-S001-01` | AC-W00-E01-S001-01 | executed 2026-07-13 | pass |
| W00-E01-S001-T002 | Re-verify AR-04 T1 boot-time unknown-namespace rejection | worker W00E01S001 | done | none | `evidence/tests/ar04-boot-run-boot.log` + `ar04-full-suite.log`; evidence `EV-W00-E01-S001-02` | AC-W00-E01-S001-02 | executed 2026-07-13 | pass |
| W00-E01-S001-T003 | Re-verify AR-06 T1 `authzStore` composition (no duplicate `authz.NewStore()` call) | worker W00E01S001 | done | none | `evidence/tests/ar06-authz-race.log` + `ar06-kernel-rules-race.log`; evidence `EV-W00-E01-S001-03` | AC-W00-E01-S001-03 | executed 2026-07-13 (see DEV-01: `-run` pattern equivalent) | pass |
| W00-E01-S001-T004 | Re-verify AR-05 T1/T2 documentation-drift fixes (phantom-API grep + Context method-set diff) | worker W00E01S001 | blocked | none | `evidence/tests/ar05-doc-drift.log`; evidence `EV-W00-E01-S001-04` | AC-W00-E01-S001-04 | executed 2026-07-13 | **failed as worded** — T2 diff pass; T1 grep found 7 pre-existing future-state hits (identical at fix commit `345e4ce`); conductor adjudication pending (DEV-02) |

## Notes

All three tasks are parallel-safe: they target disjoint packages and disjoint test commands (see
`plan.md` "Implementation sequence" and epic `dependencies.md`). No task blocks another. This
index is a derived view; if it ever disagrees with a task file's front matter, the
front matter wins (`status-model.md`).
