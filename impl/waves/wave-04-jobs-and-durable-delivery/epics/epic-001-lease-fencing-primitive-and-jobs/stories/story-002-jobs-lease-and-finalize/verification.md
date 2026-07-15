---
id: VER-W04-E01-S002
type: verification-record
parent_story: W04-E01-S002
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Verification record — W04-E01-S002

## Planned verification procedure

Per mandate §8.8. One row per acceptance criterion for this story.

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W04-E01-S002-01 | Run migration + unit test for claim SQL lease assignment | Local dev or CI, PostgreSQL instance, Go toolchain | Claim assigns fresh token + `generation+1`; `claimedJob` carries lease context | migration + unit-test report | unassigned |
| AC-W04-E01-S002-02 | Run stale-finalize rejection test alongside a legitimate-finalize positive-case test | Local dev or CI, PostgreSQL instance | Stale finalize affects 0 rows, observably rejected; legitimate finalize succeeds unregressed | integration-test report | unassigned |
| AC-W04-E01-S002-03 | Run the same test as AC-W04-E01-S002-02, asserting the reclaim generation delta | Local dev or CI, PostgreSQL instance | `ReclaimStalled` bumps `lease_generation`; delta is provable | integration-test report | unassigned |

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
