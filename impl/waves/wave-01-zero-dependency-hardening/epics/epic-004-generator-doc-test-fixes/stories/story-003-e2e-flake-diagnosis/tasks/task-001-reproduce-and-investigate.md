---
id: W01-E04-S003-T001
type: task
title: Reproduce and investigate (decision task)
status: done
parent_story: W01-E04-S003
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on: []
acceptance_criteria:
  - AC-W01-E04-S003-01
artifacts:
  - ART-W01-E04-S003-001
  - ART-W01-E04-S003-002
evidence:
  - EV-W01-E04-S003-001
  - EV-W01-E04-S003-002
---

# W01-E04-S003-T001 — Reproduce and investigate (decision task)

## Task Definition

### Task objective

Execute T-TEST-01's re-scoped 3-step protocol's first two steps — reproduce the intermittent
`internal/e2e` full-suite failure under `-count`+parallel runs, and determine whether `internal/e2e`
actually uses `testkit.NewDB` cloning or its own DB wiring — and produce a decision record stating what
(if anything) the third step's fix, owned by sibling task T002, should be.

### Parent story

W01-E04-S003 — E2E flake diagnosis.

### Owner

Unassigned.

### Status

`done`.

### Dependencies

None.

### Detailed work

This task implements exactly steps 1–2 of T-TEST-01's re-scoped protocol. Step 3 ("fix what the
reproduction shows") is explicitly NOT part of this task — it belongs to sibling task T002.

1. **Reproduce.** Run `internal/e2e`'s full test suite via `go test -count=N -parallel=P
   ./internal/e2e/...`, with N (repeat count) and P (parallelism) chosen at implementation time within
   a reasonable CI/local time budget. Repeat across multiple invocations if the first does not surface a
   failure, up to the planned budget — do not treat a single clean run as conclusive either way.
2. **Investigate DB wiring.** Independently of the reproduction outcome, read `internal/e2e`'s test-
   setup/fixture code directly to determine whether it calls `testkit.NewDB` (the per-test database
   cloning mechanism confirmed to exist at `testkit/db.go:83-144,313`) or wires its own, separate
   database setup. Record this determination as a plain fact, not an inference from the reproduction's
   pass/fail outcome.
3. **Synthesize.** Combine both findings into a diagnosis. This task's OUTPUT is a decision about what
   T002 should do — not the fix itself. The decision is recorded in this task's own Implementation
   Record section (below) once executed, functioning as a task-level decision record (per this epic's
   own framing: an investigation story's decision output belongs inside the investigating task, not as
   a separate story-level ADR, since this is not a programme-level architecture decision).

**Explicit non-goal:** this task does not pre-commit to a cause before the reproduction is attempted.
The withdrawn "shared-DB concurrency" attribution is not to be re-asserted as this task's conclusion
unless the investigation actually produces new evidence specifically supporting it.

### Expected files or components affected

No production code files. `internal/e2e`'s test-setup code is read, not modified, by this task.

### Expected output

A reproduction-run log collection (every run's pass/fail outcome and full logs) and a diagnosis/decision
record stating: (a) whether the failure reproduced, with evidence if so; (b) whether `internal/e2e` uses
`testkit.NewDB` or its own DB wiring; (c) what T002 should do as a result.

### Required artifacts

The reproduction-run log collection and the diagnosis/decision record, registered in
`../artifacts/index.md`.

### Required evidence

The reproduction-run logs, at path `evidence/premier/T-TEST-01/`, registered in `../evidence/index.md`.

### Related acceptance criteria

AC-W01-E04-S003-01.

### Completion criteria

The reproduction protocol has actually been executed (not skipped or assumed); its outcome is recorded
with evidence; the DB-wiring determination is recorded; a decision for T002 is stated, whether that
decision is "implement fix X," "no code fix needed, downgrade to monitoring," or another outcome the
investigation surfaces.

### Verification method

Review of the reproduction-run logs and diagnosis record by the framework architecture lead, confirming
the protocol was genuinely executed (not merely asserted) and that the diagnosis does not re-assert the
withdrawn cause without new supporting evidence.

### Risks

RISK-W01-004 (the reproduction step fails to reproduce the failure at all) — see
`../../../risks.md` (epic level). This is an accepted, valid possible outcome for this task, not a
failure of the task itself, provided the investigation was genuinely attempted at the planned budget.

### Rollback or recovery considerations

Not applicable — this task performs no code or configuration change; only test execution and reading.

## Implementation Record

### What was actually implemented

Both protocol steps executed 2026-07-13 against pinned commit
`0a31186cada5c275a588c74081cf977adf346e61`:

1. **Reproduction:** budget fixed at implementation time as 4 invocations of
   `go test -count=5 -parallel=4 ./internal/e2e/` (20 executions) + 3 stress iterations
   (`-count=2` e2e concurrent with `go test ./testkit/ ./internal/cli/` on the same base DB,
   6 executions) + `-race -count=2` (2 executions) + 1 preflight — all with
   `WOWAPI_REQUIRE_DB=1` so skips could not masquerade as passes. First 4 invocations ran in
   the main working tree and were contaminated by a sibling wave worker's in-flight edits
   (16 failures at the `go vet` step, cause fully identified — see `../deviations.md`); the
   protocol was re-executed in a detached git worktree pinned at the SHA above: **29/29 PASS**
   under clean conditions. The historical failure did NOT reproduce.
2. **DB-wiring determination (direct code reading):** `internal/e2e` does NOT use
   `testkit.NewDB` — single file `e2e_test.go`, no testkit import, no `t.Parallel()`; it
   consumes raw `DATABASE_URL` (line 113), the scaffolded product's migrate applies kernel
   migrations directly to the base database (lines 127-137), and the api binary connects to
   that same base DB (lines 207-210). Own wiring, no per-test clone.

Full diagnosis: `../evidence/premier/T-TEST-01/diagnosis-note.md`.

### Components changed

None yet.

### Files changed

None yet.

### Interfaces introduced or changed

None yet.

### Configuration changes

None yet.

### Schema or migration changes

None yet.

### Security changes

None yet.

### Observability changes

None yet.

### Tests added or modified

None — this task executed existing tests repeatedly; no test code was added or modified.

### Commits

None yet.

### Pull requests

None yet.

### Implementation dates

2026-07-13 (single day).

### Technical debt introduced

None yet.

### Known limitations

None yet.

### Follow-up items

None yet — this task's own output (the decision record) becomes T002's direct input once produced.

### Relationship to the approved plan

Followed as approved, with one environment adaptation recorded in `../deviations.md`: the
`-count`+parallel runs were re-executed in an isolated worktree pinned at HEAD because the
shared working tree was being concurrently mutated by sibling W01 workers (their in-flight
edits compile into this suite via the product `replace` directive), which contaminated the
first 4 invocations. The plan's protocol itself was unchanged.

**Decision record (populated on execution):**

- Reproduction outcome: **not reproduced** — 29/29 clean executions at
  `0a31186cada5c275a588c74081cf977adf346e61` under `-count=5 -parallel=4` ×4, stress ×3, and
  `-race -count=2`. (16 additional failures in the contaminated main-tree invocations are
  fully explained by sibling-worker tree mutation, preserved as `failed` evidence, and are not
  the historical flake.)
- DB-wiring determination: `internal/e2e` uses **its own wiring** (raw `DATABASE_URL` → base
  database; product migrate + api against it directly), NOT `testkit.NewDB` cloning. The
  withdrawn "shared-DB concurrency" cause is NOT re-asserted: despite the bypass being real,
  deliberate concurrent shared-DB stress produced zero failures.
- Decision for T002: **monitoring-only, no code fix** (task-002 illustrative branch 3), with
  the programme-level monitoring item defined in `diagnosis-note.md` §6 (preserve the failure
  log before rerunning; classify tree-compilation vs runtime step first).

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W01-E04-S003-01 | Execute reproduction protocol; record pass/fail per run; determine DB-wiring mechanism | CI or local, DB available | Reproduced-with-evidence OR documented non-reproduction; DB-wiring determination recorded | Test execution log + diagnosis note | framework architecture lead |

### Actual result

Documented non-reproduction (29/29 clean at pinned SHA); DB-wiring determination recorded
(own wiring, not `testkit.NewDB`); decision for T002 recorded (monitoring-only).

### Pass or fail

PASS (AC-W01-E04-S003-01 satisfied: protocol executed, result recorded, determination made,
withdrawn cause not re-asserted).

### Evidence identifier

EV-W01-E04-S003-001, EV-W01-E04-S003-002.

### Execution date

2026-07-13.

### Commit or revision

`0a31186cada5c275a588c74081cf977adf346e61` (isolated worktree pinned at this SHA for runs
05-09 + stress; conductor owns commits, so the repo SHA is unchanged by this story).

### Environment

Local darwin/arm64; go1.26.5; PostgreSQL 16.14 + MinIO via `make up` compose stack.

### Reviewer

Framework architecture lead (pending — conductor acceptance gate).

### Findings

See decision record above and `../evidence/premier/T-TEST-01/diagnosis-note.md` (§3 for the
contaminated-run analysis, §4 for the demonstrated failure-domain insight).

### Retest status

Runs 01-04 (`failed`, contamination) retested as runs 05-08 (`retested`, 20/20 PASS) at the
pinned SHA — both preserved per mandate §10.

### Final conclusion

AC-W01-E04-S003-01 satisfied; T001 complete; decision handed to T002 (monitoring-only branch).

## Deviations Record

One environment adaptation (isolated-worktree re-execution after sibling-worker tree
contamination) — recorded in `../deviations.md`; the protocol itself was executed as planned.
