---
id: W02-E01-STORIES-INDEX
type: stories-index
epic: W02-E01
wave: W02
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W02-E01 — Stories index

| Story | Title | Status | Priority | Source requirement | Tasks | Objective |
|---|---|---|---|---|---|---|
| [W02-E01-S001](story-001-manifest-and-lock-budget/story.md) | manifest-and-lock-budget | planned | P0 | DATA-09 (T1, T2) | 3 | The migration manifest schema and CI validation; 2-second lock-timeout enforcement with bounded abort-and-retry |
| [W02-E01-S002](story-002-expand-backfill-validate/story.md) | expand-backfill-validate | planned | P0 | DATA-09 (T3, T4, T5) | 4 | Expand-phase tooling; resumable backfill-job harness (interim checkpoint lease, forward-dependency-flagged); validation-phase tooling with machine-checked artifacts |
| [W02-E01-S003](story-003-canary-switch-contract-drills/story.md) | canary-switch-contract-drills | planned | P0 | DATA-09 (T6, T7, T8, T9) | 6 | Canary/deploy-N soak tooling; switch-phase tooling with application rollback; contract-phase tooling gated on evidenced safety; full 6-drill CI pipeline plus evidence aggregation |
