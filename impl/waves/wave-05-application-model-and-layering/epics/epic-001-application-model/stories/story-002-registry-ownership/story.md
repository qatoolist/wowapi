---
id: W05-E01-S002
type: story
title: Owner-bound registry wrappers across all declaration classes
status: planned
wave: W05
epic: W05-E01
owner: unassigned
reviewer: unassigned
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-12
source_requirements:
  - AR-01
depends_on:
  - W05-E01-S001
blocks:
  - W05-E01-S003
  - W05-E01-S004
acceptance_criteria:
  - AC-W05-E01-S002-01
  - AC-W05-E01-S002-02
  - AC-W05-E01-S002-03
artifacts: []
evidence: []
decisions: []
risks:
  - RISK-W05-001
  - RISK-W05-002
---

# W05-E01-S002 — Owner-bound registry wrappers across all declaration classes

## Story ID

W05-E01-S002

## Title

Owner-bound registry wrappers across all declaration classes

## Objective

Build owner-bound registrar wrappers, using S001's `Registrar` capability type, for
`resource.Registry`, `rules.Registry`, and — the widest gap — `authz.Registry` permission
registration (which today has no owner parameter at all), plus the remaining ~9+ declaration
classes (events, jobs, workflow actions, providers, templates, health checks, migrations, seeds,
OpenAPI), so that every registration surface in the framework is ownership-checked, not just the
three headline registries.

## Value to the framework

PLAN's own AR-01 T5 risk column names this story's central task as closing "the actual security
boundary" gap at its widest point: "`Register(p Permission)` currently has no owner parameter at
all — widest gap of the six." Without this story, S001's `Registrar` capability type is a boundary
with nothing enforcing it — a lock with no door. This story is what makes AR-01's security property
real across the framework's actual registration surface, not just a type that exists in isolation.

## Problem statement

`requirement-inventory.md` row AR-01 groups this story's scope explicitly: "S002 registry-ownership
(T3, T4, T5, T6)." PLAN's own AR-01 task table: T3 — "Owner-bound registrar wrapper for
`resource.Registry` | T1, T2 | `ctx.Resources()` exposes a registrar bound to the module's own
identity; ownership is structural, not string-compared | Adversarial: cross-module claim attempt
fails even with a matching key prefix | `AR-01/resource_ownership_adversarial_test.go` | Medium."
T4 — "Owner-bound registrar wrapper for `rules.Registry` | T1, T2 | Same shape as T3 for rule points
| Adversarial cross-owner rule-point claim | `AR-01/rules_ownership_adversarial_test.go` | Medium."
T5 — "Owner-bound registrar wrapper for `authz.Registry` permission registration | T1, T2 |
`Register(p Permission)` currently has no owner parameter at all — widest gap of the six; new API
derives module prefix from the bound registrar | Adversarial: cross-module permission claim rejected
at registrar boundary | `AR-01/authz_ownership_adversarial_test.go` | High — only registry with zero
existing ownership check." T6 — "Owner-bound registrar wrappers for the remaining ~9+ declaration
classes (events, jobs, workflow actions, providers, templates, health checks, migrations, seeds,
OpenAPI) | T1, T2, T3-T5 pattern | Every declaration class in AR-01's acceptance gate is
ownership-checked, not just the three headline registries | Table-driven adversarial suite, one
fixture per class | `AR-01/full_declaration_class_matrix_test.go` | Medium — easy to under-scope."

## Source requirements

AR-01 (T3, T4, T5, T6).

## Current-state assessment

Per PLAN's own evidence: `resource.Registry` and `rules.Registry` currently check ownership, where
checked at all, by string comparison rather than structural capability — `ctx.Resources()` does not
expose a registrar bound to the module's own identity today. `authz.Registry.Register(p Permission)`
has, per PLAN's own explicit statement, "no owner parameter at all" — this is a confirmed total
absence of ownership checking for that specific registry, the widest of the six registration
surfaces PLAN examines. The remaining ~9+ declaration classes (events, jobs, workflow actions,
providers, templates, health checks, migrations, seeds, OpenAPI) have not yet been individually
audited for their current ownership-check state within this programme's own analysis; this story's
own T6 re-confirmation step is to enumerate and audit each at implementation time.

## Desired state

`ctx.Resources()`, `ctx.Rules()` (or their equivalents), and the permission-registration API each
expose or require a registrar bound to the calling module's own identity, structurally, not by
string comparison — an adversarial cross-module claim attempt fails even with a matching key prefix,
proven by a dedicated test per registry. `authz.Registry`'s new permission-registration API derives
the module prefix from the bound registrar rather than accepting an arbitrary owner string. Every
declaration class in AR-01's own acceptance gate — not just the three headline registries — is
ownership-checked, proven by a table-driven adversarial suite with one fixture per class.

## Scope

- Owner-bound registrar wrapper for `resource.Registry` (T3).
- Owner-bound registrar wrapper for `rules.Registry` (T4).
- Owner-bound registrar wrapper for `authz.Registry` permission registration, including the new API
  shape that derives the module prefix from the bound registrar rather than an arbitrary owner
  string parameter (T5).
- Owner-bound registrar wrappers for the remaining ~9+ declaration classes: events, jobs, workflow
  actions, providers, templates, health checks, migrations, seeds, OpenAPI (T6).
- The table-driven adversarial suite proving every declaration class is ownership-checked (T6's own
  acceptance criterion).

## Out of scope

- **S001's `ApplicationModel`/`Compiler` lifecycle and `Registrar` capability type themselves** —
  already built; this story consumes them.
- **Snapshot immutability, post-seal rejection, model hash, race tests** — S003's scope.
- **The legacy compatibility adapter** — S004's scope; this story's wrappers are what the adapter
  must route through without bypassing.
- **Any change to `authz.Registry`'s actual permission-evaluation logic** — this story changes only
  the registration-time ownership check, not runtime authorization decisions (a separate concern,
  SEC-04's own scope in W05-E04-S002).

## Assumptions

- The exact enumeration of "the remaining ~9+ declaration classes" is taken directly from PLAN T6's
  own parenthetical list (events, jobs, workflow actions, providers, templates, health checks,
  migrations, seeds, OpenAPI) — nine named classes. PLAN's own phrasing "~9+" leaves open whether
  additional classes exist beyond this named list; this story's T6 task record requires an explicit
  audit against the framework's actual registration surface to confirm the count, not a silent
  assumption that exactly nine exist.
- T5's "new API derives module prefix from the bound registrar" is taken as a confirmed design
  direction from PLAN's own T5 acceptance-criteria column, not an invented detail — the exact
  function signature is this story's own implementation-time design work.

## Dependencies

Depends on W05-E01-S001 (T3-T6 all require T1's `ApplicationModel` and T2's `Registrar` capability
type to exist). Blocks W05-E01-S003 (T7's snapshot-immutability conversion applies to the registries
this story wraps) and W05-E01-S004 (T11's legacy adapter must route through this story's ownership
wrappers without bypassing them).

## Affected packages or components

`kernel/resource` (registry wrapper), `kernel/rules` (registry wrapper), `kernel/authz` (permission
registration API change — the widest-impact change in this story, since it alters an existing,
zero-owner-parameter API surface), plus the packages backing the remaining ~9+ declaration classes
(events, jobs, workflow, providers, templates, health checks, migrations, seeds, OpenAPI — exact
package list to be confirmed by T6's own audit).

## Compatibility considerations

T5's `authz.Registry.Register(p Permission)` API change is the story's most compatibility-sensitive
change, since it alters an existing zero-owner-parameter signature. Per this epic's own S004 legacy
adapter, existing callers must continue to compile and function through the legacy path — this
story's own T5 implementation must be designed so the legacy adapter (built afterward, in S004) can
route existing calls through the new ownership-bound API without requiring every existing caller to
be rewritten immediately.

## Security considerations

This entire story is a security story — closing the framework's remaining unowned or
string-compared registration surfaces. T5 is explicitly the highest-risk task in this story (PLAN's
own "High — only registry with zero existing ownership check"). See `risks.md` (epic-level) for
RISK-W05-001's full detail.

## Performance considerations

None material — registration-time, not request-hot-path, concern.

## Observability considerations

A rejected cross-module ownership claim (in any of the wrapped registries) should be observable
(logged, at minimum) at boot time, so a module author gets a clear diagnostic rather than a bare
compile or runtime failure — a reasonable implementation-time addition, consistent with this
programme's broader observability posture, though not separately mandated by PLAN's own T3-T6
acceptance criteria beyond the adversarial test requirements themselves.

## Migration considerations

None — no schema or data migration; this is registration-API surface work only.

## Documentation requirements

Document each wrapped registry's new ownership-bound API, with particular emphasis on T5's changed
`authz.Registry` permission-registration signature (the widest-impact change), and document the
full declaration-class enumeration T6 audits and wraps.

## Acceptance criteria

- **AC-W05-E01-S002-01**: `ctx.Resources()` and the equivalent for `rules.Registry` each expose a
  registrar bound to the module's own identity; a cross-module claim attempt fails even with a
  matching key prefix — proven by `AR-01/resource_ownership_adversarial_test.go` and
  `AR-01/rules_ownership_adversarial_test.go`.
- **AC-W05-E01-S002-02**: `authz.Registry`'s permission-registration API derives the module prefix
  from the bound registrar, not an arbitrary owner string; a cross-module permission claim is
  rejected at the registrar boundary — proven by `AR-01/authz_ownership_adversarial_test.go`.
- **AC-W05-E01-S002-03**: Every declaration class in AR-01's acceptance gate (events, jobs, workflow
  actions, providers, templates, health checks, migrations, seeds, OpenAPI, plus any additional class
  the T6 audit confirms) is ownership-checked, proven by a table-driven adversarial suite with one
  fixture per class — `AR-01/full_declaration_class_matrix_test.go`.

## Required artifacts

- Owner-bound `resource.Registry` wrapper (code).
- Owner-bound `rules.Registry` wrapper (code).
- Owner-bound `authz.Registry` permission-registration API (code).
- Owner-bound wrappers for the remaining ~9+ declaration classes (code).
- Declaration-class enumeration/audit record (T6).
See `artifacts/index.md`.

## Required evidence

- `AR-01/resource_ownership_adversarial_test.go` output.
- `AR-01/rules_ownership_adversarial_test.go` output.
- `AR-01/authz_ownership_adversarial_test.go` output.
- `AR-01/full_declaration_class_matrix_test.go` output.
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`: `story.md`
and `plan.md` complete, acceptance criteria numbered and measurable, dependency on S001 recorded,
owner/reviewer assignment pending, the T6 declaration-class enumeration recorded as an
implementation-time audit rather than a silently-assumed fixed list.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; all three acceptance
criteria verified with evidence in `evidence/index.md`; `closure.md` completed; independent review
passed per mandate §14, specifically confirming T5's adversarial test genuinely proves cross-module
permission-claim rejection and T6's declaration-class enumeration is genuinely complete against
AR-01's own acceptance gate.

## Risks

RISK-W05-001 (T5's previously-zero-ownership-check authz-registration gap) and RISK-W05-002 (T6's
explicit under-scoping risk) — see epic-level `risks.md` for full detail and mitigation/contingency.

## Residual-risk expectations

Once T5's and T6's adversarial tests are executed as planned and confirmed by independent review,
residual risk is expected to be low. T5's risk is inherently bounded by the specificity of its own
named test; T6's risk depends on the completeness of the declaration-class enumeration, which this
story's own review step is designed to catch if incomplete.

## Plan

See `plan.md`.
