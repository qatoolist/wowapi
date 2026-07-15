---
id: W03-E02-S001-T004
type: task
title: JWKS-client governance gate, D-07 enactment (SEC-06 T4)
status: done
parent_story: W03-E02-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on:
  - W03-E02-S001-T001
  - W03-E02-S001-T002
  - W03-E02-S001-T003
acceptance_criteria:
  - AC-W03-E02-S001-04
artifacts:
  - ART-W03-E02-S001-004
evidence:
  - EV-W03-E02-S001-004
---

# W03-E02-S001-T004 — JWKS-client governance gate, D-07 enactment (SEC-06 T4)

## Task Definition

### Task objective

Extend equivalent governance to the JWKS `Client` injection path per D-07 (`ADR-W00-E02-S003-007`): a
`prod`-profile boot with a custom JWKS client injected and no declared trusted-issuer allowlist fails
readiness.

### Parent story

W03-E02-S001 — Outbound-security escape-hatch governance.

### Owner

unassigned

### Status

done

### Dependencies

W03-E02-S001-T001, W03-E02-S001-T002, W03-E02-S001-T003 — PLAN's own Depends-on column for T4:
"T1-T3." Also depends on D-07 (`ADR-W00-E02-S003-007`) being confirmed ratified before this task's
implementation begins, per `story.md` "Assumptions."

### Detailed work

1. Confirm D-07 (`ADR-W00-E02-S003-007`) is ratified.
2. Design and implement the trusted-issuer config field per D-07: a declared, fingerprinted `config`
   field on `JWKSConfig` (or the appropriate config struct), exact shape finalized against D-07's
   ratified text and the existing `JWKSConfig` struct's conventions.
3. Implement `prod`-profile boot validation: a custom JWKS client injected (`JWKSConfig.Client !=
   nil`) with no declared trusted-issuer allowlist causes readiness to fail closed, not merely a
   warning log — consistent with the framework's established fail-closed config-validation pattern
   (per `epic.md`'s citation of a comparable SEC-04 T6 precedent in `config.go`).
4. Write the negative-fixture test: boot with a `prod` profile, a custom JWKS client injected, and no
   declared trusted-issuer allowlist; assert readiness fails.
5. Write a positive-path test: boot with a `prod` profile, a custom JWKS client injected, and a
   declared trusted-issuer allowlist; assert readiness succeeds.

### Expected files or components affected

`kernel/auth/jwks.go:59` (`JWKSConfig.Client`); the config layer (new trusted-issuer field); the
readiness/boot-validation layer.

### Expected output

A `prod`-profile boot with a custom JWKS client and no declared trusted-issuer allowlist fails
readiness; the same boot with a declared trusted-issuer allowlist succeeds.

### Required artifacts

ART-W03-E02-S001-004 (JWKS trusted-issuer config-gate implementation, D-07 enactment).

### Required evidence

EV-W03-E02-S001-004 (JWKS-client-governance negative-fixture test output).

### Related acceptance criteria

AC-W03-E02-S001-04.

### Completion criteria

The negative-fixture test proves readiness fails under the ungoverned-injection case; a positive-path
test proves readiness succeeds once the trusted-issuer allowlist is declared.

### Verification method

Direct negative-fixture and positive-path test execution, logged output retained as evidence.

### Risks

Per PLAN's own T4 risk note, this is the "Highest-risk task" in SEC-06 — an "open design decision,
not yet made" at PLAN-authoring time, now resolved by D-07's ratification. **This is explicitly
"Breaking only for T4, only if wowsociety currently injects a custom JWKS client with no declaration
path (unconfirmed)"** — see `story.md` "Compatibility considerations" for the full, honestly-recorded
evidence-gap framing. This task does not soften the fail-closed gate to avoid the hypothetical break.

### Rollback or recovery considerations

If the fail-closed gate is found post-rollout to break a legitimate wowsociety JWKS-client-injection
usage, coordinate with wowsociety to declare the trusted-issuer allowlist rather than reverting the
gate itself, consistent with `story.md`'s framing that the gate is correct behavior independent of
whether it is triggered today.

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
