---
id: W04-E03-S001-TASKS-INDEX
type: tasks-index
parent_story: W04-E03-S001
status: accepted
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W04-E03-S001 — Tasks index

Per mandate §16.4. Task files are single-file per the repository's documented adaptation (see
`governance/naming-conventions.md` "Adaptation 1") — the one task file below contains its task
definition, implementation record, verification record, and deviations record as internal sections.

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W04-E03-S001-T001](task-001-stopgap-fix-and-concurrency-test.md) | Correct false migration comment; enforce single-processor via advisory lock/CAS; concurrency test | W04BulkSafety | done | none | Corrected migration comment + enforcement mechanism + 2-processor concurrency test | AC-W04-E03-S001-01, AC-W04-E03-S001-02 | done | done |

## Grouping rationale

Per mandate §12: DATA-04 T1 is a single, tightly-scoped, single-owner unit of work — correcting one
false documentation claim and adding one enforcement check at one API boundary, proven by one named
concurrency test. The source's own framing underscores this: "Ships independently and fast — closes
the false-documentation P0 sub-issue before the full rewrite." Splitting the documentation
correction from the enforcement mechanism would produce two trivial tasks with no independent
tracking value — the comment correction is not meaningfully reviewable or verifiable on its own
without the enforcement mechanism it describes, and vice versa; mandate §12's own guidance is to
avoid "excessive fragmentation into trivial tasks that provide no tracking value."

**No separate independent-review task is added for this story**, unlike every P0/P1-boundary story
elsewhere in this program (e.g. `W02-E01-S001-T003`, `W02-E01-S002-T004`). This is a deliberate
judgment call, recorded here per the wave-planning brief's instruction to document the reasoning
either way:

- This story's scope is genuinely small — one documentation correction plus one enforcement
  mechanism guarded by one named concurrency test, not a multi-component build. The mandate's own
  guidance (§14) is to "define an independent review step" "for critical stories" — this story is
  P0 by priority classification (per the source's "P1; P0 before advertising multi-worker" framing
  and its fast-track role), but its P0 status derives from *urgency* (closing a false-documentation
  hazard quickly), not from *architectural complexity* or *breadth of blast radius*. Unlike
  `W02-E01-S001` (which locks a schema every future migration must satisfy) or `W04-E03-S002` (which
  rewrites the entire claim/fencing/lifecycle path), a flawed review-worthy decision in this story
  has a narrow, easily-correctable blast radius: the stopgap is explicitly, by this epic's own
  design (RISK-W04-E03-001), superseded by `W04-E03-S002`'s T2 lease-column mechanism shortly
  afterward. A defect here does not compound into downstream architecture the way an under-reviewed
  manifest schema or an under-reviewed fencing design would.
- T001's own "Completion criteria" and "Verification method" (below) already require confirming both
  acceptance criteria against source-derived, testable bars (the exact false claim removed; the
  exact concurrency scenario rejected) — this gives the single task itself a review-equivalent
  completion gate without a separate task record.
- If, during implementation, the mechanism choice (advisory lock vs. CAS) or the migration-comment
  correction proves more involved than this plan anticipates (e.g. the CAS path requires a
  non-trivial new migration), that discovery is recorded as a deviation in `deviations.md`, and the
  omission of a review task should be revisited at that point rather than assumed to remain correct
  regardless of how implementation unfolds.
