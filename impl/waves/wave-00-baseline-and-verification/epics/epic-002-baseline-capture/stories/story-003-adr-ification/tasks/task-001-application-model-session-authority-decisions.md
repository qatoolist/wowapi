---
id: W00-E02-S003-T001
type: task
title: Application-model / session-authority decisions (D-01, D-02, D-03)
status: done
parent_story: W00-E02-S003
owner: W00-E02-S003 execution worker (agent)
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on: []
acceptance_criteria:
  - AC-W00-E02-S003-01
  - AC-W00-E02-S003-03
artifacts:
  - ART-W00-E02-S003-001
  - ART-W00-E02-S003-002
  - ART-W00-E02-S003-003
evidence: []
---

# W00-E02-S003-T001 — Application-model / session-authority decisions (D-01, D-02, D-03)

## Task Definition

Per mandate §8.6. This section defines the task before work begins.

### Task objective

Author three ADR files formalizing D-01 (framework owns grant validity/expiry/revocation
authority), D-02 (single generic owner-bound `Registrar` capability type with per-subsystem typed
keys), and D-03 (post-seal mutation errors in production, panics only under an explicit dev/test
build tag) — the three decisions in this story's "application-model / session-authority" cluster
(`plan.md` task-grouping rationale: all three concern the framework's core application-model and
session/capability-authority surface).

### Parent story

W00-E02-S003 — ADR-ification of D-01 through D-09.

### Owner

Unassigned.

### Status

`done` — implemented, verified, and output registered (`impl/governance/status-model.md` §7.3).

### Dependencies

None (hard). Soft, non-blocking: this story's own recommended sequencing note in
`../../dependencies.md` (S003 logically follows S001), which does not gate this task's start.

### Detailed work

For each of D-01, D-02, D-03:

1. Locate the exact source text in `plan.md`'s "Per-decision REVIEW-section mapping" section
   (D-01 = REVIEW §F row 2; D-02 = REVIEW §F row 3; D-03 = REVIEW §F row 4), plus the REVIEW §U
   closing-sentence cross-reference for owner attribution.
2. Populate `impl/governance/templates/decision-template.md`'s shape exactly: front matter (`id:
   ADR-W00-E02-S003-00N`, `type: decision`, `title`, `status: accepted`, `context`, `date:
   2026-07-12`, `deciders: [Fable 5]`, `related_source_items: [D-0N, <owning downstream epic ID>]`)
   and body sections (Decision ID, Title, Status, Context, Options considered, Decision, Rationale,
   Consequences, Related source items, Date, Deciders).
3. Add the "Formalization note" paragraph (per `plan.md` "Proposed architecture") stating this ADR
   formalizes an already-made Fable 5 decision, not a new decision-making act.
4. Add the "Safe default" subsection under Decision — populated for D-01 (REVIEW §F row 1's safe
   default is the same framework-owns-the-grant-record premise D-01 states as resolved); for D-02
   and D-03, state "no distinct safe-default stated beyond the decision itself" since REVIEW does
   not separately name one for these two.
5. Populate "Options considered" from the rejected alternative named in each REVIEW row: D-01
   rejects wowsociety holding grant authority; D-02 rejects per-subsystem registrar types; D-03
   rejects unconditional panic in production builds.
6. State "Related source items" as `D-0N` plus the downstream epic each unblocks (D-01 → W03-E01;
   D-02 → W05-E02; D-03 → W05-E01), per `story.md` "Dependencies" table.

### Expected files or components affected

New files only:

- `decisions/adr-001-framework-owns-grant-authority.md` (D-01, `id:
  ADR-W00-E02-S003-001`)
- `decisions/adr-002-single-registrar-typed-keys.md` (D-02, `id: ADR-W00-E02-S003-002`)
- `decisions/adr-003-post-seal-mutation-error-not-panic.md` (D-03, `id:
  ADR-W00-E02-S003-003`)

### Expected output

Three complete, internally consistent ADR files ready for independent review and for inclusion in
`decisions/index.md`.

### Required artifacts

Three ADR files, type "architecture decision / design document," lifecycle stage
"implementation." See `../artifacts/index.md`.

### Required evidence

Independent-review fidelity-check coverage for these three ADRs (may be part of a consolidated
nine-ADR review report — see `../evidence/index.md`).

### Related acceptance criteria

AC-W00-E02-S003-01 (all nine ADRs internally complete — this task's three-ADR slice),
AC-W00-E02-S003-03 (no ADR adds content beyond its REVIEW §F/§U source — this task's three-ADR
slice).

### Completion criteria

All three ADR files exist, pass the `decision-template.md` completeness check (no unfilled
section), and their Decision/Rationale/Consequences/Safe-default text traces to the cited REVIEW
§F row with no unlabeled added content.

### Verification method

Independent reviewer reads each of the three ADRs against REVIEW §F rows 2, 3, 4 side-by-side, per
`../verification.md`'s planned procedure.

### Risks

RISK-W00-004 (ADR-ification inadvertently adds design content beyond REVIEW §F/§U) applies
directly — mitigated by the transcription discipline in "Detailed work" above and checked in
verification.

### Rollback or recovery considerations

Not applicable — documentation-only task; a factual error found post-creation is corrected in
place and, if found after story acceptance, tracked as a deviation per `../deviations.md`.

## Implementation Record

Per mandate §8.7. Do not pre-populate implementation claims for work that has not yet occurred.

### What was actually implemented

The 3 ADR files below were authored 2026-07-12 by the story authoring pass, exactly
per this task's Detailed work steps. On 2026-07-13 this execution pass verified each against its
cited source line-by-line, corrected the ADR status vocabulary from `accepted` to `ratified`
(story-level deviation DEV-W00-E02-S003-001 — `accepted` is not in `decision-template.md`'s
status vocabulary), and fixed the round-1 independent-review findings recorded in
`../evidence/reviews/adr-fidelity-review-2026-07-13.md`.

### Components changed

None — documentation only, as planned.

### Files changed

New files (plus 2026-07-13 in-place corrections):

- `decisions/adr-001-framework-owns-grant-authority.md`
- `decisions/adr-002-single-registrar-typed-keys.md`
- `decisions/adr-003-post-seal-mutation-error-not-panic.md`

### Interfaces introduced or changed

Not applicable.

### Configuration changes

Not applicable.

### Schema or migration changes

Not applicable.

### Security changes

Not applicable — D-01 is a security-relevant *decision* being recorded, not a security change.

### Observability changes

Not applicable.

### Tests added or modified

Not applicable.

### Commits

None yet — the story's files are uncommitted working-tree additions on top of commit
`0a31186cada5c275a588c74081cf977adf346e61` (main). Committing is the programme conductor's
integration step, not this task's.

### Pull requests

None — no PR workflow for this working tree yet.

### Implementation dates

Authored 2026-07-12; verified and corrected 2026-07-13.

### Technical debt introduced

None anticipated.

### Known limitations

None recorded yet.

### Follow-up items

None recorded yet.

### Relationship to the approved plan

Matches `../plan.md` (same files, same template shape, same source mapping) with one recorded
deviation: ADR status vocabulary `ratified` instead of the plan/task text's literal
`status: accepted` — see `../deviations.md` DEV-W00-E02-S003-001.

## Verification Record

Per mandate §8.8. Table below is planned before execution; fields after it are filled after
execution.

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W00-E02-S003-01 | Reviewer checks ADR-001/002/003 against `decision-template.md`'s required sections | local checkout | all sections populated, no placeholders | review report | independent reviewer |
| AC-W00-E02-S003-03 | Reviewer checks ADR-001/002/003 text against REVIEW §F rows 2/3/4 line-by-line | local checkout, REVIEW doc open | zero unlabeled added content | review report | independent reviewer |

### Actual result

Both planned checks executed 2026-07-13. Completeness: all 3 ADRs populate every
`decision-template.md` section, zero placeholders. Fidelity: all decision content traces to the
cited sources after the round-1 fixes; elaborations are explicitly labeled Wave-00-added
clarifications. Round-1 findings on this task's slice: (a) `status: accepted` off-vocabulary on all three ADRs (story-level DEV-W00-E02-S003-001, fixed to `ratified`); (b) ADR-001's claimed-verbatim decision quote interpolated the table name — bracketed and flagged as editorial; (c) ADR-003 presented the `s.rulesReg` case identification as REVIEW §F row 4 content — re-attributed to `premier-framework-implementation-plan.md` §7 item 4 / MATRIX CS-06 and labeled a Wave-00-added clarification; (d) ADR-001 front matter gained the D-01-tuning decider so `decisions/index.md` owner cells match front matter. ADR-002 passed clean.

### Pass or fail

Pass (after in-pass round-1 fixes; round-2 re-review clean).

### Evidence identifier

EV-W00-E02-S003-001..003 (consolidated review report) + EV-W00-E02-S003-010 (scripted structure/index check) — see `../evidence/index.md`.

### Execution date

2026-07-13.

### Commit or revision

`0a31186cada5c275a588c74081cf977adf346e61` (main; story files uncommitted working-tree additions).

### Environment

Local checkout, Darwin arm64 (macOS 25.5.0); documentation-only review, no runtime needed;
concurrent sibling W00 workers active (non-timing work).

### Reviewer

W00-E02-S003 execution worker + dedicated reviewer subagent (`AdrFidelityReview`) — both
independent of the 2026-07-12 ADR authoring pass.

### Findings

Round-1 findings on this task's slice: (a) `status: accepted` off-vocabulary on all three ADRs (story-level DEV-W00-E02-S003-001, fixed to `ratified`); (b) ADR-001's claimed-verbatim decision quote interpolated the table name — bracketed and flagged as editorial; (c) ADR-003 presented the `s.rulesReg` case identification as REVIEW §F row 4 content — re-attributed to `premier-framework-implementation-plan.md` §7 item 4 / MATRIX CS-06 and labeled a Wave-00-added clarification; (d) ADR-001 front matter gained the D-01-tuning decider so `decisions/index.md` owner cells match front matter. ADR-002 passed clean.

### Retest status

Round-2 re-check after the round-1 fixes: scripted structure/index check pass
(`../evidence/logs/adr-structure-check-2026-07-13.log`); fidelity re-verified in the consolidated
report.

### Final conclusion

Pass — this task's ADRs are complete, source-faithful, and registered; AC slices satisfied.

## Deviations Record

Per mandate §8.9. Initially state that deviations are not yet known. The approved plan must not be
silently altered to hide deviations.

Story-level deviation DEV-W00-E02-S003-001 (`ratified` status vocabulary instead of the literal
`status: accepted` in this task's Detailed work step 2) applies to this task — recorded once in
`../deviations.md`, not duplicated here. No other deviations.

### Deviation ID

Not applicable yet.

### Approved plan

Not applicable yet.

### Actual implementation

Not applicable yet.

### Reason

Not applicable yet.

### Impact

Not applicable yet.

### Risks

Not applicable yet.

### Approval

Not applicable yet.

### Compensating controls

Not applicable yet.

### Follow-up work

Not applicable yet.
