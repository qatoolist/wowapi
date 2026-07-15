---
id: W04-E03-STORIES-INDEX
type: stories-index
epic: W04-E03
wave: W04
status: accepted
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W04-E03 — Stories index

| Story | Title | Status | Priority | Source requirement | Tasks | Objective |
|---|---|---|---|---|---|---|
| [W04-E03-S001](story-001-stopgap/story.md) | stopgap | planned | P0 | DATA-04 (T1) | 1 | Correct the false "safe across replicas" migration comment; enforce single-processor via advisory lock/CAS at the `Service` API boundary — ships independently and fast, closing the false-documentation sub-issue before the full rewrite |
| [W04-E03-S002](story-002-leased-claims-and-lifecycle/story.md) | leased-claims-and-lifecycle | planned | P1 | DATA-04 (T2, T3, T4, T5, T6) | 5 | Lease columns via the shared primitive; atomic `SKIP LOCKED` leased claim; item idempotency, finalize fencing, retry policy, cancellation; pause/resume/cancel lifecycle controls; the named multi-worker chaos test reusing the shared chaos harness |
