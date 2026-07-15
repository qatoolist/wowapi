---
id: W01-E01-S002-TASKS-INDEX
type: tasks-index
parent_story: W01-E01-S002
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W01-E01-S002 — Tasks index

Per mandate §16.4. Task files are single-file per the repository's documented adaptation (see
`governance/naming-conventions.md` "Adaptation 1") — each task file below contains its task
definition, implementation record, verification record, and deviations record as internal sections.

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W01-E01-S002-T001](task-001-gosec-g704-annotation.md) | gosec G704 annotation (JWKS taint, 2 sites) | W01Lint | done | none | `kernel/auth/jwks.go:204,210` annotated with SEC-06-referencing `#nosec` justification | AC-W01-E01-S002-02 | complete (2026-07-13) | verified |
| [W01-E01-S002-T002](task-002-gosec-g115-review.md) | gosec G115 multi-site review (int-overflow conversions) | W01Lint | done | none | Enumerated site list across audit/database/jobs/mfa/pagination; each site annotated or bounds-checked | AC-W01-E01-S002-02 | complete (2026-07-13) | verified |
| [W01-E01-S002-T003](task-003-gosec-g304-annotation.md) | gosec G304 annotation (buildinfo file read) | W01Lint | done | none | Buildinfo-read site annotated as tool-only/low-risk | AC-W01-E01-S002-02 | complete (2026-07-13) | verified |
| [W01-E01-S002-T004](task-004-errorlint-fix.md) | errorlint fix (`kernel/httpx/middleware.go:54`) | W01Lint | done | none | `errors.Is` in place of `==` against `http.ErrAbortHandler` | AC-W01-E01-S002-03 | complete (2026-07-13) | verified |
| [W01-E01-S002-T005](task-005-exhaustive-annotations.md) | exhaustive annotations (2 sites, workflow package) | W01Lint | done | none | `kernel/workflow/definition.go:313` and `kernel/workflow/runtime.go:170` annotated, fail-closed default arm preserved | AC-W01-E01-S002-04 | complete (2026-07-13) | verified |
| [W01-E01-S002-T006](task-006-forcetypeassert-fixes.md) | forcetypeassert fixes (2 sites) | W01Lint | done | none | Checked (comma-ok) type assertions at `kernel/auth/jwks.go:112` and `kernel/config/bind.go:150` | AC-W01-E01-S002-05 | complete (2026-07-13) | verified |
| [W01-E01-S002-T007](task-007-usestdlibvars-nilerr-and-final-enablement.md) | usestdlibvars fixes, nilerr annotation, and final judged-set enablement | W01Lint | done | T001-T006 | usestdlibvars sites fixed (list from T001's fresh run); `kernel/policy/policy.go:166` nilerr annotation; permanent `.golangci.yml` enablement + final confirmation run + wrapcheck/revive-absence check | AC-W01-E01-S002-01, AC-W01-E01-S002-06, AC-W01-E01-S002-07 | complete (2026-07-13) | verified |

Note: T007 is the story's closure task, grouping three low-individual-risk work items no per-analyzer
triage task owns: the usestdlibvars mechanical fixes (no named sites in the source material — the
site list is produced by T001's fresh-run baseline), the `nilerr` non-finding annotation at
`kernel/policy/policy.go:166` (a real comment-only code change, adjudicated annotate-not-fix), and
the permanent `.golangci.yml` enablement with the final full-module-tree confirmation run and the
wrapcheck/revive-absence check. Grouping these three into one task rather than three preserves the
mandate §2.4 traceability chain (every acceptance criterion — including AC-01, AC-06, and AC-07 —
now has an owning task) without the excessive fragmentation mandate §12 warns against: each item
alone is a small, mechanical or documentation-adjacent unit whose separate tracking would add file
count but no independent blocking, ownership, or evidence value, while together they form a natural
"land everything, flip the switch, prove it" closure step that structurally mirrors
`W01-E01-S001`-T001's enablement-plus-confirmation shape.

## Task grouping rationale

Per the mandate §12 decomposition criteria this story's authoring guidance explicitly called out
("split when items need separate ownership... need separate evidence... have materially different
risks," balanced against "avoid excessive fragmentation into trivial tasks"), this story's gosec
triage is decomposed into **three** separate tasks (T001 G704, T002 G115, T003 G304) rather than one
combined "gosec triage" task, and the story as a whole has **seven** tasks rather than the five the
authoring instructions suggested as a starting point (three gosec tasks instead of one, plus the T007
closure task so no acceptance criterion lacks an owning task). The rationale, following the same discipline
`W01-E01-S001`'s own `tasks/index.md` used to document its four-task grouping:

- **G704 (T001) and G304 (T003) are single-site, single-disposition annotation tasks.** Each has one
  named site, one known disposition (annotate, referencing a specific governing fact — SEC-06 for
  G704, "tool-only/low-risk" for G304), and a narrow, already-fully-specified evidence requirement.
  Combining them into one task would not create tracking value beyond what combining them with G115
  would already fail to provide (see next point) — they are kept as two tasks, not one, because they
  reference different governing facts (SEC-06 vs. a general low-risk characterization) and touch
  different code areas (`kernel/auth/jwks.go` vs. a build-tooling file), which is enough material
  difference to warrant independent completion/evidence tracking without being trivial busywork —
  each is a single, complete, reviewable unit of triage.
- **G115 (T002) is deliberately split out from G704/G304 into its own task, rather than folded into
  one combined "gosec triage" task**, because it has a materially different risk profile and scope
  shape than the other two gosec items: it is a **multi-site review across five different packages**
  (audit, database, jobs, mfa, pagination) with **no site list enumerated yet** in the source
  material — the site list itself is a work product of this task, not an input to it. This is exactly
  the mandate §12 criterion "have materially different risks": a single-site annotation citing an
  already-ratified decision (G704/SEC-06) carries essentially no review risk, while an unenumerated,
  multi-package, per-site bounded-vs-unbounded judgment call (G115) carries real risk of an
  individual site being mis-disposed (annotated as "bounded" when it is not actually bounded by prior
  validation). Combining G115 into a single "gosec triage" task with G704/G304 would obscure this
  materially higher-risk, materially larger-scope item behind two much smaller, much lower-risk items
  in the same task's completion/evidence record — exactly the failure mode mandate §12 is warning
  against ("need separate evidence... can block independently"). G115 alone could plausibly block or
  slip the story's schedule while G704/G304 do not; tracking that independently has real value.
- **This is a deliberate deviation from the suggested "grouped gosec triage" 5-task starting point**
  offered as one option in this story's authoring instructions. The instructions explicitly permitted
  this judgment call ("your call, document it... if you split G115 out separately... adjust and
  document why you deviated from the suggested 5-task count"). The deviation is: 3 gosec tasks
  (T001/T002/T003) instead of 1, plus the T007 closure task described in the note above, making the
  story's total 7 tasks instead of 5. This is recorded here,
  in this index, as the authoritative rationale — no separate `deviations.md` entry is needed for this
  choice since it is a planning-time judgment call within the discretion this story's authoring
  instructions explicitly granted, not a divergence from an already-approved plan (`deviations.md` is
  reserved for divergence between an approved `plan.md` and actual implementation, per mandate §2.6;
  this choice is made once, here, before the plan is otherwise finalized).
- **errorlint (T004), exhaustive (T005), and forcetypeassert (T006) are each kept as their own task**
  for the same reason `W01-E01-S001` kept its noctx and copyloopvar fixes as separate tasks from its
  linter-enablement task: each is a materially different analyzer with its own named site(s), its own
  fail-before/pass-after evidence pair, and — critically for exhaustive and forcetypeassert — its own
  distinct fix/annotation *shape* (exhaustive is annotation-preserving-design-intent; forcetypeassert
  is a real code fix with a false-ok error-handling decision at each site). Splitting further (e.g.
  one task per forcetypeassert site) would cross into the fragmentation mandate §12 warns against,
  since both forcetypeassert sites share one disposition shape (checked assertion) and are naturally
  reviewed and evidenced together.
- No further splitting (e.g. one task per G115 site once enumerated) is planned — the G115 site list,
  once produced by T002's fresh run, is reviewed and disposed within T002 as a single task whose
  *output* is a multi-row triage record, not as N separate tasks. Splitting per-site would multiply
  file count against an as-yet-unknown N with no additional tracking value beyond what T002's own
  per-site triage record (required by AC-W01-E01-S002-02) already provides.
