---
title: Module DSL target design
status: target-not-implemented
source_requirement: DX-03
story: W06-E01-S001
created_at: 2026-07-13
updated_at: 2026-07-13
---

# Module DSL target design

> **Target, not implemented.**

This document defines a future module-authoring contract. It does not describe an API that exists in
this repository today. No compiler, generator, runtime adapter, or exported DSL type accompanies this
design record.

## Purpose and boundaries

The target DSL makes module contracts declarative while keeping business behavior explicit. A module
must be reviewable as data: identity, dependencies, capabilities, schemas, operations, and owned
extension points are visible before boot. Handlers remain ordinary typed Go functions. The DSL does not
become a low-code workflow engine, hide I/O, or move business logic into reflection.

This record covers three connected concepts:

- `port`: an owner-bound, typed capability contract built on the existing `port.Key[T]` API;
- `Manifest[TConfig]`: the complete, typed declaration of one module; and
- `Operation[Request, Response]`: one typed transport-and-execution contract whose projections drive
  HTTP, OpenAPI, validation, authorization, jobs, audit, and compatibility metadata.

Actual implementation, package placement, compiler code, generated code, migration tooling, and a
cutover from `module.Module`/`module.Context` are future work.

## Landed baseline this design builds on

The W05 baseline is intentionally preserved rather than replaced:

1. `kernel/appmodel.Compiler` owns the `collecting -> validating -> sealed` lifecycle and returns an
   immutable `ApplicationModel` snapshot.
2. `Compiler.GetRegistrar(owner)` is the sole mint authority for
   `appmodel.Registrar[T]`. The registrar carries immutable owner identity, rejects its zero value, and
   is the capability through which declarations enter the model.
3. `kernel/port.Key[T]` carries a stable string ID and a phantom Go type. `port.Define`, `Provide`,
   `Require`, and `Resolve` translate that key through the owner-bound registrar; runtime reflection is
   confined behind that typed public surface.
4. The current `module.Module.Register(module.Context)` path is mutable and broad. It remains the
   current implementation until an explicit migration story replaces it.

The future DSL is a typed front end to those lifecycle and ownership guarantees. It must project into
(or evolve compatibly from) `ApplicationModel`; it must not create a second ownership graph or an
independent port container.

## Design principles

1. **One canonical declaration, many checked projections.** JSON Schema, OpenAPI, runtime validation,
   compatibility fingerprints, and generated adapters come from one typed schema/operation declaration.
2. **Owner identity is minted once.** A manifest supplies one `ModuleID`; nested declarations never
   accept a free-form owner string.
3. **Typed at author boundaries, erased only inside the compiler.** Authors work with
   `Key[T]`, `Schema[T]`, and `Operation[Q,R]`. Heterogeneous declarations may be type-erased only after
   construction, behind a sealed compiler-facing declaration interface.
4. **Registration and runtime are different capabilities.** Boot code receives an owner-bound
   declaration registrar. Handlers receive narrow immutable runtime handles, never a mutable registry.
5. **Invalid combinations fail model compilation.** No silent defaults for tenant scope,
   credentials, authorization, schema identity, asynchronous execution, or ownership.
6. **Business behavior stays explicit Go.** The handler, authorization target resolver,
   idempotency-key function, workflow action, and compensation behavior remain named typed functions.

## Proposed author model

The following names describe the contract rather than current package declarations. Final package
placement may change during the future implementation, but the ownership and typing rules are fixed by
this design.

A module exports a single immutable `Manifest[TConfig]` value. Typed constructors create port,
operation, event, job, workflow, rule, notification, document, and webhook declarations. A
`Contracts(...)` collector accepts only framework-created, sealed declaration values. This permits a
heterogeneous manifest without exposing `any` to module authors.

Conceptually:

```text
Manifest[Config]
  identity: ModuleID + module contract version
  dependencies: versioned module requirements
  config: one canonical Schema[Config]
  contracts: sealed declarations created by typed constructors
  assets: migrations, seeds, locales
```

The future compiler performs four phases:

1. **Collect:** validate manifest identity and mint exactly one owner-bound `Registrar` from
   `Manifest.ID`; collect sealed declarations.
2. **Resolve:** construct the dependency and typed-port provider graph. Definitions, providers, and
   requirements use the landed `port.Key[T]` identity and the same `ApplicationModel` graph.
3. **Validate/project:** enforce cross-field invariants and project schemas/operations into OpenAPI,
   route metadata, validators, job/event registrations, compatibility records, and generated adapters.
4. **Seal:** produce one immutable `ApplicationModel`; reject post-seal mutation. Runtime handles are
   derived from the sealed model and do not expose registries.

Compilation is deterministic. Declaration order does not affect the model hash or generated artifacts;
stable identities and explicitly ordered output do.

## `port`: typed capability declaration

### Chosen shape

`port` is not a new service locator. It is the manifest-level declaration vocabulary around the landed
`port.Key[T]` mechanism:

- the defining module creates one stable `Key[T]` with a globally serialized ID;
- a definition binds that key to the manifest owner;
- a provider binds exactly one implementation of `T` to a defined key;
- a requirement names the same typed key and may declare version/optional-profile constraints;
- resolving occurs only from the sealed model through a narrow runtime handle.

A nested port declaration never carries `Owner string`. The compiler derives ownership from the
registrar minted for `Manifest.ID`. Serialized IDs remain available for diagnostics and compatibility,
but module code does not assemble identity strings repeatedly.

### Invariants

- key ID is non-empty, stable, and unique across the application;
- a definition's Go type is immutable for that ID;
- every required port has one definition and exactly one compatible provider in the active profile;
- duplicate providers and type conflicts are compilation errors;
- optional ports are modeled as explicit optional requirements, not failed lookups ignored at runtime;
- resolution before seal and registration after seal are errors;
- module code cannot construct or forge a registrar for another owner.

### Options considered

| Option | Benefit | Cost | Decision |
|---|---|---|---|
| Raw string keys with `any` implementations | Minimal machinery; matches the legacy `module.Context` surface | Type errors move to boot/runtime; owner strings are forgeable; refactors are unsafe | Rejected |
| A distinct registrar type for every subsystem | Strong nominal separation | Multiplies APIs and compiler plumbing without adding safety beyond owner binding + phantom types | Rejected |
| One owner-bound registrar plus `port.Key[T]` | Reuses W05 guarantees; small public surface; compile-time author ergonomics | Compiler retains an internal type-erasure boundary | Chosen |

## `Manifest[TConfig]`: typed module declaration

### Chosen shape

`Manifest[TConfig]` contains:

- `ID`: stable `ModuleID`, used to mint ownership;
- `Version`: the module contract/schema version, independent of the wowapi release;
- `Dependencies`: module ID plus an explicit supported version constraint;
- `Config`: canonical `Schema[TConfig]`, including validation and compatibility identity;
- `Capabilities`: typed port requirements and other explicit capability requirements;
- `Contracts`: the sealed heterogeneous declaration set;
- `Migrations`, `Seeds`, and `Locales`: immutable asset bundles with stable fingerprints.

`TConfig` is the module-owned decoded configuration type. The compiler validates the config schema and
binds the decoded immutable value into the module's runtime handle. There is no untyped
`map[string]any` configuration escape hatch in the author contract.

The manifest is data, not an executable registration callback. Typed constructors may validate local
shape when the value is created, but only the application compiler may validate cross-module facts or
seal the model.

### Invariants

- module ID satisfies the existing module-name/identity policy and is unique;
- module version parses under the compatibility policy and is independent of framework version;
- dependencies exist, satisfy declared constraints, and form an acyclic graph;
- every declaration is owned by the manifest that contains it;
- contract IDs are unique within their owner namespace;
- schema and asset fingerprints are reproducible;
- migrations and event/schema versions never regress;
- a manifest cannot retain the mutable registrar after collection.

### Options considered

| Option | Benefit | Cost | Decision |
|---|---|---|---|
| Keep `Register(Context)` as the primary manifest | No migration | The contract remains hidden in imperative calls; mutable mega-interface survives | Rejected as target |
| Generated manifest only | Perfectly regular compiler input | Makes ordinary Go authoring dependent on code generation and obscures handwritten behavior | Rejected |
| Typed immutable manifest with optional generated construction | Declarative review surface; works for generated and handwritten modules; ordinary handlers remain Go | Requires a sealed heterogeneous declaration representation inside the compiler | Chosen |

## `Operation[Request, Response]`: typed operation contract

### Chosen shape

An operation is an immutable generic value with these fields or equivalent typed components:

- stable `ID`, transport method, and path;
- operation kind: synchronous, asynchronous, or stream;
- tenant scope: none, current tenant, or platform;
- authentication policy: allowed credential schemes and assurance requirements;
- authorization policy typed to `Request`, including a required target resolver for resource-scoped
  checks;
- canonical input `Schema[Request]` and output `Schema[Response]`;
- declared error contracts;
- emitted event keys, schema versions, and compatibility modes;
- execution policy, including job key/idempotency/retry/budget for async work and bounded
  backpressure/record/byte/duration limits for streams;
- idempotency, concurrency, audit, rate-limit, and observability policies; and
- a typed handler `func(context.Context, RequestContext, Request) (Response, error)`.

`Schema[T]` has one author-owned typed source. JSON Schema, fingerprint, decoding, and validation are
compiler projections; they are not independent fields that authors manually synchronize.

### Compile-time/model-time invariants

- anonymous operations use no tenant and only anonymous credentials, with no permission;
- current-tenant operations reject anonymous credentials and require an owned permission;
- platform operations accept only approved service/privileged credentials and explicit platform
  authorization;
- webhook credentials require the verified-envelope runtime contract;
- resource-scoped authorization requires a non-zero typed target resolver;
- asynchronous operations require a job key, stable idempotency function, positive attempts, and a
  bounded execution budget;
- stream operations require finite record, byte, and duration bounds plus backpressure policy;
- emitted events have owned typed keys and non-zero schema versions;
- declared errors have unique stable codes and valid protocol mappings;
- method/path/operation IDs are unique after route normalization;
- handler request and response types are exactly those carried by the schemas.

### Projection flow

```text
Schema[Request] + Schema[Response] + Operation[Request,Response]
                 |
                 v
        application-model compiler
          /       |        \
 HTTP route   OpenAPI   compatibility record
 validator    operation schema fingerprints
          \       |        /
           narrow runtime operation handle
```

The OpenAPI and compatibility outputs are therefore consequences of the same declaration used to bind
the handler. A mismatch is a compiler defect, not an author synchronization task.

### Options considered

| Option | Benefit | Cost | Decision |
|---|---|---|---|
| Structs containing `any`, reflection, and independent JSON Schema blobs | Easy heterogeneous storage | Type mismatch and schema drift remain possible; author burden is high | Rejected |
| Generate all handlers from a schema language | Uniform output | Business behavior becomes generator-owned; difficult escape hatches and review | Rejected |
| Generic typed operation values compiled to erased immutable descriptors | Typed handler/schema boundary; one source for projections; explicit behavior | Requires careful internal erasure and compiler diagnostics | Chosen |

## Runtime capability boundary

Registration receives declaration capability only. A request handler receives an immutable
`RequestContext` plus only the runtime handles its operation declares, such as transactions, events,
operations, and explicit typed ports. It cannot reach boot registries, declare new routes, change
providers, or retain a mutable `module.Context`.

The future migration should be a clean cutover only after generated and handwritten consumers can be
translated. Until then the existing `module.Module` path remains authoritative; this document creates
no dual runtime and no compatibility shim.

## Diagnostics and error model

Compiler errors are deterministic, stable, and owner-qualified. Each error identifies the manifest,
declaration kind, declaration ID, violated invariant, and conflicting declaration where applicable.
Independent errors are accumulated so one compile reports the complete actionable set. Security
invariants fail closed; there is no warning-only mode for missing authorization, ambiguous provider
selection, schema mismatch, unbounded stream/async policy, or unknown workflow action version.

## Compatibility and versioning

- Framework release, module contract version, schema version, and event version are distinct axes.
- Stable serialized IDs and canonical fingerprints are included in the sealed model snapshot.
- Additive compatible changes and breaking changes are classified by the repository's v1/N-1 policy.
- Port type changes, removed required operations, incompatible schema changes, regressed migration or
  event versions, and changed ownership are breaking unless an explicit policy says otherwise.
- Generated artifacts include the compiler/framework version and model hash so stale output fails
  loudly.

## Future implementation sequence

> **Target, not implemented.**

A future implementation should proceed in this order:

1. specify immutable descriptor and diagnostic formats without changing `module.Module`;
2. add typed constructors and prove owner/key forgery resistance;
3. add compiler projections into the existing `ApplicationModel` and typed-port graph;
4. add deterministic model hashing and schema/OpenAPI projections;
5. add narrow runtime handles and generated-consumer contract tests;
6. migrate every scaffold and consumer before removing the legacy `module.Context` registration path.

Each step requires fail-first tests for ownership, type conflict, missing providers, invalid security
combinations, deterministic output, post-seal mutation, and generated-consumer boot. This story does
not perform any of those steps.

## Acceptance mapping

- `port`, `Manifest[TConfig]`, and `Operation[Request,Response]` are specified with author contract,
  compiler flow, invariants, compatibility behavior, and alternatives.
- The design explicitly builds on W05's `ApplicationModel`, owner-bound `Registrar[T]`, and
  `port.Key[T]` instead of replacing them.
- The record is visibly labeled **Target, not implemented** and introduces no implementation code.
