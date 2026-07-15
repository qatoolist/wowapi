---
id: W05-E05-S001-T002
type: task
title: kernel/mfa re-home and forwarding shim
status: todo
parent_story: W05-E05-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on: []
acceptance_criteria:
  - AC-W05-E05-S001-01
  - AC-W05-E05-S001-02
artifacts:
  - ART-W05-E05-S001-002
evidence:
  - EV-W05-E05-S001-002
---

# W05-E05-S001-T002 — kernel/mfa re-home and forwarding shim

## Task Definition

### Task objective

`git mv` `kernel/mfa` to `foundation/mfa`, and leave a deprecated forwarding shim (type aliases +
var forwarding) at `kernel/mfa` for one minor version, so wowsociety migrates on its own schedule,
proven by a behavioral-equivalence test.

### Parent story

W05-E05-S001 — Foundation tree, package moves, and mfa forwarding shim.

### Owner

unassigned

### Status

todo

### Dependencies

None within this story (parallel-safe with T001 — disjoint package); depends on W05-E01 and W05-E02
at story scope.

### Detailed work

1. Re-confirm, via a repo-wide grep search (wowapi-internal and wowsociety), that `kernel/mfa` is
   imported by exactly the 5 identity-module files in wowsociety's `internal/modules/identity/`
   REVIEW §J/§O/§P describe, at this task's actual start commit.
2. `git mv` `kernel/mfa` to `foundation/mfa`, preserving history.
3. Implement the deprecated forwarding shim at `kernel/mfa`: type aliases + var forwarding onto
   `foundation/mfa`'s public surface.
4. Write a behavioral-equivalence test: calls through the shim produce identical results to direct
   `foundation/mfa` calls, for every exported symbol the shim forwards.
5. Document the shim's deprecation timeline (one minor version) and its exact forwarding mechanics.

### Expected files or components affected

`foundation/mfa` (new location); `kernel/mfa` (shim, retained).

### Expected output

`kernel/mfa` re-homed to `foundation/mfa`; a working, behaviorally-equivalent shim at `kernel/mfa`.

### Required artifacts

ART-W05-E05-S001-002.

### Required evidence

EV-W05-E05-S001-002.

### Related acceptance criteria

AC-W05-E05-S001-01, AC-W05-E05-S001-02.

### Completion criteria

The behavioral-equivalence test passes for every exported symbol the shim forwards.

### Verification method

Direct execution of the behavioral-equivalence test.

### Risks

Medium-high — REVIEW §P's own framing: "a real, security-sensitive import-path + call-site
migration ... not a mechanical zero-cost change." See RISK-W05-004 in epic-level `risks.md`.

### Rollback or recovery considerations

If the equivalence test reveals any behavioral divergence between the shim and direct
`foundation/mfa` calls, fix before proceeding — a divergent shim is an authentication-availability
risk, not a cosmetic bug.

## Implementation Record

*Not yet implemented.*

### What was actually implemented

*Not yet implemented.*

### Components changed

*Not yet implemented.*

### Files changed

*Not yet implemented.*

### Interfaces introduced or changed

*Not yet implemented.*

### Configuration changes

*Not yet implemented.*

### Schema or migration changes

*Not applicable.*

### Security changes

*Not yet implemented — the shim's behavioral-equivalence guarantee for auth-critical code; recorded
here once implemented.*

### Observability changes

*Not applicable.*

### Tests added or modified

*Not yet implemented.*

### Commits

*Not yet implemented.*

### Pull requests

*Not yet implemented.*

### Implementation dates

*Not yet implemented.*

### Technical debt introduced

*The shim itself is a planned, time-bounded technical-debt item (one minor version) — recorded here
once implemented, not treated as unplanned debt.*

### Known limitations

*Not yet implemented.*

### Follow-up items

*Not yet implemented.*

### Relationship to the approved plan

*Not yet implemented.*

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W05-E05-S001-01 | Run a full repository build post-move | Local dev or CI, Go toolchain | Build succeeds | build-output report | unassigned |
| AC-W05-E05-S001-02 | Run the shim behavioral-equivalence test | Local dev or CI, Go toolchain | Shim calls identical to direct foundation/mfa calls | equivalence-test report | unassigned |

### Actual result

*Not yet executed.*

### Pass or fail

*Not yet executed.*

### Evidence identifier

*Not yet executed.*

### Execution date

*Not yet executed.*

### Commit or revision

*Not yet executed.*

### Environment

*Not yet executed.*

### Reviewer

*Not yet executed.*

### Findings

*Not yet executed.*

### Retest status

*Not yet executed.*

### Final conclusion

*Not yet executed.*

## Deviations Record

*No deviations recorded yet.*

### Deviation ID

*Not applicable.*

### Approved plan

*Not applicable.*

### Actual implementation

*Not applicable.*

### Reason

*Not applicable.*

### Impact

*Not applicable.*

### Risks

*Not applicable.*

### Approval

*Not applicable.*

### Compensating controls

*Not applicable.*

### Follow-up work

*Not applicable.*
