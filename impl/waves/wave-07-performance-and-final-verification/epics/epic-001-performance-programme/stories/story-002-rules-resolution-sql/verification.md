---
id: VER-W07-E01-S002
type: verification-record
parent_story: W07-E01-S002
status: passed
created_at: 2026-07-12
updated_at: 2026-07-14
---

# Verification record — W07-E01-S002

## Planned verification procedure

Per mandate §8.8. One row per acceptance criterion for this story.

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W07-E01-S002-01 | Run the index-definition audit against actual migrations | Local dev, repository grep | Directive's indexing claim confirmed or refuted, before T1 design proceeds | audit report | unassigned |
| AC-W07-E01-S002-02 | Run result-parity unit tests against the old and new query implementations | Local dev or CI, Go toolchain | Single SQL statement preserves exact precedence semantics | unit test report | unassigned |
| AC-W07-E01-S002-03 | Run EXPLAIN (ANALYZE, BUFFERS) against the new query | Local dev or CI, real Postgres | Index access shown, not sequential scan | EXPLAIN fixture report | unassigned |
| AC-W07-E01-S002-04 | Inspect committed EXPLAIN fixtures for shallow/deep and low/high cardinality coverage | Documentation + fixture review | Fixtures exist for all 4 named cardinality combinations | fixture inventory report | unassigned |
| AC-W07-E01-S002-05 | Run the parametrized parity + SQL-count-constant tests across 3/10/50-level ancestries | Local dev or CI, real Postgres | Parity holds; SQL count constant across depths | parametrized test report | unassigned |
| AC-W07-E01-S002-06 | Re-run existing rule-update-visibility tests; inspect the published before/after report | Local dev or CI, real Postgres | No stale-read regression; comparison published | regression test report + published comparison | unassigned |

## Post-execution record

| Acceptance criterion | Actual result | Evidence | Status |
|---|---|---|---|
| AC-W07-E01-S002-01 | `00008_rules.sql` audit confirmed active-only indexing before any design/edit | EV-W07-E01-S002-001 | passed |
| AC-W07-E01-S002-02 | One live statement matched the legacy result for six precedence/history cases | EV-W07-E01-S002-002 | passed |
| AC-W07-E01-S002-03 | Current exclusion and new historical index used; no `rule_versions` seq scan in four plan pairs | EV-W07-E01-S002-003 | passed |
| AC-W07-E01-S002-04 | 4/4 shallow/deep × low/high real EXPLAIN fixtures generated | EV-W07-E01-S002-004 | passed |
| AC-W07-E01-S002-05 | Legacy 11/18/58 statements; set-based 8/8/8 at depths 3/10/50; parity held | EV-W07-E01-S002-005 | passed |
| AC-W07-E01-S002-06 | Live-update suite passed; validated DEC-Q9-honest report references accepted baseline hash | EV-W07-E01-S002-006, EV-W07-E01-S002-007 | passed |

### Commands executed

```text
DATABASE_URL=<local-postgres> WOWAPI_REQUIRE_DB=1 go test ./kernel/rules -count=1 -v
DATABASE_URL=<local-postgres> WOWAPI_REQUIRE_DB=1 go test ./migrations -count=1 -v
DATABASE_URL=<local-postgres> WOWAPI_REQUIRE_DB=1 WOWAPI_UPDATE_EXPLAIN_FIXTURES=1 go test ./kernel/rules -run TestIntegrationResolverExplainFixtures -count=1 -v
```

All commands exited 0 against PostgreSQL 16.14. Required DB execution was enforced; no material test
skipped.

### Execution date

2026-07-14.

### Commit or revision

Base `733ef3e930cbb3f89f5bbc53d8f562c60e426513` plus the explicitly recorded working-tree changes.

### Environment

Go 1.26.5 darwin/arm64; PostgreSQL 16.14 in the repository's local `postgres:16-alpine` service.
This is relative/container evidence, not the accepted linux/amd64 runner and not an absolute-SLO run.

### Reviewer

`W05ReviewGateFinal` — independent; authored no W07-E01-S002 change.

### Findings

No verification failure remains. DEC-Q9 remains open by design.

### Retest status

Focused rules and migrations packages passed after the final query/index shape.

### Final conclusion

PASS. All six story ACs are verified and the independent gate found no open actionable issue.
