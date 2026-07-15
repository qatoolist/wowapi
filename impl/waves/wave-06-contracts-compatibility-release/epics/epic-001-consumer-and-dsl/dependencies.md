---
id: W06-E01-DEPS
type: epic-dependencies
epic: W06-E01
wave: W06
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W06-E01 — Dependencies

## Upstream (epics/waves this epic depends on)

- **W05** (full wave) — per `../../dependencies.md` (wave-level), W06 depends on W05's exit gate.
  DX-03-T0's own dependency row states "Wave 1 (AR-01 ApplicationModel, AR-02 typed ports) complete" —
  satisfied by W05-E01 (AR-01) and W05-E02 (AR-02) landing before this epic's S001 begins.
- **W01-E04-S001** (DX-01 T5 isolated-temp-dir scaffold harness) — this epic's S002 (DX-04 T1) reuses
  this harness as its shared primitive, per DX-01's own row note ("T5 harness = shared primitive for
  DX-02/DX-04") and per W01-E04-S001's own `story.md` forward reference to a "future DX-04 story."

## Downstream (epics/waves that depend on this epic)

| Downstream item | Depends on (from this epic) | Why |
|---|---|---|
| W06-E02-S003 (REL-03b, this wave) | W06-E01-S001 (DX-03 design) | MATRIX CS-15's own framing: REL-03 T5 ("event/schema compatibility") is "Blocked on DX-03/AR-03 — the concept doesn't exist in current source." |
| W06-E02-S003 (REL-03b, this wave) | W06-E01-S002 (DX-04) | PLAN REL-03 T7's own dependency: "Hard-blocked on DX-04." REL-03's generated-consumer upgrade check reuses DX-04's own drill. |

## Internal (within this epic)

S001 and S002 are independent of each other — DX-03's design work and DX-04's fixture-building target
disjoint surfaces (a design document vs. an installed-binary test fixture) and share no code-level
dependency. They may proceed in either order or in parallel, subject only to each satisfying its own
upstream dependency (S001 on W05's AR-01/AR-02; S002 on W01-E04-S001's harness).

## Cross-wave dependencies

W01-E04-S001 (see "Upstream" above) is the epic's one cross-wave dependency beyond the W05 entry gate.

## External dependencies

None new. DX-04's fixture uses the same Postgres/MinIO/Mailpit/OTel infrastructure already used
elsewhere in the framework's integration-test tier; no new external service is introduced.

## Repository dependencies

None cross-repo for this epic's own closure. DX-03 has zero wowsociety impact by explicit directive
design constraint (PLAN: "wowsociety's identity/policy modules compile unmodified through Waves 1-3").
DX-04 has zero wowsociety code-change impact (PLAN: "No direct code change required. wowsociety's
`framework-verify` and richer hand-completed modules are a useful secondary signal but explicitly not a
substitute for the CI-authoritative fixture").

## Tooling dependencies

None beyond the already-available Go toolchain, `go install`, and the CI infrastructure DX-04 T5 wires
into (`ci/release-gates.yaml` at its Wave-4 boundary, REL-01 — this epic's S002 only wires the gate
reference; REL-01's own manifest mechanics are W06-E03-S001's scope).

## Decision dependencies

None. See `epic.md` "Required decisions."
