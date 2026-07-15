---
id: W06-E04-ACCEPTANCE
type: epic-acceptance
epic: W06-E04
wave: W06
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W06-E04 — Epic-level acceptance

Reproduces and elaborates `epic.md`'s "Epic acceptance criteria" section as a standalone,
independently-referenceable acceptance record, consistent with the wave-level `../../acceptance.md`
pattern (AC-W06-07 there maps onto this epic).

## AC-W06-E04-01 — Doc-example compile gate enforced in CI

The `internal/tools/docexamples` extractor scans the normative doc set for `<!-- doc-example: compile
-->`-tagged fenced Go blocks, writes each into a generated throwaway package, and `go build`s them; wired
as a CI step in the `unit` job and via `make docs-check`. A deliberately staled example (calling a
removed symbol) fails the gate. Traces to W06-E04-S001.

## AC-W06-E04-02 — Generated reference docs byte-match the model export

Once W05-E03 (AR-03) is `accepted`, generated reference tables byte-match the model export, proven by an
integration golden-diff test. Traces to W06-E04-S002.

## AC-W06-E04-03 — Future-state prose labeled, not silently presented as implemented

A lint over `docs/blueprint/` fails on an unlabeled normative-sounding future-state block. Traces to
W06-E04-S002.

## AC-W06-E04-04 — Independent review passed

Both S001 and S002 have passed independent review per mandate §14.

## Acceptance authority

Developer-experience lead, consistent with PF-DX's own accountable role (AR-05 traces back to PF-ARCH's
own cross-cutting documentation concern, but its practical delivery — CLI/generator-adjacent doc
tooling — sits with the same developer-experience lead accountable for DX-01 through DX-07).
