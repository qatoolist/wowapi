---
id: W02-E02-STORIES-INDEX
type: stories-index
epic: W02-E02
wave: W02
status: accepted
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W02-E02 — Stories index

| Story | Title | Status | Priority | Source requirement | Tasks | Objective |
|---|---|---|---|---|---|---|
| [W02-E02-S001](story-001-parent-indexes-scanner-gate/story.md) | parent-indexes-scanner-gate | accepted | P0 | DATA-01 (T1, T2, T6) | 4 | `UNIQUE (tenant_id, id)` on every referenced parent; the tenant-FK catalog scanner; the permanent CI gate |
| [W02-E02-S002](story-002-audit-fk-validate-negatives/story.md) | audit-fk-validate-negatives | accepted | P0 | DATA-01 (T3, T4, T5, T7, T8) | 6 | Mismatch audit; composite `NOT VALID` FK add and validation; cross-tenant negative tests; FK cleanup |
