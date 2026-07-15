---
id: W05-E05-S002
type: story
title: Kernel package-count and wowsociety identity-suite verification
status: planned
wave: W05
epic: W05-E05
owner: unassigned
reviewer: unassigned
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-12
source_requirements:
  - FBL-01
depends_on:
  - W05-E05-S001
blocks: []
acceptance_criteria:
  - AC-W05-E05-S002-01
  - AC-W05-E05-S002-02
artifacts: []
evidence: []
decisions: []
risks:
  - RISK-W05-003
  - RISK-W05-004
---

# W05-E05-S002 — Kernel package-count and wowsociety identity-suite verification

## Story ID

W05-E05-S002

## Title

Kernel package-count and wowsociety identity-suite verification

## Objective

Verify MATRIX CS-01's own acceptance bar: `go list ./kernel/... | wc -l` at or below the
target-list count; depguard and boundaries lint green; wowsociety's identity/authz suite green on
`foundation/mfa` or the shim — coordinating with wowsociety per PROD-02 without performing
wowsociety's own code migration.

## Value to the framework

This story is FBL-01's own acceptance proof — the point at which "the framework re-homed nine
packages" becomes a verified, evidenced fact rather than an assertion. Given FBL-01's own status as
"the largest single architectural correction" that "must precede v1 stabilisation," this
verification is the gate the whole wave's kernel-layering work rests on.

## Problem statement

MATRIX CS-01's own acceptance bar: "`go list ./kernel/... | wc -l` ≤ target-list count; depguard +
boundaries lint green; wowsociety identity suite green on `foundation/mfa` (or on the shim during
the grace window)." PLAN's own detailed task register (§O): "Tests: boundary lint asserting a kernel
import allowlist that rejects the re-homed paths; `wowsociety` build + identity suite green on new
mfa path." REVIEW §P's own framing: "Sequence the mfa move deliberately with an identity-module
migration task + full re-run of wowsociety's identity/authz test suite."

## Source requirements

FBL-01 (acceptance bar).

## Current-state assessment

Prior to S001's own work landing, no `foundation/` tree exists, so this story's own verification
cannot meaningfully run until S001 has completed. This story's own re-confirmation step (once S001
has landed) is to actually run the `go list` count, the lint suite, and wowsociety's identity/authz
suite against S001's actual output, not to assume its success from S001's own self-reported
completion.

## Desired state

`go list ./kernel/... | wc -l` returns a count at or below the target-list count established by
S001-T004's own final retained kernel-package enumeration. Depguard and boundaries lint are both
green. wowsociety's build and full identity/authz test suite run green against the new
`foundation/mfa` path or the `kernel/mfa` shim.

## Scope

- Running and recording `go list ./kernel/... | wc -l` against the post-move state.
- Running and recording the depguard and boundaries-lint suites' green status.
- Coordinating with wowsociety (PROD-02) to run its build and full identity/authz test suite against
  the shim (or, if wowsociety has already migrated by this point, against `foundation/mfa`
  directly).
- Recording this story's own verification results as the epic's own acceptance evidence.

## Out of scope

- **Performing wowsociety's own code migration off `kernel/mfa`** — PROD-02, product-level, out of
  framework scope. This story runs wowsociety's suite against the framework's shim; it does not edit
  wowsociety's own source.
- **Any fix to S001's own implementation** — if this story's verification finds a failure, the fix
  is recorded as a finding requiring S001's own follow-up (or a deviation), not silently patched
  within this verification-only story.

## Assumptions

- wowsociety's repository is accessible for this story's own verification run (build + identity/authz
  suite) — this is a cross-repo verification step, consistent with this wave's own broader posture on
  wowsociety-facing acceptance criteria (e.g. AR-01-S004's own legacy-adapter compatibility proof).
  If wowsociety's repository state at this story's own execution time differs materially from what
  REVIEW §J/§O/§P describe, that difference is recorded as a finding, not silently reconciled.

## Dependencies

Depends on W05-E05-S001 (the move and shim this story verifies). No dependency on any other W05
epic beyond what S001 itself already carries (W05-E01, W05-E02).

## Affected packages or components

None directly modified by this story — a verification-only story. `kernel/`, `foundation/`, and
wowsociety's own identity module are all read/tested, not changed.

## Compatibility considerations

This story's own wowsociety-suite-green verification is itself the compatibility proof for the
`kernel/mfa` shim.

## Security considerations

Running wowsociety's full identity/authz suite (not merely an mfa-scoped subset) against the shim is
the required security-adjacent verification, per REVIEW §P's own explicit instruction: "full re-run
of wowsociety's identity/authz test suite."

## Performance considerations

None material.

## Observability considerations

None beyond this story's own evidence-recording requirements.

## Migration considerations

None — this story performs no migration of its own.

## Documentation requirements

Document the verification results: the actual `go list` count, the lint status, and wowsociety's
suite results, with commit SHAs for both repositories.

## Acceptance criteria

- **AC-W05-E05-S002-01**: `go list ./kernel/... | wc -l` is at or below the target-list count
  established by S001-T004's final enumeration; depguard and boundaries lint are both green.
- **AC-W05-E05-S002-02**: wowsociety's build and full identity/authz test suite run green against the
  `foundation/mfa` path or the `kernel/mfa` shim, with the actual commit SHA of both repositories
  recorded.

## Required artifacts

- The verification results record (documentation).
See `artifacts/index.md`.

## Required evidence

- The `go list` count output.
- The depguard/boundaries-lint green-run output.
- wowsociety's build and full identity/authz suite run output, with both repositories' commit SHAs.
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`: `story.md`
and `plan.md` complete, acceptance criteria numbered and measurable, dependency on S001 recorded,
wowsociety repository accessibility for this story's own verification run confirmed, owner/reviewer
assignment pending.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; both acceptance
criteria verified with evidence in `evidence/index.md`; `closure.md` completed; independent review
passed per mandate §14, specifically confirming wowsociety's suite was genuinely run in full (not a
narrowed subset), given REVIEW §P's own explicit instruction.

## Risks

RISK-W05-003 (schedule dependency on S001 landing) and RISK-W05-004 (the `kernel/mfa` shim's
coordination risk) — see epic-level `risks.md` for full detail and mitigation/contingency.

## Residual-risk expectations

Residual risk is expected to be low once wowsociety's full identity/authz suite is confirmed green
and independently re-checked by this story's own review task.

## Plan

See `plan.md`.
