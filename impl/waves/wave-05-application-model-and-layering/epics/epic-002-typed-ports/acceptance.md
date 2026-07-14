---
id: W05-E02-ACCEPTANCE
type: epic-acceptance
epic: W05-E02
wave: W05
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W05-E02 — Epic-level acceptance

Reproduces and elaborates `epic.md`'s "Epic acceptance criteria" section as a standalone,
independently-referenceable acceptance record.

## AC-W05-E02-01 — Port-key API and registrar-forge safety proven

`port.Key[T]` and the four generic free functions compile and resolve correctly bound to W05-E01's
`Registrar`; the internal compiler factory mints registrars with immutable owner identity; the
adversarial compile-fail fixture (`AR-02/registrar_forge_compile_fail_fixture/`) proves capability
confusion is impossible. Traces to W05-E02-S001.

## AC-W05-E02-02 — Zero-reflection graph, boot-time validation, and profile projection proven

Zero `reflect.*` calls occur at `Resolve` time, proven by benchmark and static lint; boot-time
validation rejects duplicate providers, missing requirements, undeclared edges, cycles, and invalid
scope/lifetime edges, one adversarial fixture per failure class; API/worker/migrate profiles compile
as three projections of one graph, no hand-copied wiring template remains. Traces to W05-E02-S002.

## AC-W05-E02-03 — Lifecycle manifest retired; legacy adapter compatible

The hand-maintained `kernel/lifecycle` manifest is retired in favor of the generated graph, existing
lint-failure classes pass, now data-driven; the legacy port adapter compiles/resolves unchanged
(confirmed zero external callers). Traces to W05-E02-S003.

## AC-W05-E02-04 — Independent review passed

All three stories have passed independent review per mandate §14. S001's review specifically
confirms T2's capability-confusion compile-fail fixture is genuine, not merely claimed.

## Acceptance authority

Framework architecture lead, per `../../wave.md`'s wave-level acceptance authority.
