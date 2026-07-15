---
id: TRACK-INDEX
type: index
title: Tracking directory index — purpose and canonical-vs-derived map
status: active
created_at: 2026-07-12
updated_at: 2026-07-12
derived: true
---

# Tracking index

This file is itself a DERIVED view (it describes the registers, it does not hold status data).
No register in this directory is a second source of truth for status — see
`governance/status-model.md` for the controlled status vocabulary and canonical-status rule
(mandate §6: "Do not create multiple manually maintained sources of truth for status.").

Per mandate §3, `tracking/` holds registers and matrices that are either generated from
canonical front matter or clearly marked as derived views. At programme-planning time (no
`waves/` directory yet), every register below is derived from `impl/analysis/requirement-inventory.md`
(the canonical allocation) and the four primary source documents. Once `waves/` exists, canonical
status moves to each wave/epic/story/task's own front matter and these registers become roll-ups
of that front matter instead.

## Registers in this directory

| Register | Purpose | Canonical source (today) | Canonical source (once `waves/` exists) |
|---|---|---|---|
| `status-register.md` | Roll-up of every wave/epic/story's current status | `requirement-inventory.md` Target column + disposition | Each item's own front matter (`wave.md`/`epic.md`/`story.md`) |
| `source-traceability-matrix.md` | Source document+section → requirement/finding → disposition → planned implementation item | `requirement-inventory.md` + the four primary docs | Same (source documents don't change retroactively) |
| `requirement-traceability-matrix.md` | Requirement → wave → epic → story → AC → task → artifact → evidence → final result | `requirement-inventory.md` Target column | Story front matter `acceptance_criteria` + each story's `tasks/`, `artifacts/index.md`, `evidence/index.md` |
| `dependency-register.md` | Structured dependency edges (story/epic/cross-wave/external/tooling/decision) | `impl/index.md` wave map + `requirement-inventory.md` notes columns + PLAN §7 | Same, cross-checked against `waves/*/dependencies.md` once populated |
| `risk-register.md` | Structured risk records (likelihood/impact/severity/mitigation/owner/status) | REVIEW §T + PLAN §7 + `impl/index.md` programme risks | Same, extended by `waves/*/risks.md` once populated |
| `decision-register.md` | Ratified architecture decisions (D-01..D-09) + open human decisions (DEC-Qx) + provisional planning assumptions (PA-0x) | REVIEW §U + the three DEC-Qx rows + planning-assumptions | Same, until each decision gets its own ADR under a story's `decisions/` directory |
| `technical-debt-register.md` | Accepted debt with origin, reason, impact, resolution target | Named seed items in this file (grounded against repo state at HEAD) | Extended as stories close and introduce/resolve debt (`implementation.md` "technical debt introduced" field) |
| `deviation-register.md` | Plan-vs-actual deviations | Currently empty (no story has entered implementation) | Each story's own `deviations.md` once populated |
| `deferred-items-register.md` | Items explicitly deferred outside this implementation programme | `requirement-inventory.md` deferred-disposition rows + `framework-backlog-p2-decisions.md` | Same (deferred items are stable once approved; reopening changes status here) |
| `change-log.md` | **NOT derived** — append-only source of record for programme-structure changes | — (this file IS the source) | — (unchanged) |

## Reading order

New readers should start at `status-register.md` for a snapshot, then `source-traceability-matrix.md`
or `requirement-traceability-matrix.md` depending on whether they are working forward (source →
implementation) or backward (implementation → source). `dependency-register.md` and
`risk-register.md` are consulted before starting any wave. `decision-register.md` is consulted
whenever a story's scope touches D-01..D-09 or one of the three open human decisions.

## Planning-time snapshot

At the time these registers were created (HEAD `0a31186`), `impl/waves/` does not yet exist.
Every register in this directory reflects a planning-time snapshot: statuses are `planned`
(nothing has entered implementation), and the "target" columns point at wave/epic/story IDs that
are named in `requirement-inventory.md` but do not yet have their own directories or front matter.
This is expected and correct per the mandate's phased structure — waves are populated after the
programme (this `impl/` tree) is accepted.
