---
id: VER-W05-E01-S002
type: verification-record
parent_story: W05-E01-S002
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
---

# Verification record — W05-E01-S002

## Planned verification procedure

Per mandate §8.8. One row per acceptance criterion for this story.

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W05-E01-S002-01 | Run `resource_ownership_adversarial_test.go` and `rules_ownership_adversarial_test.go` | Local dev or CI, Go toolchain | Cross-module claim attempt fails even with a matching key prefix, for both registries | adversarial-test report | unassigned |
| AC-W05-E01-S002-02 | Run `authz_ownership_adversarial_test.go` | Local dev or CI, Go toolchain | Cross-module permission claim rejected at the registrar boundary | adversarial-test report | unassigned |
| AC-W05-E01-S002-03 | Run `full_declaration_class_matrix_test.go` | Local dev or CI, Go toolchain | Every declaration class fixture (one per class) rejects a cross-module claim | adversarial-test report (table-driven) | unassigned |

## Post-execution record

*Fill in after verification is actually executed. Do not record results that were not actually
observed.*

### Actual result

*Not yet executed.*

### Pass or fail

*Not yet executed.*

### Evidence identifier

*Not yet executed.*

### Execution date

*Not yet executed.*

### Commit or revision

*Not yet executed.*

### Environment

*Not yet executed.*

### Reviewer

*Not yet executed.*

### Findings

*Not yet executed.*

### Retest status

*Not yet executed.*

### Final conclusion

*Not yet executed.*
