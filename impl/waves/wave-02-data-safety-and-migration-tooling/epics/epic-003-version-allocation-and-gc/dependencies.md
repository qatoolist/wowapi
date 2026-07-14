---
id: W02-E03-DEPS
type: epic-dependencies
epic: W02-E03
wave: W02
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W02-E03 — Dependencies

## Upstream (epics/waves this epic depends on)

- **W00** (full wave) — per `../../dependencies.md` (wave-level), W02 depends on W00's exit gate.
  This epic has no additional upstream dependency beyond that gate — DATA-05 has no D-0N ADR
  dependency and no dependency on any W01 finding.

Unlike **W02-E02**, which is explicitly gated on **W02-E01**'s protocol for its riskiest steps, this
epic carries no internal-wave dependency at all. PLAN's own DATA-05 T1 row has an empty Depends-on
column ("—"), and `requirement-inventory.md`'s DATA-05 row Notes column is blank — it cites no
dependency on DATA-09, DATA-01, or any other finding. `../../dependencies.md` (wave-level) confirms
this explicitly under "Internal (within this wave, between epics)": "W02-E03, W02-E04, W02-E05 are
independent of W02-E01, W02-E02, and of each other... They may execute in any order or in parallel
with E01/E02 and with each other."

## Downstream (epics/waves that depend on this epic)

None identified in the source. No downstream epic or wave cites a dependency on W02-E03/DATA-05 in
`impl/analysis/wave-allocation-detail.md`'s cross-wave sequencing notes, `impl/index.md`'s wave map,
or any other epic's `dependencies.md` reviewed while building this programme's W02 epics.

## Internal (within this epic)

Single story (S001); no internal epic-level story dependency to record. Within S001, per PLAN
DATA-05's own Depends-on column: T2 depends on T1; T3 depends on T1 and T2; T4 depends on T2 and T3;
T5 depends on T1. T1 and T5 both depend only on the counter/sequence mechanism itself — T5 does not
depend on T2/T3/T4, since `kernel/artifact.Generate` has no upload-session or GC surface. See
`stories/story-001-version-races-and-blob-gc/story.md` "Dependencies" for the story-scoped statement.

## Cross-wave dependencies

None. This epic does not depend on any W01, W03, W04, W05, W06, or W07 item, and no downstream item
in any later wave has been found to depend on this epic.

## External dependencies

None new. This epic's counter/sequence mechanism and upload-session table are built on the existing
PostgreSQL/pgx toolchain already used elsewhere in the framework's persistence layer. The GC sweep
operates against the existing object-storage backend already used by `kernel/document`'s upload path
(no new storage provider is introduced).

## Repository dependencies

None cross-repo. `requirement-inventory.md`'s DATA-05 row and PLAN's own wowsociety-impact note both
confirm: "No `kernel/artifact`/`kernel/document` import found anywhere in wowsociety." This epic has
no product-level coordination item to record.

## Tooling dependencies

None beyond the existing Go/PostgreSQL toolchain. The scheduled GC sweep (T4) may extend an existing
scheduled-job mechanism if one already exists in the framework, or introduce a minimal one scoped to
this epic's own need — to be determined at implementation time (see
`stories/story-001-version-races-and-blob-gc/plan.md` "Unresolved questions").

## Decision dependencies

None. See `epic.md` "Required decisions."
