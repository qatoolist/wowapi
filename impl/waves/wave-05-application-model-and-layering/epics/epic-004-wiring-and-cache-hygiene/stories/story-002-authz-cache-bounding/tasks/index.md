---
id: W05-E04-S002-TASKS-INDEX
type: tasks-index
parent_story: W05-E04-S002
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W05-E04-S002 — Tasks index

Per mandate §16.4. Task files are single-file per the repository's documented adaptation.

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W05-E04-S002-T001](task-001-bounded-sharded-cache.md) | Bounded, sharded cache | unassigned | todo | none | golang-lru/v2 cache + tests | AC-W05-E04-S002-01 | not started | not started |
| [W05-E04-S002-T002](task-002-eviction-metrics.md) | Eviction with admission/eviction metrics | unassigned | todo | T001 | Eviction + metrics + test | AC-W05-E04-S002-01 | not started | not started |
| [W05-E04-S002-T003](task-003-singleflight-collapse.md) | Singleflight-collapse of concurrent misses | unassigned | todo | T001 | Singleflight + test | AC-W05-E04-S002-02 | not started | not started |
| [W05-E04-S002-T004](task-004-authz-epoch-table-and-wiring.md) | authz_epoch table and cross-pod epoch-bump wiring | unassigned | todo | T001, T002, T003 | Epoch table + wiring + cross-pod test | AC-W05-E04-S002-02 | not started | not started |
| [W05-E04-S002-T005](task-005-decision-provenance-and-prod-config-gate.md) | Decision provenance and prod-config gate | unassigned | todo | T004 | Decision metadata + config gate + tests | AC-W05-E04-S002-03 | not started | not started |
| [W05-E04-S002-T006](task-006-independent-review.md) | Independent review | unassigned | todo | T001, T002, T003, T004, T005 | Independent-review record per mandate §14 | AC-W05-E04-S002-01, AC-W05-E04-S002-02, AC-W05-E04-S002-03, AC-W05-E04-S002-04 | not started | not started |

## Grouping rationale

Per mandate §12: T001-T003 (T1-T3) follow PLAN SEC-04's own task split, each independent enough to
land separately (MATRIX CS-17: "T1 (LRU swap) is independent, land any time"). T004 (T4) is kept
separate given its own materially larger scope and "Highest-risk task" status. T005 groups T5+T6
together (both Low-risk, small, closely-related "expose state correctly" concerns per PLAN's own
risk column) rather than as two trivial separate tasks, per mandate §12's own "avoid excessive
fragmentation" guidance. T006 adds independent review per mandate §14, justified by T004's own
"Highest-risk task" framing and by the DATA-07 T4 cross-wave AC-closure relationship this story's
own AC-04 records.
