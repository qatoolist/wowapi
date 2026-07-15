---
id: W06-E04-S001
type: story
title: Doc-example compile gate — CI-enforced normative Go example compilation
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
depends_on: []
blocks: []
acceptance_criteria:
  - AC-W06-E04-S001-01
  - AC-W06-E04-S001-02
  - AC-W06-E04-S001-03
artifacts:
  - ART-W06-E04-S001-001
  - ART-W06-E04-S001-002
  - ART-W06-E04-S001-003
  - ART-W06-E04-S001-004
evidence:
  - EV-W06-E04-S001-001
  - EV-W06-E04-S001-002
  - EV-W06-E04-S001-003
  - EV-W06-E04-S001-003-R1
  - REV-W06-E04-S001-001
decisions: []
risks: []
---

# W06-E04-S001 — Doc-example compile gate — CI-enforced normative Go example compilation

## Story ID

W06-E04-S001

## Title

Doc-example compile gate — CI-enforced normative Go example compilation

## Objective

Build a small extractor tool (`internal/tools/docexamples`) that scans the normative doc set for fenced
` ```go ` blocks tagged with a `<!-- doc-example: compile -->` marker, writes each into a generated
throwaway package, and `go build`s them; wire it as a CI step in the `unit` job and a `make docs-check`
target; prove it with an adversarial fixture — a deliberately staled example (calling a removed symbol)
that must fail the gate.

## Value to the framework

MATRIX CS-22's own evidence names two concrete instances of the exact failure mode this story exists to
prevent: "`README.md:148-153`/blueprint 11 described phantom `RunAPI`/`RunWorker`/`RunMigrate` APIs, and
blueprint 06 listed five `Context` methods that don't exist (both fixed by AR-05 T1/T2 at `345e4ce`,
verified §D). Nothing prevents recurrence." MATRIX CS-22's own consequence framing: "a consumer following
the documented API writes code that doesn't compile; reviewer trust in all docs drops to zero after the
first phantom API." This story converts documentation from a prose claim that can silently rot into a
CI-enforced property — the only way to prevent AR-05 T1/T2's fix from being undone by drift a third time.

## Problem statement

MATRIX CS-22's own evidence: "zero `//go:generate` directives repo-wide, no generated-code-currency
check, no doc-example compile gate in any workflow or Makefile (toolchain inventory, all zero-hit-
verified)." PLAN's own AR-05 T3 task row: "CI gate compiling every normative doc example against the
current API | T1, T2 | A deliberately staled example fails CI | Adversarial CI fixture |
`AR-05/doc_compile_ci_gate_test_output.txt` | Medium — needs new doc-example-extraction tooling." MATRIX
CS-22's own target-state/fix specification is the authoritative mechanics spec this story implements
verbatim: "a small extractor tool (`internal/tools/docexamples`) that scans the normative doc set for
fenced ` ```go ` blocks tagged normative (an HTML comment marker above the fence, e.g. `<!-- doc-example:
compile -->`, so illustrative pseudo-code opts out explicitly), writes each into a generated throwaway
package, and `go build`s them; wired as a CI step in the `unit` job and a `make docs-check` target."

## Source requirements

AR-05 (T3). MATRIX CS-22 is the consolidated closure spec.

## Current-state assessment

Per MATRIX CS-22's own evidence (to be re-confirmed at this story's own execution commit): zero
`//go:generate` directives repo-wide; no generated-code-currency check; no doc-example compile gate in
any workflow or Makefile — all confirmed zero-hit via toolchain inventory. AR-05 T1/T2's own README/
blueprint drift fixes already landed at commit `345e4ce` (verified per REVIEW §D), but nothing today
prevents a future doc edit from reintroducing a phantom API reference.

## Desired state

Every normative Go code example in `docs/blueprint/*.md` and `README.md` is either tagged with
`<!-- doc-example: compile -->` (and genuinely compiles, verified in CI) or is illustrative pseudo-code
that explicitly opts out of the marker (and is therefore not checked, by deliberate, visible choice, not
by accidental omission). A `make docs-check` target runs the extractor and fails if any tagged example
fails to compile. A deliberately staled example (a fixture calling a removed symbol, resurrectable from
the git history of the pre-AR-05 `RunAPI` text per MATRIX CS-22's own suggestion) fails the gate,
proving the gate actually works, not merely exists.

## Scope

- Build the `internal/tools/docexamples` extractor: scans the normative doc set for `<!-- doc-example:
  compile -->`-tagged fenced Go blocks.
- Write each tagged block into a generated throwaway package and `go build` it.
- Wire the extractor as a CI step in the `unit` job.
- Add a `make docs-check` target invoking the same check locally.
- Write an adversarial fixture: a deliberately staled example (calling a removed symbol) that fails the
  gate, proving fail-first.
- Tag the current normative Go examples in `docs/blueprint/*.md` and `README.md` with the marker
  (or explicitly leave non-normative pseudo-code untagged).

## Out of scope

- **AR-05 T4** (generated reference docs byte-matching AR-03's model export) — W06-E04-S002's own scope,
  dependent on AR-03.
- **AR-05 T5** (future-state-labeling lint) — W06-E04-S002's own scope; this story's own extractor tool
  is distinct tooling from that lint, even though both are documentation-quality gates.
- **Re-fixing AR-05 T1/T2's own already-executed drift corrections** — not this story's scope; this
  story only builds the gate preventing recurrence, it does not re-audit the already-fixed content
  beyond what tagging requires.

## Assumptions

- The exact set of currently-existing normative Go examples that should be tagged `<!-- doc-example:
  compile -->` versus left as untagged illustrative pseudo-code is not pre-determined by any source
  document — MATRIX CS-22's own framing ("so illustrative pseudo-code opts out explicitly") implies a
  judgment call per example, to be made at implementation time, not invented here.
- The extractor's own exact implementation approach (MATRIX CS-22's own estimate: "~150 LOC... stdlib
  (`go/parser` not even needed — build failure is the check)") is confirmed from source as a scale/
  approach guideline, not an exact LOC target this story must hit precisely.

## Dependencies

None — this story has no dependency on any other story in this wave or an earlier wave beyond the
transitive W05 entry gate. It does not depend on W06-E04-S002 or on AR-03/W05-E03.

## Affected packages or components

New: `internal/tools/docexamples` (the extractor tool). Extended: the CI `unit` job workflow
configuration; the `Makefile` (new `docs-check` target). Modified: `docs/blueprint/*.md` and
`README.md` (adding `<!-- doc-example: compile -->` markers to normative examples).

## Compatibility considerations

Tagging existing doc examples with the marker is additive and non-breaking. Once the gate is wired into
the `unit` job as a required check, any future doc edit introducing a phantom API reference in a tagged
example will newly fail CI — this is the intended, correct behavior change this story exists to produce.

## Security considerations

Not directly applicable — this is a documentation-tooling story with no runtime security surface.

## Performance considerations

Not applicable — this is a CI-time gate operating on a small, bounded set of doc examples.

## Observability considerations

The gate should report clearly which specific doc example failed to compile and why, so a documentation
author can fix it without needing to understand the extractor tool's own internals.

## Migration considerations

Not applicable.

## Documentation requirements

Document the `<!-- doc-example: compile -->` marker convention itself, so a future documentation author
knows how to tag a new normative example (or deliberately leave a pseudo-code example untagged).

## Acceptance criteria

- **AC-W06-E04-S001-01**: The `internal/tools/docexamples` extractor scans the normative doc set, writes each
  tagged example into a generated throwaway package, and `go build`s it; every currently-tagged
  normative example compiles.
- **AC-W06-E04-S001-02**: `make docs-check` exists and invokes the same check locally; the extractor is wired as
  a CI step in the `unit` job.
- **AC-W06-E04-S001-03**: A deliberately staled example (calling a removed symbol) fails the gate, proving
  fail-first; the current, corrected docs pass.

## Required artifacts

- The `internal/tools/docexamples` extractor tool.
- The `<!-- doc-example: compile -->` marker convention, applied to existing normative examples.
- The `make docs-check` target and CI wiring.
- The adversarial staled-example fixture.
See `artifacts/index.md`.

## Required evidence

- Extractor-run output confirming every tagged example compiles.
- `make docs-check` execution output.
- The staled-example fixture's fail-before/pass-after test output.
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`: `story.md`
and `plan.md` complete, all three acceptance criteria numbered and measurable, no dependency, owner/
reviewer assignment pending, the exact tag-vs-untag judgment call for existing examples recorded as an
unresolved question rather than silently pre-decided.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; all three acceptance
criteria verified with evidence in `evidence/index.md`; `closure.md` completed; independent review
passed per mandate §14, specifically confirming the staled-example fixture genuinely fails the gate (not
merely claimed to) and that no normative example was left deliberately untagged to avoid a compile
failure rather than because it is genuinely illustrative pseudo-code.

## Risks

None recorded at this story's own scope — this is a well-bounded, source-derived closure story with a
clear MATRIX CS-22 mechanics spec and a small, stdlib-only implementation surface.

## Residual-risk expectations

Once all three acceptance criteria are verified, residual risk is expected to be low.

## Plan

See `plan.md`.
