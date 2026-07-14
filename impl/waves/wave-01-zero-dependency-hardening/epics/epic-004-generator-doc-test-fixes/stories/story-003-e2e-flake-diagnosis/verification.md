---
id: VER-W01-E04-S003
type: verification-record
parent_story: W01-E04-S003
status: final
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Verification record — W01-E04-S003

## Planned verification procedure

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W01-E04-S003-01 | Execute the reproduction protocol (see `plan.md`); record pass/fail per run and any failure logs; separately determine and record `internal/e2e`'s actual DB-wiring mechanism | CI or local, with a database available | EITHER a reproduced failure with root-cause evidence OR a documented non-reproduction after the planned run budget; DB-wiring determination recorded either way | Test execution log + diagnosis note | framework architecture lead |
| AC-W01-E04-S003-02 | Review T002's implemented outcome against T001's actual findings and the illustrative decision branches in `tasks/task-002-conditional-fix.md` | Local or CI, matching whatever T002's branch requires | **Conditional — cannot be stated precisely until T001 completes.** The outcome is expected to be one of: a code/test fix matching a specific reproduced root cause; a documented monitoring-only outcome with no code change, if T001 found no reproduction; or a fix targeting an unforeseen branch T001 surfaces. Whichever it is, "expected result" is that T002's outcome is *consistent with and traceable to* T001's actual recorded findings, not invented independently of them | Depends on branch: test execution log (if code change), or diagnosis-note update (if monitoring-only) | framework architecture lead |

This table's AC-W01-E04-S003-02 row is deliberately incomplete in its "expected result" column — mandate
§8.5's investigation-plan guidance and this epic's own governing instruction both require that the
conditionality be stated explicitly rather than silently resolved by inventing a fixed expected result
in advance.

## Post-execution record

Executed 2026-07-13. Per-AC results:

| Acceptance criterion | Actual result | Pass/fail | Evidence |
|---|---|---|---|
| AC-W01-E04-S003-01 | Protocol executed at pinned SHA `0a31186cada5c275a588c74081cf977adf346e61`: `-count=5 -parallel=4` ×4 invocations (20 executions, isolated worktree) + 3 shared-DB stress iterations (6 executions) + `-race -count=2` (2) + preflight (1) = **29/29 PASS — documented non-reproduction**. DB-wiring determination recorded: `internal/e2e` uses its own wiring (raw `DATABASE_URL` → base DB; product migrate/api directly against it), NOT `testkit.NewDB`; single test, no `t.Parallel` (so `-parallel` is inert — recorded as a protocol fact). Withdrawn "shared-DB concurrency" cause NOT re-asserted: deliberate concurrent template-clone/DDL stress on the shared base DB produced zero failures. 16 contaminated main-tree failures preserved as `failed` evidence with cause fully identified (sibling-worker in-flight edits), superseded by `retested` runs at the pinned SHA. | **PASS** | EV-W01-E04-S003-001, EV-W01-E04-S003-002 |
| AC-W01-E04-S003-02 | T002 resolved to the monitoring-only branch (task-002 illustrative branch 3), strictly derived from T001's non-reproduction finding; the actual branch taken is recorded against the illustrative decision branches in task-002's Implementation Record and diagnosis-note §5; monitoring protocol defined in §6. No code change — consistent with T001's findings by construction. | **PASS** | EV-W01-E04-S003-003 |

### Actual result

Both ACs satisfied — see table above.

### Pass or fail

PASS (2/2 ACs).

### Evidence identifier

EV-W01-E04-S003-001..003, at `evidence/premier/T-TEST-01/` (see `evidence/index.md`).

### Execution date

2026-07-13.

### Commit or revision

`0a31186cada5c275a588c74081cf977adf346e61` — reproduction runs 05-09 + stress executed in a
detached worktree pinned at exactly this SHA. The repo SHA is unchanged by this story
(investigation + governance records only; conductor owns commits). Working-tree delta
attributable to this story: files under this story directory only; `git status` for
`internal/e2e/` and `testkit/` is clean.

### Environment

Local darwin/arm64; go1.26.5; PostgreSQL 16.14 + MinIO (`make up` compose stack, both healthy);
`WOWAPI_REQUIRE_DB=1` on every run.

### Reviewer

Framework architecture lead role — review executed by W01ReviewGate; accepted by conductor 2026-07-13 (story front matter reviewer
remains role-based).

### Findings

Two facts beyond the ACs, recorded in the diagnosis note: (1) demonstrated failure-domain
insight — the suite compiles the whole framework tree at run time via the `replace` directive,
so any transient tree inconsistency fails it in a full-suite-only pattern (16 real instances
produced during this investigation); (2) the original failure's log was never preserved, which
is why its cause is permanently unassignable — the §6 monitoring protocol closes that gap.

### Retest status

Runs 01-04 (`failed`, environment contamination) retested as runs 05-08 (20/20 PASS,
`retested`) at the pinned SHA; both preserved per mandate §10.

### Final conclusion

Story verified: honest, evidence-backed bounded non-reproduction; DB-wiring determination
recorded; monitoring-only outcome implemented consistently with the findings. Ready for
architecture-lead acceptance.

### Revision carry-forward note (evidence-policy §revision-pinning, option 2)

After this story's evidence was captured at `0a31186cada5c275a588c74081cf977adf346e61`, HEAD
advanced to `05dce5c` (conductor commit "impl: Wave 00 baseline-and-verification executed and
accepted"). What changed in between: governance documents under `impl/` only —
`git diff --stat 0a31186..05dce5c -- internal/e2e testkit migrations` is empty, and no
production Go file changed at all. The evidence is explicitly carried forward: nothing material
to either AC changed between the pinned SHA and current HEAD, so re-running the reproduction
protocol against `05dce5c` would exercise byte-identical code. Any FUTURE commit that touches
`internal/e2e/`, `testkit/`, or the migration set before story acceptance re-opens this note.

