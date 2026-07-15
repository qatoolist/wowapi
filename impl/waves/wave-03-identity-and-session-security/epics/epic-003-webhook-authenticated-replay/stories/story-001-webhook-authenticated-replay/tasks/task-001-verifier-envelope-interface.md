---
id: W03-E03-S001-T001
type: task
title: Verifier interface change to (Envelope, error) (SEC-03 T1)
status: done
parent_story: W03-E03-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on: []
acceptance_criteria:
  - AC-W03-E03-S001-01
artifacts:
  - ART-W03-E03-S001-001
  - ART-W03-E03-S001-002
evidence:
  - EV-W03-E03-S001-001
---

# W03-E03-S001-T001 — Verifier interface change to (Envelope, error) (SEC-03 T1)

## Task Definition

### Task objective

Change the `Verifier` interface to return `(Envelope, error)` instead of a bare
`error`. Define `Envelope{CanonicalBody, EventID, OccurredAt, SignatureVersion,
KeyID}`. Update `HMACVerifier` and `FakeVerifier` to satisfy the new interface.

### Parent story

W03-E03-S001 — Bind webhook replay and dedup to provider-authenticated data.

### Owner

unassigned

### Status

done

### Dependencies

None. This is the foundational interface change all subsequent tasks in this
story build on.

### Detailed work

1. Read `kernel/webhook/verifier.go` at this task's actual start commit,
   confirming the exact current interface signature and both implementations.
2. Fresh-confirm PLAN's own "zero custom `Verifier` implementation anywhere in
   wowsociety" claim via a current grep against the codebase, per RISK-W03-006's
   mitigation.
3. Define `Envelope{CanonicalBody, EventID, OccurredAt, SignatureVersion,
   KeyID}`, following existing `kernel/webhook` naming/type idioms.
4. Change `Verifier.Verify`'s signature to `(Envelope, error)`.
5. Update `HMACVerifier` to compile against and satisfy the new interface
   (synthesis logic itself is T002's scope — this task establishes the interface
   and a correct, if minimal, `Envelope` population).
6. Update `FakeVerifier` to compile against and satisfy the new interface,
   preserving its existing test-double behavior as closely as possible.
7. Write unit tests for both implementations against the new interface.

### Expected files or components affected

`kernel/webhook/verifier.go`, `kernel/webhook/webhook.go`.

### Expected output

The `Verifier` interface returns `(Envelope, error)`; `Envelope` is defined;
`HMACVerifier` and `FakeVerifier` both compile against and satisfy the new
interface; unit tests for both pass.

### Required artifacts

ART-W03-E03-S001-001 (`Envelope` type definition and the changed `Verifier`
interface), ART-W03-E03-S001-002 (updated `HMACVerifier` and `FakeVerifier`
implementations).

### Required evidence

EV-W03-E03-S001-001 (unit test output for both `Verifier` implementations).

### Related acceptance criteria

AC-W03-E03-S001-01.

### Completion criteria

The new interface compiles; both implementations satisfy it; unit tests pass for
both.

### Verification method

Direct unit test execution, logged output retained as evidence.

### Risks

**Breaking interface change**, per PLAN's own T1 risk note, verbatim. A compile
error is the safe failure mode for any undiscovered consumer — see RISK-W03-006.

### Rollback or recovery considerations

If an undiscovered consumer is found via the fresh wowsociety re-confirmation
(step 2), coordinate the interface change with that consumer before merge rather
than proceeding unilaterally.

## Implementation Record

### What was actually implemented

- Added `Envelope` type to `kernel/webhook/verifier.go` with the five mandated
  fields and godoc describing the authenticated-fields-only contract.
- Updated the `Verifier` interface in `kernel/webhook/webhook.go` to return
  `(Envelope, error)` and documented the failure contract.
- Updated `HMACVerifier.Verify` to return a populated `Envelope` on success and
  a zero `Envelope` on authentication failure.
- Updated `FakeVerifier.Verify` to return a populated `Envelope` on success and
  a zero `Envelope` on failure.
- Added `kernel/webhook/verifier_envelope_test.go` with unit tests for both
  verifiers against the new interface.
- Updated existing `kernel/webhook/coverage_test.go` tests to handle the new
  `(Envelope, error)` signature.

### Components changed

`kernel/webhook` package: `verifier.go`, `webhook.go`, `verifier_envelope_test.go`,
`coverage_test.go`.

### Files changed

- `kernel/webhook/verifier.go`
- `kernel/webhook/webhook.go`
- `kernel/webhook/verifier_envelope_test.go` (new)
- `kernel/webhook/coverage_test.go`

### Interfaces introduced or changed

- `Verifier.Verify(secret string, body []byte, headers map[string]string)
  (Envelope, error)` (changed from returning `error` only).
- New `Envelope` type.

### Configuration changes

Not applicable.

### Schema or migration changes

Not applicable.

### Security changes

This task introduces the `Envelope` contract that underpins the story's security
property. The full enforcement occurs in T003.

### Observability changes

Not applicable.

### Tests added or modified

- New `kernel/webhook/verifier_envelope_test.go`:
  - `TestUnitHMACVerifier_Envelope`
  - `TestUnitHMACVerifier_OccurredAtIgnoresTimestampHeader`
  - `TestUnitHMACVerifier_BadSignature`
  - `TestUnitFakeVerifier_Envelope`
- Updated `TestUnitFakeVerifier` and `TestUnitHMACVerifier_MissingHeader` in
  `kernel/webhook/coverage_test.go`.

### Commits

TBD at story closure.

### Pull requests

TBD at story closure.

### Implementation dates

2026-07-13.

### Technical debt introduced

None.

### Known limitations

None beyond those documented in T002.

### Follow-up items

None.

### Relationship to the approved plan

Implemented as planned.

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W03-E03-S001-01 | Run unit tests for `HMACVerifier` and `FakeVerifier` against the new `(Envelope, error)` interface | Local dev or CI | Both implementations compile against and satisfy the new interface; unit tests pass | unit test report | unassigned |

### Actual result

```
go test ./kernel/webhook -run 'TestUnit(HMACVerifier|FakeVerifier)' -v
=== RUN   TestUnitFakeVerifier
--- PASS: TestUnitFakeVerifier (0.00s)
=== RUN   TestUnitHMACVerifier_MissingHeader
--- PASS: TestUnitHMACVerifier_MissingHeader (0.00s)
=== RUN   TestUnitHMACVerifier_Envelope
--- PASS: TestUnitHMACVerifier_Envelope (0.00s)
=== RUN   TestUnitHMACVerifier_OccurredAtIgnoresTimestampHeader
--- PASS: TestUnitHMACVerifier_OccurredAtIgnoresTimestampHeader (0.00s)
=== RUN   TestUnitHMACVerifier_BadSignature
--- PASS: TestUnitHMACVerifier_BadSignature (0.00s)
=== RUN   TestUnitFakeVerifier_Envelope
--- PASS: TestUnitFakeVerifier_Envelope (0.00s)
PASS
ok      github.com/qatoolist/wowapi/kernel/webhook
```

### Pass or fail

PASS.

### Evidence identifier

EV-W03-E03-S001-001.

### Execution date

2026-07-13.

### Commit or revision

TBD at story closure.

### Environment

Local dev with `DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable`.

### Reviewer

unassigned.

### Findings

None.

### Retest status

Not retested.

### Final conclusion

Task complete; AC-W03-E03-S001-01 satisfied.

## Deviations Record

No deviations recorded.
