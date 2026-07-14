---
id: W00-E02-S003
type: story
title: ADR-ification of D-01 through D-09
status: accepted
wave: W00
epic: W00-E02
owner: W00-E02-S003 execution worker (agent)
reviewer: unassigned
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-13
source_requirements:
  - D-01
  - D-02
  - D-03
  - D-04
  - D-05
  - D-06
  - D-07
  - D-08
  - D-09
depends_on: []
blocks:
  - W03-E01
  - W05-E02
  - W05-E01
  - W04-E04
  - W06-E03
  - W05-E04
  - W03-E02
  - W01-E02
  - W01
acceptance_criteria:
  - AC-W00-E02-S003-01
  - AC-W00-E02-S003-02
  - AC-W00-E02-S003-03
artifacts:
  - ART-W00-E02-S003-001
  - ART-W00-E02-S003-002
  - ART-W00-E02-S003-003
  - ART-W00-E02-S003-004
  - ART-W00-E02-S003-005
  - ART-W00-E02-S003-006
  - ART-W00-E02-S003-007
  - ART-W00-E02-S003-008
  - ART-W00-E02-S003-009
evidence:
  - EV-W00-E02-S003-001
  - EV-W00-E02-S003-002
  - EV-W00-E02-S003-003
  - EV-W00-E02-S003-004
  - EV-W00-E02-S003-005
  - EV-W00-E02-S003-006
  - EV-W00-E02-S003-007
  - EV-W00-E02-S003-008
  - EV-W00-E02-S003-009
  - EV-W00-E02-S003-010
decisions:
  - ADR-W00-E02-S003-001
  - ADR-W00-E02-S003-002
  - ADR-W00-E02-S003-003
  - ADR-W00-E02-S003-004
  - ADR-W00-E02-S003-005
  - ADR-W00-E02-S003-006
  - ADR-W00-E02-S003-007
  - ADR-W00-E02-S003-008
  - ADR-W00-E02-S003-009
risks:
  - RISK-W00-004
---

# W00-E02-S003 — ADR-ification of D-01 through D-09

## Story ID

W00-E02-S003.

## Title

ADR-ification of D-01 through D-09.

## Objective

Turn the nine architecture decisions D-01 through D-09 — already made by Fable 5 in
`docs/implementation/fable5-final-architecture-review-2026-07-11.md` (REVIEW) §F (rows 2-8) and §U
(D-08, D-09) — into nine durable, individually-addressable ADR files under this story's
`decisions/` directory, and register them in the decision register, so that downstream stories can
cite a stable `ADR-...` ID instead of re-reading and re-interpreting REVIEW prose each time one of
these decisions is a design input.

## Value to the framework

Nine downstream epics across five later waves (W01, W03, W04, W05, W06 — see "Dependencies" below)
each treat one of D-01..D-09 as a fixed design premise before their own stories can be planned in
detail. Mandate §11.8 states explicitly: "Record architectural and implementation decisions,
including unresolved decisions. Do not bury decisions only in prose." Leaving these nine decisions
as review-document paragraphs means every consuming story re-derives the same conclusion from the
same source text, with no single point of truth and no place to record status if a decision is
later superseded. This story is pure programme-infrastructure value — it changes no production
code — but it is a load-bearing prerequisite: without it, W03-E01, W05-E01, W05-E02, W05-E04,
W04-E04, W06-E03, W03-E02, W01-E02, and W01's secrets documentation work would each have to
individually re-litigate a decision this programme has already made.

## Problem statement

D-01 through D-09 exist today only as table rows and prose sentences in REVIEW §F ("Resolution of
the 10 unresolved questions") and §U ("Decision register"). They are real, adjudicated decisions —
each has a stated recommendation and an owner — but they have no stable identifier a downstream
`story.md` `source_requirements` or `depends_on` field can cite, no independently reviewable
document boundary, and no place in this programme's `decisions/index.md` /
`tracking/decision-register.md` structure. `impl/analysis/requirement-inventory.md` row
"D-01..D-09 | Nine ratified architecture decisions" already assigns this story (W00-E02-S003) as
their target; this story is the mechanism that closes that assignment.

## Source requirements

D-01, D-02, D-03, D-04, D-05, D-06, D-07, D-08, D-09 — `impl/analysis/requirement-inventory.md` §B,
row "D-01..D-09 — Nine ratified architecture decisions." Primary source text: REVIEW §F (rows 2-8,
covering D-01 through D-07) and REVIEW §U (covering D-08 and D-09, and cross-referencing D-01
through D-07).

## Current-state assessment

Confirmed facts (read directly from the source documents during this story's planning):

- REVIEW §F rows 2-8 state D-01 through D-07, each with a "Fable 5 decision + safe default" cell.
  Row 1 (Q1, IdP `grant_id` claim contract) is a genuine human decision tracked separately as
  `DEC-Q1` in `requirement-inventory.md` §B — it is not one of D-01..D-09 and is out of scope here.
- REVIEW §U states D-08 (pgx query tracer over `otelpgx`) and D-09 (secrets boot-time-once
  resolution, restart-based rotation) in full, and restates D-01 through D-07 in condensed
  cross-reference form (same decisions, not new content).
- No ADR file for any of D-01..D-09 exists anywhere in the repository today — `decisions/` under
  this story did not exist before this story's own file creation.
- `impl/tracking/decision-register.md` is the programme-level register these nine ADRs must
  ultimately also appear in; this story populates the story-scoped `decisions/index.md` as the
  producing unit — cross-registration into the wave/programme-level tracking register is tracked as
  a `tracking/` maintenance action, not duplicated content invented by this story.

## Desired state

Nine ADR files exist at `decisions/adr-001-framework-owns-grant-authority.md` through
`decisions/adr-009-secrets-boot-time-rotation-contract.md`, each following the shape of
`impl/governance/templates/decision-template.md`, each citing its exact REVIEW §F row or §U
sentence, each stating recommendation, options considered (where the source names an alternative),
decision, rationale, safe default (where the source states one), consequences, and owner.
`decisions/index.md` registers all nine. No ADR adds substantive design content beyond what REVIEW
§F/§U already states (mandate §18: "Do not silently resolve ambiguous architecture decisions" —
here read as its corollary: do not silently *add* to an already-resolved one either).

## Scope

- Authoring nine ADR files, one per D-01 through D-09, under this story's `decisions/` directory.
- Writing `decisions/index.md` registering all nine.
- Grouping the nine ADR-authoring tasks into three tasks (T001, T002, T003) per the task-grouping
  decision recorded in `plan.md`.
- Populating this story's `story.md`, `plan.md`, `implementation.md`, `verification.md`,
  `deviations.md`, `closure.md`, `tasks/index.md`, `artifacts/index.md`, `evidence/index.md` per
  the mandate §8 required-content shapes and this repository's Adaptation 1/2 (flat task files;
  `index.md`-only artifact/evidence directories).

## Out of scope

- Making any new architecture decision, resolving any ambiguity REVIEW §F/§U leaves open, or
  extending a decision's stated safe default beyond what the source says. Per the epic's own scope
  statement (`../../epic.md` "Out of scope"): "this epic's S003 formalizes an already-made decision
  into the programme's ADR structure, it does not re-litigate or extend it."
- Resolving `DEC-Q1`, `DEC-Q9`, `DEC-Q10` — the three genuine human decisions tracked separately in
  `requirement-inventory.md` §B. These are not part of the D-01..D-09 set.
- Implementing any of the nine decisions in code (e.g. building the `identity_grant` table for
  D-01, or the `authz_epoch` table for D-06). Implementation happens in each decision's owning
  downstream story (see "Dependencies" below); this story only formalizes the decision record.
- Cross-registering these nine ADRs into `impl/tracking/decision-register.md`. That programme-level
  register is maintained as part of `impl/tracking/` upkeep, not authored fresh by this story; this
  story's own `decisions/index.md` is the authoritative story-scoped registration mandate §11.8
  requires.

## Assumptions

- REVIEW §F and §U, as they exist in the repository today at
  `docs/implementation/fable5-final-architecture-review-2026-07-11.md`, are the final, ratified
  text of these nine decisions — this story does not verify REVIEW's own internal accuracy, only
  transcribes it faithfully into ADR form.
- "Status: accepted" on each ADR file asserts that the underlying decision was already made by
  Fable 5 in REVIEW — it does not assert that this programme's own tracking-execution process
  (task `todo` → `done`) has completed. This distinction is stated explicitly in each ADR body (see
  "ADR shape" in `plan.md`) so the two senses of "accepted" cannot be conflated by a reader who
  only sees the front matter.
- The nine downstream epic dependencies listed below are drawn from `../../dependencies.md`
  (epic-level) and `../../../../dependencies.md` (wave-level), both already written by the
  programme conductor; this story does not re-derive them independently, only confirms consistency.

## Dependencies

| Depends on | Type | Notes |
|---|---|---|
| (none — hard) | — | This story has no hard `depends_on`. |
| W00-E02-S001 (quality-baselines) | soft, non-blocking | Per `../../dependencies.md` "Internal sequencing recommendation": S003 should logically follow S001 so ADR authors work from a freshly-confirmed baseline snapshot, but this is a recommendation, not an enforced `depends_on`. |

Downstream items this story unblocks (reproduced from `../../dependencies.md` and
`../../../../dependencies.md`, scoped to the ADR each depends on):

| Downstream epic | Depends on | Why |
|---|---|---|
| W03-E01 (SEC-01) | ADR D-01 | D-01 resolves grant-table authority split (framework vs wowsociety) that SEC-01's design assumes |
| W05-E02 (AR-02) | ADR D-02 | D-02 resolves the single-Registrar-type-with-typed-keys design AR-01 T2/AR-02 T1 implement directly |
| W05-E01 (AR-01) | ADR D-03 | D-03 resolves post-seal-mutation error-vs-panic policy that AR-01 T8/AR-04 T4 implement |
| W04-E04 (DATA-08 W6) | ADR D-04 | D-04 resolves the hash_version discriminator design W6-T1 implements |
| W06-E03 (REL-01) | ADR D-05 | D-05 resolves GoReleaser split-mode approach for REL-01 T6 |
| W05-E04 (SEC-04) | ADR D-06 | D-06 resolves cross-pod cache invalidation transport (epoch table, not message bus) |
| W03-E02 (SEC-06) | ADR D-07 | D-07 resolves JWKS-client governance model |
| W01-E02 (FBL-06) | ADR D-08 | D-08 resolves pgx query tracing approach (thin in-kernel tracer, not otelpgx) |
| W01 (secrets docs, CS-25) | ADR D-09 | D-09 resolves secrets rotation contract (restart-based, v1) |

## Affected packages or components

None. This story produces only planning/governance documentation under `impl/`. No Go package,
build file, or runtime configuration is touched.

## Compatibility considerations

Not applicable — no code or schema is changed by this story. Each decision's own compatibility
implications are the concern of its owning downstream implementation story (e.g. D-01's
compatibility impact on wowsociety's `identity_impersonation_session` table is SEC-01's concern,
not this story's).

## Security considerations

D-01 and D-07 are security-classified decisions (grant authority; JWKS trusted-issuer config gate).
This story's own security consideration is limited to transcription fidelity: an ADR that
understates or overstates a security decision's safe default would misinform every downstream
consumer. The verification procedure (`verification.md`) specifically checks each security-bearing
ADR (D-01, D-07) for fidelity against its source line.

## Performance considerations

Not applicable to this story directly. D-06 (authz epoch table) has downstream performance
implications (poll latency vs LISTEN/NOTIFY), but those are SEC-04's (W05-E04) concern to design
and measure, not this story's.

## Observability considerations

D-08 is itself an observability decision (pgx query tracing). This story's own observability
consideration is again transcription fidelity — D-08's ADR must accurately state the rejection of
`otelpgx` and the reason (vendor types leaking into `kernel/database`), since that rationale is
what a future contributor would otherwise be tempted to silently revisit.

## Migration considerations

Not applicable. D-04 concerns a future database migration (`hash_version` column) but that
migration itself is DATA-08 W6's (W04-E04) implementation work, not this story's.

## Documentation requirements

This story's entire output is documentation: nine ADR files plus `decisions/index.md`. No
additional documentation beyond the mandate §8 required story/task files and this decisions set is
required.

## Acceptance criteria

- **AC-W00-E02-S003-01** — All nine ADR files (`decisions/adr-001-...md` through
  `decisions/adr-009-...md`) exist and are internally complete: each has front matter (`id`,
  `type: decision`, `title`, `status: accepted`, `context`, `date`, `deciders`,
  `related_source_items`) and every body section required by `decision-template.md` (Decision ID,
  Title, Status, Context, Options considered, Decision, Rationale, Consequences, Related source
  items, Date, Deciders) populated — no section left as an unfilled template placeholder.
- **AC-W00-E02-S003-02** — `decisions/index.md` registers all nine ADRs with D-0N ID, ADR file
  name, title, status (`accepted`), and owner; the index is internally consistent with the nine
  ADR files' own front matter (no ID, title, or owner mismatch).
- **AC-W00-E02-S003-03** — No ADR adds content beyond its REVIEW §F/§U source: independent review
  (per `verification.md`) confirms every recommendation, safe default, and consequence stated in
  each ADR traces to specific REVIEW §F/§U text, with any necessary elaboration explicitly flagged
  as a Wave-00-added clarification rather than folded in as if it were original decision text
  (mitigates RISK-W00-004).

## Required artifacts

Nine ADR files, type "architecture decision / design document" per `artifact-template.md`'s type
list, lifecycle stage "implementation" (mandate §9.3: "architecture decisions" is a named
implementation-stage example). See `artifacts/index.md`.

## Required evidence

One review report per ADR (nine total) — the independent-review fidelity-check record confirming
each ADR against its REVIEW §F/§U source line-by-line. See `evidence/index.md`.

## Definition of ready

This story satisfies `impl/governance/definition-of-ready.md`'s Story DoR: specific (nine named
decisions, not an aspirational theme), bounded (scope/out-of-scope stated above), implementable
(the content to transcribe is fully known and quoted in `plan.md`), independently reviewable and
verifiable (each ADR can be checked against its own REVIEW citation), traceable
(`source_requirements` lists D-01..D-09), measurable acceptance criteria (three ACs above),
dependencies identified (table above), assumptions recorded (above), plan drafted (`plan.md`),
artifacts/evidence anticipated (above), and all mandate §8.4 consideration sections addressed or
marked not-applicable with rationale (above).

## Definition of done

This story will satisfy `impl/governance/definition-of-done.md` when: all three tasks (T001, T002,
T003) reach `done`; all three acceptance criteria have a `pass` verification result with a
registered evidence ID in `verification.md`; `artifacts/index.md` and `evidence/index.md` list all
nine ADRs and nine review reports respectively with the required fields; `deviations.md` states "no
deviations" or lists them; `closure.md` is complete; and the independent-review checklist
(mandate §14, reproduced in `definition-of-done.md`) has passed clean, specifically confirming no
ADR silently adds content beyond its REVIEW §F/§U source (RISK-W00-004 resolved or explicitly
accepted).

## Risks

- **RISK-W00-004** (epic/wave-level, reproduced here as it applies directly to this story's only
  failure mode): ADR-ification inadvertently introduces new design content beyond what D-01..D-09
  already state in REVIEW §F/§U, silently resolving an ambiguity the mandate requires to stay
  explicit. Mitigation: each ADR cites its REVIEW §F/§U source verbatim for recommendation/safe
  default/owner; any elaboration beyond the source is flagged as a Wave-00-added clarification.
  Contingency: independent review checks each ADR line-by-line against its source before this story
  moves to `accepted`.

## Residual-risk expectations

Even after mitigation, some residual risk remains that a future reader conflates an ADR's
`status: accepted` (meaning: the underlying decision was already made by Fable 5) with this
programme's own tracking-execution status (meaning: the task that authored/registered the file has
completed its lifecycle). This story mitigates that by stating the distinction explicitly in the
body of every ADR (see "ADR shape" in `plan.md`) and in this story's own "Status discipline"
framing, but cannot eliminate the risk that a downstream reader skips the body text and reads only
front matter. This residual risk is accepted, not further mitigated, because the alternative —
withholding `status: accepted` until this programme's own process completes — would misrepresent
the decision's actual, already-ratified state, which is a worse inaccuracy.

## Plan

See `plan.md` for the full proposed approach, including the task-grouping rationale (mandate §12)
and the per-decision REVIEW-section mapping.
