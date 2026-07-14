---
id: VER-W07-E01-S001
type: verification-record
parent_story: W07-E01-S001
status: verified
created_at: 2026-07-12
updated_at: 2026-07-14
---

# Verification record — W07-E01-S001

## Acceptance matrix

| Acceptance criterion | Actual verification | Result | Evidence |
|---|---|---|---|
| AC-W07-E01-S001-01 | Full-field contract test plus `actionlint .github/workflows/perf-reference.yml` | PASS | EV-W07-E01-S001-001 |
| AC-W07-E01-S001-02 | Real-PostgreSQL request matrix and seed/auth/RLS checks with the mandated local environment | PASS: six profiles, real isolated DB, `app_rt`, RLS guard | EV-W07-E01-S001-002 |
| AC-W07-E01-S001-03 | Pinned Linux/amd64 container benchmark publication inspected for unique dimensions | PASS: 36/36 cells, including 100 tenants | EV-W07-E01-S001-003 |
| AC-W07-E01-S001-04 | Publication schema and result rows inspected | PASS: six separate cost components plus SQL/bytes/tx/lock/plan metrics | EV-W07-E01-S001-004 |
| AC-W07-E01-S001-05 | Pinned Go/PostgreSQL container capture and publication-contract test | PASS: ratios published; `absolute_slo_status=conditional-on-DEC-Q9` | EV-W07-E01-S001-005 |

## Exact focused command

`DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable WOWAPI_REQUIRE_DB=1 WOWAPI_REQUIRE_S3=1 go test ./perf/requestbench -run '^(TestReferenceV1FullFieldContract|TestWorkloadFixtureDefinesCompleteMatrix|TestPublicationRequiresCompleteMatrixAndAttribution|TestPublicationPathIsRepositoryRelative|TestReferenceWarmupDurationConfiguration|TestEachProfileHasARepresentativeQueryPlanHash|TestSeedMatchesReferenceDatasetCardinality|TestSeedGrantsRealPostgresAuthorization|TestCostAttributionIsNonOverlapping|TestSQLCountExcludesWarmupAndPriorBatches|TestMatrixContractHasSixVariantsPerProfile|TestRealPostgresRequestProfileMatrix)$' -count=1 -v`

Result: PASS, `ok github.com/qatoolist/wowapi/perf/requestbench`, focused output `artifact://2665`; required DB/S3 flags prevented meaningful skips.

## Container publication command

Pinned digest containers were joined on an isolated Docker bridge: PostgreSQL 16.9 with the reference config and Go 1.26.5 Linux/amd64 running `go test ./perf/requestbench -run '^$' -bench '^BenchmarkRealPostgresRequests$' -benchtime=1x -count=1 -benchmem`.

Result: PASS in 11.808s; 36 rows, 6 profiles, 6 distinct representative plan hashes. Raw output: `perf/results/request-reference-v1.txt`; machine publication: `perf/results/request-reference-v1.json`.

## Workflow verification

`actionlint .github/workflows/perf-reference.yml` — PASS (no output). The coverage/race/fuzz executor confirmed over IRC that its changes are confined to `ci.yml`; this story changed only the separate performance workflow.

## Environment

- Focused contracts: Go 1.26.5 darwin/arm64, PostgreSQL 16.14, exact required DB/S3 environment flags.
- Publication: Go 1.26.5 Linux/amd64 image digest `sha256:5ab8e9…f85316`; PostgreSQL 16.9 image digest `sha256:ef463f9…852257a`; pinned server config and Docker bridge recorded in the JSON.
- Revision: working tree based on entry SHA `1626b11`; exact artifact checksums are in the five evidence records.

## Findings and retest status

Early red tests exposed fixture-DSN misbinding, delayed trace leakage, repository-relative report paths, missing profile-specific plan hashes, missing observed PostgreSQL metadata, and a reference/seed mismatch (1 resource per tenant vs the declared 10). Each was fixed at source; `TestSeedMatchesReferenceDatasetCardinality` now guards the last issue. The full focused suite and corrected pinned-container capture passed. No production RLS or authorization implementation changed.

## Reviewer

Independent Review Gate: PASS after 2 iterations. Iteration 1 reported no external finding; the executor's mandatory one-pass check then found the **Medium** dataset-cardinality mismatch (impact: unrepresentative reference data), fixed it with deterministic 10-resource seeding and a real-DB contract test, and reran all focused/container evidence. `W05ReviewGateFinal` re-reviewed the fix and reported no open actionable issue.

## Final conclusion

AC-01 through AC-05 are accepted within the provisional relative/container scope. Absolute SLO acceptance remains intentionally conditional on DEC-Q9.
