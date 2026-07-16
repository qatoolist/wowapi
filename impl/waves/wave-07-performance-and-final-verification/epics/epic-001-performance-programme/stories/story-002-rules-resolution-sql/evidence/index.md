---
id: W07-E01-S002-EVIDENCE-INDEX
type: evidence-index
parent_story: W07-E01-S002
status: complete
derived: false
created_at: 2026-07-12
updated_at: 2026-07-14
---

# W07-E01-S002 — Evidence index

Per mandate §10. All seven evidence items were produced from actual repository inspection, real
PostgreSQL execution, or validated publication output and accepted after clean independent review.

| Evidence ID | Type | Task | Acceptance criteria proven | Execution command | Commit SHA | Result | Status | Re-pin addendum (H-8/R-6) |
|---|---|---|---|---|---|---|---|---|
| EV-W07-E01-S002-001 | audit output | W07-E01-S002-T001 | AC-W07-E01-S002-01 | migration glob + line-numbered regex inspection | `733ef3e` + working tree | active-only claim confirmed before design | passed | `audits/EV-W07-E01-S002-001-repin-2026-07-16.md` |
| EV-W07-E01-S002-002 | result-parity real-PostgreSQL test output | W07-E01-S002-T002 | AC-W07-E01-S002-02 | `go test ./kernel/rules -run 'TestIntegrationResolver(QueryCountConstantWithDepth\|SetBasedParity)$' -count=1 -v` | `733ef3e` + working tree | six precedence/history cases pass | passed | `tests/EV-W07-E01-S002-002-repin-2026-07-16.json` |
| EV-W07-E01-S002-003 | EXPLAIN index-access confirmation | W07-E01-S002-T003 | AC-W07-E01-S002-03 | `go test ./kernel/rules -run TestIntegrationResolverExplainFixtures -count=1 -v` | `733ef3e` + working tree | current and historical predicates use index access; zero rule_versions seq scans | passed | `explain/EV-W07-E01-S002-003-repin-2026-07-16.json` |
| EV-W07-E01-S002-004 | fixture inventory + output | W07-E01-S002-T004 | AC-W07-E01-S002-04 | same EXPLAIN harness with `WOWAPI_UPDATE_EXPLAIN_FIXTURES=1` | `733ef3e` + working tree | four cardinality fixtures produced | passed | `explain/EV-W07-E01-S002-004-repin-2026-07-16.json` |
| EV-W07-E01-S002-005 | parametrized SQL-count report | W07-E01-S002-T005 | AC-W07-E01-S002-05 | `go test ./kernel/rules -run TestIntegrationResolverQueryCountConstantWithDepth -count=1 -v` | `733ef3e` + working tree | legacy 11/18/58; set-based 8/8/8 | passed | `tests/EV-W07-E01-S002-005-repin-2026-07-16.json` |
| EV-W07-E01-S002-006 | live-update regression report | W07-E01-S002-T006 | AC-W07-E01-S002-06 | `go test ./kernel/rules -count=1 -v` | `733ef3e` + working tree | package and all live visibility tests pass, no skips | passed | `tests/EV-W07-E01-S002-006-repin-2026-07-16.json` |
| EV-W07-E01-S002-007 | published relative comparison | W07-E01-S002-T007 | AC-W07-E01-S002-06 | parse/hash/contract validation of `perf/results/perf-03-comparison.json` | `733ef3e` + working tree | valid, real relative data, DEC-Q9 conditional | passed | `published/EV-W07-E01-S002-007-repin-2026-07-16.md` |

Evidence status vocabulary (per mandate §10): `not yet produced` is this programme's pre-execution
state, outside the mandate's own failed/superseded/retested/resolved/accepted-exception vocabulary,
which applies only once an evidence item has actually been produced at least once.

**Revision re-pin note (autopsy finding H-8, remediation R-6, 2026-07-16):** the `733ef3e` pin above
(full SHA `733ef3e930cbb3f89f5bbc53d8f562c60e426513`) is not an ancestor of current HEAD (a side
effect of the e8cda6b squash). Per `impl/governance/evidence-policy.md`'s revision-pinning rule,
each record above has been re-run against current HEAD
(`43b6e128672f0b0997adcebc92703884deba5684`); the results are recorded, without overwriting these
original rows, in the sibling `*-repin-2026-07-16` addendum files listed above (status `retested`).
No divergence from the original claims was found. EV-W07-E01-S002-003's addendum was re-run in
read-only mode (without `WOWAPI_UPDATE_EXPLAIN_FIXTURES=1`) to avoid mutating checked-in fixture
files, which are outside this remediation's evidence-files-only scope; EV-W07-E01-S002-004's
addendum instead verifies the four fixture files still exist at their recorded paths with fresh
checksums.
