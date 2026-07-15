---
id: W02-E01-S003-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W02-E01-S003
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W02-E01-S003 — Artifacts index

Per mandate §9.2. Structure adaptation per `governance/naming-conventions.md` "Adaptation 2":
lifecycle subdirectories are created on first real content, not pre-populated empty. All entries
below are `not yet produced`.

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Path | Status |
|---|---|---|---|---|---|---|---|---|
| ART-W02-E01-S003-001 | Canary/deploy-N tooling | source-code package | implementation | N-1/N alongside deployment orchestration and soak-metric collection | DATA-09 | W02-E01-S003-T001 | TBD at implementation time | not yet produced |
| ART-W02-E01-S003-002 | Switch-phase tooling | source-code package | implementation | Observable compatibility flag, dual-schema-version consumer support, application-rollback mechanics | DATA-09 | W02-E01-S003-T002 | TBD at implementation time | not yet produced |
| ART-W02-E01-S003-003 | Contract-phase gate | source-code package | implementation | Evidenced no-N-1-remains precondition check, fail-closed | DATA-09 | W02-E01-S003-T003 | TBD at implementation time | not yet produced |
| ART-W02-E01-S003-004 | CI drill pipeline definition | deployment manifest / CI workflow | implementation | Scheduled workflow running all six directive-named drills | DATA-09 | W02-E01-S003-T004 | `.github/workflows/` (exact file TBD) | not yet produced |
| ART-W02-E01-S003-005 | Consolidated 6-drill evidence bundle | compatibility matrix / evidence bundle | post-implementation | Aggregates T001–T004's individual drill outputs into one reviewable record | DATA-09 | W02-E01-S003-T005 | TBD at implementation time | not yet produced |
| ART-W02-E01-S003-006 | Canary/switch/contract/pipeline documentation | documentation | post-implementation | Configuration surfaces, human-decision boundaries, soak-calibration gap | DATA-09 | W02-E01-S003-T001, T002, T003, T004 | TBD at implementation time | not yet produced |
