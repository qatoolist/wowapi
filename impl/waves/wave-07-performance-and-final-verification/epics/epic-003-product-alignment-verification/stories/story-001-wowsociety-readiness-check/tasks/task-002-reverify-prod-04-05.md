---
id: W07-E03-S001-T002
type: task
title: Re-verify PROD-04/05's enabling capabilities and rollout plan
status: blocked
parent_story: W07-E03-S001
owner: W07-Phase-A-Execution.W07E03S001
created_at: 2026-07-12
updated_at: 2026-07-14
depends_on: []
acceptance_criteria:
  - AC-W07-E03-S001-04
  - AC-W07-E03-S001-05
artifacts:
  - ART-W07-E03-S001-001
evidence:
  - EV-W07-E03-S001-002
---

# W07-E03-S001-T002 — Re-verify PROD-04/05's enabling capabilities and rollout plan

## Task Definition

### Task objective

Directly re-verify SEC-01 T1/T5's grant contract and the coordinated rollout plan (PROD-04), and D-04's hash_version branch verification (PROD-05).

### Parent story

W07-E03-S001

### Owner

W07-Phase-A-Execution.W07E03S001

### Status

blocked

### Dependencies

W03-E01 (SEC-01), W04-E04/W00-E02 (D-04) must all be `accepted` — already satisfied by this wave's own entry gate.

### Detailed work

1. Inspect SEC-01 T1/T5's grant contract directly.
2. Confirm W03-E01-S004's own coordinated-rollout-plan artifact exists and is current, or document the
   gap.
3. Inspect D-04's hash_version branch-verification logic directly.
4. Confirm zero wowsociety-repository code change is performed anywhere in this task's own execution.

### Expected files or components affected

No wowapi source-code file changed, no wowsociety file touched; a documentation draft (feeding into T003's own consolidated record).

### Expected output

Confirmed existence of both enabling capabilities; confirmed or gap-noted rollout plan; confirmed zero wowsociety code change.

### Required artifacts

ART-W07-E03-S001-001 (draft coordination content, feeding into the consolidated record).

### Required evidence

EV-W07-E03-S001-002 (re-verification report, PROD-04/05).

### Related acceptance criteria

AC-W07-E03-S001-04, AC-W07-E03-S001-05.

### Completion criteria

Both capabilities are confirmed to genuinely exist; the rollout plan's status is honestly recorded; zero wowsociety code change confirmed.

### Verification method

Direct inspection of each capability's own current implementation, plus an explicit confirmation that no wowsociety repository was touched.

### Risks

RISK-W07-E03-001 (a documentation gap found in one of these two) — see epic-level `risks.md`.

### Rollback or recovery considerations

Not applicable — a re-verification finding is recorded, not rolled back; a found gap is fixed with a small documentation addition.

## Implementation Record

### What was actually implemented

Inspected SEC-01's migration, live schema/RLS/privileges/indexes, resolver and actor authority path;
ran the focused grant integration tests; and cross-checked all three W03-E01-S004 rollout documents
against that current behavior. Inspected and tested D-04's migration, v1/v2 dispatch, unknown-version
failure and per-field tamper behavior. Published the findings and product paths in the consolidated
artifact.

### Components changed

Only W07-E03-S001 documentation/evidence records.

### Files changed

`evidence/tests/EV-W07-E03-S001-002.md`, the consolidated artifact, and lifecycle records. No source
or wowsociety file was changed.

### Interfaces introduced or changed

None.

### Configuration changes

None.

### Schema or migration changes

None.

### Security changes

No implementation change. A security-relevant stale rollback/claim-authority plan was recorded as a
blocker rather than repeated as safe guidance.

### Observability changes

None.

### Tests added or modified

None; existing focused DB integration tests were executed.

### Commits

No commit created; verified revision `733ef3e930cbb3f89f5bbc53d8f562c60e426513`.

### Pull requests

None.

### Implementation dates

2026-07-14.

### Technical debt introduced

None.

### Known limitations

No product staging drill was run, by scope. The SEC-01 rollout document is not safe/current enough
for product execution and lacks wowsociety sign-off.

### Follow-up items

Correct and jointly review W03-E01-S004; then execute its product-side flow. Run and archive the
D-04 all-chain staging drill.

### Relationship to the approved plan

Matched `plan.md`; the plan explicitly required stale rollout material to be gap-noted.

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W07-E03-S001-04 | Live grant catalog, resolver tests, rollout-document cross-check | Required local PostgreSQL | Contract PASS; coordinated plan FAIL/stale | re-verification report | W05ReviewGateFinal |
| AC-W07-E03-S001-05 | Live column probe, v1/v2/unknown/tamper/manifest tests, scope audit | Required local PostgreSQL | PASS; zero wowsociety change | re-verification report | W05ReviewGateFinal |

### Actual result

SEC-01's framework contract passes. W03-E01-S004 contradicts the current schema, direct-claim
behavior and safe rollback boundary, so PROD-04 is blocked. D-04 passes and its staging path is
documented. No wowsociety repository was read or changed.

### Pass or fail

Partial FAIL: AC04 fails; AC05 passes.

### Evidence identifier

`EV-W07-E03-S001-002`.

### Execution date

2026-07-14.

### Commit or revision

`733ef3e930cbb3f89f5bbc53d8f562c60e426513`.

### Environment

Darwin arm64; Go 1.26.5; PostgreSQL 18.4 client; required local DB/S3 enforcement.

### Reviewer

`W05ReviewGateFinal` — no open package issue; upstream PROD-04 blocker confirmed.

### Findings

W03-E01-S004 has a wrong migration path, nonexistent schema columns, false compatibility assumptions,
an unsafe rollback premise, and incomplete product sign-off.

### Retest status

No retest can clear AC04 until the rollout artifact is corrected. D-04 needs only its out-of-scope
product staging drill.

### Final conclusion

Inspection work is complete, but this task remains blocked on PROD-04's rollout artifact.

## Deviations Record

No deviation. The stale-artifact branch was explicitly contemplated by the plan.

### Deviation ID

Not applicable.

### Approved plan

Verify the contract and either confirm or gap-note the rollout artifact.

### Actual implementation

Verified the contract and documented the rollout gap.

### Reason

Not applicable.

### Impact

Not applicable.

### Risks

RISK-W07-E03-001 is realized for PROD-04.

### Approval

Not applicable.

### Compensating controls

The product cutover remains blocked; direct claims are not proposed as a fallback.

### Follow-up work

Correct and independently re-review W03-E01-S004.
