---
id: W04-E01-S001-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W04-E01-S001
status: accepted
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W04-E01-S001 — Artifacts index

Per mandate §9.2. Structure adaptation per `governance/naming-conventions.md` "Adaptation 2":
lifecycle subdirectories (`pre-implementation/`, `implementation/`, `post-implementation/`) are
created on first real content, not pre-populated empty. All entries below are `not yet produced`.

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Path | Status |
|---|---|---|---|---|---|---|---|---|
| ART-W04-E01-S001-001 | Shared lease/fencing primitive | source-code package | implementation | The reusable kernel building block (`lease_token`, `lease_generation`, `lease_expires_at`, optional heartbeat) | DATA-02 | W04-E01-S001-T001 | TBD at implementation time | not yet produced |
| ART-W04-E01-S001-002 | Interim-checkpoint-lease migration tooling | source-code package | implementation | Reads W02-E01-S002's interim-lease checkpoint state and re-expresses it under the shared primitive's schema | DATA-02 | W04-E01-S001-T002 | TBD at implementation time | not yet produced |
| ART-W04-E01-S001-003 | Shared primitive contract and migration documentation | documentation | post-implementation | Documents the primitive's field set, comparison semantics, package location, and the completed interim-lease migration | DATA-02 | W04-E01-S001-T001, W04-E01-S001-T002 | TBD at implementation time | not yet produced |
