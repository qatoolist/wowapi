---
id: W04-E01-S001
type: story
title: Shared lease/fencing primitive
status: accepted
wave: W04
epic: W04-E01
owner: unassigned
reviewer: unassigned
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-13
source_requirements:
  - DATA-02
depends_on: []
blocks:
  - W04-E01-S002
  - W04-E01-S003
acceptance_criteria:
  - AC-W04-E01-S001-01
  - AC-W04-E01-S001-02
  - AC-W04-E01-S001-03
artifacts: []
evidence: []
decisions: []
risks:
  - RISK-W04-001
  - RISK-W04-E01-001
---

# W04-E01-S001 — Shared lease/fencing primitive

## Story ID

W04-E01-S001

## Title

Shared lease/fencing primitive

## Objective

Design and implement a shared lease/fencing primitive — `lease_token`, monotonic
`lease_generation`, `lease_expires_at`, and an optional heartbeat — as a single, reusable kernel
building block, and execute the planned migration of W02-E01-S002's interim checkpoint lease onto
this primitive.

## Value to the framework

This story is, in PLAN's own words, **"the single highest-leverage build in this package — staff
and design-review it first"** (PF-DATA cross-cutting note (1)). PLAN DATA-02 T1's own acceptance
criterion states the primitive must be reused "≥3 times, not three independent copies" and
classifies it as "architecturally load-bearing across all three findings" (DATA-02, DATA-03,
DATA-04). Nothing else in this epic, in W04-E02 (remote-io-outside-tx), or in W04-E03
(bulk-multi-worker-safety) can begin its own fencing work until this primitive exists — S002's
fenced `jobs_queue` claim/finalize/reclaim (this epic), W04-E02's three-stage notify/webhook
protocol, and W04-E03's leased `SKIP LOCKED` bulk claim path all consume this same primitive rather
than inventing their own lease/fencing type. This is the wave's keystone story in the same sense
that W02-E01-S001 (manifest-and-lock-budget) was W02's foundation story: everything downstream in
this epic and two sibling epics is gated on this one build landing correctly, once, as a genuinely
shared kernel type — not three parallel, subtly-divergent lease implementations.

This story also carries a second, explicitly planned responsibility: it **supersedes**
W02-E01-S002's interim checkpoint lease. W02-E01-S002 built a minimal checkpoint-lease mechanism
(checkpoint token + resumability only, no job-claim fencing or heartbeat) as a bounded interim
substitute, precisely because this primitive did not exist yet when DATA-09 T4's backfill harness
needed checkpoint safety. That interim lease is recorded — in W02's own `risks.md` as
RISK-W02-001, and in this wave's `risks.md` as the mirrored receiving-side risk RISK-W04-001 — as a
genuine, planned technical-debt-bearing deviation, not a permanent fork. This story is the landing
point for that planned transition: it must read any migration-checkpoint state written under
W02-E01-S002's interim lease and correctly re-express it under this primitive's schema, before the
interim lease code path is removed.

## Problem statement

`requirement-inventory.md` row DATA-02 states: "Lease generations/fencing + idempotency (T1–T7) |
IMPL | P0 | planned | W04-E01-S001..S003 | T1 shared primitive is keystone." PLAN DATA-02's evidence
for the epic as a whole is the confirmed race this primitive exists to close: "claim SQL returns no
lease token/generation; completion/failure match only `id`; `ReclaimStalled` blind-resets every
stale row with no per-row fencing check. Confirmed race: A stalls, gets reclaimed by B, B completes,
A's eventual finalize silently overwrites B's outcome." T1's own row states the acceptance bar
directly: "Implement a **shared** lease/fencing primitive (`lease_token`, monotonic
`lease_generation`, `lease_expires_at`, optional heartbeat) as a reusable kernel building block for
DATA-02/03/04 | — | One primitive reused ≥3 times, not three independent copies | Unit tests on
token/generation comparison | `DATA-02/lease-primitive/` | Architecturally load-bearing across all
three findings." No lease/fencing primitive of any kind exists anywhere in the repository today —
not the general kernel type this story builds, and (until W02-E01-S002 landed) not even a
narrow-purpose interim substitute.

## Source requirements

DATA-02 (T1).

## Current-state assessment

Per PLAN's own DATA-02 evidence (to be re-confirmed at this story's own execution commit): no
lease/fencing primitive exists in `kernel/jobs` or anywhere else in the repository. `jobs_queue`'s
claim SQL returns no lease token or generation; completion and failure paths match only `id`;
`ReclaimStalled` performs a blind reset of every stale row with no per-row fencing check. The one
adjacent precedent is W02-E01-S002's interim checkpoint lease — a narrow, bounded substitute built
specifically for DATA-09 T4's backfill-harness checkpoint safety, explicitly scoped to exclude
fencing generations for job claims and heartbeats (per W02-E01-S002's own `plan.md` and
RISK-W02-001's mitigation description) — which is not a prior version of this primitive to extend,
but a different, narrower mechanism this story must migrate away from, not build on top of. This
story's own re-confirmation step (per this programme's fail-first convention applied elsewhere,
e.g. W02-E01-S001) is to read `kernel/jobs`'s claim/finalize/reclaim SQL and W02-E01-S002's interim
lease implementation at this story's actual start commit and confirm both facts still hold before
building the shared primitive.

## Desired state

A single, reusable kernel-level lease/fencing type exists, exposing `lease_token` (opaque,
unguessable), a monotonically increasing `lease_generation`, `lease_expires_at`, and an optional
heartbeat mechanism, with unit-tested token/generation comparison semantics. This epic's own S002
consumes it to fence `jobs_queue`; the primitive's field set and semantics are validated against
DATA-03's (W04-E02) and DATA-04's (W04-E03) own stated needs before being treated as locked, per
RISK-W04-E01-001's mitigation, so that those epics can consume the same type rather than each
building their own. Any migration-checkpoint state written under W02-E01-S002's interim lease has
been read, correctly re-expressed under this primitive's schema, and the interim lease code path
has been removed, with a test proving no in-flight backfill checkpoint state was lost or duplicated
across the cutover.

## Scope

- Designing and implementing the shared lease/fencing primitive: `lease_token`,
  `lease_generation`, `lease_expires_at`, optional heartbeat, plus token/generation comparison
  semantics.
- Locating the primitive as a genuine kernel building block (exact package location TBD, see
  `plan.md`'s "Unresolved questions") — not embedded inside `kernel/jobs` as a jobs-only type.
- Unit tests on token/generation comparison semantics, per PLAN T1's own "Tests" column.
- Validating the primitive's field set against DATA-03's and DATA-04's own stated needs (not just
  DATA-02's), per RISK-W04-E01-001's mitigation, before treating the design as locked.
- The planned migration of W02-E01-S002's interim checkpoint lease onto this primitive: reading any
  existing interim-lease checkpoint state, re-expressing it under this primitive's schema, and
  removing the interim lease code path, with a test proving no checkpoint state is lost or
  duplicated across the cutover.

## Out of scope

- **Applying the primitive to `jobs_queue`'s claim/finalize/reclaim SQL** — W04-E01-S002's scope.
  This story produces the primitive; S002 consumes it.
- **Applying the primitive to `kernel/notify`/`kernel/webhook` or to bulk multi-worker processing**
  — W04-E02's and W04-E03's scope respectively. This story validates the primitive's field set
  against their stated needs (so the design does not have to be retrofitted later) but does not
  itself implement either consumer.
- **The worker idempotency-declaration contract (inbox/effect ledger, domain CAS, provider
  idempotency key) and the named chaos test** — W04-E01-S003's scope (PLAN DATA-02 T5, T7).
- **Building a new backfill harness or altering DATA-09's expand/backfill/validate protocol beyond
  the checkpoint-lease migration itself** — W02-E01-S002's own scope remains as built; this story
  changes only the lease mechanism underneath it, not the backfill harness's external behavior.

## Assumptions

- The primitive's exact package location (a new `kernel/lease` package, a subpackage of
  `kernel/jobs`, or elsewhere) is not determined by the source documents beyond "a reusable kernel
  building block" — this is a genuinely open design question recorded as an unresolved question in
  `plan.md`, not invented here.
- The heartbeat mechanism is explicitly "optional" per PLAN T1's own acceptance criterion wording
  — this story's plan records whether S002/S003's own needs require it to be exercised now or left
  as a documented extension point, per `plan.md`'s "Unresolved questions."
- The interim-lease migration's exact mechanics (read-then-translate-then-remove vs. a
  dual-write transition window) are not specified by the source beyond RISK-W04-001's own
  mitigation description ("an explicit migration step (not a big-bang cutover)") — the specific
  mechanism is an implementation-time decision recorded in `plan.md`, not pre-answered here.

## Dependencies

None within W04-E01 (this is the epic's first story). Depends on W04's own entry gate (W00's exit
criteria) at wave scope — DATA-02 has no dependency on W02's online-migration protocol. Blocks
W04-E01-S002 (jobs-queue fencing consumes this story's primitive) and W04-E01-S003 (the idempotency
contract and chaos harness both operate on the fenced chain S002 builds atop this story's
primitive). Also the landing point for the planned supersession of W02-E01-S002's interim checkpoint
lease (RISK-W04-001, mirroring RISK-W02-001).

## Affected packages or components

New: the shared lease/fencing primitive package (exact location TBD, per `plan.md`'s "Unresolved
questions" — expected under a new `kernel/lease`-style package or equivalent). Modified: W02-E01-S002's
interim checkpoint-lease code path (migrated onto the new primitive, then removed).

## Compatibility considerations

The interim-lease migration is itself a compatibility-sensitive operation: any migration checkpoint
state already written under W02-E01-S002's interim lease format (in production, staging, or test
fixtures) must remain correctly interpretable across the cutover. Per RISK-W04-001's mitigation,
this story's plan must define an explicit migration step, not a big-bang cutover that risks losing
or misinterpreting in-flight checkpoint state.

## Security considerations

None distinct to this story beyond the general correctness of the fencing primitive itself — an
incorrect token/generation comparison would undermine the entire epic's security-relevant guarantee
(no stale-worker overwrite), so the unit tests on comparison semantics (PLAN T1's own "Tests"
column) are load-bearing, not incidental.

## Performance considerations

None separately mandated by the source. The primitive's own token/generation fields are simple
scalar comparisons; no performance-sensitive design decision beyond avoiding an unnecessarily heavy
heartbeat mechanism if S002/S003's needs do not require one (see "Assumptions").

## Observability considerations

None separately mandated by the source for this story specifically (S002's fenced finalize/reclaim
paths are where observable fencing-rejection events matter operationally, per that story's own
scope).

## Migration considerations

This story's own "migration" is the planned interim-checkpoint-lease cutover described above under
"Desired state" and "Compatibility considerations" — not a database schema migration of its own
(the primitive itself does not require new persisted columns; `jobs_queue`'s own lease columns are
S002's scope).

## Documentation requirements

Document the primitive's field set, token/generation comparison semantics, and package location, so
that W04-E02 and W04-E03 can consume it without re-deriving its contract from source. Document the
interim-lease migration's mechanics and completion, so a future reader understands why
W02-E01-S002's interim lease existed and confirms it has been fully retired.

## Acceptance criteria

- **AC-W04-E01-S001-01**: The shared lease/fencing primitive (`lease_token`, monotonic
  `lease_generation`, `lease_expires_at`, optional heartbeat) is implemented as a single, reusable
  kernel building block, with unit tests proving correct token/generation comparison semantics.
- **AC-W04-E01-S001-02**: The primitive's field set and semantics have been validated against
  DATA-03's (W04-E02) and DATA-04's (W04-E03) own stated needs — not merely DATA-02's — before being
  treated as locked, per RISK-W04-E01-001's mitigation, so it can be reused "≥3 times, not three
  independent copies" per PLAN T1's own acceptance criterion.
- **AC-W04-E01-S001-03**: Any migration-checkpoint state written under W02-E01-S002's interim
  checkpoint lease has been correctly read and re-expressed under this primitive's schema, the
  interim lease code path has been removed, and a test proves no in-flight backfill checkpoint state
  was lost or duplicated across the cutover.

## Required artifacts

- The shared lease/fencing primitive (source-code package).
- The interim-checkpoint-lease migration tooling/code.
- Documentation of the primitive's contract and the completed migration.
See `artifacts/index.md`.

## Required evidence

- Unit-test output for token/generation comparison semantics.
- The cross-consumer field-set review record (DATA-03/DATA-04 needs validated against this
  primitive's design).
- Migration test output proving no checkpoint state lost or duplicated across the cutover.
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`: `story.md`
and `plan.md` complete, acceptance criteria numbered and measurable, dependencies (none within this
epic) recorded, owner/reviewer assignment pending, unresolved questions (package location, heartbeat
exercise scope, migration mechanics) explicitly recorded rather than silently assumed.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; all three acceptance
criteria verified with evidence in `evidence/index.md`; `closure.md` completed; independent review
passed per mandate §14, specifically confirming the interim-checkpoint-lease migration genuinely
occurred (AC-W04-E01-S001-03) and the cross-consumer field-set review genuinely occurred
(AC-W04-E01-S001-02), not merely claimed.

## Risks

RISK-W04-001 (the interim-checkpoint-lease migration carries a correctness risk on the receiving
side — mirrors W02's own RISK-W02-001) and RISK-W04-E01-001 (the primitive, once locked, is
load-bearing across three epics at once — an under-specified design is costly to retrofit
post-consumption) — see epic-level `risks.md` for full detail and mitigation/contingency.

## Residual-risk expectations

Once the cross-consumer field-set review (AC-W04-E01-S001-02) and the interim-lease migration
(AC-W04-E01-S001-03) are executed as planned, residual risk is expected to be low — this is a
foundational but well-bounded design-and-migration story with a clear, source-derived acceptance
bar and an already-documented mitigation path for both of its named risks.

## Plan

See `plan.md`.
