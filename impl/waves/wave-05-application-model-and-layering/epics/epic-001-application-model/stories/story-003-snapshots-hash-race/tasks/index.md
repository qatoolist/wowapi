---
id: W05-E01-S003-TASKS-INDEX
type: tasks-index
parent_story: W05-E01-S003
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W05-E01-S003 — Tasks index

Per mandate §16.4. Task files are single-file per the repository's documented adaptation (see
`governance/naming-conventions.md` "Adaptation 1").

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W05-E01-S003-T001](task-001-snapshot-immutability-conversion.md) | Snapshot-immutability conversion | unassigned | todo | none (depends on S002) | Cloned/immutable exported readers across all wrapped registries | AC-W05-E01-S003-01 | not started | not started |
| [W05-E01-S003-T002](task-002-post-seal-retention-rejection.md) | Post-seal Context/registrar retention rejection | unassigned | todo | none (depends on S001) | Explicit-error retention rejection, validated against wowsociety's pattern | AC-W05-E01-S003-02 | not started | not started |
| [W05-E01-S003-T003](task-003-deterministic-model-hash.md) | Deterministic model hash | unassigned | todo | T001, T002 (full T1-T8 surface) | Model-hash function emitted at startup/readiness | AC-W05-E01-S003-03 | not started | not started |
| [W05-E01-S003-T004](task-004-race-safety-tests.md) | Race-test suite | unassigned | todo | T001, T002, T003 | Race-detector-clean proof of sealed-model integrity | AC-W05-E01-S003-03 | not started | not started |

## Grouping rationale

Per mandate §12: T001 (T7, snapshot immutability) and T002 (T8, post-seal retention rejection) are
kept separate given they touch disjoint concerns (reader-return-value mutability vs. registrar
retention) and have separate named tests. T003 (T9, model hash) and T004 (T10, race tests) are kept
separate given their own distinct evidence (determinism test vs. race-detector output) despite both
depending on the fuller preceding surface. No independent-review task is added for this story — per
this wave's task-brief guidance, this story's task-level risk values (PLAN's own Low-medium/Medium/
Low/Low for T7/T8/T9/T10) are materially lower than S001/S002's High-risk items; epic-level review
coverage from S001/S002 already covers the model's core security-boundary properties this story
builds upon.
