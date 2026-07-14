---
id: W05-E03-S002-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W05-E03-S002
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W05-E03-S002 — Artifacts index

Per mandate §9.2. Structure adaptation per `governance/naming-conventions.md` "Adaptation 2". All
entries below are `not yet produced`.

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Path | Status |
|---|---|---|---|---|---|---|---|---|
| ART-W05-E03-S002-001 | Duplicate-collector rejection | source-code change | implementation | Every collector rejects a second write; preserves legitimate multi-locale accumulation | AR-04 | W05-E03-S002-T001 | TBD at implementation time | not yet produced |
| ART-W05-E03-S002-002 | Empty-required-fragment rejection | source-code change | implementation | A required-but-empty fragment fails boot | AR-04 | W05-E03-S002-T002 | TBD at implementation time | not yet produced |
| ART-W05-E03-S002-003 | Post-seal config/namespace/collector rejection extension | source-code change | implementation | Extends AR-01 T8's error-not-panic contract | AR-04 | W05-E03-S002-T003 | TBD at implementation time | not yet produced |
| ART-W05-E03-S002-004 | Shared no-op-adapter waiver mechanism | source-code package | implementation | Prod readiness gate on required-but-no-op/missing adapter without waiver; shared primitive for SEC-06/DX-07 | AR-04 | W05-E03-S002-T004 | TBD at implementation time | not yet produced |
