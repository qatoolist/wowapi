---
id: W06-E04
type: epic
title: Documentation gates
status: in-progress
wave: W06
owner: unassigned
reviewer: unassigned
priority: medium
created_at: 2026-07-12
updated_at: 2026-07-12
source_requirements:
  - AR-05
depends_on: []
stories:
  - W06-E04-S001
  - W06-E04-S002
decisions: []
risks: []
---

# W06-E04 — Documentation gates

## Epic objective

Build a CI-enforced doc-example compile gate (AR-05 T3, per MATRIX CS-22's full mechanics spec) so that
a normative Go example in the framework's own documentation is proven to compile, not merely asserted
correct in prose; generate reference/API docs from AR-03's authoritative manifest so they byte-match the
model export (AR-05 T4); and label any remaining future-state design prose as "target, not implemented"
via a lint (AR-05 T5).

## Problem being solved

MATRIX CS-22's own evidence: "documentation drift has already happened twice at reviewer-visible
severity: `README.md:148-153`/blueprint 11 described phantom `RunAPI`/`RunWorker`/`RunMigrate` APIs, and
blueprint 06 listed five `Context` methods that don't exist (both fixed by AR-05 T1/T2 at `345e4ce`,
verified §D). Nothing prevents recurrence: zero `//go:generate` directives repo-wide, no generated-code-
currency check, no doc-example compile gate in any workflow or Makefile." MATRIX CS-22's own defect
framing: "normative Go examples in `docs/blueprint/*.md` and `README.md` are prose — they can silently
rot against the live API, and did." `requirement-inventory.md` row AR-05 confirms T1/T2 are already
`EXECUTED`, with T3 (this epic's S001) and T4/T5 (this epic's S002, dependent on AR-03) as the remaining
scope.

## Scope

- AR-05 T3 (MATRIX CS-22's full spec): the `internal/tools/docexamples` extractor tool; the
  `<!-- doc-example: compile -->` marker convention; a `make docs-check` target; an adversarial staled-
  example fixture (S001).
- AR-05 T4: generated reference docs byte-matching AR-03's authoritative model export (S002).
- AR-05 T5: a lint labeling remaining future-state design prose as "target, not implemented" (S002).

## Out of scope

- **AR-05 T1, T2** — already `EXECUTED` per `requirement-inventory.md` (the README/blueprint drift fixes
  at commit `345e4ce`); this epic does not re-do that work, only re-confirms it still holds if relevant
  to T5's own lint scope.
- **AR-03's own manifest work** — W05-E03's scope; this epic's S002 consumes AR-03's model export, it
  does not build it.

## Source requirements

AR-05 (T3, T4, T5). MATRIX CS-22 is the consolidated closure spec for T3 specifically.

## Architectural context

This epic groups AR-05's remaining T3/T4/T5 scope because all three concern the same underlying
property: the framework's documentation must be provably correct, not merely asserted correct.
`impl/analysis/wave-allocation-detail.md`'s own W06-E04 grouping states this exactly: "S001
doc-example-compile-gate (CS-22/AR-05 T3 spec); S002 generated-docs-and-labels (AR-05 T4, T5 — dep
E02/W05-E03 manifest)." This two-way split (T3's compile-gate mechanics standing alone; T4/T5's
generated-docs-and-labeling work depending on AR-03) is fixed by the canonical allocation.
`requirement-inventory.md`'s own AR-05 row lists its overall target as "W06-E04-S002" while
`wave-allocation-detail.md` gives the more precise per-task split (T3→S001, T4/T5→S002) — this tree
follows `wave-allocation-detail.md` as the canonical per-epic/story authority, per the task brief's own
explicit instruction.

## Included stories

- **W06-E04-S001 — doc-example-compile-gate** (PLAN AR-05 T3, MATRIX CS-22's full mechanics spec): the
  `internal/tools/docexamples` extractor, marker convention, `make docs-check`, adversarial fixture.
- **W06-E04-S002 — generated-docs-and-labels** (PLAN AR-05 T4, T5): generated reference docs matching
  AR-03's model export; future-state-labeling lint.

## Dependencies

No dependency on any other W06 epic for S001's own entry. **S002 depends cross-wave on W05-E03** (AR-03
T1/T5, per AR-05 T4's own dependency row: "dep AR-03" — AR-03's own target story is W05-E03, this wave's
own upstream). This epic depends transitively on this wave's own W05 entry gate.

## Risks

None recorded at this epic's own scope. AR-05's own PLAN risk classification for T3 is "Medium — needs
new doc-example-extraction tooling"; for T4 is "Medium — depends on AR-03"; for T5 is "Low."

## Required decisions

None. AR-05 carries no D-0N architecture-decision dependency in `requirement-inventory.md` §B or REVIEW
§F/§U.

## Epic acceptance criteria

- **AC-W06-E04-01**: The doc-example-compile-gate runs in CI via `make docs-check`; every tagged
  normative example compiles; a deliberately staled example (calling a removed symbol) fails the gate.
- **AC-W06-E04-02**: Generated reference docs byte-match AR-03's authoritative model export, once AR-03
  (W05-E03) is `accepted`.
- **AC-W06-E04-03**: A lint over `docs/blueprint/` fails on an unlabeled normative-sounding future-state
  block.
- **AC-W06-E04-04**: Both stories have passed independent review per mandate §14.

## Closure conditions

Both stories reach `accepted` (S002's own closure is gated on W05-E03 having reached `accepted` first,
per its own cross-wave dependency); AC-W06-E04-01 through AC-W06-E04-04 above are all satisfied;
`closure-report.md` for this epic is completed with reviewer conclusion and acceptance date.

## Status update (2026-07-16)

`status: in-progress` — S001 (doc-example compile gate) accepted (artifact-presence level). S002
(generated docs and labels) accept-with-conditions: T5 (labeling/lint) is fully implemented,
evidenced, and reviewed with no W05 dependency; T4 (generate reference/API docs from AR-03's
manifest) remains open/blocked pending W05-E03 reaching `accepted` — see this story's own
`closure.md` scoping note. Epic cannot reach `accepted` while S002's T4 remains open.

— dated 2026-07-16, conductor adjudication (Fable 5), per review-gate-2026-07-16.md records
