---
id: W03-E01-S003-T001
type: task
title: Assurance freshness and step-up enforcement (SEC-01 T6)
status: done
parent_story: W03-E01-S003
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-16
depends_on: []
acceptance_criteria:
  - AC-W03-E01-S003-01
artifacts:
  - ART-W03-E01-S003-001
evidence:
  - EV-W03-E01-S003-001
---

# W03-E01-S003-T001 — Assurance freshness and step-up enforcement (SEC-01 T6)

## Task Definition

### Task objective

Bind `auth_time`/`acr`/`amr` into the framework's assurance model and enforce freshness for
step-up, so that a stale `auth_time` with an otherwise-valid `amr` still fails step-up.

### Parent story

W03-E01-S003 — Assurance freshness and credential-scheme distinction.

### Owner

unassigned

### Status

todo

### Dependencies

None within this story (it is the first task). Story-level dependency on W03-E01-S001 (PLAN: T6
depends on T2).

### Detailed work

1. Read `kernel/auth`'s current step-up/assurance code path at this task's actual start commit to
   confirm the exact current extent of AMR plumbing (per T6's own risk note, "already exists") and
   whether any `auth_time` freshness check currently exists.
2. Implement or correct freshness enforcement so that a stale `auth_time` fails step-up regardless
   of `amr` validity — bind `auth_time`, `acr`, and `amr` together into the assurance evaluation.
3. Write the fail-first test: confirm today's actual behavior (a stale `auth_time` with valid `amr`
   currently passes, or is incompletely checked) before the fix.
4. Write the step-up freshness test proving the fix: stale `auth_time`, valid `amr` → step-up
   fails.

### Expected files or components affected

`kernel/auth` (step-up/assurance code path — exact file to be confirmed at implementation time).

### Expected output

A stale `auth_time` with an otherwise-valid `amr` fails step-up.

### Required artifacts

ART-W03-E01-S003-001 (assurance-freshness binding/enforcement code change).

### Required evidence

EV-W03-E01-S003-001 (step-up freshness test report).

### Related acceptance criteria

AC-W03-E01-S003-01.

### Completion criteria

The step-up freshness test passes: stale `auth_time` + valid `amr` → step-up correctly fails.

### Verification method

Direct test execution (fail-first, then pass-after), logged output retained as evidence.

### Risks

Low-moderate — PLAN's own risk note: "`AMR` plumbing already exists — additive, moderate risk."

### Rollback or recovery considerations

Revert if the freshness enforcement rejects a currently-valid, non-stale step-up flow due to an
implementation error in the freshness threshold or comparison logic — investigate the specific
failure before reverting.

## Implementation Record

### What was actually implemented

Bound `auth_time`/`acr`/`amr` into the framework's assurance model and enforced freshness for
step-up.

- Added `AuthTime *jwt.NumericDate` and `ACR string` to `auth.Claims`.
- Added `AuthTime time.Time`, `ACR string`, and `CredentialScheme` to `authz.Actor`.
- `Verifier.Actor` now propagates `AuthTime`, `ACR`, `AMR`, and sets `CredentialScheme = CredentialUser`.
- Added `MaxAge time.Duration` to `authz.StepUpPolicy`.
- Added `StepUpMaxAge time.Duration` to `authz.Options` as the deployment default for the plain
  `step_up: true` shorthand.
- Updated the evaluator's step-up gate: when a MaxAge is configured, a stale (or zero) `AuthTime`
  fails step-up before the AMR check, producing `Reason = "step_up_freshness_required"` and
  `StepUpRequired = true`.
- Added `WithAuthTime` and `WithACR` options to `testkit.TokenIssuer`.

### Components changed

`kernel/auth`, `kernel/authz`, `testkit`.

### Files changed

- `kernel/auth/auth.go`
- `kernel/authz/authz.go`
- `kernel/authz/registry.go`
- `kernel/authz/evaluator.go`
- `testkit/auth.go`
- `kernel/authz/assurance_freshness_test.go` (new)
- `kernel/auth/assurance_internal_test.go` (new)

### Interfaces introduced or changed

- `auth.Claims` gains `AuthTime` and `ACR`.
- `authz.Actor` gains `AuthTime`, `ACR`, and `CredentialScheme`.
- `authz.StepUpPolicy` gains `MaxAge`.
- `authz.Options` gains `StepUpMaxAge`.

### Configuration changes

*Not applicable.*

### Schema or migration changes

*Not applicable.*

### Security changes

Closes the "expired step-up" gap in SEC-01's required test-class list: a stale `auth_time` with an
otherwise-valid `amr` now fails step-up.

### Observability changes

*Not applicable.*

### Tests added or modified

- `kernel/authz/assurance_freshness_test.go`: stale/fresh/zero AuthTime, default MaxAge shorthand,
  backward compatibility.
- `kernel/auth/assurance_internal_test.go`: AuthTime/ACR/CredentialScheme propagation from Claims
  to Actor.

### Commits

*Local working changes; no commit authored in this session.*

### Pull requests

*Not applicable for this session.*

### Implementation dates

2026-07-13.

### Technical debt introduced

*None.*

### Known limitations

DB-backed tests in `./kernel/authz/...` are skipped when `DATABASE_URL` is unavailable. All
non-DB tests pass.

### Follow-up items

- Re-run DB-backed tests with `DATABASE_URL` set.

### Relationship to the approved plan

Matches `plan.md`: T6 bound assurance freshness and enforced it at step-up.
## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W03-E01-S003-01 | Stale `auth_time`, valid `amr`, attempt step-up | Local dev or CI | Step-up fails | unit/adversarial test report | unassigned |

### Actual result

PASS — DB-backed test re-run 2026-07-16 confirmed stale `auth_time` with valid `amr` correctly fails step-up.

### Pass or fail

PASS

### Evidence identifier

EV-W03-E01-S003-001 (kernel/authz/assurance_freshness_test.go, kernel/auth/assurance_internal_test.go)

### Execution date

2026-07-16

### Commit or revision

HEAD (at time of closure.md); tests in kernel/authz/ and kernel/auth/

### Environment

Local dev with `DATABASE_URL` set to test database

### Reviewer

Independent review (closure.md generated 2026-07-16)

### Findings

None — assurance freshness enforcement working as designed

### Retest status

PASS (2026-07-16 re-run)

### Final conclusion

Acceptance criterion AC-W03-E01-S003-01 verified by independent re-run 2026-07-16.

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
