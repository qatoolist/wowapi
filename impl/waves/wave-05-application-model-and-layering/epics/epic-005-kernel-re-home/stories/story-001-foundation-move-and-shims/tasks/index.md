---
id: W05-E05-S001-TASKS-INDEX
type: tasks-index
parent_story: W05-E05-S001
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W05-E05-S001 — Tasks index

Per mandate §16.4. Task files are single-file per the repository's documented adaptation.

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W05-E05-S001-T001](task-001-foundation-tree-and-mechanical-moves.md) | Foundation tree creation and 8 mechanical package moves | unassigned | todo | none (depends on W05-E01, W05-E02) | foundation/ tree + 8 packages moved + build success | AC-W05-E05-S001-01 | not started | not started |
| [W05-E05-S001-T002](task-002-mfa-rehome-and-shim.md) | kernel/mfa re-home and forwarding shim | unassigned | todo | none (depends on W05-E01, W05-E02) | mfa moved + shim + equivalence test | AC-W05-E05-S001-01, AC-W05-E05-S001-02 | not started | not started |
| [W05-E05-S001-T003](task-003-depguard-extension.md) | Depguard extension | unassigned | todo | T001, T002 | Extended depguard rule + adversarial fixture | AC-W05-E05-S001-03 | not started | not started |
| [W05-E05-S001-T004](task-004-boundaries-lint-extension.md) | Boundaries-lint allowlist extension | unassigned | todo | T001, T002 | Extended allowlist + adversarial fixture | AC-W05-E05-S001-03 | not started | not started |
| [W05-E05-S001-T005](task-005-independent-review.md) | Independent review | unassigned | todo | T001, T002, T003, T004 | Independent-review record per mandate §14 | AC-W05-E05-S001-01, AC-W05-E05-S001-02, AC-W05-E05-S001-03 | not started | not started |

## Grouping rationale

Per mandate §12: T001 (the 8 zero-consumer mechanical moves) and T002 (the auth-critical `mfa` move
+ shim) are kept separate given their materially different risk profiles — T001 is low-risk and
mechanical; T002 is security-sensitive per REVIEW §P. T003 and T004 (the two lint extensions) are
kept separate given they touch different tooling (depguard vs. the boundaries script) with separate
adversarial fixtures. T005 adds independent review per mandate §14, justified by FBL-01's own
"largest single architectural correction" status and T002's auth-critical nature.
