---
id: W03-E04-S001
type: story
title: Relationship semantics — party-subject evaluation, full subject-kind matrix, mutation governance
status: accepted
wave: W03
epic: W03-E04
owner: unassigned
reviewer: unassigned
priority: high
created_at: 2026-07-12
updated_at: 2026-07-13
source_requirements:
  - DATA-07
depends_on:
  - W03-E01
blocks: []
acceptance_criteria:
  - AC-W03-E04-S001-01
  - AC-W03-E04-S001-02
  - AC-W03-E04-S001-03
artifacts: []
evidence: []
decisions: []
risks:
  - RISK-W03-003
---

# W03-E04-S001 — Relationship semantics — party-subject evaluation, full subject-kind matrix, mutation governance

## Story ID

W03-E04-S001

## Title

Relationship semantics — party-subject evaluation, full subject-kind matrix, mutation governance

## Objective

Extend `Checker.Has` to resolve actor → active capacity → optional party through the post-SEC-01
authoritative principal model (so party-subject edges are evaluated, not silently skipped); extend
`Checker.Has` to cover every schema-enumerated `subject_kind` with unsupported kinds failing closed;
and ensure every authorization-input mutation (edge create/revoke) is ownership-checked, attributed
(via DATA-06 T2's mechanism, consumed not reimplemented), audited, and versioned, with the
cache-invalidation portion of that acceptance criterion deferred-linked to W05-E04-S002. This is PLAN
§5.3 DATA-07, T1, T2, and T4 — T3 is explicitly excluded, cross-referenced to DATA-06 T2
(W02-E04-S001) per PLAN's own "reuse... do not reimplement" instruction.

## Value to the framework

Per PLAN §5.3's own evidence, `Checker.Has` (`kernel/relationship/relationship.go:42-66`) today
filters `subject_kind='capacity'` only — party-subject edges exist in the schema but are, per the
code's own comment, "not consulted yet." This is a genuine functional gap: a relationship-based
authorization check that should evaluate `true` for a party-subject edge silently evaluates `false`
instead, because the evaluation logic simply never looks at that edge type. This story closes that
gap on top of the authoritative principal model W03-E01 establishes, extends evaluation to every
schema-enumerated `subject_kind` (not just `capacity` and party), and closes the parallel governance
gap on relationship mutations themselves (ownership check, attribution, audit, versioning) — while
explicitly not duplicating DATA-06 T2's actor-attribution fix, which shares the exact same file
(`registrar_pg.go`) and is owned by W02-E04-S001, not this story.

## Problem statement

PLAN §5.3's own DATA-07 task table, quoted exactly:

- "T1. Resolve actor → active capacity → optional party through the authoritative principal model |
  **Hard dependency on PF-SEC's SEC-01 — do not schedule before it lands** | `Checker.Has` can
  evaluate party-subject edges | Test: seed a party-subject edge, resolve an actor carrying a party,
  assert previously-false now true | `DATA-07/party-subject-eval/` | Blocked, not merely related."
- "T2. Extend `Checker.Has` to cover every schema-enumerated `subject_kind` | T1 | Every enumerated
  kind has an evaluation branch; unsupported kind fails closed | Matrix test |
  `DATA-07/subject-kind-matrix/` | Confirm which enumerated kinds are live requirements vs. dead
  schema surface first."
- "T3. Source real actor for `Relate`/mirror `Upsert` — **reuse DATA-06 T2's mechanism directly, same
  file, do not reimplement** | DATA-06 T2 | Real `created_by`; same missing-actor rule as DATA-06 |
  Shared test helper | `DATA-07/actor-attribution/` | High duplication risk if staffed
  independently — sequence as one shared task."
- "T4. Every authorization-input mutation is ownership-checked, attributed, audited, versioned, and
  invalidates relevant caches | T1-T3; **also depends on SEC-04's cache-epoch work** | Edge
  create/revoke writes audit rows and triggers observable cache invalidation | Test |
  `DATA-07/mutation-audit-cache/` | Second cross-work-package dependency — do not assume PF-SEC
  delivers on PF-DATA's timeline."

Per `impl/analysis/wave-allocation-detail.md`'s own grouping instruction: "S001 T1, T2, T4 (T3
consumed from DATA-06 T2 — DATA-06 is W02 scope, cross-reference only, do not reimplement; hard dep
W03-E01 accepted; SEC-04 epoch dep noted soft — cache-invalidation AC deferred-linked to
W05-E04-S002)." PLAN's own PF-DATA cross-cutting note (3): "DATA-07 has a hard dependency on SEC-01,
secondary on SEC-04 — sequence accordingly, not in parallel."

## Source requirements

DATA-07 (T1, T2, T4 — T3 excluded, cross-referenced to DATA-06 T2 / W02-E04-S001).

## Current-state assessment

Per PLAN §5.3's own evidence citation (to be re-confirmed at this story's own execution commit,
consistent with this programme's fail-first re-confirmation convention):

- `Checker.Has` (`kernel/relationship/relationship.go:42-66`) filters `subject_kind='capacity'`
  only — party-subject edges exist in the schema but are never consulted, per the code's own comment.
- The same nil-actor gap DATA-06 identifies exists in the same file (`registrar_pg.go`) for
  `Relate`/mirror `Upsert` — this is T3's scope, owned by DATA-06 T2 (W02-E04-S001), not
  reimplemented here.
- No confirmed mutation-governance mechanism (ownership check, attribution, audit, versioning, cache
  invalidation) currently exists for relationship-edge create/revoke operations — to be re-confirmed
  at this story's own start commit.
- Per `requirement-inventory.md` row DATA-07: "No confirmed usage" in wowsociety —
  `grep -rn "kernel/relationship"` returns zero matches across wowsociety, including
  `committeeseat.go` (which mentions "ReBAC" conceptually in a comment but does not import or call
  `kernel/relationship`) — to be re-verified at this story's own ship time, not assumed from the
  cited snapshot.

## Desired state

`Checker.Has` resolves actor → active capacity → optional party through the post-SEC-01 authoritative
principal model; a seeded party-subject edge, when resolved through an actor carrying a party, is
evaluated as `true` where it was previously (incorrectly) `false`. Every schema-enumerated
`subject_kind` has an explicit evaluation branch in `Checker.Has`; an unsupported/unenumerated kind
fails closed. Every authorization-input mutation (edge create/revoke) is ownership-checked,
attributed (via DATA-06 T2's mechanism, consumed via the shared file, not reimplemented here),
audited, and versioned — the cache-invalidation sub-criterion is explicitly deferred-linked to
W05-E04-S002 and tracked as such, not silently dropped nor silently assumed complete.

## Scope

- T1 — resolve actor → active capacity → optional party through the authoritative principal model
  (the model W03-E01 establishes); `Checker.Has` can evaluate party-subject edges.
- T2 — extend `Checker.Has` to cover every schema-enumerated `subject_kind`; confirm which enumerated
  kinds are live requirements versus dead schema surface first; unsupported kinds fail closed.
- T4 — every authorization-input mutation is ownership-checked, attributed (via DATA-06 T2's
  mechanism), audited, and versioned. The cache-invalidation portion is deferred-linked to
  W05-E04-S002 (SEC-04's epoch table, D-06).

## Out of scope

- **T3 (source real actor for `Relate`/mirror `Upsert`)** — per PLAN's own instruction: "reuse
  DATA-06 T2's mechanism directly, same file, do not reimplement." T3 is DATA-06 T2's shared fix
  surface (W02-E04-S001, another worker's scope). This story's T4 *consumes* that mechanism once it
  exists; it does not reimplement or re-plan it.
- **SEC-01 itself (W03-E01)** — a hard, blocking dependency, not this story's own implementation
  scope.
- **SEC-04's cache-bounding/epoch-table work (W05-E04-S002)** — this story's T4 cache-invalidation
  acceptance criterion is deferred-linked to it, but the epoch table itself is built in W05, not
  here.

## Assumptions

- **W03-E01 is accepted, not merely started, before this story's implementation work begins.**
  PLAN's own words, verbatim: "do not schedule before it lands." If W03-E01 has not reached
  `accepted` at this story's intended start time, this story's work does not begin — see "Risks" and
  RISK-W03-E04-002 (epic-level `risks.md`).
- **DATA-06 T2 (W02-E04-S001) has landed, or T4's dependency on it is tracked as a blocking input, not
  worked around by an independent reimplementation.** If DATA-06 T2 has not landed by the time this
  story's T4 is ready, T4 is blocked on that specific external input.
- **The cache-invalidation sub-criterion of T4 is not assumed to land on this story's own timeline.**
  Per PLAN's own cross-cutting note: "do not assume PF-SEC delivers on PF-DATA's timeline" (applied
  here as: do not assume W05 delivers the epoch table on W03's timeline). This story's T4 is
  structured so its non-cache-invalidation portions (ownership check, attribution, audit, versioning)
  can close independently of the cache-invalidation portion's landing time.
- Which schema-enumerated `subject_kind` values are live requirements versus dead schema surface is
  not assumed here — T2's own risk note requires this be confirmed first, at implementation time, not
  invented in this document.
- PLAN's own citation of "no confirmed direct usage" of `kernel/relationship` in wowsociety is treated
  as a snapshot to re-confirm (fresh grep) at this story's own execution time, not blindly trusted.

## Dependencies

**Hard, blocking: W03-E01 must be `accepted`.** This story's T1 cannot meaningfully implement
actor-resolution against the pre-SEC-01 claim-trusting `Actor` shape — PLAN's own words: "Blocked, not
merely related." **Soft, cross-reference: W02-E04-S001 (DATA-06 T2).** This story's T4 consumes
DATA-06 T2's actor-attribution mechanism in `registrar_pg.go`; if DATA-06 T2 has not landed, T4 is
blocked on that specific input, tracked explicitly. **Soft, deferred-link: W05-E04-S002 (SEC-04's
epoch table, D-06).** T4's cache-invalidation acceptance criterion is deferred-linked to it — see
"Out of scope" and "Assumptions."

## Affected packages or components

`kernel/relationship/relationship.go` (`Checker.Has`); the SEC-01 principal-resolution surface
(`kernel/auth`), consumed as an input, not modified by this story; for T4's cache-invalidation
sub-criterion, a future authz-cache interface from W05-E04 (consumed once it exists, not built here).

## Compatibility considerations

Per `requirement-inventory.md` row DATA-07 and PLAN's own wowsociety-impact note: "No confirmed
direct usage" — `grep -rn "kernel/relationship"` returns zero matches across wowsociety today,
including `committeeseat.go` (mentions "ReBAC" conceptually but does not import or call
`kernel/relationship`). This is re-verified fresh at this story's own execution time (not merely
trusted from the cited snapshot) per PLAN's own instruction: "Re-verify at DATA-07 ship time." Given
zero confirmed usage, this story carries materially lower wowsociety-compatibility risk than
W03-E01/E03; the extension of `Checker.Has`'s evaluation logic (T1/T2) is additive from any current
consumer's perspective (a previously-`false` result for a party-subject or newly-supported
`subject_kind` edge becomes correctly `true`, or fails closed for a genuinely unsupported kind — never
a previously-`true` result silently flipping to `false` for a kind that was already correctly
evaluated).

## Security considerations

T1 closes a real authorization-evaluation gap: a party-subject edge that should grant access was
silently evaluated as denying it, which is a fail-closed-by-accident outcome today but becomes a
correctness gap once party-subject edges are an active part of the authorization model this story
enables. T2's "unsupported kind fails closed" requirement is itself a security control — an
unenumerated or newly-introduced `subject_kind` must never be silently treated as `true` or silently
ignored (which could itself be a fail-open risk depending on how "ignored" is implemented). T4's
ownership-check, attribution, audit, and versioning requirements make relationship-edge mutations a
governed, auditable operation rather than an ungoverned one.

## Performance considerations

T1's actor → active capacity → optional party resolution adds a resolution step to `Checker.Has`'s
evaluation path beyond what exists today — not separately benchmarked as part of this story's
acceptance criteria unless the fail-first/implementation-time work surfaces an unacceptable latency
regression, in which case it is recorded as a finding, not silently absorbed.

## Observability considerations

T4's audit-write requirement is itself an observability deliverable for relationship-edge mutations.
The cache-invalidation sub-criterion's own observability ("triggers observable cache invalidation,"
per PLAN's own T4 acceptance wording) is deferred along with the rest of that sub-criterion to
W05-E04-S002's landing.

## Migration considerations

None anticipated for T1/T2 (evaluation-logic changes only, no schema change). T4's audit-write
requirement may require a new or extended audit table/column if the existing `kernel/audit`
mechanism does not already cover relationship-edge mutations — to be confirmed at implementation
time, not assumed here.

## Documentation requirements

Document `Checker.Has`'s extended evaluation logic (party-subject resolution path, the full
subject-kind matrix, and the fail-closed behavior for unsupported kinds) and the mutation-governance
contract (ownership check, attribution, audit, versioning, and the deferred-linked cache-invalidation
status) in whatever documentation currently covers the `kernel/relationship` module.

## Acceptance criteria

- **AC-W03-E04-S001-01**: `Checker.Has` resolves actor → active capacity → optional party through the
  post-SEC-01 authoritative principal model; a test seeds a party-subject edge, resolves an actor
  carrying a party, and asserts the previously-false evaluation is now correctly `true`.
- **AC-W03-E04-S001-02**: Every schema-enumerated `subject_kind` has an explicit evaluation branch in
  `Checker.Has`; a matrix test confirms every enumerated kind; an unsupported/unenumerated kind fails
  closed, not silently `true` or silently ignored.
- **AC-W03-E04-S001-03**: Every authorization-input mutation (edge create/revoke) is
  ownership-checked, attributed (via DATA-06 T2's mechanism, consumed not reimplemented), writes an
  audit row, and is versioned, proven by a test. The cache-invalidation portion of this criterion is
  explicitly deferred-linked to W05-E04-S002 and tracked as such (not silently dropped, not silently
  assumed complete) if W05-E04-S002 has not landed by this story's own closure time.

## Required artifacts

- `Checker.Has`'s extended party-subject evaluation logic (T1).
- `Checker.Has`'s full subject-kind evaluation matrix, with fail-closed handling (T2).
- The mutation-governance implementation (ownership check, attribution consumption, audit write,
  versioning) for relationship-edge create/revoke (T4).
See `artifacts/index.md`.

## Required evidence

- Party-subject-edge seeded test output (previously-false now true) (AC-01).
- Subject-kind matrix test output, including the fail-closed case (AC-02).
- Mutation-governance test output (ownership check, attribution, audit, versioning); the
  cache-invalidation sub-criterion's status recorded explicitly as deferred-linked if not yet landed
  (AC-03).
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`: `story.md`
and `plan.md` complete, acceptance criteria numbered and measurable, dependencies recorded
(`depends_on: [W03-E01]`), **W03-E01 confirmed `accepted` (not merely started)** as an explicit entry
condition, owner/reviewer assignment pending.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; all three acceptance
criteria verified with evidence in `evidence/index.md` (AC-W03-E04-S001-03's cache-invalidation
sub-criterion recorded as deferred-linked if W05-E04-S002 has not landed, per its own documented
deferral); `closure.md` completed; independent review passed per mandate §14, with specific
confirmation T3's scope was correctly cross-referenced to DATA-06 (W02-E04-S001) rather than
reimplemented in this story.

## Risks

RISK-W03-003 (T4's cache-invalidation acceptance criterion depends on W05-E04-S002 landing, which
may not occur on this story's timeline) — see epic-level `risks.md` for full detail and
mitigation/contingency. Also see RISK-W03-E04-002 (epic-level) for the risk of this story's
implementation starting before W03-E01 reaches `accepted`.

## Residual-risk expectations

RISK-W03-003 is expected to reduce to low-medium residual risk once the deferred-link framing for
T4's cache-invalidation sub-criterion is honored (tracked explicitly, not silently dropped).
RISK-W03-E04-002 is expected to reduce to low residual risk provided the hard W03-E01-acceptance gate
is genuinely honored before this story's implementation work begins.

## Plan

See `plan.md`.
