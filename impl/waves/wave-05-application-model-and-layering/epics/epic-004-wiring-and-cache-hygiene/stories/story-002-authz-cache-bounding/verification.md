---
id: VER-W05-E04-S002
type: verification-record
parent_story: W05-E04-S002
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
---

# Verification record — W05-E04-S002

## Planned verification procedure

Per mandate §8.8.

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W05-E04-S002-01 | Run `SEC-04/bounded-cache-tests.md` and `SEC-04/eviction-metrics-tests.md`'s producing tests | Local dev or CI, Go toolchain (`-race`) | Cache never exceeds configured max; idle entries evicted with full metrics | test + race-test report | unassigned |
| AC-W05-E04-S002-02 | Run `SEC-04/singleflight-tests.md` and `SEC-04/cross-pod-epoch-tests.md`'s producing tests | Local dev or CI, Go toolchain, PostgreSQL instance | N misses → 1 DB load; cross-pod revocation visible without full TTL wait, across every enumerated mutation path | test report | unassigned |
| AC-W05-E04-S002-03 | Run `SEC-04/decision-provenance-tests.md` and `SEC-04/prod-config-gate-tests.md`'s producing tests | Local dev or CI, Go toolchain | Decision metadata differs hit vs. miss; prod+cache-enabled+no-bound fails boot | test report | unassigned |
| AC-W05-E04-S002-04 | Inspect this story's `story.md` "Dependencies" section for the DATA-07 T4 cross-reference | Documentation review | DATA-07 T4's cache-invalidation AC-closure relationship recorded by ID | cross-reference confirmation | unassigned |

## Post-execution record

*Fill in after verification is actually executed.*

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
