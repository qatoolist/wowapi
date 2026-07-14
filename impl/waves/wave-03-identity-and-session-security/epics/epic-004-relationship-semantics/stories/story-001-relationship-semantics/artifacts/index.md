---
id: W03-E04-S001-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W03-E04-S001
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W03-E04-S001 — Artifacts index

Per mandate §9.2. Structure adaptation per `governance/naming-conventions.md` "Adaptation 2":
lifecycle subdirectories are created on first real content, not pre-populated empty. All entries
below are `not yet produced`.

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Path | Status |
|---|---|---|---|---|---|---|---|---|
| ART-W03-E04-S001-001 | `Checker.Has` party-subject evaluation logic | source-code change | implementation | Actor → active capacity → optional party resolution through the post-SEC-01 principal model | DATA-07 | W03-E04-S001-T001 | `kernel/relationship/relationship.go` | not yet produced |
| ART-W03-E04-S001-002 | Full subject-kind evaluation matrix | source-code change | implementation | Explicit evaluation branch for every live-requirement `subject_kind`; fail-closed default for unenumerated kinds | DATA-07 | W03-E04-S001-T002 | `kernel/relationship/relationship.go` | not yet produced |
| ART-W03-E04-S001-003 | Mutation-governance implementation | source-code change | implementation | Ownership check, attribution consumption (DATA-06 T2), audit write, versioning for edge create/revoke; cache-invalidation deferred-linked to W05-E04-S002 | DATA-07 | W03-E04-S001-T003 | TBD (mutation call sites) at implementation time | not yet produced |
