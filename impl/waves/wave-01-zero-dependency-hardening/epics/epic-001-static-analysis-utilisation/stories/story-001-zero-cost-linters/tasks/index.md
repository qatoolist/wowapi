---
id: W01-E01-S001-TASKS-INDEX
type: tasks-index
parent_story: W01-E01-S001
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W01-E01-S001 — Tasks index

Per mandate §16.4. Task files are single-file per the repository's documented adaptation (see
`governance/naming-conventions.md` "Adaptation 1") — each task file below contains its task
definition, implementation record, verification record, and deviations record as internal sections.

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W01-E01-S001-T001](task-001-zero-cost-linter-enablement.md) | Zero-cost linter enablement | W01Lint | done | none | `.golangci.yml` enabling 7 zero-cost analyzers; fail-before/pass-after run logs | AC-W01-E01-S001-01 | complete (2026-07-13) | verified |
| [W01-E01-S001-T002](task-002-noctx-fix.md) | noctx fix (2 sites) | W01Lint | done | none | `exec.CommandContext` at both named sites | AC-W01-E01-S001-02 | complete (2026-07-13) | verified |
| [W01-E01-S001-T003](task-003-copyloopvar-fix.md) | copyloopvar fix (1 site) | W01Lint | done | none | Removed pre-1.22 capture idiom at `app/maintenance.go:148` | AC-W01-E01-S001-03 | complete (2026-07-13) | verified |
| [W01-E01-S001-T004](task-004-pool-lifetime-config-keys.md) | MaxConnLifetime/MaxConnIdleTime config keys | W01Lint | done | none | New config keys + validation + unit test | AC-W01-E01-S001-04 | complete (2026-07-13) | verified |

## Grouping rationale

Per mandate §12 ("tasks must be decomposed when they... produce multiple unrelated outputs... need
separate evidence... have materially different risks"): the linter-enablement task (T001) is kept
separate from the two named code fixes (T002/T003) because the enablement task's evidence is a
config-plus-run-log pair covering seven analyzers at once, while each code fix has its own narrow
fail-before/pass-after evidence pair and touches a different file. T004 (new config keys) is kept
separate from all three linter-related tasks because it is not linter-driven at all — it is a new
capability (pool-lifetime configuration) that happens to share this story's CS-10/pgx-pool context,
not a triage output. Four tasks is the natural grain here: no further splitting (e.g. one task per
zero-cost analyzer) would add tracking value, since all seven share one enablement action and one
evidence artifact.
