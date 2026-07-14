---
id: IMPL-W07-E01-S004
type: implementation-record
parent_story: W07-E01-S004
status: complete
created_at: 2026-07-12
updated_at: 2026-07-14
---

# Implementation record — W07-E01-S004

## What was actually implemented

Framework document uploads now require a canonical lowercase-hex SHA-256 before
presigning and persist the same checksum in the pending session. S3 signs and
stores canonical checksum metadata; ordinary `Stat` uses object metadata only.
Legacy hashing is isolated behind an explicitly labeled, byte/time-bounded
repair capability. A cursor-based, context-cancellable backfill inventories,
skips, repairs, interrupts, and resumes without duplicate work.

Repair emits bounded-cardinality hit, byte, and duration observations through
the expanded metrics port and Prometheus adapter. Seven exact CS-16 hot-path
benchmarks and budgets were added and wired into `BENCH_PKGS`.

## Components and files changed

- `foundation/document`: checksum-required initiation and migrated upload tests.
- `kernel/storage`: optional checksum upload/repair capabilities and Memory support.
- `adapters/storage/s3`: signed checksum upload, metadata-only Stat, repair/backfill.
- `kernel/observability` and `adapters/metrics/prometheus`: histograms.
- `kernel/{database,jobs,outbox,workflow,auth,mfa,httpclient}`: seven benchmarks.
- `Makefile`, `bench-budgets.txt`, `perf/results/perf-05-*.json`.

## Interfaces introduced or changed

Added `storage.ChecksumUploader`, `storage.ChecksumRepairer`,
`Service.InitiateUploadChecksum`, checksum headers on `PresignedURL`, and
`Metrics.ObserveHistogram`. Optional storage capabilities preserve third-party
adapter source compatibility; framework upload initiation fails closed if the
required upload capability is absent.

## Configuration, schema, security, and observability

No schema, migration, or new operator configuration was required. Security now
requires canonical integrity metadata and prevents implicit body downloads.
Observability adds `storage_checksum_repair_hits_total`,
`storage_checksum_repair_bytes`, and
`storage_checksum_repair_duration_seconds`, labeled only by repair label.

## Tests and verification

Real MinIO tests cover canonical upload, zero-GET Stat, bounded repair, metrics,
and interrupt/resume backfill. Focused document/storage/metrics tests pass.
The seven targeted benchmarks execute against their real paths; the combined
`make bench-budget` gate passes against local PostgreSQL.

## Revision and dates

Implemented and verified 2026-07-14 in a working tree based on `733ef3e`.
No PR was created.

## Technical debt, limitations, and follow-up

No implementation debt was introduced. Production legacy-object cardinality was
not measured without production credentials. The accepted reference has no
before samples for the new benchmarks, so the published local run is explicitly
not like-for-like. Re-run at a source SHA in the accepted Linux/amd64 environment;
absolute SLO claims remain conditional on DEC-Q9.

## Relationship to the approved plan

Implementation matched `plan.md`; `deviations.md` remains empty.
