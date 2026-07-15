---
id: W03-E01-S001-DECISIONS-INDEX
type: decisions-index
parent_story: W03-E01-S001
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W03-E01-S001 — Decisions index

This story references one existing, already-ratified decision. No new ADR is authored here.

| Decision ID | Title | Status | Origin | Relationship to this story |
|---|---|---|---|---|
| ADR-W00-E02-S003-001 | Framework owns grant validity/expiry/revocation (D-01) | accepted | W00-E02-S003 | Referenced — this story's `identity_grant` table and unconditional membership check are the direct enactment of D-01's framework-boundary decision. Not authored here; see `../../../../../wave-00-baseline-and-verification/epics/epic-002-baseline-capture/stories/story-003-adr-ification/` for the ADR itself. |

Per `governance/naming-conventions.md` and this story's task brief: "W03-E01-S001→D-01+DEC-Q1-safe-
default" — D-01 is a referenced, already-ratified decision (registered here). DEC-Q1 is a distinct,
still-open human decision this story proceeds against via its documented safe default; DEC-Q1 is
recorded in `story.md`'s "Assumptions" section, not in this decisions index, since DEC-Q1 has no
ADR — it is not yet a resolved decision at all.
