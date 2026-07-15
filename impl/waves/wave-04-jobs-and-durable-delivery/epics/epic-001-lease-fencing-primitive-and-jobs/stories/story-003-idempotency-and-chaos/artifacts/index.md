---
id: W04-E01-S003-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W04-E01-S003
status: accepted
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W04-E01-S003 — Artifacts index

Per mandate §9.2. Structure adaptation per `governance/naming-conventions.md` "Adaptation 2":
lifecycle subdirectories (`pre-implementation/`, `implementation/`, `post-implementation/`) are
created on first real content, not pre-populated empty. All entries below are `not yet produced`.

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Path | Status |
|---|---|---|---|---|---|---|---|---|
| ART-W04-E01-S003-001 | Idempotency-declaration contract | source-code package | implementation | Worker-registration-time enforcement requiring exactly one declared duplicate-safety mechanism; stable idempotency key + lease context threaded to worker invocation | DATA-02 | W04-E01-S003-T001 | TBD at implementation time | not yet produced |
| ART-W04-E01-S003-002 | Effect-ledger-vs-fencing test | source-code package | implementation | Testable proof that queue-row fencing alone does not undo a committed stale-worker domain transaction | DATA-02 | W04-E01-S003-T002 | TBD at implementation time | not yet produced |
| ART-W04-E01-S003-003 | Named chaos test and reusable chaos harness | source-code package | implementation | `DATA-02/chaos/duplicate_worker_lease_expiry_test.go`, built as a harness shared with W04-E02/W04-E03 | DATA-02 | W04-E01-S003-T003 | TBD at implementation time | not yet produced |
| ART-W04-E01-S003-004 | Consolidated evidence bundle | evidence-bundle | post-implementation | Aggregates T001–T003's individual test outputs into one consolidated record | DATA-02 | W04-E01-S003-T004 | TBD at implementation time | not yet produced |
| ART-W04-E01-S003-005 | Idempotency-contract, effect-ledger, harness-reuse, and T5-coordination documentation | documentation | post-implementation | Documents the idempotency contract, fencing/effect-ledger distinction, chaos-harness reuse contract for W04-E02/W04-E03, and the T5 breaking-change coordination note | DATA-02 | W04-E01-S003-T001, W04-E01-S003-T002, W04-E01-S003-T003 | TBD at implementation time | not yet produced |
