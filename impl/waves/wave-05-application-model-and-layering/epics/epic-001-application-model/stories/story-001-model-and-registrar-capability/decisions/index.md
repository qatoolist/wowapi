---
id: W05-E01-S001-DECISIONS-INDEX
type: decisions-index
parent_story: W05-E01-S001
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W05-E01-S001 — Decisions index

This story references two existing, already-ratified decisions. No new ADR is authored here.

| Decision ID | Title | Status | Origin | Relationship to this story |
|---|---|---|---|---|
| ADR-W00-E02-S003-002 | One generic owner-bound Registrar type with per-subsystem typed keys (D-02) | accepted | W00-E02-S003 | Referenced — this story's `Registrar` capability type is the direct enactment of D-02's resolution of PLAN's own PF-ARCH cross-cutting note (5): "do all AR-01 per-subsystem registrars share one `Registrar` type... or does each get a distinct type." Not authored here; see `../../../../../../wave-00-baseline-and-verification/epics/epic-002-baseline-capture/stories/story-003-adr-ification/` for the ADR itself. |
| ADR-W00-E02-S003-003 | Post-seal mutation error not panic (D-03) | accepted | W00-E02-S003 | Referenced — this story's post-seal mutation behavior (error in production, panic only under an explicit dev/test build tag) is the direct enactment of D-03's resolution of PLAN's own PF-ARCH cross-cutting note (6): "should post-seal mutation panic in production builds, or only error?" Not authored here; see the same ADR-ification story directory. |

Per `impl/analysis/wave-allocation-detail.md`'s own W05-E01 story brief: "S001
model-and-registrar-capability (T1, T2; D-02/D-03 enacted as story ADR refs)" — both decisions are
referenced, not re-decided, exactly per the pattern established in
`W03-E01-S001/decisions/index.md` for D-01.
