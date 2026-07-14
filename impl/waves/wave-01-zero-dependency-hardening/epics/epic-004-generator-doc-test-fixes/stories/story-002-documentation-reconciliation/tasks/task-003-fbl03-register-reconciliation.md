---
id: W01-E04-S002-T003
type: task
title: FBL-03 wowsociety upstream register reconciliation
status: done
parent_story: W01-E04-S002
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on:
  - W01-E04-S001
acceptance_criteria:
  - AC-W01-E04-S002-03
artifacts:
  - ART-W01-E04-S002-005
evidence:
  - EV-W01-E04-S002-003
---

# W01-E04-S002-T003 — FBL-03 wowsociety upstream register reconciliation

## Task Definition

### Task objective

Produce a precise, actionable PROD-level coordination recommendation for reconciling the wowsociety
repository's upstream finding register — marking PF-2 closeable once sibling story W01-E04-S001's DX-02
fix lands, and correcting PF-6/RFF-001 to already-resolved status per REVIEW Answer 18.

### Parent story

W01-E04-S002 — Documentation reconciliation.

### Owner

Unassigned.

### Status

`done`.

### Dependencies

`W01-E04-S001` — specifically, this task's PF-2 sub-item cannot be finalized as "closeable" until
S001's DX-02 task (the generator permission-verb fix) has actually landed. The PF-6/RFF-001 sub-items
have no such dependency and can be finalized independently.

### Detailed work

**This task does not edit the wowsociety repository.** The wowsociety upstream register
(`docs/upstream/` in the `wowsociety` repository) is a file this repository does not own. Per mandate
§2.3's framework/product boundary discipline, this task's deliverable is a **PROD-level coordination
recommendation** — a precisely-worded document, analogous in shape to `requirement-inventory.md` §D's
PROD-01 through PROD-05 rows, that a wowsociety-repository maintainer (or a future cross-repository
task) can act on directly.

The recommendation must state, for each of the three named findings:

1. **PF-2** — the wowsociety-documented instance of DX-02's `.delete`-verb generator defect (see
   `wowsociety/docs/upstream/12-sf-7-init-gomod-invalid-and-gitignored-local-overlay.md`, cited
   informationally by DX-01's own row note, though PF-2 itself is the DX-02-verb-defect finding, not the
   init-gomod finding — confirm the exact PF-2 document reference at implementation time rather than
   assuming it is the same document as the DX-01 citation). Recommend marking PF-2 **closed** only once
   S001's DX-02 task has landed — state this contingency explicitly in the recommendation text, do not
   recommend closing it prematurely.
2. **PF-6** — per REVIEW Answer 18 ("no active workarounds remain... mark the 2 stale upstream docs
   resolved"), recommend correcting PF-6's register entry to already-resolved status. This task does
   not re-derive why PF-6 is resolved; it acts on the epic's own governing citation of REVIEW Answer 18.
3. **RFF-001** — the second of REVIEW Answer 18's "2 stale upstream docs," recommend the same
   already-resolved correction.

The recommendation document should be precise enough that whoever applies it in the wowsociety
repository does not need to re-derive the reasoning — it should name the exact register file (to the
extent knowable without direct wowsociety-repository access at planning time — flag as TBD if not
confirmable) and the exact status change requested for each of the three entries.

### Expected files or components affected

No `wowapi` production files. No `wowsociety` files (out of this repository's reach). This task's own
output is the recommendation document, stored within this story's `artifacts/` tree.

### Expected output

A single coordination-recommendation document naming PF-2 (contingent on S001), PF-6, and RFF-001, each
with a precise target status and rationale.

### Required artifacts

The coordination-recommendation document, registered in `../artifacts/index.md`.

### Required evidence

The recommendation document itself, registered in `../evidence/index.md`, functions as both artifact
and evidence — there is no separate "test" for a cross-repository coordination note beyond reviewer
confirmation of its precision and accuracy.

### Related acceptance criteria

AC-W01-E04-S002-03.

### Completion criteria

The recommendation document exists, names all three findings with correct target statuses, and
explicitly states PF-2's contingency on S001 rather than recommending premature closure.

### Verification method

Reviewer confirms the recommendation's phrasing precisely reflects REVIEW Answer 18 for PF-6/RFF-001,
and that PF-2's contingency on S001 is stated (not omitted or softened into an unconditional recommendation).

### Risks

RISK-W01-E04-003 (this task cannot verify the wowsociety-side edit is ever actually applied, since it
lives outside this repository's control) — see `../../../risks.md` (epic level). Mitigation: this
task's completion criteria are scoped to producing a correct recommendation, not to confirming a
downstream edit — this keeps the task's own closure independent of an uncontrollable external factor.

### Rollback or recovery considerations

Not applicable — this task produces a recommendation document, not a live system change.

## Implementation Record

### What was actually implemented

PROD-level coordination recommendation produced:
`../artifacts/fbl03-wowsociety-register-coordination-recommendation.md`. It names the exact
register files and README index rows (from read-only inspection of `wowsociety/docs/upstream/`):
PF-2 = `06-pf-2-gen-crud-emits-out-of-set-verb.md` — recommended RESOLVED **only after** S001's
DX-02 lands (fix confirmed in-flight-not-landed via IRC with the S001 owner; contingency stated,
not softened); PF-6 = `01-pf-6-step-up-seedability.md` — entry body already carries a RESOLVED
header (`d2a4164`), so the recommendation reconciles the still-open README index row and the
"PF-6 is prioritized" posting instruction with it; RFF-001 =
`03-rff-001-production-object-storage-adapter.md` — recommended RESOLVED, corroborated by
`adapters/storage/s3` present at HEAD. The inventory note's "etc." resolved conservatively: no
recommendation for findings whose fixes are unverified. No wowsociety file was edited.

### Components changed

None — recommendation document only, inside this story's `artifacts/` tree.

### Files changed

`../artifacts/fbl03-wowsociety-register-coordination-recommendation.md` (new).

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

None — reviewer confirmation is the verification method for a coordination note.

### Commits

None (conductor commits at wave close).

### Pull requests

None.

### Implementation dates

2026-07-13.

### Technical debt introduced

None.

### Known limitations

Cannot verify the wowsociety-side edit is ever applied (RISK-W01-E04-003) — restated here as the
planned known limitation, tracked at programme level if never applied.

### Follow-up items

wowsociety-side application of the recommendation; PF-2's row only after S001's DX-02 lands and
the shipping commit SHA is known (the recommendation leaves `<commit>` as the one deliberate
placeholder for that reason).

### Relationship to the approved plan

Matches plan. The story-level dependency on S001 was honored by the contingency framing rather
than by waiting: the plan's approval condition (b) permits proceeding with the dependency
"explicitly still-pending with that status recorded," which the recommendation does.

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W01-E04-S002-03 | Review recommendation against REVIEW Answer 18 and S001's completion state | Local doc review | PF-2 contingency stated; PF-6/RFF-001 correctly targeted as already-resolved | Doc diff / coordination note | developer-experience lead |

### Actual result

Recommendation names all three findings with exact files, index rows, target statuses, and
rationale; PF-2 contingency explicit; PF-6/RFF-001 already-resolved per REVIEW Answer 18 with
independent corroboration for each.

### Pass or fail

PASS.

### Evidence identifier

EV-W01-E04-S002-003.

### Execution date

2026-07-13.

### Commit or revision

`0a31186cada5c275a588c74081cf977adf346e61` (carry-forward to `05dce5c8` in `../evidence/index.md`);
wowsociety register inspected read-only the same day.

### Environment

Local doc review, macOS.

### Reviewer

Developer-experience lead (role); conductor acceptance at wave close.

### Findings

None.

### Retest status

Not required.

### Final conclusion

AC-W01-E04-S002-03 satisfied.

## Deviations Record

None specific to this task (story-level deviations in `../deviations.md`).
