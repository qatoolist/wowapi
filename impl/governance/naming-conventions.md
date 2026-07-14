---
id: GOV-NAMING-CONVENTIONS
type: governance
title: Naming conventions — identifier rules and documented adaptations to the mandate baseline
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
derived: false
---

# Naming conventions

Mandate §5, "Identifier and naming rules," plus two documented adaptations to the mandate's
baseline directory structure (§3), recorded per mandate §18: "Record assumptions explicitly."

## Identifier rules (mandate §5, verbatim)

- Never reuse an identifier.
- Never renumber existing identifiers merely to improve ordering.
- Use zero-padded numbers.
- Use descriptive directory names.
- Keep filenames stable and generic within an identified directory.
- Avoid ambiguous names such as `story_0.md`, `impl_0.md`, `related artifacts.md`, or `misc.md`.

These rules are absolute: once `W02-E01-S001` exists, it keeps that ID even if the epic is later
reordered, split, or partially cancelled. Gaps in numbering (e.g. a cancelled story leaving
`S002` unused) are acceptable and expected — do not renumber later stories to close the gap.

## Example ID patterns (mandate §5, verbatim)

```text
W00
W00-E01
W00-E01-S001
W00-E01-S001-T001

REQ-ARCH-001
AC-W00-E01-S001-01
ART-W00-E01-S001-001
EV-W00-E01-S001-001
ADR-W00-E01-S001-001
RISK-W00-E01-S001-001
DEV-W00-E01-S001-001
TD-W00-E01-S001-001
```

Reading the grammar: wave (`W<NN>`) → epic (`-E<NN>`) → story (`-S<NNN>`) → task (`-T<NNN>`).
Acceptance criteria, artifacts, evidence, ADRs, risks, deviations, and technical debt are each
their own ID namespace, scoped by suffixing the owning wave/epic/story/task ID — never reusing a
bare running counter across the whole programme (that would violate "never reuse an identifier"
the moment two stories both wanted "item 001").

Source-level requirement IDs (e.g. `REQ-ARCH-001`, or the existing `AR-01`/`SEC-01`/`DATA-01`-
style IDs already used in `impl/analysis/requirement-inventory.md`) are retained unchanged from
the source documents per mandate §1.1/§5 — they are not renumbered into the `W/E/S/T` grammar.

## Preferred file layout example (mandate §5, verbatim)

```text
story-001-runtime-lifecycle/
├── story.md
├── plan.md
├── implementation.md
├── verification.md
└── deviations.md
```

Directory names are descriptive (`story-001-runtime-lifecycle`, not `story-001`); filenames
inside an identified directory are stable and generic (`story.md`, not
`runtime-lifecycle-story.md`) — the directory name carries the descriptive part, the filename
carries the document type.

---

## Adaptations to the mandate's baseline directory structure

Documented per mandate §18: "Record assumptions explicitly." Two adaptations are made to the
example tree in mandate §3. Both preserve every required-content field from §8/§9/§10 — nothing
mandated is dropped; only file/directory granularity is coarsened where it added no traceability
value.

### Adaptation 1 — flat task files, not 4-file task directories

**What the mandate's example shows.** Mandate §3's tree gives each task its own directory:

```text
tasks/
└── task-001-<name>/
    ├── task.md
    ├── implementation.md
    ├── verification.md
    └── deviations.md
```

**What this programme does instead.** One file per task:

```text
story-XXX/tasks/task-NNN-<descriptive-name>.md
```

containing all four required sections — task definition (§8.6), implementation record (§8.7),
verification record (§8.8), deviations record (§8.9) — as internal `##` sections within the
single file.

**Rationale.**

- Mandate §12, verbatim: "avoid excessive fragmentation into trivial tasks that provide no
  tracking value." A task is normally a bounded, single-owner unit of work; splitting its own
  record across 4 near-empty files multiplies file count roughly 4x with no traceability benefit.
- Mandate §1.1/design intent favors doability over theoretical completeness (§2.1). Four files
  per task across a programme with potentially hundreds of tasks is exactly the kind of
  structural overhead that trades tracking value for file-count churn.
- All four §8.6–§8.9 required content fields are preserved verbatim as sections within the one
  file — nothing from the mandate's required content is dropped. The front-matter `id`, `status`,
  `depends_on`, and `related acceptance criteria` fields still make each section independently
  queryable (a tool can parse `## Implementation` out of `task-003-foo.md` exactly as it would
  parse `implementation.md` out of a `task-003-foo/` directory).
- Only the file-boundary is coarsened at the **task** level. Stories, epics, and waves keep their
  multi-file structure exactly as specified in the mandate's example tree, since those carry
  materially more independent content (a story's `plan.md` is substantial and reviewed
  separately from its `verification.md`; a task's four sections are typically short enough that
  splitting them adds no reviewability).

### Adaptation 2 — `evidence/` and `artifacts/` subdirectories created on first use, not pre-populated empty

**What the mandate's example shows.** Mandate §3's tree pre-creates, at story creation time:

```text
evidence/
├── index.md
├── baselines/
├── tests/
├── coverage/
├── logs/
├── screenshots/
├── benchmarks/
├── security/
├── static-analysis/
├── compatibility/
├── regression/
├── reviews/
└── acceptance/

artifacts/
├── index.md
├── pre-implementation/
├── implementation/
└── post-implementation/
```

— 12 empty evidence subdirectories and 3 empty artifact subdirectories per story, before any
work has started.

**What this programme does instead.** Story creation ships only `evidence/index.md` and
`artifacts/index.md`. A specific subdirectory (e.g. `evidence/coverage/`) is created only when
the first item of that category actually exists for that story.

**Rationale.**

- Mandate §18, verbatim: "Do not create placeholder files containing only headings where
  meaningful planning content can be derived." Empty directories are the directory-level
  equivalent of that anti-pattern.
- Git does not track empty directories at all. Pre-creating all 15 subdirectories per story
  either requires placeholder `.gitkeep` files in every one (pure noise, especially multiplied
  across dozens of stories) or the directories silently fail to exist in the repository until
  first use regardless — so "pre-creating" them is not even achievable without the `.gitkeep`
  workaround, and that workaround has no informational value.
- `index.md` in both `evidence/` and `artifacts/` still declares and describes **all applicable
  categories up front**, per the §9.2/§10 required-field lists — so nothing about the story's
  evidence/artifact *scope* is hidden or deferred. Only the physical subdirectory creation is
  deferred to first real content; the plan for what categories will be populated is visible from
  story creation onward.

### Scope of the adaptations

Both adaptations apply only to the two specific structural points described above. Every other
part of the mandate's §3 directory tree — wave/epic/story file sets, `decisions/` under each
story, the `tasks/index.md` roll-up, the tracking-level registers — is implemented exactly as the
mandate specifies.
