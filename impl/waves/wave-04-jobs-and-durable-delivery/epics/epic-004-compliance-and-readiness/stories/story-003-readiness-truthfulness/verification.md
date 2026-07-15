---
id: VER-W04-E04-S003
type: verification-record
parent_story: W04-E04-S003
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Verification record — W04-E04-S003

## Planned verification procedure

Per mandate §8.8. One row per acceptance criterion for this story.

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W04-E04-S003-01 | Boot the service against a stale-migrated database and query `/readyz` | Local dev environment or CI, PostgreSQL instance at a lagging migration version | `/readyz` returns 503 | stale-migration integration-test report | unassigned |
| AC-W04-E04-S003-02 | Boot the service against a correctly-migrated database and query `/readyz`, inspecting the full payload | Local dev environment or CI, PostgreSQL instance at expected migration version | Payload reports migration version, seed/rule hash, and model hash (if AR-01 available) | full-readiness-payload integration-test report | unassigned |
| AC-W04-E04-S003-03 | Run `config doctor` from a nested subdirectory and from outside the repo with `--project` | Local dev environment or CI, Go toolchain | Discovery succeeds in both cases; explicitly reports whether product validation ran | config-doctor discovery test report | unassigned |

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
