---
id: W06-E03-S001-T005
type: task
title: release.yml verify job
status: done
parent_story: W06-E03-S001
owner: W06E03Impl
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on:
  - W06-E03-S001-T004
acceptance_criteria:
  - AC-W06-E03-S001-05
artifacts:
  - ART-W06-E03-S001-005
evidence:
  - EV-W06-E03-S001-005
---

# W06-E03-S001-T005 — release.yml verify job

## Task Definition

### Task objective

Add a verify job to release.yml calling required-gates.yml with the tag event's exact SHA, never trusting a same-named check on another ref.

### Parent story

W06-E03-S001

### Owner

unassigned

### Status

todo

### Dependencies

W06-E03-S001-T004 (the verify job calls the same required-gates.yml wiring).

### Detailed work

1. Add the verify job to release.yml, calling required-gates.yml with the tag event's exact SHA.
2. Confirm the checked-out SHA equals the tag's target commit.
3. Write a seeded-failure fixture: tag a commit with a deliberately broken test.
4. Confirm verify fails and build-candidate never runs.

### Expected files or components affected

.github/workflows/release.yml.

### Expected output

A verify job that fails on a deliberately broken tagged commit, blocking build-candidate.

### Required artifacts

ART-W06-E03-S001-005 (release.yml verify job).

### Required evidence

EV-W06-E03-S001-005 (seeded-failure tag test output).

### Related acceptance criteria

AC-W06-E03-S001-05

### Completion criteria

The seeded-failure fixture proves verify fails and build-candidate never runs.

### Verification method

Direct execution against a scratch/throwaway repo with a deliberately broken tagged commit.

### Risks

High — this is the core trust boundary, per PLAN T5's own risk note.

### Rollback or recovery considerations

If verify can be bypassed under any condition, treat as a critical defect and escalate immediately — this is the pipeline's core trust boundary.

## Implementation Record

Implemented the release exact-tag-SHA verifier, attested required-gate caller, and exact-SHA hosted security barrier. `build-candidate` depends on both; moving-tag and failed-gate fixtures passed. Evidence: EV-W06-E03-S001-005.
## Verification Record

Pass — temporary Git repository moving-tag fixture rejected and failed gates block build. Revision `733ef3e930cbb3f89f5bbc53d8f562c60e426513`; EV-W06-E03-S001-005.
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
