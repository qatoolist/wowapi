---
id: W01-E03-S001-T003
type: task
title: CSRF MaxBytesReader defensive bound (gosec G120 fix)
status: done
parent_story: W01-E03-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on: []
acceptance_criteria:
  - AC-W01-E03-S001-04
artifacts: []
evidence: []
---

# W01-E03-S001-T003 — CSRF MaxBytesReader defensive bound (gosec G120 fix)

## Task Definition

### Task objective

Wrap `kernel/httpx/csrf.go:118`'s `r.FormValue(p.FieldName)` call in a defensive
`http.MaxBytesReader` so the CSRF middleware bounds its own body read regardless of its position in
a product's own middleware chain, closing gosec's G120 finding.

### Parent story

W01-E03-S001 — Server timeouts and body bounds.

### Owner

Unassigned.

### Status

todo.

### Dependencies

None — independent of T001/T002.

### Detailed work

1. Confirm the exact current call site: `kernel/httpx/csrf.go:118`, inside `CSRFProtect`'s unsafe-
   method branch, `supplied = r.FormValue(p.FieldName)`.
2. Wrap `r.Body` in an `http.MaxBytesReader` before the `FormValue` call, choosing a bound consistent
   with the framework's existing body-size conventions (e.g. reusing `HTTP.MaxBodyBytes` if
   accessible in this middleware's scope, or a CSRF-appropriate smaller constant if `MaxBodyBytes`
   is not threaded through to `CSRFPolicy` — to be determined at implementation time; do not invent
   the exact bound value here).
3. Verify the fix does not change behavior for any request whose form body is already within the
   framework's other body-size guardrails (regression safety, per `plan.md` "Regression strategy").
4. Write a functional test proving an oversized form body is rejected (or its read is capped) rather
   than fully buffered.
5. Run a scoped gosec check against `kernel/httpx/csrf.go` (ad hoc, ahead of W01-E01-S002's broader
   linter enablement) confirming the G120 hit is resolved.

### Expected files or components affected

- `kernel/httpx/csrf.go` (`CSRFProtect`, unsafe-method branch).
- `kernel/httpx/csrf.go`'s existing test file(s) (`csrf_test.go`, `csrf_internal_test.go`) — new test
  case(s) added there, following the existing test-file split (external-behavior tests vs.
  internal-package tests).

### Expected output

`csrf.go`'s `FormValue` call is preceded by a defensive `http.MaxBytesReader` wrap; a passing test
proves an oversized form body is rejected; a scoped gosec run shows no G120 finding at this call
site.

### Required artifacts

None beyond the code diff.

### Required evidence

gosec G120 resolution (scoped run); functional oversized-body-rejection test output. See
`../../evidence/index.md`.

### Related acceptance criteria

AC-W01-E03-S001-04.

### Completion criteria

The defensive bound is implemented, tested, and the G120 finding is confirmed resolved by a scoped
gosec run.

### Verification method

Run the new CSRF test locally and in CI; run a scoped `gosec ./kernel/httpx/...` (or file-scoped
equivalent) and confirm the `csrf.go:118` finding is absent.

### Risks

Low — this is a narrowly-scoped defensive addition with no expected behavior change for in-bound
requests.

### Rollback or recovery considerations

Revert the single-file commit; no running-system state implicated.

## Implementation Record

Implemented 2026-07-13 by W01Http at SHA 0a31186cada5c275a588c74081cf977adf346e61 (working tree; conductor owns the wave commit).

- `kernel/httpx/csrf.go`: form-field fallback now wraps `r.Body` in
  `http.MaxBytesReader(w, r.Body, limit)` before `r.FormValue`; new `CSRFPolicy.MaxFormBytes`
  (0 → 1 MiB default constant `csrfDefaultMaxFormBytes`, matching default `http.max_body_bytes`
  so default-config behavior is unchanged for in-bound requests).
- Bound resolution: policy-level knob + constant default (the task's sanctioned option);
  `HTTP.MaxBodyBytes` is NOT threaded through SecurityChain — recorded as a known limitation in
  the story implementation.md, not a config-key invention.
- Tests: `TestCSRFOversizedFormBodyRejected` (fail-first captured — pre-fix the oversized body
  was fully buffered and served), `TestCSRFCustomMaxFormBytesOverridesDefault`; full CSRF suite
  green (in-bound regression proof via pre-existing `TestCSRFUnsafeMethodAcceptsFormField`).
- Scoped gosec run: 0 findings in kernel/httpx (rule-id caveat recorded in EV-004).

### Commits / pull requests

None yet — conductor owns the wave commit; working-tree diff recorded in the story's
implementation.md.

## Verification Record

| Acceptance criterion | Verification method | Environment | Result | Evidence | Reviewer |
|---|---|---|---|---|---|
| AC-W01-E03-S001-04 | Oversized-body fail-first pair + custom-bound test + scoped gosec | Local | PASS | EV-W01-E03-S001-004, -005 | pending |

Execution date 2026-07-13; revision 0a31186cada5c275a588c74081cf977adf346e61; environment local darwin/arm64 (go1.26.5);
reviewer pending (W01 wave review gate). Findings: none open. Retest: not required.
Final conclusion: task complete, ACs verified.

## Deviations Record

None — see the story-level deviations.md.
