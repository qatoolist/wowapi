---
id: GOV-DOR
type: governance
title: Definition of Ready — entry gate for story and task work
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
derived: false
---

# Definition of Ready

Gate an item must satisfy before it may move to `ready` (see `lifecycle.md` for the transition
rule, `status-model.md` for the status vocabulary). Derived from mandate §4.3 (story
requirements), §8.4 (`story.md` required content), and §8.5 (`plan.md` required content).

## Story DoR

A story may move `planned` → `ready` only when every item below is true. Per mandate §4.3, each
story must be:

- [ ] **Specific** — the story targets one coherent capability, not an aspirational theme (mandate
      §4.3 explicitly rejects vague stories such as "improve architecture" or "enhance quality").
- [ ] **Bounded** — scope and out-of-scope are both stated in `story.md`; the story does not
      require broad, undefined, or unbounded implementation (mandate §2.1).
- [ ] **Implementable** — a concrete approach exists or can be determined without re-planning the
      programme; unresolved unknowns are recorded as assumptions or "must be determined during the
      story" per mandate §18, not silently assumed.
- [ ] **Independently reviewable** — the story's changes can be reviewed on their own, without
      requiring simultaneous review of unrelated work.
- [ ] **Independently verifiable** — the story's acceptance criteria can be proven with evidence
      that does not depend on another story's completion first (or that dependency is explicit).
- [ ] **Traceable to source requirements** — `story.md` front matter `source_requirements` lists
      every `impl/analysis/requirement-inventory.md` ID this story addresses.
- [ ] **Measurable acceptance criteria** — `acceptance_criteria` front matter lists numbered IDs
      (`AC-<story-id>-NN`); each is measurable (produces a pass/fail or quantitative result), not
      aspirational prose.
- [ ] **Scope and out-of-scope stated** — both sections present in `story.md`, not just scope.
- [ ] **Dependencies identified** — `depends_on` front matter and the dependency-register entry
      (if cross-epic/cross-wave) are populated; no silent dependency.
- [ ] **Assumptions recorded** — where facts are uncertain (e.g. repository state not yet
      confirmed), `story.md` records the assumption explicitly rather than asserting it as fact
      (mandate §18: "Record assumptions explicitly").
- [ ] **Plan drafted** — `plan.md` exists with proposed architecture/approach, task breakdown, and
      unresolved questions/approval conditions listed (mandate §8.5); it need not be final, but it
      must exist and distinguish confirmed facts from planned changes from assumptions.
- [ ] **Required artifacts and evidence anticipated** — `story.md` states which artifact types and
      evidence types this story is expected to produce (even if the concrete items don't exist
      yet — see `artifact-policy.md` / `evidence-policy.md` for why subdirectories are deferred).
- [ ] **Compatibility, security, performance, observability, migration considerations addressed**
      — each mandate §8.4 consideration section in `story.md` is either filled in or explicitly
      marked not-applicable with a one-line reason; it is not left blank.

A story failing any box stays in `planned` (or reverts to it) until closed.

## Task DoR

A task may move `todo` → `ready` only when:

- [ ] **Parent story is `ready` or `in-progress`** — a task cannot be readied against a story that
      is still `draft`/`planned` (unreadied) or already `accepted`/`cancelled`/`deferred`.
- [ ] **Dependencies resolved or explicitly waived** — every item in the task's `depends_on` list
      (mandate §8.6 "dependencies") is `done`, or a waiver is recorded with rationale.
- [ ] **Owner assignable** — a specific owner can be named (need not yet be assigned in front
      matter, but the work is not blocked on organizational ambiguity).
- [ ] **Mapped acceptance criteria identified** — the task's `related acceptance criteria`
      (mandate §8.6) reference specific `AC-...` IDs from the parent story; a task with no mapped
      AC is a signal it may be untracked busywork (mandate §12: tasks must not substitute for
      acceptance criteria — every task should trace to at least one AC it helps satisfy, or be
      justified as a pure enabling/closure activity per mandate §4.4's task-type list).
- [ ] **Expected files/components and completion criteria stated** — task.md §objective and
      §detailed work are specific enough that "done" is checkable without re-interpretation.

A task failing any box stays in `todo` until closed. Per mandate §12, avoid decomposing further
merely to satisfy this checklist mechanically — "avoid excessive fragmentation into trivial tasks
that provide no tracking value."

## Not part of DoR

DoR does not require implementation to have started, tests to exist yet, or evidence to be
collected — those are Definition of Done concerns (`definition-of-done.md`). DoR is an entry
gate; DoD is an exit gate.
