---
id: W01-E03-S001-T002
type: task
title: Prod-profile zero-timeout config.Validate rejection
status: done
parent_story: W01-E03-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on:
  - W01-E03-S001-T001
acceptance_criteria:
  - AC-W01-E03-S001-03
artifacts: []
evidence: []
---

# W01-E03-S001-T002 — Prod-profile zero-timeout config.Validate rejection

## Task Definition

### Task objective

Reject an explicit zero value for any of the 4 new HTTP timeout keys, following the validation
policy resolved for this story (see `plan.md` unresolved question 1: unconditional rejection,
matching the 3 pre-existing HTTP timeout keys, vs. prod-profile-only rejection, matching the
SSRF-disable precedent), and prove it with a test extending the existing
`unsafe_config_matrix_test.go` table-driven pattern.

### Parent story

W01-E03-S001 — Server timeouts and body bounds.

### Owner

Unassigned.

### Status

todo.

### Dependencies

W01-E03-S001-T001 — the 4 config keys must exist before this task can add rejection logic for them.

### Detailed work

1. Resolve `plan.md` unresolved question 1 (validation policy) before implementing — confirm with
   review whether unconditional or prod-profile-only rejection applies to the new keys, and record
   the decision.
2. Implement the resolved rejection in `Framework.Validate()` (`kernel/config/config.go`), following
   the existing code shape at either lines 192-200 (unconditional pattern) or 254-263 (prod-gated
   pattern), whichever is chosen.
3. Extend `kernel/config/unsafe_config_matrix_test.go`'s existing table (which already covers
   `HTTP.ReadHeaderTimeout`, `HTTP.RequestTimeout`, `HTTP.MaxBodyBytes` at lines 106, 111, 116, 121)
   with equivalent cases for the 4 new keys, consistent with the resolved policy.
4. If the resolved policy is prod-profile-only (per the task brief's original framing), also add a
   test proving a *non-prod* profile with a zero-value new-key config does **not** fail Validate (to
   distinguish the two policies unambiguously in the test suite).

### Expected files or components affected

- `kernel/config/config.go` (`Framework.Validate()`).
- `kernel/config/unsafe_config_matrix_test.go` (or an equivalent existing config test file).

### Expected output

A prod-profile (or, per the resolved policy, any-profile) `Framework` config carrying an explicit
zero value for any of the 4 new timeout keys fails `Validate()` with a descriptive error naming the
offending key; unset config (falling through to the safe non-zero defaults) does not fail.

### Required artifacts

None beyond the code diff itself (this task does not introduce a new artifact type beyond what T001
already registers for the config schema).

### Required evidence

Prod-profile (or resolved-policy) zero-timeout rejection test output. See
`../../evidence/index.md`.

### Related acceptance criteria

AC-W01-E03-S001-03.

### Completion criteria

The resolved validation policy is implemented and proven by a passing test; the non-rejection case
(unset config falling through to defaults) is also proven not to fail, directly addressing
RISK-W01-003.

### Verification method

Run the extended `unsafe_config_matrix_test.go` (or equivalent) table locally and in CI.

### Risks

RISK-W01-003 (interaction with an unset-config deployment) — directly mitigated by the explicit
non-rejection test case for unset/default config.

### Rollback or recovery considerations

Revert the validation-logic commit; no running-system state implicated.

## Implementation Record

Implemented 2026-07-13 by W01Http at SHA 0a31186cada5c275a588c74081cf977adf346e61 (working tree; conductor owns the wave commit).

- Plan Q1 resolved: **prod-profile-only** rejection (matches task brief, AC-03, and the
  SSRF-disable precedent). Implemented in `Framework.Validate()`'s existing `IsProd()` block:
  `<= 0` rejection for read/write/idle timeouts with descriptive per-key messages.
- `unsafe_config_matrix_test.go`: 3 new prod-gated rows (fail-first captured: "knob is NOT gated
  in prod" pre-implementation); `TestConnectionTimeoutZeroToleratedOutsideProd` proves non-prod
  zero passes (policy disambiguation per detailed-work step 4); prod baseline with unset config
  (defaults) still validates (RISK-W01-003).

### Commits / pull requests

None yet — conductor owns the wave commit; working-tree diff recorded in the story's
implementation.md.

## Verification Record

| Acceptance criterion | Verification method | Environment | Result | Evidence | Reviewer |
|---|---|---|---|---|---|
| AC-W01-E03-S001-03 | Extended matrix + non-prod tolerance test, fail-first pair | Local | PASS | EV-W01-E03-S001-003 | pending |

Execution date 2026-07-13; revision 0a31186cada5c275a588c74081cf977adf346e61; environment local darwin/arm64 (go1.26.5);
reviewer pending (W01 wave review gate). Findings: none open. Retest: not required.
Final conclusion: task complete, ACs verified.

## Deviations Record

None — see the story-level deviations.md.
