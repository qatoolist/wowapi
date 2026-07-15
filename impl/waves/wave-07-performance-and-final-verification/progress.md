---
id: W07-PROGRESS
type: wave-progress
wave: W07
status: in-progress
derived: false
created_at: 2026-07-12
updated_at: 2026-07-14
---

# W07 progress

Phase A execution status as of 2026-07-14.

## Epic status

| Epic | Title | Status | Stories | Story status breakdown |
|---|---|---|---|---|
| W07-E01 | performance-programme | accepted | 4 | 4 accepted |
| W07-E02 | verification-hardening | blocked | 2 | 1 accepted, 1 blocked |
| W07-E03 | product-alignment-verification | blocked | 1 | 1 blocked |
| W07-E04 | programme-closure | planned | 2 | 2 planned |

## Story status

| Story | Title | Status | Task count | Task status breakdown |
|---|---|---|---|---|
| W07-E01-S001 | request-benchmarks-real-pg | accepted | 6 | 6 done (incl. independent review) |
| W07-E01-S002 | rules-resolution-sql | accepted | 7 | 7 done; clean independent story review |
| W07-E01-S003 | sweeper-materialization | accepted | 9 | 9 done; clean independent story review |
| W07-E01-S004 | checksum-behaviour-and-bench-coverage | accepted | 7 | 7 done; clean independent story review |
| W07-E02-S001 | security-verification-profile | blocked | 2 | agent-reachable work complete; external/prerequisite gates blocked |
| W07-E02-S002 | coverage-truthfulness-completion | accepted | 5 | 5 done (incl. independent review) |
| W07-E03-S001 | wowsociety-readiness-check | blocked | 3 | package complete; 2 substantive criteria blocked |
| W07-E04-S001 | final-verification-gate | planned | 4 | 4 todo (incl. 1 independent-review task) |
| W07-E04-S002 | closure-and-claim-decision | planned | 3 | 3 todo |

## Blocked items

- W07-E02-S001: the control map and internal story review are complete, but 0/7 upstream SEC
  lifecycle pairs are accepted, no external professional assessor/report exists, and clean-revision
  revalidation remains pending.
- W07-E03-S001: PROD-01 lacks `UNIQUE (tenant_id,id)` on `rule_versions`; PROD-04's rollout artifact
  contradicts the current grant schema/authority model and lacks product sign-off.

## Critical dependencies

- W07-E02-S001 (SEC-05) hard-depends on SEC-01/03/04/06's own acceptance — per PLAN SEC-05 T1's own
  dependency row: "SEC-01–04 substantially complete."
- W07-E01's own absolute-SLO acceptance criteria are conditional on DEC-Q9 — tracked at epic level, not
  resolved by this wave's own execution capacity.
- W07-E04-S001 (final verification gate) depends on every other epic in this wave, and transitively on
  every prior wave, having reached its own closure state first.
- W07-E04-S002 (closure-and-claim-decision) depends on W07-E04-S001.

## Open decisions

DEC-Q9 (reference-performance-environment ownership) is open and tracked, with REVIEW §F row 9's own
provisional default already in effect (relative/container benchmarking proceeds now; absolute-SLO
gating waits). No other new decision opens in this wave.

## Open risks

See `risks.md`.

## Artifact completeness

7/9 story-level artifact sets populated; 5 are accepted and 2 truthfully document blockers.

## Evidence completeness

40 produced evidence records (27 E01, 8 E02, 5 E03); E04 has six planned slots but no produced evidence.

## Review state

All five accepted W07 stories have clean independent review; E01 also passed its fresh epic gate with no open findings.

## Exit-gate readiness

Not ready. 5 of 9 stories and E01 are accepted; 2 stories remain blocked and E04's 2 stories remain planned.
