---
id: VER-W06-E02-S001
type: verification-record
parent_story: W06-E02-S001
status: verified
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Verification record — W06-E02-S001

## Planned verification procedure

Per mandate §8.8. One row per acceptance criterion for this story.

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W06-E02-S001-01 | Run the fixture-driven per-field test suite (one fragment per OpenAPI 3.1 top-level/components.* field type) | Local dev or CI, Go toolchain | Every field either merges correctly per its documented policy or is explicitly rejected with a field-specific error | fixture test report | unassigned |
| AC-W06-E02-S001-02 | Validate the merged document against 3.1.1/2020-12 using the selected validator; run a malformed-output negative fixture | Local dev or CI, Go toolchain | Valid merged output passes; malformed output fails the command | structural-validation test report | unassigned |
| AC-W06-E02-S001-03 | Run the seeded intentional-breaking-change fixture through the semantic-diff gate | CI gate | The breaking-change fixture fails the gate | CI gate test report | unassigned |
| AC-W06-E02-S001-04 | Inspect the validator-dependency decision record for a security/licence review outcome predating its use as a hard dependency | Documentation review | A dated review record exists and predates the dependency's use | review report | unassigned |

## Post-execution record

Focused verification executed; raw output and metadata are in `evidence/openapi-focused-tests.txt`.

### Actual result

All four criteria passed local focused verification.

### Pass or fail

PASS.

### Evidence identifier

EV-W06-E02-S001-001 through EV-W06-E02-S001-004.

### Execution date

2026-07-13.

### Commit or revision

Working tree based on `733ef3e`.

### Environment

Darwin arm64; Go 1.26.5.

### Reviewer

W06-E01-E04-Execution.W06E02ReviewFinal — PASS, confidence 1.

### Findings

No verification failures after libopenapi-validator was wired; legacy OpenAPI expectations were updated to 3.1.1.

### Retest status

Focused retest PASS: three packages.

### Final conclusion

Verified and independently reviewed with no open issues; hosted CI evidence remains pending.
