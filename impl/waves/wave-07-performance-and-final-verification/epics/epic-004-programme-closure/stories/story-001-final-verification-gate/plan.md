---
id: PLAN-W07-E04-S001
type: plan
parent_story: W07-E04-S001
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
---

# Plan — W07-E04-S001

Per mandate §8.5. This plan's own methodology mirrors REVIEW's own §30 gate exactly, applied fresh
against the programme's final state rather than restated from the original review. Confirmed facts,
planned changes, and assumptions are distinguished explicitly below.

## Proposed architecture

A fresh, documented re-run of REVIEW's own two-layer capability assessment (§H's 30-area capability
matrix, §I's 20-capability mandatory-readiness check), plus a traceability-completeness pass over
`requirement-inventory.md`'s own five tables (§A-E), plus a sampled disposition audit — each producing
its own report, together forming this story's own consolidated gate-re-run output.

## Implementation strategy

1. Re-assess each of REVIEW §H's original 30 capability areas against the programme's actual final
   implementation, not against REVIEW's own original classification.
2. Re-assess each of REVIEW §I's original 20 mandatory capabilities the same way.
3. Walk `requirement-inventory.md`'s own §A-E tables row by row, confirming each has a disposition and
   cross-checking that disposition against the item's own actual closure state (where applicable).
4. Select a disposition-audit sample, weighted toward P0/critical-priority stories, and for each sampled
   story, independently re-check its own `accepted` claim against its own `evidence/index.md` and
   `closure.md`.
5. Compile all of the above into the gate re-run's own consolidated report.

## Expected package or module changes

None — zero code change.

## Expected file changes where determinable

New documentation files (exact locations TBD): the gate re-run report; the traceability-completeness
check output; the disposition-audit report.

## Contracts and interfaces

None affected.

## Data structures

None.

## APIs

None affected.

## Configuration changes

None.

## Persistence changes

None.

## Migration strategy

Not applicable.

## Concurrency implications

None.

## Error-handling strategy

Not applicable.

## Security controls

None new — this story verifies existing controls' own closure state, it does not add new controls.

## Observability changes

None.

## Testing strategy

Not applicable in the code-test sense — this story's own "tests" are the capability re-assessments, the
traceability walk, and the disposition-audit spot-checks themselves, each producing documented evidence
of what was actually re-checked.

## Regression strategy

Not applicable — this is a one-time, terminal-wave verification exercise.

## Compatibility strategy

Not applicable.

## Rollout strategy

Single story, landed as its own reviewable documentation unit, sequenced after every other epic in this
wave reaches closure.

## Rollback strategy

Not applicable — a verification report is a factual record; if a re-assessment is later found incorrect,
it is corrected directly, and any downstream consumer (W07-E04-S002) of the incorrect version must be
notified.

## Implementation sequence

As listed under "Implementation strategy" above (steps 1–5) — steps 1-4 may proceed in parallel since
they target disjoint verification surfaces; step 5 (compilation) follows all four.

## Task breakdown

- **W07-E04-S001-T001** — Re-run the §H/§I-style capability assessments.
- **W07-E04-S001-T002** — Traceability-completeness check.
- **W07-E04-S001-T003** — Disposition audit (sampled).
- **W07-E04-S001-T004** — Independent review.

## Expected artifacts

The fresh gate re-run report; the traceability-completeness check output; the disposition-audit report.

## Expected evidence

The gate re-run's own evidence trail; the traceability-completeness check's own row-by-row output; the
disposition audit's own sampled-claim evidence trail.

## Unresolved questions

- Exact sample size/selection methodology for the disposition audit.
- The exact format for documenting the §H/§I-style re-assessment (a full replica of REVIEW's own
  30-row/20-row tables, or a narrower delta-only report against the original) — this story's own
  implementation determines the format, favoring whichever gives the clearest evidence trail.

## Approval conditions

This plan is approved for implementation once: (a) W07-E01, W07-E02, W07-E03 have all reached `accepted`
(this story's own entry gate), and (b) the owner and reviewer are assigned.
