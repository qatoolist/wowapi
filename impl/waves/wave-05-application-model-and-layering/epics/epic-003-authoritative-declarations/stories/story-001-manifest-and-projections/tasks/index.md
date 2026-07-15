---
id: W05-E03-S001-TASKS-INDEX
type: tasks-index
parent_story: W05-E03-S001
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W05-E03-S001 — Tasks index

Per mandate §16.4. Task files are single-file per the repository's documented adaptation.

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W05-E03-S001-T001](task-001-manifest-schema-definition.md) | Manifest schema definition | unassigned | todo | none (depends on W05-E01) | Schema + round-trip test | AC-W05-E03-S001-01 | not started | not started |
| [W05-E03-S001-T002](task-002-route-derivation-golden-delta-gate.md) | Route derivation and golden-declaration-delta acceptance gate | unassigned | todo | T001 (depends on W05-E01, W05-E02) | Derivation tooling + golden-delta test | AC-W05-E03-S001-02 | not started | not started |
| [W05-E03-S001-T003](task-003-duplicate-omission-lint-rule.md) | Duplicate-identity/omitted-projection lint rule | unassigned | todo | T001, T002 | Lint rule + adversarial fixtures | AC-W05-E03-S001-03 | not started | not started |
| [W05-E03-S001-T004](task-004-doc-test-manifest-export-projections.md) | Documentation/test/manifest export projections | unassigned | todo | T001, T002, T003 | Extended golden-delta coverage | AC-W05-E03-S001-03 | not started | not started |
| [W05-E03-S001-T005](task-005-independent-review.md) | Independent review | unassigned | todo | T001, T002, T003, T004 | Independent-review record per mandate §14 | AC-W05-E03-S001-01, AC-W05-E03-S001-02, AC-W05-E03-S001-03 | not started | not started |

## Grouping rationale

Per mandate §12: T001-T004 follow PLAN AR-03's own T1/T3/T4/T5 task split exactly, matching the
source's own dependency chain. T005 adds independent review per mandate §14, justified specifically
by T002's (PLAN T3's) own "this test IS the acceptance gate" High-risk framing.
