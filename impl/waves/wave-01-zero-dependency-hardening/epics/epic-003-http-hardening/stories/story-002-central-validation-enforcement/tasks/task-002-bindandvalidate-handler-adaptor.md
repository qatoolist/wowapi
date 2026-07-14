---
id: W01-E03-S002-T002
type: task
title: BindAndValidate-calling generic handler adaptor
status: done
parent_story: W01-E03-S002
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on:
  - W01-E03-S002-T001
acceptance_criteria:
  - AC-W01-E03-S002-03
artifacts: []
evidence: []
---

# W01-E03-S002-T002 — BindAndValidate-calling generic handler adaptor

## Task Definition

### Task objective

Add a generic handler adaptor that calls the existing `BindAndValidate[T]` internally, so that
declaring a route's request contract (via T001's resolved mechanism) and wiring validation are the
same act — no separate manual `BindAndValidate` call is required for a handler built through this
adaptor.

### Parent story

W01-E03-S002 — Central validation enforcement.

### Owner

Unassigned.

### Status

todo.

### Dependencies

W01-E03-S002-T001 — the adaptor's design depends on which contract-declaration shape (Candidate A or
B) T001 resolves.

### Detailed work

1. Confirm T001's resolved shape before designing the adaptor's exact signature.
2. Implement the adaptor in `kernel/httpx` (likely `decode.go`, alongside `BindAndValidate`, or a new
   file — location `[ASSUMPTION]` per `plan.md`), composing `BindAndValidate[T]` internally without
   changing `BindAndValidate`'s own existing signature.
3. Ensure the adaptor surfaces `BindAndValidate`'s existing `KindValidation` error path unchanged (no
   new error shape) — the 400 response a caller sees through the adaptor must be byte-for-byte
   consistent with what a direct `BindAndValidate` caller already produces today.
4. Write a unit test proving the adaptor correctly binds and validates a valid request, and correctly
   rejects an invalid one with the expected field-error shape.
5. Write the adversarial integration-level test: register a route through the adaptor with a declared
   contract, POST an invalid DTO, assert HTTP 400 with field errors (AC-W01-E03-S002-03).

### Expected files or components affected

- `kernel/httpx/decode.go` (or a new file in the same package).
- A new or extended test file proving the adaptor's binding/validation behavior and the adversarial
  400 case.

### Expected output

A working generic adaptor function; a passing adversarial 400-with-field-errors test.

### Required artifacts

Handler adaptor (source code). See `../../artifacts/index.md`.

### Required evidence

Adversarial invalid-DTO 400 test output. See `../../evidence/index.md`.

### Related acceptance criteria

AC-W01-E03-S002-03.

### Completion criteria

The adaptor exists, composes `BindAndValidate` correctly, and the adversarial 400 test passes against
a fixture route built through it.

### Verification method

Run the adaptor unit test and the adversarial integration-level test locally and in CI.

### Risks

Low, contingent on T001's resolved shape being stable before this task starts — if T001's shape
changes after T002 has started, T002 may need rework (a natural consequence of the sequencing, not a
new risk beyond ordinary task dependency risk).

### Rollback or recovery considerations

Revert the adaptor commit; no running-system state implicated (this is a new, additive function).

## Implementation Record

Implemented 2026-07-13 by W01Http at SHA 0a31186cada5c275a588c74081cf977adf346e61 (working tree; conductor owns the wave commit).

- `httpx.ValidatedHandler[T](v *validation.Validator, maxBytes int64, fn func(w, r, req T))
  http.HandlerFunc` added to `kernel/httpx/decode.go`, composing the UNCHANGED
  `BindAndValidate[T]`; failures route through the existing `WriteError` KindValidation path —
  400 problem-details with field errors, byte-identical to a direct caller.
- Tests: adversarial `TestValidatedHandlerRejectsInvalidDTOWith400FieldErrors` (400, errors[0].
  field=="name", handler never ran) and `TestValidatedHandlerPassesValidDTOToBusinessLogic` —
  both RED under a deliberately validation-skipping stub (exactly the FBL-08 defect class),
  green after wiring BindAndValidate (fail-first captured).

### Commits / pull requests

None yet — conductor owns the wave commit; working-tree diff recorded in the story's
implementation.md.

## Verification Record

| Acceptance criterion | Verification method | Environment | Result | Evidence | Reviewer |
|---|---|---|---|---|---|
| AC-W01-E03-S002-03 | Adversarial invalid-DTO 400-with-field-errors, fail-first pair | Local | PASS | EV-W01-E03-S002-002 | pending |

Execution date 2026-07-13; revision 0a31186cada5c275a588c74081cf977adf346e61; environment local darwin/arm64 (go1.26.5);
reviewer pending (W01 wave review gate). Findings: none open. Retest: not required.
Final conclusion: task complete, ACs verified.

## Deviations Record

None — see the story-level deviations.md.
