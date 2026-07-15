---
id: PLAN-W01-E03-S002
type: plan
parent_story: W01-E03-S002
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
---

# Plan — W01-E03-S002 — Central validation enforcement

Per mandate §8.5: "Do not invent precise code changes where the repository does not yet provide
enough information. Clearly distinguish confirmed facts, planned changes, and implementation
assumptions." As in the sibling S001 plan, facts below are marked `[CONFIRMED]` (verified directly
against repository state on 2026-07-12), `[PLANNED]` (this story's intended change, shape not yet
finalized), or `[ASSUMPTION]` (an implementation-time judgment call, listed under "Unresolved
questions").

## Proposed architecture

`[CONFIRMED]` `RouteMeta` (`kernel/httpx/router.go:18-33`) is the existing boot-validated mandatory-
metadata seam: every route passes through `Router.Handle`, which calls `meta.validate()`
(`router.go:36-45`) and accumulates any error into `r.errs`, surfaced at app boot via `Router.Err()`.
This is not a proposal — it is the existing architecture this story extends.

`[PLANNED]` This story extends `meta.validate()` (or adds an equivalent check invoked from
`Router.Handle` at the same point) with a second invariant: a route whose `Method` is POST, PUT, or
PATCH must have a declared request contract on `RouteMeta`, unless a waiver is present, and unless
the profile flag gating this whole check is off. The check plugs into the exact same error-
accumulation mechanism the existing `Public`/`Permission` invariant uses — no new enforcement
mechanism is architecturally introduced, only a second condition inside (or alongside) the existing
one.

`[PLANNED]` A new generic handler adaptor (naming/location `[ASSUMPTION]`, likely alongside
`BindAndValidate` in `kernel/httpx/decode.go` or a new file in the same package) that:
1. Reads the declared `Request` contract off the route's `RouteMeta` (or is itself parameterized by
   the contract type `T`, with the route registration wiring `RouteMeta.Request` to the same `T`).
2. Calls the existing `BindAndValidate[T]` internally.
3. Passes the validated value to the handler's actual business logic.

The exact mechanism for "the type declared in `RouteMeta.Request` is the same type the adaptor binds
into" is the core open design question of this story — see "Unresolved questions" below. Two
candidate shapes exist and neither is invented as final here:

- **Candidate A — type-token field**: `RouteMeta.Request` holds something like a `reflect.Type` or a
  zero-value `any` used only for the boot-time presence check (proving *a* contract was declared),
  while the actual binding still happens via a generic adaptor function called with an explicit type
  parameter at the route-registration call site (`httpx.ValidatedHandler[CreateOrderRequest](...)`).
  This keeps Go's static generics model intact (Go cannot express "the type stored in this struct
  field" as a compile-time type parameter) at the cost of the declared type and the bound type being
  two separate pieces of code that could theoretically drift.
- **Candidate B — adaptor-only declaration, no separate RouteMeta field**: instead of a
  `RouteMeta.Request` field holding a type, the act of registering a route *through* the new generic
  adaptor is itself what satisfies the "has a declared contract" boot check — e.g. the adaptor sets a
  boolean/marker on the `RouteMeta` it hands back to `Router.Handle`, so `RouteMeta.Request` becomes
  more like `RouteMeta.HasValidatedRequest bool`, set automatically by the adaptor rather than
  hand-declared by the caller. This removes the type/binding-drift risk of Candidate A but makes the
  boot check purely presence-based (did *any* validation run) rather than contract-shape-based.

This plan does not choose between Candidate A and Candidate B — that choice belongs to
implementation-time design review, informed by which shape composes better with AR-03's future
RouteMeta-projection consumption (out of scope to resolve now, but the choice should not foreclose
AR-03's later options without recording why).

## Implementation strategy

Three tasks, corresponding to the three architectural pieces above:

1. T001 — the `RouteMeta` extension (contract-declaration field, boot-time check, waiver field) —
   the foundational piece; T002 depends on T001's resolved field shape.
2. T002 — the handler adaptor, built against whichever shape T001 resolves.
3. T003 — crud/scaffold template migration to use T002's adaptor, the last step since it depends on
   both T001 and T002 existing and being stable.

## Expected package or module changes

- `kernel/httpx` — `router.go` (`RouteMeta` struct, `validate()` method, possibly `Router.Handle`);
  `decode.go` or a new file (the adaptor).
- `internal/cli/templates/...` — the crud/scaffold template set (exact files `[ASSUMPTION]`, to be
  located at T003's start; likely distinct from S001's `init` template set since crud generation is a
  separate `wowapi gen crud` code path per DX-02's own story scope in W01-E04).
- `kernel/config` (if the profile flag is framework-level) or a product-level config location (if
  the flag is product-owned) — `[ASSUMPTION]`, see "Unresolved questions."

## Expected file changes where determinable

- `kernel/httpx/router.go` — near `RouteMeta` (line 18-33) and `validate()` (line 36-45).
- `kernel/httpx/decode.go` — near `BindAndValidate` (line 52-67), if the adaptor is added to the same
  file.
- A new fixture-route test file (or extension of an existing `kernel/httpx` test file) proving the
  fail-first boot-passes-today / boot-fails-after-T1 sequence.
- crud/scaffold template file(s) — not yet located; T003 must first identify them.

## Contracts and interfaces

`[PLANNED]` `RouteMeta` gains a new field (exact name/type per the Candidate A/B resolution above)
and possibly a new waiver field. This is an additive struct change — existing `RouteMeta{...}`
literals across the codebase (and any product's already-generated code) that do not set the new
field continue to compile (Go's zero-value struct-literal semantics), and continue to boot
successfully as long as the profile flag defaults to off.

`[PLANNED]` A new generic adaptor function, exported from `kernel/httpx`, with a signature
resembling `func ValidatedHandler[T any](v *validation.Validator, maxBytes int64, fn func(*http.Request, T) (...)) http.HandlerFunc`
or similar — exact signature `[ASSUMPTION]`, not invented as final here; it must compose with
`BindAndValidate[T]`'s existing signature (`func BindAndValidate[T any](r *http.Request, v
*validation.Validator, maxBytes int64) (T, error)`) without changing that function's own signature
(no reason to break existing direct callers of `BindAndValidate`).

## Data structures

See "Proposed architecture" Candidates A and B above. Neither is finalized.

## APIs

No externally-visible HTTP API surface changes — this story changes internal wiring (how a handler
validates), not the shape of any response. The 400 field-error response shape (AC-W01-E03-S002-03)
is already the existing `KindValidation` error shape `BindAndValidate` produces today; this story
does not change that shape, only guarantees it is produced for every route using the new adaptor.

## Configuration changes

`[PLANNED]` A new profile flag gating the boot-time rejection. Location (`kernel/config` framework-
level vs. a product-owned config section) is an open question — see "Unresolved questions."

## Persistence changes

None.

## Migration strategy

None — no data or schema migration. The profile flag itself is this story's compatibility/rollout
mechanism (see `story.md` "Compatibility considerations").

## Concurrency implications

None beyond the existing `Router` construction-time (single-goroutine, boot-only) concurrency model —
the boot-time check runs once, synchronously, at the same point `meta.validate()` already runs today.

## Error-handling strategy

The boot-time rejection reuses `Router`'s existing accumulate-and-surface-at-`Err()` pattern
(`router.go:73-89`), consistent with every other route-registration error. No new error-handling
pattern is introduced at the boot layer. At the request layer, the adaptor's validation failures
reuse `BindAndValidate`'s existing `KindValidation` error path unchanged.

## Security controls

The entire boot-time-mandatory-contract mechanism is itself the security control this story adds —
see `story.md` "Security considerations."

## Observability changes

None planned.

## Testing strategy

- **Fail-first fixture test (T001)**: a fixture route with a POST handler and no declared contract,
  proven to boot successfully today (pre-fix), then proven to fail boot once T1's check + the profile
  flag are both in place (mandate §13 fail-first).
- **Waiver test (T001)**: a fixture route using the waiver field, proven to boot successfully with
  the enforcement flag enabled.
- **Adaptor unit test (T002)**: the new adaptor correctly binds and validates a request, correctly
  rejects an invalid one with the expected field-error shape.
- **Adversarial 400 test (T002 or story-level integration test)**: an end-to-end (or handler-level)
  test posting an invalid DTO through a route built with the new adaptor and asserting HTTP 400 with
  field errors — this is the story's headline adversarial proof (AC-W01-E03-S002-03).
- **crud-template test (T003)**: whatever existing generator-output test infrastructure DX-02 (W01-
  E04) builds or already has (the generator-output-boots test referenced at wave level) should be
  extended, not duplicated, to assert the migrated crud template's generated handlers use the new
  adaptor and boot successfully.

## Regression strategy

The profile flag defaulting to off is the primary regression guard: no existing route's boot
behavior changes until a product explicitly opts in. The adaptor is additive (a new function; no
existing function signature changes), so no existing direct `BindAndValidate` caller is affected.

## Compatibility strategy

Profile-flag-first, per FBL-08's own explicit note and `story.md` "Compatibility considerations." No
breaking change ships in this story; the breaking-if-misused change (enforcement) is opt-in and
downstream-audited before any product flips it on.

## Rollout strategy

Ships in a wowapi release with the flag defaulting off. A downstream product (wowsociety) audits its
own mutating routes for a declared contract, then flips the flag — this rollout step is downstream
coordination, not part of this story's own rollout.

## Rollback strategy

If the boot-time check itself proves defective (false-positive rejections), the profile flag can be
turned back off without any code rollback, since the check's entire activation is flag-gated. A code-
level rollback (reverting the commit) is the fallback only if the flag mechanism itself is found to
be broken.

## Implementation sequence

1. T001 — `RouteMeta.Request` field (Candidate A or B, resolved before implementation) + boot-time
   check + waiver field, forward-compatible with AR-04 T5.
2. T002 — `BindAndValidate`-calling generic handler adaptor, built against T001's resolved shape.
3. T003 — crud/scaffold template migration to the adaptor.

## Task breakdown

- **W01-E03-S002-T001** — `RouteMeta.Request` field + boot-time rejection + waiver field,
  forward-compatible with AR-04 T5.
- **W01-E03-S002-T002** — `BindAndValidate`-calling generic handler adaptor.
- **W01-E03-S002-T003** — crud/scaffold template migration to the adaptor.

## Expected artifacts

- `RouteMeta.Request` contract type (or resolved-shape equivalent).
- Handler adaptor.
- Updated crud/scaffold template(s).
- See `../artifacts/index.md`.

## Expected evidence

- Boot-rejection fail-first test pair.
- Adversarial invalid-DTO 400 test.
- Waiver-exemption boot-success test.
- See `../evidence/index.md`.

## Unresolved questions

1. **Candidate A vs. Candidate B (contract-declaration shape)** — see "Proposed architecture." This
   is the single largest open design question in this story and must be resolved by implementation-
   time design review (referencing `codebase-design` skill guidance on seam placement, per this
   repository's own `CLAUDE.md` routing) before T001 is implemented.
2. **Profile-flag location** — framework-level (`kernel/config`) vs. product-owned config section.
   `kernel/config.Framework` is the more consistent location given the other prod-safety-floor flags
   already living there (e.g. `WebhookOutbound.SSRFProtectionDisabled`'s `unsafe:"true"` pattern), but
   this enforcement flag is arguably closer to a build-time/product-adoption concern than a per-
   environment runtime toggle — not resolved here.
3. **Waiver field shape** — minimal (a `bool`) vs. richer (a reason string, an approver field). This
   plan recommends minimal, per `story.md` "Assumptions," to reduce the surface AR-04 T5 later
   reconciles, but this is a recommendation, not a decision recorded as final.
4. **Whether the boot-time check lives inside `meta.validate()` itself or as a separate check invoked
   from `Router.Handle` alongside it** — `meta.validate()` currently has no method-awareness (it does
   not know if the route is POST/PUT/PATCH; that information lives on the `Route` struct's `Method`
   field, not on `RouteMeta` itself, per `router.go:49-54`). This means the new check likely cannot
   live purely inside `RouteMeta.validate()` (a method on `RouteMeta` alone, with no access to the
   HTTP method) without either passing the method in as a parameter or relocating the check to
   `Router.Handle` (which already has `method` in scope at the call site, `router.go:73`). This is a
   confirmed structural fact `[CONFIRMED]`, not an assumption — `RouteMeta.validate()`'s signature is
   `func (m RouteMeta) validate() error`, no parameters. T001 must account for this when placing the
   new check.

## Approval conditions

This plan is approved for implementation once unresolved question 1 (Candidate A vs. B) is resolved
by design review, since T001 and T002 cannot proceed on a coherent shape without it. Questions 2 and
3 may be resolved concurrently with or shortly after T001's start; question 4 is a structural
constraint T001 must satisfy regardless of which candidate is chosen, not a blocking decision.
