---
id: W03-E04
type: epic
title: Relationship semantics
status: accepted
wave: W03
owner: unassigned
reviewer: unassigned
priority: high
created_at: 2026-07-12
updated_at: 2026-07-12
source_requirements:
  - DATA-07
depends_on:
  - W03-E01
stories:
  - W03-E04-S001
decisions: []
risks:
  - RISK-W03-003
---

# W03-E04 — Relationship semantics

## Epic objective

Complete `Checker.Has`'s relationship evaluation for party-subject edges (today, per the code's own
comment, "not consulted yet") and every schema-enumerated `subject_kind`, and source real actor
attribution for `Relate`/mirror `Upsert` by reusing DATA-06 T2's mechanism directly rather than
reimplementing it. This is PLAN §5.3's DATA-07, scoped to T1, T2, and T4 per
`impl/analysis/wave-allocation-detail.md` ("T1, T2, T4 (T3 consumed from DATA-06 T2)").

## Problem being solved

`requirement-inventory.md` row DATA-07 records: "Relationship semantics + actor attribution (T1–T4)"
— class IMPL, priority P1, disposition "blocked→planned," target `W03-E04-S001`, notes "HARD dep
SEC-01; secondary SEC-04." PLAN §5.3's own evidence: "`Checker.Has`
(`kernel/relationship/relationship.go:42-66`) filters `subject_kind='capacity'` only — party-subject
edges are, per the code's own comment, 'not consulted yet.' Same nil-actor gap as DATA-06, same
file." This epic closes both gaps: it extends `Checker.Has` to evaluate party-subject edges and every
enumerated `subject_kind` (not just `capacity`), and it sources real actor attribution for
relationship mutations — but the actor-attribution mechanism itself is not reimplemented here, it is
DATA-06 T2's mechanism, consumed directly.

**This epic has a hard, blocking dependency on W03-E01's acceptance.** PLAN's own words, verbatim:
"Hard dependency on PF-SEC's SEC-01 — do not schedule before it lands." T1's own resolution path
("Resolve actor → active capacity → optional party through the authoritative principal model")
requires the authoritative principal model SEC-01 builds — this epic cannot meaningfully implement
T1 against the pre-SEC-01 claim-trusting `Actor` shape.

## Scope

- T1 — resolve actor → active capacity → optional party through the authoritative principal model
  (the model W03-E01 establishes); `Checker.Has` can evaluate party-subject edges.
- T2 — extend `Checker.Has` to cover every schema-enumerated `subject_kind`; unsupported kinds fail
  closed.
- T4 — every authorization-input mutation is ownership-checked, attributed, audited, versioned, and
  invalidates relevant caches. The cache-invalidation portion of this acceptance criterion is
  deferred-linked to W05-E04-S002 (SEC-04's epoch table, D-06) — see "Dependencies."

## Out of scope

- T3 (source real actor for `Relate`/mirror `Upsert`) — per PLAN's own instruction: "reuse DATA-06
  T2's mechanism directly, same file, do not reimplement." T3 is DATA-06 T2's shared fix surface
  (W02-E04-S001, another worker's scope); this epic's T4 *consumes* that mechanism once it exists,
  it does not reimplement or re-plan it. See "Dependencies" for the cross-reference.
- SEC-01 itself (W03-E01) — a hard dependency, not this epic's own implementation scope.
- SEC-04's cache-bounding/epoch-table work (W05-E04-S002) — this epic's T4 cache-invalidation
  acceptance criterion is deferred-linked to it, but the epoch table itself is built in W05, not
  here.

## Source requirements

DATA-07 (T1, T2, T4 — T3 excluded per the cross-reference above).

## Architectural context

`kernel/relationship`'s `Checker.Has` today only evaluates `subject_kind='capacity'` edges — this is
a genuine functional gap, not merely an untested one: party-subject edges exist in the schema but
are never consulted. This epic completes that evaluation logic on top of the principal model W03-E01
establishes (hence the hard dependency: a party-subject edge resolution needs "actor → active
capacity → optional party" resolution through an *authoritative* principal model, which does not
exist in a trustworthy form before SEC-01 lands). The actor-attribution portion (T3, cross-referenced
but not reimplemented here) shares its exact fix surface with DATA-06 T2 in `registrar_pg.go` — PLAN
is explicit that this is "High duplication risk if staffed independently — sequence as one shared
task," which is why the task ownership sits entirely with DATA-06 (W02-E04-S001), not duplicated
into this epic.

The affected layer is `kernel/relationship/relationship.go` (`Checker.Has`) and its interaction with
the SEC-01 principal-resolution surface (`kernel/auth`) and, for T4's cache-invalidation portion, a
future authz-cache interface from W05-E04.

## Included stories

- **W03-E04-S001 — relationship-semantics** (DATA-07 T1, T2, T4, single story per
  `impl/analysis/wave-allocation-detail.md`: "S001 T1, T2, T4 (T3 consumed from DATA-06 T2 — DATA-06
  is W02 scope, cross-reference only, do not reimplement; hard dep W03-E01 accepted; SEC-04 epoch dep
  noted soft — cache-invalidation AC deferred-linked to W05-E04-S002)").

## Dependencies

**Hard**: W03-E01 must be **accepted**, not merely started, before this epic's implementation work
begins (PLAN's own "do not schedule before it lands" language). **Soft/cross-reference**: DATA-06 T2
(W02-E04-S001) — this epic's T4 consumes DATA-06 T2's actor-attribution mechanism in
`registrar_pg.go`, it does not reimplement it; if DATA-06 T2 has not landed by the time this epic's
T4 is ready, T4 is blocked on that specific dependency, tracked explicitly rather than worked around
by an independent reimplementation. **Soft/deferred-link**: W05-E04-S002 (SEC-04's epoch table,
D-06) — T4's cache-invalidation acceptance criterion is deferred-linked to it per PLAN's own
cross-cutting note: "do not assume PF-SEC delivers on PF-DATA's timeline" (inverted at this epic's
scope: do not assume W05 delivers the epoch table on W03's timeline).

## Risks

RISK-W03-003 (DATA-07 T4's cache-invalidation AC depends on W05-E04-S002, which may not land on
W03's timeline) — inherited from `../../risks.md` (wave-level); see `risks.md` (epic-level) for
elaboration.

## Required decisions

None new. This epic's T1 depends on SEC-01's authoritative principal model but makes no new
architecture decision of its own.

## Epic acceptance criteria

- **AC-W03-E04-01**: `Checker.Has` resolves actor → active capacity → optional party through the
  post-SEC-01 authoritative principal model; a seeded party-subject edge, when resolved through an
  actor carrying a party, is evaluated as `true` where it was previously (incorrectly) `false`.
- **AC-W03-E04-02**: Every schema-enumerated `subject_kind` has an evaluation branch in
  `Checker.Has`; an unsupported/unenumerated kind fails closed (not silently `true` or silently
  ignored).
- **AC-W03-E04-03**: Every authorization-input mutation (edge create/revoke) is ownership-checked,
  attributed (via DATA-06 T2's mechanism, consumed not reimplemented), audited, and versioned; the
  cache-invalidation portion is explicitly deferred-linked to W05-E04-S002, not silently dropped nor
  silently assumed complete.
- **AC-W03-E04-04**: The story has passed independent review per mandate §14, with specific
  confirmation T3's scope was correctly cross-referenced to DATA-06, not reimplemented.

## Closure conditions

W03-E04-S001 reaches `accepted`; AC-W03-E04-01 through AC-W03-E04-04 above are all satisfied (with
AC-W03-E04-03's cache-invalidation sub-criterion permitted to close as "deferred-linked, tracked" per
the epic's own documented deferral); `closure-report.md` for this epic is completed with reviewer
conclusion and acceptance date; this epic must not begin substantive implementation before W03-E01's
`closure-report.md` records `accepted`.

## Status update (2026-07-16)

`status: accepted` — W03-E04-S001 accepted.

— dated 2026-07-16, conductor adjudication (Fable 5), per review-gate-2026-07-16.md records
