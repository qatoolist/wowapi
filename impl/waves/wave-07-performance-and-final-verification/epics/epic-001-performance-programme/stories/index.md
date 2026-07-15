---
id: W07-E01-STORIES-INDEX
type: stories-index
epic: W07-E01
wave: W07
status: accepted
derived: false
created_at: 2026-07-12
updated_at: 2026-07-14
---

# W07-E01 — Stories index

| Story | Title | Status | Priority | Source requirement | Tasks | Objective |
|---|---|---|---|---|---|---|
| [W07-E01-S001](story-001-request-benchmarks-real-pg/story.md) | request-benchmarks-real-pg | accepted | P1 | PERF-02 (T1-T5) | 6 | The §14 reference-env stand-up (shared prerequisite) plus DB-backed complete-request benchmarks |
| [W07-E01-S002](story-002-rules-resolution-sql/story.md) | rules-resolution-sql | accepted | P1 | PERF-03 (T0-T6) | 7 | Collapse rules resolution from a per-ancestor query loop into bounded, index-verified SQL |
| [W07-E01-S003](story-003-sweeper-materialization/story.md) | sweeper-materialization | accepted | P1 | PERF-04 (T1-T8) | 9 | Remove N+1/unbounded materialization from sweepers/workers; leased-state-machine outbox rework |
| [W07-E01-S004](story-004-checksum-behaviour-and-bench-coverage/story.md) | checksum-behaviour-and-bench-coverage | accepted | P2 | PERF-05 (T1-T5); CS-16 | 7 | Explicit required-checksum behavior; 7-package hot-path benchmark coverage expansion |
