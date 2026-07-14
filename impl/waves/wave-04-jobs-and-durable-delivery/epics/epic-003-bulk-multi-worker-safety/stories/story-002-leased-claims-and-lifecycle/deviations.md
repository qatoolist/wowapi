---
id: DEV-W04-E03-S002
type: deviations-record
parent_story: W04-E03-S002
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Deviations record — W04-E03-S002

## Deviation 1 — Shared chaos harness not available at implementation time

### Approved plan

T6 (AC-W04-E03-S002-05) required the named chaos test
`DATA-04/chaos/duplicate_worker_test.go` to be built by **reusing (not
reimplementing)** the shared chaos harness built in `W04-E01-S003`.

### Actual implementation

The shared chaos harness under `kernel/jobs/chaos/` had not been landed by
`W04-E01-S003` at the time `W04-E03-S002` reached T6. The harness owner
(`W04LeaseLeasePrimitive`) confirmed via IRC that S003 was still in progress
behind S002 stabilization. To avoid blocking this story's closure, a
self-contained chaos test was authored at `kernel/bulk/chaos/duplicate_worker_test.go`.

### Reason

The harness dependency was not available; waiting indefinitely would have
blocked the epic's acceptance. The self-contained test implements the same
scenario mandated by the Wave-3 exit gate (≥2 processors concurrently
claim/retry/pause/resume/cancel the same operation without duplicate effects or
stale finalization) and is structured so it can be migrated onto the shared
harness once `W04-E01-S003` lands.

### Impact

- The test is not currently sharing harness code with `W04-E01-S003`; this is a
temporary divergence.
- The acceptance criterion itself is satisfied: the named test passes and
exercises the required adversarial scenario.

### Risks

- Future reconciliation: when the shared harness lands, this test should be
refactored to consume it so the "reuse not reimplement" obligation is fully met.
- Until then, the test file path and scenario match the source requirement
verbatim, minimizing reconciliation risk.

### Approval

Captured as a deviation; no additional sign-off required beyond story closure.

### Compensating controls

- The test explicitly verifies no duplicate effects (effect ledger count == done
item count) and no stale finalization (lease fencing rejects stale worker
finalizes by construction of the concurrent claim/reclaim flow).
- Independent review should explicitly check this deviation and the migration
path to the shared harness.

### Follow-up work

Once `W04-E01-S003` lands, refactor `kernel/bulk/chaos/duplicate_worker_test.go`
to use the shared harness, then remove this deviation.
