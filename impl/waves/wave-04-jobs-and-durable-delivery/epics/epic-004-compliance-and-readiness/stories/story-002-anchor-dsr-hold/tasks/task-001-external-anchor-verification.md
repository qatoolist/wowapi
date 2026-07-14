---
id: W04-E04-S002-T001
type: task
title: External anchor verification
status: done
parent_story: W04-E04-S002
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on: []
acceptance_criteria:
  - AC-W04-E04-S002-01
artifacts:
  - ART-W04-E04-S002-001
  - ART-W04-E04-S002-006
evidence:
  - EV-W04-E04-S002-001
---

# W04-E04-S002-T001 — External anchor verification

## Task Definition

### Task objective

Build a mechanism that periodically anchors the audit chain's head hash externally, and implement
detection logic proving that tampering with the local chain is detectable via the external anchor
even if the local `head_hash` were compromised.

### Parent story

W04-E04-S002 — External anchoring, DSR export artifact, central legal-hold, and explicit per-class
status.

### Owner

unassigned

### Status

done

### Dependencies

None.

### Detailed work

1. Draft external-anchoring mechanism options (e.g. a public timestamping service, a separate
   append-only log, a third-party notarization service) with trade-offs; select one and document the
   rationale, per PLAN DATA-08 W6-T2's own risk note ("Genuinely new subsystem — vendor/design
   decision needed").
2. Implement the anchoring mechanism: periodically publish the chain head externally.
3. Implement detection logic that cross-checks the local chain against the external anchor.
4. Write the anchor-then-tamper detection test: anchor the chain, tamper with a local row, confirm
   detection via the anchor.
5. Document the anchoring mechanism and its verification procedure.

### Expected files or components affected

A new file or extension within `kernel/audit` implementing the external anchoring mechanism (exact
path TBD, dependent on the vendor/protocol decision); a new test file for the anchor-then-tamper
test.

### Expected output

A working external anchoring mechanism; a passing anchor-then-tamper detection test; documentation
of the mechanism.

### Required artifacts

ART-W04-E04-S002-001 (external anchor mechanism), ART-W04-E04-S002-006 (documentation, shared with
T002/T003/T004).

### Required evidence

EV-W04-E04-S002-001 (anchor-tamper-detection report).

### Related acceptance criteria

AC-W04-E04-S002-01.

### Completion criteria

The anchoring mechanism periodically publishes the chain head externally; the anchor-then-tamper test
confirms tampering is detectable via the anchor even where local `head_hash` alone would not reveal
it.

### Verification method

Direct execution of the anchor-then-tamper detection test.

### Risks

The vendor/design decision itself (PLAN W6-T2's own risk note) — an under-specified anchoring
mechanism could fail to actually provide independent tamper-evidence if the external target is not
genuinely outside the attacker's compromise surface.

### Rollback or recovery considerations

If the chosen anchoring vendor/protocol proves unworkable post-implementation, the anchoring
mechanism can be replaced without affecting `chainHash` or `Verify`'s core logic — it is an additive
layer, not a replacement of the existing hash-chain mechanism.

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

*Not applicable unless the chosen anchoring mechanism requires a local record of anchor events — TBD
at implementation time.*

### Security changes

*Not yet implemented.*

### Observability changes

*Not yet implemented.*

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
| AC-W04-E04-S002-01 | Run anchor-then-tamper detection test | Local dev or CI, external anchor target | Tampering detected via the anchor | anchor-tamper-detection report | unassigned |

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
