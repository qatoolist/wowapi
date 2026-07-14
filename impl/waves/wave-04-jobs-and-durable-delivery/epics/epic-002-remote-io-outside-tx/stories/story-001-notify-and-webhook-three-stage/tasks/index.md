---
id: W04-E02-S001-TASKS-INDEX
type: tasks-index
parent_story: W04-E02-S001
status: accepted
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W04-E02-S001 — Tasks index

Per mandate §16.4. Task files are single-file per the repository's documented adaptation (see
`governance/naming-conventions.md` "Adaptation 1") — each task file below contains its task
definition, implementation record, verification record, and deviations record as internal sections.

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W04-E02-S001-T001](task-001-shared-primitive-reuse-and-lease-columns.md) | Shared-primitive reuse: claim-row lease-column migration | unassigned | todo | W04-E01-S001 | Lease-column migration for notify/webhook delivery-tracking tables | AC-W04-E02-S001-01 | not started | not started |
| [W04-E02-S001-T002](task-002-notify-three-stage-protocol.md) | Three-stage protocol for `kernel/notify` | unassigned | todo | T001 | Claim-tx/effect/finalize-tx protocol; self-documented comment deleted/updated | AC-W04-E02-S001-02 | not started | not started |
| [W04-E02-S001-T003](task-003-webhook-three-stage-protocol.md) | Three-stage protocol for `kernel/webhook.deliverToEndpoint` | unassigned | todo | T001 | Claim-tx/effect/finalize-tx protocol; current-row-state check moved to claim stage | AC-W04-E02-S001-03 | not started | not started |
| [W04-E02-S001-T004](task-004-independent-review.md) | Independent review | unassigned | todo | T001, T002, T003 | Independent-review record per mandate §14 | AC-W04-E02-S001-01, AC-W04-E02-S001-02, AC-W04-E02-S001-03 | not started | not started |

## Grouping rationale

Per mandate §12: T001 (shared-primitive reuse / migration) is kept separate from T002/T003 because
it produces a distinct output (a schema migration) consumed by both notify and webhook, with its
own evidence (`DATA-03/lease-columns/`) and its own risk profile (a migration touching live
delivery-tracking tables, distinct from the transaction-boundary restructuring risk T002/T003
carry). T002 and T003 are kept separate from each other because they touch disjoint packages
(`kernel/notify` vs `kernel/webhook`), have separate acceptance criteria (AC-...-02 vs AC-...-03),
and separate evidence paths (`DATA-03/notify/` vs `DATA-03/webhook/`) — per PLAN DATA-03's own task
table treating T2 and T3 as independent rows despite sharing the same protocol shape. This story is
P0 (DATA-03 as a whole is P0, and this story is the epic's foundation, gating S002's T4/T6/T8) per
this epic's priority, so T004 adds an independent-review task per mandate §14, scoped to confirming
the shared primitive was genuinely reused (not copied), the self-documented comment was genuinely
resolved, and both packages' no-remote-call-while-tx-open assertions genuinely hold. No separate
evidence-aggregation task is added — three tasks' evidence (the migration test, the notify
assertion, the webhook assertion) is already a consolidated, story-scope-sized record; a fifth
aggregation task would add no tracking value for a story this size.
