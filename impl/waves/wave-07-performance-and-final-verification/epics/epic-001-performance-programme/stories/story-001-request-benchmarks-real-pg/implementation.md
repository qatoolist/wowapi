---
id: IMPL-W07-E01-S001
type: implementation-record
parent_story: W07-E01-S001
status: implemented
created_at: 2026-07-12
updated_at: 2026-07-14
---

# Implementation record — W07-E01-S001

## What was actually implemented

A provisional Linux amd64 GitHub reference workflow, a full-field `perf/reference-v1.json`, deterministic 100-tenant fixtures, and a real-PostgreSQL request benchmark suite covering all six PERF-02 profiles across cold/warm and 1/10/100-tenant concurrency. The suite records p50/p95/p99, allocation samples, SQL count, bytes, pool/transaction/lock timing, six profile-specific query-plan hashes, and non-overlapping pool-wait/tx-setup/authz-query/handler-query/serialization/middleware attribution. It publishes a machine-readable initial container reference with explicit DEC-Q9 conditionality.

## Components changed

- `.github/workflows/perf-reference.yml`: reference-contract, pinned-container smoke/publication, and manually dispatched full 5m-warmup/15m-measurement/3-repeat profile jobs.
- `perf/reference-v1.json` and `perf/fixtures/request-workloads-v1.json`: policy and deterministic workload contract.
- `perf/requestbench/`: real request handlers, isolated real-PostgreSQL fixture, matrix runner, attribution, publication validator, and focused contract tests.
- `perf/results/request-reference-v1.{json,txt}`: pinned Linux/amd64 Go 1.26.5 + PostgreSQL 16.9 initial reference capture.

## Interfaces introduced or changed

No production API changed. Benchmark-only environment inputs are `PERF_REPORT`, `PERF_SOURCE_SHA`, `PERF_CONTAINER_IMAGE`, `PERF_POSTGRES_IMAGE`, `PERF_POSTGRES_VERSION`, `PERF_POSTGRES_CONFIG`, `PERF_NETWORK`, and optional `PERF_WARMUP_DURATION`.

## Configuration changes

A standalone performance workflow was added. Existing `ci.yml` coverage/race/fuzz logic was not modified. Images and PostgreSQL settings are digest/config pinned; raw and machine-readable outputs are source-SHA-addressed CI artifacts.

## Schema or migration changes

None. The suite uses the existing migrations through `testkit.NewDB` and drops only its isolated test database during test cleanup.

## Security changes

No RLS policy or runtime guard changed. The measured pool uses `SET ROLE app_rt`, `WithConnRLSGuard`, tenant/actor context binding, and asserts both `current_user=app_rt` and connection to the isolated fixture database before measuring.

## Observability changes

Benchmark-only pgx/OTel query spans count SQL statements; synchronized phase timers and pgxpool acquire deltas provide cost attribution; `pg_stat_activity` lock sampling records lock-wait observations.

## Tests added or modified

Focused contracts cover the full reference field set, fixture/matrix completeness, publication schema/environment/path, warmup configuration, per-profile plan hashes, seed authorization, RLS/runtime binding, non-overlapping attribution, and trace-batch isolation.

## Commits and pull requests

No commit or pull request was created. Evidence is pinned to the working tree based on entry SHA `1626b11` and to artifact checksums.

## Implementation dates

2026-07-13 through 2026-07-14.

## Technical debt introduced

None in production code. The full-duration reference capture is intentionally manual and expensive; CI shards it by profile.

## Known limitations

The committed result is an initial 1x pinned-container reference, not an absolute SLO and not a substitute for the manual full-duration capture. Absolute numeric ceilings remain pending DEC-Q9.

## Follow-up items

A performance/SRE lead may run the `full_reference` workflow and approve dedicated-runner absolute ceilings only after DEC-Q9 resolves.

## Relationship to the approved plan

T1-T5 were implemented as planned. The provisional runner is GitHub-hosted Linux amd64 and the publication is explicitly relative/container-only while DEC-Q9 remains open.
