---
id: GOV-ARTIFACT-POLICY
type: governance
title: Artifact policy — definition, index fields, and lifecycle stages
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
derived: false
---

# Artifact policy

Mandate §9, "Artifact-management requirements." Artifacts and evidence are tracked separately:
an artifact is a *thing*; evidence (`evidence-policy.md`) is *proof about* a thing or behavior.

## 9.1 Artifact definition (mandate §9.1, verbatim)

> An artifact is something consumed, produced, modified, or delivered as part of implementation.

Examples (mandate §9.1, verbatim list):

source-code packages · interfaces · schemas · migrations · API specifications · architecture
diagrams · generated code · configuration examples · deployment manifests · runbooks ·
compatibility matrices · design documents · release notes · binaries · benchmark definitions ·
migration utilities.

## 9.2 Artifact index — required fields (mandate §9.2, verbatim list)

Every story must contain an `artifacts/index.md` with entries containing:

- artifact ID;
- title;
- type;
- lifecycle stage;
- description;
- source requirement;
- producing task;
- repository path or storage location;
- version;
- checksum where appropriate;
- status;
- reviewer;
- retention requirement.

Artifact IDs follow the naming pattern `ART-<story-id>-NNN` (see `naming-conventions.md`).
"Checksum where appropriate" applies to generated/binary artifacts where content-addressability
matters (see the no-duplication rule below); it is not required for e.g. a design document whose
authoritative copy is the repository file itself.

## 9.3 Artifact lifecycle stages (mandate §9.3, verbatim)

Artifacts are organised into three stage directories:

```text
pre-implementation/
implementation/
post-implementation/
```

### Pre-implementation (mandate §9.3, verbatim example list)

current-state baseline · existing behaviour report · architecture baseline · compatibility
baseline · performance baseline · source inventory · risk assessment · design inputs.

### Implementation (mandate §9.3, verbatim example list)

schemas · interfaces · migrations · generated outputs · code-generation definitions ·
configuration changes · architecture decisions · implementation notes.

### Post-implementation (mandate §9.3, verbatim example list)

migration results · release notes · runbooks · upgrade guidance · compatibility reports · final
architecture diagrams · acceptance packages · closure reports.

An artifact's "lifecycle stage" index field must be one of these three values and must match the
stage subdirectory it is registered under.

## No-duplication rule (mandate §9.3, verbatim)

> Do not duplicate large generated artifacts unnecessarily. Register their authoritative path,
> version, size, checksum, and generation command.

Practically: a generated OpenAPI spec, a large migration output log, or a binary is not copied
into the story's `artifacts/` tree. `artifacts/index.md` records where the authoritative copy
lives (repository path, or external storage location), its version, size, checksum, and the exact
command used to (re)generate it — enough for anyone to reproduce or verify it without the story
directory carrying a second copy.

## Why `artifacts/` subdirectories are not pre-populated

Story creation ships only `artifacts/index.md` (declaring the applicable categories per §9.2/§9.3
up front) — not empty `pre-implementation/`, `implementation/`, `post-implementation/`
subdirectories. The specific subdirectory is created only when the first artifact of that stage
actually exists. This is Adaptation 2 to the mandate's baseline directory tree; full rationale
(git's non-tracking of empty directories, mandate §18's placeholder-content rule) is in
`naming-conventions.md`.
