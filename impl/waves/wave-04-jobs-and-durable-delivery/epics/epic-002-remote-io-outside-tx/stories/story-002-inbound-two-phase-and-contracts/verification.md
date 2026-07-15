---
id: VER-W04-E02-S002
type: verification-record
parent_story: W04-E02-S002
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Verification record — W04-E02-S002

## Planned verification procedure

Per mandate §8.8. One row per acceptance criterion for this story.

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W04-E02-S002-01 | Run the rotation-during-verification test: rotate/deactivate the endpoint's secret in the snapshot-to-verification window | Local dev environment or CI, PostgreSQL instance | Discard+retry occurs on mismatch; no accept-under-stale-policy; retry attempts bounded | integration-test report | unassigned |
| AC-W04-E02-S002-02 | Run the empty-body-field test on a deliberately failed signature verification | Local dev environment or CI, Go toolchain | Audit row is written in its own short tx; body field is empty | test report | unassigned |
| AC-W04-E02-S002-03 | Run the boot-time fixture test with a deliberately undeclared adapter; confirm `Sender` inventory is complete | Local dev environment or CI, Go toolchain | Boot sequence rejects the undeclared adapter with a clear error; inventory confirms all existing `Sender` implementations correctly declared | boot-time fixture test report + inventory report | unassigned |
| AC-W04-E02-S002-04 | Run the 6-boundary chaos test for both notify and webhook, reusing W04-E01-S003's harness | Local dev environment or CI, PostgreSQL instance, chaos-test harness | Zero duplicate external effects observed across all 6 named boundaries, for both notify and webhook | chaos-test report | unassigned |

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
