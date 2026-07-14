---
id: W03-E02-S001-T001
type: task
title: Fingerprint-scope confirmation (SEC-06 T1)
status: done
parent_story: W03-E02-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on: []
acceptance_criteria:
  - AC-W03-E02-S001-01
artifacts:
  - ART-W03-E02-S001-001
evidence:
  - EV-W03-E02-S001-001
---

# W03-E02-S001-T001 — Fingerprint-scope confirmation (SEC-06 T1)

## Task Definition

### Task objective

Confirm, or extend, `SharedFingerprint()`'s scope to cover the outbound allowlist (`AllowedHosts`/
`AllowedCIDRs`), proven by a fingerprint-diff regression test.

### Parent story

W03-E02-S001 — Outbound-security escape-hatch governance.

### Owner

unassigned

### Status

done

### Dependencies

None.

### Detailed work

1. Confirm `SharedFingerprint()`'s current field coverage at this task's actual start commit — read
   the implementation directly, not PLAN's own hedge ("likely already covers these fields
   structurally").
2. Write a fingerprint-diff test that mutates the allowlist (`AllowedHosts`/`AllowedCIDRs`) and
   asserts the fingerprint's output changes.
3. If the test fails (fingerprint does not change), extend `SharedFingerprint()`'s scope to cover the
   allowlist fields, then re-run the test.

### Expected files or components affected

The config layer's `SharedFingerprint()` implementation (exact file TBD at implementation time).

### Expected output

A fingerprint-diff regression test proving `SharedFingerprint()`'s scope covers the outbound
allowlist — either because it already did, or because this task extended it.

### Required artifacts

ART-W03-E02-S001-001 (`SharedFingerprint()` scope confirmation/extension and its regression test).

### Required evidence

EV-W03-E02-S001-001 (fingerprint-diff regression test output).

### Related acceptance criteria

AC-W03-E02-S001-01.

### Completion criteria

The fingerprint-diff test passes, proving a mutation to `AllowedHosts`/`AllowedCIDRs` changes
`SharedFingerprint()`'s output.

### Verification method

Direct test execution, logged output retained as evidence.

### Risks

Low, per PLAN's own T1 risk note: "likely already correct, may close as 'add regression test only'."

### Rollback or recovery considerations

If an extension to `SharedFingerprint()`'s scope is required, it is a low-risk additive change to an
existing fingerprint function — revertible independently of this story's other tasks.

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
