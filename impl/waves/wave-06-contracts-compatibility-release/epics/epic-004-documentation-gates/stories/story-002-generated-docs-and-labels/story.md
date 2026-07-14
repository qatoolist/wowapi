---
id: W06-E04-S002
type: story
title: Generated docs and labels — model-export byte-match and future-state labeling lint
status: accepted
wave: W06
epic: W06-E04
owner: W06E04Impl
reviewer: W06-E01-E04-Execution.W06E04ReviewR
priority: P2
created_at: 2026-07-12
updated_at: 2026-07-13
source_requirements:
  - AR-05
depends_on:
  - W05-E03
blocks: []
acceptance_criteria:
  - AC-W06-E04-S002-01
  - AC-W06-E04-S002-02
artifacts:
  - ART-W06-E04-S002-001
  - ART-W06-E04-S002-002
evidence:
  - EV-W06-E04-S002-001
  - EV-W06-E04-S002-002
  - REV-W06-E04-S002-001
decisions: []
risks:
  - RISK-W06-E04-001
---

# W06-E04-S002 — Generated docs and labels — model-export byte-match and future-state labeling lint

## Story ID

W06-E04-S002

## Title

Generated docs and labels — model-export byte-match and future-state labeling lint

## Objective

Generate reference/API documentation from AR-03's authoritative manifest so the generated reference
tables byte-match the model export (AR-05 T4), and build a lint labeling remaining future-state design
prose as "target, not implemented" (AR-05 T5).

## Value to the framework

AR-05 T3 (W06-E04-S001) proves normative *code examples* compile; this story extends the same
"documentation must be provably correct" principle to two further documentation classes: reference/API
tables (which must match the framework's own authoritative model, not a hand-maintained copy that can
drift) and future-state design prose (which must be visibly distinguished from currently-implemented
behavior, so a reader does not mistake a design aspiration for a shipped capability — exactly the class
of confusion DX-03's own design record, W06-E01-S001, must itself avoid triggering).

## Problem statement

PLAN's own AR-05 task table: "T4. Generate reference/API docs from AR-03's authoritative manifest | AR-03
T1, T5 | Generated reference tables byte-match the model export | Integration golden-diff |
`AR-05/generated_docs_byte_match_test.go` | Medium — depends on AR-03." "T5. Label remaining future-state
design prose as 'target, not implemented' | T1-T4 | Lint over `docs/blueprint/` for unlabeled
normative-sounding future-state blocks | Lint | `AR-05/future_state_labeling_lint_test.go` | Low."
`requirement-inventory.md`'s own AR-05 row notes: "T4/T5 dep AR-03." Neither mechanism exists today:
there is no generated-reference-doc pipeline consuming AR-03's model export (because AR-03's own
manifest work, W05-E03, has not yet landed at this wave's own planning time), and no lint exists to
catch an unlabeled future-state design block in `docs/blueprint/`.

## Source requirements

AR-05 (T4, T5).

## Current-state assessment

Per PLAN's own evidence (to be re-confirmed at this story's own execution commit): no generated-
reference-doc pipeline exists; no future-state-labeling lint exists. T4's own dependency (AR-03 T1, T5)
is not yet satisfied at this wave's own planning time — W05-E03 (AR-03's remainder) is this wave's own
upstream dependency, landing in the prior wave.

## Desired state

Once W05-E03 (AR-03) is `accepted`: generated reference tables byte-match AR-03's own model export,
proven by an integration golden-diff test — the reference docs are generated *from* the authoritative
source, not hand-maintained separately and hoped to stay in sync. A lint runs over `docs/blueprint/`
and fails on any unlabeled normative-sounding future-state block, so a design-only concept (like DX-03's
own module-DSL design record, W06-E01-S001) is always visibly distinguished from shipped behavior.

## Scope

- **T4** — Generate reference/API docs from AR-03's authoritative manifest (once W05-E03 is `accepted`);
  an integration golden-diff test proving byte-match.
- **T5** — A lint over `docs/blueprint/` (and, by extension, any other future-state design documents
  this programme produces, including W06-E01-S001's own DX-03 design record) failing on an unlabeled
  normative-sounding future-state block.

## Out of scope

- **AR-03's own manifest implementation** — W05-E03's scope; this story consumes AR-03's model export,
  it does not build it.
- **AR-05 T1, T2, T3** — T1/T2 already `EXECUTED`; T3 is W06-E04-S001's own scope, not duplicated here.
- **Retroactively re-labeling every existing document in the repository** — T5's lint is scoped to
  `docs/blueprint/` per PLAN's own acceptance criterion; whether the lint's scope should extend further
  (e.g. to `impl/` itself, which this very generation task produces future-state-labeled content for,
  such as W06-E01-S001's design record) is recorded as an implementation-time scoping decision, not
  invented here.

## Assumptions

- T4 cannot begin implementation until W05-E03 (AR-03) reaches `accepted` — this is a hard, structural
  dependency, not a scheduling preference; this story's own T4 task is recorded with an explicit
  blocked-entry criterion, per this epic's own risk register (RISK-W06-E04-001).
- T5's own lint can proceed independently of T4, since it does not depend on AR-03's model export — only
  on the existing `docs/blueprint/` content and this programme's own labeling convention (already
  established practice, e.g. "target, not implemented" as used by W06-E01-S001's own DX-03 design
  record).
- The exact "generated reference/API docs" format and pipeline mechanism (a `go generate` directive, a
  standalone tool, a Makefile target) is not specified by any source document beyond "generate reference/
  API docs from AR-03's authoritative manifest" — recorded as an implementation-time decision.

## Dependencies

**T4 depends on W05-E03 (AR-03, cross-wave) reaching `accepted`.** T5 has no dependency on W05-E03 or on
this story's own T4 — it may proceed independently. No dependency on W06-E04-S001 (this epic's sibling
story) — the two stories' mechanisms are distinct (compile-checking code examples vs. generating
reference docs and linting labeling).

## Affected packages or components

New: a reference-doc-generation pipeline (exact location TBD, dependent on AR-03's own delivered
package structure); a future-state-labeling lint (exact location TBD, likely alongside other lint
tooling in `internal/tools/` or `scripts/`).

## Compatibility considerations

T4's generated docs, once wired in, replace any hand-maintained reference documentation that duplicates
AR-03's model — this is a strict correctness improvement (generated-from-source cannot drift the way
hand-maintained content can), not a breaking change to any code surface.

## Security considerations

Not applicable — documentation tooling only.

## Performance considerations

Not applicable.

## Observability considerations

The lint (T5) should report clearly which specific unlabeled block triggered the failure and its exact
location, so a documentation author can add the missing label without needing to interpret the lint's
own internals.

## Migration considerations

Not applicable.

## Documentation requirements

Document the reference-doc-generation pipeline's own invocation (T4) and the future-state-labeling
convention the lint enforces (T5), so both are discoverable by a future documentation author.

## Acceptance criteria

- **AC-W06-E04-S002-01**: Once W05-E03 (AR-03) is `accepted`, generated reference tables byte-match the model
  export, proven by an integration golden-diff test. If W05-E03 is not yet `accepted` at this story's
  own closure attempt, this AC is recorded as deferred with the unblocking condition (W05-E03's
  acceptance) restated, not silently dropped.
- **AC-W06-E04-S002-02**: A lint over `docs/blueprint/` fails on an unlabeled normative-sounding future-state
  block; correctly-labeled blocks (e.g. content following the "target, not implemented" convention) pass.

## Required artifacts

- The reference-doc-generation pipeline (T4, once unblocked).
- The future-state-labeling lint (T5).
See `artifacts/index.md`. T4's artifact is recorded as "not yet produced — blocked" if W05-E03 has not
landed.

## Required evidence

- Integration golden-diff test output proving byte-match (T4).
- Lint fixture test output, unlabeled-block-fails / labeled-block-passes (T5).
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`: `story.md`
and `plan.md` complete, both acceptance criteria numbered and measurable, T4's dependency on W05-E03
recorded explicitly, owner/reviewer assignment pending. Per this story's own non-standard readiness
posture (mirroring W06-E02-S003's own framing), T5 may independently become `ready` for implementation
without waiting for T4's own entry criterion.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted` (in full) or
`partially-accepted` (if T4 remains blocked): T5 is implemented, evidenced, and independently reviewed;
T4 is implemented, evidenced, and independently reviewed if its entry criterion (W05-E03 `accepted`) was
satisfied during this story's execution window, or recorded in `closure.md` as deferred-with-restated-
unblocking-condition if not.

## Risks

RISK-W06-E04-001 (T4 may remain blocked past this story's own closure attempt if W05-E03 is delayed) —
see epic-level `risks.md` for full detail and mitigation/contingency.

## Residual-risk expectations

T5's residual risk is expected to be low once its own acceptance criterion is verified. T4's residual
risk cannot be fully eliminated within this story's own scope — it depends on W05-E03's own completion
timing, tracked honestly via partial-acceptance status if not satisfied by this story's own closure
attempt.

## Plan

See `plan.md`.
