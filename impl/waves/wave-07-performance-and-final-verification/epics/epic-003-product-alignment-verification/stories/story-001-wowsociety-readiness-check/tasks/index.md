---
id: W07-E03-S001-TASKS-INDEX
type: tasks-index
parent_story: W07-E03-S001
status: blocked
derived: false
created_at: 2026-07-12
updated_at: 2026-07-14
---

# W07-E03-S001 — Tasks index

Per mandate §16.4. Task files are single-file per the repository's documented adaptation (see
`governance/naming-conventions.md` "Adaptation 1") — each task file below contains its task
definition, implementation record, verification record, and deviations record as internal sections.

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W07-E03-S001-T001](task-001-reverify-prod-01-02-03.md) | Re-verify PROD-01/02/03's enabling capabilities | W07-Phase-A-Execution.W07E03S001 | blocked | none | Direct findings + product paths | AC01, AC02, AC03 | implemented | AC01 fail; AC02/03 pass |
| [W07-E03-S001-T002](task-002-reverify-prod-04-05.md) | Re-verify PROD-04/05's enabling capabilities and rollout plan | W07-Phase-A-Execution.W07E03S001 | blocked | none | Direct findings + product paths | AC04, AC05 | implemented | AC04 fail; AC05 pass |
| [W07-E03-S001-T003](task-003-assemble-consolidated-record.md) | Assemble the consolidated coordination-artifact record | W07-Phase-A-Execution.W07E03S001 | done | T001, T002 findings | `ART-W07-E03-S001-001` | AC01 .. AC05 | implemented | independent review pass; no package issue |

## Grouping rationale

Per mandate §12: T001 (PROD-01/02/03) and T002 (PROD-04/05) are split by which prior wave's own
capability they re-verify (W02/W05/W01+W04 vs. W03/W04+W00), allowing the two tasks to proceed in
parallel since they target disjoint framework capabilities. T003 assembles both into one consolidated
record rather than five separate documents, since all five PROD-0N items share the identical
verification shape and a single record is easier for a future wowsociety maintainer to consult as one
reference. No independent-review task is added: this story is P2, framework-side-only, with no code
change of any kind — the standard closure-time review required by `governance/definition-of-done.md`
already provides adequate rigor for a story of this risk profile.
