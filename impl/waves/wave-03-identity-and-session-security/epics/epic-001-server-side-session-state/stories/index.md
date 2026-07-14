---
id: W03-E01-STORIES-INDEX
type: stories-index
epic: W03-E01
wave: W03
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W03-E01 — Stories index

| Story | Title | Status | Priority | Source requirement | Tasks | Objective |
|---|---|---|---|---|---|---|
| [W03-E01-S001](story-001-grant-schema-and-membership/story.md) | grant-schema-and-membership | planned | P0 | SEC-01 (T1–T3) | 4 | Build the `identity_grant` table; make tenant-membership verification unconditional; reject zero/unknown tenant claims pre-transaction |
| [W03-E01-S002](story-002-capacity-and-privileged-resolver/story.md) | capacity-and-privileged-resolver | planned | P0 | SEC-01 (T4–T5) | 3 | Require explicit server-side capacity selection; replace direct claim copy of impersonation/break-glass with a verified grant-table resolver |
| [W03-E01-S003](story-003-assurance-and-credential-schemes/story.md) | assurance-and-credential-schemes | planned | P0 | SEC-01 (T6–T7) | 3 | Bind auth_time/acr/amr into assurance and enforce step-up freshness; distinguish credential schemes explicitly |
| [W03-E01-S004](story-004-cross-repo-cutover-plan/story.md) | cross-repo-cutover-plan | planned | P0 | SEC-01 (PROD-04 coordination) | 3 | Produce the sequencing, staging-validation, and rollback plan for wowsociety's two-repo impersonation cutover |
