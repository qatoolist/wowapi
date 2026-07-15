---
id: W01-E04-S003-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W01-E04-S003
status: complete
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W01-E04-S003 — Artifacts index

Per mandate §9.2. All artifacts produced 2026-07-13 by W01-E04-S003-T001/T002.

## Pre-implementation

| Artifact ID | Title | Type | Description | Source requirement | Producing task | Path | Version | Status |
|---|---|---|---|---|---|---|---|---|
| ART-W01-E04-S003-001 | Reproduction-run log collection | logs / existing-behaviour report | Full logs from every reproduction run, pass or fail — 16 files: preflight, 4× contaminated main-tree `-count=5 -parallel=4`, 4× isolated worktree `-count=5 -parallel=4`, 3× stress (e2e + concurrent testkit/cli companion), 1× `-race -count=2` | T-TEST-01 | W01-E04-S003-T001 | `../evidence/premier/T-TEST-01/logs/` (+ record `reproduction-runs.md`) | n/a | produced |
| ART-W01-E04-S003-002 | DB-wiring determination + diagnosis/decision record | design document (decision record) | `internal/e2e` uses its OWN wiring (raw `DATABASE_URL`, product migrate applies kernel migrations directly to the base DB; no `testkit.NewDB`, no `t.Parallel`); historical failure not reproduced (29/29 clean at pinned SHA); decision for T002: monitoring-only | T-TEST-01 | W01-E04-S003-T001 | `../evidence/premier/T-TEST-01/diagnosis-note.md` | n/a | produced |

## Implementation (conditional)

| Artifact ID | Title | Type | Description | Source requirement | Producing task | Path | Version | Status |
|---|---|---|---|---|---|---|---|---|
| ART-W01-E04-S003-003 | Monitoring-decision note (no code change) | design document | T002 resolved to task-002 illustrative branch 3 (non-reproduction → monitoring-only): `diagnosis-note.md` §5 (decision) + §6 (programme-level monitoring item); no production file changed | T-TEST-01 | W01-E04-S003-T002 | `../evidence/premier/T-TEST-01/diagnosis-note.md` §5-§6 | n/a | produced |

## Retention

All artifacts above are retained for the lifetime of the programme's traceability record per
`governance/artifact-policy.md`. The reproduction-run logs are retained even if a later run passes,
consistent with mandate §10's "do not delete earlier failed verification merely because a later run
passes" — every run's outcome is preserved.
