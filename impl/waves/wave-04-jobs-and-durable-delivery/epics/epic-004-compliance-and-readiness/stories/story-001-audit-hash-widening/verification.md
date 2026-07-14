---
id: VER-W04-E04-S001
type: verification-record
parent_story: W04-E04-S001
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Verification record — W04-E04-S001

## Planned verification procedure

Per mandate §8.8. One row per acceptance criterion for this story.

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W04-E04-S001-01 | Run the per-field tamper test: mutate each declared field independently (metadata, tx_id, and every other field in the widened scheme) on a chained row | Local dev environment or CI, PostgreSQL instance | Every independent field mutation causes verification to fail; no field mutation is silently undetected | tamper-test report (per-field results) | unassigned |
| AC-W04-E04-S001-02 | Run the version-branch verification test: verify a hash_version=1 historical-row fixture under the v1 branch, and a new row under the v2 branch | Local dev environment or CI, PostgreSQL instance with the hash_version migration applied | Historical row verifies correctly under v1; new row verifies correctly under v2, including metadata and tx_id coverage | version-branch verification report | unassigned |
| AC-W04-E04-S001-03 | Confirm the hash_version migration has a complete, valid manifest entry per W02-E01-S001's schema and was executed within its lock-timeout budget | Migration-manifest inspection + CI validation output from W02-E01's tooling | Manifest entry present and complete; migration classified and executed through W02-E01's protocol, not ad hoc | migration-classification report | unassigned |

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
