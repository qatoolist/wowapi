---
id: W06-E02-S001
type: story
title: OpenAPI merge complete-or-loud — full-field merge, validation, semantic diff
status: verified
wave: W06
epic: W06-E02
owner: W06E02Impl
reviewer: W06-E01-E04-Execution.W06E02ReviewFinal
priority: P1
created_at: 2026-07-12
updated_at: 2026-07-13
source_requirements:
  - DX-06
  - AR-03
depends_on: []
blocks:
  - W06-E02-S003
acceptance_criteria:
  - AC-W06-E02-S001-01
  - AC-W06-E02-S001-02
  - AC-W06-E02-S001-03
  - AC-W06-E02-S001-04
artifacts:
  - ART-W06-E02-S001-001
  - ART-W06-E02-S001-002
  - ART-W06-E02-S001-003
  - ART-W06-E02-S001-004
evidence:
  - EV-W06-E02-S001-001
  - EV-W06-E02-S001-002
  - EV-W06-E02-S001-003
  - EV-W06-E02-S001-004
  - EV-W06-E02-S001-005
decisions:
  - DEC-W06-E02-S001-VALIDATOR
risks:
  - RISK-W06-E02-001
---

# W06-E02-S001 — OpenAPI merge complete-or-loud — full-field merge, validation, semantic diff

## Story ID

W06-E02-S001

## Title

OpenAPI merge complete-or-loud — full-field merge, validation, semantic diff

## Objective

Expand the OpenAPI merge struct to cover every OpenAPI 3.1 top-level field and every `components.*`
field with an explicit per-field merge policy, validate the merged document against 3.1.1/2020-12, and
gate a semantic diff against DX-05's v1 policy — closing DX-06's own defect and, by single ownership per
`impl/analysis/conflict-resolution.md` CONFLICT-01, the identical AR-03 T2 closure contract. AR-03's own
target story (W05-E03) proceeds without T2; this story owns that scope in full.

## Value to the framework

MATRIX CS-15 states the defect's consequence bluntly: "a module declaring `security` on a fragment ships
an API with that requirement silently absent from the published contract — a *security-adjacent*
documentation lie; breaking API changes are detectable only by consumers breaking." This story converts
a merge mechanism that silently discards most of the OpenAPI 3.1 specification's top-level surface into
one that either merges every field correctly or explicitly rejects a fragment it cannot merge — the only
two acceptable outcomes for a framework claiming to publish a correct API contract. It also resolves a
genuine duplicate-effort risk: PLAN's own §7 cross-cutting note #11 states "AR-03 T2 and DX-06 T1
(OpenAPI full-field merge) — identical closure contract," and this story is the programme's single
assigned owner of that shared contract, preventing the same merge-completeness work from being built
twice under two different finding names.

## Problem statement

MATRIX CS-15's evidence: "`internal/cli/openapi_cmd.go:139-144` `mergeFragment` unmarshals only `paths`
+ `components.schemas` into an anonymous struct — all other OpenAPI 3.1 top-level fields silently
dropped (duplicate paths/schemas *do* fail loudly, `:148-158` — the loud-on-collision half already
exists); zero `apidiff`/`gorelease` hits in Makefile/CI/docs; zero `/vN` path versioning." PLAN's own
DX-06 evidence confirms the same defect from the plan's own reading: "`openapi_cmd.go`'s merge-target
struct captures only `Paths` and `Components.Schemas` — every other top-level 3.1 field (`security`,
`tags`, `servers`, `webhooks`, callbacks, non-schema `components.*`) is silently discarded by
`json.Unmarshal` with no error or warning." `impl/analysis/conflict-resolution.md` CONFLICT-01 confirms
the duplicate: "AR-03 T2 vs DX-06 (T1) — Duplicate scope — both findings independently specify the
identical closure contract for OpenAPI full-field merge, complete-or-loud behaviour... Resolution:
Single owner DX-06."

## Source requirements

DX-06 (T1–T3); AR-03 (T2 scope only, owned via CONFLICT-01 — AR-03's own T1/T3/T4/T5 remain W05-E03's
scope, not this story's).

## Current-state assessment

Per PLAN's own evidence and MATRIX CS-15 (to be re-confirmed at this story's own execution commit):
`internal/cli/openapi_cmd.go:139-144`'s `mergeFragment` merge-target struct captures only `Paths` and
`Components.Schemas`. Every other OpenAPI 3.1 top-level field — `security`, `tags`, `servers`,
`webhooks`, callbacks, and non-schema `components.*` entries — is silently discarded by
`json.Unmarshal` with no error or warning, because the anonymous struct has no field to receive them.
The loud-on-collision half of the merge already exists and works correctly: duplicate paths/schemas do
fail loudly (`:148-158`). Zero `apidiff`/`gorelease` hits exist anywhere in the Makefile, CI, or docs.
No OpenAPI 3.1.1/2020-12 structural validator is wired into the merge command today.

## Desired state

The merge-target struct covers every OpenAPI 3.1 top-level field (`security`, `tags`, `servers`,
`webhooks`, `paths`, all of `components.*`, and any other 3.1 top-level field) with an explicit,
documented per-field merge policy — each field either merges correctly according to its own semantics
(e.g. `tags` might union by name, `security` might require identical declarations across fragments or
fail loudly on conflict) or the fragment is explicitly rejected with a clear error naming the offending
field, never silently dropped. The final merged document is validated against OpenAPI 3.1.1 and JSON
Schema 2020-12 before being accepted as output. A semantic-diff gate, keyed to DX-05's already-ratified
v1/N-1 compatibility policy, fails an intentional breaking-change fixture.

## Scope

- **T1** — Expand the merge struct to cover all OpenAPI 3.1 top-level fields plus `components.*`, with
  an explicit per-field merge policy; a fixture-driven test suite, one fragment per field type, proving
  each field either merges correctly or is explicitly rejected.
- **T2** — Validate the final merged document against OpenAPI 3.1.1 / JSON Schema 2020-12; select and
  wire in a validator dependency (candidate: `pb33f/libopenapi`, per MATRIX CS-15 — decided at
  implementation time with security/licence review, not pre-selected here).
- **T3** — Semantic API diffing gated to DX-05's already-ratified v1 policy; a seeded intentional
  breaking-change fixture must fail the gate.
- **AR-03 T2's identical contract** — owned in full by T1/T2 above; AR-03's own target story (W05-E03)
  does not implement a second version of this merge-completeness work.

## Out of scope

- **AR-03's T1, T3, T4, T5** (the authoritative manifest and its derived projections beyond the merge-
  completeness contract) — W05-E03's own scope, not touched here.
- **REL-03 T3** (OpenAPI semantic diff as a compatibility-gate task) — this story's own T3 *is* the
  semantic-diff mechanism; W06-E02-S003's REL-03b T3 leg consumes this story's output as its unblocking
  dependency, it does not duplicate the mechanism.
- **`/vN` path versioning** — MATRIX CS-15 records this as a zero-hit evidence item but does not assign
  it a task in DX-06's own T1–T3 table; not implemented here.

## Assumptions

- The exact per-field merge policy for each newly-covered field (e.g. whether `security` fragments must
  be identical across modules or merge by union, whether `webhooks` merge by name with collision
  rejection) is not specified by any source document beyond MATRIX CS-15's framing that "every field
  either merges correctly or the fragment is explicitly rejected." This story's own T1 design work
  determines the exact policy per field, recorded in `plan.md` as an implementation-time decision, not
  invented here.
- The OpenAPI 3.1 validator dependency is not yet selected — MATRIX CS-15 names `pb33f/libopenapi` only
  as an "evaluate" candidate, with an explicit "decision at implementation, security-review licence"
  caveat. This story's T2 records the decision as its own task, not a pre-made choice.
- DX-05's v1 policy (the compatibility-class definition T3's semantic diff is keyed to) is already
  ratified and accepted at W01-E04-S002 — this story consumes that policy, it does not re-derive it.

## Dependencies

None within W06-E02 for this story's own entry (it is the epic's first story in dependency order).
Depends transitively on this wave's W05 entry gate and, for T3's semantic-diff policy, on DX-05's
already-`accepted` v1 policy from W01. Blocks W06-E02-S003's T3 leg (REL-03b's OpenAPI semantic diff,
MATRIX CS-15: "Blocked on DX-06 — a lossy merge can't be meaningfully diffed").

## Affected packages or components

`internal/cli/openapi_cmd.go` (the merge-target struct and `mergeFragment` function); a new or extended
validation step wired into the same command; a new semantic-diff CI job (exact location TBD).

## Compatibility considerations

This story's own semantic-diff gate (T3) is itself a compatibility-enforcement mechanism, keyed to
DX-05's already-ratified v1 policy — it does not introduce a new compatibility model, it enforces the
existing one against the OpenAPI surface specifically. The merge-struct expansion (T1) is a strict
correctness improvement: any fragment field that was previously silently dropped is now either
correctly merged or explicitly rejected — no prior "working" behavior depended on the silent drop, since
a silently-dropped field was never actually present in the merged output to begin with.

## Security considerations

MATRIX CS-15's own framing makes this story security-adjacent: a module declaring `security` on a
fragment that gets silently dropped ships an API whose published contract lies about its own security
requirements. T1's per-field merge policy for `security` specifically must not silently drop a
declared requirement — this is a required correctness property of this story's own acceptance bar, not
an optional hardening add-on.

## Performance considerations

Not applicable — this is a build-time/CI-time tool, not a runtime request path.

## Observability considerations

The merge command should report, per field, whether it merged or was rejected, so a module author
understands exactly what happened to their fragment — an implementation-time detail, not separately
mandated by acceptance criteria beyond the "explicitly rejected... never silently dropped" requirement
itself.

## Migration considerations

Not applicable — no schema or data migration is involved.

## Documentation requirements

Document the per-field merge policy for every OpenAPI 3.1 top-level field and `components.*` entry, so
a module author knows in advance what happens to each field type in their fragment; document the
validator dependency choice and its rationale once T2's decision is made; document the semantic-diff
gate's behavior and how it relates to DX-05's v1 policy.

## Acceptance criteria

- **AC-W06-E02-S001-01**: The merge struct covers every OpenAPI 3.1 top-level field and every `components.*`
  field; a fixture-driven test suite (one fragment per field type) proves each field either merges
  correctly per its documented policy or the fragment is explicitly rejected with a field-specific
  error — no field is silently dropped.
- **AC-W06-E02-S001-02**: The final merged document is validated against OpenAPI 3.1.1 / JSON Schema 2020-12; a
  malformed merged output fails the command.
- **AC-W06-E02-S001-03**: A seeded, intentional breaking-API-change fixture fails the semantic-diff gate, which
  is keyed to DX-05's already-ratified v1 policy.
- **AC-W06-E02-S001-04**: The validator-dependency decision (T2) is recorded with its security/licence review
  outcome, before being wired into the merge command as a hard dependency.

## Required artifacts

- The expanded merge struct with per-field policy (T1).
- The fixture-driven per-field test suite (T1).
- The 3.1.1/2020-12 structural validator wiring (T2).
- The semantic-diff CI gate (T3).
- Per-field merge-policy documentation.
See `artifacts/index.md`.

## Required evidence

- Per-field-type fixture test output, one fragment per field type (T1).
- Structural-validation test output, including a malformed-output negative fixture (T2).
- Seeded-breaking-fixture semantic-diff test output (T3).
- The validator-dependency security/licence review record (T2).
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`: `story.md`
and `plan.md` complete, all four acceptance criteria numbered and measurable, no dependency within this
epic, owner/reviewer assignment pending, the per-field merge-policy design and the validator-dependency
choice recorded as unresolved questions rather than silently assumed.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; all four acceptance
criteria verified with evidence in `evidence/index.md`; `closure.md` completed; independent review
passed per mandate §14, specifically confirming no OpenAPI 3.1 field is silently dropped by the expanded
merge (re-testing against the full field list, not trusting T1's own self-reported coverage) and that
the validator-dependency decision genuinely received a security/licence review before being wired in.

## Risks

RISK-W06-E02-001 (the OpenAPI validator dependency decision made without adequate security/licence
review if rushed) — see epic-level `risks.md` for full detail and mitigation/contingency.

## Residual-risk expectations

Once T1's per-field policy design, T2's reviewed validator selection, and T3's DX-05-keyed diff gate are
all verified, residual risk is expected to be low — this is a well-bounded, source-derived closure story
with a clear MATRIX CS-15 acceptance bar and no ambiguity about which fields must be covered (the
OpenAPI 3.1 specification's own top-level field list is a closed, enumerable set).

## Plan

See `plan.md`.
