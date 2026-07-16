---
id: W07-E01-S004-EVIDENCE-INDEX
type: evidence-index
parent_story: W07-E01-S004
status: verified
derived: false
created_at: 2026-07-12
updated_at: 2026-07-14
---

# W07-E01-S004 — Evidence index

Per mandate §10. Commands were executed against local PostgreSQL and MinIO
service containers from a working tree based on `733ef3e`.

| Evidence ID | Type | Task | Acceptance criteria proven | Execution command | Commit SHA | Result | Status | Re-pin addendum (H-8/R-6) |
|---|---|---|---|---|---|---|---|---|
| EV-W07-E01-S004-001 | integration test report (no GetObject on Stat) | W07-E01-S004-T001 | AC-W07-E01-S004-01 | `WOWAPI_REQUIRE_S3=1 go test ./adapters/storage/s3 -count=1` | working tree based on `733ef3e` | PASS: canonical upload Stat issued zero GetObject requests | produced | `repin/EV-W07-E01-S004-001-repin-2026-07-16.json` |
| EV-W07-E01-S004-002 | labeled-repair-path test report | W07-E01-S004-T002 | AC-W07-E01-S004-02 | `WOWAPI_REQUIRE_S3=1 go test ./adapters/storage/s3 -count=1` | working tree based on `733ef3e` | PASS: unlabeled and oversized repair rejected; labeled bounded repair succeeded | produced | `repin/EV-W07-E01-S004-002-repin-2026-07-16.json` |
| EV-W07-E01-S004-003 | metric-emission test report | W07-E01-S004-T003 | AC-W07-E01-S004-03 | `WOWAPI_REQUIRE_S3=1 go test ./adapters/storage/s3 -count=1` | working tree based on `733ef3e` | PASS: hit counter plus byte and duration histograms observed | produced | `repin/EV-W07-E01-S004-003-repin-2026-07-16.json` |
| EV-W07-E01-S004-004 | backfill interrupt/resume test report | W07-E01-S004-T004 | AC-W07-E01-S004-04 | `WOWAPI_REQUIRE_S3=1 go test ./adapters/storage/s3 -count=1` | working tree based on `733ef3e` | PASS: three real MinIO legacy objects repaired across interrupt/resume without duplicates | produced | `repin/EV-W07-E01-S004-004-repin-2026-07-16.json` |
| EV-W07-E01-S004-005 | published report | W07-E01-S004-T005 | AC-W07-E01-S004-05 | inspect `perf/results/perf-05-comparison-v1.json` against `perf/reference-v1.json` | working tree based on `733ef3e` | PASS: measurements published; comparison explicitly not like-for-like and absolute SLO remains conditional on DEC-Q9 | produced | `repin/EV-W07-E01-S004-005-repin-2026-07-16.json` |
| EV-W07-E01-S004-006 | benchmark report (per package) | W07-E01-S004-T006 | AC-W07-E01-S004-06 | `go test -v` on the seven named packages with exact seven-benchmark regex, `-benchtime=10x -benchmem -count=1` | working tree based on `733ef3e` | PASS: all seven benchmarks executed with ns/op, B/op, allocs/op | produced | `repin/EV-W07-E01-S004-006-repin-2026-07-16.json` |
| EV-W07-E01-S004-007 | make bench-budget passing output | W07-E01-S004-T006 | AC-W07-E01-S004-07 | `DATABASE_URL=<local> WOWAPI_REQUIRE_DB=1 make bench-budget` | working tree based on `733ef3e` | PASS: all budgeted benchmarks, including all seven CS-16 additions | produced | `repin/EV-W07-E01-S004-007-repin-2026-07-16.json` |

Fresh passing evidence uses status `produced`; retained failed or superseded
records continue to use the mandate §10 replacement vocabulary.

**Revision re-pin note (autopsy finding H-8, remediation R-6, 2026-07-16):** the `733ef3e` pin above
is not an ancestor of current HEAD (a side effect of the e8cda6b squash). These records were
originally captured inline in this index only (no separate per-record evidence file existed). Per
`impl/governance/evidence-policy.md`'s revision-pinning rule, each row above has now been re-run
against current HEAD (`43b6e128672f0b0997adcebc92703884deba5684`); the results are recorded, without
overwriting these original rows, in new sibling files under `evidence/repin/` (status `retested`).
EV-W07-E01-S004-006's addendum notes that the original's cited `-benchtime=10x` was capped to
`-benchtime=2x` for this re-run per this remediation's bench-time cap policy. No divergence from the
original claims was found.
