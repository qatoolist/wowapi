---
id: W00-E02-S003-T002
type: task
title: Data / release / security decisions (D-04, D-05, D-06, D-07)
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
  - ART-W00-E02-S003-004
  - ART-W00-E02-S003-005
  - ART-W00-E02-S003-006
  - ART-W00-E02-S003-007
evidence: []
---

# W00-E02-S003-T002 — Data / release / security decisions (D-04, D-05, D-06, D-07)

## Task Definition

Per mandate §8.6. This section defines the task before work begins.

### Task objective

Author four ADR files formalizing D-04 (audit `hash_version` column), D-05 (GoReleaser
`--skip=publish` + separate `publish` step), D-06 (per-tenant `authz_epoch` table, not a message
bus), and D-07 (JWKS trusted-issuer config gate) — the four decisions in this story's "data /
release / security" cluster (`plan.md` task-grouping rationale: four standalone infrastructure/
policy decisions, each consumed by a different downstream epic).

### Parent story

W00-E02-S003 — ADR-ification of D-01 through D-09.

### Owner

Unassigned.

### Status

`done` — implemented, verified, and output registered (`impl/governance/status-model.md` §7.3).

### Dependencies

None (hard). No dependency on T001 or T003 — each task draws from a disjoint slice of REVIEW
§F/§U.

### Detailed work

For each of D-04, D-05, D-06, D-07:

1. Locate the exact source text in `../plan.md`'s "Per-decision REVIEW-section mapping" section
   (D-04 = REVIEW §F row 5; D-05 = REVIEW §F row 6; D-06 = REVIEW §F row 7; D-07 = REVIEW §F row
   8), plus the REVIEW §U closing-sentence cross-reference for owner attribution.
2. Populate `impl/governance/templates/decision-template.md`'s shape exactly: front matter (`id:
   ADR-W00-E02-S003-00N`, `type: decision`, `title`, `status: accepted`, `context`, `date:
   2026-07-12`, `deciders: [Fable 5]`, `related_source_items: [D-0N, <owning downstream epic ID>]`)
   and body sections (Decision ID, Title, Status, Context, Options considered, Decision, Rationale,
   Consequences, Related source items, Date, Deciders).
3. Add the "Formalization note" paragraph stating this ADR formalizes an already-made Fable 5
   decision, not a new decision-making act.
4. Add the "Safe default" subsection under Decision — for D-04, D-06, D-07, state "no distinct
   safe-default stated beyond the decision itself" (REVIEW does not separately name one). For D-05,
   note the source's own stated caveat as the closest analogue to a safe default: "verify against
   the pinned GoReleaser version at implementation time (this is a caveat, not yet independently
   confirmed)" — record it under Consequences/Rationale rather than forcing it into "Safe default"
   if it does not read as a distinct fallback option, since REVIEW frames it as a verification
   caveat on the decision, not an alternate safe path. Use judgment consistent with "do not force
   text onto a field that doesn't apply"; if in doubt, state explicitly which framing was chosen and
   why.
5. Populate "Options considered": D-04 — REVIEW names no explicit rejected alternative beyond
   "answerable by technical analysis"; state this rather than inventing one. D-05 rejects a
   hand-rolled release pipeline. D-06 rejects a new message bus in the kernel (LISTEN/NOTIFY
   retained only as optional, non-load-bearing latency optimization). D-07 rejects an
   ungoverned/undeclared custom JWKS `*http.Client` in `prod` profile.
6. State "Related source items" as `D-0N` plus the downstream epic each unblocks (D-04 → W04-E04;
   D-05 → W06-E03; D-06 → W05-E04; D-07 → W03-E02), per `../story.md` "Dependencies" table.

### Expected files or components affected

New files only:

- `decisions/adr-004-audit-hash-version-column.md` (D-04, `id: ADR-W00-E02-S003-004`)
- `decisions/adr-005-goreleaser-skip-publish-split.md` (D-05, `id: ADR-W00-E02-S003-005`)
- `decisions/adr-006-authz-epoch-table-not-message-bus.md` (D-06, `id:
  ADR-W00-E02-S003-006`)
- `decisions/adr-007-jwks-trusted-issuer-config-gate.md` (D-07, `id:
  ADR-W00-E02-S003-007`)

### Expected output

Four complete, internally consistent ADR files ready for independent review and for inclusion in
`decisions/index.md`.

### Required artifacts

Four ADR files, type "architecture decision / design document," lifecycle stage "implementation."
See `../artifacts/index.md`.

### Required evidence

Independent-review fidelity-check coverage for these four ADRs (may be part of a consolidated
nine-ADR review report — see `../evidence/index.md`).

### Related acceptance criteria

AC-W00-E02-S003-01 (all nine ADRs internally complete — this task's four-ADR slice),
AC-W00-E02-S003-03 (no ADR adds content beyond its REVIEW §F/§U source — this task's four-ADR
slice).

### Completion criteria

All four ADR files exist, pass the `decision-template.md` completeness check (no unfilled
section), and their Decision/Rationale/Consequences/Safe-default text traces to the cited REVIEW
§F row with no unlabeled added content.

### Verification method

Independent reviewer reads each of the four ADRs against REVIEW §F rows 5, 6, 7, 8 side-by-side,
per `../verification.md`'s planned procedure.

### Risks

RISK-W00-004 applies directly — mitigated by the transcription discipline in "Detailed work" above
and checked in verification. D-05 carries an additional note-worthy nuance (REVIEW's own stated
caveat about verifying against the pinned GoReleaser version) that must be preserved, not smoothed
away, since dropping it would understate the source's own acknowledged uncertainty.

### Rollback or recovery considerations

Not applicable — documentation-only task; a factual error found post-creation is corrected in
place and, if found after story acceptance, tracked as a deviation per `../deviations.md`.

## Implementation Record

Per mandate §8.7. Do not pre-populate implementation claims for work that has not yet occurred.

### What was actually implemented

The 4 ADR files below were authored 2026-07-12 by the story authoring pass, exactly
per this task's Detailed work steps. On 2026-07-13 this execution pass verified each against its
cited source line-by-line, corrected the ADR status vocabulary from `accepted` to `ratified`
(story-level deviation DEV-W00-E02-S003-001 — `accepted` is not in `decision-template.md`'s
status vocabulary), and fixed the round-1 independent-review findings recorded in
`../evidence/reviews/adr-fidelity-review-2026-07-13.md`.

### Components changed

None — documentation only, as planned.

### Files changed

New files (plus 2026-07-13 in-place corrections):

- `decisions/adr-004-audit-hash-version-column.md`
- `decisions/adr-005-goreleaser-skip-publish-split.md`
- `decisions/adr-006-authz-epoch-table-not-message-bus.md`
- `decisions/adr-007-jwks-trusted-issuer-config-gate.md`

### Interfaces introduced or changed

Not applicable.

### Configuration changes

Not applicable.

### Schema or migration changes

Not applicable — D-04's actual `hash_version` migration is DATA-08 W6's (W04-E04) implementation
work, not this task's.

### Security changes

Not applicable — D-07 is a security-relevant *decision* being recorded, not a security change.

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
| AC-W00-E02-S003-01 | Reviewer checks ADR-004/005/006/007 against `decision-template.md`'s required sections | local checkout | all sections populated, no placeholders | review report | independent reviewer |
| AC-W00-E02-S003-03 | Reviewer checks ADR-004/005/006/007 text against REVIEW §F rows 5/6/7/8 line-by-line | local checkout, REVIEW doc open | zero unlabeled added content | review report | independent reviewer |

### Actual result

Both planned checks executed 2026-07-13. Completeness: all 4 ADRs populate every
`decision-template.md` section, zero placeholders. Fidelity: all decision content traces to the
cited sources after the round-1 fixes; elaborations are explicitly labeled Wave-00-added
clarifications. Round-1 findings on this task's slice: (a) `status: accepted` off-vocabulary on all four ADRs (story-level DEV-W00-E02-S003-001, fixed to `ratified`); (b) ADR-004's decision block claimed "REVIEW §F row 5, quoted verbatim" while blending MATRIX CS-20 phrasing — attribution corrected; (c) ADR-006 attributed "P1, not on the critical path" to `requirement-inventory.md` when it is REVIEW §F row 7's blocks-column text — citation corrected. ADR-005 and ADR-007 passed clean.

### Pass or fail

Pass (after in-pass round-1 fixes; round-2 re-review clean).

### Evidence identifier

EV-W00-E02-S003-004..007 (consolidated review report) + EV-W00-E02-S003-010 (scripted structure/index check) — see `../evidence/index.md`.

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

Round-1 findings on this task's slice: (a) `status: accepted` off-vocabulary on all four ADRs (story-level DEV-W00-E02-S003-001, fixed to `ratified`); (b) ADR-004's decision block claimed "REVIEW §F row 5, quoted verbatim" while blending MATRIX CS-20 phrasing — attribution corrected; (c) ADR-006 attributed "P1, not on the critical path" to `requirement-inventory.md` when it is REVIEW §F row 7's blocks-column text — citation corrected. ADR-005 and ADR-007 passed clean.

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
