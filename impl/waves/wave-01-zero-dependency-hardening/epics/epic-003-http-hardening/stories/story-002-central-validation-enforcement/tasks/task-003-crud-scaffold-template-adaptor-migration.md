---
id: W01-E03-S002-T003
type: task
title: crud/scaffold template migration to the adaptor
status: done
parent_story: W01-E03-S002
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on:
  - W01-E03-S002-T001
  - W01-E03-S002-T002
acceptance_criteria:
  - AC-W01-E03-S002-03
artifacts: []
evidence: []
---

# W01-E03-S002-T003 — crud/scaffold template migration to the adaptor

## Task Definition

### Task objective

Update the crud/scaffold code-generation templates so generated mutating-route handlers use the new
adaptor (T002) and declare a request contract (T001), rather than the previous manual
`BindAndValidate` call pattern (if the current templates use one) or no validation call at all.

### Parent story

W01-E03-S002 — Central validation enforcement.

### Owner

Unassigned.

### Status

todo.

### Dependencies

W01-E03-S002-T001, W01-E03-S002-T002 — both the contract-declaration mechanism and the adaptor must
exist before templates can be migrated to use them.

### Detailed work

1. Locate the crud/scaffold template file(s) that currently generate mutating-route handler code —
   not yet confirmed as of this task's writing; likely inside `internal/cli/templates/` in a
   `crud`-specific subdirectory distinct from S001's `init` templates, but this must be confirmed at
   task start, not assumed.
2. Determine the current state: does the existing crud template already call `BindAndValidate`
   manually, or does it currently skip validation entirely? (This determines whether this task is a
   "migrate an existing call" or "add a call that wasn't there" — the story's own "Problem statement"
   suggests the framework-wide pattern is opt-in/often-skipped, but the crud template's specific
   current state should be confirmed, not assumed, before editing.)
3. Update the template to declare the request contract on the generated route's `RouteMeta` (per
   T001's resolved mechanism) and to wire the handler through T002's adaptor.
4. Extend whatever generator-output-boots test infrastructure DX-02 (W01-E04-S001, a sibling epic's
   story) builds or already has, so the migrated crud template's generated handlers are proven to
   boot successfully and to actually validate — reuse that shared test harness rather than building a
   parallel one, consistent with the "shared primitive" framing `wave.md` uses for DX-01 T5/DX-02.
5. Update any generated documentation/comments in the crud template that reference the old pattern.

### Expected files or components affected

- crud/scaffold template file(s) (exact path(s) to be confirmed at task start).
- Generator-output test infrastructure (shared with DX-02, W01-E04-S001 — coordinate, do not
  duplicate).

### Expected output

The crud template generates mutating-route handlers that declare a request contract and use the new
adaptor; generated output boots successfully and enforces validation, proven by an extended
generator-output test.

### Required artifacts

Updated crud/scaffold template. See `../../artifacts/index.md`.

### Required evidence

Generator-output test proving the migrated template's output boots and validates correctly. See
`../../evidence/index.md`.

### Related acceptance criteria

AC-W01-E03-S002-03 (transitively — the crud-generated route is one concrete instance of a route built
through the adaptor; the story's adversarial 400 proof may be satisfied either by a dedicated fixture
route (T001/T002's own tests) or additionally by the crud-generated route this task produces — the
story's AC does not require both, but this task's own evidence should show the crud-generated route
also passes the same adversarial check as a template-correctness proof).

### Completion criteria

The crud template is migrated; the generator-output test (shared or extended) proves the migrated
output boots and enforces validation.

### Verification method

Run the generator-output test locally and in CI, coordinating with DX-02's (W01-E04-S001) test
infrastructure if it exists by this task's start.

### Risks

Low — this is the last task in the sequence, depending on both T001 and T002 being stable; the main
risk is coordination overhead with the sibling W01-E04-S001 story's own generator-output-boots test
infrastructure if that story's timeline diverges from this one's.

### Rollback or recovery considerations

Revert the template commit; no running-system state implicated (affects only future `wowapi gen crud`
invocations, same delivery-model reasoning as S001's scaffold-template fix).

## Implementation Record

Implemented 2026-07-13 by W01Http at SHA 0a31186cada5c275a588c74081cf977adf346e61 (working tree; conductor owns the wave commit).

- Located: `internal/cli/templates/crud/resource.go.tmpl` (detailed-work step 1). Current state
  (step 2): the pre-fix template skipped validation entirely (TODO-stub create/update) — this was
  an "add the call that wasn't there" migration.
- Migration: emits `Create<R>Request`/`Update<R>Request` structs (`validate:"required"` starter
  tags + adjust-me comment), declares them on POST/PUT `RouteMeta.Request`, wires handlers
  through `httpx.ValidatedHandler(v, 1<<20, ...)` with `v := mc.Validator()`; registration doc
  comment explains the enforcement flag. W01Gen's `.deactivate` DELETE line preserved
  (irc-coordinated; disjoint edits, no conflict).
- Test: `TestGenCRUDMutatingRoutesDeclareContractsAndUseValidatedHandler` in the shared
  scaffold_test.go harness (coordinated with W01Gen's new gen-crud boots test in the same
  package — no duplicate harness); full `go test -race ./internal/cli/` green.

### Commits / pull requests

None yet — conductor owns the wave commit; working-tree diff recorded in the story's
implementation.md.

## Verification Record

| Acceptance criterion | Verification method | Environment | Result | Evidence | Reviewer |
|---|---|---|---|---|---|
| AC-W01-E03-S002-03 (transitive) | Generator-output assertion test + package suite | Local | PASS | EV-W01-E03-S002-004 | pending |

Execution date 2026-07-13; revision 0a31186cada5c275a588c74081cf977adf346e61; environment local darwin/arm64 (go1.26.5);
reviewer pending (W01 wave review gate). Findings: none open. Retest: not required.
Final conclusion: task complete, ACs verified.

## Deviations Record

None — see the story-level deviations.md.
