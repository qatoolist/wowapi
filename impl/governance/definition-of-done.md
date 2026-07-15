---
id: GOV-DOD
type: governance
title: Definition of Done — evidence-driven completion gate
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
derived: false
---

# Definition of Done

Mandate §2.5, "Evidence-driven completion":

> No story should be considered complete merely because code was written.

## The completion requirement list (mandate §2.5, verbatim)

Completion must require:

- implementation;
- required artifacts;
- tests;
- evidence;
- acceptance-criteria verification;
- review;
- documentation;
- resolution or acknowledgement of deviations and residual risks.

All eight items are mandatory for any item claiming completion at story level or above. None may
be skipped by asserting "not applicable" without a one-line rationale recorded in the relevant
section (`story.md` risk/residual-risk sections, or `deviations.md`).

## Task-level done

A task is `done` (see `lifecycle.md`) when:

1. **Implemented** — the task's `implementation.md` section (Adaptation 1, see
   `naming-conventions.md`) records what was actually done, matching or documenting deviation
   from the task's own definition.
2. **Verified** — the task's stated verification method has been executed and its
   `verification.md` section records actual result, evidence identifier, execution date/revision,
   and reviewer where applicable.

Task `done` is a narrower claim than story `accepted`: it proves this task's own completion
criteria were met, nothing more.

## Story-level done — `accepted`

Mandate §7, quoted verbatim, is the binding constraint:

> A story must not be accepted solely because all tasks are marked complete.

A story reaches `accepted` only when, in addition to every contained task being `done`:

- Every acceptance criterion in `story.md` has a corresponding entry in `verification.md` with
  `pass` result and a valid evidence ID (`evidence-policy.md`).
- Required artifacts (per `story.md` "required artifacts") are registered in `artifacts/index.md`
  with the fields specified in `artifact-policy.md`.
- Required evidence (per `story.md` "required evidence") is registered in `evidence/index.md`
  with the fields specified in `evidence-policy.md`.
- `deviations.md` either states "no deviations" or lists every deviation with reason, impact,
  approval, and compensating controls (mandate §8.9) — deviations are never silently absorbed
  into a rewritten plan (mandate §2.6).
- Documentation requirements stated in `story.md` are satisfied (doc files updated, or explicit
  not-applicable rationale).
- `closure.md` is complete: acceptance-criteria completion, task completion, artifact
  completeness, evidence completeness, unresolved findings, accepted risks, deferred work,
  reviewer conclusion, acceptance authority, closure date, final status (mandate §8.10).
- The independent-review checklist below has been run and passed clean.

## Epic- and wave-level done

An epic or wave reaches `accepted` (or `partially-accepted`, per `status-model.md`) only when:

- `closure-report.md` is complete for that epic/wave.
- All mandatory stories/epics in scope are `accepted`, OR are explicitly `deferred` with a
  recorded approval and target milestone (`tracking/deferred-items-register.md`), OR
  `partially-accepted` status is used with the gap and rationale stated in `closure-report.md`.
- Epic/wave-level acceptance criteria (mandate §8.2/§8.3) — distinct from any single story's AC —
  are themselves verified (e.g. cross-story integration behavior).
- Programme-level rule from `impl/index.md` "Programme acceptance" applies at the top: "no later
  wave starts while a mandatory predecessor capability is unaccepted" absent a documented
  exception/deviation record.

## Independent-review checklist (mandate §14) — final gate before `accepted`

For critical stories (and any story about to move `verified` → `accepted`), an independent
reviewer — someone who did not write the implementation — must verify all of the following before
acceptance. This is the same checklist invoked by the global `independent-review-gate` skill; it
applies at the level of an individual story's acceptance, not only at whole-goal completion.

- [ ] Implementation matches the approved plan (`plan.md`), or deviations are documented
      (`deviations.md`).
- [ ] Acceptance criteria are complete — every numbered AC in `story.md` has a verification
      outcome, none silently dropped.
- [ ] Tests meaningfully prove behaviour — not merely present, and not padding for coverage
      (mandate §13: "Do not create tests merely to increase numerical coverage").
- [ ] Evidence references the correct code revision (see `evidence-policy.md` revision-pinning
      rule).
- [ ] Artifacts are registered (`artifacts/index.md`, per `artifact-policy.md`).
- [ ] Regression risk is addressed (existing behavior not silently broken, or breakage is a
      documented, approved deviation).
- [ ] Architecture boundaries are preserved (framework/product boundary per mandate §2.3; no
      society/committee/policy vocabulary leaking into `wowapi` kernel packages).
- [ ] Security implications are handled (threat surface considered, not merely "tests pass").
- [ ] No unsupported completion claims are made (no "verified" without an evidence ID; no
      "tested" without an execution record).
- [ ] No source requirement has been silently dropped (cross-check against
      `impl/analysis/requirement-inventory.md` — every `source_requirements` entry this story
      claims to address is actually addressed or the gap is a recorded deviation).

Review findings from this checklist must be recorded (in `closure.md` or a linked review note)
and resolved or explicitly accepted as a residual risk before the story moves to `accepted`. A
checklist that "passes" with unresolved findings quietly waved through is not a pass.
