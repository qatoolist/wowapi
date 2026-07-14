---
id: ADR-W00-E02-S003-002
type: decision
title: One generic owner-bound Registrar type with per-subsystem typed keys
status: ratified
context: AR-01 Registrar type design — one shared type vs a per-subsystem registrar type per subsystem?
date: 2026-07-12
deciders:
  - Fable 5 (framework architecture lead role)
related_source_items:
  - D-02
  - W05-E02
created_at: 2026-07-12
updated_at: 2026-07-12
derived: false
---

# ADR-W00-E02-S003-002 — One generic owner-bound Registrar type with per-subsystem typed keys

**Formalization note:** This ADR formalizes a decision Fable 5 already made in
`docs/implementation/fable5-final-architecture-review-2026-07-11.md` §F/§U; this ADR is the
programme's own durable record of it, not a new decision-making act.

## Decision ID

ADR-W00-E02-S003-002.

## Title

One generic owner-bound Registrar type with per-subsystem typed keys.

## Status

ratified — the underlying decision was already made by Fable 5 in REVIEW §F row 3 (Q3); this ADR
file's own creation/registration is tracked separately by task W00-E02-S003-T001's own `status:
todo`→`done` lifecycle (see `../story.md` "Status discipline").

## Context

REVIEW §F question 3 asks: should AR-01's `Registrar` capability type be a single generic type
shared across the framework's subsystems, or should each subsystem define its own registrar type?
This is a public-contract decision — it fixes a shape that AR-01 (application-model, W05-E01) and
AR-02 (typed port keys, W05-E02) implement directly.

## Options considered

- **Per-subsystem registrar types** (one distinct `Registrar` type per subsystem) — rejected, per
  the question framing itself ("one shared vs per-subsystem"; the chosen option is explicitly the
  "one" side, meaning "per-subsystem" is the rejected alternative).
- **One generic owner-bound `Registrar` capability type, with per-subsystem typed keys (`Key[T]`)**
  — chosen. See Decision below.

## Decision

**One generic owner-bound `Registrar` capability type, with per-subsystem *typed keys* (`Key[T]`)
rather than per-subsystem registrar types.** (REVIEW §F row 3, quoted verbatim.)

### Safe default

No distinct safe-default stated beyond the decision itself — REVIEW §F row 3 states an
unconditional resolution ("resolved"), not a recommendation with a separate fallback path.

## Rationale

REVIEW §F row 3, quoted verbatim: "Capability confusion is prevented by the key's phantom type +
owner binding, not by multiplying registrar types." That is, the safety property the
per-subsystem-type alternative would have provided (subsystem A cannot accidentally register into
subsystem B's registrar) is achieved instead by the type parameter on `Key[T]` plus owner binding —
so the safety goal is met without multiplying the number of distinct registrar types the framework
must define and maintain.

## Consequences

- AR-01 (W05-E01) and AR-02 (W05-E02) implement one `Registrar` type as their public contract,
  parameterized by typed keys, rather than a family of subsystem-specific registrar types.
- The framework's public API surface for capability registration stays small (one type) while still
  preventing cross-subsystem capability confusion, via the type system rather than via a
  proliferation of nominal types.
- Downstream subsystems each define their own `Key[T]` types (phantom-typed per subsystem) rather
  than their own `Registrar` types — this is the concrete design constraint AR-02's implementation
  work (typed port keys + compiled provider graph) inherits from this ADR.

## Related source items

D-02; downstream epic W05-E02 (AR-02) — unblocked by this ADR per
`../../../../dependencies.md` and `../story.md` "Dependencies." (W05-E01/AR-01 also implements this
decision directly, per `requirement-inventory.md`'s AR-01 row note: "D-02/D-03 resolve its open
questions.")

## Date

2026-07-12.

## Deciders

Fable 5 (framework, public contract).
