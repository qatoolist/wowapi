---
id: W06-E03-DEPS
type: epic-dependencies
epic: W06-E03
wave: W06
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W06-E03 — Dependencies

## Upstream (epics/waves this epic depends on)

- **W00-E02-S003** (ADR-005, `ADR-W00-E02-S003-005`) — already ratified; S001's T6 consumes this
  decision, does not re-decide it.
- **W05** (full wave, transitively) — per this wave's own entry gate.

## Downstream (epics/waves that depend on this epic)

| Downstream item | Depends on (from this epic) | Why |
|---|---|---|
| W06-E02-S002-T005 (container architecture smoke, this wave) | This epic's S001 (build-candidate split) | REL-03 T8 must run against the candidate image produced by REL-01 T6/T7, not an already-published image. |
| W06-E02-S002-T006 (SBOM/provenance fold-in, this wave) | This epic's S001 (REL-01 T8/T9) | REL-03 T9's own framing: "Folds directly into REL-01 T8/T9 — not separate work." |
| W07-E02-S002 (REL-04 coverage-truthfulness) | This epic's S003 (REL-02 manifest wiring) | REL-04's T5 manifest work assumes REL-01/REL-02's gate manifest already exists and enforces blocking behavior. |

## Internal (within this epic)

- **S002 has no code-level dependency on S001** beyond consuming the same overall REL-01 finding — S002
  is REL-01's *remainder* (T9), not a consumer of S001's T1-T8 output in the code sense, though
  practically the activation only makes sense once S001's pipeline exists to protect.
- **S003 has no dependency on S001 or S002 for its own T1-T4** (Trivy flip, waiver schema, visibility-
  guard check, private-repo fallback) — these target Trivy/CodeQL/Scorecard configuration, disjoint from
  REL-01's own pipeline mechanics. **S003's T5 (manifest wiring) depends on S001's T1/T2** (the manifest
  schema and Wave-0 entries must exist before REL-02's checks can be wired into it).

## Cross-wave dependencies

None beyond W00-E02-S003 (ADR-005) and the transitive W05 entry gate.

## External dependencies

GoReleaser (already in use, governed by ADR-005's split-mode decision). Trivy (already in use, S003
flips its blocking behavior). No new external service.

## Repository dependencies

None cross-repo for this epic's own closure. REL-01/REL-02 confirmed not affected for wowsociety: its
CI checks out wowapi via plain `git checkout` at a pinned SHA, never invoking wowapi's release pipeline
via `workflow_call`. If REL-01/REL-02 work ever flips wowapi private again, wowsociety's checkout would
need `WOWAPI_CHECKOUT_TOKEN` populated — a cross-repo coordination note, not a code change this epic
must make.

## Tooling dependencies

None new beyond GitHub Actions, GoReleaser, and Trivy — all already in use.

## Decision dependencies

ADR-005 (already ratified). DEC-Q10 (open, human-blocked, tracked by S002).
