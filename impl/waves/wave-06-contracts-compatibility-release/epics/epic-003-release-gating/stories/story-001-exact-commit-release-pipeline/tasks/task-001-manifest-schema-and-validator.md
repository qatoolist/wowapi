---
id: W06-E03-S001-T001
type: task
title: Manifest schema and validator
status: done
parent_story: W06-E03-S001
owner: W06E03Impl
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on: []
acceptance_criteria:
  - AC-W06-E03-S001-01
artifacts:
  - ART-W06-E03-S001-001
evidence:
  - EV-W06-E03-S001-001
---

# W06-E03-S001-T001 — Manifest schema and validator

## Task Definition

### Task objective

Design ci/release-gates.yaml's manifest schema (ID, command/job ref, owner, required_from_wave, timeout, evidence-artifact path) and a JSON Schema validator rejecting a malformed entry.

### Parent story

W06-E03-S001

### Owner

unassigned

### Status

todo

### Dependencies

None.

### Detailed work

1. Design the manifest schema with the required fields.
2. Implement a JSON Schema validator.
3. Write malformed-vs-valid manifest fixture tests.

### Expected files or components affected

ci/release-gates.yaml schema definition; a JSON Schema validator (exact location TBD).

### Expected output

A schema and validator that rejects a manifest entry missing a required field.

### Required artifacts

ART-W06-E03-S001-001 (manifest schema + validator).

### Required evidence

EV-W06-E03-S001-001 (malformed-manifest-fixture test output).

### Related acceptance criteria

AC-W06-E03-S001-01

### Completion criteria

Schema rejects a manifest entry missing a required field.

### Verification method

Direct execution of the validator against malformed and valid fixtures.

### Risks

Low — pure config/tooling, per PLAN T1's own risk classification.

### Rollback or recovery considerations

Revert to the prior (nonexistent) state if the schema proves wrong; this is additive tooling with no runtime behavior to roll back.

## Implementation Record

Implemented `ci/release-gates.schema.json` and fail-closed validation in `scripts/validation/release_contract.py`. Focused malformed-field tests passed; evidence: EV-W06-E03-S001-001.
## Verification Record

Pass — `python3 -m unittest scripts.validation.tests.test_release_contracts`; malformed required field rejected. Revision `733ef3e930cbb3f89f5bbc53d8f562c60e426513`; EV-W06-E03-S001-001.
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
