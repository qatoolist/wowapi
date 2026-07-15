---
id: W04-E02-S001-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W04-E02-S001
status: accepted
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W04-E02-S001 — Artifacts index

Per mandate §9.2. Structure adaptation per `governance/naming-conventions.md` "Adaptation 2":
lifecycle subdirectories (`pre-implementation/`, `implementation/`, `post-implementation/`) are
created on first real content, not pre-populated empty. All entries below are `not yet produced`.

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Path | Status |
|---|---|---|---|---|---|---|---|---|
| ART-W04-E02-S001-001 | Claim-row lease-column migration | migration | implementation | Adds W04-E01's shared primitive's lease columns to notify/webhook delivery-tracking tables | DATA-03 (T1) | W04-E02-S001-T001 | `DATA-03/lease-columns/` | not yet produced |
| ART-W04-E02-S001-002 | Notify three-stage protocol implementation | source-code package | implementation | Claim-tx / effect-outside-tx / finalize-tx protocol for `kernel/notify` | DATA-03 (T2) | W04-E02-S001-T002 | `DATA-03/notify/` | not yet produced |
| ART-W04-E02-S001-003 | Webhook three-stage protocol implementation | source-code package | implementation | Claim-tx / effect-outside-tx / finalize-tx protocol for `kernel/webhook.deliverToEndpoint` | DATA-03 (T3) | W04-E02-S001-T003 | `DATA-03/webhook/` | not yet produced |
| ART-W04-E02-S001-004 | Three-stage protocol documentation | documentation | post-implementation | Documents claim/effect/finalize stage boundaries for notify and webhook | DATA-03 (T2, T3) | W04-E02-S001-T002, W04-E02-S001-T003 | TBD at implementation time | not yet produced |
