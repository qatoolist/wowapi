---
id: W04-E04-S001-TASKS-INDEX
type: tasks-index
parent_story: W04-E04-S001
status: accepted
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W04-E04-S001 — Tasks index

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W04-E04-S001-T001](task-001-hash-widening-and-version-migration.md) | Audit hash-chain widening, hash_version migration, and version-branched verification | W04Compliance | done | none (story-level dependency on W02-E01) | Widened chainHash + hash_version migration + version-branched Verify + per-field tamper test | AC-W04-E04-S001-01, AC-W04-E04-S001-02, AC-W04-E04-S001-03 | implemented | verified |
| [W04-E04-S001-T002](task-002-independent-review.md) | Independent review | reviewer | done | T001 | Independent-review record per mandate §14 | AC-W04-E04-S001-01, AC-W04-E04-S001-02, AC-W04-E04-S001-03 | n/a | passed (no open issues) |

## Grouping rationale

Per mandate §12: this story has a single substantive implementation task (T001) because D-04's own
decision text requires the `hash_version` migration and the `chainHash` widening to land as one
atomic unit. T002 adds a mandatory independent-review task per mandate §14, with specific attention
to per-field tamper coverage and D-04 fidelity.
