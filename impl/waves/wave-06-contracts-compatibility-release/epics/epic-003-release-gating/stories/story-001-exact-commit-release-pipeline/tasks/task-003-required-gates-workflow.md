---
id: W06-E03-S001-T003
type: task
title: required-gates.yml reusable workflow
status: done
parent_story: W06-E03-S001
owner: W06E03Impl
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on:
  - W06-E03-S001-T002
acceptance_criteria:
  - AC-W06-E03-S001-03
artifacts:
  - ART-W06-E03-S001-003
evidence:
  - EV-W06-E03-S001-003
---

# W06-E03-S001-T003 — required-gates.yml reusable workflow

## Task Definition

### Task objective

Build required-gates.yml (workflow_call, parameterized on SHA), emitting attested gate-results.json.

### Parent story

W06-E03-S001

### Owner

unassigned

### Status

todo

### Dependencies

W06-E03-S001-T002 (the workflow consumes the populated manifest).

### Detailed work

1. Build required-gates.yml as a workflow_call workflow, parameterized on SHA.
2. Ensure exact-SHA checkout, not branch HEAD.
3. Emit attested gate-results.json, with each manifest entry individually reported.
4. Write a seeded-failure fixture through the workflow, confirming attested failure.

### Expected files or components affected

.github/workflows/required-gates.yml.

### Expected output

A reusable workflow producing an attested gate-results.json for any given SHA.

### Required artifacts

ART-W06-E03-S001-003 (required-gates.yml).

### Required evidence

EV-W06-E03-S001-003 (seeded-failure gate-results attestation output).

### Related acceptance criteria

AC-W06-E03-S001-03

### Completion criteria

A seeded failing entry produces an attested failure in gate-results.json.

### Verification method

Direct execution of the workflow against a seeded-failure fixture.

### Risks

Medium — must guarantee exact-SHA checkout, not branch HEAD, per PLAN T3's own risk note.

### Rollback or recovery considerations

If exact-SHA checkout proves unreliable, escalate rather than silently falling back to branch HEAD.

## Implementation Record

Implemented `.github/workflows/required-gates.yml`: full-SHA checkout, isolated matrix execution, complete failure results, deterministic aggregation, artifact retention, GitHub provenance attestation, and fail-closed summary. Evidence: EV-W06-E03-S001-003.
## Verification Record

Pass — seeded failed exact-SHA gate emitted failure and blocked candidate creation; actionlint passed. Revision `733ef3e930cbb3f89f5bbc53d8f562c60e426513`; EV-W06-E03-S001-003.
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
