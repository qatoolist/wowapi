---
id: W01-E01-S003-TASKS-INDEX
type: tasks-index
parent_story: W01-E01-S003
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W01-E01-S003 — Tasks index

Per mandate §16.4. Task files are single-file per the repository's documented adaptation (see
`governance/naming-conventions.md` "Adaptation 1") — each task file below contains its task definition,
implementation record, verification record, and deviations record as internal sections.

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W01-E01-S003-T001](task-001-go-mod-verify-ci-step.md) | `go mod verify` CI step | W01Lint | done | none | Updated `ci.yml` with a passing `go mod verify` step | AC-W01-E01-S003-01 | complete (2026-07-13) | verified |
| [W01-E01-S003-T002](task-002-license-signal-enablement.md) | License-scanning signal enablement | W01Lint | done | none | Updated `security-scan.yml` (Trivy `license` scanner, planned choice carried through) with documented choice | AC-W01-E01-S003-02 | complete (2026-07-13) | verified |
| [W01-E01-S003-T003](task-003-pre-push-hook-db-skip-fix.md) | Pre-push hook DB-silent-skip fix | W01Lint | done | none | Updated `.githooks/pre-push` failing loudly instead of silently skipping DB tests | AC-W01-E01-S003-04 | complete (2026-07-13) | verified |
| [W01-E01-S003-T004](task-004-nightly-fuzz-schedule-confirmation.md) | Nightly fuzz-schedule confirmation | W01Lint | done | none | Confirmation/audit note on `ci.yml`'s nightly schedule wiring (observed run 29229288699), with the `-fuzz=` gap restated as W07 scope | AC-W01-E01-S003-03 | complete (2026-07-13; no code change needed) | verified |

## Grouping rationale

Per mandate §12 ("split when materially different risk/evidence/ownership; don't over-fragment") and
the epic's own instruction that S003 decomposes into 3 tasks (go-mod-verify CI step; license-signal
enablement; pre-push hook fix), with the nightly-fuzz-schedule confirmation either folded into one of
those three or given a 4th task, at this story's discretion, documented:

**Decision: four tasks.** T001 (`go mod verify`), T002 (license signal), and T003 (pre-push hook) are
the three baseline tasks the epic names directly — each touches a different file (`ci.yml`,
`security-scan.yml`, `.githooks/pre-push` respectively) and each has a materially different fix
mechanism (a new CI command; a scanner-list configuration change; a shell-script logic change), so none
of the three is a candidate for merging with another.

The nightly-fuzz-schedule confirmation is given its own task, **T004**, rather than being folded into
T001 even though both are `ci.yml`-focused. `plan.md`'s "Task breakdown grouping decision" section gives
the full reasoning; in summary:

- **Evidence type differs.** T001's evidence is a standard CI run log (a new step executing and
  passing). T004's evidence is a confirmation/audit note — a record of what was inspected and observed
  about pre-existing wiring, not a diff-driven pass/fail log for a newly-added step. Folding a
  confirmation-type record into a code-addition task's evidence would blur two evidence types under one
  task ID.
- **Risk profile differs.** T001's risk is near-zero (a standard, well-understood toolchain command).
  T004's risk is a scope-boundary risk — the danger of either under-verifying (accepting the existing
  header-comment claims about the nightly schedule at face value without confirming) or over-verifying
  (drifting into implementing the `-fuzz=` flag wiring, which is W07's job, not this story's). This
  distinct risk shape justifies independent tracking.
- **Independent reviewability.** The epic's own acceptance criterion AC-W01-E01-04 calls out, by name,
  that S003 must be "specifically checked for the nightly-fuzz scope boundary being honestly stated" —
  a dedicated task makes that specific check independently traceable in `tasks/index.md` and
  independently closable, rather than bundled inside T001's own completion record.

This keeps the task count at four without over-fragmenting: no further splitting (e.g., separating the
license-signal *decision* from its *implementation*, or separating "inspect the schedule" from "observe
a run" within T004) is warranted, since each of those pairs shares one owner, one evidence artifact, and
one risk profile.
