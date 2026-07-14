---
id: PLAN-W06-E04-S001
type: plan
parent_story: W06-E04-S001
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
---

# Plan — W06-E04-S001

Per mandate §8.5. This plan follows MATRIX CS-22's own target-state/fix specification verbatim as its
architecture. Confirmed facts, planned changes, and assumptions are distinguished explicitly below.

## Proposed architecture

A small extractor tool, `internal/tools/docexamples`, that scans `docs/blueprint/*.md` and `README.md`
for fenced ` ```go ` blocks preceded by an HTML comment marker `<!-- doc-example: compile -->`, writes
each into a generated throwaway package, and invokes `go build` on it — build failure is the check
itself, per MATRIX CS-22's own framing ("`go/parser` not even needed — build failure is the check").

## Implementation strategy

1. Review the current normative doc set (`docs/blueprint/*.md`, `README.md`) for existing Go code
   examples; judge, per example, whether it is normative (should compile) or illustrative pseudo-code
   (should not).
2. Tag normative examples with `<!-- doc-example: compile -->`; leave pseudo-code untagged, by
   deliberate, visible choice.
3. Build the `internal/tools/docexamples` extractor tool.
4. Wire it as a CI step in the `unit` job.
5. Add a `make docs-check` target invoking the same check locally.
6. Write an adversarial fixture: a deliberately staled example (calling a removed symbol,
   resurrectable from the pre-AR-05 `RunAPI` text's git history per MATRIX CS-22's own suggestion) and
   confirm it fails the gate.
7. Confirm the current, corrected docs pass the gate.

## Expected package or module changes

New: `internal/tools/docexamples`. Extended: `Makefile` (new `docs-check` target); CI `unit` job
workflow configuration.

## Expected file changes where determinable

- `internal/tools/docexamples/` (new package).
- `Makefile` (new `docs-check` target).
- CI workflow configuration for the `unit` job (extended).
- `docs/blueprint/*.md` and `README.md` (markers added to normative examples).

## Contracts and interfaces

The `<!-- doc-example: compile -->` marker convention itself is the primary new "interface" — a
documentation-authoring convention, not a Go API.

## Data structures

None new at the framework level.

## APIs

None affected.

## Configuration changes

None.

## Persistence changes

None.

## Migration strategy

Not applicable.

## Concurrency implications

None.

## Error-handling strategy

The extractor must report clearly which specific doc example (file, line) failed to compile and why.

## Security controls

None new.

## Observability changes

None beyond the extractor's own clear-failure-reporting requirement.

## Testing strategy

- Positive: every currently-tagged normative example compiles.
- Negative (fail-first): a deliberately staled example (calling a removed symbol) fails the gate.
- `make docs-check` execution confirms the local invocation matches the CI invocation.

## Regression strategy

Once wired into the `unit` job as a required CI step, this gate becomes the ongoing regression guard
against any future doc edit reintroducing a phantom API reference in a tagged example.

## Compatibility strategy

Not applicable — this is additive documentation tooling.

## Rollout strategy

Single story, landed as its own reviewable unit.

## Rollback strategy

If the gate produces false positives against a legitimate example (e.g. an example that genuinely
compiles but the extractor mis-parses), fix the extractor directly; do not silently untag a genuinely
normative example merely to avoid a gate failure.

## Implementation sequence

As listed under "Implementation strategy" above (steps 1–7).

## Task breakdown

- **W06-E04-S001-T001** — Build the docexamples extractor tool and tag existing normative examples.
- **W06-E04-S001-T002** — Wire into CI (`unit` job) and `make docs-check`; write the adversarial staled-example
  fixture.
- **W06-E04-S001-T003** — Independent review.

## Expected artifacts

The `internal/tools/docexamples` extractor tool; the marker convention applied to existing examples; the
`make docs-check` target and CI wiring; the adversarial staled-example fixture.

## Expected evidence

Extractor-run output confirming every tagged example compiles; `make docs-check` execution output; the
staled-example fixture's fail-before/pass-after test output.

## Unresolved questions

- Exact tag-vs-untag judgment call for each currently-existing normative Go example — to be decided at
  implementation time, per example.

## Approval conditions

This plan is approved for implementation once the owner and reviewer are assigned.
