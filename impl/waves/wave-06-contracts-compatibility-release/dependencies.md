---
id: W06-DEPS
type: wave-dependencies
wave: W06
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W06 — Dependencies

## Upstream (waves this wave depends on)

- **W05** — full wave dependency per `impl/index.md`'s wave map row for W06 ("Depends on: W05 (AR-03
  unblocks REL-03b legs)"). W06 consumes W05's AR-03 remainder (authoritative declaration + derived
  projections, W05-E03-S001..S002) as the specific unblocking dependency for W06-E02-S003's T5 leg and
  as a prerequisite for W06-E01-S001's DX-03 design work (PLAN DX-03-T0's own dependency: "Wave 1
  (AR-01 ApplicationModel, AR-02 typed ports) complete" — satisfied transitively by W05's own AR-01/
  AR-02 stories landing before AR-03).

## Downstream (waves that depend on this wave)

| Downstream item | Depends on (from W06) | Why |
|---|---|---|
| W07 (full wave) | W06 (full wave) | `impl/index.md`'s wave map states W07 "Depends on: all prior" — the strict W00→W07 entry ordering places W06 immediately before W07's entry gate. |
| W07-E02-S002 (REL-04 coverage-truthfulness) | W06-E03-S003 (REL-02 blocking scans) | REL-04's own T5's manifest wiring assumes REL-01/REL-02's gate manifest already exists and enforces blocking behavior, consistent with REL-02 T5's own task ("Wire all REL-02 blocking checks into REL-01's Wave-0 manifest"). |

## Cross-wave dependencies

- **W01-E04-S001** (DX-01 T5 isolated-temp-dir scaffold harness) — W06-E01-S002 (DX-04) reuses this
  harness as its shared primitive rather than rebuilding it, per DX-01's own row note in
  `requirement-inventory.md` ("T5 harness = shared primitive for DX-02/DX-04") and per W01-E04-S001's
  own `story.md` "Dependencies" section, which records this exact forward reference.
- **W00-E02-S003** (ADR-ification story) — W06-E03-S001's T6 (GoReleaser split-mode) consumes
  `ADR-W00-E02-S003-005` (D-05) by reference; this wave does not mint a new ADR for that decision.

## Internal (within this wave, between epics)

- **W06-E02 depends partly on W06-E01.** W06-E02-S003's T5 leg depends on W06-E01-S001 (DX-03 design)
  in addition to W05-E03 (AR-03 remainder, cross-wave). W06-E02-S003's T7 leg depends on W06-E01-S002
  (DX-04).
- **W06-E02-S003 depends on W06-E02-S001** (its own epic-sibling) for its T3 leg — DX-06's merge-
  complete-or-loud closure must be `accepted` before REL-03's OpenAPI semantic-diff task (T3) can begin,
  per MATRIX CS-15's own framing: "T3 (Blocked on DX-06 — a lossy merge can't be meaningfully diffed)."
- **W06-E03 has no internal dependency on W06-E01/E02** — REL-01's exact-commit release pipeline
  (S001), its human-gated activation remainder (S002), and REL-02's blocking-security-scans (S003)
  target disjoint CI/workflow surface from the DX-03/DX-04/DX-06/REL-03 work in E01/E02, and may
  proceed in parallel with them once this wave's own entry gate (W05) is satisfied.
- **W06-E04 has no internal dependency on W06-E01/E02/E03** for its S001 (doc-example-compile-gate);
  W06-E04-S002 (generated-docs-and-labels) depends cross-wave on W05-E03 (AR-03 T1/T5) for its T4 leg,
  recorded explicitly below.
- **W06-E04-S002 depends cross-wave on W05-E03** (AR-03's authoritative manifest, T1/T5) — AR-05 T4's
  own acceptance criterion is "generated reference tables byte-match the model export," which requires
  AR-03's model export to exist first. This crosses back to W05, recorded explicitly here per this
  wave's task brief instruction not to let a cross-wave dependency go unstated merely because the
  target wave has already closed.

## External dependencies

None new for W06-E01/E04. W06-E02-S001 (DX-06 T2) introduces a new external dependency candidate — an
OpenAPI 3.1 validator (MATRIX CS-15 names `pb33f/libopenapi` as the evaluation candidate) — not yet
approved; the decision is recorded as an implementation-time task, not assumed. W06-E03 introduces no
new external dependency beyond GoReleaser, already in use and governed by ADR-005.

## Repository dependencies

None cross-repo for this wave's own closure. wowsociety impact is real but non-blocking:

- **DX-06** — PLAN's own recommended follow-up: "audit `wowsociety/internal/modules/*/openapi.json` for
  silently-dropped fields once T1's stricter validator ships." Not required for this wave's closure.
- **REL-01/REL-02** — confirmed not affected: wowsociety's CI checks out wowapi via plain `git checkout`
  at a pinned SHA, never invoking wowapi's release pipeline via `workflow_call`.
- **REL-03** — affected at consumption/upgrade time only, per PLAN's own framing: "wowsociety's CI never
  invokes REL-03's gates... What changes is *what a wowapi release means*." Not a blocking dependency
  for this wave.
- **AR-05** — not affected; pure wowapi-internal documentation.

## Tooling dependencies

None new beyond the OpenAPI validator (DX-06 T2, undecided) and the existing GitHub Actions/GoReleaser/
Trivy toolchain already in use elsewhere in the framework's CI.

## Decision dependencies

DEC-Q10 (repo-admin activation) directly gates W06-E03-S002's entry. ADR-005 (D-05, GoReleaser
split-mode) is already ratified at W00 and is consumed, not re-decided, by W06-E03-S001.
