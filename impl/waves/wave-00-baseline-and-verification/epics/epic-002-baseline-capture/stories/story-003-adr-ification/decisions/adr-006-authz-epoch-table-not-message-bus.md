---
id: ADR-W00-E02-S003-006
type: decision
title: Per-tenant authz_epoch table, polled on the existing authz read path; LISTEN/NOTIFY optional only
status: ratified
context: SEC-04 cross-pod authz cache invalidation transport — LISTEN/NOTIFY vs an epoch-poll design, and whether a new message bus is needed
date: 2026-07-12
deciders:
  - Fable 5 (framework architecture lead role)
related_source_items:
  - D-06
  - W05-E04
created_at: 2026-07-12
updated_at: 2026-07-12
derived: false
---

# ADR-W00-E02-S003-006 — Per-tenant authz_epoch table, polled on the existing authz read path; LISTEN/NOTIFY optional only

**Formalization note:** This ADR formalizes a decision Fable 5 already made in
`docs/implementation/fable5-final-architecture-review-2026-07-11.md` §F/§U; this ADR is the
programme's own durable record of it, not a new decision-making act.

## Decision ID

ADR-W00-E02-S003-006.

## Title

Per-tenant authz_epoch table, polled on the existing authz read path; LISTEN/NOTIFY optional only.

## Status

ratified — the underlying decision was already made by Fable 5 in REVIEW §F row 7 (Q7); this ADR
file's own creation/registration is tracked separately by task W00-E02-S003-T002's own `status:
todo`→`done` lifecycle (see `../story.md` "Status discipline").

## Context

REVIEW §F question 7 asks: SEC-04 needs to bound authz cache staleness across multiple pods/
processes sharing a tenant's authorization state — when one pod invalidates its local cache (e.g.
on a permission change), how do other pods learn to invalidate theirs? The two transports
considered are Postgres `LISTEN/NOTIFY` (push-based) and a polled epoch counter (pull-based).

## Options considered

- **A new message bus in the kernel** (a general-purpose pub/sub mechanism for cross-pod
  invalidation signaling) — rejected. REVIEW §F row 7, quoted verbatim: "Avoids a new message bus
  in the kernel." This is also consistent with REVIEW §M's broader rejected-dependency stance
  (`requirement-inventory.md` §C, `M-REJ` row: "Rejected deps (viper/envconfig, kernel message bus,
  custom crypto)").
- **Postgres `LISTEN/NOTIFY` as the sole/primary transport** — not rejected outright, but
  downgraded: retained only as an *optional* latency optimization, not the correctness mechanism.
- **Per-tenant epoch integer in a small `authz_epoch` table, polled on the existing authz read
  path** — chosen as the correctness mechanism. See Decision below.

## Decision

**Per-tenant epoch integer in a small `authz_epoch` table, polled on the existing authz read path;
Postgres `LISTEN/NOTIFY` as an optional latency optimisation, not a correctness dependency — avoids
a new message bus in the kernel.** (REVIEW §F row 7, quoted verbatim.)

### Safe default

No distinct safe-default stated beyond the decision itself — REVIEW §F row 7 states an
unconditional resolution ("resolved"), not a recommendation with a separate fallback path.

## Rationale

REVIEW §F row 7, quoted verbatim (classification): "Fable 5 decision (concurrency)." Polling an
epoch counter on a path the authz system already reads (rather than standing up a new push-based
notification mechanism as the load-bearing correctness path) keeps correctness dependent only on
the database the framework already requires, not on a new infrastructure component. `LISTEN/NOTIFY`
can still be layered on top later purely to reduce the *latency* between an invalidation and other
pods noticing it, but the system remains correct even if `LISTEN/NOTIFY` delivery is missed or
unavailable, because the poll against `authz_epoch` is the actual source of truth.

## Consequences

- SEC-04 (W05-E04, per `requirement-inventory.md`'s SEC-04 row: "CS-17: LRU (approved dep) + epoch
  table (D-06); P0 if cache prod-enabled") implements the `authz_epoch` table and epoch-poll logic
  as the cross-pod cache-invalidation correctness mechanism.
- No new kernel dependency (message-bus client) is introduced by this decision.
- `LISTEN/NOTIFY` remains available as a future, purely-optional latency optimization layered on
  top of the epoch-poll design — its absence or failure does not compromise correctness, only
  invalidation latency.
- Per REVIEW §F row 7's blocks column, this whole finding is "P1, not on the critical path" for
  build-blocking purposes, though per `requirement-inventory.md`'s SEC-04 row it becomes P0 "if
  cache prod-enabled" — this ADR's decision does not itself
  change that priority classification.

## Related source items

D-06; downstream epic W05-E04 (SEC-04) — unblocked by this ADR per
`../../../../dependencies.md` and `../story.md` "Dependencies."

## Date

2026-07-12.

## Deciders

Fable 5 (concurrency decision).
