---
id: W04-E01-S003-TASKS-INDEX
type: tasks-index
parent_story: W04-E01-S003
status: accepted
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W04-E01-S003 — Tasks index

Per mandate §16.4. Task files are single-file per the repository's documented adaptation (see
`governance/naming-conventions.md` "Adaptation 1").

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W04-E01-S003-T001](task-001-idempotency-declaration-contract.md) | Idempotency-declaration contract | unassigned | todo | W04-E01-S002 | Registration-time contract + key/lease-context threading + duplicate-effect test | AC-W04-E01-S003-01 | not started | not started |
| [W04-E01-S003-T002](task-002-effect-ledger-vs-fencing-test.md) | Effect-ledger-vs-fencing test | unassigned | todo | T001 | Testable proof fencing alone does not undo a committed stale-worker transaction | AC-W04-E01-S003-02 | not started | not started |
| [W04-E01-S003-T003](task-003-chaos-harness-and-named-test.md) | Chaos harness and named chaos test | unassigned | todo | T001 | Reusable harness + `DATA-02/chaos/duplicate_worker_lease_expiry_test.go` | AC-W04-E01-S003-03 | not started | not started |
| [W04-E01-S003-T004](task-004-evidence-aggregation.md) | Evidence aggregation (consolidated bundle) | unassigned | todo | T001, T002, T003 | Consolidated evidence bundle registered in `evidence/index.md` | AC-W04-E01-S003-01, AC-W04-E01-S003-02, AC-W04-E01-S003-03 | not started | not started |
| [W04-E01-S003-T005](task-005-independent-review.md) | Independent review | unassigned | todo | T001–T004 | Independent-review record per mandate §14 | AC-W04-E01-S003-01, AC-W04-E01-S003-02, AC-W04-E01-S003-03 | not started | not started |

## Grouping rationale

Per mandate §12: T001 (idempotency contract), T002 (effect-ledger-vs-fencing test), and T003 (chaos
harness + named test) map onto PLAN DATA-02's own T5/T6/T7 rows, each carrying its own required
artifact path (`DATA-02/idempotency/`, `DATA-02/worker-contract/`,
`DATA-02/chaos/duplicate_worker_lease_expiry_test.go`) and materially different risks (T5's is the
confirmed-breaking worker-signature change requiring wowsociety coordination, RISK-W04-003; T6's is
low per PLAN's own risk column; T7's is "must exercise all 3 named boundaries" plus the harness's
cross-epic reuse obligation) — no merging is warranted. T002 depends on T001 because the
effect-ledger test needs the idempotency contract's key/lease-context threading to construct a
realistic stale-worker scenario; T003 depends on T001 for the same reason (the chaos test invokes
workers using the new idempotency key/lease-context shape). T002 and T003 do not depend on each
other and may proceed in parallel once T001 lands.

T004 (evidence aggregation) is added by this story's own judgment, following the same reasoning
pattern as W02-E01-S003-T005: this story's evidence is produced by three separate tasks (a
registration test, an integration test, and a chaos test), and nothing in PLAN DATA-02's own T5–T7
rows owns assembling them into the one consolidated record this story's closure and the epic's
AC-W04-E01-03 both consume. This differs from a case like W04-E01-S002 (this epic's own prior
story), where three tasks' evidence remained naturally separable and small enough not to warrant
aggregation — here, three materially different evidence types (unit/registration test, integration
test, multi-boundary chaos test) plus the harness's explicit cross-epic reuse obligation make a
single consolidated bundle genuinely useful for W04-E02/W04-E03 reviewers who will need to confirm
this story's chaos-test evidence before trusting their own reuse of the harness. T004's own Task
Definition records this reasoning. This story is P0, so T005 adds an independent-review task per
mandate §14, with specific attention to the chaos test genuinely exercising all three named
boundaries and the T5 coordination note being honestly recorded as open, not silently resolved
(epic-level AC-W04-E01-04's story-specific review focus).
