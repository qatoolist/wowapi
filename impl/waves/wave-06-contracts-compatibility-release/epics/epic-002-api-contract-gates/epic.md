---
id: W06-E02
type: epic
title: API contract gates
status: in-progress
wave: W06
owner: unassigned
reviewer: unassigned
priority: high
created_at: 2026-07-12
updated_at: 2026-07-12
source_requirements:
  - DX-06
  - AR-03
  - REL-03
depends_on:
  - W06-E01
stories:
  - W06-E02-S001
  - W06-E02-S002
  - W06-E02-S003
decisions: []
risks:
  - RISK-W06-E02-001
---

# W06-E02 — API contract gates

## Epic objective

Make the framework's OpenAPI merge either capture every 3.1 field or fail loudly instead of silently
discarding them (DX-06), resolving the identical AR-03 T2 closure contract by single ownership per
`impl/analysis/conflict-resolution.md` CONFLICT-01; and build the full compatibility-gate programme
(REL-03), split honestly into a buildable-now half (REL-03a) and a still-blocked half (REL-03b) with
explicit per-leg unblocking criteria, per PLAN's own recommendation not to schedule REL-03 "as one
monolithic P1 item, or 5 of 9 sub-tasks silently block the other 4."

## Problem being solved

`requirement-inventory.md` row DX-06 states: "OpenAPI merge complete-or-loud (T1-T3) | IMPL | P1 |
planned | W06-E02-S001 | Single owner of AR-03 T2 scope; validator dep decision at impl." MATRIX CS-15
gives the exact evidence: "`internal/cli/openapi_cmd.go:139-144` `mergeFragment` unmarshals only
`paths` + `components.schemas` into an anonymous struct — all other OpenAPI 3.1 top-level fields
silently dropped (duplicate paths/schemas *do* fail loudly, `:148-158` — the loud-on-collision half
already exists); zero `apidiff`/`gorelease` hits in Makefile/CI/docs; zero `/vN` path versioning." The
defect's consequence, per MATRIX CS-15: "a module declaring `security` on a fragment ships an API with
that requirement silently absent from the published contract — a *security-adjacent* documentation
lie; breaking API changes are detectable only by consumers breaking." Row REL-03 states: "Compatibility
gates (split a/b) | QG | P1 | planned | W06-E02-S002..S003 | a=T1,T2,T4,T6,T8,T9 now; b=T3(DX-06),
T5(AR-03/DX-03),T7(DX-04)." PLAN's own recommendation is explicit: "Recommend splitting into REL-03a
(buildable now: Go API diff, module compile matrix, config compat, migration-upgrade drill, arch smoke,
SBOM/provenance-verify) and REL-03b (hard-blocked on Wave 1/4 architecture work...) — do not schedule as
one monolithic P1 item."

## Scope

- DX-06 T1–T3: the full-field OpenAPI merge struct with per-field policy, 3.1.1/2020-12 structural
  validation, and a semantic-diff gate keyed to DX-05's v1 policy (S001). This story also owns AR-03
  T2's identical closure contract, per CONFLICT-01's resolution — AR-03's own target story (W05-E03)
  proceeds without T2.
- A validator-dependency decision task (evaluate `pb33f/libopenapi` per MATRIX CS-15, decided at
  implementation time with security/licence review — not pre-decided by this planning document) (S001).
- REL-03a T1, T2, T4, T6, T8, T9: Go public API diff, module compile matrix, config-schema
  compatibility, migration upgrade-from-oldest-supported drill, container architecture smoke, SBOM/
  provenance/signature verification fold-in from REL-01 T8/T9 (S002).
- REL-03b T3, T5, T7: OpenAPI semantic diff (blocked on DX-06), event/schema compatibility (blocked on
  DX-03/AR-03), generated-consumer upgrade check (blocked on DX-04) — recorded with explicit per-leg
  blocked-entry criteria naming the exact unblocking story (S003).

## Out of scope

- **AR-03's own T1, T3, T4, T5** (manifest and derived projections beyond the merge-completeness
  contract) — W05-E03's own scope, not duplicated here; this epic owns only the identical merge-
  completeness contract AR-03 T2 and DX-06 T1 both independently specified.
- **DX-03's implementation** — W06-E01-S001 produces only a design record, not implementation; REL-03b's
  T5 leg remains blocked on that design record plus AR-03's remainder regardless.
- **REL-01's own gate-manifest mechanics** — W06-E03-S001's scope; REL-03a T9's SBOM/provenance-verify
  fold-in shares evidence with REL-01 T8/T9 but does not itself build the manifest schema.

## Source requirements

DX-06, AR-03 (T2 scope only, via CONFLICT-01), REL-03. MATRIX CS-15 is the consolidated closure spec
covering all three.

## Architectural context

This epic groups DX-06 and REL-03 because both concern proving the framework's *published API
contract* is correct and stable — one at the OpenAPI-document level (DX-06), the other at the Go
public-API/config/migration/container/consumer-upgrade level (REL-03). `impl/analysis/wave-allocation-
detail.md`'s own W06-E02 grouping states this exactly: "S001 openapi-merge-complete-or-loud (DX-06
T1–T3; owns AR-03 T2 scope; validator dependency decision task); S002 compat-gates-buildable-now
(REL-03a: T1, T2, T4, T6, T8, T9); S003 compat-gates-unblocked (REL-03b: T3, T5, T7 — entry criteria
reference their unblocking stories)." This three-way split (S001 for the merge-completeness contract
itself, S002 for what REL-03 can build today, S003 for what REL-03 cannot build yet) is fixed by the
canonical allocation and is not to be regrouped.

## Included stories

- **W06-E02-S001 — openapi-merge-complete-or-loud** (PLAN DX-06 T1–T3; owns AR-03 T2 via CONFLICT-01):
  full-field OpenAPI merge, structural validation, semantic diff; validator-dependency decision task.
- **W06-E02-S002 — compat-gates-buildable-now** (PLAN REL-03a: T1, T2, T4, T6, T8, T9): the six
  compatibility-gate tasks buildable without any unresolved upstream dependency.
- **W06-E02-S003 — compat-gates-unblocked** (PLAN REL-03b: T3, T5, T7): the three still-blocked
  compatibility-gate tasks, each with an explicit per-leg entry criterion naming its unblocking story.

## Dependencies

Depends on W06-E01 within this wave (S003's T5 leg depends on W06-E01-S001's DX-03 design; S003's T7 leg
depends on W06-E01-S002's DX-04). Depends on W05-E03 cross-wave (AR-03 remainder, for S003's T5 leg).
S003's T3 leg depends internally on this epic's own S001 (DX-06). This epic depends on W06's own W05
entry gate transitively.

## Risks

RISK-W06-E02-001 (the DX-06 T2 validator-dependency decision made without adequate security/licence
review if rushed) originates at this epic's S001. See `risks.md` for the epic-scoped elaboration; the
REL-03b scheduling risk (RISK-W06-003) is tracked at wave scope and lands entirely within this epic's
S003.

## Required decisions

None in the D-0N architecture-decision sense — DX-06/AR-03/REL-03 carry no D-0N dependency in
`requirement-inventory.md` §B or REVIEW §F/§U. The DX-06 T2 validator choice is an implementation-time
task decision, not a programme-level D-0N ADR, and does not warrant a `decisions/` directory under
S001 (it is recorded directly in S001's own task record once made, per this story's own plan).

## Epic acceptance criteria

- **AC-W06-E02-01**: The OpenAPI merge struct covers every 3.1 top-level field and `components.*` field
  with explicit per-field merge policy; the merged document validates against 3.1.1/2020-12; a seeded
  breaking-API fixture fails the semantic-diff gate. AR-03's own target story proceeds without its T2
  task per CONFLICT-01.
- **AC-W06-E02-02**: REL-03a's six tasks (Go API diff, compile matrix, config compat, migration
  upgrade-drill, arch smoke, SBOM/provenance-verify fold-in) are complete and evidenced.
- **AC-W06-E02-03**: REL-03b's three legs are recorded with explicit per-leg blocked-entry criteria; any
  leg that unblocks during this epic's execution is completed and evidenced; any leg still blocked at
  this epic's closure is recorded as deferred-with-restated-unblocking-condition, not silently dropped.
- **AC-W06-E02-04**: S001 and S002 have passed independent review per mandate §14; S003's review (for
  whichever legs unblock and complete) specifically confirms the still-blocked legs' entry criteria are
  honestly stated, not silently bypassed.

## Closure conditions

S001 and S002 reach `accepted`; S003 reaches `accepted` or `partially-accepted` depending on how many of
its three legs unblocked during this epic's execution, per `governance/definition-of-done.md`'s
partially-accepted status; AC-W06-E02-01 through AC-W06-E02-04 above are satisfied to the extent each
story's own status allows; `closure-report.md` for this epic is completed with reviewer conclusion and
acceptance date; any still-blocked REL-03b leg is recorded with its exact unblocking condition restated,
not silently dropped.

## Status update (2026-07-16)

`status: in-progress` — S001 (OpenAPI merge complete-or-loud) and S002 (compat gates buildable now)
both independently reviewed and accepted per `review-gate-2026-07-16.md`. S003 (compat gates
unblocked) remains genuinely `blocked` — its blocking dependencies (E02-S001, E01-S001 + W05-E03,
E01-S002) are confirmed genuine and unresolved as of this review; the `blocked` classification is
accurate, not a defect.

— dated 2026-07-16, conductor adjudication (Fable 5), per review-gate-2026-07-16.md records
