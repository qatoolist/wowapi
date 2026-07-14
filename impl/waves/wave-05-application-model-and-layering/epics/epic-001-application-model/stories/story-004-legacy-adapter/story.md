---
id: W05-E01-S004
type: story
title: Legacy module/context compatibility adapter
status: planned
wave: W05
epic: W05-E01
owner: unassigned
reviewer: unassigned
priority: high
created_at: 2026-07-12
updated_at: 2026-07-12
source_requirements:
  - AR-01
depends_on:
  - W05-E01-S003
blocks:
  - W05-E02-S001
  - W05-E05-S001
acceptance_criteria:
  - AC-W05-E01-S004-01
  - AC-W05-E01-S004-02
artifacts: []
evidence: []
decisions: []
risks:
  - RISK-W05-E01-003
---

# W05-E01-S004 — Legacy module/context compatibility adapter

## Story ID

W05-E01-S004

## Title

Legacy module/context compatibility adapter

## Objective

Build the legacy adapter wrapping the current `module.Module`/`Context` surface so existing modules
(wowapi-internal and wowsociety) compile and boot unchanged through S001-S003's ownership-bound,
immutable, race-safe `ApplicationModel`, with the adapter itself deriving owner from
`Module.Name()` and routing through the same owner-bound registrars as the non-legacy path — never
bypassing S002's ownership checks.

## Value to the framework

Without this story, every existing module in wowapi and wowsociety would need to be rewritten
against the new ownership-bound API before this epic's work could land — an all-or-nothing cutover
PLAN's own Wave-1 framing explicitly avoids. This story is the compatibility strategy that lets
AR-01's security and correctness properties ship without forcing an immediate, coordinated rewrite
of every consuming module.

## Problem statement

`requirement-inventory.md` row AR-01 groups this story's scope: "S004 legacy-adapter (T11 —
compatibility story)." PLAN's own AR-01 task table: T11 — "Legacy adapter wrapping current
`module.Module`/`Context` so existing modules compile unchanged (Wave 1 compatibility strategy) |
T1-T10 | Existing modules (wowapi internal + wowsociety) boot unchanged through the adapter; the
adapter itself derives owner from `Module.Name()` and routes through the same owner-bound registrar
— it must not bypass T2-T6 | Integration: existing contract tests pass unmodified through the legacy
path | `AR-01/legacy_adapter_compat_test_output.txt` | Medium — the adapter is itself a trust
boundary."

## Source requirements

AR-01 (T11).

## Current-state assessment

Per this epic's own S001-S003, the ownership-bound `ApplicationModel`, `Registrar` capability type,
per-registry ownership wrappers, immutability, post-seal rejection, model hash, and race safety all
now exist — but no compatibility path yet exists for the current `module.Module`/`Context` surface
to route through them. Every existing module today registers through the pre-AR-01 surface; this
story's own re-confirmation step is to re-read the current `module.Module`/`Context` implementation
at this story's actual start commit and confirm it has not itself changed since S001-S003 began.

## Desired state

Existing modules (wowapi-internal and wowsociety) boot unchanged through the legacy adapter — no
source change required in any existing module. The adapter derives each module's owner from
`Module.Name()` and routes every registration call through the same owner-bound registrars S002
built, never bypassing their ownership checks. Existing contract tests pass unmodified through the
legacy path.

## Scope

- The legacy adapter implementation: wraps `module.Module`/`Context` and routes registration calls
  through S001-S003's ownership-bound, immutable model.
- Owner derivation from `Module.Name()` for each module routed through the adapter.
- Confirmation that the adapter does not bypass any ownership check established by S002 — proven by
  running S002's own adversarial fixtures through the legacy path, not only the non-legacy path.
- The integration test proving existing contract tests pass unmodified through the legacy path.

## Out of scope

- **Migrating any existing module off the legacy path onto the new ownership-bound API directly** —
  this story provides the compatibility bridge; it does not require or perform any module's
  migration. Per PLAN's own wowsociety-impact note for AR-01: "No wowsociety change required
  before/during Wave 1 landing; cleanup is low-risk and can happen on wowsociety's own schedule."
- **AR-02's typed port-key API's own legacy adapter** — a separate, later concern (W05-E02-S003's
  own legacy port-adapter task, AR-02 T7), not built here.

## Assumptions

- The set of "existing modules" this story must prove compatible against includes both
  wowapi-internal modules and wowsociety's modules, per PLAN T11's own acceptance criterion
  ("wowapi internal + wowsociety"). wowsociety's own contract tests are consumed as a regression
  check (verification dependency), not a code dependency this story modifies.

## Dependencies

Depends on W05-E01-S003 (T11 depends on T1-T10 in full — the legacy adapter wraps the complete,
race-safe, deterministic-hash model). Blocks W05-E02-S001 (AR-02's own port system, at wave scope,
is built to work alongside the legacy adapter's compatibility strategy) and W05-E05-S001 (FBL-01's
kernel re-home is sequenced after this epic completes in full, including this compatibility story).

## Affected packages or components

The current `module.Module`/`Context` package(s) (exact location: existing `kernel/module`) — wrapped
by the new adapter, not replaced. No existing module's own source code is modified by this story.

## Compatibility considerations

This story IS the compatibility mechanism for AR-01's entire epic — its own compatibility
consideration is its acceptance criterion: existing modules boot unchanged, existing contract tests
pass unmodified.

## Security considerations

PLAN's own risk note: "the adapter is itself a trust boundary." An adapter that derives owner
incorrectly, or that bypasses S002's ownership checks for convenience, would silently reintroduce
the exact unowned-registration gap this epic exists to close, just hidden behind a compatibility
shim — see RISK-W05-E01-003 in epic-level `risks.md`.

## Performance considerations

None material — boot-time compatibility wrapping, not a request-hot-path concern.

## Observability considerations

None beyond the epic's own boot-time logging conventions (S001's `collect → validate → seal`
transition logging, which the legacy adapter's routed calls pass through).

## Migration considerations

None — no schema or data migration; this is a compatibility-shim story.

## Documentation requirements

Document the legacy adapter's owner-derivation mechanism (`Module.Name()`-based) and its guarantee
that it routes through, never bypasses, S002's ownership checks — so a future reader auditing the
adapter's trust-boundary status has a clear reference.

## Acceptance criteria

- **AC-W05-E01-S004-01**: Existing modules (wowapi-internal and wowsociety) boot unchanged through
  the legacy adapter — proven by existing contract tests passing unmodified through the legacy path
  (`AR-01/legacy_adapter_compat_test_output.txt`).
- **AC-W05-E01-S004-02**: The adapter derives owner from `Module.Name()` and routes through the same
  owner-bound registrars as the non-legacy path — proven by running S002's own adversarial fixtures
  (resource, rules, authz, and the full declaration-class matrix) through the legacy path and
  confirming identical rejection behavior to the non-legacy path.

## Required artifacts

- The legacy adapter implementation (code).
- Legacy-adapter documentation.
See `artifacts/index.md`.

## Required evidence

- `AR-01/legacy_adapter_compat_test_output.txt`.
- Adversarial-fixture-through-legacy-path test output (AC-02's own proof).
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`: `story.md`
and `plan.md` complete, acceptance criteria numbered and measurable, dependency on S003 recorded,
owner/reviewer assignment pending.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; both acceptance
criteria verified with evidence in `evidence/index.md`; `closure.md` completed; independent review
passed per mandate §14, specifically confirming the adapter does not bypass any S002 ownership check
— PLAN's own "the adapter is itself a trust boundary" framing is the direct basis for adding review
to this story despite its otherwise Medium risk profile.

## Risks

RISK-W05-E01-003 (the legacy adapter, as a trust boundary, could silently bypass S002's ownership
checks if incorrectly implemented) — see epic-level `risks.md` for full detail and mitigation/
contingency.

## Residual-risk expectations

Residual risk is expected to be low once the adversarial-fixtures-through-legacy-path proof (AC-02)
is executed and confirmed by independent review — this is exactly the mechanism designed to catch a
bypassing adapter before it ships.

## Plan

See `plan.md`.
