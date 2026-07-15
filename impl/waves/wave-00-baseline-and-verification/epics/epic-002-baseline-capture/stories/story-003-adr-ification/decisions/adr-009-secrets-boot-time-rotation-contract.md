---
id: ADR-W00-E02-S003-009
type: decision
title: "Secrets: boot-time-once resolution + restart-based rotation is the documented v1 contract"
status: ratified
context: Secrets rotation contract — hot-reload every secret consumer, embed a vault client in the kernel, or accept restart-based rotation as the v1 contract?
date: 2026-07-12
deciders:
  - Fable 5 (framework architecture lead role)
related_source_items:
  - D-09
  - W01
created_at: 2026-07-12
updated_at: 2026-07-12
derived: false
---

# ADR-W00-E02-S003-009 — Secrets: boot-time-once resolution + restart-based rotation is the documented v1 contract

**Formalization note:** This ADR formalizes a decision Fable 5 already made in
`docs/implementation/fable5-final-architecture-review-2026-07-11.md` §F/§U; this ADR is the
programme's own durable record of it, not a new decision-making act.

## Decision ID

ADR-W00-E02-S003-009.

## Title

Secrets: boot-time-once resolution + restart-based rotation is the documented v1 contract.

## Status

ratified — the underlying decision was already made by Fable 5 in REVIEW §U (cross-referenced from
MATRIX CS-25 in `requirement-inventory.md` §C); this ADR file's own creation/registration is
tracked separately by task W00-E02-S003-T003's own `status: todo`→`done` lifecycle (see
`../story.md` "Status discipline").

## Context

The framework must define a secrets-rotation contract: when a secret (e.g. a database credential or
API key) changes, how does the framework pick up the new value? Options range from full hot-reload
plumbing through every secret-consuming component, to accepting that a process restart is required,
to delegating secret storage/rotation entirely to an external vault system reachable via a client
embedded in the kernel.

## Options considered

- **Hot-reload plumbing through every secret consumer** — rejected for v1. MATRIX CS-25
  (reproduced in `../plan.md`'s D-09 mapping), elaborating REVIEW §U's condensed line: "most orchestrators roll pods on secret change; hot-reload
  plumbing through every consumer is real complexity with modest v1 payoff."
- **A vault client embedded in the kernel** — rejected outright (not merely deferred to a later
  version). REVIEW §U, quoted verbatim: "no vault client in the kernel."
- **Boot-time-once resolution + restart-based rotation, with a file-provider (K8s mounted-secret
  pattern) as a later, non-kernel-vault increment** — chosen. See Decision below.

## Decision

**Secrets: boot-time-once resolution + restart-based rotation is the documented v1 contract (most
orchestrators roll pods on secret change; hot-reload plumbing through every consumer is real
complexity with modest v1 payoff). File-provider (K8s mounted-secret pattern) is the next increment
when needed — NOT a vault client in the kernel.** (REVIEW §U, combining the condensed
decision-register line — "secrets: boot-time-once resolution + restart-based rotation is the
documented v1 contract; file-provider is the next increment, no vault client in the kernel" — with
the fuller phrasing from MATRIX CS-25, reproduced in `../plan.md`'s D-09 mapping, which elaborates
without contradicting the §U summary; both describe the same decision restated at two points.)

### Safe default

No distinct safe-default stated beyond the decision itself — REVIEW §U states a direct
recommendation-plus-rejection, not a recommendation with a separate fallback path.

## Rationale

Most container orchestrators (e.g. Kubernetes) already roll (restart) pods when a mounted secret or
referenced config changes, so a restart-based rotation contract piggybacks on infrastructure
behavior that already exists in typical deployments, rather than the framework having to
independently solve live secret hot-swapping. Building hot-reload plumbing through every secret
consumer is real, non-trivial engineering complexity, and REVIEW's own judgment is that this
complexity buys only "modest v1 payoff" relative to simply documenting the restart-based contract.
Embedding a vault client in the kernel is rejected outright — not deferred (REVIEW §U: "no vault
client in the kernel"; MATRIX CS-25: "*not* a vault client in the kernel"). Neither source states
a reason for that rejection. *Wave-00-added clarification, not source text:* a kernel-embedded
vault client would tie the kernel to a specific external secrets-management product — the kind of
vendor-specific dependency the kernel/adapter boundary discipline (mandate §2.3) keeps out of the
kernel — but this reasoning is this programme's own elaboration, not REVIEW's stated rationale.

## Consequences

- The framework documents boot-time-once secret resolution and restart-based rotation as the v1
  secrets contract (tracked as MATRIX CS-25 in `requirement-inventory.md` §C: "Secrets rotation
  contract (D-09) | DOC/OPS | planned | Restart-based rotation documented; file-provider = deferred
  (DEF-01)").
- No vault-client dependency is added to the kernel as a consequence of this decision.
- A file-provider (K8s mounted-secret pattern) remains an explicitly acknowledged future increment
  — tracked as a deferred item (`DEF-01` per `requirement-inventory.md` §C) — not implemented as
  part of this decision's v1 scope, and not to be conflated with the rejected vault-client option:
  the file-provider is a *deferred, later* capability, whereas the vault client is *rejected
  outright* for the kernel.
- Any secret consumer built against this contract must tolerate values being fixed for the lifetime
  of a process and only refreshed via restart — this is a documented behavioral constraint on
  downstream implementation, not merely an implementation detail.

## Related source items

D-09; downstream item W01 (secrets documentation, MATRIX CS-25) — unblocked by this ADR per
`../../../../dependencies.md` and `../story.md` "Dependencies."

## Date

2026-07-12.

## Deciders

Fable 5.
