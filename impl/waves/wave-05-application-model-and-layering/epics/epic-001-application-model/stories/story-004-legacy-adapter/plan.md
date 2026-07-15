---
id: PLAN-W05-E01-S004
type: plan
parent_story: W05-E01-S004
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
---

# Plan — W05-E01-S004

Per mandate §8.5. Confirmed facts, planned changes, and implementation assumptions are distinguished
explicitly below.

## Proposed architecture

An adapter layer wrapping the current `module.Module`/`Context` interface, translating each existing
module's registration calls into calls against S001-S003's ownership-bound `ApplicationModel` and
its `Registrar` capability type — deriving the `Registrar`'s owner identity from `Module.Name()`
rather than requiring the existing module interface to change.

## Implementation strategy

1. Re-read the current `module.Module`/`Context` implementation at this story's start commit to
   confirm it has not materially changed since S001-S003 began.
2. Implement the adapter: for each existing registration call surface (`ctx.Resources()`,
   `ctx.Rules()`, permission registration, and the remaining ~9+ declaration classes), route the
   call through the corresponding S002 wrapper, using a `Registrar` minted for the module's own
   `Module.Name()`.
3. Confirm the adapter's owner-derivation is structurally tied to `Module.Name()` — not an
   independently-suppliable string the adapter itself could get wrong.
4. Run S002's own adversarial fixtures (resource, rules, authz, full declaration-class matrix)
   through the legacy path, confirming identical rejection behavior to the non-legacy path — this is
   the story's primary trust-boundary proof.
5. Run existing contract tests (wowapi-internal and wowsociety) unmodified through the legacy path,
   capturing `AR-01/legacy_adapter_compat_test_output.txt`.
6. Document the adapter's owner-derivation mechanism and its non-bypass guarantee.

## Expected package or module changes

A new adapter layer wrapping the existing `kernel/module` package's `Module`/`Context` surface. No
existing module's own source is changed.

## Expected file changes where determinable

- A new adapter file (or files) within or adjacent to `kernel/module` (exact path TBD).
- A new integration test capturing `AR-01/legacy_adapter_compat_test_output.txt`.
- Adversarial-fixture-through-legacy-path test files (reusing or extending S002's own fixtures).

## Contracts and interfaces

The existing `module.Module`/`Context` interface is preserved unchanged (this is the entire point of
the adapter) — no existing module needs to change its own implementation of this interface. The
adapter itself is new code sitting between the existing interface and S001-S003's new
ownership-bound API.

## Data structures

None new beyond the adapter's own internal owner-mapping (module → minted `Registrar`), which is
itself derived, not separately stored state requiring its own persistence.

## APIs

None externally facing.

## Configuration changes

None.

## Persistence changes

None.

## Migration strategy

Not applicable — no existing module is migrated by this story; the adapter is the mechanism that
makes migration optional/deferred.

## Concurrency implications

None beyond what S001-S003 already establish for the underlying model — the adapter itself operates
at boot time, single-threaded by the existing framework's own registration-phase convention.

## Error-handling strategy

Any rejection the adapter surfaces (via routing through S002's wrappers) must be the same error the
non-legacy path would produce — no adapter-specific error wrapping that could obscure the underlying
ownership-check failure.

## Security controls

The adapter's non-bypass guarantee — proven by running S002's adversarial fixtures through the
legacy path — is itself the required security control for this story, per PLAN's own "it must not
bypass T2-T6" acceptance criterion.

## Observability changes

None beyond what S001-S003 already establish; the adapter's routed calls pass through the same
boot-time logging.

## Testing strategy

- Adversarial-fixtures-through-legacy-path: re-run S002's own resource/rules/authz/full-matrix
  fixtures through the legacy adapter path, confirming identical rejection behavior.
- Existing-contract-tests-through-legacy-path: wowapi-internal and wowsociety's own module contract
  tests pass unmodified, captured as `AR-01/legacy_adapter_compat_test_output.txt`.

## Regression strategy

The adversarial-fixtures-through-legacy-path test is itself the permanent regression guard against
the adapter silently becoming a bypass in the future — any change to the adapter that reintroduces a
bypass would be caught by re-running S002's own fixtures through it.

## Compatibility strategy

This story's entire purpose is compatibility — existing modules boot unchanged, no source change
required. See "Proposed architecture" above.

## Rollout strategy

Single story, landed as its own reviewable unit, sequenced after S001-S003. Once landed, existing
modules continue functioning through the legacy path indefinitely (or until each migrates on its own
schedule, per PLAN's own wowsociety-impact note) — no forced migration timeline is imposed by this
story.

## Rollback strategy

Revert the adapter if the adversarial-fixtures-through-legacy-path test reveals a bypass — do not
ship an adapter that silently reintroduces the unowned-registration gap AR-01 exists to close.

## Implementation sequence

As listed under "Implementation strategy" above (steps 1-6).

## Task breakdown

- **W05-E01-S004-T001** — Legacy adapter implementation and owner-derivation mechanism (steps 2-3
  above).
- **W05-E01-S004-T002** — Independent review (per mandate §14, scoped to this story, given PLAN's own
  explicit "the adapter is itself a trust boundary" framing).

Note: the adversarial-fixtures-through-legacy-path proof and the existing-contract-tests proof
(steps 4-5) are both included within T001's own scope as its required evidence, not split into a
separate task — they are two halves of the same "prove the adapter doesn't bypass anything and
doesn't break anything" objective for a single, bounded piece of code.

## Expected artifacts

The legacy adapter implementation (code); legacy-adapter documentation.

## Expected evidence

`AR-01/legacy_adapter_compat_test_output.txt`; adversarial-fixture-through-legacy-path test output.

## Unresolved questions

- Exact adapter package location (within `kernel/module` or a new adjacent package) — to be decided
  at implementation time.
- Whether the adapter is a permanent compatibility layer or has a planned deprecation timeline (PLAN
  frames it as "Wave 1 compatibility strategy," implying it is not necessarily permanent, but no
  explicit deprecation plan is stated in the source) — recorded as an open question, not invented
  here.

## Approval conditions

This plan is approved for implementation once the owner and reviewer are assigned.
