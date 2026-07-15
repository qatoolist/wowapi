---
id: W04-E02-S002-TASKS-INDEX
type: tasks-index
parent_story: W04-E02-S002
status: accepted
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W04-E02-S002 — Tasks index

Per mandate §16.4. Task files are single-file per the repository's documented adaptation (see
`governance/naming-conventions.md` "Adaptation 1") — each task file below contains its task
definition, implementation record, verification record, and deviations record as internal sections.

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W04-E02-S002-T001](task-001-inbound-two-phase-verification.md) | Inbound two-phase verification for `HandleInbound` | unassigned | todo | W04-E02-S001-T001 | Two-phase read-verify-recheck protocol; rotation-during-verification test | AC-W04-E02-S002-01 | not started | not started |
| [W04-E02-S002-T002](task-002-failed-signature-audit.md) | Failed-signature audit path | unassigned | todo | T001 | Body-free audit row, own short tx | AC-W04-E02-S002-02 | not started | not started |
| [W04-E02-S002-T003](task-003-adapter-idempotency-contract.md) | Per-adapter idempotency-safety contract declaration | unassigned | todo | W04-E02-S001-T002, W04-E02-S001-T003 | Boot-time-enforced contract; `Sender` inventory | AC-W04-E02-S002-03 | not started | not started |
| [W04-E02-S002-T004](task-004-six-boundary-chaos-test.md) | Named 6-boundary chaos test (notify and webhook) | unassigned | todo | W04-E02-S001-T002, W04-E02-S001-T003, T001, W04-E01-S003 | 6-boundary chaos-test suite, both packages | AC-W04-E02-S002-04 | not started | not started |
| [W04-E02-S002-T005](task-005-evidence-aggregation.md) | Evidence aggregation and T7 cross-reference | unassigned | todo | T001, T002, T003, T004 | Consolidated evidence package; T7 cross-reference record | AC-W04-E02-S002-01, AC-W04-E02-S002-02, AC-W04-E02-S002-03, AC-W04-E02-S002-04 | not started | not started |
| [W04-E02-S002-T006](task-006-independent-review.md) | Independent review | unassigned | todo | T001, T002, T003, T004, T005 | Independent-review record per mandate §14 | AC-W04-E02-S002-01, AC-W04-E02-S002-02, AC-W04-E02-S002-03, AC-W04-E02-S002-04 | not started | not started |

## Grouping rationale

Per mandate §12: T001 (inbound two-phase verification), T002 (failed-signature audit), T003
(adapter idempotency contract), and T004 (6-boundary chaos test) are kept as four separate tasks
because each produces a distinct output with its own named evidence path
(`DATA-03/webhook/inbound-two-phase/`, `DATA-03/webhook/failed-sig-audit/`,
`DATA-03/adapter-contract/`, `DATA-03/chaos/`) and materially different risk profiles — T001 carries
the breaking-change risk (RISK-W04-E02-S002-001), T003 carries an inventory-completeness risk (PLAN
T6's own risk note: "Inventory all existing `Sender` implementations first"), and T004 is, per the
source's own framing, "the most labor-intensive requirement in PF-DATA," warranting isolation from
the other three so its scope and evidence are not diluted. This mirrors
`story-003-canary-switch-contract-drills/tasks/index.md`'s pattern in W02-E01 of keeping a
named-drill-heavy task separate from lighter-weight tasks in the same story.

T005 (evidence aggregation) is added here — unlike W02-E01-S001, which judged a fourth aggregation
task unnecessary for a 2-task story — because this story spans four substantive tasks (T001–T004)
producing evidence across four distinct named paths plus the T7 cross-reference, which is itself a
reference to a different epic's (DATA-08's) evidence location
(`DATA-08/wave0/legal-audit/`). Mirroring `story-003-canary-switch-contract-drills/tasks/index.md`'s
own rationale for its evidence-aggregation task (T6-T9's spread across multiple named drill outputs
warranting consolidation into one record), this story's T4/T5/T6/T8 evidence spread plus the T7
cross-reference is large enough and heterogeneous enough (test reports, an inventory report, a
cross-repo-scope reference) that a dedicated consolidation task adds real tracking value: it is the
single place a reviewer confirms all four AC's evidence exists and the T7 cross-reference is present
and correctly scoped, rather than needing to independently cross-check four separate task records
plus `story.md`'s own prose.

This story is P0 (DATA-03 as a whole is P0), so T006 adds an independent-review task per mandate
§14, scoped to confirming all four AC's evidence, the T4 breaking-change compatibility note, and the
T7 cross-reference-only treatment are all genuinely correct, not merely claimed.
