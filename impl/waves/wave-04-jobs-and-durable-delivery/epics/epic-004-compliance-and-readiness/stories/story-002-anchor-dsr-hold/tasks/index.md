---
id: W04-E04-S002-TASKS-INDEX
type: tasks-index
parent_story: W04-E04-S002
status: recorded
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W04-E04-S002 — Tasks index

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W04-E04-S002-T001](task-001-external-anchor-verification.md) | External anchor verification | W03-E02-E03-E04-E05-Rerun | done | none | Anchor mechanism + anchor-then-tamper detection test | AC-W04-E04-S002-01 | implemented | verified |
| [W04-E04-S002-T002](task-002-dsr-export-artifact.md) | Encrypted immutable DSR export artifact | W03-E02-E03-E04-E05-Rerun | done | none | Artifact writer (manifest, checksum, encryption, access policy, download audit) + completion/checksum test | AC-W04-E04-S002-02 | implemented | verified |
| [W04-E04-S002-T003](task-003-central-legal-hold-enforcement.md) | Central legal-hold enforcement wrapper | W03-E02-E03-E04-E05-Rerun | done | none | RecordClass enumeration + wrapper + negative test | AC-W04-E04-S002-03, AC-W04-E04-S002-04 (enumeration half) | implemented | verified |
| [W04-E04-S002-T004](task-004-explicit-per-class-status.md) | Explicit partial/not-applicable per-class DSR status | W03-E02-E03-E04-E05-Rerun | done | T002, T003 | Explicit-status reporting mechanism + test | AC-W04-E04-S002-04 (status half) | implemented | verified |
| [W04-E04-S002-T005](task-005-independent-review.md) | Independent review | reviewer | todo | T001, T002, T003, T004 | Independent-review record per mandate §14 | AC-W04-E04-S002-01, AC-W04-E04-S002-02, AC-W04-E04-S002-03, AC-W04-E04-S002-04 | n/a | pending |

## Grouping rationale

Per mandate §12: T001/T002/T003/T004 map one-to-one onto PLAN DATA-08's own W6-T2/T3/T4/T5 rows,
each producing unrelated outputs with separate evidence and materially different risks. T004 depends
on T002 (manifest shape) and T003 (enumeration/wrapper). T005 adds a mandatory independent-review
task per mandate §14, with specific attention to the `RecordClass` enumeration predating the wrapper
and the artifact being gated on write success.
