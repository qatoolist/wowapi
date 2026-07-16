---
id: W06-E03-S001-TASKS-INDEX
type: tasks-index
parent_story: W06-E03-S001
status: verified
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W06-E03-S001 — Tasks index

Per mandate §16.4. Task files are single-file per the repository's documented adaptation (see
`governance/naming-conventions.md` "Adaptation 1") — each task file below contains its task
definition, implementation record, verification record, and deviations record as internal sections.

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W06-E03-S001-T001](task-001-manifest-schema-and-validator.md) | Manifest schema and validator | W06E03Impl | done | none | schema + validator | AC-01 | implemented | verified |
| [W06-E03-S001-T002](task-002-wave0-manifest-entries.md) | Wave-0 manifest entries | W06E03Impl | done | T001 | complete gate catalog | AC-02 | implemented | verified |
| [W06-E03-S001-T003](task-003-required-gates-workflow.md) | reusable required gates | W06E03Impl | done | T002 | SHA-bound attested results | AC-03 | implemented | verified |
| [W06-E03-S001-T004](task-004-ci-yml-wiring.md) | CI wiring | W06E03Impl | done | T003 | shared caller path | AC-04 | implemented | verified |
| [W06-E03-S001-T005](task-005-release-verify-job.md) | release verification | W06E03Impl | done | T004 | failed gate blocks build | AC-05 | implemented | verified |
| [W06-E03-S001-T006](task-006-build-candidate-split.md) | immutable candidate | W06E03Impl | done | T005 | no-publish build | AC-06 | implemented | verified |
| [W06-E03-S001-T007](task-007-publish-job-scaffolding.md) | exact-byte protected publisher | W06E03Impl | done | T006 | draft `gh`/ORAS publisher | AC-07 | implemented | verified |
| [W06-E03-S001-T008](task-008-verify-release-script.md) | clean release verifier | W06E03Impl | done | T007 | golden failures + clean job | AC-08 | implemented | verified |
| [W06-E03-S001-T009](task-009-independent-review.md) | Independent review | W06-E01-E04-Execution.W06E03ReviewR | done | T001-T008 | review-only, no-open-issues | AC-01..AC-08 | complete | passed |

## Grouping rationale

Per mandate §12: T001-T008 follow PLAN REL-01's own T1-T8 task table exactly, in the same strict
dependency order PLAN itself specifies (each task's own "Depends-on" column chains T1→T2→T3→T4→T5→T6→
T7→T8). This is a P0/critical story — REL-01 is the framework's core release-time trust boundary — so
T009 adds an independent-review task per mandate §14, specifically scoped to confirm this story's own
buildable-now boundary is honestly stated against W06-E03-S002's real-protected-environment remainder,
not silently overclaimed as a complete end-to-end proof.
