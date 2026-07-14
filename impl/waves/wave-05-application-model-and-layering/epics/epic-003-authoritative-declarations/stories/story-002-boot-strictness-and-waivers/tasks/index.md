---
id: W05-E03-S002-TASKS-INDEX
type: tasks-index
parent_story: W05-E03-S002
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W05-E03-S002 — Tasks index

Per mandate §16.4. Task files are single-file per the repository's documented adaptation.

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W05-E03-S002-T001](task-001-duplicate-collector-rejection.md) | Duplicate-collector rejection | unassigned | todo | none (depends on W05-E01) | Rejection + adversarial fixtures | AC-W05-E03-S002-01 | not started | not started |
| [W05-E03-S002-T002](task-002-empty-required-fragment-rejection.md) | Empty-required-fragment rejection | unassigned | todo | none (depends on W05-E01) | Rejection + adversarial fixture | AC-W05-E03-S002-01 | not started | not started |
| [W05-E03-S002-T003](task-003-post-seal-config-rejection.md) | Post-seal config/namespace/collector rejection | unassigned | todo | none (depends on W05-E01-S003) | Rejection extension + regression test | AC-W05-E03-S002-02 | not started | not started |
| [W05-E03-S002-T004](task-004-shared-waiver-mechanism.md) | Shared no-op-adapter waiver mechanism | unassigned | todo | T001, T002, T003 (depends on W05-E01, W05-E02) | Waiver mechanism + integration matrix test | AC-W05-E03-S002-03 | not started | not started |

## Grouping rationale

Per mandate §12: T001-T003 (T2-T4) are kept separate given distinct collector-type/fragment/
config-state concerns and separate named tests, matching PLAN's own task split. T004 (T5) is kept
separate given its own materially larger scope (a new shared, cross-consumer primitive) and its own
integration-matrix test shape. No independent-review task — PLAN's own risk column values (Medium,
Low-medium, Low, Medium) are moderate, and T5's forward-shared-consumer concern is addressed through
design-care emphasis in `plan.md` rather than a dedicated review task.
