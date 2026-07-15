---
id: W04-E02-STORIES-INDEX
type: stories-index
epic: W04-E02
wave: W04
status: accepted
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W04-E02 — Stories index

| Story | Title | Status | Priority | Source requirement | Tasks | Objective |
|---|---|---|---|---|---|---|
| [W04-E02-S001](story-001-notify-and-webhook-three-stage/story.md) | notify-and-webhook-three-stage | planned | P0 | DATA-03 (T1, T2, T3) | 4 | Reuse the shared lease primitive for notify/webhook claim rows; the three-stage claim-tx → effect-outside-tx → fenced-finalize-tx protocol for `kernel/notify` and `kernel/webhook.deliverToEndpoint` |
| [W04-E02-S002](story-002-inbound-two-phase-and-contracts/story.md) | inbound-two-phase-and-contracts | planned | P0 | DATA-03 (T4, T5, T6, T8; T7 cross-ref only) | 6 | Inbound two-phase webhook verification; failed-signature audit; per-adapter idempotency-contract declaration; the named 6-boundary chaos test (reusing W04-E01-S003's harness) |
| [W04-E02-S003](story-003-retry-adoption/story.md) | retry-adoption | planned | P1 | FBL-04 | 3 | Adopt `cenkalti/backoff/v5` in place of the framework's two duplicated hand-rolled retry implementations, with retry-schedule parity and fault-injection tests |
