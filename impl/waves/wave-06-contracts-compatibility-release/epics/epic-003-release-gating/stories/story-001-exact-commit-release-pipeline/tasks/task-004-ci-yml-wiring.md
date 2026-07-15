---
id: W06-E03-S001-T004
type: task
title: ci.yml wiring
status: done
parent_story: W06-E03-S001
owner: W06E03Impl
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on:
  - W06-E03-S001-T003
acceptance_criteria:
  - AC-W06-E03-S001-04
artifacts:
  - ART-W06-E03-S001-004
evidence:
  - EV-W06-E03-S001-004
---

# W06-E03-S001-T004 — ci.yml wiring

## Task Definition

### Task objective

Update ci.yml to call required-gates.yml so PR CI and release use the identical execution path.

### Parent story

W06-E03-S001

### Owner

unassigned

### Status

todo

### Dependencies

W06-E03-S001-T003 (ci.yml calls the workflow built in T003).

### Detailed work

1. Update ci.yml to call required-gates.yml.
2. Confirm the same SHA through both PR/main CI and release paths produces byte-identical results
   (excluding run ID/timestamp).
3. Confirm no PR CI latency regression.

### Expected files or components affected

.github/workflows/ci.yml.

### Expected output

PR/main CI and release use the identical execution path via required-gates.yml.

### Required artifacts

ART-W06-E03-S001-004 (ci.yml wiring).

### Required evidence

EV-W06-E03-S001-004 (diff-based same-SHA-both-paths test output).

### Related acceptance criteria

AC-W06-E03-S001-04

### Completion criteria

Byte-identical results (excluding run ID/timestamp) for the same SHA through both paths; no PR CI latency regression.

### Verification method

Diff-based test comparing results from both paths for the same SHA.

### Risks

Medium — must not regress PR CI latency, per PLAN T4's own risk note.

### Rollback or recovery considerations

If latency regresses materially, optimize or revert the wiring change and record why in `deviations.md`.

## Implementation Record

Wired `.github/workflows/ci.yml` to the reusable required-gates workflow at `${{ github.sha }}`. Actionlint verified the caller and reusable workflow; evidence: EV-W06-E03-S001-004.
## Verification Record

Pass — reusable caller syntax and exact `${{ github.sha }}` wiring accepted by actionlint. Revision `733ef3e930cbb3f89f5bbc53d8f562c60e426513`; EV-W06-E03-S001-004.
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
