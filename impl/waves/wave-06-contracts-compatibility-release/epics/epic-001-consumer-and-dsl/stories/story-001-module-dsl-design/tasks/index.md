---
id: W06-E01-S001-TASKS-INDEX
type: tasks-index
parent_story: W06-E01-S001
status: done
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W06-E01-S001 — Tasks index

Per mandate §16.4. Task files are single-file per the repository's documented adaptation (see
`governance/naming-conventions.md` "Adaptation 1") — each task file below contains its task
definition, implementation record, verification record, and deviations record as internal sections.

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W06-E01-S001-T001](task-001-draft-dsl-design-options.md) | Draft module-DSL design options and trade-offs | W06E01Impl | done | none | Module-DSL design document | AC-W06-E01-S001-01 | implemented | verified |
| [W06-E01-S001-T002](task-002-formalize-labeled-decision-record.md) | Formalize design into a labeled ADR-style decision record | W06E01Impl | done | T001 | ADR-style decision record, labeled target-not-implemented | AC-W06-E01-S001-02 | implemented | verified |

## Grouping rationale

Per mandate §12: T001 (draft the design) and T002 (formalize into a labeled decision record) are
kept as two sequential tasks rather than one, because they produce genuinely distinct outputs — a
design document (exploratory, trade-off-oriented) versus a decision record (a formalized, labeled,
durable artifact) — and because the labeling correctness in T002 is itself a distinct, separately-
checkable concern from the design content's completeness in T001. No independent-review task (T003, per
the pattern used in P0/critical stories elsewhere in this programme) is added here: this story is P1,
not P0/critical, and per mandate §14 the independent-review task is scoped to critical stories — this
design-investigation story with a two-task, no-code, prose-only output does not carry the same
acceptance risk profile that would justify adding a third review-only task on top of the DoD's own
`governance/definition-of-done.md` closure discipline (which already requires reviewer conclusion at
story-closure time regardless of whether a dedicated review task exists).
