---
id: W01-E04-S002-T002
type: task
title: DX-05 residual reconciliation — T3 blueprint-11 examples, T4 version-gate design, T5 deferral
status: done
parent_story: W01-E04-S002
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on: []
acceptance_criteria:
  - AC-W01-E04-S002-02
artifacts:
  - ART-W01-E04-S002-002
  - ART-W01-E04-S002-003
  - ART-W01-E04-S002-004
evidence:
  - EV-W01-E04-S002-002
---

# W01-E04-S002-T002 — DX-05 residual reconciliation

## Task Definition

### Task objective

Close DX-05's three residual sub-tasks: reconcile blueprint-11's CLI examples against
`internal/cli/cli.go`'s real commands/flags (T3), design a version-compatibility gate for mutating
generator commands sharing S001's version-verification plumbing (T4), and record T5 as an explicit
deferral to W06/REL-03 (T5).

### Parent story

W01-E04-S002 — Documentation reconciliation.

### Owner

Unassigned.

### Status

`done`.

### Dependencies

Soft, task-level, plumbing dependency: this task's T4 sub-item should not begin *implementation* before
sibling story `W01-E04-S001`'s T001 (DX-01 version-resolution flags/plumbing) has landed, since T4
reuses that plumbing's version-comparison logic. This does not block T3's or T5's work, and does not
block this task's own *planning*.

### Detailed work

**T3 — blueprint-11 CLI example reconciliation.** Walk every CLI example documented in blueprint-11
(exact document path to be confirmed at implementation time — not independently verified by this
planning task) against `internal/cli/cli.go`'s actual, current commands and flags. For each example,
record one of two decisions: (a) **implement** — the example is still valid/intended but the
documentation is stale relative to what the CLI actually does, so the documentation should be corrected
to match; or (b) **delete** — the example describes a command/flag surface that no longer exists or was
never implemented, so the example should be removed rather than corrected. This is a per-example
judgment call, not a mechanical fix — `plan.md`'s "unresolved questions" section flags that the volume
of stale examples is not known at planning time (see RISK-W01-E04-002).

**T4 — version-compatibility gate design.** `requirement-inventory.md`'s DX-05 row targets a design
where `wapi version` (or, more precisely per this epic's framing, the mutating generator commands
themselves) fails when the resolved framework version and the target module's declared version
constraint have an incompatible major/minor pairing. This task's deliverable is the **design**, not the
implementation: describe the check's trigger point (before generation proceeds, mirroring DX-01's
fail-closed-before-any-file-write discipline), the comparison logic it needs (major/minor
compatibility, not exact-version equality), and its explicit reuse of S001's DX-01 version-verification
plumbing (the `go list -m`-based resolution check) rather than a parallel, independently-built
mechanism.

**T5 — explicit deferral.** Record, in this task's own output (and in the story's `story.md`
out-of-scope section), that DX-05 T5 (public API/config/event compatibility gates enforcing v1 rules)
is deferred to W06, citing that it is explicitly "shared with REL-03" per the plan document and that
REL-03 targets `W06-E02-S002..S003` per `requirement-inventory.md`. This is a recorded deferral, not a
silent drop — mandate §11.10's deferred-items-register discipline applies at programme level; this
task's output feeds that register.

### Expected files or components affected

Blueprint-11's CLI-examples document (path TBD at implementation time) for T3. No production code files
for T4 (design note only, not implementation). No files for T5 beyond this task's own recorded deferral
note.

### Expected output

A per-example decision table for T3 (implement/delete, one row per blueprint-11 CLI example); a design
note for T4 describing the version-gate's trigger point, comparison logic, and S001-plumbing reuse; a
deferral note for T5 citing the REL-03/W06 cross-reference.

### Required artifacts

T3's decision table, T4's design note, and T5's deferral note, registered in `../artifacts/index.md`.

### Required evidence

The decision table and design note themselves serve as evidence of completeness (every example
accounted for; the design explicitly states the S001 dependency), registered in `../evidence/index.md`.

### Related acceptance criteria

AC-W01-E04-S002-02.

### Completion criteria

Every blueprint-11 CLI example has a recorded implement-or-delete decision; the T4 design note exists
and explicitly states its dependency on S001's plumbing; the T5 deferral note exists and cites the
REL-03/W06 cross-reference.

### Verification method

Review of the decision table for completeness (no example left undecided); review of the design note
for the explicit S001-dependency statement; review of the deferral note for the correct cross-reference.

### Risks

RISK-W01-E04-002 (blueprint-11 drift could make T3's scope larger than estimated) — see `../../../risks.md`
(epic level). Mitigation: re-confirm `internal/cli/cli.go`'s actual commands/flags fresh at
implementation time rather than trusting the blueprint's age; if the stale-example count is large
enough to threaten this task's boundedness, split into a follow-up rather than silently absorbing it.

### Rollback or recovery considerations

T3's documentation corrections can be reverted per-example if a delete decision is later found wrong.
T4's design note carries no runtime rollback concern (nothing is deployed by this task). T5's deferral
note is a recording action with no rollback need.

## Implementation Record

### What was actually implemented

**T3:** all 20 CLI examples/claims in `docs/blueprint/11-framework-distribution-and-consumption.md`
reconciled against `internal/cli/` — 14 keep, 6 implement-corrections applied (init flags,
new-module `--name`, gen-crud module directory, migrate-create `--dir`, seed-validate required
flags, openapi-merge `--check` removed, init config-seeding claim), 2 deletes applied (bare
`wowapi gen`, `wowapi config init`). Per-example record with code citations:
`../artifacts/dx05-t3-cli-example-decision-table.md`. Every stale form's failure was captured
before correction and every corrected form executed against binaries built from the working tree
AND a clean `git archive HEAD` extraction (`../evidence/reviews/ev-002-command-log.md`).
**T4:** design note produced (`../artifacts/dx05-t4-version-gate-design-note.md`) — trigger point
(mutating generator commands, fail-closed pre-write), major/minor comparison per the v1/N-1
policy, explicit reuse of S001's DX-01 plumbing; implementation deliberately NOT started (S001
owner confirmed over IRC the plumbing has not landed).
**T5:** deferral note produced (`../artifacts/dx05-t5-deferral-note.md`) citing REL-03 /
`W06-E02-S002..S003`; no deferred-items-register row per DEV-04 (ratified).

### Components changed

Blueprint-11 documentation only.

### Files changed

`docs/blueprint/11-framework-distribution-and-consumption.md` (diff:
`../evidence/reviews/ev-002-t002-blueprint11.diff`).

### Interfaces introduced or changed

None.

### Configuration changes

None.

### Schema or migration changes

None.

### Security changes

None.

### Observability changes

None.

### Tests added or modified

None (no code); verification by executing every documented command form against HEAD-built
binaries — fail-first captures for all stale forms.

### Commits

None (conductor commits at wave close).

### Pull requests

None.

### Implementation dates

2026-07-13.

### Technical debt introduced

None.

### Known limitations

Blueprint config examples parse but do not fully run on a pristine scaffold at HEAD due to two
pre-existing generator defects (DEV-03, reassigned to W01-E04-S001 by Main).

### Follow-up items

T4 implementation as a follow-on task once S001's DX-01 plumbing lands (the design note is the
contract); blueprint init example gains the version-pin flag in the change that ships DX-01's
flags (DEV-05).

### Relationship to the approved plan

Matches plan; deviations DEV-02/DEV-03/DEV-05 recorded in `../deviations.md`.

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W01-E04-S002-02 | Review decision table, design note, and deferral note for completeness and correctness | Local doc review | Every example decided; T4 dependency explicit; T5 deferral recorded | Doc diff / decision table | developer-experience lead |

### Actual result

Decision table complete (no example undecided); T4 note states the S001 dependency verbatim as a
load-bearing constraint; T5 note cites REL-03/W06.

### Pass or fail

PASS.

### Evidence identifier

EV-W01-E04-S002-002.

### Execution date

2026-07-13.

### Commit or revision

`0a31186cada5c275a588c74081cf977adf346e61`; HEAD-clean retest at `05dce5c8` (carry-forward note in
`../evidence/index.md`).

### Environment

Local macOS (darwin/arm64), go1.26.5.

### Reviewer

Developer-experience lead (role); conductor acceptance at wave close.

### Findings

Two out-of-scope generator defects surfaced and routed (DEV-03).

### Retest status

HEAD-clean retest performed; all results reproduced.

### Final conclusion

AC-W01-E04-S002-02 satisfied.

## Deviations Record

DEV-02, DEV-03, DEV-05 (see `../deviations.md`); DEV-03 ratified and reassigned by Main.
