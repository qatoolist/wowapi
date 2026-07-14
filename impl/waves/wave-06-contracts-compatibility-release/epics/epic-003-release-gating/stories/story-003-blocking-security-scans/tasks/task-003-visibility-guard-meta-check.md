---
id: W06-E03-S003-T003
type: task
title: Visibility-guard regression meta-check
status: done
parent_story: W06-E03-S003
owner: W06E03Impl
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on: []
acceptance_criteria:
  - AC-W06-E03-S003-03
artifacts:
  - ART-W06-E03-S003-003
evidence:
  - EV-W06-E03-S003-003
---

# W06-E03-S003-T003 — Visibility-guard regression meta-check

## Task Definition

### Task objective

Build a meta-check asserting dependency-review/codeql/scorecard actually ran whenever the repository is public, as a regression safety net.

### Parent story

W06-E03-S003

### Owner

unassigned

### Status

todo

### Dependencies

None.

### Detailed work

1. Re-confirm the repository's current public visibility via a live gh api call.
2. Implement the meta-check asserting the three scanners actually ran (not merely configured) whenever
   the repository is public.
3. Test against a forced-private test branch to confirm the guard logic itself, not just current
   visibility.

### Expected files or components affected

A new visibility-guard meta-check workflow step.

### Expected output

A meta-check that fails if any of the three scanners did not run while the repository is public.

### Required artifacts

ART-W06-E03-S003-003 (visibility-guard regression meta-check).

### Required evidence

EV-W06-E03-S003-003 (forced-private-test-branch guard-regression test report).

### Related acceptance criteria

AC-W06-E03-S003-03.

### Completion criteria

The forced-private test branch confirms the guard logic itself.

### Verification method

Direct execution against a forced-private test branch.

### Risks

Low, per PLAN T3's own risk classification.

### Rollback or recovery considerations

If the guard produces false positives against a legitimate configuration, revise the guard logic.

## Implementation Record

Implemented public exact-SHA hosted scanner polling and fail-closed result validation, tag triggers, forced visibility tests, and explicit private fallback selection. Concurrent hosted runs are retried with bounded configuration. Evidence: EV-W06-E03-S003-003.
## Verification Record

Pass — forced public missing result fails, forced private selects fallback, and concurrent public hosted results are retried then accepted. Revision `733ef3e930cbb3f89f5bbc53d8f562c60e426513`; EV-W06-E03-S003-003.
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
