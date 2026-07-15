---
id: W05-E05-S002-TASKS-INDEX
type: tasks-index
parent_story: W05-E05-S002
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W05-E05-S002 — Tasks index

Per mandate §16.4. Task files are single-file per the repository's documented adaptation.

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W05-E05-S002-T001](task-001-package-count-and-lint-verification.md) | Kernel package-count and lint verification | unassigned | todo | none (depends on W05-E05-S001) | Count + lint results | AC-W05-E05-S002-01 | not started | not started |
| [W05-E05-S002-T002](task-002-wowsociety-identity-suite-verification.md) | wowsociety identity-suite verification | unassigned | todo | none (depends on W05-E05-S001) | Cross-repo test results | AC-W05-E05-S002-02 | not started | not started |
| [W05-E05-S002-T003](task-003-independent-review.md) | Independent review | unassigned | todo | T001, T002 | Independent-review record per mandate §14 | AC-W05-E05-S002-01, AC-W05-E05-S002-02 | not started | not started |

## Grouping rationale

Per mandate §12: T001 (framework-side count/lint verification) and T002 (cross-repo wowsociety
verification) are kept separate given their distinct execution environments (local/CI vs. a
cross-repo coordination step) and separate evidence. T003 adds independent review per mandate §14,
justified by FBL-01's own "largest single architectural correction" status and by REVIEW §P's own
explicit instruction requiring wowsociety's full (not partial) identity/authz suite re-run.
