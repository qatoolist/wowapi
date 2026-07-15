---
id: W03-E02-S001-T005
type: task
title: No-tenant-controlled-allowlist fitness check (SEC-06 T5)
status: done
parent_story: W03-E02-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on:
  - W03-E02-S001-T001
  - W03-E02-S001-T002
  - W03-E02-S001-T003
  - W03-E02-S001-T004
acceptance_criteria:
  - AC-W03-E02-S001-05
artifacts:
  - ART-W03-E02-S001-005
evidence:
  - EV-W03-E02-S001-005
---

# W03-E02-S001-T005 — No-tenant-controlled-allowlist fitness check (SEC-06 T5)

## Task Definition

### Task objective

Codify "never tenant/user-controlled data populates allowlists/JWKS clients" as a static fitness
check — a mechanically enforced invariant, not merely a documented convention.

### Parent story

W03-E02-S001 — Outbound-security escape-hatch governance.

### Owner

unassigned

### Status

done

### Dependencies

W03-E02-S001-T001, W03-E02-S001-T002, W03-E02-S001-T003, W03-E02-S001-T004 — PLAN's own Depends-on
column for T5: "T1-T4." Sequenced last per `plan.md`'s own rationale so the fitness check covers
T4's newly-added JWKS-client governance surface too.

### Detailed work

1. Decide the fitness-check mechanism: a custom linter analyzer (golangci-lint-compatible) versus a
   dedicated CI-time test walking construction call sites — per `plan.md`'s own framing, a dedicated
   test is likely simpler for this single, narrow invariant, but this is a judgment call for
   implementation time.
2. Implement the fitness check: assert that allowlist/JWKS-client construction call sites never read
   from a request-scoped or tenant-scoped context value.
3. Write the fitness-check test proving the assertion actually fires against a deliberately
   introduced violation (a fail-first-style proof that the check is not vacuously true).

### Expected files or components affected

A new fitness-check test file (exact path TBD), or `.golangci.yml` if implemented as a custom
analyzer.

### Expected output

A static fitness check that fails if allowlist/JWKS-client construction reads request- or
tenant-scoped data; proven to actually fire against a deliberately introduced violation.

### Required artifacts

ART-W03-E02-S001-005 (the no-tenant-controlled-allowlist fitness check).

### Required evidence

EV-W03-E02-S001-005 (fitness-check test output).

### Related acceptance criteria

AC-W03-E02-S001-05.

### Completion criteria

The fitness check passes against the current, correct codebase, and is proven to fail against a
deliberately introduced violation (fail-first-style proof it is not vacuously true).

### Verification method

Direct test/analyzer execution, including a deliberately-introduced-violation proof run, logged
output retained as evidence.

### Risks

Low, per PLAN's own T5 risk note: "codifying an already-true invariant."

### Rollback or recovery considerations

Additive check — independently revertible without affecting runtime behavior of the allowlist/JWKS
client construction itself.

## Implementation Record

Implementation details are recorded in the story-level `implementation.md`.

## Verification Record

Verification details are recorded in the story-level `verification.md`; evidence is in `evidence/index.md`.

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
