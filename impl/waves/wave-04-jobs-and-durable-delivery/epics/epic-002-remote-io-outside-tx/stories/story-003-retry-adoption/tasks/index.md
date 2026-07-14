---
id: W04-E02-S003-TASKS-INDEX
type: tasks-index
parent_story: W04-E02-S003
status: accepted
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W04-E02-S003 — Tasks index

Per mandate §16.4. Task files are single-file per the repository's documented adaptation (see
`governance/naming-conventions.md` "Adaptation 1") — each task file below contains its task
definition, implementation record, verification record, and deviations record as internal sections.

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W04-E02-S003-T001](task-001-locate-and-replace-hand-rolled-retry.md) | Locate and replace both hand-rolled retry implementations | W04-Rerun | done | none | `cenkalti/backoff/v5` integration at both call sites | AC-W04-E02-S003-01 | done | done |
| [W04-E02-S003-T002](task-002-parity-and-fault-injection-tests.md) | Retry-schedule-parity and fault-injection tests | W04-Rerun | done | T001 | Parity test + fault-injection test suites | AC-W04-E02-S003-01, AC-W04-E02-S003-02 | done | done |
| [W04-E02-S003-T003](task-003-lightweight-review.md) | Lightweight review | unassigned | todo | T001, T002 | Lightweight review record | AC-W04-E02-S003-01, AC-W04-E02-S003-02 | pending | pending |

## Grouping rationale

Per this story's own prompt framing — "a SMALL, WELL-BOUNDED item. Do not artificially inflate
scope. Two tasks are plenty (adoption; parity/fault-injection tests)" — T001 (adoption: locating and
replacing both hand-rolled implementations) and T002 (the parity and fault-injection tests) are kept
as exactly two substantive tasks, matching REVIEW §O's own two-part test requirement ("retry-schedule
parity + fault injection") as a single combined test task rather than splitting into two, since both
tests exercise the same two replaced call sites and are naturally authored together once T001's
replacement lands.

T003 adds a lightweight review, not a full mandate-§14 independent-review task of the depth used in
S001/S002. Rationale for the lighter approach, per this story's own priority and risk profile:
FBL-04 is P1 (not P0, unlike DATA-03's S001/S002), REVIEW's own framing treats it as "a genuine, but
small and well-bounded" reuse item (an approved, already-transitive dependency swap, not new
architecture), and this story's own residual-risk expectation (`story.md`) is low once the parity/
fault-injection tests pass. A full independent-review task mirroring S001-T004/S002-T006's exhaustive
mandate-§14 checklist would be disproportionate to this story's size and risk (mandate §2.1,
"doability over theoretical completeness," and mandate §12's caution against "excessive
fragmentation into trivial tasks that provide no tracking value" apply here in the opposite
direction — over-provisioning review process for a two-task story is itself a form of
disproportionate structure). T003 is still a genuine, evidence-checking review — not skipped
entirely, per this story's own instruction to "not skip evidence/artifacts entirely" — but is scoped
narrowly to confirming both hand-rolled implementations are genuinely gone (not left in place
alongside the new library) and both tests are meaningful (assert real schedule/fault behavior, not
merely "does not error"), rather than running the full multi-point mandate §14 checklist used for
this epic's two P0 stories.
