---
id: W03-E03-S001-T002
type: task
title: HMACVerifier authenticated-data synthesis (SEC-03 T2)
status: done
parent_story: W03-E03-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on:
  - W03-E03-S001-T001
acceptance_criteria:
  - AC-W03-E03-S001-02
artifacts:
  - ART-W03-E03-S001-003
evidence:
  - EV-W03-E03-S001-002
---

# W03-E03-S001-T002 — HMACVerifier authenticated-data synthesis (SEC-03 T2)

## Task Definition

### Task objective

Make `HMACVerifier` synthesize `EventID` and `OccurredAt` from the authenticated
body and/or receipt time only — never from caller-supplied request fields.
Document this synthesis approach as unsuitable for timestamped-provider
protocols requiring provider-asserted timestamps.

### Parent story

W03-E03-S001 — Bind webhook replay and dedup to provider-authenticated data.

### Owner

unassigned

### Status

done

### Dependencies

W03-E03-S001-T001 — PLAN's own Depends-on column for T2: "T1."

### Detailed work

1. Implement `EventID` synthesis from the authenticated body (a stable SHA-256
   hash) — never from `InboundIn.ExternalEventID`.
2. Implement `OccurredAt` synthesis from the receipt time (server-side clock at
   verification time) — never from `InboundIn.Timestamp`.
3. Document, in code comments and the T004 contract document, that this
   synthesis approach is unsuitable for provider protocols requiring a
   provider-asserted timestamp rather than a receipt-time approximation.
4. Write the dedicated test proving `Envelope.OccurredAt` is immune to a
   manipulated `in.Timestamp`.

### Expected files or components affected

`kernel/webhook/verifier.go` (`HMACVerifier`'s `Verify` implementation);
`artifacts/provider-verifier-contract.md` (T004).

### Expected output

`Envelope` never surfaces caller-supplied fields; `OccurredAt` is proven immune
to a manipulated `in.Timestamp`.

### Required artifacts

ART-W03-E03-S001-003 (`HMACVerifier`'s authenticated-data synthesis
implementation).

### Required evidence

EV-W03-E03-S001-002 (`OccurredAt`-immune-to-manipulated-timestamp test output).

### Related acceptance criteria

AC-W03-E03-S001-02.

### Completion criteria

The dedicated test proves `OccurredAt` does not change when `in.Timestamp` is
manipulated.

### Verification method

Direct test execution, logged output retained as evidence.

### Risks

Moderate, per PLAN's own T2 risk note. The receipt-time-based `OccurredAt`
synthesis is a real behavioral limitation for timestamped-provider protocols —
this task does not silently paper over that limitation, it documents it
explicitly.

### Rollback or recovery considerations

If the receipt-time synthesis approach is found unsuitable for a specific
provider integration post-rollout, that provider requires its own `Verifier`
implementation (consistent with the T004 contract document's guidance) rather
than a change to `HMACVerifier`'s own documented, limited behavior.

## Implementation Record

### What was actually implemented

- `HMACVerifier.Verify` computes `EventID` as `sha256:<hex(SHA-256(body))>`,
  derived entirely from the authenticated body.
- `HMACVerifier.Verify` sets `OccurredAt` to `time.Now()` at verification time,
  independent of any caller-supplied timestamp.
- Added explicit godoc on `HMACVerifier.Verify` documenting the body-only
  authentication and the resulting limitation for timestamped-provider
  protocols.
- Added `TestUnitHMACVerifier_OccurredAtIgnoresTimestampHeader` (unit-level) and
  `TestIntegrationHandleInbound_TimestampManipulationImmune` (integration-level)
  proving `OccurredAt` is immune to manipulated timestamps/headers.

### Components changed

`kernel/webhook` package: `verifier.go`, `verifier_envelope_test.go`,
`webhook_test.go`.

### Files changed

- `kernel/webhook/verifier.go`
- `kernel/webhook/verifier_envelope_test.go`
- `kernel/webhook/webhook_test.go`

### Interfaces introduced or changed

None; consumes the interface from T001.

### Configuration changes

Not applicable.

### Schema or migration changes

Not applicable.

### Security changes

`HMACVerifier` no longer surfaces any caller-supplied field in `Envelope`. This
closes the successful-signature gap for body-only HMAC providers.

### Observability changes

Not applicable.

### Tests added or modified

- `TestUnitHMACVerifier_OccurredAtIgnoresTimestampHeader` in
  `kernel/webhook/verifier_envelope_test.go`.
- `TestIntegrationHandleInbound_TimestampManipulationImmune` in
  `kernel/webhook/webhook_test.go`.

### Commits

TBD at story closure.

### Pull requests

TBD at story closure.

### Implementation dates

2026-07-13.

### Technical debt introduced

None — the timestamped-provider limitation is a documented design constraint,
not technical debt.

### Known limitations

`HMACVerifier` is unsuitable for provider protocols requiring a
provider-asserted timestamp, because the body-only HMAC does not authenticate a
timestamp. Such providers need a dedicated `Verifier`.

### Follow-up items

None.

### Relationship to the approved plan

Implemented as planned.

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W03-E03-S001-02 | Run a dedicated test that manipulates `in.Timestamp` and asserts `Envelope.OccurredAt` is unaffected | Local dev or CI | `OccurredAt` is immune to a manipulated `in.Timestamp` | targeted test report | unassigned |

### Actual result

```
go test ./kernel/webhook -run 'TestIntegrationHandleInbound_TimestampManipulationImmune' -v
=== RUN   TestIntegrationHandleInbound_TimestampManipulationImmune
--- PASS: TestIntegrationHandleInbound_TimestampManipulationImmune (0.10s)
PASS
ok      github.com/qatoolist/wowapi/kernel/webhook
```

### Pass or fail

PASS.

### Evidence identifier

EV-W03-E03-S001-002.

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

Task complete; AC-W03-E03-S001-02 satisfied.

## Deviations Record

No deviations recorded.
