---
id: W07-E03-S001-EVIDENCE-INDEX
type: evidence-index
parent_story: W07-E03-S001
status: failed
derived: false
created_at: 2026-07-12
updated_at: 2026-07-14
---

# W07-E03-S001 — Evidence index

Per mandate §10 and `governance/evidence-policy.md`. Evidence files carry the full required metadata;
the index preserves the result/status distinction. A failed acceptance result is retained rather than
rewritten as successful because other rows passed.

| Evidence ID | Type | Task | Acceptance criteria | Execution command | Commit SHA | Result | Status | Path |
|---|---|---|---|---|---|---|---|---|
| EV-W07-E03-S001-001 | re-verification report (PROD-01/02/03) | T001 | AC01, AC02, AC03 | focused tenant-FK/migration, MFA, readiness/scaffold tests + live catalog probe | `733ef3e930cbb3f89f5bbc53d8f562c60e426513` | AC01 FAIL; AC02/03 PASS | failed | `tests/EV-W07-E03-S001-001.md` |
| EV-W07-E03-S001-002 | re-verification report (PROD-04/05) | T002 | AC04, AC05 | focused grant/audit/manifest tests + live schema/RLS/grant probes + rollout cross-check | `733ef3e930cbb3f89f5bbc53d8f562c60e426513` | AC04 FAIL; AC05 PASS | failed | `tests/EV-W07-E03-S001-002.md` |
| EV-W07-E03-S001-003 | failed execution record | T001 | AC01 | `go test ./kernel/migration -count=1 -v` | `733ef3e930cbb3f89f5bbc53d8f562c60e426513` | FAIL: required PostgreSQL unavailable | failed | `tests/EV-W07-E03-S001-003.md` |
| EV-W07-E03-S001-004 | retest execution record | T001 | AC01 | `go test ./kernel/migration -count=1 -v` | `733ef3e930cbb3f89f5bbc53d8f562c60e426513` | PASS after declared services/migrations started | retested | `tests/EV-W07-E03-S001-004.md` |
| EV-W07-E03-S001-005 | independent review + focused reruns | closure review | AC01..AC05 package, epic AC03 | exact commands in review record | `733ef3e930cbb3f89f5bbc53d8f562c60e426513` | PASS package; upstream AC01/04 blockers corroborated | retested | `reviews/EV-W07-E03-S001-005.md` |

Failed EV-003 is preserved permanently; EV-004 is its separate passing infrastructure retest.
Neither record resolves the substantive PROD-01 capability failure.
