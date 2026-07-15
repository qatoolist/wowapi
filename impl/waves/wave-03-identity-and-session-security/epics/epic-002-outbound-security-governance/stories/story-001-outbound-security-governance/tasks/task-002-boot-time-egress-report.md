---
id: W03-E02-S001-T002
type: task
title: Boot-time egress-exception report (SEC-06 T2)
status: done
parent_story: W03-E02-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on: []
acceptance_criteria:
  - AC-W03-E02-S001-02
artifacts:
  - ART-W03-E02-S001-002
evidence:
  - EV-W03-E02-S001-002
---

# W03-E02-S001-T002 — Boot-time egress-exception report (SEC-06 T2)

## Task Definition

### Task objective

Implement a boot-time startup report enumerating every enabled egress exception (`AllowedHosts`/
`AllowedCIDRs` and any other configured escape hatch), with no credentials exposed in the output.

### Parent story

W03-E02-S001 — Outbound-security escape-hatch governance.

### Owner

unassigned

### Status

done

### Dependencies

None.

### Detailed work

1. Identify the existing readiness/boot-reporting layer this report should extend (exact file TBD at
   implementation time).
2. Implement the report: enumerate `AllowedHosts`/`AllowedCIDRs` and any other configured egress
   exception, formatted for readiness/log output.
3. Explicitly review the output format to confirm no credential or secret value is included —
   document this review step's outcome as part of the task's evidence.

### Expected files or components affected

The readiness/boot-reporting layer (exact file TBD).

### Expected output

A boot-time report enumerating every enabled egress exception, confirmed credential-free.

### Required artifacts

ART-W03-E02-S001-002 (boot-time egress-exception report implementation).

### Required evidence

EV-W03-E02-S001-002 (report-output sample, confirmed credential-free).

### Related acceptance criteria

AC-W03-E02-S001-02.

### Completion criteria

The report enumerates every configured egress exception; a review confirms no credential or secret
value appears in the output.

### Verification method

Direct report-output test against a fixture configuration with multiple exceptions enabled, logged
output retained as evidence.

### Risks

Low, per PLAN's own T2 risk note.

### Rollback or recovery considerations

Additive boot-time reporting change — independently revertible.

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
