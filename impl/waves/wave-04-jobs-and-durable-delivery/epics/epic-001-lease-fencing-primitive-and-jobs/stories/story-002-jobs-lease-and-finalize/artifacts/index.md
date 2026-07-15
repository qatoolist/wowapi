---
id: W04-E01-S002-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W04-E01-S002
status: accepted
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W04-E01-S002 — Artifacts index

Per mandate §9.2. Structure adaptation per `governance/naming-conventions.md` "Adaptation 2":
lifecycle subdirectories (`pre-implementation/`, `implementation/`, `post-implementation/`) are
created on first real content, not pre-populated empty. All entries below are `not yet produced`.

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Path | Status |
|---|---|---|---|---|---|---|---|---|
| ART-W04-E01-S002-001 | `jobs_queue` lease-column migration | migration | implementation | Adds lease columns backed by W04-E01-S001's shared primitive; claim SQL assigns fresh token + `generation+1` | DATA-02 | W04-E01-S002-T001 | TBD at implementation time | not yet produced |
| ART-W04-E01-S002-002 | Fenced finalize code | source-code package | implementation | `complete`/`fail` paths compare lease token/generation and reject mismatch | DATA-02 | W04-E01-S002-T002 | TBD at implementation time | not yet produced |
| ART-W04-E01-S002-003 | Fenced `ReclaimStalled` code | source-code package | implementation | Bumps `lease_generation` on every reclaimed row | DATA-02 | W04-E01-S002-T003 | TBD at implementation time | not yet produced |
| ART-W04-E01-S002-004 | Lease-column schema and fencing-behavior documentation | documentation | post-implementation | Documents the lease-column schema, claim/finalize/reclaim fencing behavior, and rejection semantics | DATA-02 | W04-E01-S002-T001, W04-E01-S002-T002, W04-E01-S002-T003 | TBD at implementation time | not yet produced |
