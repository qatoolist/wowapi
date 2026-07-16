---
id: W02-E04-S001-TASKS-INDEX
type: tasks-index
parent_story: W02-E04-S001
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W02-E04-S001 — Tasks index

Per mandate §16.4. Task files are single-file per the repository's documented adaptation (see
`governance/naming-conventions.md` "Adaptation 1") — each task file below contains its task
definition, implementation record, verification record, and deviations record as internal sections.

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W02-E04-S001-T001](task-001-aggregate-repository-helper.md) | Typed aggregate repository/unit-of-work helper | unassigned | done | none | Helper bundling write+mirror+audit+outbox atomically; fault-injection test suite | AC-W02-E04-S001-01 | implemented | verified |
| [W02-E04-S001-T002](task-002-actor-attribution.md) | Actor-attribution fix (single owner, shared with DATA-07 T3) | unassigned | done | T001 | `registrar_pg.go` fix sourcing `created_by` from context; actor-attribution test | AC-W02-E04-S001-02 | implemented | verified |
| [W02-E04-S001-T003](task-003-reference-handler-migration.md) | Reference-handler migration | unassigned | done | T001, T002 | Reference handler migrated onto the new helper | AC-W02-E04-S001-03 | implemented | verified |
| [W02-E04-S001-T004](task-004-kernel-resource-docs-update.md) | `kernel/resource` documentation update | unassigned | done | T001 | Updated package documentation | AC-W02-E04-S001-04 | implemented | verified |
| [W02-E04-S001-T005](task-005-independent-review.md) | Independent review | unassigned | done | T001, T002, T003, T004 | Independent-review record per mandate §14 | AC-W02-E04-S001-01, AC-W02-E04-S001-02, AC-W02-E04-S001-03, AC-W02-E04-S001-04 | reviewed | verified |

## Grouping rationale

Per mandate §12, this task grouping follows PLAN DATA-06's own T1–T4 task table directly — each
PLAN task becomes one task file, since each already produces a distinct output with its own
acceptance criterion, test approach, and evidence type (T1's fault-injection suite; T2's actor-
attribution test; T3's regression-test confirmation; T4's manual documentation review), and each
carries a materially different risk (T1's is atomicity correctness; T2's is "must not break
legitimate system-actor call sites," PLAN's own named risk; T3's is "fix the reference pattern
before it's copied further"; T4's is "stale docs created this defect class"). No further splitting
or merging improves tracking value. T005 adds an independent-review task even though DATA-06 is P1
(not P0): T2 touches actor attribution — a security/accountability-relevant fix explicitly shared
with DATA-07 T3 as a single-owner fix surface (PLAN's own cross-cutting note 2) — and T1 establishes
a new atomicity guarantee other modules will come to rely on. Both properties warrant independent
review even at P1, consistent with this wave's general pattern of adding review tasks where a
task's blast radius (a shared fix surface consumed by a later, separate finding) or its correctness
property (transactional atomicity across 4 stages) is high-consequence if silently wrong.
