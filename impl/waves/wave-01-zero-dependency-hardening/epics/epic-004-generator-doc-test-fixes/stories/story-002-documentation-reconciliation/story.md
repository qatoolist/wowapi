---
id: W01-E04-S002
type: story
title: Documentation reconciliation — plan traceability fix, DX-05 residual, wowsociety upstream register
status: accepted
wave: W01
epic: W01-E04
owner: W01Docs
reviewer: unassigned
priority: P2
created_at: 2026-07-12
updated_at: 2026-07-13
source_requirements:
  - T-DOC-01
  - DX-05
  - FBL-03
depends_on:
  - W01-E04-S001
blocks: []
acceptance_criteria:
  - AC-W01-E04-S002-01
  - AC-W01-E04-S002-02
  - AC-W01-E04-S002-03
artifacts:
  - ART-W01-E04-S002-001
  - ART-W01-E04-S002-002
  - ART-W01-E04-S002-003
  - ART-W01-E04-S002-004
  - ART-W01-E04-S002-005
evidence:
  - EV-W01-E04-S002-001
  - EV-W01-E04-S002-002
  - EV-W01-E04-S002-003
decisions: []
risks:
  - RISK-W01-E04-002
  - RISK-W01-E04-003
---

# W01-E04-S002 — Documentation reconciliation

## Story ID

W01-E04-S002.

## Title

Documentation reconciliation — plan traceability fix, DX-05 residual, wowsociety upstream register.

## Objective

Fix the implementation plan document's own internal traceability contradiction (§6 vs §9 on DX-05's
status), close DX-05's three residual sub-tasks (T3 reconciliation, T4 version-gate planning, T5
recorded as a deferred cross-reference), and produce a precise, PROD-level coordination recommendation
for reconciling the wowsociety upstream finding register — without performing the wowsociety-side edit
directly, since that register lives outside this repository's boundary.

## Value to the framework

A programme that claims to track "what was implemented, deferred, rejected, or superseded" (mandate
§1, objective list) is only as trustworthy as its own internal consistency. A plan document whose
traceability matrix (§6) disagrees with its own execution record (§9) about the same finding (DX-05)
undermines the evidence-driven-completion discipline mandate §2.5 requires of every story in this
programme — a reader cannot tell, from the authoritative document, whether DX-05's T1/T2 shipped or
not. This story is not cosmetic: mandate §2.4's traceability chain ("Source document → Source
requirement → Wave → Epic → Story → Acceptance criterion → Task → Implementation change → Artifact →
Verification evidence → Review → Acceptance") only holds if the source documents it starts from are
themselves internally consistent. Closing this gap protects the integrity of every other story's
traceability claims that cite the plan document.

## Problem statement

`impl/analysis/requirement-inventory.md` row T-DOC-01 states: "Fix plan §6-vs-§9 DX-05 inconsistency
... §6 (traceability matrix) marked DX-05 `PLANNED` while §9 (execution record) reports DX-05 T1/T2
`EXECUTED`." This is confirmed by REVIEW §E as a traceability defect in the PLAN document itself — not
a disagreement about whether the work happened (§9's prose is accurate), but a stale/uncorrected matrix
cell that was never updated when T1/T2 actually executed. Separately, DX-05's row in
`requirement-inventory.md` states "T1/T2 EXECUTED; §6-vs-§9 status inconsistency = T-DOC-01," meaning
DX-05 itself has three residual sub-tasks (T3, T4, T5) beyond the already-executed T1/T2 that this
story must plan. Separately again, FBL-03 ("Reconcile wowsociety upstream register," disposition
`planned`, target `W01-E04-S002`) requires marking wowsociety's own `docs/upstream/`-documented
findings PF-2/PF-6/RFF-001 as closed once their fixes land or are confirmed already-resolved.

All three problems are documentation/traceability-integrity problems, not code-correctness problems —
this is why they are grouped into one story distinct from S001's code-level generator fixes.

## Source requirements

T-DOC-01, DX-05 (residual scope: T3, T4, T5-deferred), FBL-03.

## Current-state assessment

**Confirmed from source documents (not independently re-verified against the live plan document's
current line numbers by this story's planning — see "Assumptions" below):**

- `impl/analysis/requirement-inventory.md` row DX-05: "T1/T2 EXECUTED; §6-vs-§9 status inconsistency =
  T-DOC-01" — this is the canonical allocation's own statement that T1 (README status banner rewrite)
  and T2 (upgrade-policy rewrite) are done and were verified in W00-E01, not this story's job to
  re-verify.
- `impl/analysis/requirement-inventory.md` row T-DOC-01: "Fix plan §6-vs-§9 DX-05 inconsistency," P3,
  disposition `planned`, target `W01-E04-S002`.
- `impl/analysis/requirement-inventory.md` row FBL-03: "Reconcile wowsociety upstream register," DOC,
  P2, disposition `planned`, target `W01-E04-S002`, note: "Mark PF-2/PF-6/RFF-001 etc. as closed when
  their fixes land."
- REVIEW Answer 18 (cited by this epic's governing task instructions, not independently re-read by this
  story's author): "no active workarounds remain... mark the 2 stale upstream docs resolved" — meaning
  PF-6 and RFF-001 specifically are already resolved and only need their register entries corrected,
  not new fixes.
- PF-2's closure is explicitly contingent on sibling story W01-E04-S001's DX-02 task (the generator
  permission-verb fix) actually landing, since PF-2 is the wowsociety-documented instance of that exact
  bug.

**Not yet confirmed (planning-time assumptions, to be confirmed at implementation time):**

- The exact current line numbers of `docs/implementation/premier-framework-implementation-plan.md`'s §6
  table and §9 execution record for the DX-05 row. This story's task-001 describes the required
  correction by its semantic content (the DX-05 row's status cell), not by a line number, since the
  document may have shifted since the review that identified this defect.
- The exact current state of `internal/cli/cli.go`'s commands and flags, needed for DX-05 T3's
  blueprint-11 example reconciliation. This story assumes the CLI has evolved since blueprint-11 was
  authored (that is the premise of the finding) but does not assume how many examples are stale.

## Desired state

The plan document's §6 table and §9 execution record agree on DX-05's status (both show T1/T2
`EXECUTED`). Every blueprint-11 CLI example either matches `internal/cli/cli.go`'s real commands/flags
or has been explicitly marked for deletion, with no example left silently wrong. A version-
compatibility gate design exists for `wowapi version` rejecting mutating generator commands on
incompatible major/minor pairing, with its dependency on S001's version-verification plumbing made
explicit. DX-05 T5 is explicitly recorded as deferred to W06/REL-03, not silently dropped. The
wowsociety upstream register has a precise, actionable coordination recommendation for PF-2 (contingent
on S001), PF-6, and RFF-001 (both corrected to already-resolved).

## Scope

- **T-DOC-01**: describing the exact §6 table-row correction needed for DX-05 (status cell corrected
  from `PLANNED` to reflect T1/T2 `EXECUTED`, matching §9), as planning documentation — the edit itself
  happens later when the task moves to `in-progress`, not as part of this planning exercise.
- **DX-05 T3**: reconciling blueprint-11's CLI examples against `internal/cli/cli.go`'s real commands
  and flags, with a per-example implement-or-delete decision.
- **DX-05 T4**: planning a version-compatibility gate so `wowapi version` (or the generator commands
  themselves) fail on an incompatible major/minor framework-version pairing, built on S001's DX-01
  version-verification plumbing.
- **DX-05 T5**: recorded, not implemented — deferred to W06, cross-referenced to REL-03's shared
  compat-gate plumbing.
- **FBL-03**: a precise PROD-level coordination recommendation for the wowsociety upstream register's
  PF-2 (contingent on S001), PF-6, and RFF-001 entries.

## Out of scope

- **DX-05 T1/T2** — already `EXECUTED` per `requirement-inventory.md`'s DX-05 row, verified in W00-E01.
  Not re-verified by this story.
- **DX-05 T5's full compat-gate build** — explicitly "shared with REL-03" per the plan, and REL-03 is
  W06 scope (`requirement-inventory.md` row REL-03: target `W06-E02-S002..S003`). This story records
  the deferral with the cross-reference; it does not attempt the build.
- **The actual edit to `docs/implementation/premier-framework-implementation-plan.md`** — a source
  document outside `impl/waves/`. This story's T001 describes the required edit as planning
  documentation; the edit itself is deferred to implementation time per this epic's own governing
  constraint.
- **The actual edit to the wowsociety upstream register** — that register is a file in the
  `wowsociety` repository, not `wowapi`. Per mandate §2.3's framework/product boundary discipline, this
  story can only plan/recommend that edit (a PROD-level coordination note, following the pattern of
  `requirement-inventory.md` §D's PROD-01 through PROD-05 rows), not execute it.
- **Sibling story W01-E04-S001's own DX-02 fix** — this story's T003 (FBL-03) merely records PF-2's
  dependency on S001's fix; it does not re-implement or re-plan that fix.

## Assumptions

- The plan document's §6/§9 sections still contain the DX-05 inconsistency at the time this story
  begins implementation; if the document has already been corrected by other means, T001 becomes a
  verification-only task (confirm no inconsistency remains) rather than a fix.
- `internal/cli/cli.go`'s exact current command/flag surface is not independently re-read as part of
  authoring this story; T002's DX-05 T3 sub-item re-confirms it fresh at implementation time rather
  than trusting blueprint-11's age.
- REVIEW Answer 18's characterization of PF-6/RFF-001 as already-resolved is taken as given from this
  epic's own governing task instructions; T003 does not re-derive that conclusion, only acts on it.

## Dependencies

- **Story-level, front matter `depends_on: ["W01-E04-S001"]`**: T003's FBL-03 sub-task for PF-2's
  closure cannot honestly recommend closing PF-2 in the wowsociety register until S001's DX-02
  permission-verb fix has actually landed — the register entry would be recommending closure of a bug
  that is not yet fixed. This is a genuine story-level dependency, not merely a note.
- **Task-level, soft/plumbing dependency**: T002's DX-05 T4 sub-item (version-compatibility gate) reuses
  the version-verification plumbing S001's T001 builds for DX-01 (the `go list -m` resolution check,
  version-comparison logic). T002 should not begin implementing T4 before S001's T001 has landed, to
  avoid duplicating or diverging from that logic. This does not block T002's *planning*, only its
  *implementation* ordering — recorded at task level in `tasks/task-002-dx05-residual-reconciliation.md`.
- **No dependency on W01-E02 or W01-E03** — this story's scope is entirely documentation and CLI-example
  reconciliation, disjoint from observability and HTTP-hardening work.

## Affected packages or components

- `docs/implementation/premier-framework-implementation-plan.md` (source document, edited later, not by
  this planning story).
- Documentation referencing blueprint-11's CLI examples (exact file path to be confirmed at
  implementation time — likely under a `docs/blueprint/` or similar directory per this repository's
  `wowapi-framework-blueprint` convention).
- `internal/cli/cli.go` — read-only reference for T3's reconciliation and T4's design, not modified by
  this story's planning.
- The wowsociety repository's `docs/upstream/` register — out of this repository's write boundary
  entirely; referenced only for the coordination recommendation.

## Compatibility considerations

DX-05 T4's version-compatibility gate is itself a compatibility-enforcement mechanism (rejecting
incompatible major/minor framework-version pairings on mutating generator commands) — its design must
not itself introduce a compatibility break for currently-valid version pairings. This is a design
consideration for T002, not an implementation this story performs.

## Security considerations

Largely not applicable — this is a documentation-correction and planning story. The one
security-adjacent consideration: DX-05 T4's version-compatibility gate is a supply-chain-correctness
control (preventing a mutating generator command from running against an incompatible framework
version, which could otherwise silently generate code incompatible with the target kernel's contracts).

## Performance considerations

Not applicable.

## Observability considerations

Not applicable.

## Migration considerations

Not applicable — no schema, data, or configuration migration is involved in documentation reconciliation.

## Documentation requirements

This entire story is a documentation-requirements story: the plan document's §6 table correction, the
blueprint-11 example reconciliation, the DX-05 T4 design note, and the FBL-03 coordination
recommendation are all documentation deliverables in their own right.

## Acceptance criteria

- **AC-W01-E04-S002-01**: The plan document's §6 traceability-matrix row for DX-05 is described
  precisely enough (by this story's T001) that its correction — showing T1/T2 as `EXECUTED`, matching
  §9's record — can be executed without re-deriving the fix from source documents again.
- **AC-W01-E04-S002-02**: DX-05 T3's blueprint-11 CLI examples each have a recorded implement-or-delete
  decision against `internal/cli/cli.go`'s real commands/flags; DX-05 T4's version-compatibility gate is
  planned with its dependency on S001's version-verification plumbing stated explicitly; DX-05 T5 is
  recorded as deferred-to-W06 with the REL-03 cross-reference, not silently dropped.
- **AC-W01-E04-S002-03**: FBL-03's target register plan states precisely that PF-2 is closeable
  contingent on S001's DX-02 task landing (with the cross-story dependency recorded), and that PF-6/
  RFF-001 are corrected to already-resolved status per REVIEW Answer 18 — delivered as a documented,
  precise PROD-level coordination recommendation, not a direct edit to a file this repository does not
  own.

## Required artifacts

See `artifacts/index.md`. Expected: the corrected plan-document §6 row (recorded as a doc diff once
executed), a DX-05 T3 example-reconciliation decision table, a DX-05 T4 version-gate design note, and
an FBL-03 coordination-recommendation document.

## Required evidence

See `evidence/index.md`. Expected: doc diffs for T-DOC-01 and DX-05 T3's reconciliation, and the FBL-03
coordination note itself (which functions as both artifact and evidence that the recommendation was
produced).

## Definition of ready

Per `governance/definition-of-ready.md`: story is specific (three distinct, named documentation
defects), bounded (explicit out-of-scope list above), implementable (each task's detailed work is
either fully specified or explicitly conditional/described-not-performed per mandate §8.5's guidance),
independently reviewable and verifiable (each of the three tasks produces its own reviewable artifact),
traceable (source_requirements front matter populated), measurable acceptance criteria (three ACs
above), dependencies identified (`depends_on: ["W01-E04-S001"]` in front matter).

## Definition of done

Per `governance/definition-of-done.md`: all three tasks reach `done`; the plan-document correction is
actually applied (once this task moves past `todo`) and verified against §9; DX-05 T3/T4 decisions are
recorded; DX-05 T5's deferral is documented; the FBL-03 coordination recommendation is delivered and
reviewed; no source requirement (T-DOC-01, DX-05, FBL-03) is silently dropped.

## Risks

RISK-W01-E04-002 (DX-05 T3's per-example reconciliation could expand beyond estimated scope if
blueprint-11 has drifted significantly) and RISK-W01-E04-003 (FBL-03's cross-repository nature means
this story cannot verify the wowsociety-side edit was actually applied) — see `../../risks.md` (epic
level) for full elaboration.

## Residual-risk expectations

RISK-W01-E04-003 is expected to remain as a permanently-accepted residual risk even after this story's
acceptance: this story's closure is scoped to producing a correct, actionable recommendation, not to
verifying a downstream repository was actually edited. If the wowsociety-side edit is never applied,
that is tracked at programme level in `impl/tracking/deferred-items-register.md`, not as an open item
against this story.
