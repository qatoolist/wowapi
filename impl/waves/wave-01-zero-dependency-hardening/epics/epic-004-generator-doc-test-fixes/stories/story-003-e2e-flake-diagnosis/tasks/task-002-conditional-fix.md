---
id: W01-E04-S003-T002
type: task
title: Conditional fix, per T001's findings
status: done
parent_story: W01-E04-S003
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on:
  - W01-E04-S003-T001
acceptance_criteria:
  - AC-W01-E04-S003-02
artifacts:
  - ART-W01-E04-S003-003
evidence:
  - EV-W01-E04-S003-003
---

# W01-E04-S003-T002 — Conditional fix, per T001's findings

## Task Definition

### Task objective

Implement whatever fix (or explicit no-fix, monitoring-only outcome) T001's investigation determines is
warranted — strictly derived from T001's actual findings, not invented in advance of them.

### Parent story

W01-E04-S003 — E2E flake diagnosis.

### Owner

Unassigned.

### Status

`done`.

### Dependencies

`W01-E04-S003-T001` — hard, blocking dependency. This task cannot be meaningfully scoped, let alone
started, until T001 has completed both its reproduction protocol and its DB-wiring determination.

### Detailed work

**This task's detailed work cannot be fully specified in advance.** Per mandate §8.5: "where
implementation details cannot yet be known, state what must be determined during the story rather than
inventing specifics." What follows is the currently-foreseeable decision space this task will resolve
into once T001 completes — stated as illustrative branches, not a commitment to any one of them, and not
exhaustive of every possibility T001's investigation might actually surface.

**Illustrative decision branches:**

- **If T001 finds `internal/e2e` bypasses `testkit.NewDB` and has its own, non-isolated (or
  differently-isolated) DB wiring**, THEN this task's fix is expected to route `internal/e2e` through
  `testkit.NewDB` or an equivalent isolation mechanism, consistent with how other test packages in the
  repository already use it.
- **If T001 finds a genuine race condition or defect in `testkit`'s own cloning/cleanup mechanism**
  (not currently suspected, but not ruled out), THEN this task's fix targets that mechanism directly,
  inside `testkit/db.go`.
- **If T001 cannot reproduce the failure after the planned investigation budget**, THEN this task's
  outcome is explicitly NOT a code fix — it is a documented, monitored, non-blocking known-flaky-item
  record (or, more precisely, a record that a single historical failure could not be reproduced and is
  downgraded to a monitoring item, since "known-flaky" would itself overstate confidence about
  recurrence). This is a legitimate, complete outcome for this task, not a fallback of last resort.
- **If T001 surfaces a cause unrelated to database isolation entirely** (e.g., a resource-exhaustion or
  timing issue orthogonal to `testkit`), THEN this task's fix targets whatever that actual cause is —
  this branch is explicitly left open-ended because T001's plan does not pre-suppose the reproduced
  failure (if any) will be DB-related at all.

These branches are recorded here as the currently-foreseeable decision space only. T001's actual
findings govern which branch (or an unforeseen one not listed above) is taken — this task's eventual
Implementation Record section will state explicitly which branch was actually followed and why.

### Expected files or components affected

Not determinable until T001 completes and a branch is selected. Candidates depending on branch:
`internal/e2e/`'s test-setup code, `testkit/db.go`, or no code file at all (monitoring-only branch).

### Expected output

Either a code/test change addressing T001's confirmed root cause, or a documented monitoring-only
outcome with no code change — whichever T001's findings warrant.

### Required artifacts

Whatever artifact the selected branch produces (a code diff, or a monitoring-decision note),
registered in `../artifacts/index.md` once known.

### Required evidence

Test execution evidence (if a code change results) or a diagnosis-note update (if monitoring-only),
registered in `../evidence/index.md` once known.

### Related acceptance criteria

AC-W01-E04-S003-02.

### Completion criteria

The implemented outcome (code fix or documented monitoring-only decision) is traceable to and
consistent with T001's actual recorded findings — not invented independently of them.

### Verification method

Review by the framework architecture lead confirming the selected branch's outcome matches what T001
actually found, per `../verification.md`'s deliberately-conditional AC-W01-E04-S003-02 row.

### Risks

The primary risk is scope drift if this task's implementer is tempted to implement a fix that "seems
reasonable" rather than one strictly derived from T001's actual findings — mitigated by this task's own
completion criteria requiring explicit traceability to T001's decision record.

### Rollback or recovery considerations

Not determinable until the selected branch (and any resulting code change) is known.

## Implementation Record

### What was actually implemented

**Branch taken: illustrative branch 3 — monitoring-only, no code fix.** T001's investigation
could not reproduce the historical failure after the planned budget (29/29 clean executions at
`0a31186cada5c275a588c74081cf977adf346e61` under `-count=5 -parallel=4` ×4 invocations,
3 shared-DB stress iterations, and `-race -count=2`). Per this task's own definition, the
outcome is a documented monitoring record, not a code change: the single historical failure is
downgraded to a programme-level monitoring item, defined in
`../evidence/premier/T-TEST-01/diagnosis-note.md` §5 (decision) and §6 (monitoring protocol:
preserve the failure log before any rerun; classify the failing step — tree-compilation vs
runtime — before attributing cause).

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

None — the monitoring-only branch makes no code or test change by definition.

### Commits

None yet.

### Pull requests

None yet.

### Implementation dates

2026-07-13.

### Technical debt introduced

None yet.

### Known limitations

None yet.

### Follow-up items

The programme-level monitoring item (diagnosis-note §6). No code follow-up.

### Relationship to the approved plan

Branch 3 of the illustrative decision space ("cannot reproduce after the planned investigation
budget → documented, monitored, non-blocking record") was taken, exactly as foreseen, strictly
traceable to T001's decision record (task-001 Implementation Record + diagnosis-note §5). The
plan's deferred-approval condition for T002 is satisfied trivially: no code change was
implemented, so nothing beyond the documented monitoring outcome required approval; the
architecture-lead review at story acceptance covers this record.

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W01-E04-S003-02 | Review implemented outcome against T001's actual findings | Depends on selected branch | Conditional — see `../verification.md` | Depends on branch | framework architecture lead |

### Actual result

Monitoring-only outcome implemented: decision + monitoring item recorded in diagnosis-note
§5-§6; zero production files changed (verified: `git status` clean for `internal/e2e/`,
`testkit/`).

### Pass or fail

PASS (AC-W01-E04-S003-02: outcome traceable to and consistent with T001's actual findings;
actual branch recorded against the illustrative branches).

### Evidence identifier

EV-W01-E04-S003-003.

### Execution date

2026-07-13.

### Commit or revision

`0a31186cada5c275a588c74081cf977adf346e61` (unchanged by this story; conductor owns commits).

### Environment

Not applicable beyond T001's (no code executed; documentation-only outcome).

### Reviewer

Framework architecture lead (pending — conductor acceptance gate).

### Findings

None beyond T001's — this task consumed T001's decision record without modification.

### Retest status

Not applicable (no code change to retest).

### Final conclusion

AC-W01-E04-S003-02 satisfied; story-level outcome complete (both tasks done).

## Deviations Record

No deviations recorded yet. Note: because this task's plan is explicitly conditional on T001, selecting
a branch not among the illustrative options above is not automatically a "deviation" in the traditional
sense — see `../deviations.md` for how this story treats that distinction.
