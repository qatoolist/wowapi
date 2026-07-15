---
id: TRACK-SOURCE-TRACE-MATRIX
type: matrix
title: Source traceability matrix — source document+section to planned implementation item
status: active
created_at: 2026-07-12
updated_at: 2026-07-12
derived: true
---

# Source traceability matrix

DERIVED VIEW. Per mandate §11.4: `Source document and section → Requirement or finding →
Disposition → Planned implementation item`. Built by walking every row of
`impl/analysis/requirement-inventory.md` tables A (PLAN findings), B (REVIEW findings/decisions),
C (MATRIX verify-outcomes/constraints), and D (product-level items). Canonical source =
`requirement-inventory.md` + the four primary documents (`impl/index.md` §Source documents).
Every row below mirrors a row in `requirement-inventory.md`; no row is sampled or omitted.

## PLAN-sourced rows (source: `premier-framework-implementation-plan.md` §5, PLAN findings A)

| Source doc+section | Requirement/finding ID | Disposition | Planned implementation item |
|---|---|---|---|
| PLAN §5 (AR-01) | AR-01 | planned | W05-E01-S001..S004 |
| PLAN §5 (AR-02) | AR-02 | planned | W05-E02-S001..S003 |
| PLAN §5 (AR-03) | AR-03 | planned | W05-E03-S001..S002 |
| PLAN §5 (AR-04) | AR-04 | partial | W05-E03-S002 |
| PLAN §5 (AR-05) | AR-05 | partial | W06-E04-S002 |
| PLAN §5 (AR-06) | AR-06 | partial | W05-E04-S001 |
| PLAN §5 (SEC-01) | SEC-01 | planned | W03-E01-S001..S004 |
| PLAN §5 (SEC-02) | SEC-02 | partial | W03-E05-S001 |
| PLAN §5 (SEC-03) | SEC-03 | planned | W03-E03-S001 |
| PLAN §5 (SEC-04) | SEC-04 | planned | W05-E04-S002 |
| PLAN §5 (SEC-05) | SEC-05 | planned | W07-E02-S001 |
| PLAN §5 (SEC-06) | SEC-06 | planned | W03-E02-S001 |
| PLAN §5 (DATA-01) | DATA-01 | planned | W02-E02-S001..S002 |
| PLAN §5 (DATA-02) | DATA-02 | planned | W04-E01-S001..S003 |
| PLAN §5 (DATA-03) | DATA-03 | planned | W04-E02-S001..S002 |
| PLAN §5 (DATA-04) | DATA-04 | planned | W04-E03-S001 |
| PLAN §5 (DATA-05) | DATA-05 | planned | W02-E03-S001 |
| PLAN §5 (DATA-06) | DATA-06 | planned | W02-E04-S001 |
| PLAN §5 (DATA-07) | DATA-07 | blocked→planned | W03-E04-S001 |
| PLAN §5 (DATA-08) | DATA-08 | partial | W04-E04-S001..S002 |
| PLAN §5 (DATA-09) | DATA-09 | planned | W02-E01-S001..S003 |
| PLAN §5 (DX-01) | DX-01 | planned | W01-E04-S001 |
| PLAN §5 (DX-02) | DX-02 | planned | W01-E04-S001 |
| PLAN §5 (DX-03) | DX-03 | deferred | W06-E01-S001 |
| PLAN §5 (DX-04) | DX-04 | planned | W06-E01-S002 |
| PLAN §5 (DX-05) | DX-05 | partial | W01-E04-S002 |
| PLAN §5 (DX-06) | DX-06 | planned | W06-E02-S001 |
| PLAN §5 (DX-07) | DX-07 | planned | W04-E04-S003 |
| PLAN §5 (PERF-01) | PERF-01 | INV | W00-E01-S002 |
| PLAN §5 (PERF-02) | PERF-02 | accepted | W07-E01-S001 |
| PLAN §5 (PERF-03) | PERF-03 | accepted | W07-E01-S002 |
| PLAN §5 (PERF-04) | PERF-04 | accepted | W07-E01-S003 |
| PLAN §5 (PERF-05) | PERF-05 | accepted | W07-E01-S004 |
| PLAN §5 (PERF-06) | PERF-06 | INV | W00-E01-S002 |
| PLAN §5 (REL-01) | REL-01 | planned | W06-E03-S001..S002 |
| PLAN §5 (REL-02) | REL-02 | planned | W06-E03-S003 |
| PLAN §5 (REL-03) | REL-03 | planned | W06-E02-S002..S003 |
| PLAN §5 (REL-04) | REL-04 | partial | W07-E02-S002 |

## REVIEW-sourced rows (source: `fable5-final-architecture-review-2026-07-11.md` §O/§U, REVIEW findings B)

| Source doc+section | Requirement/finding ID | Disposition | Planned implementation item |
|---|---|---|---|
| REVIEW §O (FBL-01) | FBL-01 | planned | W05-E05-S001..S002 |
| REVIEW §O (FBL-02) | FBL-02 | planned | W02-E05-S001 |
| REVIEW §O (FBL-03) | FBL-03 | planned | W01-E04-S002 |
| REVIEW §O (FBL-04) | FBL-04 | planned | W04-E02-S003 |
| REVIEW §O (FBL-05) | FBL-05 | planned | W01-E01-S001 |
| REVIEW §O (FBL-06) | FBL-06 | planned | W01-E02-S001..S002 |
| REVIEW §O (FBL-07) | FBL-07 | partial | W01-E01-S002..S003 |
| REVIEW §O (FBL-08) | FBL-08 | planned | W01-E03-S002 |
| REVIEW §O (FBL-09) | FBL-09 | planned | W01-E03-S001 |
| REVIEW §U (D-01..D-09) | D-01..D-09 | planned | W00-E02-S003 |
| REVIEW §O (T-DOC-01) | T-DOC-01 | planned | W01-E04-S002 |
| REVIEW §O (T-TEST-01) | T-TEST-01 | planned | W01-E04-S003 |
| REVIEW §F Q1 | DEC-Q1 | blocked (human) | W03-E01 (tracked) |
| REVIEW (OPS) | DEC-Q9 | blocked (human) | W07-E01 (tracked) |
| REVIEW / session fact | DEC-Q10 | blocked (human) | W06-E03 (tracked) |

## MATRIX-sourced rows (source: `fable5-closure-depth-matrix-2026-07-11.md`, verify-outcomes/constraints C)

| Source doc+section | Requirement/finding ID | Disposition | Planned implementation item |
|---|---|---|---|
| MATRIX CS-03 | CS-03 | INV→verified | W00 evidence pointer (no dedicated story; re-earned with citations in MATRIX) |
| MATRIX CS-19 | CS-19 | INV→verified | W00 evidence pointer |
| MATRIX CS-24 | CS-24 | INV→verified | W00 evidence pointer; gosec G704 annotation task inside FBL-07 (W01-E01-S002..S003) |
| MATRIX CS-10 | CS-10 | planned | W01-E01-S001 (FBL-05 mechanical enforcement; pool lifetime config keys task) |
| MATRIX CS-25 | CS-25 | planned | D-09 documentation; DEF-01 (file-provider deferred) |
| REVIEW §K | K-RETAIN | not-applicable | none — justified retentions, no work planned |
| REVIEW §K | K-P2 | deferred | DEF-02 (gobreaker), DEF-03 (jwx) |
| REVIEW §M | M-REJ | rejected | none — rationale in REVIEW §M (viper/envconfig, kernel message bus, custom crypto) |
| `framework-backlog-p2-decisions.md` | B11/B12/B13 | deferred | DEF-04 (B11), DEF-05 (B12), DEF-06 (B13) |

## Product-boundary rows (source: `requirement-inventory.md` §D — mandate §2.3 framework boundary)

| Source doc+section | Requirement/finding ID | Disposition | Planned implementation item |
|---|---|---|---|
| `requirement-inventory.md` §D | PROD-01 | product-level (excluded from framework impl) | Enabled by DATA-01 T1 (parent unique index) + DATA-09 protocol |
| `requirement-inventory.md` §D | PROD-02 | product-level (excluded from framework impl) | Enabled by FBL-01 re-home (deprecated forwarding shim) |
| `requirement-inventory.md` §D | PROD-03 | product-level (excluded from framework impl) | Enabled by DX-07 T1 + FBL-09 template fixes |
| `requirement-inventory.md` §D | PROD-04 | product-level (excluded from framework impl) | Enabled by SEC-01 T1/T5 grant contract + coordinated rollout plan |
| `requirement-inventory.md` §D | PROD-05 | product-level (excluded from framework impl) | Enabled by DATA-08 hash_version branch verification (D-04) |

## Session-delta rows (source: `requirement-inventory.md` §E — git history `4eca9f4..0a31186`)

| Source doc+section | Requirement/finding ID | Disposition | Planned implementation item |
|---|---|---|---|
| git history (#23) | SD-01 | informational | W00 baseline captures new wall-clocks |
| git history (#24) | SD-02 | informational | Partially advances FBL-07 (W01-E01-S002..S003); REL-04 T8 fuzz remains (W07-E02-S002) |
| git history (#25) | SD-03 | informational | W00-E01-S002 verifies PERF-01 against new budgets |
| git history (#22) | SD-04 | informational | none — archive index is the provenance record |

## Row count

38 PLAN rows + 15 REVIEW rows + 9 MATRIX rows + 5 product-boundary rows + 4 session-delta rows =
**71 rows**, matching `requirement-inventory.md`'s own totals line (38 + 15 + 9 + 5 + 4 = 71
source items, no item dropped).
