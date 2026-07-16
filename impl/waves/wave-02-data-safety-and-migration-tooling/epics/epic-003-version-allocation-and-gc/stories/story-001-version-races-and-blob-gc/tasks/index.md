---
id: W02-E03-S001-TASKS-INDEX
type: tasks-index
parent_story: W02-E03-S001
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W02-E03-S001 — Tasks index

Per mandate §16.4. Task files are single-file per the repository's documented adaptation (see
`governance/naming-conventions.md` "Adaptation 1") — each task file below contains its task
definition, implementation record, verification record, and deviations record as internal sections.

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W02-E03-S001-T001](task-001-locked-counter-version-allocation.md) | Locked-counter/sequence-row version allocation (both packages) | unassigned | done | none | Race-free version-allocation mechanism applied to `kernel/artifact` and `kernel/document` | AC-W02-E03-S001-01 | implemented | verified |
| [W02-E03-S001-T002](task-002-durable-upload-session-records.md) | Durable upload-session records | unassigned | done | T001 | Upload-session table + persistence before URL issuance | AC-W02-E03-S001-02 | implemented | verified |
| [W02-E03-S001-T003](task-003-atomic-confirmation-cas.md) | Atomic confirmation CAS | unassigned | done | T001, T002 | Confirmation path that CASes session + version atomically | AC-W02-E03-S001-03 | implemented | verified |
| [W02-E03-S001-T004](task-004-scheduled-gc-sweep.md) | Scheduled GC sweep | unassigned | done | T002, T003 | GC sweep mechanism with metrics/audit | AC-W02-E03-S001-04 | implemented | verified |
| [W02-E03-S001-T005](task-005-artifact-version-mirror-fix.md) | `kernel/artifact.Generate` mirror fix and dedicated test | unassigned | done | T001 | Same counter mechanism applied to `kernel/artifact.Generate`, with its own dedicated concurrency test | AC-W02-E03-S001-05 | implemented | verified |

## Grouping rationale

Per PLAN DATA-05's own task table, T1–T5 are five tasks with distinct outputs and (mostly) distinct
evidence: T1 (`DATA-05/version-allocation/`), T2 (`DATA-05/upload-session/`), T3
(`DATA-05/confirm-cas/`), T4 (`DATA-05/gc-sweep/`), T5 (`DATA-05/artifact-version/`) — five separate
evidence directories named directly in the source, which this programme keeps as five separate tasks
rather than merging any pair, consistent with mandate §12's decomposition guidance ("Tasks must be
decomposed when they... need separate evidence... can block independently... have materially
different risks"). T5 is kept as its own task rather than folded into T1 despite both targeting "the
same counter fix": T1's own row scopes the mechanism to "both `kernel/artifact` and
`kernel/document`," but T5 exists specifically to hold `kernel/artifact.Generate` to "the same
concurrency bar as T1" via its own dedicated mirror test — i.e., T5 is the acceptance gate proving
`kernel/artifact`'s application of the shared mechanism is independently verified, not merely
implemented as an assumed side effect of T1's own (kernel/document-focused) concurrency test. This
distinction is treated as real, not cosmetic, per this story's `epic.md` "Architectural context"
section, and is preserved as separate task tracking so `kernel/artifact`'s own proof cannot be
silently skipped if T1's test happens to only exercise `kernel/document`.

**No dedicated independent-review task is added.** DATA-05 is P1 per `requirement-inventory.md`
("Version allocation races + blob GC (T1–T5) | IMPL | P1 | planned | W02-E03-S001 |"). This
programme's convention, established at W02-E01-S001 (a P0 story), is that only P0 stories
automatically receive a dedicated independent-review task; for a P1 story, the default — per this
wave's task brief and mandate §12's "avoid excessive fragmentation into trivial tasks that provide no
tracking value" — is no separate review task unless task count or risk genuinely warrants one. This
story's five tasks form a single, tightly-coupled persistence-correctness fix in one package pair
(`kernel/artifact`, `kernel/document`), explicitly framed by `epic.md` as a "single reviewer domain"
per `impl/analysis/wave-allocation-detail.md`'s own allocation note — meaning one reviewer with
domain expertise in this exact persistence path is expected to review the story as a coherent whole,
not that a dedicated review task is required to formalize that review. RISK-W02-E03-001 (counter-row
contention) is a real but singular, already-named risk with its own required evidence (measured lock
wait as part of T1/T5's own tests), not a diffuse set of risks across unrelated tasks that would
justify a dedicated review task the way W02-E01-S001's DoS-adjacent unbounded-retry concern
(alongside a separate schema-design risk) did. Judgment applied: **5 tasks only.**
