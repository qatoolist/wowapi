---
id: VER-W02-E02-S002
type: verification-record
parent_story: W02-E02-S002
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
---

# Verification record — W02-E02-S002

## Planned verification procedure

Per mandate §8.8. One row per acceptance criterion for this story.

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W02-E02-S002-01 | Run the mismatch-audit tool via a platform-role connection against staging/prod-shaped data; run the seeded-mismatch integration test | Staging or prod-shaped environment, platform-role DB connection | Zero-mismatch report across all 8 edges (or a documented, resolved remediation decision per RISK-W02-002); seeded-mismatch fixture is correctly detected | audit report + integration-test report | unassigned |
| AC-W02-E02-S002-02 | Run the composite FK `NOT VALID` add per-table, measure lock duration against the DATA-09 budget — **only after confirming W02-E01-S001 and W02-E01-S002 are `accepted`** | CI or staging environment with DATA-09 lock-timeout tooling available | Each per-table `NOT VALID` add stays under the 2-second lock-timeout budget | migration lock-duration report | unassigned |
| AC-W02-E02-S002-03 | Run `VALIDATE CONSTRAINT` for each of the 8 composite FKs under concurrent writer load — **only after confirming W02-E01-S001 and W02-E01-S002 are `accepted`, and after AC-W02-E02-S002-02 passes** | Staging or prod-shaped environment with concurrent-writer-load test harness | Validation completes without blocking concurrent DML; second zero-mismatch confirmation produced | load-test report + second audit report | unassigned |
| AC-W02-E02-S002-04 | Run the extended catalog-driven RLS matrix test with seeded cross-tenant inserts under both `app_rt` and `app_platform` | CI or staging environment, both role connections available | Insert fails under both roles; platform-role result explicitly asserted, not assumed | RLS matrix test report | unassigned |
| AC-W02-E02-S002-05 (optional) | If pursued: grep sweep for references to the old single-column FK name, plus a full regression run, before the FK-removal migration | CI environment | No code relies on the old FK name for cascade behavior; full regression passes | regression report + grep sweep output | unassigned |

## Post-execution record

*Fill in after verification is actually executed. Do not record results that were not actually
observed — in particular, do not record a mismatch-audit result before T001 has actually run.*

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
