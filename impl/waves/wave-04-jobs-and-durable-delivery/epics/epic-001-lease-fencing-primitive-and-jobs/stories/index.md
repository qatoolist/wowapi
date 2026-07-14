---
id: W04-E01-STORIES-INDEX
type: stories-index
epic: W04-E01
wave: W04
status: accepted
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W04-E01 — Stories index

| Story | Title | Status | Priority | Source requirement | Tasks | Objective |
|---|---|---|---|---|---|---|
| [W04-E01-S001](story-001-shared-primitive/story.md) | shared-primitive | planned | P0 | DATA-02 (T1) | 3 | The shared lease/fencing primitive itself — the wave's keystone build — plus the planned supersession of W02-E01-S002's interim checkpoint lease |
| [W04-E01-S002](story-002-jobs-lease-and-finalize/story.md) | jobs-lease-and-finalize | planned | P0 | DATA-02 (T2, T3, T4) | 4 | Lease columns on `jobs_queue`; fenced finalize; fenced reclaim with generation bump |
| [W04-E01-S003](story-003-idempotency-and-chaos/story.md) | idempotency-and-chaos | planned | P0 | DATA-02 (T5, T6, T7) | 5 | Worker idempotency-declaration contract; effect-ledger-survives-fencing test; the named chaos test, built as a harness shared with W04-E02/W04-E03 |
