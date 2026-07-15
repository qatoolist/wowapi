---
id: VER-W02-E05-S001
type: verification-record
parent_story: W02-E05-S001
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
---

# Verification record — W02-E05-S001

## Planned verification procedure

Per mandate §8.8. One row per acceptance criterion for this story.

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W02-E05-S001-01 | Inspect T001's design document against `plan.md`'s "Unresolved questions" list; confirm each has a documented decision + rationale dated before any T002–T005 implementation commit | Documentation review + git history | Every listed question resolved with rationale, predating implementation; any D-0N-caliber decision escalated | design-decision record / review report | unassigned |
| AC-W02-E05-S001-02 | Run the sync twice against the same database (repeat-run test); run dry-run against an unsynced database with a no-writes assertion | Local dev or CI, PostgreSQL instance | Second run converges with no spurious writes; dry-run produces a change plan with zero writes | integration-test report | unassigned |
| AC-W02-E05-S001-03 | Run the sync against a schema-valid and a schema-invalid manifest fixture; retrieve the recorded manifest version after a successful sync | Local dev or CI, PostgreSQL instance | Invalid manifest rejected before any write; applied version recorded and retrievable | schema-validation + integration-test report | unassigned |
| AC-W02-E05-S001-04 | Run the RLS-posture verification test per T001's documented role decision | Local dev or CI, PostgreSQL with RLS roles configured | Sync runs under the documented role; tenant-table RLS enforcement preserved; no undocumented superuser bypass | integration-test report | unassigned |
| AC-W02-E05-S001-05 | Boot prod-profile against an empty catalog DB before the fix (fail-first capture) and after: assert named readiness failure until sync, then ready with seed/catalog hash in the payload | Local dev or CI, PostgreSQL instance, prod-profile boot path | Before: silent ready (defect captured); after: named 503 until sync, then 200 with hash reported | fail-first/pass-after integration-test report pair | unassigned |
| AC-W02-E05-S001-06 | Run a sync (and a dry-run, per T001's dry-run-auditing decision) and assert the audit record's presence and required fields | Local dev or CI, PostgreSQL instance | Durable audit record per run with manifest version, hash, actor, outcome per T001's shape | integration-test report (audit-row assertion) | unassigned |

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
