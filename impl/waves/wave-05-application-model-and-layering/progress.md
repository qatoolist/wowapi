---
id: W05-PROGRESS
type: wave-progress
wave: W05
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W05 progress (initial state)

Per mandate §16.2. Populated at programme-creation time; every item below is at its initial status.

## Epic status

| Epic | Title | Status | Stories | Story status breakdown |
|---|---|---|---|---|
| W05-E01 | application-model | planned | 4 | 4 planned |
| W05-E02 | typed-ports | planned | 3 | 3 planned |
| W05-E03 | authoritative-declarations | planned | 2 | 2 planned |
| W05-E04 | wiring-and-cache-hygiene | planned | 2 | 2 planned |
| W05-E05 | kernel-re-home | planned | 2 | 2 planned |

## Story status

| Story | Title | Status | Task count | Task status breakdown |
|---|---|---|---|---|
| W05-E01-S001 | model-and-registrar-capability | planned | 3 | 3 todo (incl. 1 independent-review task) |
| W05-E01-S002 | registry-ownership | planned | 5 | 5 todo (incl. 1 independent-review task) |
| W05-E01-S003 | snapshots-hash-race | planned | 4 | 4 todo |
| W05-E01-S004 | legacy-adapter | planned | 2 | 2 todo |
| W05-E02-S001 | port-api-and-forge-proofs | planned | 3 | 3 todo (incl. 1 independent-review task) |
| W05-E02-S002 | graph-validation-and-profiles | planned | 3 | 3 todo |
| W05-E02-S003 | lifecycle-manifest-retirement | planned | 2 | 2 todo |
| W05-E03-S001 | manifest-and-projections | planned | 5 | 5 todo (incl. 1 independent-review task) |
| W05-E03-S002 | boot-strictness-and-waivers | planned | 4 | 4 todo |
| W05-E04-S001 | constructor-bypass-closure | planned | 2 | 2 todo |
| W05-E04-S002 | authz-cache-bounding | planned | 6 | 6 todo (incl. 1 independent-review task) |
| W05-E05-S001 | foundation-move-and-shims | planned | 5 | 5 todo (incl. 1 independent-review task) |
| W05-E05-S002 | re-home-verification | planned | 3 | 3 todo (incl. 1 independent-review task) |

## Blocked items

None yet — no story has entered `in-progress`. Note for future readers: this wave's own entry is
gated on W03-E01 acceptance (see `dependencies.md`); internally, W05-E02 and W05-E03 are gated on
W05-E01 acceptance, and W05-E05 is gated on both W05-E01 and W05-E02 acceptance — these are recorded
as planned internal dependencies, not blocked items, until an upstream epic reaches `in-progress`
without a downstream epic waiting correctly.

## Critical dependencies

- W05 (full wave) depends on W03-E01 acceptance — `impl/analysis/wave-allocation-detail.md`'s
  explicit cross-wave sequencing note: "W05 entry requires W03-E01 acceptance (actor model
  stability)."
- W05-E02 and W05-E03 depend on W05-E01 (AR-01 T1/T2 are "the load-bearing prerequisite for AR-02's
  `Registrar` reuse and AR-03's manifest-consumes-model dependency," per PLAN's own PF-ARCH
  cross-cutting note).
- W05-E05 depends on W05-E01 and W05-E02 — MATRIX CS-01: "Dependencies: AR-01/02 first (re-homing
  mid-registration-rework causes double churn)."

## Open decisions

None new to W05 beyond the two already-ratified decisions this wave enacts: D-02, D-03 (enacted in
W05-E01-S001) and D-06 (enacted in W05-E04-S002) — all ratified in W00-E02-S003, referenced here,
not re-decided. No W05 story other than these two carries a `decisions/` directory — confirmed
against `requirement-inventory.md` §B (see `wave.md` "Assumptions").

## Open risks

See `risks.md`.

## Artifact completeness

0/13 story-level artifact sets populated.

## Evidence completeness

0 evidence records registered.

## Review state

Not yet reviewed.

## Exit-gate readiness

Not ready. 0 of 13 stories accepted.
