---
id: W04-E02-S001
type: story
title: Notify and webhook three-stage remote-I/O protocol
status: accepted
wave: W04
epic: W04-E02
owner: unassigned
reviewer: unassigned
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-13
source_requirements:
  - DATA-03
depends_on:
  - W04-E01-S001
blocks:
  - W04-E02-S002
acceptance_criteria:
  - AC-W04-E02-S001-01
  - AC-W04-E02-S001-02
  - AC-W04-E02-S001-03
artifacts: []
evidence: []
decisions: []
risks:
  - RISK-W04-E02-S001-001
---

# W04-E02-S001 — Notify and webhook three-stage remote-I/O protocol

## Story ID

W04-E02-S001

## Title

Notify and webhook three-stage remote-I/O protocol

## Objective

Reuse W04-E01's shared lease/fencing primitive for `kernel/notify` and `kernel/webhook` claim rows
(not a bespoke copy), and implement a three-stage claim-tx → effect-outside-tx → fenced-finalize-tx
protocol for both `kernel/notify` and `kernel/webhook.deliverToEndpoint`, so that no remote
network/secret I/O call executes while a database transaction is open.

## Value to the framework

This story closes the specific defect wowapi's own source code already documents against itself:
`notify/service.go:456-586`'s own comment (446-449) states "Real production deployments should move
the network call outside the tx." `webhook/service.go`'s delivery loop and secret resolution both
run inside `plat.WithTenant(...)` today — meaning a slow or hung remote call currently holds a
database transaction open for its full duration, and a rollback/retry after a partial remote effect
can duplicate that effect with no protection. This story is the epic's foundation: T4's inbound
two-phase verification (W04-E02-S002) and T8's 6-boundary chaos test both assume T2/T3's three-stage
protocol already exists as the thing being chaos-tested.

## Problem statement

`requirement-inventory.md` row DATA-03 states: "Remote I/O outside DB transactions (T1–T8) | IMPL |
P0 | planned | W04-E02-S001..S002 | Scope refined by MATRIX CS-11: external effects only; T7 =
DATA-08 W0-T2 duplicate (done)." This story's source task rows, verbatim:

- **T1**: "Reuse DATA-02's shared lease primitive for notify/webhook claim rows | Depends on:
  DATA-02 T1 | Acceptance: Lease columns via shared primitive, not a bespoke copy | Evidence:
  `DATA-03/lease-columns/` | Risk: None beyond DATA-02's own risk."
- **T2**: "Three-stage protocol for `kernel/notify`: claim-tx (assigns lease) → `sender.Send`
  outside any tx, delivery ID as idempotency key → finalize-tx comparing lease token | Depends on:
  T1 | Acceptance: No `sender.Send` call while a DB tx is open | Tests: See T8 boundary matrix |
  Evidence: `DATA-03/notify/` | Risk: Delete/update the self-documented 'should move outside tx'
  comment as part of this task."
- **T3**: "Same three-stage protocol for `kernel/webhook.deliverToEndpoint` | Depends on: T1 |
  Acceptance: No DNS/secret-resolve/POST call while a tx is open | Tests: See T8 boundary matrix |
  Evidence: `DATA-03/webhook/` | Risk: Current-row-state check must move into claim stage so
  Execute needs no mid-flight DB reads."

## Source requirements

DATA-03 (T1, T2, T3).

## Current-state assessment

Per the source evidence, self-documented in wowapi's own code: `notify/service.go:456-586`'s own
comment (446-449) already states "Real production deployments should move the network call outside
the tx" — this is a confirmed, self-acknowledged defect, not a hypothesis this story must first
establish. `webhook/service.go`'s delivery loop and secret resolution both run inside
`plat.WithTenant(...)`, per the same source evidence. This story's own re-confirmation step, per
this programme's fail-first convention, is to re-read `notify/service.go:446-586` and
`webhook/service.go`'s delivery/secret-resolution code paths at this story's actual start commit
and confirm the network calls are still inside the transaction boundary before restructuring them.

## Desired state

`kernel/notify` and `kernel/webhook.deliverToEndpoint` both claim their delivery row inside a short
transaction that assigns a lease (via W04-E01's shared primitive), perform the remote call
(`sender.Send` for notify; DNS resolution, secret resolution, and the POST for webhook) entirely
outside any open transaction, and finalize in a second short transaction that compares the lease
token before committing the outcome. Delivery ID is used as the idempotency key for the effect
stage. The self-documented "should move outside tx" comment at `notify/service.go:456-586`
(446-449) is deleted or updated to reflect that this is now the implemented protocol, not a TODO.

## Scope

- Migrating notify/webhook claim rows onto W04-E01's shared lease/fencing primitive's columns
  (`lease_token`, `lease_generation`, `lease_expires_at`), not a bespoke copy (T1).
- The three-stage claim-tx → effect-outside-tx → fenced-finalize-tx protocol for `kernel/notify`,
  using delivery ID as the effect stage's idempotency key (T2).
- The same three-stage protocol for `kernel/webhook.deliverToEndpoint`, with the current-row-state
  check moved into the claim stage so the effect stage needs no mid-flight DB reads (T3).
- Deleting or updating the self-documented "should move outside tx" comment at
  `notify/service.go:456-586` (446-449) as part of T2's own change.

## Out of scope

- Inbound webhook verification's two-phase protocol (T4), the failed-signature audit path (T5), the
  per-adapter idempotency-contract declaration (T6), and the named 6-boundary chaos test (T8) — all
  W04-E02-S002's scope. This story builds the outbound three-stage protocol those later tasks verify
  and extend; it does not itself implement inbound verification or the chaos test.
- T7 (stale "app_platform lacks INSERT on events_outbox" comment removal and legal-delivery audit
  wiring) — not this story's scope; already executed under DATA-08 W0-T2, cross-referenced in
  W04-E02-S002's story.md, not implemented anywhere in this epic.
- W04-E01's shared lease/fencing primitive itself — this story consumes it; it does not build or
  extend it.
- FBL-04's retry-library adoption — W04-E02-S003's scope, independently sequenced.

## Assumptions

- W04-E01-S001 (the shared lease/fencing primitive) is assumed to expose a reusable claim/lease/
  finalize API surface generic enough for notify/webhook to consume without notify/webhook-specific
  modification to the primitive itself — this is the explicit design intent of W04-E01-S001 per
  DATA-02 T1's own acceptance criterion ("One primitive reused ≥3 times," per `wave.md`), but the
  exact call-site integration shape (how notify/webhook's claim-tx invokes the primitive) is not yet
  determined by the source and is recorded as an implementation-time design question in `plan.md`.
- The exact bounded-retry/finalize-token-comparison mechanics mirror W04-E01's own fencing
  semantics (comparing a lease token at finalize time to detect a stale/reclaimed worker) — this is
  confirmed by DATA-03 T2's own acceptance criterion structure ("finalize-tx comparing lease
  token"), not invented by this story.
- Delivery ID is confirmed, not assumed, as the idempotency key for the effect stage — DATA-03 T2's
  own acceptance criterion states this explicitly ("delivery ID as idempotency key").

## Dependencies

Depends on **W04-E01-S001** (the shared lease/fencing primitive) — T1's own dependency column
states "DATA-02 T1." No dependency within this epic (this is the epic's first story). Blocks
W04-E02-S002 (T4/T5/T6/T8 all build on or verify this story's three-stage protocol).

## Affected packages or components

`kernel/notify` (specifically `notify/service.go:446-586`'s claim/send/finalize logic), `kernel/
webhook` (specifically `webhook/service.go`'s `deliverToEndpoint` function and its delivery loop /
secret resolution). New: a claim-row schema migration adding the shared primitive's lease columns
to notify/webhook's delivery-tracking tables (exact table names TBD at implementation time — see
`plan.md`).

## Compatibility considerations

The three-stage protocol changes notify/webhook's internal claim/send/finalize control flow but is
not expected to change either package's external API surface (the caller-facing `Send`/deliver
entry points) — this is a confirmed assumption pending implementation-time verification, recorded
in `plan.md`'s "Unresolved questions" if it proves incorrect.

## Security considerations

None beyond what the shared lease primitive itself already provides (fencing against a reclaimed
worker's stale finalize) — this story does not introduce a new security control beyond correctly
consuming that primitive.

## Performance considerations

Moving the remote call outside the transaction boundary is itself the performance/availability
control this story exists to provide — it prevents a slow or hung remote call from holding a
database transaction (and its locks) open for the call's full duration, which is the root cause
condition DATA-03 exists to close.

## Observability considerations

Claim, effect, and finalize stage transitions should be observable (logged, at minimum) so an
operator can distinguish "delivery succeeded on finalize" from "delivery is stuck between claim and
finalize" — a reasonable implementation-time addition consistent with W04-E01's own lease-primitive
observability expectations, not separately mandated by DATA-03's source text beyond the boundary
requirement itself.

## Migration considerations

The claim-row schema migration (adding the shared primitive's lease columns to notify/webhook's
delivery-tracking tables) is this story's only schema change. No data backfill is anticipated
beyond the new columns' default values, since notify/webhook delivery rows are transient
(delivery-cycle-scoped), not long-lived state requiring historical backfill — to be confirmed at
implementation time.

## Documentation requirements

Document the three-stage protocol's stage boundaries (what happens in claim-tx, what happens
outside any tx, what happens in finalize-tx) for both notify and webhook, so a future maintainer
does not reintroduce a remote call inside the transaction boundary. Document the deletion/update of
the `notify/service.go:446-586` self-referential comment and why it is now resolved rather than
still a TODO.

## Acceptance criteria

- **AC-W04-E02-S001-01**: `kernel/notify` and `kernel/webhook` claim rows use W04-E01's shared
  lease/fencing primitive's columns directly, not a bespoke copy — proven by a code-level
  inspection confirming the same primitive's schema/API is invoked, not a parallel implementation.
- **AC-W04-E02-S001-02**: No `sender.Send` call executes while a database transaction is open in
  `kernel/notify`'s claim → send → finalize flow, proven per the T8 boundary-matrix tests (executed
  in W04-E02-S002, referenced here as this story's own completion gate for T2's acceptance
  criterion). The self-documented "should move outside tx" comment at `notify/service.go:456-586`
  (446-449) is deleted or updated as part of this change.
- **AC-W04-E02-S001-03**: No DNS resolution, secret resolution, or POST call executes while a
  database transaction is open in `kernel/webhook.deliverToEndpoint`'s claim → deliver → finalize
  flow, proven per the T8 boundary-matrix tests; the current-row-state check is confirmed to occur
  in the claim stage, with no mid-flight DB read during the effect stage.

## Required artifacts

- The claim-row schema migration adding shared-primitive lease columns to notify/webhook's
  delivery-tracking tables.
- The three-stage protocol implementation for `kernel/notify`.
- The three-stage protocol implementation for `kernel/webhook.deliverToEndpoint`.
- Protocol-boundary documentation for both packages.
See `artifacts/index.md`.

## Required evidence

- Lease-column migration test output (`DATA-03/lease-columns/`).
- Notify three-stage protocol test output, specifically a no-send-while-tx-open assertion
  (`DATA-03/notify/`).
- Webhook three-stage protocol test output, specifically a no-network-call-while-tx-open assertion
  (`DATA-03/webhook/`).
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`: `story.md`
and `plan.md` complete, acceptance criteria numbered and measurable, dependency on W04-E01-S001
recorded and its status tracked, owner/reviewer assignment pending, unresolved questions (claim-row
schema integration shape, exact table names) explicitly recorded rather than silently assumed.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; all three acceptance
criteria verified with evidence in `evidence/index.md`; `closure.md` completed; independent review
passed per mandate §14, specifically confirming the self-documented comment was genuinely
deleted/updated (not merely claimed) and that the shared primitive was genuinely reused, not
copied.

## Risks

RISK-W04-E02-S001-001 (the claim-row schema migration touches live notify/webhook delivery-tracking
tables; an incorrect migration could disrupt in-flight deliveries) — see epic-level `risks.md` for
the epic-scoped W04-E01-dependency risk (RISK-W04-E02-002), which this story's own start depends on
resolving.

## Residual-risk expectations

Once W04-E01-S001 is accepted and this story's own migration/protocol work is independently
verified, residual risk is expected to be low — this is a well-bounded protocol-restructuring story
with a source-confirmed defect (the self-documented comment) as its own proof that the change is
correct and necessary.

## Plan

See `plan.md`.
