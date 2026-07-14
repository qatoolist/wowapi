---
id: GOV-TEMPLATE-ARTIFACT
type: template
title: Artifact index entry template
status: template
created_at: 2026-07-12
updated_at: 2026-07-12
derived: false
---

<!--
Template for a single artifact index entry. Copy into the appropriate lifecycle subdirectory
under `.../story-<NNN>-<name>/artifacts/{pre-implementation,implementation,post-implementation}/`
and reference it from `artifacts/index.md`. Fields per mandate §9.2. Do not duplicate large
generated artifacts unnecessarily — register their authoritative path, version, size, checksum,
and generation command instead of copying the content here.
-->

---
id: <ART-W NN-E NN-S NNN-NNN>
type: artifact
title: <Artifact title>
status: template
created_at: 2026-07-12
updated_at: 2026-07-12
derived: false
---

# <ART-W NN-E NN-S NNN-NNN> — <Artifact title>

## Artifact ID

*State the stable artifact identifier.*

## Title

*State the artifact title.*

## Type

*State the artifact type (e.g. source-code package, interface, schema, migration, API specification, architecture diagram, generated code, configuration example, deployment manifest, runbook, compatibility matrix, design document, release notes, binary, benchmark definition, migration utility).*

## Lifecycle stage

*State the lifecycle stage: pre-implementation, implementation, or post-implementation.*

## Description

*Describe what this artifact is and what it captures.*

## Source requirement

*State the source requirement ID(s) this artifact relates to.*

## Producing task

*State the task ID that produced this artifact.*

## Repository path or storage location

*State the authoritative path or storage location of this artifact.*

## Version

*State the version of this artifact.*

## Checksum where appropriate

*State a checksum for this artifact, where appropriate.*

## Status

*State the status of this artifact (e.g. draft, current, superseded).*

## Reviewer

*State who reviewed this artifact.*

## Retention requirement

*State how long this artifact must be retained and why.*
