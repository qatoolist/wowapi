---
id: W05-E04-S001-TASKS-INDEX
type: tasks-index
parent_story: W05-E04-S001
status: complete
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W05-E04-S001 — Tasks index

Per mandate §16.4. Task files are single-file per the repository's documented adaptation.

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W05-E04-S001-T001](task-001-constructor-boundary-lint.md) | Constructor-boundary lint tool | task | done | none | `internal/tools/constructorlint` + aliased adversarial fixture + enforced Make target | AC-W05-E04-S001-01 | implemented | verified |
| [W05-E04-S001-T002](task-002-kernel-constructor-audit.md) | kernel/kernel.go audit | task | done | none | `evidence/AR-06/kernel_constructor_audit.md` | AC-W05-E04-S001-02 | complete | verified |

## Grouping rationale

Per mandate §12: T001 (T2, lint) and T002 (T3, audit) are kept separate given their distinct
outputs (code vs. investigative document) and separate PLAN task rows. No independent-review task —
both carry PLAN's lowest risk ratings among this wave's stories not explicitly named for review.
