---
id: W05-E01-S004-TASKS-INDEX
type: tasks-index
parent_story: W05-E01-S004
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W05-E01-S004 — Tasks index

Per mandate §16.4. Task files are single-file per the repository's documented adaptation (see
`governance/naming-conventions.md` "Adaptation 1").

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W05-E01-S004-T001](task-001-legacy-adapter-implementation.md) | Legacy adapter implementation | unassigned | todo | none (depends on S003) | Adapter + compat-test output + adversarial-fixtures-through-legacy-path proof | AC-W05-E01-S004-01, AC-W05-E01-S004-02 | not started | not started |
| [W05-E01-S004-T002](task-002-independent-review.md) | Independent review | unassigned | todo | T001 | Independent-review record per mandate §14 | AC-W05-E01-S004-01, AC-W05-E01-S004-02 | not started | not started |

## Grouping rationale

Per mandate §12: this story's single substantive implementation task (T001) covers the adapter
itself and both required proofs (existing-contract-test compatibility and adversarial-fixture
non-bypass) as one bounded, single-owner unit of work — splitting them would multiply file count with
no added tracking value for a story this focused. T002 adds independent review per mandate §14,
justified specifically by PLAN's own explicit framing of T11: "the adapter is itself a trust
boundary" — this is the same category of risk (a compatibility shim silently reintroducing a closed
security gap) that justifies review on S001/S002 despite this story's own PLAN risk column reading
only "Medium."
