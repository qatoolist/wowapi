---
id: W04-E04-S003-TASKS-INDEX
type: tasks-index
parent_story: W04-E04-S003
status: accepted
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W04-E04-S003 — Tasks index

**No task exists for DX-07 T4** — it is explicitly out of scope for this story; see `story.md`
"Out of scope" and epic-level `risks.md` RISK-W04-004.

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W04-E04-S003-T001](task-001-migration-currency-check.md) | Migration-currency readiness check | W04Compliance | done | none | Readiness template gains migration-currency check + stale-migration 503 test | AC-W04-E04-S003-01 | implemented | verified |
| [W04-E04-S003-T002](task-002-seed-rule-model-hash-readiness.md) | Seed/rule/model-hash readiness reporting | W04Compliance | done | none (model-hash portion contingent on AR-01) | Readiness payload gains migration version + seed/rule hash + model hash reporting | AC-W04-E04-S003-02 | implemented with deviation DEV-W04-E04-S003-001 | verified |
| [W04-E04-S003-T003](task-003-config-doctor-discovery-fix.md) | config doctor product-root discovery fix | W04Compliance | done | none | go env GOMOD/--project discovery + explicit product-validation-ran reporting | AC-W04-E04-S003-03 | implemented | verified |
| [W04-E04-S003-T004](task-004-independent-review.md) | Independent review | reviewer | done | T001, T002, T003 | Independent-review record per mandate §14, incl. confirming T4's correct exclusion | AC-W04-E04-S003-01, AC-W04-E04-S003-02, AC-W04-E04-S003-03 | n/a | passed (no open issues; T4 absent) |

## Grouping rationale

Per mandate §12: T001/T002/T003 map one-to-one onto PLAN DX-07's own T1/T2/T3 rows. T2 carries an
internal contingency (model-hash portion depends on AR-01) recorded as deviation DEV-W04-E04-S003-001
rather than split further. T004 adds a mandatory independent-review task per mandate §14 to confirm
DX-07 T4 was correctly scoped out.
