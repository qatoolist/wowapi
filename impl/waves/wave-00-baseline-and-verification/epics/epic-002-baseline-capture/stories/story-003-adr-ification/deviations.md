---
id: DEV-INDEX-W00-E02-S003
type: deviations-record
parent_story: W00-E02-S003
status: recorded
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Deviations — W00-E02-S003

One deviation recorded (below). This file is populated only if this story's actual implementation
diverges from the approved `plan.md` (mandate §8.9, §2.6). Per mandate §2.6, verbatim: "The
approved implementation plan must not be rewritten after implementation to make it appear that the
final implementation always matched the plan. Differences must be recorded in a separate deviation
record."

When a deviation occurs, it is recorded here using the shape in
`impl/governance/templates/deviation-template.md`: deviation ID (`DEV-W00-E02-S003-NNN`), approved
plan (quoted from `plan.md`), actual implementation, reason, impact, risks, approval, compensating
controls, follow-up work.

## DEV-W00-E02-S003-001 — ADR status vocabulary `ratified`, not `accepted`

- **Deviation ID:** DEV-W00-E02-S003-001.
- **Approved plan (quoted):** `story.md` AC-W00-E02-S003-01 requires each ADR front matter to carry
  "`status: accepted`" (AC-02 likewise says "status (`accepted`)"); task Detailed-work step 2 in all
  three task files says "`status: accepted`".
- **Actual implementation:** all nine ADRs (front matter and `## Status` body section) and
  `decisions/index.md` use **`status: ratified`**.
- **Reason:** `accepted` is not in the decision-status vocabulary. `decision-template.md` states
  verbatim: "Status vocabulary aligns with `decision-register.md`: proposed / ratified-pending-adr /
  ratified / superseded / rejected." `impl/tracking/decision-register.md` holds all nine D-0N rows
  at `ratified-pending-ADR` pending exactly this story — the post-ADR state is `ratified`. The
  Wave-00 execution brief for this story also mandates "status (ratified)". Using `accepted` would
  additionally collide with the story/epic lifecycle term `accepted` — the exact conflation
  `story.md` "Residual-risk expectations" warns about; `ratified` eliminates it.
- **Impact:** none on decision content; downstream consumers cite the ADR ID, not the status string.
  The literal wording of AC-01/AC-02 in `story.md` is satisfied in substance (status field populated,
  consistent across all nine ADRs and the index) with `ratified` as the vocabulary-correct value.
  `story.md`'s AC text itself is left unedited per mandate §2.6 (no post-hoc rewriting of the
  approved record); this deviation record is the required mechanism instead.
- **Risks:** a reviewer reading AC-01 hyper-literally could flag the mismatch — mitigated by this
  record and by `verification.md` stating the substitution explicitly in its post-execution record.
- **Approval:** pending conductor review at story acceptance (this story cannot self-approve; the
  substitution follows binding governance sources, so it is recorded, not silently absorbed).
- **Compensating controls:** scripted check EV-W00-E02-S003-010 enforces the `ratified` vocabulary
  across all nine ADRs and the index; the residual-risk note in every ADR's `## Status` section is
  retained.
- **Follow-up work:** conductor updates `impl/tracking/decision-register.md` rows D-01..D-09 from
  `ratified-pending-ADR` to `ratified` with the ADR paths (register is conductor-owned, out of this
  story's scope).
