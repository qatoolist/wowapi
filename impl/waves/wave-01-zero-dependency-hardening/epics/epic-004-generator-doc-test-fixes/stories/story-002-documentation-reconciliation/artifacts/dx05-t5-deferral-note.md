---
id: ART-W01-E04-S002-004
type: artifact
title: DX-05 T5 — explicit deferral to W06/REL-03
parent_story: W01-E04-S002
producing_task: W01-E04-S002-T002
source_requirement: DX-05 (T5)
status: produced
created_at: 2026-07-13
---

# DX-05 T5 — deferral note

**DX-05 T5** (public API/config/event compatibility gates enforcing v1 rules —
`docs/implementation/premier-framework-implementation-plan.md` §5.4 DX-05 task table, T5 row) is
**deferred to W06**, not implemented in W01-E04-S002 and not silently dropped.

- **Cross-reference:** the plan document's own T5 row marks it "shared with REL-03/DX-06"; per
  `impl/analysis/requirement-inventory.md` row REL-03, that shared compat-gate infrastructure
  targets `W06-E02-S002..S003`. T5 lands there, on REL-03's shared plumbing (Go public API diff,
  config-schema compatibility, OpenAPI semantic diff gated on DX-06).
- **Recorded per** mandate §11.10's deferred-items discipline: this note feeds the programme-level
  `impl/tracking/deferred-items-register.md`; `../story.md` "Out of scope" carries the same
  deferral.
- **Nothing about T5 is blocked by this story;** the deferral is a sequencing decision the plan
  document itself already made ("High — large, shared with REL-03/DX-06").
