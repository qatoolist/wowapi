---
id: W04-E04-S003-T001
type: task
title: Migration-currency readiness check
status: done
parent_story: W04-E04-S003
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on: []
acceptance_criteria:
  - AC-W04-E04-S003-01
artifacts:
  - ART-W04-E04-S003-001
  - ART-W04-E04-S003-004
evidence:
  - EV-W04-E04-S003-001
---

# W04-E04-S003-T001 — Migration-currency readiness check

## Task Definition

### Task objective

Add a migration-currency check to the generated readiness template so `/readyz` fails (503) when the
applied-migration version lags the expected version, closing the confirmed gap between the health
contract's own documentation and the generated template's actual `"db"`/`"seeds"`-only checks.

### Parent story

W04-E04-S003 — Readiness and configuration diagnostics truthfulness.

### Owner

unassigned

### Status

todo

### Dependencies

None.

### Detailed work

1. Re-read the generated `cmd/api/main.go.tmpl`'s readiness map and `app/health.go:9-14`'s
   documented DB/migration-check contract comment at this task's actual start commit to confirm the
   current-state gap still holds.
2. Determine the "expected migration version" source (PLAN T1's own risk note: "needs a stable
   'expected migration version' source") — likely derived from the existing migration-registration
   mechanism's own highest-numbered/latest migration.
3. Implement the migration-currency check; wire it into the generated readiness template's map
   alongside the existing `"db"`/`"seeds"` checks.
4. Write the stale-migration 503 integration test: boot against a database at a lagging migration
   version, assert the readiness endpoint returns 503.
5. Document the check's failure condition and its expected-version source.

### Expected files or components affected

The generated `cmd/api/main.go.tmpl`; `app/health.go` (or its readiness-check registration logic); a
new integration test file for the stale-migration 503 test.

### Expected output

A working migration-currency readiness check wired into the generated template; a passing
stale-migration 503 integration test; documentation of the check.

### Required artifacts

ART-W04-E04-S003-001 (migration-currency readiness check), ART-W04-E04-S003-004 (documentation,
shared with T002/T003).

### Required evidence

EV-W04-E04-S003-001 (stale-migration integration-test report).

### Related acceptance criteria

AC-W04-E04-S003-01.

### Completion criteria

`/readyz` returns 503 when booted against a stale-migrated database, proven by the stale-migration
integration test.

### Verification method

Direct execution of the stale-migration 503 integration test.

### Risks

Medium — per PLAN T1's own risk column: "needs a stable 'expected migration version' source." An
unstable or incorrectly-derived expected-version source could produce false 503s on a correctly-
migrated database, or false 200s on a genuinely stale one.

### Rollback or recovery considerations

Additive, non-schema change — a code-level revert restores the previous (contract-violating)
behavior if the check proves unstable in production; this is a deliberate, documented regression if
ever exercised, not a silent one.

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
| AC-W04-E04-S003-01 | Run stale-migration 503 integration test | Local dev or CI, PostgreSQL instance at lagging migration version | `/readyz` returns 503 | stale-migration integration-test report | unassigned |

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
