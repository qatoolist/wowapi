---
id: W03-E01-S004-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W03-E01-S004
status: accepted
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W03-E01-S004 — Artifacts index

Per mandate §9.2. Structure adaptation per `governance/naming-conventions.md` "Adaptation 2":
lifecycle subdirectories are created on first real content, not pre-populated empty. All entries
below are `not yet produced`.

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Path | Status |
|---|---|---|---|---|---|---|---|---|
| ART-W03-E01-S004-001 | Sequencing plan | runbook / design document | post-implementation | Repo-by-repo cutover order; named wowsociety files/tests needing rework | SEC-01 (PROD-04) | W03-E01-S004-T001 | `sequencing-plan.md` | produced |
| ART-W03-E01-S004-002 | Staging-validation plan | runbook | post-implementation | How T2/T5 are validated against wowsociety staging data before production enforcement | SEC-01 (PROD-04) | W03-E01-S004-T002 | `staging-validation-plan.md` | produced |
| ART-W03-E01-S004-003 | Rollback plan | runbook | post-implementation | Revert steps for both cutover failure directions | SEC-01 (PROD-04) | W03-E01-S004-T003 | `rollback-plan.md` | produced |
