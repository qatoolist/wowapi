---
id: W04-E04-S001-DECISIONS-INDEX
type: decisions-index
parent_story: W04-E04-S001
status: accepted
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W04-E04-S001 — Decisions index

This story references one existing, already-ratified decision. No new ADR is authored here.

| Decision ID | Title | Status | Origin | Relationship to this story |
|---|---|---|---|---|
| ADR-W00-E02-S003-004 | Audit hash_version smallint column, version-branched verification (D-04) | accepted | W00-E02-S003 | Referenced — this story's widened `chainHash` and `hash_version` migration are the direct enactment of D-04's version-discriminator design. Not authored here; see `../../../../../../wave-00-baseline-and-verification/epics/epic-002-baseline-capture/stories/story-003-adr-ification/decisions/adr-004-audit-hash-version-column.md` for the ADR itself. |

Per `governance/naming-conventions.md` and this epic's own `epic.md` "Required decisions" section:
D-04 is a referenced, already-ratified decision (registered here), enacted by this story only — no
other story in W04-E04 carries a `decisions/` directory. This story's own contribution is
implementation and evidence (the per-field tamper test, the version-branch verification test), not
re-deciding the discriminator design D-04 already settled.
