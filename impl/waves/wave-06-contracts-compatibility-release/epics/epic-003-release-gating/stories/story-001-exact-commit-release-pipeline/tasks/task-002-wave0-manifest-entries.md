---
id: W06-E03-S001-T002
type: task
title: Wave-0 manifest entries
status: done
parent_story: W06-E03-S001
owner: W06E03Impl
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on:
  - W06-E03-S001-T001
acceptance_criteria:
  - AC-W06-E03-S001-02
artifacts:
  - ART-W06-E03-S001-002
evidence:
  - EV-W06-E03-S001-002
---

# W06-E03-S001-T002 — Wave-0 manifest entries

## Task Definition

### Task objective

Populate Wave-0 manifest entries mapping to today's workflow-lint/unit/gate/coverage/reference-smoke jobs + vuln.yml + REL-02's blocking scanners.

### Parent story

W06-E03-S001

### Owner

unassigned

### Status

todo

### Dependencies

W06-E03-S001-T001 (entries must conform to the schema before they can be added).

### Detailed work

1. Enumerate every job currently required for green ci-container.
2. Add a manifest entry for each, with completed_wave: 0.
3. Diff-review to confirm no existing check is silently dropped.

### Expected files or components affected

ci/release-gates.yaml (populated with Wave-0 entries).

### Expected output

A manifest whose entry count matches the current required-check count.

### Required artifacts

ART-W06-E03-S001-002 (Wave-0 manifest entries).

### Required evidence

EV-W06-E03-S001-002 (manifest-entry-count diff-review output).

### Related acceptance criteria

AC-W06-E03-S001-02

### Completion criteria

Every job currently required for green ci-container has a manifest entry.

### Verification method

Diff review comparing manifest entries against the current CI required-check set.

### Risks

Medium — must not silently drop an existing check, per PLAN T2's own risk note.

### Rollback or recovery considerations

If an entry is later found missing, add it and record why it was initially omitted in `deviations.md`.

## Implementation Record

Implemented the complete required-check catalog in `ci/release-gates.yaml`, including the REL-02 classes and per-gate evidence metadata. Cross-reference tests passed; evidence: EV-W06-E03-S001-002.
## Verification Record

Pass — focused catalog/schema/cross-reference tests; every enumerated required security class occurs exactly once. Revision `733ef3e930cbb3f89f5bbc53d8f562c60e426513`; EV-W06-E03-S001-002.
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
