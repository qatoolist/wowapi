---
id: VER-W07-E01-S004
type: verification-record
parent_story: W07-E01-S004
status: passed
created_at: 2026-07-12
updated_at: 2026-07-14
---

# Verification record — W07-E01-S004

## Planned verification procedure

Per mandate §8.8. One row per acceptance criterion for this story.

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W07-E01-S004-01 | Framework upload then Stat with request counter | Real MinIO | zero GetObject calls | EV-W07-E01-S004-001 | W07-Scoping-Dispatch.W07E01S004ReviewR |
| AC-W07-E01-S004-02 | Ambient Stat and labeled bounded repair | Real MinIO | only labeled repair hashes body | EV-W07-E01-S004-002 | W07-Scoping-Dispatch.W07E01S004ReviewR |
| AC-W07-E01-S004-03 | Trigger repair and record metrics | Real MinIO | hit, byte, duration observations | EV-W07-E01-S004-003 | W07-Scoping-Dispatch.W07E01S004ReviewR |
| AC-W07-E01-S004-04 | Interrupt/resume three-object backfill | Real MinIO | complete with exactly three unique repairs | EV-W07-E01-S004-004 | W07-Scoping-Dispatch.W07E01S004ReviewR |
| AC-W07-E01-S004-05 | Inspect publication against reference | Data review | honest comparability and DEC-Q9 posture | EV-W07-E01-S004-005 | W07-Scoping-Dispatch.W07E01S004ReviewR |
| AC-W07-E01-S004-06 | Exact seven-name benchmark run | Local PostgreSQL where needed | all seven hot paths execute | EV-W07-E01-S004-006 | W07-Scoping-Dispatch.W07E01S004ReviewR |
| AC-W07-E01-S004-07 | `make bench-budget` | Local PostgreSQL | exit 0 with all seven entries | EV-W07-E01-S004-007 | W07-Scoping-Dispatch.W07E01S004ReviewR |

## Post-execution record

### Actual result

All seven acceptance procedures passed. S3 checksum behavior ran against real
local MinIO; focused framework/storage/metrics tests passed; the exact seven
benchmarks emitted measurements; the combined budget gate exited 0.

### Pass or fail

**PASS**

### Evidence identifiers

EV-W07-E01-S004-001 through EV-W07-E01-S004-007.

### Execution date and revision

2026-07-14; working tree based on `733ef3e`.

### Environment

Darwin/arm64 Apple M3 Max, Go toolchain, local PostgreSQL and MinIO service
containers. The publication explicitly does not treat this as like-for-like
with the accepted Linux/amd64 reference.

### Reviewer and findings

`W07-Scoping-Dispatch.W07E01S004ReviewR`: correctness `correct`, confidence 1,
findings `[]`.

### Retest status

Not required: independent review found no issue. The final focused S3 tests,
package tests, exact benchmark run, and `make bench-budget` were all green.

### Final conclusion

**Accepted — all seven ACs are evidenced and independently reviewed with no open
issues. Absolute SLO claims remain conditional on DEC-Q9.**
