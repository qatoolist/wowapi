---
id: VER-W04-E03-S001
type: verification-record
parent_story: W04-E03-S001
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Verification record — W04-E03-S001

## Planned verification procedure

Per mandate §8.8. One row per acceptance criterion for this story.

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W04-E03-S001-01 | Inspect migration `00016`'s header comment for the corrected claim | Documentation / source inspection | Comment no longer states "safe across replicas"; states the actual single-processor-enforced property | documentation-diff record | unassigned |
| AC-W04-E03-S001-02 | Run the 2-processor concurrency test against the same `bulkID` | Local dev environment or CI, PostgreSQL instance | Exactly one processor succeeds; the second is cleanly rejected, not silently racing | concurrency-test report (`DATA-04/stopgap/`) | unassigned |

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
