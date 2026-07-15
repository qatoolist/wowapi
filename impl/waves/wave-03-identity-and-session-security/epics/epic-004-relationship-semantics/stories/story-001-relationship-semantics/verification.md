---
id: VER-W03-E04-S001
type: verification-record
parent_story: W03-E04-S001
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
---

# Verification record — W03-E04-S001

## Planned verification procedure

Per mandate §8.8. One row per acceptance criterion for this story.

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W03-E04-S001-01 | Seed a party-subject edge; resolve an actor carrying a party through the post-SEC-01 principal model; run `Checker.Has` | Local dev or CI, testkit DB, W03-E01's principal model available and `accepted` | The previously-false evaluation is now correctly `true` | party-subject-edge test report | unassigned |
| AC-W03-E04-S001-02 | Run the subject-kind matrix test across every schema-enumerated `subject_kind`, including a deliberately unenumerated kind | Local dev or CI, testkit DB | Every enumerated kind has a correct evaluation branch; the unenumerated kind fails closed | subject-kind matrix test report | unassigned |
| AC-W03-E04-S001-03 | Run the mutation-governance test: create/revoke a relationship edge, assert ownership check, attribution (via DATA-06 T2's mechanism), audit-row write, and version bump | Local dev or CI, testkit DB, DATA-06 T2 (W02-E04-S001) landed | Ownership-checked, attributed, audited, and versioned mutation confirmed; cache-invalidation sub-criterion tested if W05-E04-S002 has landed, otherwise recorded as deferred-linked | mutation-governance test report (+ cache-invalidation test report if applicable) | unassigned |

## Post-execution record

*Fill in after verification is actually executed. Do not record results that were not actually
observed — in particular, do not record AC-W03-E04-S001-03's cache-invalidation sub-criterion as
verified unless W05-E04-S002 has actually landed and the test actually ran against it.*

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
