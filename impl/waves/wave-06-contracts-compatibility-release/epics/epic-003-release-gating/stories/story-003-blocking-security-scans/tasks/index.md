---
id: W06-E03-S003-TASKS-INDEX
type: tasks-index
parent_story: W06-E03-S003
status: verified
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W06-E03-S003 — Tasks index

Per mandate §16.4. Task files are single-file per the repository's documented adaptation (see
`governance/naming-conventions.md` "Adaptation 1") — each task file below contains its task
definition, implementation record, verification record, and deviations record as internal sections.

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W06-E03-S003-T001](task-001-trivy-blocking-flip.md) | Blocking scoped Trivy | W06E03Impl | done | none | CRITICAL/HIGH and seeded defects block | AC-01 | implemented | verified |
| [W06-E03-S003-T002](task-002-waiver-mechanism.md) | Expiring scoped waivers | W06E03Impl | done | none | schema + ignore synchronization | AC-02 | implemented | verified |
| [W06-E03-S003-T003](task-003-visibility-guard-meta-check.md) | Hosted visibility meta-check | W06E03Impl | done | none | exact-SHA hosted proof | AC-03 | implemented | verified |
| [W06-E03-S003-T004](task-004-local-scanner-fallback.md) | Private local fallback | W06E03Impl | done | none | fail-closed SAST/posture fallback | AC-04 | implemented | verified |
| [W06-E03-S003-T005](task-005-manifest-wiring.md) | Manifest wiring | W06E03Impl | done | T001-T004 | one gate per scanner + artifact reports | AC-05 | implemented | verified |
| [W06-E03-S003-T006](task-006-independent-review.md) | Independent review | W06-E01-E04-Execution.W06E03ReviewR | done | T001-T005 | review-only, no-open-issues | AC-01..AC-05 | complete | passed |

## Grouping rationale

Per mandate §12: T001-T004 follow PLAN REL-02's own T1-T4 task table exactly, kept as four
separate tasks because each targets a genuinely disjoint scanner/mechanism (Trivy, the waiver schema,
the visibility-guard meta-check, the local-scanner fallback) with its own separately-evidenced
acceptance bar. T005 (manifest wiring) is sequenced last because it depends on all four existing.
This story is P0/P1 (REL-02's own PLAN priority: "P0/P1") — T006 adds an independent-review task per
mandate §14, given this story's direct security-posture consequence for every future PR once its gates
are required.
