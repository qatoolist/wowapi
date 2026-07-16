---
id: W02-E01-S003-T001
type: task
title: Canary/deploy-N tooling
status: done
parent_story: W02-E01-S003
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-16
depends_on:
  - W02-E01-S002
acceptance_criteria:
  - AC-W02-E01-S003-01
artifacts:
  - ART-W02-E01-S003-001
  - ART-W02-E01-S003-006
evidence:
  - EV-W02-E01-S003-001
---

# W02-E01-S003-T001 — Canary/deploy-N tooling

## Task Definition

### Task objective

Implement canary/deploy-N tooling — N alongside N-1 with soak metrics — and prove, via the named
canary test, both explicitly-required legs: N-1 code runs correctly against the N-expanded schema,
and N code runs correctly both before and after backfill (PLAN DATA-09 T6's bolded acceptance
criterion).

### Parent story

W02-E01-S003 — Canary, switch, and contract-phase tooling with the full CI drill pipeline.

### Owner

unassigned

### Status

todo

### Dependencies

W02-E01-S002 (PLAN T6's "Depends-on" column names T5 — canary runs after validation).

### Detailed work

1. Decide the mechanism for materializing an N-1 application version in a test environment
   (prior-release container image vs. checked-out prior build) — `plan.md`'s "Unresolved questions"
   step 2, a prerequisite for everything below.
2. Implement canary orchestration: deploy N-1 alongside N against a schema expanded by S002's
   tooling; collect soak metrics.
3. Make soak duration/threshold parameters configurable — do not hardcode guessed values. PLAN T6's
   own risk column: "No production telemetry baseline exists — soak duration/thresholds are a
   genuine, currently unresolvable judgment gap" (RISK-W02-003); the go/no-go decision is human per
   T6's own classification column.
4. Write the named canary test (`DATA-09/canary-soak/`) covering both required legs: N-1 on the
   expanded N schema, and N code before and after backfill.
5. Document the configuration surface with an explicit note that value calibration is a per-rollout
   human judgment.

### Expected files or components affected

New canary-tooling package (location TBD per `plan.md`).

### Expected output

Canary tooling with configurable soak parameters and a passing named canary test covering both
required legs.

### Required artifacts

ART-W02-E01-S003-001 (canary tooling), ART-W02-E01-S003-006 (documentation, shared with
T002/T003/T004).

### Required evidence

EV-W02-E01-S003-001 (named canary test output, both legs).

### Related acceptance criteria

AC-W02-E01-S003-01.

### Completion criteria

The named canary test passes both required legs, evidenced against a named commit SHA; soak
parameters are demonstrably configurable.

### Verification method

Direct execution of the named canary test against a live PostgreSQL instance with both application
versions materialized; configuration-surface inspection.

### Risks

RISK-W02-003 (the soak-calibration judgment gap) — this task's mitigation is the configurable-
parameter design; the gap itself is accepted, not resolved, per the wave-level risk register.

### Rollback or recovery considerations

The canary tooling executes no production migration; reverting it is a plain code revert. A failed
canary within the tooling's own operation must leave a forward-recoverable state (the forward-
recovery drill in T003's scope exercises this across all phases).

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

*Not applicable — canary tooling orchestrates deploys against fixture migrations; it adds no
application schema.*

### Security changes

*Not applicable.*

### Observability changes

*Not yet implemented — soak-metric collection is anticipated.*

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

*Not yet implemented — the soak-calibration gap is expected to be recorded here as a known,
accepted limitation once the tooling lands.*

### Follow-up items

*Not yet implemented.*

### Relationship to the approved plan

*Not yet implemented.*

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W02-E01-S003-01 | Named canary test, both legs | Local dev or CI, PostgreSQL, N-1 and N versions materialized | N-1 accepts expanded N schema; N correct before/after backfill; parameters configurable | integration-test report | unassigned |

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
