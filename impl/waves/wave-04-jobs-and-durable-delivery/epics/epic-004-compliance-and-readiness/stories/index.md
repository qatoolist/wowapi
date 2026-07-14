---
id: W04-E04-STORIES-INDEX
type: stories-index
epic: W04-E04
wave: W04
status: accepted
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W04-E04 — Stories index

| Story | Title | Status | Priority | Source requirement | Tasks | Objective |
|---|---|---|---|---|---|---|
| [W04-E04-S001](story-001-audit-hash-widening/story.md) | audit-hash-widening | planned | P0 | DATA-08 (W6-T1); enacts D-04 | 2 | Widen `chainHash` to cover every persisted field (metadata, tx_id, all nullable fields, sequence, ID, timestamps, previous hash); add `hash_version` discriminator column; version-branch verification. Single highest-risk task in PF-DATA's Wave-6 scope; hits wowsociety's live audit rows |
| [W04-E04-S002](story-002-anchor-dsr-hold/story.md) | anchor-dsr-hold | planned | P1 | DATA-08 (W6-T2, T3, T4, T5) | 5 | External anchor verification for the audit chain; encrypted immutable DSR export artifact; central legal-hold enforcement wrapper; explicit per-class DSR status reporting |
| [W04-E04-S003](story-003-readiness-truthfulness/story.md) | readiness-truthfulness | planned | P1 | DX-07 (T1, T2, T3; T4 explicitly deferred-linked to W05-E03-S002's AR-04 T5) | 4 | Migration-currency readiness check; seed/rule/model-hash readiness reporting; `config doctor` product-root discovery fix |
