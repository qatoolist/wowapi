---
id: W07-E01-S002-TASKS-INDEX
type: tasks-index
parent_story: W07-E01-S002
status: complete
derived: false
created_at: 2026-07-12
updated_at: 2026-07-14
---

# W07-E01-S002 — Tasks index

Per mandate §16.4. Task files are single-file per the repository's documented adaptation (see
`governance/naming-conventions.md` "Adaptation 1") — each task file below contains its task
definition, implementation record, verification record, and deviations record as internal sections.

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W07-E01-S002-T001](task-001-index-definition-audit.md) | Index-definition audit (gap-fill) | W07-Scoping-Dispatch.W07E01S002 | done | none | Active-only claim confirmed before design | AC-W07-E01-S002-01 | complete | passed |
| [W07-E01-S002-T002](task-002-set-based-query-design.md) | Set-based query design | W07-Scoping-Dispatch.W07E01S002 | done | T001 | One precedence-preserving SQL statement | AC-W07-E01-S002-02 | complete | passed |
| [W07-E01-S002-T003](task-003-index-confirmation.md) | Index confirmation/addition | W07-Scoping-Dispatch.W07E01S002 | done | T001, T002 | Current/history index access proven | AC-W07-E01-S002-03 | complete | passed |
| [W07-E01-S002-T004](task-004-explain-fixtures.md) | EXPLAIN fixtures at representative cardinality | W07-Scoping-Dispatch.W07E01S002 | done | T003 | Four real PostgreSQL fixtures | AC-W07-E01-S002-04 | complete | passed |
| [W07-E01-S002-T005](task-005-parity-and-sql-count-tests.md) | Parity and SQL-count-constant tests | W07-Scoping-Dispatch.W07E01S002 | done | T002 | 3/10/50 parity and constant count | AC-W07-E01-S002-05 | complete | passed |
| [W07-E01-S002-T006](task-006-live-update-regression.md) | Live-update-visibility regression confirmation | W07-Scoping-Dispatch.W07E01S002 | done | T002 | No stale-read regression | AC-W07-E01-S002-06 | complete | passed |
| [W07-E01-S002-T007](task-007-publication.md) | Publication against perf/reference-v1.json | W07-Scoping-Dispatch.W07E01S002 | done | T004, T005, T006; cross-story W07-E01-S001-T001 | DEC-Q9-honest relative comparison | AC-W07-E01-S002-06 | complete | passed |

## Grouping rationale

Per mandate §12: T001-T006 follow PLAN PERF-03's own T0-T5 task table exactly, preserving PLAN's own
explicit T0-must-precede-T2 sequencing note. T007 (publication, PLAN's own T6) is sequenced last since
it consumes T004/T005/T006's own evidence and W07-E01-S001's shared reference environment. No
independent-review task is added: this story is P1 but its own scope is a single, well-bounded query
rewrite with a clear correctness bar (result parity) that this story's own T002/T005/T006 already
directly verify — a dedicated review task would not add tracking value beyond what
`governance/definition-of-done.md`'s own closure-time review already requires for a 7-task story of this
size.
