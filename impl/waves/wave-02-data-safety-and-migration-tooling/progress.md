---
id: W02-PROGRESS
type: wave-progress
wave: W02
status: accepted
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W02 progress

All W02 epics and stories are accepted.

## Epic status

| Epic | Title | Status | Stories | Story status breakdown |
|---|---|---|---|---|
| W02-E01 | online-migration-protocol | accepted | 3 | 3 accepted |
| W02-E02 | tenant-fk-integrity | accepted | 2 | 2 accepted |
| W02-E03 | version-allocation-and-gc | accepted | 1 | 1 accepted |
| W02-E04 | aggregate-write-contract | accepted | 1 | 1 accepted |
| W02-E05 | production-seed-sync | accepted | 1 | 1 accepted |

## Story status

| Story | Title | Status | Task count | Task status breakdown |
|---|---|---|---|---|
| W02-E01-S001 | manifest-and-lock-budget | accepted | 3 | done |
| W02-E01-S002 | expand-backfill-validate | accepted | 4 | done |
| W02-E01-S003 | canary-switch-contract-drills | accepted | 6 | done |
| W02-E02-S001 | parent-indexes-scanner-gate | accepted | 4 | done |
| W02-E02-S002 | audit-fk-validate-negatives | accepted | 6 | done |
| W02-E03-S001 | version-races-and-blob-gc | accepted | 5 | done |
| W02-E04-S001 | aggregate-write-contract | accepted | 4 | done |
| W02-E05-S001 | prod-seed-sync-path | accepted | 6 | done |

## Blocked items

None yet — no story has entered `in-progress`. Note for future readers: W02-E02-S002's tasks T4/T5
(PLAN DATA-01 T4/T5, the `NOT VALID` composite-FK add and `VALIDATE CONSTRAINT`) are recorded as
gated on W02-E01-S001+S002 acceptance in `story.md`'s own `depends_on` and in
`epics/epic-002-tenant-fk-integrity/dependencies.md` — this is a planned internal-epic dependency,
not a blocked item, until E01 actually reaches `in-progress` without E02 waiting correctly.

## Critical dependencies

- W02-E02 (DATA-01) depends on W02-E01 (DATA-09): specifically, E02-S002's T4/T5 (the risky
  `NOT VALID` add and `VALIDATE CONSTRAINT` steps) must not start before E01-S001+S002 acceptance,
  per `impl/analysis/wave-allocation-detail.md`'s explicit cross-wave sequencing note.
- W02-E01-S002's backfill harness (T4) has a forward dependency on DATA-02 T1's lease primitive,
  which does not yet exist (DATA-02 is W04 scope). S002 builds a minimal checkpoint lease as an
  interim measure; W04-E01-S001 replaces it. This is recorded as a planned deviation-risk, not a
  silent shortcut — see `risks.md`.
- W02-E05 (FBL-02) contains a design-investigation task (catalog manifest format) that must
  complete, with its design decision recorded, before any of its implementation tasks can begin.

## Open decisions

None new to W02. No W02 story enacts a new ADR (confirmed — see `wave.md` "Assumptions" and each
epic's `epic.md` "Required decisions," all stating none). FBL-02's catalog-manifest-format question
is a design-investigation output, not an architecture decision in the D-0N sense — it does not
require ADR-ification under this programme's existing D-01..D-09 register.

- W02-E05-S001 completed; no open risks.

## Open risks

See `risks.md`. RISK-W02-004 resolved within scope.

## Artifact completeness

1/8 story-level artifact sets populated (W02-E05-S001).

## Evidence completeness

6 evidence records registered for W02-E05-S001.

## Review state

Not yet reviewed.

## Exit-gate readiness

Partial. 1 of 8 stories accepted (W02-E05-S001); remaining 7 stories still planned/in-progress.
