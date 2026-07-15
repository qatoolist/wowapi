---
id: ADR-W00-E02-S003-008
type: decision
title: Thin in-kernel pgx.QueryTracer over the existing observability port; otelpgx rejected
status: ratified
context: pgx query tracing implementation approach — bind an existing third-party OTel bridge, or a thin in-kernel tracer over the framework's own observability port?
date: 2026-07-12
deciders:
  - Fable 5 (framework architecture lead role)
related_source_items:
  - D-08
  - W01-E02
created_at: 2026-07-12
updated_at: 2026-07-12
derived: false
---

# ADR-W00-E02-S003-008 — Thin in-kernel pgx.QueryTracer over the existing observability port; otelpgx rejected

**Formalization note:** This ADR formalizes a decision Fable 5 already made in
`docs/implementation/fable5-final-architecture-review-2026-07-11.md` §F/§U; this ADR is the
programme's own durable record of it, not a new decision-making act.

## Decision ID

ADR-W00-E02-S003-008.

## Title

Thin in-kernel `pgx.QueryTracer` over the existing observability port; `otelpgx` rejected.

## Status

ratified — the underlying decision was already made by Fable 5 in REVIEW §U (cross-referenced from
FBL-06 / MATRIX CS-05 in `requirement-inventory.md` §B); this ADR file's own creation/registration
is tracked separately by task W00-E02-S003-T003's own `status: todo`→`done` lifecycle (see
`../story.md` "Status discipline").

## Context

`kernel/database`'s pgx integration needs query tracing (for trace/log correlation, per FBL-06,
`requirement-inventory.md` §B: "OTel trace/log correlation + pgx tracer (D-08)"). The framework
already exposes an internal observability `Tracer` port used elsewhere in the kernel. The question
is how to wire pgx's own `pgx.QueryTracer` interface into that observability system: adopt the
existing third-party `otelpgx` bridge library, or implement a small, purpose-built
`pgx.QueryTracer` inside the kernel that calls the existing `Tracer` port directly.

## Options considered

- **`otelpgx`** (a third-party OTel-to-pgx tracing bridge library) — rejected. REVIEW §U, quoted
  verbatim: "`otelpgx` rejected to keep vendor types out of `kernel/database`." (The "**not**
  `otelpgx`" phrasing is MATRIX CS-05's.) MATRIX CS-05 and `../plan.md`'s D-08 mapping elaborate
  the reason: "a third-party bridge
  would bind OTel vendor types into `kernel/database`, breaking the port discipline the adapters
  layer gets right."
- **A thin in-kernel `pgx.QueryTracer` implementation over the existing observability `Tracer`
  port** — chosen. See Decision below.

## Decision

**pgx query tracing via a thin in-kernel `pgx.QueryTracer` implementation (~50 LOC) over the
existing observability `Tracer` port — NOT `otelpgx` (a third-party bridge would bind OTel vendor
types into `kernel/database`, breaking the port discipline the adapters layer gets right).**
(REVIEW §U, combining the condensed decision-register line — "pgx query tracing via a thin
in-kernel `pgx.QueryTracer` over the existing observability port — `otelpgx` rejected to keep
vendor types out of `kernel/database`" — with the fuller phrasing from MATRIX CS-05,
reproduced in `../plan.md`'s D-08 mapping, which elaborates without contradicting the §U summary;
both describe the same decision restated at two points.)

### Safe default

No distinct safe-default stated beyond the decision itself — REVIEW §U states a direct
recommendation-plus-rejection, not a recommendation with a separate fallback path.

## Rationale

The framework's adapters layer already maintains a discipline of keeping vendor-specific types
(e.g. OTel SDK types) out of `kernel/*` packages, confining them instead to adapter packages that
implement the kernel's own ports. Adopting `otelpgx` directly inside `kernel/database` would import
OTel vendor types into a kernel package, breaking exactly the port-discipline boundary the adapters
layer is designed to preserve. A ~50-line, purpose-built `pgx.QueryTracer` that calls the
framework's own existing `Tracer` port avoids this: it is small enough to own and maintain directly,
and it keeps `kernel/database` dependent only on the framework's own abstraction, not on a
third-party vendor-specific bridge.

## Consequences

- FBL-06 (W01-E02, per `requirement-inventory.md`'s FBL-06 row: "OTel trace/log correlation + pgx
  tracer (D-08) | CS-05 T1–T3") implements the thin in-kernel tracer rather than adopting `otelpgx`.
- `otelpgx` is not added as a dependency of `kernel/database` — the framework's approved-dependency
  register (REVIEW §L) is not extended to include it as a consequence of this decision.
- Any OTel-specific tracing behavior remains confined to whatever adapter ultimately implements the
  framework's `Tracer` port for OTel — `kernel/database` itself stays OTel-agnostic, consistent with
  the framework/product and kernel/adapter boundary disciplines this programme otherwise enforces
  (mandate §2.3).

## Related source items

D-08; downstream epic W01-E02 (FBL-06) — unblocked by this ADR per
`../../../../dependencies.md` and `../story.md` "Dependencies." MATRIX CS-05 (T1-T3) is the
closure-depth spec implementing this decision's mechanics.

## Date

2026-07-12.

## Deciders

Fable 5.
