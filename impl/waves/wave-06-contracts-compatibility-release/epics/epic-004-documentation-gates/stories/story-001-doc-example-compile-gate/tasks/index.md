---
id: W06-E04-S001-TASKS-INDEX
type: tasks-index
parent_story: W06-E04-S001
status: complete
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W06-E04-S001 — Tasks index

Per mandate §16.4. Task files are single-file per the repository's documented adaptation (see
`governance/naming-conventions.md` "Adaptation 1") — each task file below contains its task
definition, implementation record, verification record, and deviations record as internal sections.

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W06-E04-S001-T001](task-001-build-extractor-and-tag-examples.md) | Build the docexamples extractor tool and tag existing normative examples | W06E04Impl | complete | none | Extractor tool + classified examples | AC-W06-E04-S001-01 | complete | PASS (EV-W06-E04-S001-001) |
| [W06-E04-S001-T002](task-002-ci-wiring-and-adversarial-fixture.md) | Wire into CI and make docs-check; adversarial staled-example fixture | W06E04Impl | complete | T001 | CI-enforced gate proven fail-first | AC-W06-E04-S001-02, AC-W06-E04-S001-03 | complete | PASS (EV-W06-E04-S001-002/003) |
| [W06-E04-S001-T003](task-003-independent-review.md) | Independent review | W06-E01-E04-Execution.W06E04ReviewR | complete | T001, T002 | Independent-review PASS, no issues | AC-W06-E04-S001-01, AC-W06-E04-S001-02, AC-W06-E04-S001-03 | complete | PASS (REV-W06-E04-S001-001) |

## Grouping rationale

Per mandate §12: T001 (build the tool + tag examples) and T002 (CI wiring + adversarial fixture) are
kept separate because they produce distinct outputs — T001's a working extractor and a tagged doc set;
T002's a CI-enforced gate with fail-first proof. This story is P2 but this epic's task brief still adds
an independent-review task (T003) given the recurrence-prevention purpose this story serves (MATRIX
CS-22's own framing: documentation drift "already happened twice at reviewer-visible severity") — a
story whose entire purpose is preventing a repeat of a twice-occurred defect class warrants review even
at P2.
