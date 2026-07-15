---
id: PLAN-W05-E01-S002
type: plan
parent_story: W05-E01-S002
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
---

# Plan — W05-E01-S002

Per mandate §8.5. Confirmed facts, planned changes, and implementation assumptions are distinguished
explicitly below.

## Proposed architecture

Four owner-bound registrar wrappers, each consuming S001's `Registrar` capability type and
per-subsystem typed keys: `resource.Registry`, `rules.Registry`, `authz.Registry` (permission
registration), and a table-driven set covering the remaining ~9+ declaration classes. Each wrapper
follows the same structural pattern established by T3 (the first, reference implementation): the
registry's registration entry point accepts (or is scoped by) a `Registrar` rather than an arbitrary
owner string, and ownership is checked by the capability's own binding, not a string comparison.

## Implementation strategy

1. Implement T3 (`resource.Registry` wrapper) first as the reference pattern — `ctx.Resources()`
   exposes a registrar bound to the module's own identity.
2. Implement T4 (`rules.Registry` wrapper) following the same shape as T3.
3. Implement T5 (`authz.Registry` permission-registration wrapper) — the widest gap, requiring an
   API-signature change since `Register(p Permission)` today has no owner parameter at all; the new
   API derives the module prefix from the bound registrar.
4. Audit the framework's actual registration surface to confirm the full list of remaining
   declaration classes beyond the three headline registries (starting from PLAN's own named list:
   events, jobs, workflow actions, providers, templates, health checks, migrations, seeds, OpenAPI).
5. Implement T6's wrappers for each confirmed declaration class, following the T3-T5 pattern.
6. Write the adversarial tests: `AR-01/resource_ownership_adversarial_test.go`,
   `AR-01/rules_ownership_adversarial_test.go`, `AR-01/authz_ownership_adversarial_test.go` (T3-T5);
   `AR-01/full_declaration_class_matrix_test.go` (T6, table-driven, one fixture per class).
7. Document each wrapper's ownership-bound API, with emphasis on T5's changed signature.

## Expected package or module changes

`kernel/resource`, `kernel/rules`, `kernel/authz` (T3-T5); the packages backing each confirmed
remaining declaration class (T6, exact list TBD by the audit in step 4 above).

## Expected file changes where determinable

- `kernel/resource`'s registration entry point — extended or wrapped to accept a `Registrar`.
- `kernel/rules`'s registration entry point — same shape.
- `kernel/authz`'s `Register(p Permission)` — signature change to derive owner from the bound
  `Registrar` rather than accepting none.
- New adversarial test files as named above.
- New wrapper files for each T6 declaration class (exact paths TBD by the audit).

## Contracts and interfaces

Each wrapped registry's registration entry point changes from an unowned or string-owned call to a
`Registrar`-bound call. `authz.Registry`'s permission-registration signature changes materially
(from no owner parameter to a bound-registrar-derived owner) — this is the story's most significant
interface change and the one requiring the most careful compatibility handling in S004's legacy
adapter.

## Data structures

No new persistent data structures — this is registration-API surface work over in-memory registries.

## APIs

Internal Go APIs only (registration entry points); no external HTTP/gRPC surface change.

## Configuration changes

None anticipated.

## Persistence changes

None.

## Migration strategy

Not applicable.

## Concurrency implications

Registration happens at boot time, single-threaded by convention in the existing framework
(collect-phase, per S001's lifecycle skeleton) — no new concurrency concern introduced by this
story's wrappers themselves; S003's own race tests (T10) cover the sealed model's read-time
concurrency, not this story's collect-time wrapper logic.

## Error-handling strategy

A cross-module ownership claim attempt must fail — for the resource/rules/authz wrappers (T3-T5),
this failure surfaces as an adversarial-test-provable rejection at the registrar boundary (a runtime
error path, since these are registration calls at boot, not a compile-fail fixture like S001 T2's
scenario). The specific error must be clear enough to diagnose (naming the attempted owner mismatch),
consistent with this programme's broader error-handling posture (e.g. W02-E01-S001's field-specific
manifest-validation error requirement).

## Security controls

T5's fix closes the framework's widest zero-ownership-check gap — this is the central security
control this story delivers. T3/T4/T6's fixes close the remaining string-compared or otherwise
non-structural ownership checks.

## Observability changes

Boot-time logging of a rejected cross-module ownership claim (implementation-time addition, per
`story.md` "Observability considerations").

## Testing strategy

- One adversarial test per headline registry (T3, T4, T5) proving cross-module claim rejection, per
  PLAN's own named test files.
- A table-driven adversarial suite (T6) with one fixture per remaining declaration class, so the
  audit's completeness is directly visible in the test's own fixture count.
- No integration or performance test separately required — these are registration-time correctness
  tests, not runtime-behavior tests.

## Regression strategy

The adversarial tests, once established, are permanent regression guards: any future change that
reintroduces a string-comparable or unchecked ownership path in any wrapped registry should cause
the corresponding adversarial test to start failing (or, for a genuinely new bypass path, to reveal
a gap the existing fixture didn't cover — a signal to extend the fixture, not silently accept the
gap).

## Compatibility strategy

T5's `authz.Registry` signature change is the story's most compatibility-sensitive item. This story
does not itself provide a compatibility shim — that is explicitly S004's scope (T11, the legacy
adapter) — but this story's own implementation must be structured so S004's adapter can route
existing calls through the new signature without requiring every existing caller to be rewritten
immediately. This dependency is recorded here, not silently assumed to resolve itself.

## Rollout strategy

Single story, landed as its own reviewable unit, but with an explicit awareness that T5's signature
change is not usable by existing callers until S004's legacy adapter lands — this story's own scope
ends at "the new ownership-bound API exists and is adversarially proven," not "existing callers are
migrated."

## Rollback strategy

Revert the T5 signature change if it is found to be incompatible with the legacy-adapter approach
S004 will build — escalate for redesign rather than silently narrowing T5's own ownership-check
scope to make S004 easier.

## Implementation sequence

As listed under "Implementation strategy" above (steps 1-7). Step 3 (T5) should not be delayed past
T3/T4, despite being higher-risk, since PLAN's own task table lists no ordering dependency of T5 on
T3/T4 beyond the shared T1/T2 prerequisite — treating T5 first or in parallel with T3/T4 is
consistent with the source, though this plan sequences it after T3/T4 to establish the reference
pattern first, a documented implementation choice, not a source requirement.

## Task breakdown

- **W05-E01-S002-T001** — `resource.Registry` and `rules.Registry` owner-bound wrappers (T3, T4;
  grouped as the "reference pattern" pair).
- **W05-E01-S002-T002** — `authz.Registry` permission-registration owner-bound wrapper (T5; the
  widest-gap, highest-risk task, kept separate given its distinct risk profile and API-signature
  change).
- **W05-E01-S002-T003** — Owner-bound wrappers for the remaining ~9+ declaration classes (T6).
- **W05-E01-S002-T004** — Independent review (per mandate §14, scoped to this story, given T5's
  High-risk status and T6's under-scoping risk).

## Expected artifacts

Owner-bound wrappers for `resource.Registry`, `rules.Registry`, `authz.Registry`, and the remaining
~9+ declaration classes; the declaration-class enumeration/audit record.

## Expected evidence

The four named adversarial test outputs (resource, rules, authz, full declaration-class matrix).

## Unresolved questions

- The exact confirmed count of "the remaining ~9+ declaration classes" — PLAN's own "~9+" phrasing
  is not a closed count; T6's own audit must confirm the exact list at implementation time.
- Exact API signature for `authz.Registry`'s new permission-registration entry point (T5) — PLAN
  states the design direction ("derives module prefix from the bound registrar") but not the
  concrete Go signature.
- Whether T3/T4/T6's wrappers share a common generic implementation (given they follow "the same
  shape as T3" per PLAN T4's own wording) or are implemented as separate, registry-specific code —
  to be decided at implementation time; a shared generic implementation would reduce duplication but
  is not mandated by the source.

## Approval conditions

This plan is approved for implementation once: (a) T6's declaration-class audit is complete and its
list is recorded, and (b) the owner and reviewer are assigned.
