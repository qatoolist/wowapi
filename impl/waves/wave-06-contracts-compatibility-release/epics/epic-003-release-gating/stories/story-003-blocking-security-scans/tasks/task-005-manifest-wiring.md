---
id: W06-E03-S003-T005
type: task
title: Manifest wiring
status: done
parent_story: W06-E03-S003
owner: W06E03Impl
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on:
  - W06-E03-S003-T001
  - W06-E03-S003-T002
  - W06-E03-S003-T003
  - W06-E03-S003-T004
acceptance_criteria:
  - AC-W06-E03-S003-05
artifacts:
  - ART-W06-E03-S003-005
evidence:
  - EV-W06-E03-S003-005
---

# W06-E03-S003-T005 — Manifest wiring

## Task Definition

### Task objective

Wire all four REL-02 blocking checks into W06-E03-S001's Wave-0 manifest, with a cross-reference test confirming exactly one entry per scanner class.

### Parent story

W06-E03-S003

### Owner

unassigned

### Status

todo

### Dependencies

W06-E03-S003-T001, W06-E03-S003-T002, W06-E03-S003-T003, W06-E03-S003-T004 (all four checks must exist first); cross-story on W06-E03-S001's T001/T002 (the manifest schema and Wave-0 entries must exist).

### Detailed work

1. Confirm W06-E03-S001's manifest schema and Wave-0 entries exist.
2. Add manifest entries for T1-T4's four checks.
3. Write a cross-reference test confirming every enumerated scanner class has exactly one manifest
   entry.

### Expected files or components affected

W06-E03-S001's ci/release-gates.yaml (extended with REL-02's entries).

### Expected output

Every REL-02 blocking check has exactly one manifest entry.

### Required artifacts

ART-W06-E03-S003-005 (manifest wiring).

### Required evidence

EV-W06-E03-S003-005 (cross-reference test report).

### Related acceptance criteria

AC-W06-E03-S003-05.

### Completion criteria

The cross-reference test confirms exactly one entry per scanner class.

### Verification method

Direct execution of the cross-reference test.

### Risks

Low, per PLAN T5's own risk classification.

### Rollback or recovery considerations

If an entry is later found duplicated or missing, correct the manifest and record why in `deviations.md`.

## Implementation Record

Wired exactly one gate entry per required security-result class and required blocking artifact/image security reports in release manifests and clean verification. Cross-reference/adversarial tests passed; evidence: EV-W06-E03-S003-005.
## Verification Record

Pass — gate catalog cross-reference rejects duplicates and release manifests reject missing artifact/image security reports. Revision `733ef3e930cbb3f89f5bbc53d8f562c60e426513`; EV-W06-E03-S003-005.
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
