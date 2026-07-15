---
id: VER-W04-E01-S001
type: verification-record
parent_story: W04-E01-S001
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Verification record — W04-E01-S001

## Planned verification procedure

Per mandate §8.8. One row per acceptance criterion for this story.

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W04-E01-S001-01 | Run unit tests on token/generation comparison semantics | Local dev environment or CI, Go toolchain | Current token/generation pair compares valid; stale/expired pair compares rejected | unit-test report | unassigned |
| AC-W04-E01-S001-02 | Inspect the cross-consumer field-set review record against DATA-03/DATA-04's stated needs | Documentation / review-record inspection | A dated, attributed review record exists confirming the field set covers DATA-03/DATA-04's stated needs, predating the design being locked | review report | unassigned |
| AC-W04-E01-S001-03 | Run the interim-checkpoint-lease migration test | Local dev or CI, PostgreSQL instance with simulated interim-lease checkpoint state | Checkpoint state correctly re-expressed under the new primitive's schema; no loss or duplication across the cutover | migration-test report | unassigned |

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
