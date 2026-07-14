---
id: VER-W07-E03-S001
type: verification-record
parent_story: W07-E03-S001
status: blocked
created_at: 2026-07-12
updated_at: 2026-07-14
---

# Verification record — W07-E03-S001

## Planned verification procedure

Per mandate §8.8. One row per acceptance criterion for this story.

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W07-E03-S001-01 | Direct source, live-catalog, scanner, migration-protocol and manifest checks | Local PostgreSQL + MinIO, required DB/S3 env | DATA-09 PASS; DATA-01 parent-key prerequisite FAIL | `EV-W07-E03-S001-001`, `-003`, `-004`, `-005` | W05ReviewGateFinal |
| AC-W07-E03-S001-02 | Direct shim/canonical source inspection and focused package tests | Go 1.26.5 | PASS; shim exists and path documented | `EV-W07-E03-S001-001`, `-005` | W05ReviewGateFinal |
| AC-W07-E03-S001-03 | Readiness integration, template assertions, rendered-product compile | Local PostgreSQL + Go 1.26.5 | PASS; readiness and all four timeout paths documented | `EV-W07-E03-S001-001`, `-005` | W05ReviewGateFinal |
| AC-W07-E03-S001-04 | Grant schema/RLS/privileges probe, resolver tests, rollout-document cross-check | Local PostgreSQL + Go 1.26.5 | Framework contract PASS; rollout artifact FAIL/stale | `EV-W07-E03-S001-002`, `-005` | W05ReviewGateFinal |
| AC-W07-E03-S001-05 | Version-branch/tamper/manifest/audit-CLI tests and live column probe | Local PostgreSQL + Go 1.26.5 | PASS; staging path documented; zero wowsociety change | `EV-W07-E03-S001-002`, `-005` | W05ReviewGateFinal |

## Post-execution record

Verification was executed directly at the pinned revision. Exact commands, outputs, environment,
timestamps, and the resolved initial infrastructure failure are retained in the indexed evidence.

### Actual result

AC02, AC03, and AC05 pass. AC01 fails because `rule_versions` has no unique/exclusion key on
`(tenant_id, id)`. AC04 fails because the existing W03-E01-S004 rollout documents use nonexistent
columns and contradict current `Verifier.Actor` behavior and the safe rollback boundary.

### Pass or fail

**FAIL / blocked**: three criteria pass and two fail. No exception is accepted.

### Evidence identifier

`EV-W07-E03-S001-001` through `EV-W07-E03-S001-005`.

### Execution date

2026-07-14 (`2026-07-13T20:39Z` through `2026-07-13T21:14:39Z`).

### Commit or revision

`733ef3e930cbb3f89f5bbc53d8f562c60e426513` on `main`.

### Environment

Darwin arm64; Go 1.26.5; PostgreSQL client 18.4; local PostgreSQL/MinIO services declared by
`deployments/compose.yaml`; required DB and S3 enforcement enabled. DSN password is redacted.

### Reviewer

`W05ReviewGateFinal` — PASS for the package; no open actionable package issue. The reviewer
independently reran the focused commands and confirmed the two upstream blockers remain.

### Findings

PROD-01 and PROD-04 blockers are detailed in `ART-W07-E03-S001-001`. PROD-02 has a non-blocking
sunset-scheduling gap. PROD-03 and PROD-05 require product-side execution/evidence, as expected.

### Retest status

The initial DB-unavailable failure is preserved as `EV-W07-E03-S001-003` and resolved by the passing
exact-command retest `EV-W07-E03-S001-004`. No retest can resolve the substantive PROD-01/04 findings
until their owners change the missing parent key and stale rollout material.

### Final conclusion

The framework-only verification package is complete and honestly records every row, but the story
remains blocked: AC01 and AC04 are not satisfied. No wowsociety repository was read or changed.
