# Architecture review evidence — 2026-07-11

## Identity and environment

- Review ID: `architecture-review-2026-07-11`
- Reviewed commit: `d3c2640dbe1a0fe27e826cdf053945c4f49bc034`
- Recorded at: `2026-07-11T06:04:16Z` (`2026-07-11T11:34:16+0530`, IST)
- Last content validation: `2026-07-11T06:44:41Z` (`2026-07-11T12:14:41+0530`, IST)
- Host: Apple M3 Max, arm64, Darwin 25.5.0
- Go: `go1.26.5 darwin/arm64`
- Docker client: `29.4.0`
- Repository state before review artifacts: clean `main`, with no divergence from the locally available `origin/main` tracking ref
- Review artifacts are documentation only; no application, kernel, module, migration, workflow, or test source was changed.

This record distinguishes a command that returned successfully from a suite that necessarily exercised every optional integration. Go may serve package test results from its cache; targeted migration and S3 drills below therefore use `-count=1`. The general container gate forces Postgres tests but does not force MinIO tests.

## Command ledger

| ID   | Command                                                                                                                                                          | Exit | Result                                                                                                                                                                                                      |
| ---- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---: | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| E-01 | `scripts/graphify_refresh.sh check`                                                                                                                              |    0 | The checked-in graph corpus matched the current source revision.                                                                                                                                            |
| E-02 | `make ci`                                                                                                                                                        |    0 | Vet, boundary lint, lifecycle lint, unit tests, race tests, benchmark budget, and build passed. Some package tests were cache-served.                                                                       |
| E-03 | `make ci-container`                                                                                                                                              |    0 | Authoritative toolbox gate passed with `WOWAPI_REQUIRE_DB=1`; Postgres-backed tests were required rather than allowed to skip. This target does not set `WOWAPI_REQUIRE_S3=1`.                              |
| E-04 | `make lint-new`                                                                                                                                                  |    0 | `golangci-lint`: zero issues.                                                                                                                                                                               |
| E-05 | `miscellaneous/check_migrations.sh`                                                                                                                              |    0 | Thirty migrations were registered and contiguous and had the repository's Up/Down markers. This static script does not execute the rollback path.                                                           |
| E-06 | `docker compose -f deployments/compose.yaml run --rm -e WOWAPI_REQUIRE_DB=1 tools go test ./migrations -run '^TestIntegrationMigrationsReversible$' -count=1 -v` |    0 | Uncached disposable down/up reconstruction passed.                                                                                                                                                          |
| E-07 | `miscellaneous/check_test_skips.sh`                                                                                                                              |    0 | Exactly 22 explicit skip sites were reported; inventory follows below.                                                                                                                                      |
| E-08 | `docker compose -f deployments/compose.yaml run --rm -e WOWAPI_REQUIRE_S3=1 tools go test ./adapters/storage/s3 -count=1 -v`                                     |    1 | Expected diagnostic failure: tests used their `localhost:9000` fallback because the toolbox exports `S3_ENDPOINT`, while the tests read `S3_TEST_ENDPOINT`. Forced mode failed closed rather than skipping. |
| E-09 | `docker compose -f deployments/compose.yaml run --rm -e WOWAPI_REQUIRE_S3=1 -e S3_TEST_ENDPOINT=minio:9000 tools go test ./adapters/storage/s3 -count=1 -v`      |    0 | Uncached MinIO contract, presign, checksum, document round-trip, bucket-provisioning, and concurrency tests passed.                                                                                         |
| E-10 | `go test ./kernel/httpx -run '^$' -bench '^BenchmarkDispatch$' -benchmem -count=3`                                                                               |    0 | Three uncached host iterations passed; raw output follows below.                                                                                                                                            |
| E-11 | `go build ./...`                                                                                                                                                 |    0 | All packages built.                                                                                                                                                                                         |
| E-12 | `go vet ./...`                                                                                                                                                   |    0 | Vet passed.                                                                                                                                                                                                 |
| E-13 | Source-citation resolver over the directive                                                                                                                      |    0 | All 124 distinct backtick source citations resolved to existing files and in-range lines.                                                                                                                   |
| E-14 | `npx --yes prettier --check` over the three review artifacts                                                                                                     |    0 | Markdown and JSON formatting passed.                                                                                                                                                                        |
| E-15 | `jq empty evidence.json` plus a trailing-whitespace scan over all three artifacts                                                                                |    0 | JSON parsed and no trailing whitespace was found.                                                                                                                                                           |
| E-16 | Final post-edit source-citation, Prettier, JSON, trailing-whitespace, and Graphify-current checks                                                                |    0 | Latest directive and evidence content passed; all 124 source citations resolved and the graph freshness check remained clean.                                                                               |
| E-17 | Finding-heading/closure-row set equality validator                                                                                                               |    0 | Exactly 38 finding headings and 38 unique matching §13.2 closure contracts were present, with no missing, duplicate, or extra IDs.                                                                          |

The E-08 failure is evidence of a real integration-gate wiring gap, not a failed final validation. E-09 supplies the missing test endpoint explicitly and proves the adapter path against the running MinIO service. The architectural directive records the gap under REL-04.

## Independent review gate

- Completed: `2026-07-11T06:43:13Z` (`2026-07-11T12:13:13+0530`, IST)
- Reviewer: fresh low-cost, high-reasoning, read-only independent agent
- Scope: factual support, severity, ownership/DSL design, finding-level closure completeness, wave/release/migration/webhook consistency, B11/B12/B13 decisions, and evidence semantics
- Result: **PASS**
- Open findings against the deliverable: Critical 0, High 0, Medium 0, Low 0

The reviewer explicitly treated unresolved implementation defects and absent future roadmap artifacts as the subject of the review, not as failures of the review deliverable. It confirmed all 38 closure contracts, the owner-bound typed-port enforcement locus, DSL invariants, Wave 0/later-wave separation, release and online-migration choreography, parked-item consistency, reviewed-SHA semantics, and the 124-citation validation record.

## Uncached migration drill

```text
=== RUN   TestIntegrationMigrationsReversible
--- PASS: TestIntegrationMigrationsReversible (0.70s)
PASS
ok  github.com/qatoolist/wowapi/v2/migrations  0.707s
```

## Uncached S3/MinIO drill

The first forced invocation failed closed with the following repeated cause:

```text
WOWAPI_REQUIRE_S3=1 but S3/minio unreachable at localhost:9000:
dial tcp [::1]:9000: connect: connection refused
FAIL
FAIL github.com/qatoolist/wowapi/v2/adapters/storage/s3 0.004s
```

With the test-specific endpoint bound to the Compose service, the integration-bearing cases executed and passed:

```text
=== RUN   TestContract_S3
--- PASS: TestContract_S3 (0.02s)
=== RUN   TestDocument_UploadRoundTrip_S3
--- PASS: TestDocument_UploadRoundTrip_S3 (0.12s)
=== RUN   TestS3_PresignedRoundTrip
--- PASS: TestS3_PresignedRoundTrip (0.01s)
=== RUN   TestS3_WrongChecksum_StatReportsTrueBytes
--- PASS: TestS3_WrongChecksum_StatReportsTrueBytes (0.00s)
=== RUN   TestS3_MissingObject_IsKindNotFound
--- PASS: TestS3_MissingObject_IsKindNotFound (0.00s)
=== RUN   TestS3_PresignTTL_ConfiguredExpiryClamp
--- PASS: TestS3_PresignTTL_ConfiguredExpiryClamp (0.00s)
=== RUN   TestS3_EmptyObject
--- PASS: TestS3_EmptyObject (0.00s)
=== RUN   TestS3_New_MissingBucketFailsClosedWithoutCreateBucket
--- PASS: TestS3_New_MissingBucketFailsClosedWithoutCreateBucket (0.00s)
=== RUN   TestS3_New_CreateBucketProvisionsFreshBucket
--- PASS: TestS3_New_CreateBucketProvisionsFreshBucket (0.01s)
=== RUN   TestS3_New_ConcurrentCreateBucketRaceIsTolerated
--- PASS: TestS3_New_ConcurrentCreateBucketRaceIsTolerated (0.04s)
=== RUN   TestS3_Peek_NonPositiveN
--- PASS: TestS3_Peek_NonPositiveN (0.01s)
=== RUN   TestS3_ConcurrentRoundTrips
--- PASS: TestS3_ConcurrentRoundTrips (0.01s)
PASS
ok  github.com/qatoolist/wowapi/v2/adapters/storage/s3  0.231s
```

## Dispatch benchmark

Raw host output:

```text
goos: darwin
goarch: arm64
pkg: github.com/qatoolist/wowapi/v2/kernel/httpx
cpu: Apple M3 Max
BenchmarkDispatch/50routes-16          2032887    588.3 ns/op    1194 B/op    14 allocs/op
BenchmarkDispatch/50routes-16          2014862    585.3 ns/op    1194 B/op    14 allocs/op
BenchmarkDispatch/50routes-16          2059642    595.8 ns/op    1194 B/op    14 allocs/op
BenchmarkDispatch/500routes-16         1959051    612.7 ns/op    1194 B/op    14 allocs/op
BenchmarkDispatch/500routes-16         1966822    609.1 ns/op    1194 B/op    14 allocs/op
BenchmarkDispatch/500routes-16         1957287    620.1 ns/op    1194 B/op    14 allocs/op
BenchmarkDispatch/2000routes-16        1858754    641.9 ns/op    1194 B/op    14 allocs/op
BenchmarkDispatch/2000routes-16        1897162    641.1 ns/op    1194 B/op    14 allocs/op
BenchmarkDispatch/2000routes-16        1858645    638.4 ns/op    1194 B/op    14 allocs/op
PASS
ok  github.com/qatoolist/wowapi/v2/kernel/httpx  16.965s
```

Median host values are 588.3, 612.7, and 641.1 ns/op for 50, 500, and 2,000 routes respectively, with 1,194 B/op and 14 allocations/op at every size. The 40-fold route-count increase moved the median by approximately 9.0%.

The single benchmark-budget invocation inside the successful toolbox gate reported:

```text
BenchmarkDispatch/50routes       1000.0 ns/op    1194 B/op    14 allocs/op
BenchmarkDispatch/500routes       990.3 ns/op    1194 B/op    14 allocs/op
BenchmarkDispatch/2000routes      933.0 ns/op    1194 B/op    14 allocs/op
```

These results support keeping B11 parked. They do not establish an end-to-end request SLO because `BenchmarkDispatch` deliberately replaces authentication, authorization, and transaction work with fakes.

## Skip inventory

`miscellaneous/check_test_skips.sh` found 22 sites:

| Area                                |  Sites | Reason class                                                      |
| ----------------------------------- | -----: | ----------------------------------------------------------------- |
| CLI scaffolding/i18n/config/migrate |      6 | Short-mode compilation or generated-product execution             |
| CLI DB helper                       |      1 | Database DSN absent                                               |
| End-to-end CLI                      |      3 | Go toolchain or cold-cache network prerequisite absent            |
| Build information                   |      1 | Root execution makes permission branch inapplicable               |
| Testkit DB environment              |      1 | Database DSN absent                                               |
| Scratch consumer                    |      3 | Database, Go toolchain, or cold-cache network prerequisite absent |
| S3 adapter                          |      3 | MinIO unavailable unless forced                                   |
| Outbox coverage                     |      1 | Database DSN absent                                               |
| RLS guard                           |      2 | Database or role topology prerequisite absent                     |
| MFA/TOTP                            |      1 | Randomly generated code happens to be `000000`                    |
| **Total**                           | **22** |                                                                   |

The inventory justifies a fail-closed, machine-readable integration-suite manifest. In particular, the random TOTP skip should be replaced by deterministic time and secret fixtures.

## Evidence boundaries

- No production traffic, cardinality distribution, query plan, or SLO trace was available. Performance findings beyond the router microbenchmark are complexity/lock/query-shape risks, not measured production regressions.
- No external penetration test, deployment IAM configuration, GitHub branch/tag protection configuration, or disaster-recovery execution record was available.
- This review did not mutate production code to test a proposed architecture. Blueprint acceptance tests are specified in the directive and remain future implementation work.
- B11, B12, and B13 were rechecked against the live source and remain parked. Their reopen conditions are preserved in the directive.
