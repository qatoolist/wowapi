---
id: ADR-W06-E01-S001-001
type: decision
title: Typed immutable manifests compiled into the landed application model
status: target-not-implemented
context: DX-03 future module DSL design
story: W06-E01-S001
date: 2026-07-13
deciders:
  - W06-E01-S001 design task
related_source_items:
  - DX-03
  - AR-01
  - AR-02
created_at: 2026-07-13
updated_at: 2026-07-13
derived: false
---

# ADR-W06-E01-S001-001 — Typed immutable manifests compiled into the landed application model

> **Target, not implemented.**

## Decision ID

ADR-W06-E01-S001-001.

## Status

**Target, not implemented.** This is a future-state design decision. No DSL type, compiler, generator,
runtime adapter, or migration is implemented by this decision record.

## Context

DX-03 requires a module DSL that is declarative about contracts and explicit about business behavior.
The current implementation is `module.Module.Register(module.Context)`, with a broad mutable context.
W05 has since landed the foundation that a future DSL must reuse:

- `kernel/appmodel.Compiler` collects, validates, and seals one immutable `ApplicationModel`;
- `Compiler.GetRegistrar(owner)` is the sole mint authority for an owner-bound
  `appmodel.Registrar[T]`; and
- `kernel/port.Key[T]` plus `port.Define`/`Provide`/`Require`/`Resolve` provides a typed author surface
  over the compiled provider graph.

The decision must give future implementers a precise direction without claiming those future APIs
exist today. Detailed shapes and invariants are recorded in
`docs/implementation/module-dsl-target-design.md`.

## Options considered

### Option A — Retain imperative `Register(Context)` as the target

Continue widening `module.Context` and let modules declare contracts by calling mutable registries.

- Advantage: no migration and minimal new compiler machinery.
- Disadvantage: the complete module contract is not inspectable as data; handlers can retain mutable
  boot capabilities; string/`any` boundaries remain; schema and OpenAPI projections can drift.

### Option B — Generate all modules and handlers from an external schema language

Make generation the only authoring path and compile that schema into runtime code.

- Advantage: uniform compiler input and deterministic generated output.
- Disadvantage: moves business behavior into a generator-centric model, makes ordinary Go extension
  awkward, and creates a second source language whose escape hatches become the real API.

### Option C — Typed immutable Go declarations with a boot compiler

Authors declare an immutable `Manifest[TConfig]`, owner-bound typed ports based on `port.Key[T]`, and
typed `Operation[Request,Response]` values. Typed constructors erase heterogeneous declarations only
inside a sealed compiler boundary. The compiler projects one declaration into routes, schemas,
OpenAPI, compatibility metadata, job/event wiring, and immutable runtime handles, then seals the
existing `ApplicationModel`.

- Advantage: declarative and reviewable contracts, typed handlers and schemas, one canonical source
  for projections, ordinary explicit Go business behavior, and direct reuse of W05 ownership and port
  guarantees.
- Cost: requires a disciplined internal type-erasure boundary, deterministic diagnostics and model
  hashing, and an eventual whole-consumer migration away from `module.Context`.

## Decision

Choose **Option C: typed immutable Go declarations compiled into the landed application model**.

Specifically:

1. `port` is manifest-level vocabulary around the existing `port.Key[T]`; it is not a second service
   locator. Definitions, providers, requirements, and resolution retain the landed typed-key and
   owner-bound registrar guarantees.
2. `Manifest[TConfig]` is immutable data containing module identity and version, versioned
   dependencies, canonical typed configuration schema, sealed typed contracts, migrations, seeds, and
   locales. Nested declarations do not accept an owner string.
3. `Operation[Request,Response]` binds typed request/output schemas, handler, transport, tenancy,
   authentication, authorization, errors, events, execution, idempotency, concurrency, audit,
   rate-limit, and observability policy in one immutable declaration.
4. Author-facing declarations remain generic and typed. Heterogeneous storage is erased only behind a
   compiler-owned sealed declaration interface after typed construction.
5. The compiler mints one registrar from `Manifest.ID`, validates the entire module/port/operation
   graph, emits deterministic projections, and seals one `ApplicationModel`.
6. Registration capability and runtime capability are separate. Handlers receive narrow immutable
   handles and cannot see mutable boot registries.
7. Business behavior remains explicit typed Go functions. The DSL does not interpret or synthesize
   domain logic.

## Required invariants

- owner identity is derived once from `Manifest.ID` and cannot be forged by nested declarations;
- port IDs are stable and uniquely typed, every requirement has one compatible provider, and
  duplicate/type-conflicting providers fail compilation;
- manifest dependency graphs are complete, version-compatible, and acyclic;
- schema/OpenAPI/validation/fingerprint outputs come from one canonical typed declaration;
- anonymous/current-tenant/platform operation combinations fail closed when credentials,
  authorization, or tenancy are inconsistent;
- asynchronous and streaming operations are explicitly bounded and idempotent as applicable;
- model output and hashes are deterministic; and
- post-seal mutation and pre-seal resolution remain errors.

## Rationale

Option C is evolutionary: it deepens the W05 `ApplicationModel`/`Registrar[T]`/`port.Key[T]`
foundation instead of discarding it or creating parallel ownership and provider graphs. It also meets
DX-03's central tension. Contracts become declarative and compiler-visible, while behavior remains
normal typed Go that can be reviewed, tested, and debugged without runtime magic.

A single generic owner-bound registrar plus typed keys is preferred over per-subsystem registrar types
because capability confusion is already prevented by immutable owner binding and phantom key types.
A typed manifest is preferred over an imperative callback because the compiler can inspect and compare
the whole contract before any runtime wiring occurs. Generic operations with internal sealed erasure
are preferred over author-facing `any` because handler, schema, and authorization target types remain
coupled at compile time.

## Consequences

### Positive

- Module contracts can be inspected, hashed, diffed, documented, and compiled before boot.
- OpenAPI, validation, and compatibility metadata cannot silently diverge from handler types by author
  input.
- Ownership and typed-port safety reuse already-landed mechanisms.
- Generated and handwritten modules can target the same declaration model.
- Runtime code receives smaller immutable capabilities than the current `module.Context`.

### Negative

- A future implementation must design safe internal generic erasure and high-quality accumulated
  compiler diagnostics.
- The existing `module.Module`/`module.Context` surface cannot be removed until every scaffold and
  supported consumer migrates.
- Compatibility tooling must distinguish framework, module-contract, schema, and event versions.

### Neutral / deferred

- Final Go package names are intentionally deferred to implementation, but the ownership, typing,
  compilation, and runtime-capability decisions above are not.
- Code generation is an optional projection/construction aid, not the source of business behavior.
- No compatibility shim or dual runtime is authorized by this ADR; a future migration plan must define
  the cutover and update every caller.

## Rejected implementation shortcuts

- author-supplied independent JSON Schema, fingerprint, and validator fields;
- free-form owner strings on nested declarations;
- raw-string/`any` ports in the author API;
- handlers retaining the boot registrar or mutable registries;
- warning-only treatment of missing authorization, unsatisfied ports, or unbounded async/stream work;
- a local DSL compiler that creates a second application/provider graph beside `ApplicationModel`.

## Implementation status

> **Target, not implemented.**

DX-03 implementation tasks T1..Tn remain outside this programme. This ADR and its design document are
the complete output of W06-E01-S001; there are no accompanying `.go` files or runtime changes.

## Related records

- `docs/implementation/module-dsl-target-design.md`
- `kernel/appmodel/appmodel.go` (current W05 ApplicationModel/Registrar baseline)
- `kernel/port/port.go` (current W05 typed-port baseline)
- W00 ADR-W00-E02-S003-002 (single owner-bound registrar with typed keys)

## Date

2026-07-13.
