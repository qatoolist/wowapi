---
id: W07-E03-S001-T001
type: task
title: Re-verify PROD-01/02/03's enabling capabilities
status: blocked
parent_story: W07-E03-S001
owner: W07-Phase-A-Execution.W07E03S001
created_at: 2026-07-12
updated_at: 2026-07-14
depends_on: []
acceptance_criteria:
  - AC-W07-E03-S001-01
  - AC-W07-E03-S001-02
  - AC-W07-E03-S001-03
artifacts:
  - ART-W07-E03-S001-001
evidence:
  - EV-W07-E03-S001-001
  - EV-W07-E03-S001-003
  - EV-W07-E03-S001-004
---

# W07-E03-S001-T001 — Re-verify PROD-01/02/03's enabling capabilities

## Task Definition

### Task objective

Directly re-verify DATA-01 T1/DATA-09's protocol (PROD-01), FBL-01's forwarding shim (PROD-02), and DX-07 T1/FBL-09's template fixes (PROD-03).

### Parent story

W07-E03-S001

### Owner

W07-Phase-A-Execution.W07E03S001

### Status

blocked

### Dependencies

W02-E01/E02, W05-E05, W04-E04, W01-E03 must all be `accepted` — already satisfied by this wave's own entry gate.

### Detailed work

1. Inspect DATA-01 T1's migration and DATA-09's protocol tooling directly.
2. Inspect FBL-01's deprecated forwarding shim at kernel/mfa directly.
3. Inspect DX-07 T1's readiness check and FBL-09's template fixes directly.
4. Draft the product-upgrade-path documentation for each of the three items.

### Expected files or components affected

No wowapi source-code file changed; a documentation draft (feeding into T003's own consolidated record).

### Expected output

Confirmed existence of all three enabling capabilities; drafted upgrade-path documentation for each.

### Required artifacts

ART-W07-E03-S001-001 (draft coordination content, feeding into the consolidated record).

### Required evidence

EV-W07-E03-S001-001 (re-verification report, PROD-01/02/03).

### Related acceptance criteria

AC-W07-E03-S001-01, AC-W07-E03-S001-02, AC-W07-E03-S001-03.

### Completion criteria

All three capabilities are confirmed to genuinely exist via direct inspection.

### Verification method

Direct inspection of each capability's own current implementation.

### Risks

RISK-W07-E03-001 (a documentation gap found in one of these three) — see epic-level `risks.md`.

### Rollback or recovery considerations

Not applicable — a re-verification finding is recorded, not rolled back; a found gap is fixed with a small documentation addition.

## Implementation Record

### What was actually implemented

Directly inspected the DATA-01 parent-index migration and live `rule_versions` indexes, executed the
DATA-09 protocol suite, inspected and compiled the FBL-01 forwarding shim/canonical package, and
executed the DX-07/FBL-09 readiness/template/rendered-product checks. Drafted the exact status, gap,
owner path, and consumer steps now published in `ART-W07-E03-S001-001`.

### Components changed

Only this story's documentation, artifact, evidence, and lifecycle records.

### Files changed

`evidence/tests/EV-W07-E03-S001-001.md`, `-003.md`, `-004.md` and the consolidated artifact, plus
governance records. No source file was modified.

### Interfaces introduced or changed

None.

### Configuration changes

None.

### Schema or migration changes

None.

### Security changes

None.

### Observability changes

None.

### Tests added or modified

None; existing focused tests were run.

### Commits

No commit created; verified revision `733ef3e930cbb3f89f5bbc53d8f562c60e426513`.

### Pull requests

None.

### Implementation dates

2026-07-14.

### Technical debt introduced

None.

### Known limitations

The intended PROD-01 composite FK cannot be created until wowapi adds a unique parent key on
`rule_versions(tenant_id,id)`. The FBL-01 shim has no exact published removal release.

### Follow-up items

wowapi data/reliability adds the parent key through DATA-09; the FBL-01 owner publishes its removal
release. Reverify affected rows afterward.

### Relationship to the approved plan

Matched `plan.md`; no deviation.

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W07-E03-S001-01 | Source/live-catalog inspection plus tenant-FK, protocol and manifest tests | Required local PostgreSQL/MinIO | DATA-09 PASS; DATA-01 parent key FAIL | re-verification + failed/retest records | W05ReviewGateFinal |
| AC-W07-E03-S001-02 | Shim/canonical source and focused tests | Go 1.26.5 | PASS | re-verification report | W05ReviewGateFinal |
| AC-W07-E03-S001-03 | Readiness/scaffold tests and rendered-product compile | PostgreSQL + Go 1.26.5 | PASS | re-verification report | W05ReviewGateFinal |

### Actual result

PROD-02 and PROD-03 are ready for their documented product actions. PROD-01 is blocked because the
live/source catalog lacks `UNIQUE (tenant_id,id)` on `rule_versions`; DATA-09 itself passes.

### Pass or fail

Partial FAIL: AC01 fails; AC02 and AC03 pass.

### Evidence identifier

`EV-W07-E03-S001-001`, `EV-W07-E03-S001-003`, `EV-W07-E03-S001-004`.

### Execution date

2026-07-14.

### Commit or revision

`733ef3e930cbb3f89f5bbc53d8f562c60e426513`.

### Environment

Darwin arm64; Go 1.26.5; local PostgreSQL/MinIO; DB/S3 enforcement enabled.

### Reviewer

`W05ReviewGateFinal` — no open package issue; upstream PROD-01 blocker confirmed.

### Findings

One substantive blocker (PROD-01 parent key) and one non-blocking scheduling gap (FBL-01 sunset).

### Retest status

The initial PostgreSQL-unavailable failure is resolved by the passing exact-command retest. The
substantive parent-key failure awaits implementation.

### Final conclusion

Inspection work is complete, but the task remains blocked because its completion criterion requires
all three capabilities and PROD-01 is absent.

## Deviations Record

No deviation. Recording a genuine failed prerequisite was explicitly required by the plan.

### Deviation ID

Not applicable.

### Approved plan

Directly verify and document gaps.

### Actual implementation

Directly verified and documented the gap.

### Reason

Not applicable.

### Impact

Not applicable.

### Risks

RISK-W07-E03-001 is realized for PROD-01.

### Approval

Not applicable.

### Compensating controls

PROD-01 is blocked; no unsafe product workaround is recommended.

### Follow-up work

Add and reverify the parent key.
