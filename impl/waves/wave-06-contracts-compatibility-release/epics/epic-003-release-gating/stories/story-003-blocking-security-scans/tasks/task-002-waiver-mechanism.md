---
id: W06-E03-S003-T002
type: task
title: Waiver mechanism
status: done
parent_story: W06-E03-S003
owner: W06E03Impl
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on: []
acceptance_criteria:
  - AC-W06-E03-S003-02
artifacts:
  - ART-W06-E03-S003-002
evidence:
  - EV-W06-E03-S003-002
---

# W06-E03-S003-T002 — Waiver mechanism

## Task Definition

### Task objective

Build a reviewed waiver-allowlist file format with owner/rationale/expiry/remediation-link per entry, CI-validated.

### Parent story

W06-E03-S003

### Owner

unassigned

### Status

todo

### Dependencies

None.

### Detailed work

1. Design the waiver-schema file format (owner/rationale/expiry/remediation-link per entry).
2. Implement a CI validator.
3. Write well-formed/missing-field/expired fixture tests.

### Expected files or components affected

A new waiver-schema file (exact path TBD) and its CI validator.

### Expected output

A waiver mechanism where well-formed entries pass and missing-field/expired entries fail CI.

### Required artifacts

ART-W06-E03-S003-002 (waiver-schema file format + CI validator).

### Required evidence

EV-W06-E03-S003-002 (waiver-schema fixture test report).

### Related acceptance criteria

AC-W06-E03-S003-02.

### Completion criteria

Well-formed entries pass; missing-field and expired entries both fail.

### Verification method

Direct execution of the validator against all three fixture types.

### Risks

Low, per PLAN T2's own risk classification.

### Rollback or recovery considerations

If a legitimate waiver is rejected due to a schema bug, fix the schema/validator directly.

## Implementation Record

Implemented strict expiring waiver schema/registry validation and exact `.trivyignore.yaml` synchronization. Active scoped entries pass; malformed, expired, missing-field, and mismatched ignore entries fail. Evidence: EV-W06-E03-S003-002.
## Verification Record

Pass — scoped active waiver accepted; malformed, missing, expired, wrong-scope, and ignore-registry mismatch cases rejected. Revision `733ef3e930cbb3f89f5bbc53d8f562c60e426513`; EV-W06-E03-S003-002.
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
