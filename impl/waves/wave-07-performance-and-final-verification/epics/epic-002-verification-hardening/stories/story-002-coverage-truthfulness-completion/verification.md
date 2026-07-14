---
id: VER-W07-E02-S002
type: verification-record
parent_story: W07-E02-S002
status: verified
created_at: 2026-07-12
updated_at: 2026-07-14
---

# Verification record — W07-E02-S002

## Planned verification procedure

Per mandate §8.8. One row per acceptance criterion for this story.

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W07-E02-S002-01 | Kill required DB and S3 dependencies under the authoritative requirement flags | Local negative fixtures | Non-zero inner test exits with actionable diagnoses | fail-closed test report | W05ReviewGateFinal |
| AC-W07-E02-S002-02 | Run exhaustive manifest plus unapproved and approved fixtures | Local | Unguarded skip fails; approved skip with rationale passes | skip-manifest fixture report | W05ReviewGateFinal |
| AC-W07-E02-S002-03 | Run seeded negative fixture, then `-race` over DB/S3 packages | Real PostgreSQL/MinIO | Seeded race caught; real suite clean | race test report | W05ReviewGateFinal |
| AC-W07-E02-S002-04 | Run PR 10s profile then scheduled 1m profile against same cache | Local native fuzz engine | Positive elapsed executions; scheduled run restores and grows corpus | fuzz-duration/corpus-mtime report | W05ReviewGateFinal |

## Post-execution record

| Acceptance criterion | Actual result | Pass/fail | Evidence | Revision |
|---|---|---|---|---|
| AC-W07-E02-S002-01 | Missing DB and S3 each forced non-zero inner `go test` with the requirement flag and dependency diagnosis. | PASS | EV-W07-E02-S002-001 | `733ef3e930cbb3f89f5bbc53d8f562c60e426513` + scoped shared-worktree provenance |
| AC-W07-E02-S002-02 | 38 repository approvals validated; unapproved fixture rejected; owner+rationale fixture accepted. | PASS | EV-W07-E02-S002-002 | same |
| AC-W07-E02-S002-03 | Seed emitted `WARNING: DATA RACE`; seven real DB/S3 integration packages passed `-race -count=1`. | PASS | EV-W07-E02-S002-003 | same |
| AC-W07-E02-S002-04 | PR corpus 0→520 files; scheduled restored 520 and grew to 761; every target emitted positive elapsed time and over one million executions. | PASS | EV-W07-E02-S002-004 | same |

### Execution date

2026-07-13T21:19:33Z through 2026-07-13T21:27:09Z (2026-07-14 local).

### Environment

macOS Darwin 25.5.0 arm64; Go 1.26.5; real local PostgreSQL and MinIO for the integration race suite;
explicit `WOWAPI_REQUIRE_DB=1` and `WOWAPI_REQUIRE_S3=1`.

### Reviewer

W05ReviewGateFinal — independent review PASS, no open actionable issue.

### Findings

Implementation verification found and corrected one proof-parser defect: minute-format native fuzz
progress (`1m0s`) was initially parsed only through its prior `57s` line. A fail-first unit test now
requires minute-duration parsing and passes. Raw scheduled logs already contain the full `1m0s` line.

### Retest status

Focused unit tools, fail-closed fixtures, manifest fixtures, workflow lint, seeded race fixture, real
integration race suite, and both real fuzz profiles passed after implementation.

### Final conclusion

All four acceptance criteria have passing evidence and independent review passed with no open issues.
Story is verified and accepted.
