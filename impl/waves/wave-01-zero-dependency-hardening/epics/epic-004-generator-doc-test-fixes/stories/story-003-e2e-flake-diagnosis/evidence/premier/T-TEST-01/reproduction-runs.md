---
id: EV-W01-E04-S003-001
type: evidence-record
title: Reproduction-run log collection — repeated -count/-parallel/-race/stress executions
parent_story: W01-E04-S003
parent_task: W01-E04-S003-T001
status: final
created_at: 2026-07-13
---

# Evidence record — T-TEST-01 reproduction runs

Mandate §10 required fields:

- **Evidence ID:** EV-W01-E04-S003-001 (this record covers the whole run collection; the
  contaminated subset run-01..04 carries status `failed`, superseded by run-05..08 `retested`).
- **Evidence type:** test execution logs (repeated-run reproduction protocol).
- **Story / task:** W01-E04-S003 / W01-E04-S003-T001.
- **Acceptance criteria proven:** AC-W01-E04-S003-01.
- **Execution commands:**
  - `go test -v -count=1 -timeout=15m ./internal/e2e/` (run-00, main tree).
  - `go test -v -count=5 -parallel=4 -timeout=25m ./internal/e2e/` ×4 (run-01..04, main tree;
    run-05..08, isolated worktree).
  - Stress ×3: `go test -v -count=2 -timeout=25m ./internal/e2e/` concurrent with
    `go test -count=1 -timeout=25m ./testkit/ ./internal/cli/` (worktree).
  - `go test -v -race -count=2 -timeout=25m ./internal/e2e/` (run-09, worktree).
  - All runs with `DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable`
    and `WOWAPI_REQUIRE_DB=1` (skip-masking guard).
- **Code revision:** `0a31186cada5c275a588c74081cf977adf346e61` — runs 05-09 and stress executed
  in a detached `git worktree` pinned at exactly this SHA (created for isolation, removed after).
  Runs 00-04 ran in the main working tree, whose production files were being concurrently
  mutated by sibling wave workers (see contamination note below); the story's own change set
  contains **zero** production-file modifications (`git status` for `internal/e2e/`, `testkit/`:
  clean; working-tree delta is governance/evidence files under this story dir only).
- **Branch or tag:** `main` (detached worktree at the same SHA for the isolated runs).
- **Execution environment:** local darwin/arm64 workstation; `make up` compose stack
  (`wowapi-postgres-1`, `wowapi-minio-1`, both healthy).
- **Tool versions:** go1.26.5 darwin/arm64; PostgreSQL 16.14; git worktree isolation.
- **Date/time:** 2026-07-13, 12:50–13:15 IST (per-run timestamps embedded in the logs).
- **Result:** 29/29 PASS under clean conditions (1 preflight + 20 isolated primary + 6 stress +
  2 race); 16 FAIL in run-01..04 from a fully-identified external cause (sibling worker's
  in-flight edit to `adapters/tracing/otel`/`kernel/observability`, compiled into the run via
  the product `replace` directive) — preserved, not the historical flake. Historical failure:
  **not reproduced** (bounded refutation).
- **File/URI:** `logs/` (16 files, ~328K) alongside this record; diagnosis in
  `diagnosis-note.md` (EV-W01-E04-S003-002 / ART-W01-E04-S003-002).
- **Checksum:** SHA-256 over the sorted per-file SHA-256 list of `logs/*`:
  `bc04020519f2bc02ac8a2d133822505e452d83b9bb8160819b0f14b4a2bb1f23`.
- **Reviewer:** framework architecture lead (pending story acceptance; conductor gate).
- **Superseded evidence:** run-01..04 (status `failed`, environment contamination) superseded by
  run-05..08 (status `retested`) at the pinned SHA. No prior T-TEST-01 evidence records exist.

Failed-evidence preservation: the 16 failing executions in run-01..04 are retained verbatim in
`logs/run-0{1..4}-count5-parallel4.log` per mandate §10 ("do not delete earlier failed
verification merely because a later run passes").

## Reviewer completion addendum — 2026-07-16

**Reviewer**: Independent review agent (Claude Sonnet 4.5), dispatched 2026-07-16 by Fable 5 conductor (autopsy remediation R-3).
**Review date**: 2026-07-16.
**Commit revision reviewed against**: HEAD 43b6e12 + remediation working tree 2026-07-16.
**Disposition**: Spot-checked, not re-run. The flaky e2e suite itself was not re-executed in this pass (diagnosis/decision record, not a simple repro command, consistent with the autopsy's own scoping). Evidence bundle (diagnosis-note.md, reproduction-runs.md, logs/) confirmed present on disk.

This addendum retroactively fills the evidence-policy-mandated "reviewer" field. The original
record above (including the "framework architecture lead (pending story acceptance; conductor
gate)" reviewer line) is left unmodified per the failed-evidence preservation convention — this
is an appended addendum, not a rewrite.
