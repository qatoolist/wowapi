---
id: W07-E04-S001-T003
type: task
title: Disposition audit (sampled)
status: todo
parent_story: W07-E04-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on: []
acceptance_criteria:
  - AC-W07-E04-S001-03
artifacts:
  - ART-W07-E04-S001-003
evidence:
  - EV-W07-E04-S001-003
---

# W07-E04-S001-T003 — Disposition audit (sampled)

## Task Definition

### Task objective

Spot-check a meaningful, weighted-toward-P0/critical sample of accepted story claims across the programme against their own actual evidence.

### Parent story

W07-E04-S001

### Owner

unassigned

### Status

todo

### Dependencies

W07-E01, W07-E02, W07-E03 must all be `accepted` first.

### Detailed work

1. Select a disposition-audit sample, weighted toward P0/critical-priority stories.
2. For each sampled story, independently re-check its own accepted claim against its evidence/index.md
   and closure.md.
3. Record whether each sampled claim is genuine or found deficient.

### Expected files or components affected

A new disposition-audit report (exact location TBD).

### Expected output

A sampled, evidenced confirmation of accepted-claim genuineness across the programme.

### Required artifacts

ART-W07-E04-S001-003 (the disposition-audit report).

### Required evidence

EV-W07-E04-S001-003 (the sampled-claim evidence trail).

### Related acceptance criteria

AC-W07-E04-S001-03.

### Completion criteria

Each sampled claim is independently re-checked, and its genuineness (or deficiency) is recorded.

### Verification method

Direct re-checking of each sampled story's own evidence and closure record.

### Risks

RISK-W07-003 (a genuine gap found in a sampled story's own claim) — see epic-level `risks.md`.

### Rollback or recovery considerations

Not applicable — if a sampled claim is found deficient, record the finding honestly and carry it forward, not silently correct the underlying story's own closure record without a proper deviation trail.

## Implementation Record

*Not yet implemented.*

### What was actually implemented

*Not yet implemented.*

### Components changed

*Not yet implemented.*

### Files changed

*Not yet implemented.*

### Interfaces introduced or changed

*Not yet implemented.*

### Configuration changes

*Not yet implemented.*

### Schema or migration changes

*Not applicable.*

### Security changes

*Not yet implemented.*

### Observability changes

*Not yet implemented.*

### Tests added or modified

*Not yet implemented.*

### Commits

*Not yet implemented.*

### Pull requests

*Not yet implemented.*

### Implementation dates

*Not yet implemented.*

### Technical debt introduced

*None anticipated.*

### Known limitations

*Not yet implemented.*

### Follow-up items

*Not yet implemented.*

### Relationship to the approved plan

*Not yet implemented. Once implementation occurs, record whether it matched `plan.md`; if not,
reference the corresponding entry in `deviations.md`.*

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W07-E04-S001-03 | Re-check each sampled claim | Documentation + evidence review | Each sampled claim genuinely re-checked | disposition-audit report | unassigned |

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

## Deviations Record

*No deviations recorded yet.*

### Deviation ID

*Not applicable.*

### Approved plan

*Not applicable.*

### Actual implementation

*Not applicable.*

### Reason

*Not applicable.*

### Impact

*Not applicable.*

### Risks

*Not applicable.*

### Approval

*Not applicable.*

### Compensating controls

*Not applicable.*

### Follow-up work

*Not applicable.*
