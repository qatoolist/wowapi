---
id: VER-W04-E01-S003
type: verification-record
parent_story: W04-E01-S003
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Verification record — W04-E01-S003

## Planned verification procedure

Per mandate §8.8. One row per acceptance criterion for this story.

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W04-E01-S003-01 | Run the duplicate-effect / registration-rejection test | Local dev or CI, Go toolchain | Worker without a declared mechanism cannot register; worker with exactly one declared mechanism registers | test report | unassigned |
| AC-W04-E01-S003-02 | Run the effect-ledger-vs-fencing test | Local dev or CI, PostgreSQL instance | Fencing alone does not undo a committed stale-worker domain transaction; effect ledger catches an idempotency-ignoring worker | integration-test report | unassigned |
| AC-W04-E01-S003-03 | Run the named chaos test `DATA-02/chaos/duplicate_worker_lease_expiry_test.go` | Local dev or CI, PostgreSQL instance, multi-goroutine/multi-process test harness | Exactly one logical effect recorded; worker A's writes rejected at all three named boundaries (domain, external, finalize) | chaos-test report | unassigned |

## Post-execution record

*Fill in after verification is actually executed. Do not record results that were not actually
observed.*

### Actual result

*Not yet executed.*

### Pass or fail

*Not yet executed.*

### Evidence identifier

*Not yet executed.*

### Execution date

*Not yet executed.*

### Commit or revision

*Not yet executed.*

### Environment

*Not yet executed.*

### Reviewer

*Not yet executed.*

### Findings

*Not yet executed.*

### Retest status

*Not yet executed.*

### Final conclusion

*Not yet executed.*
