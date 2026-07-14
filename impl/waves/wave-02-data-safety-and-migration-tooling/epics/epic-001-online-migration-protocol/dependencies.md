---
id: W02-E01-DEPS
type: epic-dependencies
epic: W02-E01
wave: W02
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W02-E01 — Dependencies

## Upstream (epics/waves this epic depends on)

- **W00** (full wave) — per `../../dependencies.md` (wave-level), W02 depends on W00's exit gate.
  This epic has no additional upstream dependency beyond that gate — DATA-09 has no D-0N ADR
  dependency and no dependency on any W01 finding.

## Downstream (epics/waves that depend on this epic)

| Downstream item | Depends on (from this epic) | Why |
|---|---|---|
| W02-E02 (DATA-01, this wave) | W02-E01-S001 + W02-E01-S002 (protocol through validation-phase tooling) | DATA-01 T4/T5 (the `NOT VALID` composite-FK add and its `VALIDATE CONSTRAINT`) are gated on this epic's S001+S002 acceptance, per `impl/analysis/wave-allocation-detail.md`'s explicit cross-wave sequencing note. |
| W03-E01-S001 (SEC-01 grant-table migration) | This epic's protocol (S001–S003) | `impl/index.md`'s wave map: W03 depends on "W02 (grant-table migration uses DATA-09)." |
| W04-E01-S001 (DATA-02 shared lease primitive) | W02-E01-S002 (interim checkpoint lease) | W04-E01-S001 is scoped to specifically replace S002's interim lease with the full shared primitive — a supersession dependency, not merely a "builds on" dependency. See RISK-W02-001. |
| W04-E04-S001 (DATA-08 W6-T1 audit-hash migration) | This epic's protocol | `impl/analysis/wave-allocation-detail.md`'s W04-E04-S001 row: "dep W02-E01 protocol." |

## Internal (within this epic)

S001 → S002 → S003 form a strict phase-pipeline dependency, matching PLAN DATA-09's own task
dependency chain (T1 → T2; T3 depends on T1; T4 depends on T3 and DATA-02 T1; T5 depends on T4;
T6 depends on T5; T7 depends on T6; T8 depends on T7; T9 depends on T1–T8). Concretely:

- S002 depends on S001 (T3 depends on T1's manifest schema being in place to classify the expand
  migration itself).
- S003 depends on S002 (T6 depends on T5's validation-phase tooling; T9 depends on all of T1–T8).

Unlike W01-E01's three internally-parallel stories, this epic's three stories are **not**
independently orderable — they must execute S001 → S002 → S003 in sequence, because each phase's
tooling is a genuine prerequisite for the next phase's tooling to have anything to operate on.

## Cross-wave dependencies

None beyond the W00→W02 entry dependency and the downstream table above.

## External dependencies

None new. This epic's tooling operates on the existing PostgreSQL/pgx toolchain already used
elsewhere in the framework's persistence layer. No new external service, queue, or third-party
migration tool is introduced.

## Repository dependencies

None cross-repo for this epic's own closure. wowsociety's eventual adoption of this protocol
(PLAN's own note: "adopt whatever manifest schema wowapi's tooling consumes... wowsociety's own
`cmd/migrate` runs the same underlying mechanics for its module migrations, making it a direct
consumer, not a bystander") is tracked as a future-adoption item, not a dependency this epic's
closure requires.

## Tooling dependencies

None beyond the already-available Go/PostgreSQL toolchain. S003-T9's CI drill pipeline extends the
existing CI infrastructure rather than introducing a new one.

## Decision dependencies

None. See `epic.md` "Required decisions."
