---
id: VER-W02-E02-S001
type: verification-record
parent_story: W02-E02-S001
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
---

# Verification record — W02-E02-S001

## Planned verification procedure

Per mandate §8.8. One row per acceptance criterion for this story.

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W02-E02-S001-01 | Query `pg_indexes` against all 4 parent tables post-migration | Local dev environment or CI, PostgreSQL instance | `UNIQUE (tenant_id, id)` present on `parties`, `organizations`, `documents`, `document_versions` | migration test report (`pg_indexes` query output) | unassigned |
| AC-W02-E02-S001-02 | Run the catalog scanner against a fixture schema mirroring the 8 known edges | Local dev environment or CI, Go toolchain | Scanner enumerates exactly 8 FKs, zero silent gaps | fixture-schema test report | unassigned |
| AC-W02-E02-S001-03 | Run CI against a negative fixture migration adding a single-column tenant FK | CI | Build fails, citing the specific non-composite FK | CI run output (negative fixture) | unassigned |

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
