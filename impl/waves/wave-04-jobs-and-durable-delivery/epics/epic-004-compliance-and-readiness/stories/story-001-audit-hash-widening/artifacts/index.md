---
id: W04-E04-S001-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W04-E04-S001
status: accepted
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W04-E04-S001 — Artifacts index

Per mandate §9.2. Structure adaptation per `governance/naming-conventions.md` "Adaptation 2":
lifecycle subdirectories (`pre-implementation/`, `implementation/`, `post-implementation/`) are
created on first real content, not pre-populated empty. All entries below are `not yet produced`.

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Path | Status |
|---|---|---|---|---|---|---|---|---|
| ART-W04-E04-S001-001 | Widened chainHash implementation | source-code change | implementation | Covers every persisted field, including canonicalized metadata and tx_id | DATA-08 W6-T1 | W04-E04-S001-T001 | TBD at implementation time | not yet produced |
| ART-W04-E04-S001-002 | hash_version column migration | migration | implementation | Adds `hash_version smallint NOT NULL DEFAULT 1`, shipped through W02-E01's protocol, per D-04 | DATA-08 W6-T1, D-04 | W04-E04-S001-T001 | TBD at implementation time | not yet produced |
| ART-W04-E04-S001-003 | Version-branched Verify implementation | source-code change | implementation | Branches verification by hash_version: v1 historical scheme, v2 widened scheme | DATA-08 W6-T1, D-04 | W04-E04-S001-T001 | TBD at implementation time | not yet produced |
| ART-W04-E04-S001-004 | Audit hash-widening documentation | documentation | post-implementation | Documents the widened field list, metadata-canonicalization approach, hash_version value, and version-branch semantics | DATA-08 W6-T1 | W04-E04-S001-T001 | TBD at implementation time | not yet produced |
