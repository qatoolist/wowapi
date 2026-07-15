---
id: VER-W02-E03-S001
type: verification-record
parent_story: W02-E03-S001
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
---

# Verification record — W02-E03-S001

## Planned verification procedure

Per mandate §8.8. One row per acceptance criterion for this story.

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W02-E03-S001-01 | Run the version-allocation concurrency test (≥20 concurrent callers) against both `kernel/artifact.Generate` and `kernel/document.InitiateUpload` | Local dev environment or CI, PostgreSQL instance | N concurrent callers produce N unique, monotonic versions, zero unexpected conflicts; lock wait measured | concurrency-test report | unassigned |
| AC-W02-E03-S001-02 | Run the crash-simulation upload-session test | Local dev environment or CI, PostgreSQL instance | Session row exists with `status='pending'` and a set expiry after simulated crash | integration-test report | unassigned |
| AC-W02-E03-S001-03 | Run the racing-confirmation concurrency test | Local dev environment or CI, PostgreSQL instance | Exactly one of two racing confirms succeeds | concurrency-test report | unassigned |
| AC-W02-E03-S001-04 | Run the mixed confirmed/expired/pending GC sweep test | Local dev environment or CI, PostgreSQL instance + object storage (or fake) | Sweep removes only past-expiry unconfirmed sessions' objects; never removes a referenced object | integration-test report | unassigned |
| AC-W02-E03-S001-05 | Run the dedicated `kernel/artifact.Generate` mirror concurrency test | Local dev environment or CI, PostgreSQL instance | Same concurrency bar as AC-W02-E03-S001-01, proven independently | concurrency-test report | unassigned |

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
