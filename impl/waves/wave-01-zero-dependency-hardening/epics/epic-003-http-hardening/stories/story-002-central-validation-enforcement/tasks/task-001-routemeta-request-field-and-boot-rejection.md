---
id: W01-E03-S002-T001
type: task
title: RouteMeta.Request field + boot-time rejection + waiver field, forward-compatible with AR-04 T5
status: done
parent_story: W01-E03-S002
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on: []
acceptance_criteria:
  - AC-W01-E03-S002-01
  - AC-W01-E03-S002-02
  - AC-W01-E03-S002-04
artifacts: []
evidence: []
---

# W01-E03-S002-T001 — RouteMeta.Request field + boot-time rejection + waiver field, forward-compatible with AR-04 T5

## Task Definition

### Task objective

Add a request-contract-declaration mechanism to `RouteMeta` (resolving `plan.md`'s Candidate A vs.
Candidate B question), a boot-time check that fails a POST/PUT/PATCH route with no declared contract
and no waiver, and a waiver field for genuinely body-less mutations — designed additively/forward-
compatible with AR-04 T5's future waiver mechanism (W05 scope, not yet built).

### Parent story

W01-E03-S002 — Central validation enforcement.

### Owner

Unassigned.

### Status

todo.

### Dependencies

Design review must resolve `plan.md` unresolved question 1 (Candidate A vs. Candidate B) before this
task's implementation proceeds — this is a design-review gate, not a task/story dependency in the
`depends_on` sense.

### Detailed work

1. Resolve Candidate A vs. Candidate B (see `plan.md` "Proposed architecture") via design review,
   consulting the `codebase-design` skill's seam-placement guidance per this repository's routing
   (`CLAUDE.md`).
2. Resolve `plan.md` unresolved question 4 (structural fact: `RouteMeta.validate()` has no access to
   the HTTP method today — `router.go`'s `RouteMeta.validate()` signature is `func (m RouteMeta)
   validate() error`, no parameters) — decide whether to extend that signature, pass the method in,
   or relocate the new check to `Router.Handle` (which already has `method` in scope).
3. Add the resolved contract-declaration field to `RouteMeta` (`kernel/httpx/router.go`, near lines
   18-33).
4. Add the waiver field, keeping its shape minimal (per `plan.md`'s Assumption/recommendation) unless
   design review determines otherwise; add a doc comment explicitly noting its forward-compatibility
   intent with AR-04 T5 (referencing `requirement-inventory.md` row AR-04) so a future AR-04 T5
   implementer finds this note rather than rediscovering the coordination need.
5. Add the boot-time check, wired through the existing `r.errs` accumulation pattern
   (`Router.Handle`, `router.go:73-89`), gated by the profile flag (location resolved per `plan.md`
   unresolved question 2).
6. Write the fail-first fixture-route test: prove a POST route with no declared contract boots
   successfully on the pre-fix codebase (capture this run as evidence before proceeding), then prove
   it fails boot post-fix with the flag enabled.
7. Write the waiver-exemption fixture-route test: prove a body-less mutating route using the waiver
   boots successfully with the flag enabled.

### Expected files or components affected

- `kernel/httpx/router.go` (`RouteMeta`, `validate()`, possibly `Router.Handle`).
- `kernel/config` or a product-level config location (the profile flag — location per resolved
  question 2).
- A new or extended `kernel/httpx` test file for the fixture-route tests.

### Expected output

`RouteMeta` carries the resolved contract-declaration field and waiver field; a boot-time check
(flag-gated) rejects an undeclared-contract mutating route; both fail-first and waiver-exemption
tests pass; the design-review resolution (Candidate A/B, flag location, waiver shape, check
placement) is recorded in this task's Implementation Record.

### Required artifacts

`RouteMeta.Request` contract type (or resolved-shape equivalent). See `../../artifacts/index.md`.

### Required evidence

Boot-rejection fail-first test pair (pre-fix boots, post-fix fails); waiver-exemption boot-success
test. See `../../evidence/index.md`.

### Related acceptance criteria

AC-W01-E03-S002-01, AC-W01-E03-S002-02, AC-W01-E03-S002-04.

### Completion criteria

All four `plan.md` unresolved questions relevant to this task (1, 2, 3, 4) are resolved and recorded;
the fail-first pair and waiver-exemption test both pass; the boot-time check correctly accumulates
its error into the existing `Router.Err()` path rather than introducing a separate failure mechanism.

### Verification method

Run the fixture-route tests locally and in CI, both with the flag off (existing behavior unchanged)
and with the flag on (new rejection active).

### Risks

RISK-W01-002 (boot-time rejection breaking an existing route) — directly mitigated by the flag
defaulting off, proven by the fail-first pair's "boots pre-fix" half remaining true even post-fix
when the flag is off (a test case this task should also cover, even though it is not separately
numbered as an AC).

### Rollback or recovery considerations

Turn the profile flag back off if the check proves defective; code-level revert only if the flag
mechanism itself is broken.

## Implementation Record

Implemented 2026-07-13 by W01Http at SHA 0a31186cada5c275a588c74081cf977adf346e61 (working tree; conductor owns the wave commit).

- Plan Q1 resolved: **Candidate A** (type-token prototype `Request any`) — AR-03 (W05) needs the
  concrete declared type reachable from the route table for projections; drift risk accepted and
  documented on the field. Q4 resolved: check lives in `Router.Handle` via
  `checkRequestContract(method, meta)` (RouteMeta.validate() has no method access). Q2 resolved:
  flag is `config.Security.EnforceRouteContracts` (framework-level), wired in `app/boot.go`
  BEFORE module registration. Q3 resolved: minimal bool waiver `NoRequestBody`, doc-commented as
  AR-04 T5 forward-compatible (references requirement-inventory row AR-04).
- Checks: POST/PUT/PATCH only; Request+NoRequestBody contradiction rejected unconditionally;
  missing both rejected only under RequireRequestContracts; errors accumulate into r.errs.
- Tests: `kernel/httpx/route_contract_test.go` (compat-default, 3-verb rejection, declared pass,
  waiver, contradiction, non-mutating exemption) + `TestEnforceRouteContractsDefaultsOff`
  (kernel/config). Three-stage fail-first captured (pristine-boots / stub-red / green).

### Commits / pull requests

None yet — conductor owns the wave commit; working-tree diff recorded in the story's
implementation.md.

## Verification Record

| Acceptance criterion | Verification method | Environment | Result | Evidence | Reviewer |
|---|---|---|---|---|---|
| AC-W01-E03-S002-01 | Fixture boots at pristine HEAD; still boots flag-off post-fix | Local | PASS | EV-W01-E03-S002-001 | pending |
| AC-W01-E03-S002-02 | Rejection test red (stub) → green (implemented) | Local | PASS | EV-W01-E03-S002-001 | pending |
| AC-W01-E03-S002-04 | Waiver-exemption boot success + contradiction guard | Local | PASS | EV-W01-E03-S002-003 | pending |

Execution date 2026-07-13; revision 0a31186cada5c275a588c74081cf977adf346e61; environment local darwin/arm64 (go1.26.5);
reviewer pending (W01 wave review gate). Findings: none open. Retest: not required.
Final conclusion: task complete, ACs verified.

## Deviations Record

None — see the story-level deviations.md.
