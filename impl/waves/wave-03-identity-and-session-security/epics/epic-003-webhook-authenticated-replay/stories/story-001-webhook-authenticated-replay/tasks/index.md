---
id: W03-E03-S001-TASKS-INDEX
type: tasks-index
parent_story: W03-E03-S001
status: planned
derived: false
created_at: 2026-07-13
updated_at: 2026-07-13
---

# W03-E03-S001 — Tasks index

Per mandate §16.4. Task files are single-file per the repository's documented adaptation — each
file below contains its task definition, implementation record, verification record, and deviations
record as internal sections.

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W03-E03-S001-T001](task-001-verifier-envelope-interface.md) | Verifier interface change to (Envelope, error) (SEC-03 T1) | unassigned | done | none | The `Verifier` interface returns `(Envelope, error)`; `Envelope` is defined; ... | AC-W03-E03-S001-01 | completed | passed |
| [W03-E03-S001-T002](task-002-hmac-envelope-synthesis.md) | HMACVerifier authenticated-data synthesis (SEC-03 T2) | unassigned | done | W03-E03-S001-T001 | `Envelope` never surfaces caller-supplied fields; `OccurredAt` is proven immu... | AC-W03-E03-S001-02 | completed | passed |
| [W03-E03-S001-T003](task-003-handleinbound-rewire.md) | HandleInbound rewire to Envelope-only (SEC-03 T3) | unassigned | done | W03-E03-S001-T001, W03-E03-S001-T002 | No security decision in `HandleInbound` reads a raw `InboundIn` field; the ad... | AC-W03-E03-S001-03 | completed | passed |
| [W03-E03-S001-T004](task-004-provider-verifier-contract-doc.md) | Provider-verifier contract document (SEC-03 T4) | unassigned | done | W03-E03-S001-T001, W03-E03-S001-T002, W03-E03-S001-T003 | A provider-verifier contract document exists, with a reference example, accur... | AC-W03-E03-S001-04 | completed | passed |
| [W03-E03-S001-T005](task-005-independent-review.md) | Independent review | unassigned | todo | W03-E03-S001-T001, W03-E03-S001-T002, W03-E03-S001-T003, W03-E03-S001-T004 | A completed review report confirming the checklist above, recorded as evidence. | AC-W03-E03-S001-01, AC-W03-E03-S001-02, AC-W03-E03-S001-03, AC-W03-E03-S001-04 | not started | not started |

## Grouping rationale

Per `plan.md`: tasks follow the PLAN task breakdown for this story (Bind webhook replay and dedup to provider-authenticated data). Each task is
tracked separately because it produces distinct output with separate evidence. The final task is an
independent-review task per mandate §14 for this P0/P1 security or governance story.
