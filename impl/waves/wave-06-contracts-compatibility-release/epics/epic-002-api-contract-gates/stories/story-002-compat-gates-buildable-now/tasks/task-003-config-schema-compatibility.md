---
id: W06-E02-S002-T003
type: task
title: Config schema compatibility
status: done
parent_story: W06-E02-S002
owner: W06E02Impl
created_at: 2026-07-12
updated_at: 2026-07-14
depends_on: []
acceptance_criteria:
  - AC-W06-E02-S002-03
artifacts:
  - ART-W06-E02-S002-003
evidence:
  - EV-W06-E02-S002-003
---

# W06-E02-S002-T003 — Config schema compatibility

## Task Definition

### Task objective

Build a config schema compatibility gate against kernel/config/schema.go; seeded breaking-config fixture fails, additive optional fields pass.

### Parent story

W06-E02-S002

### Owner

W06E02Impl

### Status

done

### Dependencies

None.

### Detailed work

1. Build the compatibility gate against kernel/config/schema.go as source of truth.
2. Write a seeded breaking-config fixture (field removal or type change) and confirm it fails the gate.
3. Write a generated fixture migration test confirming additive optional fields pass.

### Expected files or components affected

New CI workflow configuration and fixture files for the config-compat gate.

### Expected output

A gate failing breaking config changes and passing additive optional fields.

### Required artifacts

ART-W06-E02-S002-003 (config schema compatibility gate).

### Required evidence

EV-W06-E02-S002-003 (seeded breaking-config-fixture test report).

### Related acceptance criteria

AC-W06-E02-S002-03.

### Completion criteria

Breaking fixture fails; additive optional-field fixture passes.

### Verification method

Direct execution of the gate against both fixtures.

### Risks

None beyond standard tooling-integration risk.

### Rollback or recovery considerations

If the gate misclassifies a legitimate additive change as breaking, revise the classification logic.

## Implementation Record

Implemented and focused-test verified.

### What was actually implemented

Recursive JSON Schema comparison rejects removed fields, incompatible types, narrowed enums, new constraints, and newly required properties while permitting additive optional properties and relaxed required sets.

### Components changed

Compatibility library, CLI, adversarial schemas, tests, and reusable CI job.

### Files changed

`internal/compat/config_schema.go`; `internal/compat/config_schema_test.go`; `internal/compat/testdata/config-schema/`; `internal/compatcli/`; `cmd/compatcheck/`.

### Interfaces introduced or changed

Adds `compatcheck config --baseline FILE --current FILE`.

### Configuration changes

CI generates each schema from its own release source and compares oldest-supported to current.

### Schema or migration changes

*Not applicable.*

### Security changes

No runtime security change.

### Observability changes

None.

### Tests added or modified

Fixtures cover identical/additive/removal/type/enum/required changes; direction regression proves required-to-optional passes and optional-to-required fails.

### Commits

No commit; working tree based on `733ef3e`.

### Pull requests

None.

### Implementation dates

2026-07-13.

### Technical debt introduced

*None anticipated.*

### Known limitations

No known limitation; hosted CI remains authoritative for future refs and does not replace this evidence.

### Follow-up items

None.

### Relationship to the approved plan

Matches the approved generated-schema source-of-truth plan.

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W06-E02-S002-03 | Run the gate against both fixtures | CI | Breaking fixture fails; additive fixture passes | CI gate test report | unassigned |

### Actual result

Additive schemas passed; every seeded breaking schema failed at the expected property path.

### Pass or fail

PASS.

### Evidence identifier

EV-W06-E02-S002-003.

### Execution date

2026-07-13.

### Commit or revision

Working tree based on `733ef3e`.

### Environment

Darwin arm64; Go 1.26.5.

### Reviewer

W06-E02-S002-Rerun confirmed the required-direction regression and full compatibility fixtures PASS.

### Findings

No code-direction defect reproduced; explicit regression coverage added.

### Retest status

Focused required-direction retest PASS.

### Final conclusion

Implemented, independently verified, and accepted.

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
