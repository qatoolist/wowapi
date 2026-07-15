---
id: W05-E03-S002-T004
type: task
title: Shared no-op-adapter waiver mechanism
status: todo
parent_story: W05-E03-S002
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on:
  - W05-E03-S002-T001
  - W05-E03-S002-T002
  - W05-E03-S002-T003
acceptance_criteria:
  - AC-W05-E03-S002-03
artifacts:
  - ART-W05-E03-S002-004
evidence:
  - EV-W05-E03-S002-003
---

# W05-E03-S002-T004 — Shared no-op-adapter waiver mechanism

## Task Definition

### Task objective

Build the explicit optional-capability declaration and waiver mechanism: a `prod` profile with a
required-but-no-op/missing adapter and no waiver fails readiness by name; `local` with the same
configuration succeeds; a policy-approved waiver suppresses the failure with an audit record. This
mechanism is built once, as the shared primitive later consumed by SEC-06 and DX-07 — per
`impl/analysis/wave-allocation-detail.md`'s own note.

### Parent story

W05-E03-S002 — Boot-time strictness and the shared no-op-adapter waiver mechanism.

### Owner

unassigned

### Status

todo

### Dependencies

W05-E03-S002-T001, W05-E03-S002-T002, W05-E03-S002-T003 (PLAN T5's own dependency row: "AR-01, AR-02,
T1-T4"); depends on W05-E01 and W05-E02 at story scope.

### Detailed work

1. Design the explicit optional-capability declaration mechanism.
2. Design the waiver mechanism as a reusable, cross-consumer primitive, with explicit consideration
   of SEC-06's and DX-07's anticipated consumption needs.
3. Implement the readiness-gating logic: `prod` + required-but-no-op/missing adapter + no waiver →
   named readiness failure; `local` + same configuration → succeeds.
4. Implement waiver suppression with a required audit record.
5. Write `AR-04/prod_noop_adapter_readiness_test.go`: the integration matrix (profile × waiver ×
   adapter-real/no-op).
6. Document the mechanism as a shared primitive, with a clear reference for future SEC-06/DX-07
   implementers.

### Expected files or components affected

A new waiver-mechanism package (exact location TBD).

### Expected output

A working waiver mechanism proven by the named integration-matrix test, designed for reuse by
SEC-06 and DX-07.

### Required artifacts

ART-W05-E03-S002-004.

### Required evidence

EV-W05-E03-S002-003.

### Related acceptance criteria

AC-W05-E03-S002-03.

### Completion criteria

The integration matrix confirms all three named scenarios: prod+no-op+no-waiver fails named; local
succeeds; waiver present suppresses and audits.

### Verification method

Direct execution of `AR-04/prod_noop_adapter_readiness_test.go`.

### Risks

Medium, per PLAN T5's own risk column — "shares scope with SEC-06 and DX-07's readiness closure
contracts; build the waiver mechanism once."

### Rollback or recovery considerations

If the mechanism's shape proves unsuitable for SEC-06/DX-07's later consumption needs, escalate for
redesign — do not ship a narrow mechanism coupled only to this task's own scenario, since that would
force divergent waiver implementations for SEC-06/DX-07 later.

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

*Not applicable, pending implementation-time confirmation of whether waivers persist to the
database.*

### Security changes

*Not yet implemented — this mechanism is a security/compliance control; recorded here once
implemented.*

### Observability changes

*Not yet implemented — the waiver audit record; recorded here once implemented.*

### Tests added or modified

*Not yet implemented.*

### Commits

*Not yet implemented.*

### Pull requests

*Not yet implemented.*

### Implementation dates

*Not yet implemented.*

### Technical debt introduced

*None anticipated.*

### Known limitations

*Not yet implemented.*

### Follow-up items

*Not yet implemented.*

### Relationship to the approved plan

*Not yet implemented.*

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W05-E03-S002-03 | Run `AR-04/prod_noop_adapter_readiness_test.go` | Local dev or CI, Go toolchain | prod+no-op+no-waiver fails named; local succeeds; waiver suppresses + audits | integration-matrix test report | unassigned |

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
