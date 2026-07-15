---
id: W06-E03-S003-T004
type: task
title: Local-scanner fallback
status: done
parent_story: W06-E03-S003
owner: W06E03Impl
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on: []
acceptance_criteria:
  - AC-W06-E03-S003-04
artifacts:
  - ART-W06-E03-S003-004
evidence:
  - EV-W06-E03-S003-004
---

# W06-E03-S003-T004 — Local-scanner fallback

## Task Definition

### Task objective

Build a local SAST substitute + scorecard-equivalent fallback, auto-activating on guard.outputs.public == 'false', with documented coverage gap vs. CodeQL.

### Parent story

W06-E03-S003

### Owner

unassigned

### Status

todo

### Dependencies

None.

### Detailed work

1. Select a local SAST substitute tool and a scorecard-equivalent mechanism.
2. Implement the fallback workflow, auto-activating on guard.outputs.public == 'false'.
3. Write a seeded unsafe-pattern fixture, tested against a forced-private test branch, confirming the
   fallback catches it.
4. Document the fallback's coverage gap versus CodeQL, not claiming parity.

### Expected files or components affected

A new local-scanner fallback workflow.

### Expected output

A fallback that auto-activates on private-repo state and catches a seeded unsafe pattern, with an honest documented coverage gap.

### Required artifacts

ART-W06-E03-S003-004 (local-scanner fallback workflow).

### Required evidence

EV-W06-E03-S003-004 (seeded-SAST-fixture fallback test report).

### Related acceptance criteria

AC-W06-E03-S003-04.

### Completion criteria

The seeded pattern is caught by the fallback in a forced-private test branch; coverage gap documented.

### Verification method

Direct execution of the seeded fixture against a forced-private test branch.

### Risks

Medium — document coverage gap vs. CodeQL rather than claim parity, per PLAN T4's own risk note.

### Rollback or recovery considerations

If the fallback proves to have a larger coverage gap than initially documented, update the documentation honestly rather than silently narrowing the claimed scope.

## Implementation Record

Implemented auto-activating non-public actionlint, local SAST, repository-posture, govulncheck, and Trivy checks. Seeded unsafe shell/workflow fixtures fail; documentation does not claim CodeQL parity. Evidence: EV-W06-E03-S003-004.
## Verification Record

Pass — complete local fallback exited 0; seeded unsafe shell and unpinned workflow failed. Revision `733ef3e930cbb3f89f5bbc53d8f562c60e426513`; EV-W06-E03-S003-004.
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
