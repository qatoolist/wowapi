---
id: W06-E02-S002-TASKS-INDEX
type: tasks-index
parent_story: W06-E02-S002
status: accepted
derived: false
created_at: 2026-07-12
updated_at: 2026-07-14
---

# W06-E02-S002 — Tasks index

Per mandate §16.4. Task files are single-file per the repository's documented adaptation (see
`governance/naming-conventions.md` "Adaptation 1") — each task file below contains its task
definition, implementation record, verification record, and deviations record as internal sections.

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W06-E02-S002-T001](task-001-go-public-api-diff.md) | Go public API diff | W06E02Impl | done | none | API diff CI gate | AC-W06-E02-S002-01 | implemented | verified |
| [W06-E02-S002-T002](task-002-module-compile-matrix.md) | Module compile matrix | W06E02Impl | done | none | Compile matrix with explicit exclusions | AC-W06-E02-S002-02 | implemented | verified locally under both exact toolchains |
| [W06-E02-S002-T003](task-003-config-schema-compatibility.md) | Config schema compatibility | W06E02Impl | done | none | Config compat gate | AC-W06-E02-S002-03 | implemented | verified |
| [W06-E02-S002-T004](task-004-migration-upgrade-drill.md) | Migration upgrade-from-oldest-supported drill | W06E02Impl | done | none | Extended reversibility test | AC-W06-E02-S002-04 | implemented | verified |
| [W06-E02-S002-T005](task-005-container-architecture-smoke.md) | Container architecture smoke | W06E02Impl | done | REL-01 candidate layout | Exact pre-publish architecture smoke | AC-W06-E02-S002-05 | implemented | real amd64/arm64 digest run PASS |
| [W06-E02-S002-T006](task-006-sbom-provenance-fold-in.md) | SBOM/provenance/signature verification fold-in | W06E02Impl | done | shared REL-01 verifier | REL-03 T9 cross-reference | AC-W06-E02-S002-06 | implemented without duplication | 12-test golden suite PASS |
| [W06-E02-S002-T007](task-007-independent-review.md) | Independent review | W06-E02-S002-Rerun | done | T001-T006 | Independent-review record per mandate §14 | all | complete | PASS, no production findings |

## Grouping rationale

Per mandate §12: T001–T006 follow PLAN REL-03's own T1/T2/T4/T6/T8/T9 numbering exactly (renumbered
sequentially here per this programme's task-numbering convention, with each task's own header noting
its source REL-03 T-number), kept as six separate tasks because each targets a genuinely disjoint code
surface (API surface, compile matrix, config schema, migration reversibility, container images, supply-
chain evidence) with its own separately-evidenced acceptance bar — collapsing any two into one task
would blur two independently-reviewable outcomes into one. T007 adds an independent-review task per
mandate §14, consistent with how this programme treats P1 stories carrying meaningful downstream
CI-gate consequence (these six gates, once required, block every future PR).
