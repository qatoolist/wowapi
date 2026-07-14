---
id: DEV-W02-E05-S001
type: deviations-record
parent_story: W02-E05-S001
status: complete
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Deviations record — W02-E05-S001

## Plan revision (not a deviation)

T001's design decision record ratified the following revisions to `plan.md`'s initial task ordering
and boundaries; these are recorded here so the implementation trace matches the actual work:

1. **Execution order:** T003 (manifest/hash canonicalization) → T002 (Apply/idempotency/RLS) →
   T005 (audit folded into Apply + readiness hash reporting + docs) → T004 (readiness-check
   registration). The audit table and audit-row production land with `seeds.Apply` rather than as a
   separate later task, so T003/T002/T005 are tightly coupled and completed together.
2. **CLI shape:** `--env prod` from CS-21's sketch was rejected; the CLI remains a bare
   `DATABASE_URL` escape hatch (`wowapi seed sync --module ...`), with `--dry-run` added.
3. **Audit location:** a new global `seed_sync_runs` table was chosen over `kernel/audit` because
   seed-sync is tenant-less and forcing a synthetic tenant would corrupt the audit hash chain.
   This was reviewed and confirmed as a story-scoped table, not a D-0N escalation.

No implementation diverged from the ratified design.

## Deviations

None.
