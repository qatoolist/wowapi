---
id: W02-E02-S002-TASKS-INDEX
type: tasks-index
parent_story: W02-E02-S002
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W02-E02-S002 — Tasks index

Per mandate §16.4. Task files are single-file per the repository's documented adaptation (see
`governance/naming-conventions.md` "Adaptation 1") — each task file below contains its task
definition, implementation record, verification record, and deviations record as internal sections.

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W02-E02-S002-T001](task-001-mismatch-audit.md) | Mismatch audit (DATA-01 T3) | unassigned | todo | none (soft: more useful after W02-E02-S001-T002's scanner confirms the 8-edge inventory) | Zero-mismatch report against staging/prod-shaped data, or a documented remediation-decision record | AC-W02-E02-S002-01 | not started | not started |
| [W02-E02-S002-T002](task-002-composite-fk-notvalid.md) | Composite FK `NOT VALID` add, all 8 edges (DATA-01 T4) | unassigned | todo | T001; **hard gate: W02-E01-S001 and W02-E01-S002 both `accepted`** | 8 per-table `NOT VALID` composite FK adds | AC-W02-E02-S002-02 | not started | not started |
| [W02-E02-S002-T003](task-003-validate-constraint.md) | `VALIDATE CONSTRAINT` each new composite FK (DATA-01 T5) | unassigned | todo | T002; **hard gate: W02-E01-S001 and W02-E01-S002 both `accepted`** | 8 validated composite FKs; second zero-mismatch confirmation | AC-W02-E02-S002-03 | not started | not started |
| [W02-E02-S002-T004](task-004-cross-tenant-negative-tests.md) | Seeded cross-tenant insert negative tests, both roles (DATA-01 T7) | unassigned | todo | T003 | Catalog-driven RLS matrix test, cross-tenant insert fails under `app_rt` and `app_platform` | AC-W02-E02-S002-04 | not started | not started |
| [W02-E02-S002-T005](task-005-redundant-fk-cleanup.md) | Optional redundant single-column FK cleanup (DATA-01 T8) | unassigned | todo | T003, T004 | FK-removal migration + consumer/rollback verification record, or explicit deferral | AC-W02-E02-S002-05 | not started | not started |
| [W02-E02-S002-T006](task-006-independent-review.md) | Independent review | unassigned | todo | T001, T002, T003, T004, T005 | Review report confirming the W02-E01 gate was honored and the mismatch-audit outcome honestly recorded | AC-W02-E02-S002-01, AC-W02-E02-S002-02, AC-W02-E02-S002-03, AC-W02-E02-S002-04, AC-W02-E02-S002-05 | not started | not started |

## Grouping rationale

Per mandate §12: T001 (the mismatch audit) is its own task because it is read-only, has its own
distinct evidence type (a dated audit report), and its outcome branches this story's entire
sequencing — T002/T003 cannot begin until T001 resolves to zero-mismatch or a resolved remediation
decision, per RISK-W02-002. T002 (`NOT VALID` add) and T003 (`VALIDATE CONSTRAINT`) are kept as
separate tasks, matching PLAN's own separate T4/T5 rows, because they have materially different risk
profiles and evidence types — T002's evidence is a migration lock-duration report, T003's is a
concurrent-writer-load report plus a second zero-mismatch confirmation — and because both carry the
same hard cross-wave gate (W02-E01-S001 and W02-E01-S002 both `accepted`) independently restated in
each task's own "Dependencies" section per this programme's established pattern (see
`../story-001-parent-indexes-scanner-gate/tasks/` and `../../../../wave-03-identity-and-session-security/epics/epic-001-server-side-session-state/stories/story-001-grant-schema-and-membership/tasks/`),
so a task-level reader cannot miss the gate by reading only one task file. T004 (cross-tenant negative
tests) is separate because it is a distinct acceptance criterion (DATA-01 T7, its own PLAN row) with
its own adversarial-test evidence type, even though it consumes T003's validated constraints. T005
(optional FK cleanup) is kept as its own task, matching PLAN's own separate T8 row and its own
"optional — don't block P0 closure" framing, so its non-completion is trackable as an intentional
deferral rather than an unresolved task. Per this story's P0 priority and MATRIX CS-18's "top-ranked"
framing (mandate §14's independent-review requirement for critical stories), T006 adds independent
review as its own tracked unit, scoped specifically to confirming the W02-E01 gate was genuinely
honored on T002/T003 and the mismatch-audit outcome was honestly recorded — consistent with this
story's own `plan.md` "Task breakdown" (W02-E02-S002-T001 through -T006).
