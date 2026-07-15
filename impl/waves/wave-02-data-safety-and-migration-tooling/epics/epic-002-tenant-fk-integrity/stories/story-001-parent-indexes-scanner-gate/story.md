---
id: W02-E02-S001
type: story
title: Parent tenant-scoped unique indexes, FK catalog scanner, and CI gate
status: accepted
wave: W02
epic: W02-E02
owner: W02FKVerAgg
reviewer: W02ReviewGate
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-13
source_requirements:
  - DATA-01
depends_on: []
blocks:
  - W02-E02-S002
acceptance_criteria:
  - AC-W02-E02-S001-01
  - AC-W02-E02-S001-02
  - AC-W02-E02-S001-03
artifacts:
  - ART-W02-E02-S001-001
  - ART-W02-E02-S001-002
  - ART-W02-E02-S001-003
  - ART-W02-E02-S001-004
evidence:
  - EV-W02-E02-S001-001
  - EV-W02-E02-S001-002
  - EV-W02-E02-S001-003
decisions: []
risks: []
---

# W02-E02-S001 — Parent tenant-scoped unique indexes, FK catalog scanner, and CI gate

## Story ID

W02-E02-S001

## Title

Parent tenant-scoped unique indexes, FK catalog scanner, and CI gate

## Objective

Add or confirm `UNIQUE (tenant_id, id)` on every parent table referenced by the 8 tenant-scoped
child-table foreign keys named in PLAN DATA-01's evidence, built `CONCURRENTLY`; build a tenant-FK
catalog scanner that enumerates every tenant-table FK not composite on `(tenant_id, …)`; and wire
that scanner into a permanent CI gate that fails a new migration adding a single-column tenant FK.

## Value to the framework

This story is the low-risk, purely-additive, CI-durable half of DATA-01's work — it does not itself
touch RLS or referential integrity between existing rows, it only makes the *foundation* the
composite FKs will point at exist (the parent's own `UNIQUE (tenant_id, id)`), and it builds the
mechanical enumeration (the scanner) that S002's later work depends on to know it has covered every
affected edge with zero silent gaps. PLAN's own risk note for T6 states this story's CI-gate half is
"cheapest, most durable part — do first if sequencing allows" — this story exists precisely so that
guidance is honored: it does not wait on this wave's W02-E01 online-migration protocol (unlike S002's
riskiest steps), so it can and should land first and fast, converting "no tenant-FK is currently
checked mechanically" into "any future migration that adds a non-composite tenant FK fails CI
immediately," well before the higher-risk composite-FK rollout itself begins.

## Problem statement

PLAN's own DATA-01 evidence, quoted exactly: "8 tenant-scoped child tables
(persons/legal_entities/party_contacts/acting_capacities → parties; resources → organizations;
document_versions/document_access_grants/attachments → documents/document_versions) reference only
the parent's `id`, never `(tenant_id, id)`. RLS proves the child row's own tenant; nothing proves
parent and child agree." PLAN's task table gives this story's three tasks directly: "T1. Add/confirm
`UNIQUE (tenant_id, id)` on every referenced parent, built `CONCURRENTLY` | — | Every parent has the
unique index | Migration test via `pg_indexes` | `DATA-01/parent-index/` | `SHARE UPDATE EXCLUSIVE`
lock — must run non-transactionally." "T2. Build a tenant-FK catalog scanner flagging any
tenant-table FK not composite on `(tenant_id, …)` | T1 | Enumerates exactly the 8 known FKs with zero
silent gaps; becomes a permanent CI gate | Fixture-schema test | `DATA-01/fk-catalog/` | Must key off
the existing RLS-tagged tenant-table matrix, not a hand-maintained list." "T6. Wire the T2 scanner
into a permanent CI gate | T2 | A new migration adding a single-column tenant FK fails CI | Negative
fixture migration | `DATA-01/gate-test/` | Cheapest, most durable part — do first if sequencing
allows." MATRIX CS-18's fail-first framing applies to this epic as a whole: "platform-role seeded
cross-tenant parent/child insert succeeds today; fails after" — this story does not itself close that
fail-first test (S002-T7 does), but it builds the parent-side prerequisite and the durable CI gate
that prevents the gap from silently reopening once closed.

## Source requirements

DATA-01 (T1, T2, T6).

## Current-state assessment

Per PLAN's own evidence (to be re-confirmed at this story's actual start commit, consistent with this
programme's fail-first re-confirmation convention applied elsewhere, e.g. W01-E01-S001 and
W02-E01-S001): none of the 8 parent tables (`parties`, `organizations`, `documents`,
`document_versions`) is confirmed to already carry a `UNIQUE (tenant_id, id)` index — PLAN's own
acceptance criterion for T1 is phrased "add/confirm," acknowledging some parents may already have it
and others may not; this story's own T1 task must check `pg_indexes` per-parent before assuming every
index needs to be created from scratch. No tenant-FK catalog scanner exists anywhere in the
repository today, and no CI gate currently rejects a migration adding a non-composite tenant-table
FK — this is a confirmed absence per PLAN's own framing of DATA-01 as a whole ("nothing proves parent
and child agree").

## Desired state

Every one of the 8 confirmed edges' parent table (`parties`, `organizations`, `documents`,
`document_versions`) carries a `UNIQUE (tenant_id, id)` index, built `CONCURRENTLY` so it does not
block concurrent traffic. A tenant-FK catalog scanner exists that enumerates every tenant-scoped
table's foreign keys and flags any FK that is not composite on `(tenant_id, …)`, keyed off the
existing RLS-tagged tenant-table matrix (not a hand-maintained list, per PLAN T2's own risk note) —
confirming it finds exactly the 8 known FKs with zero silent gaps. That scanner is wired into CI as a
permanent gate: any future migration that adds a single-column (non-composite) tenant-table FK fails
the build, proven by a negative fixture migration.

## Scope

- `UNIQUE (tenant_id, id)` migrations for `parties`, `organizations`, `documents`, and
  `document_versions`, each built `CONCURRENTLY` (non-transactionally, per T1's own risk note).
- The tenant-FK catalog scanner tool, keyed off the existing RLS-tagged tenant-table matrix.
- Wiring the scanner into a permanent CI gate (extending the existing CI infrastructure, consistent
  with how W02-E01-S001's manifest-schema validation and lock-timeout enforcement are wired in).
- A negative fixture migration proving the CI gate actually fails a non-composite tenant FK.

## Out of scope

- The mismatch audit, the composite `NOT VALID` FK add, `VALIDATE CONSTRAINT`, and the cross-tenant
  negative tests — W02-E02-S002's scope (PLAN DATA-01 T3, T4, T5, T7, T8). This story builds the
  parent-side prerequisite and the durable scanner/gate; it does not itself add or validate any
  composite FK on the child tables.
- Any dependency on this wave's W02-E01 online-migration protocol — unlike S002's T4/T5, none of this
  story's three tasks (T1, T2, T6) is gated on W02-E01's acceptance; T1's `CONCURRENTLY` index build
  and T2/T6's scanner-and-gate work are safe, additive changes that do not require the online-
  migration protocol's expand/backfill/validate machinery.
- wowsociety's own `policy_override.rule_version_id` migration (tracked as `PROD-01`, product-level)
  — depends on this story's T1 landing in wowapi first, but its own migration is out of scope here
  per mandate §2.3 (framework-first scope).

## Assumptions

- The 8 tenant-scoped child tables and their parents are assumed to still be exactly the set named in
  PLAN's evidence at this story's actual start commit — subject to this story's own T2 scanner
  re-confirming the inventory mechanically rather than trusting the hand-written list, per PLAN T2's
  own risk note and per `wave.md`'s "Assumptions" section, which flags this as illustrative of
  programme convention, not a load-bearing dependency of the wave.
- Some of the 4 parent tables may already carry a `UNIQUE (tenant_id, id)` index (T1 says
  "add/confirm," not merely "add") — this story's own implementation must check `pg_indexes`
  per-parent before assuming every index is missing; which of the 4 parents (if any) already have the
  index is not determinable from the source documents and must be confirmed at implementation time.
- The exact mechanism for the catalog scanner to "key off the existing RLS-tagged tenant-table
  matrix" (T2's own risk note) — i.e., which existing artifact constitutes that matrix — is not named
  precisely in the source; this story's plan records it as an implementation-time investigation step,
  not an invented specific.

## Dependencies

None within W02-E02 (this is the epic's first story). Depends on W00's exit gate at wave scope, per
`../../dependencies.md` (epic-level) and `../../../../dependencies.md` (wave-level). Blocks
W02-E02-S002: S002's T3 mismatch audit is more useful once this story's T2 catalog scanner confirms
the 8-edge inventory is complete and current (an audit against an incomplete FK inventory could
silently miss an edge), and S002's T4 has an intra-epic dependency on this story's T1 (PLAN's own
Depends-on column for T4: "T1, T3").

## Affected packages or components

New: the parent `UNIQUE (tenant_id, id)` migrations (4 new or modified migration files, one per
parent table, exact locations under the existing migration directory structure); the tenant-FK
catalog scanner tool (exact package location TBD, expected under a similar location to
W02-E01-S001's manifest-schema validator, e.g. `internal/tools/` or a migration-tooling-adjacent
package). Extended: the CI configuration (e.g. `.github/workflows/`) to invoke the scanner as a gate.

## Compatibility considerations

Adding a `UNIQUE (tenant_id, id)` index to a table that does not yet have one is purely additive and
non-breaking — it does not change any existing query's result set, only adds an index a later
composite FK can reference. The CI gate this story wires in (T6) does change future migration-author
behavior: once wired, any future migration adding a single-column tenant FK will fail CI going
forward. Per T6's own acceptance criterion this enforcement is immediate upon landing (no transition
period is described in the source for T6, unlike W02-E01-S001's own open question about its manifest
gate's enforcement timing) — this story's plan should confirm this reading is correct at
implementation time rather than assume a transition period exists where the source does not describe
one.

## Security considerations

None distinct from this epic's overall security posture — this story's own scope (parent indexes,
scanner, CI gate) does not itself change any access-control or RLS behavior; it only prepares the
foundation the security-relevant composite-FK enforcement (S002) will build on. The CI gate's own
correctness (T2/T6) is itself a security-adjacent control in the sense that a gap in its enumeration
would allow a future tenant-isolation-weakening migration to land silently — this is why T2's
acceptance criterion requires "zero silent gaps," not merely "detects the known 8."

## Performance considerations

`UNIQUE (tenant_id, id)` indexes built `CONCURRENTLY` avoid blocking concurrent traffic during
creation, per T1's own risk note ("`SHARE UPDATE EXCLUSIVE` lock — must run non-transactionally").
The scanner itself is a build/CI-time tool, not a runtime component — it has no production
performance impact.

## Observability considerations

The CI gate's failure output (when a future migration is rejected for a non-composite tenant FK)
should be a clear, actionable error message identifying which migration and which FK failed — a
reasonable implementation-time requirement given the gate's purpose is to guide a migration author to
the correct pattern, not merely block them.

## Migration considerations

This story's T1 is itself the migration work: 4 new `CONCURRENTLY`-built unique-index migrations,
each run non-transactionally per T1's own risk note. No data migration or backfill is required — a
unique index addition does not move or transform existing data.

## Documentation requirements

Document the tenant-FK catalog scanner's purpose, how it is keyed off the existing RLS-tagged
tenant-table matrix, and how the CI gate behaves (what triggers a failure, what the migration author
should do instead) — so a future migration author who trips the gate has a clear path to a compliant
migration without needing to read this story's own planning documents.

## Acceptance criteria

- **AC-W02-E02-S001-01**: `UNIQUE (tenant_id, id)` exists, confirmed via `pg_indexes`, on `parties`,
  `organizations`, `documents`, and `document_versions` — each built `CONCURRENTLY`, proven by a
  migration test querying `pg_indexes` post-migration.
- **AC-W02-E02-S001-02**: The tenant-FK catalog scanner enumerates exactly the 8 known FKs
  (persons/legal_entities/party_contacts/acting_capacities → parties; resources → organizations;
  document_versions/document_access_grants/attachments → documents/document_versions) with zero
  silent gaps, keyed off the existing RLS-tagged tenant-table matrix rather than a hand-maintained
  list, proven by a fixture-schema test.
- **AC-W02-E02-S001-03**: The scanner is wired as a permanent CI gate; a negative fixture migration
  that adds a single-column (non-composite) tenant-table FK fails CI, proven by that fixture's actual
  CI run.

## Required artifacts

- The 4 parent `UNIQUE (tenant_id, id)` migrations.
- The tenant-FK catalog scanner tool.
- The CI gate wiring (workflow/configuration change).
- The negative fixture migration used to prove the gate.
See `artifacts/index.md`.

## Required evidence

- `pg_indexes` migration test output confirming all 4 parent unique indexes exist.
- Fixture-schema catalog-scanner test output confirming exactly 8 FKs are enumerated with zero gaps.
- Negative fixture migration CI run output confirming the gate fails a non-composite tenant FK.
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`: `story.md`
and `plan.md` complete, acceptance criteria numbered and measurable, dependencies (none within this
epic) recorded, owner/reviewer assignment pending, the two open implementation-time questions (which
parents already have the index; what constitutes "the existing RLS-tagged tenant-table matrix" the
scanner keys off) explicitly recorded rather than silently assumed.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; all three acceptance
criteria verified with evidence in `evidence/index.md`; `closure.md` completed; independent review
passed per mandate §14, specifically confirming the scanner's enumeration has zero silent gaps
(AC-W02-E02-S001-02) and the CI gate genuinely rejects a non-composite tenant FK, not merely claims to
(AC-W02-E02-S001-03).

## Risks

None specific to this story beyond the epic-level risks (RISK-W02-002, RISK-W02-E02-002), both of
which land in S002, not this story — this story's own scope (additive indexes, a scanner, a CI gate)
carries no distinct risk requiring its own risk-register entry beyond the general
under-specification risk already captured for the sibling W02-E01-S001 pattern (not reproduced here
since this story's schema — the scanner's inputs and the CI gate's trigger condition — is more
narrowly bounded by PLAN's own explicit "exactly the 8 known FKs" acceptance bar than W02-E01-S001's
open-ended manifest-format design question was).

## Residual-risk expectations

Once T1's per-parent `pg_indexes` confirmation and T2's zero-silent-gap enumeration are both verified
by their own named tests, residual risk is expected to be low — this is the epic's own
characterization of this story as the "low-risk, purely-additive, CI-durable half" of DATA-01's work.

## Plan

See `plan.md`.
