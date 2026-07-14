---
id: PLAN-W03-E01-S004
type: plan
parent_story: W03-E01-S004
status: ready
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Plan — W03-E01-S004

Per mandate §8.5. This story has no PLAN §5 T-row of its own — it is a coordination-artifact story
this programme constructs directly from PROD-04's `requirement-inventory.md` registration and PLAN
SEC-01's own wowsociety-impact prose (§5.2), per the task brief's explicit instruction. Confirmed
facts, planned changes, and implementation assumptions are distinguished explicitly below.

## Proposed architecture

Not applicable in the code sense — this story produces three planning documents, not a code
architecture. Its "architecture" is the document structure itself: sequencing plan → staging-
validation plan → rollback plan, each independently reviewable per mandate §4.3's story
independence requirement, but sharing the same subject matter (the SEC-01 T1/T5 cutover).

## Implementation strategy

1. Draft the sequencing plan: repo-by-repo order, named files/tests requiring wowsociety-side
   rework, explicit statement of what "coordinated cutover" means operationally (e.g. a feature
   flag, a version-gate, or a hard cutover date — to be determined, see "Unresolved questions").
2. Draft the staging-validation plan: what wowsociety staging data and test suites validate T2/T5
   before production enforcement; explicit go/no-go criteria for declaring staging validation
   successful.
3. Draft the rollback plan: for each of the two failure directions named in `story.md`'s acceptance
   criteria, the specific revert steps and how to confirm neither repo is left inconsistent.
4. Circulate all three documents for review by at least a wowapi-side reviewer (and a
   wowsociety-side reviewer where practicable, though this story does not have authority to compel
   wowsociety-repo review).
5. Record review outcomes as evidence.

## Expected package or module changes

None. This story does not touch any Go package.

## Expected file changes where determinable

None in wowapi's source tree. The three plan documents themselves are new files under this story's
own `impl/` tree (exact filenames determined at implementation time, e.g.
`sequencing-plan.md`/`staging-validation-plan.md`/`rollback-plan.md`, placed either inline in this
story's directory or referenced from it — to be finalized at implementation time).

## Contracts and interfaces

None. This story documents an operational contract (the cutover sequence) but does not define a
code-level interface.

## Data structures

None.

## APIs

None.

## Configuration changes

None directly; the sequencing plan may recommend a feature-flag or version-gate mechanism for the
cutover, which — if adopted — would be a configuration change made by whoever executes the cutover,
not by this story.

## Persistence changes

None.

## Migration strategy

The sequencing plan references DATA-09's online-migration discipline as the pattern wowsociety's
own `identity_impersonation_session` schema change should follow (per REVIEW §P's recommendation)
but this story does not author any migration.

## Concurrency implications

None directly; the staging-validation plan should consider what happens if wowsociety production
traffic continues during the staging-validation window (a concurrency/consistency concern for the
plan's content, not for this story's own execution).

## Error-handling strategy

The rollback plan is, in effect, this story's error-handling strategy for the cutover as a whole —
it specifies what happens when the cutover itself errors.

## Security controls

The staging-validation plan's access-control recommendation (see `story.md` "Security
considerations") — validating against production impersonation data requires appropriate access
controls, not ad hoc access.

## Observability changes

The staging-validation plan should recommend wowsociety-side observability for the cutover window
(see `story.md` "Observability considerations") — a recommendation within the document, not an
implementation.

## Testing strategy

This story's own "testing" is a documentation-review process, not an executable test suite —
consistent with mandate §10's "review reports" evidence type. Each of the three plan documents is
reviewed against a checklist: does it name concrete files/tests (not vague references), does it
state explicit go/no-go criteria, does it cover both rollback failure directions.

## Regression strategy

Not applicable in the code-regression sense. The plan documents themselves are the regression guard
for the *cutover process* — a future cutover attempt without consulting them risks repeating a
coordination gap this story exists to close.

## Compatibility strategy

This entire story is a compatibility strategy document for SEC-01's breaking change — see `story.md`
throughout.

## Rollout strategy

The sequencing plan document is itself the rollout strategy.

## Rollback strategy

The rollback plan document is itself the rollback strategy — this story does not have a separate
rollback strategy for its own (documentation-only) work, since there is no code to roll back.

## Implementation sequence

As listed under "Implementation strategy" above (steps 1-5).

## Task breakdown

- **W03-E01-S004-T001** — Sequencing plan document.
- **W03-E01-S004-T002** — Staging-validation plan document.
- **W03-E01-S004-T003** — Rollback plan document.

Per the task brief: each of the three plan components is its own task, clearly noted as a
coordination-artifact task, not a code-implementation task. No independent-review task is added as
a fourth task here — because this story's tasks *are* review-facing documents whose own completion
criteria already require review sign-off (see each task's own "Completion criteria"), a separate
review task would duplicate rather than add tracking value, unlike S001/S002/S003 where review
checks code the tasks themselves do not review. This is recorded as a deliberate deviation from the
"always add an independent-review task to P0 stories" default, with the rationale stated here per
mandate §12's guidance against excessive fragmentation with no added tracking value — if the actual
reviewer disagrees at implementation time, a fourth review task can be added without renumbering.

## Expected artifacts

Sequencing plan document; staging-validation plan document; rollback plan document.

## Expected evidence

Review records for each of the three documents.

## Unresolved questions

- The exact wowsociety-side engineering ownership and timeline for adopting `grant_id` — not known
  at this story's planning time; the sequencing plan should state what must be coordinated rather
  than inventing a specific date or owner.
- Whether the "coordinated cutover" mechanism is a feature flag, a version-gate, or a hard cutover
  date — to be determined during this story's own execution, informed by whatever mechanism
  wowsociety's own deployment practice supports (`wowsociety/docs/DEPLOY.md`'s single-shot deploy
  model, per DATA-09's own wowsociety-impact note, may constrain the available options — to be
  confirmed by reading that document at implementation time, which is outside this story's own
  read scope during initial planning).
- Exact availability and access-control terms for wowsociety's staging environment/data — not
  assumed here; the staging-validation plan should state the validation approach and flag the
  access-control need explicitly rather than assuming unrestricted access.
- Whether DEC-Q1's eventual resolution (if it differs materially from the safe default) would
  require revising the sequencing plan — flagged as a possibility, not resolved here.

## Approval conditions

This plan is approved for implementation once: (a) S001 and S002 are substantially planned (their
`story.md`/`plan.md` exist, even if not yet implemented) so the three plan documents have a stable
target to reference, and (b) the owner and reviewer are assigned.
