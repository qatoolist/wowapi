---
id: IMPL-W06-E04-S002
type: implementation-record
parent_story: W06-E04-S002
status: complete
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Implementation record — W06-E04-S002

Implemented 2026-07-13 by W06E04Impl against HEAD lineage
`733ef3e930cbb3f89f5bbc53d8f562c60e426513` plus the shared W05/W06 working tree.

## What was actually implemented

- `internal/tools/docexamples/reference.go` consumes W05 AR-03's authoritative
  `appmodel.GenerateProjections` export for the canonical requests manifest. The generated
  `docs/reference/application-model.md` contains only the exported table bytes plus the canonical
  final newline. `checkReference` rejects drift; `-write-reference` regenerates intentionally.
- `internal/tools/docexamples/future.go` lints `docs/blueprint/*.md` and documentation files named
  `*-target-design.md`. Future/planned/target headings must be immediately followed by a line
  containing `Target, not implemented`; fenced examples are ignored. Existing blueprint future
  sections and W06-E01's `docs/implementation/module-dsl-target-design.md` pass this convention.
- Unlabeled/labeled fixtures prove the lint's negative and positive contracts; the byte-golden test
  compares the generated reference file directly with AR-03's projection output.
- The combined gate is invoked by `make docs-check` and the CI unit job.

## Components and files changed

`internal/tools/docexamples/{reference.go,future.go,main.go,main_test.go,testdata/future-*.md}`;
`docs/reference/application-model.md`; future-state labels in `docs/blueprint/00-overview.md` and
`docs/blueprint/01-domain-model.md`; blueprint-11 invocation documentation; shared Makefile/CI hooks;
story-local artifact/evidence/lifecycle records.

## Interfaces introduced or changed

Documentation-authoring contract: a future/planned/target design heading in the lint scope is followed
by `> **Target, not implemented.**`. Generated reference changes are made with
`go run ./internal/tools/docexamples -write-reference` and verified with `make docs-check`.

## Configuration, schema, security, and observability changes

No runtime configuration, schema, or security surface changed. Diagnostics include exact file/line
locations for unlabeled future-state headings and an actionable regeneration command for stale output.

## Tests added or modified

Focused tests cover AR-03 export byte equality, on-disk currency, unlabeled rejection,
labeled/current prose acceptance, fenced-content exclusion, and repository-wide gate integration.

## Commits and pull requests

No commit or pull request was created; the conductor owns integration of the shared working tree.

## Technical debt and known limitations

W05 AR-03's implementation is present and its owner confirmed `GenerateProjections` and the AR-03
golden tests as authoritative, but W05's story/task lifecycle records remain draft/todo in this shared
workspace. The implementation proceeded under the user's explicit available-export rule; see
`deviations.md` DEV-W06-E04-S002-001.

## Relationship to the approved plan

The mechanisms match `plan.md`. Starting T4 before W05 bookkeeping reached `accepted` is the sole
recorded deviation; the technical prerequisite was present and directly exercised rather than absent
or mocked.
