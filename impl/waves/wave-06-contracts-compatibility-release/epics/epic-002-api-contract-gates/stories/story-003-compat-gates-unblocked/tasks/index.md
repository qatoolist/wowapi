---
id: W06-E02-S003-TASKS-INDEX
type: tasks-index
parent_story: W06-E02-S003
status: blocked
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W06-E02-S003 — Tasks index

Per mandate §16.4. Task files are single-file per the repository's documented adaptation (see
`governance/naming-conventions.md` "Adaptation 1") — each task file below contains its task
definition, implementation record, verification record, and deviations record as internal sections.

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W06-E02-S003-T001](task-001-openapi-semantic-diff.md) | OpenAPI semantic diff | W06E02Impl | blocked | W06-E02-S001 accepted | OpenAPI semantic-diff gate | AC-W06-E02-S003-01 | not started | blocked |
| [W06-E02-S003-T002](task-002-event-schema-compatibility.md) | Event/schema compatibility | W06E02Impl | blocked | W06-E01-S001 accepted AND W05-E03 accepted | Event/schema compatibility check | AC-W06-E02-S003-02 | not started | blocked |
| [W06-E02-S003-T003](task-003-generated-consumer-upgrade-check.md) | Generated-consumer upgrade check | W06E02Impl | blocked | W06-E01-S002 accepted | Generated-consumer upgrade check | AC-W06-E02-S003-03 | not started | blocked |
| [W06-E02-S003-T004](task-004-independent-review.md) | Independent review | W06-E01-E04-Execution.W06E02ReviewFinal | done | T001-T003 | Independent blocked-leg review | all | review complete | PASS, blockers honest |

## Grouping rationale

Per mandate §12: T001, T002, T003 map directly to PLAN REL-03's own T3, T5, T7 (kept as three
separate tasks because each is entry-gated on a genuinely distinct set of unblocking stories, per this
story's own explicit design requirement to state per-leg blocked-entry criteria rather than a single
opaque story-level "blocked" status). T004 adds an independent-review task scoped specifically to
re-checking that each completed leg's entry criterion was genuinely satisfied before implementation —
this is the exact failure mode (a bypassed entry criterion) this story's whole structure exists to
prevent, so the review task's own scope is written to catch it explicitly, not merely to re-run the
standard mandate §14 checklist generically.
