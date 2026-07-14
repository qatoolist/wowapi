---
id: W04-E04-S003-T002
type: task
title: Seed/rule/model-hash readiness reporting
status: done
parent_story: W04-E04-S003
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on: []
acceptance_criteria:
  - AC-W04-E04-S003-02
artifacts:
  - ART-W04-E04-S003-002
  - ART-W04-E04-S003-004
evidence:
  - EV-W04-E04-S003-002
---

# W04-E04-S003-T002 — Seed/rule/model-hash readiness reporting

## Task Definition

### Task objective

Add seed/rule/model-hash checks to readiness so the payload reports migration version, seed/rule
hash, and model hash — the seed/rule-hash and migration-version portions independently of AR-01, the
model-hash portion contingent on AR-01's model hash being available.

### Parent story

W04-E04-S003 — Readiness and configuration diagnostics truthfulness.

### Owner

unassigned

### Status

todo

### Dependencies

None at task level, though T1's migration-currency check can ship independently per PLAN's own
framing, and this task's model-hash portion has an internal contingency on AR-01's model hash
(W05 scope) being available at implementation time.

### Detailed work

1. Confirm AR-01's model hash's availability at this task's actual implementation time.
2. Implement seed/rule-hash reporting in the readiness payload (independent of AR-01).
3. Implement model-hash reporting in the readiness payload, contingent on step 1's confirmation; if
   unavailable, record the contingency explicitly rather than silently omitting or fabricating a
   value.
4. Write the full-readiness-payload integration test: confirm the payload reports migration version,
   seed/rule hash, and (if available) model hash.
5. Document the readiness payload's new fields.

### Expected files or components affected

The generated `cmd/api/main.go.tmpl` or `app/health.go`'s readiness-check registration logic (shared
surface with T001); a new integration test file for the full-readiness-payload test.

### Expected output

A readiness payload reporting migration version, seed/rule hash, and (if AR-01 available) model
hash; a passing full-readiness-payload integration test; documentation of the new fields.

### Required artifacts

ART-W04-E04-S003-002 (seed/rule/model-hash readiness reporting), ART-W04-E04-S003-004
(documentation, shared with T001/T003).

### Required evidence

EV-W04-E04-S003-002 (full-readiness-payload integration-test report).

### Related acceptance criteria

AC-W04-E04-S003-02.

### Completion criteria

The readiness payload reports migration version and seed/rule hash unconditionally; model hash is
reported if AR-01's model hash is available, with the contingency status honestly recorded if not.

### Verification method

Direct execution of the full-readiness-payload integration test.

### Risks

Medium-high for the model-hash portion specifically — per PLAN T2's own risk column, this portion is
"blocked on Wave 1" (AR-01). The seed/rule-hash and migration-version portions carry the same
medium risk profile as T1's own expected-version-source concern.

### Rollback or recovery considerations

Additive, non-schema payload change — revertible at the code level without a data-migration concern.

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

*Not applicable.*

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
| AC-W04-E04-S003-02 | Run full-readiness-payload integration test | Local dev or CI | Payload reports migration version, seed/rule hash, model hash (if available) | full-readiness-payload integration-test report | unassigned |

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
