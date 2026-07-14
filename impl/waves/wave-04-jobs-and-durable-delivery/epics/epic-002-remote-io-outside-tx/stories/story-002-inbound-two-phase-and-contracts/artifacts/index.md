---
id: W04-E02-S002-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W04-E02-S002
status: accepted
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W04-E02-S002 — Artifacts index

Per mandate §9.2. Structure adaptation per `governance/naming-conventions.md` "Adaptation 2":
lifecycle subdirectories (`pre-implementation/`, `implementation/`, `post-implementation/`) are
created on first real content, not pre-populated empty. All entries below are `not yet produced`.

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Path | Status |
|---|---|---|---|---|---|---|---|---|
| ART-W04-E02-S002-001 | Inbound two-phase verification implementation | source-code package | implementation | Two-phase read-tx snapshot / verify-outside-tx / write-tx re-check protocol for `HandleInbound` | DATA-03 (T4) | W04-E02-S002-T001 | `DATA-03/webhook/inbound-two-phase/` | not yet produced |
| ART-W04-E02-S002-002 | Failed-signature audit path | source-code package | implementation | Body-free audit row write, own short transaction, on failed verification | DATA-03 (T5) | W04-E02-S002-T002 | `DATA-03/webhook/failed-sig-audit/` | not yet produced |
| ART-W04-E02-S002-003 | Per-adapter idempotency-safety contract declaration mechanism | source-code package | implementation | Boot-time-enforced adapter-registration contract plus `Sender` implementation inventory | DATA-03 (T6) | W04-E02-S002-T003 | `DATA-03/adapter-contract/` | not yet produced |
| ART-W04-E02-S002-004 | 6-boundary chaos-test suite (notify and webhook) | test suite | implementation | Named chaos test at 6 boundaries, reusing W04-E01-S003's harness, applied to both notify and webhook | DATA-03 (T8) | W04-E02-S002-T004 | `DATA-03/chaos/` | not yet produced |
| ART-W04-E02-S002-005 | T7 cross-reference record | documentation | post-implementation | Records that DATA-03 T7 is already executed under DATA-08 W0-T2 and is not re-implemented here | DATA-03 (T7, cross-ref only) | W04-E02-S002-T005 | `DATA-08/wave0/legal-audit/` (referenced, not produced by this epic) | not yet produced |
| ART-W04-E02-S002-006 | Consolidated story acceptance evidence package | documentation | post-implementation | Aggregates T001–T004's evidence and the T7 cross-reference into one story-scope record | DATA-03 (T4, T5, T6, T8) | W04-E02-S002-T005 | TBD at implementation time | not yet produced |
