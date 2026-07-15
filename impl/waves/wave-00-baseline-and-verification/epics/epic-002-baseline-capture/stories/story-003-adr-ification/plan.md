---
id: PLAN-W00-E02-S003
type: plan
parent_story: W00-E02-S003
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
---

# Plan — W00-E02-S003 (ADR-ification of D-01 through D-09)

This is ADR-authoring work, not code. There is no application architecture, package change, API,
persistence change, or migration involved. The "proposed architecture" for this story is a
documentation structure decision (which files, what shape, how grouped), not a software
architecture decision.

## Proposed architecture

Nine ADR files under `decisions/`, one per D-01..D-09, each following
`impl/governance/templates/decision-template.md`'s shape exactly (front matter: `id`, `type:
decision`, `title`, `status`, `context`, `date`, `deciders`, `related_source_items`; body sections:
Decision ID, Title, Status, Context, Options considered, Decision, Rationale, Consequences, Related
source items, Date, Deciders). `decisions/index.md` is the story-scoped register mandate §11.8
requires ("do not bury decisions only in prose").

Two fields are added beyond the bare template, because this task's brief explicitly requires them
and the template's `## Context`/`## Decision` sections are the natural home:

- A **"Formalization note"** paragraph in each ADR's body (placed directly under the title, before
  `## Decision ID`) stating explicitly: "This ADR formalizes a decision Fable 5 already made in
  `docs/implementation/fable5-final-architecture-review-2026-07-11.md` §F/§U; this ADR is the
  programme's own durable record of it, not a new decision-making act." This is required so
  `status: accepted` in the front matter cannot be misread as asserting this programme's own
  process completed — see `story.md` "Residual-risk expectations."
- A **"Safe default"** subsection under `## Decision`, populated where the source states one (D-01
  explicitly; most of D-02..D-09 fold their safe default into the decision statement itself, so
  this subsection states "no distinct safe-default stated beyond the decision itself" where that's
  the case, per this task's explicit instruction not to force text onto a field that doesn't apply).

## Implementation strategy

Transcribe, not invent. For each D-0N: locate its exact REVIEW §F row (D-01..D-07) or §U sentence
(D-08, D-09) — both quoted verbatim below — and populate the ADR template fields from that text
only. Where REVIEW names a rejected alternative (explicit "vs" framing or an explicit rejection),
record it under "Options considered." Where REVIEW does not name an alternative, "Options
considered" states that explicitly rather than inventing one (mandate §18).

## Expected package or module changes

None. No Go source, `go.mod`, migration, or configuration file is touched.

## Expected file changes where determinable

All new files, fully determinable in advance (this plan's own author list, reproduced in "Task
breakdown" below and in `tasks/index.md`).

## Contracts and interfaces

Not applicable.

## Data structures

Not applicable.

## APIs

Not applicable.

## Configuration changes

Not applicable.

## Persistence changes

Not applicable.

## Migration strategy

Not applicable — D-04's `hash_version` migration is DATA-08 W6's implementation concern (W04-E04),
not this story's.

## Concurrency implications

Not applicable.

## Error-handling strategy

Not applicable.

## Security controls

Not applicable to this story's own execution. D-01's and D-07's *content* is security-relevant —
see `story.md` "Security considerations" — but authoring their ADR text introduces no new security
control itself.

## Observability changes

Not applicable to this story's own execution; D-08 is an observability *decision* being
transcribed, not an observability change this story makes.

## Testing strategy

Not applicable in the code-testing sense. The equivalent quality control is the independent-review
fidelity check defined in `verification.md`: an independent reviewer reads each ADR against its
REVIEW §F/§U source line-by-line and confirms no added content and no missing required field.

## Regression strategy

Not applicable — no existing behavior exists to regress.

## Compatibility strategy

Not applicable.

## Rollout strategy

The nine ADR files and `decisions/index.md` are committed as part of this story's normal file
creation; there is no phased or feature-flagged rollout of a documentation file.

## Rollback strategy

If an ADR is later found to misstate its source, it is corrected in place (documentation is not
versioned/released software requiring a rollback mechanism) and the correction is itself recorded
as a deviation if it occurs after this story has reached `accepted` (mandate §2.6 — the approved
plan/record is not silently rewritten; a correction after acceptance would be tracked via a new
change, not a silent edit).

## Implementation sequence

1. T001 — author ADR-001, ADR-002, ADR-003 (D-01, D-02, D-03).
2. T002 — author ADR-004, ADR-005, ADR-006, ADR-007 (D-04, D-05, D-06, D-07).
3. T003 — author ADR-008, ADR-009 (D-08, D-09).
4. After all three tasks: assemble `decisions/index.md` registering all nine, cross-checked against
   each ADR's own front matter for ID/title/owner consistency.
5. Independent review pass per `verification.md`.

Tasks 1-3 have no ordering dependency on each other (each draws from a disjoint slice of REVIEW
§F/§U) and could be parallelized across workers; they are listed sequentially above only because
this plan is authored by a single planning pass, not because T002 requires T001's output.

## Task breakdown — and the task-grouping decision (mandate §12)

**Judgment call, made explicitly per this task's brief and mandate §12's "avoid excessive
fragmentation into trivial tasks that provide no tracking value" guidance:** the nine ADRs are
grouped into **three tasks by logical cluster**, not left as nine near-identical one-ADR tasks and
not collapsed into a single nine-ADR task. Both alternatives were considered and rejected:

- **Nine tasks (one per ADR) — rejected.** Each ADR-authoring unit is small (transcribe one REVIEW
  row into one template). Nine tasks would each carry the full task.md front-matter/section
  overhead (mandate §8.6's objective/owner/status/dependencies/detailed-work/... fields, per this
  repository's Adaptation 1 single-file-per-task shape) for a unit of work that takes minutes and
  produces one file. This is exactly the "trivial tasks that provide no tracking value" mandate §12
  warns against — nine near-identical task records add filing overhead without adding independent
  reviewability, since all nine ADRs are reviewed together against the same two source sections
  (§F, §U) in one verification pass anyway.
- **One task (all nine ADRs) — rejected.** A single task covering all nine decisions mixes genuinely
  unrelated subject areas (application-model/session-authority design vs data/release/security
  policy vs observability/secrets infrastructure) under one `detailed work` / `completion criteria`
  /`verification method` description, which mandate §12 also flags as a decomposition trigger
  ("affect several unrelated framework capabilities... combine implementation with unrelated...").
  It would also make the task's own risk profile blurry — a fidelity error in D-08's transcription
  would sit in the same task record as an unrelated error in D-01's, with no way to mark one `done`
  independently of the other's fix.
- **Three tasks by logical cluster — chosen.** Groups decisions that share a subject-matter
  neighborhood and are more likely to be reviewed together by someone with the same domain context:
  - **T001 — application-model / session-authority decisions**: D-01 (framework owns grant
    authority), D-02 (single Registrar + typed keys), D-03 (post-seal mutation error not panic).
    All three concern the framework's core application-model and session/capability-authority
    surface — the same surface AR-01/AR-02/SEC-01 (W05-E01, W05-E02, W03-E01) later implement
    against.
  - **T002 — data / release / security decisions**: D-04 (audit hash_version column), D-05
    (GoReleaser skip-publish split), D-06 (authz epoch table not message bus), D-07 (JWKS
    trusted-issuer config gate). Four decisions spanning persistence, release engineering, and
    security-config governance — grouped because each is a standalone infrastructure/policy
    decision consumed by a different downstream epic (W04-E04, W06-E03, W05-E04, W03-E02
    respectively), unlike T001's shared application-model surface, but each is similarly
    self-contained and evidence-light to transcribe.
  - **T003 — observability / secrets decisions**: D-08 (pgx query tracer, not otelpgx), D-09
    (secrets boot-time-once resolution, restart-based rotation). Both are kernel cross-cutting
    infrastructure decisions (tracing, secrets) consumed by W01-scoped epics (W01-E02, W01 secrets
    docs) rather than the W03/W04/W05/W06 product-facing decisions in T001/T002 — smallest cluster,
    kept separate because it is thematically distinct from both T001 and T002, not because of size
    alone.

This grouping is stated explicitly here, per mandate §18 ("Do not silently resolve ambiguous
architecture decisions" — read here as: do not silently resolve a stated planning judgment call
either; record it).

## Expected artifacts

Nine ADR files (`ART-W00-E02-S003-001` through `ART-W00-E02-S003-009`), registered in
`artifacts/index.md`.

## Expected evidence

Nine independent-review fidelity-check reports (`EV-W00-E02-S003-001` through
`EV-W00-E02-S003-009`), registered in `evidence/index.md`.

## Unresolved questions

None blocking this story's own execution. This story does not resolve `DEC-Q1`, `DEC-Q9`, or
`DEC-Q10` (out of scope, see `story.md`) and does not need to — none of D-01..D-09 depends on those
three human decisions being resolved first.

## Approval conditions

This plan is approved for execution once: (a) the three-task grouping above is accepted as the
task-decomposition approach (no alternative grouping silently substituted later), and (b) the
per-decision REVIEW-section mapping below is confirmed complete (all nine decisions map to a cited
REVIEW location, none invented).

## Per-decision REVIEW-section mapping (source of truth for T001/T002/T003 authors)

Quoted from `docs/implementation/fable5-final-architecture-review-2026-07-11.md`.

### D-01 — REVIEW §F, row 2 (Q2)

> wowsociety `identity_impersonation_session` vs framework grant table authority — **Fable 5
> decision (framework boundary)** — resolved — **Framework owns grant validity/expiry/revocation;
> wowsociety keeps its table for product UX/audit only.** This is the correct dependency direction
> (WOW-Review §1). Recorded as decision **D-01**.

Cross-reference, REVIEW §F row 1 (Q1, `DEC-Q1`, out of scope but adjacent): the safe default that
unblocks Q1 build ("build the server-side `identity_grant` table + resolver now... If the IdP
cannot emit `grant_id`, the framework still owns the grant record") is the same framework-owns-the-
grant-record premise D-01 states as a resolved decision — D-01's ADR may note this adjacency without
importing Q1's own unresolved-human-decision content.

Owner: Fable 5 (framework); D-01 tuning (IdP claim shape) = product/security-lead (REVIEW §U:
"owner = Fable 5 (framework) except D-01 tuning = product/security-lead").

### D-02 — REVIEW §F, row 3 (Q3)

> AR-01 `Registrar` type: one shared vs per-subsystem — **Fable 5 decision (public contract)** —
> resolved — **One generic owner-bound `Registrar` capability type**, with per-subsystem *typed
> keys* (`Key[T]`) rather than per-subsystem registrar types. Capability confusion is prevented by
> the key's phantom type + owner binding, not by multiplying registrar types. Decision **D-02**.

Rejected alternative, explicit in the question framing: per-subsystem registrar types (multiple
`Registrar` types, one per subsystem).

Owner: Fable 5 (framework, public contract).

### D-03 — REVIEW §F, row 4 (Q4)

> AR-01/AR-04 post-seal mutation: error vs panic in prod — **Fable 5 decision
> (concurrency/lifecycle)** — resolved — **Error in production builds; panic only under an explicit
> `dev`/test build tag.** A framework must not convert a benign retained-handle into a prod crash.
> Decision **D-03**.

Rejected alternative, explicit in the question framing: unconditional panic on post-seal mutation
in production builds (the wowsociety `s.rulesReg` retention case cited as the reason this would be
wrong).

Owner: Fable 5.

### D-04 — REVIEW §F, row 5 (Q5)

> DATA-08 W6 audit-hash `hash_version` discriminator design — **Answerable by technical analysis**
> — resolved — **Add a `hash_version smallint NOT NULL DEFAULT 1` column in the same migration that
> widens `chainHash`'s field coverage; verification branches on it.** Historical rows verify under
> v1; new rows under v2 (metadata + tx_id included). Standard append-only-log versioning. Decision
> **D-04**.

No explicit rejected alternative named in the source beyond the general "answerable by technical
analysis" framing; the ADR states this rather than inventing a rejected option.

Owner: Fable 5.

### D-05 — REVIEW §F, row 6 (Q6)

> REL-01 GoReleaser split-mode (`--skip=publish` vs hand-rolled) — **Answerable by
> evidence/testing** — resolved — **Use GoReleaser `release --skip=publish` for build-candidate + a
> separate `goreleaser publish` step.** Supported in current GoReleaser; no hand-rolled pipeline
> needed. Decision **D-05** (verify against pinned GoReleaser version at implementation time).

Rejected alternative, explicit in the question framing: a hand-rolled release pipeline.

Owner: Fable 5 (verify against pinned GoReleaser version at implementation — REVIEW's own stated
caveat, not yet independently confirmed).

### D-06 — REVIEW §F, row 7 (Q7)

> SEC-04 cross-pod cache invalidation transport (LISTEN/NOTIFY vs epoch poll) — **Fable 5 decision
> (concurrency)** — resolved — **Per-tenant epoch integer in a small `authz_epoch` table, polled on
> the existing authz read path; Postgres `LISTEN/NOTIFY` as an optional latency optimisation, not a
> correctness dependency.** Avoids a new message bus in the kernel. Decision **D-06**.

Rejected alternative, explicit in the question framing and the decision text: a new message bus in
the kernel (LISTEN/NOTIFY is retained only as an optional, non-load-bearing latency optimization,
not as the correctness mechanism).

Owner: Fable 5 (concurrency decision).

### D-07 — REVIEW §F, row 8 (Q8)

> SEC-06 JWKS-client governance model — **Fable 5 decision (security)** — resolved — **Require
> trusted-issuer/egress config to be a declared, fingerprinted `config` field; reject a custom JWKS
> `*http.Client` in `prod` profile unless the trusted-issuer allowlist is set.** Decision **D-07**.

Rejected alternative, implicit but stated as the constraint being imposed: an ungoverned/undeclared
custom JWKS `*http.Client` in `prod` profile (permitted only once the trusted-issuer allowlist
gate is satisfied).

Owner: Fable 5 (security decision).

### D-08 — REVIEW §U

> **D-08** (pgx query tracing via a thin in-kernel `pgx.QueryTracer` over the existing observability
> port — `otelpgx` rejected to keep vendor types out of `kernel/database`)

Full sentence context from §U: "D-08 (pgx query tracing via a thin in-kernel `pgx.QueryTracer`
implementation (~50 LOC) over the existing observability `Tracer` port — NOT `otelpgx` (a
third-party bridge would bind OTel vendor types into `kernel/database`, breaking the port
discipline the adapters layer gets right)" — this fuller phrasing appears in this task's own brief
and is consistent with, and elaborates without contradicting, the §U summary line; both are used
together as the D-08 source text since they are the same decision restated at two points in the
same document section.

Rejected alternative, explicit: `otelpgx` (third-party OTel bridge).

Owner: Fable 5.

### D-09 — REVIEW §U

> **D-09** (secrets: boot-time-once resolution + restart-based rotation is the documented v1
> contract; file-provider is the next increment, no vault client in the kernel)

Fuller phrasing, consistent with the brief: "most orchestrators roll pods on secret change;
hot-reload plumbing through every consumer is real complexity with modest v1 payoff. File-provider
(K8s mounted-secret pattern) is the next increment when needed — NOT a vault client in the kernel."

Rejected alternatives, explicit: hot-reload plumbing through every secret consumer (rejected for
v1, high complexity vs modest payoff); a vault client embedded in the kernel (rejected outright, not
deferred).

Owner: Fable 5.

Cross-reference: REVIEW §U's closing sentence applies to all nine — "Each: recommendation stated,
safe default given, owner = Fable 5 (framework) except D-01 tuning = product/security-lead." This
confirms every ADR's owner field and confirms D-01 is the only decision with a split
owner/tuning-owner distinction.
