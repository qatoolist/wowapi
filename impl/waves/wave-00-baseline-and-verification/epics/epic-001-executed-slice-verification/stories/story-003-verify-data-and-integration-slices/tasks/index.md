---
id: W00-E01-S003-TASKS-INDEX
type: tasks-index
parent_story: W00-E01-S003
status: complete
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W00-E01-S003 — Tasks index

Per mandate §16.4.

| Task | Title | Owner | Status | Dependencies | Output | Related ACs | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W00-E01-S003-T001](task-001-reverify-data08-w0-attachment-notify-durability.md) | Re-verify DATA-08 W0 (attachment/notify durability) | unassigned | done | none (parallel-safe within story) | EV-W00-E01-S003-01 registered (pass); log `../evidence/logs/t001-db-gated-attachment-notify.log` | AC-W00-E01-S003-01 | complete (verification-only) | pass — 2026-07-13 @ 0a31186 |
| [W00-E01-S003-T002](task-002-reverify-rel04-s3-totp-and-ci-pipeline-state.md) | Re-verify REL-04 T1-T4 (S3/TOTP wiring) and confirm SD-01/SD-02 CI-pipeline-state | unassigned | done | none (parallel-safe within story) | EV-W00-E01-S003-02 registered (pass / confirmed-no-drift); logs + inspection note under `../evidence/logs/` | AC-W00-E01-S003-02 | complete (verification-only) | pass — 2026-07-13 @ 0a31186 |
| [W00-E01-S003-T003](task-003-repin-cs03-cs19-cs24-verify-outcomes.md) | Re-pin CS-03/CS-19/CS-24 matrix verify-outcomes | unassigned | done | none (parallel-safe within story; soft convenience overlap with T002's CI-inspection work, not a hard dependency) | EV-W00-E01-S003-03 registered (pass — all three still hold); re-pin note `../evidence/logs/t003-cs-repin-note.md` | AC-W00-E01-S003-03 | complete (verification-only) | pass — 2026-07-13 @ 0a31186 |
