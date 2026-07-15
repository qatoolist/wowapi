---
id: W05-E04-S002-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W05-E04-S002
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W05-E04-S002 — Artifacts index

Per mandate §9.2. Structure adaptation per `governance/naming-conventions.md` "Adaptation 2". All
entries below are `not yet produced`.

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Path | Status |
|---|---|---|---|---|---|---|---|---|
| ART-W05-E04-S002-001 | Bounded, sharded authz cache | source-code change | implementation | golang-lru/v2-backed, sized by config | SEC-04 | W05-E04-S002-T001 | TBD at implementation time | not yet produced |
| ART-W05-E04-S002-002 | Eviction + admission/eviction metrics | source-code change | implementation | Full metric set for cache eviction | SEC-04 | W05-E04-S002-T002 | TBD at implementation time | not yet produced |
| ART-W05-E04-S002-003 | Singleflight miss-collapse | source-code change | implementation | N concurrent misses → 1 DB load | SEC-04 | W05-E04-S002-T003 | TBD at implementation time | not yet produced |
| ART-W05-E04-S002-004 | authz_epoch table + epoch-bump wiring | schema + source-code change | implementation | Per-tenant epoch table (D-06); framework-side mutation paths bump epoch in-tx | SEC-04 | W05-E04-S002-T004 | TBD at implementation time | not yet produced |
| ART-W05-E04-S002-005 | Decision provenance metadata + prod-config gate | source-code change | implementation | CacheHit/epoch-observed on Decision; prod boot fails without explicit bound | SEC-04 | W05-E04-S002-T005 | TBD at implementation time | not yet produced |
