---
id: W00-E02-S003-T003
type: task
title: Observability / secrets decisions (D-08, D-09)
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
  - ART-W00-E02-S003-008
  - ART-W00-E02-S003-009
evidence: []
---

# W00-E02-S003-T003 — Observability / secrets decisions (D-08, D-09)

## Task Definition

Per mandate §8.6. This section defines the task before work begins.

### Task objective

Author two ADR files formalizing D-08 (pgx query tracing via a thin in-kernel `pgx.QueryTracer`
over the existing observability `Tracer` port, not `otelpgx`) and D-09 (secrets: boot-time-once
resolution + restart-based rotation as the documented v1 contract, no vault client in the kernel) —
the two decisions in this story's "observability / secrets" cluster (`plan.md` task-grouping
rationale: both are kernel cross-cutting infrastructure decisions consumed by W01-scoped epics,
thematically distinct from T001's application-model cluster and T002's data/release/security
cluster).

### Parent story

W00-E02-S003 — ADR-ification of D-01 through D-09.

### Owner

Unassigned.

### Status

`done` — implemented, verified, and output registered (`impl/governance/status-model.md` §7.3).

### Dependencies

None (hard). No dependency on T001 or T002 — each task draws from a disjoint slice of REVIEW
§F/§U. Note: unlike D-01..D-07 (sourced from REVIEW §F), D-08 and D-09 are sourced from REVIEW §U
directly — this task's source citation differs in location, not in required rigor.

### Detailed work

For each of D-08, D-09:

1. Locate the exact source text in `../plan.md`'s "Per-decision REVIEW-section mapping" section
   (both D-08 and D-09 are sourced from REVIEW §U — the condensed decision-register sentence, plus
   the fuller phrasing given in this task's own brief, which elaborates without contradicting the
   §U summary and is used as the same-decision source text per `plan.md`'s note on this point).
2. Populate `impl/governance/templates/decision-template.md`'s shape exactly: front matter (`id:
   ADR-W00-E02-S003-00N`, `type: decision`, `title`, `status: accepted`, `context`, `date:
   2026-07-12`, `deciders: [Fable 5]`, `related_source_items: [D-0N, <owning downstream epic ID>]`)
   and body sections (Decision ID, Title, Status, Context, Options considered, Decision, Rationale,
   Consequences, Related source items, Date, Deciders).
3. Add the "Formalization note" paragraph stating this ADR formalizes an already-made Fable 5
   decision, not a new decision-making act.
4. Add the "Safe default" subsection under Decision — for both D-08 and D-09, state "no distinct
   safe-default stated beyond the decision itself" since REVIEW §U does not separately name one
   (§U's format is recommendation + rejected-alternative + rationale, not
   recommendation-plus-fallback).
5. Populate "Options considered": D-08 explicitly rejects `otelpgx` (third-party OTel bridge,
   would bind OTel vendor types into `kernel/database`, breaking port discipline). D-09 explicitly
   rejects two alternatives: hot-reload plumbing through every secret consumer (rejected for v1 —
   high complexity, modest payoff) and a vault client embedded in the kernel (rejected outright,
   not deferred); the file-provider (K8s mounted-secret pattern) is noted as the *next increment*,
   not a rejected option — record it under Consequences as the stated future path, not under
   Options considered as something rejected.
6. State "Related source items" as `D-0N` plus the downstream epic each unblocks (D-08 → W01-E02;
   D-09 → W01, secrets docs/CS-25), per `../story.md` "Dependencies" table.

### Expected files or components affected

New files only:

- `decisions/adr-008-pgx-query-tracer-not-otelpgx.md` (D-08, `id: ADR-W00-E02-S003-008`)
- `decisions/adr-009-secrets-boot-time-rotation-contract.md` (D-09, `id:
  ADR-W00-E02-S003-009`)

### Expected output

Two complete, internally consistent ADR files ready for independent review and for inclusion in
`decisions/index.md`.

### Required artifacts

Two ADR files, type "architecture decision / design document," lifecycle stage "implementation."
See `../artifacts/index.md`.

### Required evidence

Independent-review fidelity-check coverage for these two ADRs (may be part of a consolidated
nine-ADR review report — see `../evidence/index.md`).

### Related acceptance criteria

AC-W00-E02-S003-01 (all nine ADRs internally complete — this task's two-ADR slice),
AC-W00-E02-S003-03 (no ADR adds content beyond its REVIEW §F/§U source — this task's two-ADR
slice).

### Completion criteria

Both ADR files exist, pass the `decision-template.md` completeness check (no unfilled section),
and their Decision/Rationale/Consequences/Safe-default text traces to the cited REVIEW §U source
with no unlabeled added content.

### Verification method

Independent reviewer reads both ADRs against REVIEW §U's D-08 and D-09 text side-by-side, per
`../verification.md`'s planned procedure.

### Risks

RISK-W00-004 applies directly — mitigated by the transcription discipline in "Detailed work" above
and checked in verification. D-09 carries an additional nuance to preserve: the file-provider
next-increment note must not be recorded as a rejected option (it is an explicitly *deferred*
future path, distinct from the two genuinely rejected alternatives).

### Rollback or recovery considerations

Not applicable — documentation-only task; a factual error found post-creation is corrected in
place and, if found after story acceptance, tracked as a deviation per `../deviations.md`.

## Implementation Record

Per mandate §8.7. Do not pre-populate implementation claims for work that has not yet occurred.

### What was actually implemented

The 2 ADR files below were authored 2026-07-12 by the story authoring pass, exactly
per this task's Detailed work steps. On 2026-07-13 this execution pass verified each against its
cited source line-by-line, corrected the ADR status vocabulary from `accepted` to `ratified`
(story-level deviation DEV-W00-E02-S003-001 — `accepted` is not in `decision-template.md`'s
status vocabulary), and fixed the round-1 independent-review findings recorded in
`../evidence/reviews/adr-fidelity-review-2026-07-13.md`.

### Components changed

None — documentation only, as planned.

### Files changed

New files (plus 2026-07-13 in-place corrections):

- `decisions/adr-008-pgx-query-tracer-not-otelpgx.md`
- `decisions/adr-009-secrets-boot-time-rotation-contract.md`

### Interfaces introduced or changed

Not applicable.

### Configuration changes

Not applicable.

### Schema or migration changes

Not applicable.

### Security changes

Not applicable.

### Observability changes

Not applicable — D-08 is an observability *decision* being recorded, not an observability change.

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
| AC-W00-E02-S003-01 | Reviewer checks ADR-008/009 against `decision-template.md`'s required sections | local checkout | all sections populated, no placeholders | review report | independent reviewer |
| AC-W00-E02-S003-03 | Reviewer checks ADR-008/009 text against REVIEW §U line-by-line | local checkout, REVIEW doc open | zero unlabeled added content | review report | independent reviewer |

### Actual result

Both planned checks executed 2026-07-13. Completeness: all 2 ADRs populate every
`decision-template.md` section, zero placeholders. Fidelity: all decision content traces to the
cited sources after the round-1 fixes; elaborations are explicitly labeled Wave-00-added
clarifications. Round-1 findings on this task's slice: (a) `status: accepted` off-vocabulary on both ADRs (story-level DEV-W00-E02-S003-001, fixed to `ratified`); (b) ADR-008 cited "NOT `otelpgx`" as REVIEW §U verbatim — §U actually reads "`otelpgx` rejected to keep vendor types out of `kernel/database`"; quote and fuller-phrasing attributions corrected to MATRIX CS-05 / `../plan.md`'s D-08 mapping; (c) ADR-009's Rationale asserted an unsourced vault-client-rejection reason — reworded with sources quoted and the elaboration explicitly labeled "Wave-00-added clarification, not source text" (the one substantive AC-03 defect found); (d) ADR-009's title disagreed between front matter, body, and index — aligned to the §U-faithful "documented v1 contract" everywhere; hot-reload quote re-attributed to MATRIX CS-25 / `../plan.md`.

### Pass or fail

Pass (after in-pass round-1 fixes; round-2 re-review clean).

### Evidence identifier

EV-W00-E02-S003-008..009 (consolidated review report) + EV-W00-E02-S003-010 (scripted structure/index check) — see `../evidence/index.md`.

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

Round-1 findings on this task's slice: (a) `status: accepted` off-vocabulary on both ADRs (story-level DEV-W00-E02-S003-001, fixed to `ratified`); (b) ADR-008 cited "NOT `otelpgx`" as REVIEW §U verbatim — §U actually reads "`otelpgx` rejected to keep vendor types out of `kernel/database`"; quote and fuller-phrasing attributions corrected to MATRIX CS-05 / `../plan.md`'s D-08 mapping; (c) ADR-009's Rationale asserted an unsourced vault-client-rejection reason — reworded with sources quoted and the elaboration explicitly labeled "Wave-00-added clarification, not source text" (the one substantive AC-03 defect found); (d) ADR-009's title disagreed between front matter, body, and index — aligned to the §U-faithful "documented v1 contract" everywhere; hot-reload quote re-attributed to MATRIX CS-25 / `../plan.md`.

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
