---
id: W06-E02-S003-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W06-E02-S003
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W06-E02-S003 — Artifacts index

Per mandate §9.2. Structure adaptation per `governance/naming-conventions.md` "Adaptation 2":
lifecycle subdirectories (`pre-implementation/`, `implementation/`, `post-implementation/`) are
created on first real content, not pre-populated empty. All entries below are `not yet produced`.

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Path | Status |
|---|---|---|---|---|---|---|---|---|
| ART-W06-E02-S003-001 | OpenAPI semantic-diff gate | CI configuration | implementation | Classifies breaking OpenAPI changes per DX-06's 3.1/2020-12 baseline (blocked on W06-E02-S001) | REL-03 | W06-E02-S003-T001 | TBD at implementation time | not yet produced |
| ART-W06-E02-S003-002 | Event/schema compatibility-check mechanism | source-code + CI configuration | implementation | Ties an incompatible-bump check to a Compatibility mode (blocked on W06-E01-S001 + W05-E03) | REL-03 | W06-E02-S003-T002 | TBD at implementation time | not yet produced |
| ART-W06-E02-S003-003 | Generated-consumer upgrade check invocation | CI configuration | implementation | REL-03-scoped invocation of DX-04's upgrade-replay drill (blocked on W06-E01-S002) | REL-03 | W06-E02-S003-T003 | TBD at implementation time | not yet produced |
