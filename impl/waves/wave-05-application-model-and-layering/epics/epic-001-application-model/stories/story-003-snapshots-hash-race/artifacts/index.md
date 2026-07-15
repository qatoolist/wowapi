---
id: W05-E01-S003-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W05-E01-S003
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W05-E01-S003 — Artifacts index

Per mandate §9.2. Structure adaptation per `governance/naming-conventions.md` "Adaptation 2":
lifecycle subdirectories are created on first real content, not pre-populated empty. All entries
below are `not yet produced`.

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Path | Status |
|---|---|---|---|---|---|---|---|---|
| ART-W05-E01-S003-001 | Snapshot-immutability conversion | source-code change | implementation | Converts every exported registry reader to cloned/immutable data | AR-01 | W05-E01-S003-T001 | TBD at implementation time | not yet produced |
| ART-W05-E01-S003-002 | Post-seal Context/registrar retention rejection | source-code change | implementation | Explicit error on retained-registrar post-boot mutation, validated against wowsociety's pattern | AR-01 | W05-E01-S003-T002 | TBD at implementation time | not yet produced |
| ART-W05-E01-S003-003 | Deterministic model-hash function | source-code package | implementation | Byte-identical hash for identical compiles; emitted at startup/readiness | AR-01 | W05-E01-S003-T003 | TBD at implementation time | not yet produced |
| ART-W05-E01-S003-004 | Race-test suite | test code | implementation | Proves no runtime mutation of the sealed model | AR-01 | W05-E01-S003-T004 | TBD at implementation time | not yet produced |
