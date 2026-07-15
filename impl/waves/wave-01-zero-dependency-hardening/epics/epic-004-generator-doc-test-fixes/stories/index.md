---
id: W01-E04-STORIES-INDEX
type: stories-index
epic: W01-E04
wave: W01
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W01-E04 — Stories index

| Story | Title | Status | Priority | Source requirement | Tasks | Objective |
|---|---|---|---|---|---|---|
| [W01-E04-S001](story-001-generator-correctness/story.md) | generator-correctness | planned | P0 | DX-01, DX-02 (Wave-0 slice) | 4 | Fix source-built CLI path version resolution (fail-closed, no `v0.0.0` fallback) and the generator's boot-breaking permission-verb defect, proven via a real generate→build→boot→smoke harness |
| [W01-E04-S002](story-002-documentation-reconciliation/story.md) | documentation-reconciliation | planned | P2/P3 | T-DOC-01, DX-05 (residual), FBL-03 | 3 | Fix the plan document's §6-vs-§9 traceability inconsistency, plan DX-05's residual reconciliation items, and plan the wowsociety upstream register's PF-2/PF-6/RFF-001 correction |
| [W01-E04-S003](story-003-e2e-flake-diagnosis/story.md) | e2e-flake-diagnosis | planned | P2 | T-TEST-01 | 2 | Reproduce and honestly diagnose the intermittent `internal/e2e` full-suite failure, re-scoped after the original "shared-DB concurrency" cause attribution was withdrawn |
