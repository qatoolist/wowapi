---
id: W07-E01-S003-T005
type: task
title: Leased-state-machine outbox rework
status: done
parent_story: W07-E01-S003
owner: W07-Scoping-Dispatch.W07E01S003
created_at: 2026-07-12
updated_at: 2026-07-16
depends_on: []
acceptance_criteria:
  - AC-W07-E01-S003-05
artifacts:
  - ART-W07-E01-S003-005
evidence:
  - EV-W07-E01-S003-005
---

# W07-E01-S003-T005 — Leased-state-machine outbox rework

## Task Definition

### Task objective

Rework outbox claim/dispatch into a leased state machine consuming W04's DATA-02/DATA-03 primitives, preserving per-aggregate ordering.

### Parent story

W07-E01-S003

### Owner

W07-Scoping-Dispatch.W07E01S003

### Status

complete

### Dependencies

Hard dependency on W04-E01 (DATA-02) and W04-E02 (DATA-03), already satisfied by this wave's own all-prior-waves entry gate — but this task must re-confirm the primitives' actual shape before beginning implementation, per RISK-W07-E01-001's own mitigation.

### Detailed work

1. Re-confirm W04-E01/E02's DATA-02/DATA-03 lease primitives' actual shape before beginning
   implementation.
2. Rework outbox claim/dispatch into a leased state machine consuming those primitives directly, not a
   parallel fencing mechanism.
3. Ensure no outer transaction spans tenant handlers.
4. Preserve per-aggregate ordering.
5. Run the inherited crash/duplicate-worker chaos tests from DATA-02/DATA-03's own gate.

### Expected files or components affected

kernel/outbox's claim/dispatch implementation.

### Expected output

A leased-state-machine outbox rework passing inherited chaos tests, no outer transaction spanning tenant handlers.

### Required artifacts

ART-W07-E01-S003-005 (leased-state-machine outbox rework).

### Required evidence

EV-W07-E01-S003-005 (inherited chaos test output).

### Related acceptance criteria

AC-W07-E01-S003-05

### Completion criteria

Inherited chaos tests pass; no outer transaction spans tenant handlers; per-aggregate ordering preserved.

### Verification method

Direct execution of the inherited crash/duplicate-worker chaos tests.

### Risks

High — cross-work-package dependency, do not attempt in isolation, per PLAN T5's own risk note; RISK-W07-E01-001 (adaptation-layer risk).

### Rollback or recovery considerations

If W04's primitives prove incompatible without adaptation, halt and escalate to the performance/SRE lead rather than silently hand-rolling a parallel fencing mechanism.

## Implementation Record

### What was actually implemented

The relay now claims an ordered batch in a short app-platform transaction, persists the accepted W04
`lease.Lease` token/generation/expiry, commits, then dispatches each event through its tenant
transaction. Success and failure finalize with the exact token+generation+unexpired predicate.
The claim still selects only the earliest undispatched event per aggregate.

### Components changed

Outbox relay and outbox schema.

### Files changed

`kernel/outbox/relay.go`, `kernel/outbox/relay_lease_test.go`,
`migrations/00047_perf04_sweeper_outbox_leases.sql`.

### Interfaces introduced or changed

`WithRelayLeaseTTL` configures deterministic lease-expiry tests; default remains 30 seconds.

### Configuration changes

None in product configuration.

### Schema or migration changes

Adds nullable outbox `lease_token`, `lease_generation`, and `lease_expires_at` plus a claim index.

### Security changes

Claims/finalization use the app-platform pool; handlers run only inside tenant-bound transactions.
No claim transaction remains open across handlers or external side effects.

### Observability changes

Covered by T6.

### Tests added or modified

Claim-commit probe, lease-expiry duplicate-worker fencing, race-detector stress, existing inbox
idempotency and per-aggregate retry ordering suites.

### Commits

Working tree based on entry SHA `733ef3e`.

### Pull requests

None.

### Implementation dates

2026-07-13.

### Technical debt introduced

None.

### Known limitations

External handler side effects remain at-least-once, as documented by the accepted W04 primitive.

### Follow-up items

None.

### Relationship to the approved plan

Matches plan T5 and directly reuses `kernel/lease.Lease`; no alternate lease primitive was introduced.

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W07-E01-S003-05 | Commit-boundary probe, lease-expiry race test, inherited W04 chaos and existing ordering tests | PostgreSQL 16.9 + race detector | No outer transaction; stale worker fenced; inbox/order contracts pass | chaos report | W07-Scoping-Dispatch.W07E01S003ReviewR |

### Actual result

PASS: a concurrent `FOR UPDATE` succeeds while the tenant handler is blocked, proving claim commit;
worker B reclaims generation 2, stale worker A cannot finalize, one effect/inbox row remains; W04
duplicate-worker chaos and existing per-aggregate ordering tests pass.

### Pass or fail

PASS.

### Evidence identifier

EV-W07-E01-S003-005.

### Execution date

2026-07-13.

### Commit or revision

Working tree based on entry SHA `733ef3e`.

### Environment

darwin/arm64, Go 1.26.5 race detector, PostgreSQL 16.9 Docker service.

### Reviewer

W07-Scoping-Dispatch.W07E01S003ReviewR.

### Findings

The race stress initially exposed a test-only second-lease expiry under instrumentation; worker B's
lease was increased to five seconds while worker A retains the 75ms reclaim trigger. Ten race runs pass.

### Retest status

Focused outbox package, ten race runs, and inherited W04 chaos passed.

### Final conclusion

AC-05 accepted after independent review found no open issues.
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
