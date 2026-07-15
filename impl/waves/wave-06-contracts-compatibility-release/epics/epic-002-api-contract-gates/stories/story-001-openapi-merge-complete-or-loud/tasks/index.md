---
id: W06-E02-S001-TASKS-INDEX
type: tasks-index
parent_story: W06-E02-S001
status: verified
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W06-E02-S001 — Tasks index

Per mandate §16.4. Task files are single-file per the repository's documented adaptation (see
`governance/naming-conventions.md` "Adaptation 1") — each task file below contains its task
definition, implementation record, verification record, and deviations record as internal sections.

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W06-E02-S001-T001](task-001-full-field-merge-struct.md) | Full-field merge struct and per-field policy | W06E02Impl | done | none | Expanded merge struct, fixture-proven | AC-W06-E02-S001-01 | implemented | verified |
| [W06-E02-S001-T002](task-002-validator-decision-and-structural-validation.md) | Validator-dependency decision and structural validation | W06E02Impl | done | T001 | Validator selected+reviewed; structural validation wired | AC-W06-E02-S001-02, AC-W06-E02-S001-04 | implemented | verified |
| [W06-E02-S001-T003](task-003-semantic-diff-gate.md) | Semantic-diff gate keyed to DX-05's v1 policy | W06E02Impl | done | T002 | Semantic-diff CI gate | AC-W06-E02-S001-03 | implemented | verified |
| [W06-E02-S001-T004](task-004-independent-review.md) | Independent review | W06-E01-E04-Execution.W06E02ReviewFinal | done | T001, T002, T003 | Independent-review record per mandate §14 | all | complete | PASS, no open issues |

## Grouping rationale

Per mandate §12: T001–T003 follow PLAN DX-06's own T1–T3 task table exactly, in the same
dependency order (T1 the merge struct itself, T2 validation which operates on T1's output, T3 the
semantic diff which operates on T2's validated output). This story is P1 but owns a duplicate-resolution
contract (AR-03 T2 via CONFLICT-01) and gates W06-E02-S003's T3 leg — T004 adds an independent-review
task per mandate §14's own framing that a story whose failure would silently propagate a field-drop
into every consumer's published API contract warrants review even at P1.
