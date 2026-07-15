---
id: DEV-W01-E01-S003
type: deviations-record
parent_story: W01-E01-S003
status: recorded
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Deviations record — W01-E01-S003

Implementation (2026-07-13, W01Lint) matched the approved `plan.md` — the "License signal decision"
(Trivy license scanner) and the "Task breakdown grouping decision" (T004 separate) were both carried
through unchanged, and the fresh re-read confirmed `story.md`'s current-state citations still hold at
HEAD `0a31186` (line numbers verified: trivy `scanners:` at `security-scan.yml:71`,
`dependency-review` gate at `:80,93`, hook test line at `.githooks/pre-push:21-22`, cron at
`ci.yml:42-46`). No plan text was altered. Two notes below the deviation threshold, recorded for
completeness rather than as approved-plan divergences:

1. **Opt-out mechanism added beyond the minimal fix** — `WOWAPI_PREPUSH_SKIP_DB=1` (loudly announced
   skip) alongside the loud-failure default. Within task T003's own step-3 language ("…or how to
   explicitly opt out"), so scope-conformant, not a deviation; noted because it introduces one new
   hook-local env var.
2. **AC-01/AC-02 evidence is local-execution + actionlint, not an in-CI run log** — workers cannot
   push (conductor owns commits), so the CI run log for the new steps cannot exist yet. Explicit
   carry-forward rationale in `verification.md`; the wave gate's first CI run is to be registered as
   `retested` evidence superseding these records.
