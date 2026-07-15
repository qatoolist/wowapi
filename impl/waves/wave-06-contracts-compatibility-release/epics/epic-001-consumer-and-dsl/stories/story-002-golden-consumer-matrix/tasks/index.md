---
id: W06-E01-S002-TASKS-INDEX
type: tasks-index
parent_story: W06-E01-S002
status: done
derived: false
created_at: 2026-07-12
updated_at: 2026-07-14
---

# W06-E01-S002 — Tasks index

Per mandate §16.4. Task files are single-file per the repository's documented adaptation (see
`governance/naming-conventions.md` "Adaptation 1") — each task file below contains its task
definition, implementation record, verification record, and deviations record as internal sections.

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W06-E01-S002-T001](task-001-golden-consumer-scaffold-job.md) | Golden-consumer scaffold job | W06E01Impl | done | none | Versioned CLI installed through `go install` with provenance and no-replace assertions | AC-W06-E01-S002-01 | implemented | verified |
| [W06-E01-S002-T002](task-002-cross-module-subsystem-generation.md) | Cross-module, cross-subsystem generation | W06E01Impl | done | T001 | Two automatically wired modules generate and exercise all eight required subsystem types | AC-W06-E01-S002-02 | implemented | verified |
| [W06-E01-S002-T003](task-003-boot-and-exercise-real-infra.md) | Boot and exercise against real infrastructure | W06E01Impl | done | T002 | Generated API/worker pass CRUD, RLS, outbox, and restart recovery against Postgres/MinIO/Mailpit/Jaeger | AC-W06-E01-S002-03 | implemented | verified |
| [W06-E01-S002-T004](task-004-upgrade-from-previous-version-replay.md) | Upgrade-from-previous-version replay | W06E01Impl | done | T003 | Tagged v1.1.0 baseline and local candidate pass the two-pass build/boot contract; upgraded fixture passes real-infrastructure contracts | AC-W06-E01-S002-04 | implemented | verified |
| [W06-E01-S002-T005](task-005-ci-gate-wiring.md) | CI gate wiring | W06E01Impl | done | T004 | Ordinary CI and exact-SHA required-gates runner invoke the Wave-4 golden-consumer gate with all services including Jaeger | AC-W06-E01-S002-05 | implemented | verified |
| [W06-E01-S002-T006](task-006-independent-review.md) | Independent review | W06-E01-S002-Verify | done | T001, T002, T003, T004, T005 | EV-013 preserves the failed review; EV-014 records the passing fresh review | AC-W06-E01-S002-01, AC-W06-E01-S002-02, AC-W06-E01-S002-03, AC-W06-E01-S002-04, AC-W06-E01-S002-05 | review complete | verified |

## Grouping rationale

Per mandate §12: T001–T005 follow PLAN DX-04's own T1–T5 task table exactly, in the same
dependency order (T1 the scaffold, T2 generation, T3 boot-and-exercise, T4 upgrade replay, T5 CI-gate
wiring) — each produces a distinct, separately-evidenced output and each genuinely depends on its
predecessor's output existing first (generation needs an installed CLI; booting needs generated
modules; upgrade-replay needs a working baseline boot; gate-wiring needs a fully working fixture). This
story is P1 but carries substantial cross-infrastructure integration risk (RISK-W06-E01-001's broad
subsystem-coverage surface) and gates downstream work (W06-E02-S003's T7 leg) — T006 adds an
independent-review task per mandate §14's own framing that a story of this scope and downstream
consequence warrants review even at P1, consistent with how this programme treats any story whose
failure would silently propagate into another story's acceptance criteria.
