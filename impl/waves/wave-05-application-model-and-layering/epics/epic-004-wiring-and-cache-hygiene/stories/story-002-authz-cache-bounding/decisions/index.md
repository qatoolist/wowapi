---
id: W05-E04-S002-DECISIONS-INDEX
type: decisions-index
parent_story: W05-E04-S002
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W05-E04-S002 — Decisions index

This story references one existing, already-ratified decision. No new ADR is authored here.

| Decision ID | Title | Status | Origin | Relationship to this story |
|---|---|---|---|---|
| ADR-W00-E02-S003-006 | Per-tenant authz_epoch table, polled; LISTEN/NOTIFY optional only (D-06) | accepted | W00-E02-S003 | Referenced — this story's T4 (per-tenant authz_epoch table and cross-pod epoch-bump wiring) is the direct enactment of D-06's resolution of SEC-04 T4's own "open architecture decision (LISTEN/NOTIFY vs. epoch-row-poll)," per PLAN's own risk column and MATRIX CS-17's own concretized closure spec. Not authored here; see `../../../../../../wave-00-baseline-and-verification/epics/epic-002-baseline-capture/stories/story-003-adr-ification/` for the ADR itself. |

Per `impl/analysis/wave-allocation-detail.md`'s own W05-E04 story brief: "S002 authz-cache-bounding
(SEC-04 all tasks per CS-17: golang-lru + epoch table D-06...)" — D-06 is referenced, not re-decided,
exactly per the pattern established in `W03-E01-S001/decisions/index.md` for D-01 and
`W05-E01-S001/decisions/index.md` for D-02/D-03.
