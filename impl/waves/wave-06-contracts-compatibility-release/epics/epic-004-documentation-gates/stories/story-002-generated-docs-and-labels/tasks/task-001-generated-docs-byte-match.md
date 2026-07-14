---
id: W06-E04-S002-T001
type: task
title: Generated reference docs byte-matching AR-03's model export (blocked on W05-E03)
status: complete
parent_story: W06-E04-S002
owner: W06E04Impl
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on:
  - W05-E03
acceptance_criteria:
  - AC-W06-E04-S002-01
artifacts:
  - ART-W06-E04-S002-001
evidence:
  - EV-W06-E04-S002-001
---

# W06-E04-S002-T001 — Generated reference docs byte-matching AR-03's model export (blocked on W05-E03)

## Task Definition

### Task objective

Generate reference/API docs from AR-03's authoritative manifest so the generated reference tables byte-match the model export. BLOCKED: cannot begin until W05-E03 (AR-03) reaches accepted.

### Parent story

W06-E04-S002

### Owner

unassigned

### Status

todo

### Dependencies

W05-E03 (AR-03 remainder, cross-wave) must be `accepted` — this task hard-depends on AR-03's own delivered model-export format existing first, per PLAN AR-05 T4's own dependency row ('AR-03 T1, T5'). This task must not begin before that entry criterion is satisfied.

### Detailed work

1. Confirm W05-E03 has reached `accepted`.
2. Confirm AR-03's own delivered model-export format.
3. Build a reference-doc-generation pipeline consuming that model export.
4. Write an integration golden-diff test proving the generated reference tables byte-match the model
   export.

### Expected files or components affected

A new reference-doc-generation pipeline (exact location TBD, dependent on AR-03's own delivered package structure).

### Expected output

Generated reference tables that byte-match AR-03's model export, proven by an integration golden-diff test.

### Required artifacts

ART-W06-E04-S002-001 (reference-doc-generation pipeline).

### Required evidence

EV-W06-E04-S002-001 (integration golden-diff test report).

### Related acceptance criteria

AC-W06-E04-S002-01.

### Completion criteria

The generated reference tables byte-match the model export, once unblocked.

### Verification method

Direct execution of the integration golden-diff test, once W05-E03 is accepted.

### Risks

RISK-W06-E04-001 (this task cannot begin until W05-E03 resolves) — see epic-level `risks.md`.

### Rollback or recovery considerations

If begun prematurely (before W05-E03 genuinely reaches accepted), halt and record a deviation; do not silently proceed against an unaccepted upstream dependency.

## Implementation Record

Implemented a generator/checker consuming W05 AR-03's delivered
`appmodel.GenerateProjections(canonicalManifest()).Doc` bytes. The generated
`docs/reference/application-model.md` is the exported table plus one canonical newline;
`-write-reference` regenerates and the normal gate rejects stale bytes.

- **Files changed:** `internal/tools/docexamples/reference.go`, `main.go`, `main_test.go`, and
  `docs/reference/application-model.md`.
- **Tests:** `TestGeneratedReferenceByteMatchesAuthoritativeExport` compares both rendered and
  on-disk bytes directly to the AR-03 export.
- **Implementation date:** 2026-07-13.
- **Commits/PRs:** none; conductor owns integration.
- **Plan relationship:** mechanism matched; entry occurred before W05 lifecycle bookkeeping reached
  `accepted`, recorded as DEV-W06-E04-S002-001.

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W06-E04-S002-01 | focused byte-golden test | macOS arm64, Go 1.26.5 | Generated table byte-matches AR-03 export | integration golden-diff report | pending W06-E04-S002-T003 |

- **Actual result:** PASS — rendered and checked-in bytes equal AR-03's `Doc` export plus canonical newline.
- **Evidence identifier:** EV-W06-E04-S002-001.
- **Execution date/revision:** 2026-07-13; `733ef3e930cbb3f89f5bbc53d8f562c60e426513` plus shared changes.
- **Environment:** macOS Darwin 25.5.0 arm64; Go 1.26.5.
- **Retest status:** focused test and combined gate passed.
- **Final conclusion:** technical proof passed; independent review and upstream bookkeeping
  disposition remain to be recorded.

## Deviations Record

DEV-W06-E04-S002-001: the exact AR-03 export and golden tests were present and owner-confirmed, but
W05 story/task status fields remained draft/todo. The user's available-export rule authorized T4;
see story `deviations.md` for impact and controls.
