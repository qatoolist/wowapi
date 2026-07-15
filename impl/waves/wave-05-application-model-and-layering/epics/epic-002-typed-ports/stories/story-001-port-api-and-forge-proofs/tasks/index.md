---
id: W05-E02-S001-TASKS-INDEX
type: tasks-index
parent_story: W05-E02-S001
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W05-E02-S001 — Tasks index

Per mandate §16.4. Task files are single-file per the repository's documented adaptation.

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W05-E02-S001-T001](task-001-port-key-api.md) | port.Key[T] and generic free functions | unassigned | todo | none (depends on W05-E01-S001) | Typed port-key API + happy-path round-trip test | AC-W05-E02-S001-01 | not started | not started |
| [W05-E02-S001-T002](task-002-registrar-forge-safety-proof.md) | Compiler factory extension and registrar-forge compile-fail fixture | unassigned | todo | none (depends on W05-E01-S001) | Port-key minting extension + compile-fail fixture | AC-W05-E02-S001-02 | not started | not started |
| [W05-E02-S001-T003](task-003-independent-review.md) | Independent review | unassigned | todo | T001, T002 | Independent-review record per mandate §14 | AC-W05-E02-S001-01, AC-W05-E02-S001-02 | not started | not started |

## Grouping rationale

Per mandate §12: T001 (the API surface) and T002 (the minting extension and its adversarial safety
proof) are kept separate given their distinct risk profiles (PLAN's own Medium vs. High risk
columns for T1/T2) and separate evidence. T003 adds independent review per mandate §14, justified by
T2's own High-risk "verify capability confusion is impossible if AR-01/AR-02 share one `Registrar`
type" framing.
