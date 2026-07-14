---
id: W04-E01-S002-TASKS-INDEX
type: tasks-index
parent_story: W04-E01-S002
status: accepted
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W04-E01-S002 — Tasks index

Per mandate §16.4. Task files are single-file per the repository's documented adaptation (see
`governance/naming-conventions.md` "Adaptation 1").

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W04-E01-S002-T001](task-001-lease-column-migration-and-fenced-claim.md) | Lease-column migration and fenced claim SQL | unassigned | todo | W04-E01-S001 | `jobs_queue` lease columns + fenced claim SQL + migration/unit test | AC-W04-E01-S002-01 | not started | not started |
| [W04-E01-S002-T002](task-002-fenced-finalize.md) | Fenced finalize paths | unassigned | todo | T001 | Fenced `complete`/`fail` + stale-rejection test + non-regression test | AC-W04-E01-S002-02 | not started | not started |
| [W04-E01-S002-T003](task-003-fenced-reclaim.md) | Fenced reclaim with generation bump | unassigned | todo | T001 | Fenced `ReclaimStalled` + generation-delta assertion (same test as T002) | AC-W04-E01-S002-03 | not started | not started |
| [W04-E01-S002-T004](task-004-independent-review.md) | Independent review | unassigned | todo | T001, T002, T003 | Independent-review record per mandate §14 | AC-W04-E01-S002-01, AC-W04-E01-S002-02, AC-W04-E01-S002-03 | not started | not started |

## Grouping rationale

Per mandate §12: T001 (lease-column migration + fenced claim), T002 (fenced finalize), and T003
(fenced reclaim) map onto PLAN DATA-02's own T2/T3/T4 rows, each carrying its own required artifact
path (`DATA-02/jobs-lease-migration/`, `DATA-02/finalize/`, `DATA-02/reclaim/`) and — per T4's own
"Same test as T3" instruction — sharing test infrastructure between T002 and T003 while remaining
separate tasks because they touch different code surfaces (finalize vs. reclaim) with materially
different risk framing (T2's risk is timeout-source duplication; T3's is at-least-once-recovery-path
regression; T4 carries no separate risk note). T002 and T003 both depend on T001 (the lease columns
and fenced claim must exist before either finalize or reclaim can compare against lease state) but
not on each other — they may proceed in parallel once T001 lands, since they touch disjoint code
paths (finalize vs. reclaim) sharing only the same underlying test fixture. This story is P0 (DATA-02
as a whole is P0, and this story is where the epic's fencing guarantee first becomes real on
`jobs_queue`) per this wave's task brief, so T004 adds an independent-review task per mandate §14,
scoped to confirming the fencing genuinely does not regress the at-least-once recovery path. No
separate evidence-collection task is added — T001/T002/T003's own evidence (migration/unit-test
output, the stale-finalize/non-regression test, the generation-delta assertion sharing that same
test) is already a consolidated, story-scope-sized record; a fifth aggregation task would add no
tracking value for a story this size (3 substantive tasks).
