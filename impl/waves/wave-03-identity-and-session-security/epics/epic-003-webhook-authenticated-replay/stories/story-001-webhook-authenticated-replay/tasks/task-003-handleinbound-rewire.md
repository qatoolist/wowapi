---
id: W03-E03-S001-T003
type: task
title: HandleInbound rewire to Envelope-only (SEC-03 T3)
status: done
parent_story: W03-E03-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on:
  - W03-E03-S001-T001
  - W03-E03-S001-T002
acceptance_criteria:
  - AC-W03-E03-S001-03
artifacts:
  - ART-W03-E03-S001-004
evidence:
  - EV-W03-E03-S001-003
---

# W03-E03-S001-T003 — HandleInbound rewire to Envelope-only (SEC-03 T3)

## Task Definition

### Task objective

Rewire `HandleInbound` so no security decision (replay-window check, dedup
check) reads a raw `InboundIn` field — every such decision is sourced
exclusively from `Envelope`.

### Parent story

W03-E03-S001 — Bind webhook replay and dedup to provider-authenticated data.

### Owner

unassigned

### Status

done

### Dependencies

W03-E03-S001-T001, W03-E03-S001-T002 — PLAN's own Depends-on column for T3:
"T1, T2."

### Detailed work

1. Read `kernel/webhook/service.go` (`HandleInbound`) at this task's actual
   start commit, identifying every read of `InboundIn.Timestamp` and
   `InboundIn.ExternalEventID` used in a replay-window or dedup decision.
2. Replace each such read with the corresponding `Envelope` field
   (`Envelope.OccurredAt`, `Envelope.EventID`).
3. Preserve the existing defense-in-depth measure forcing
   `external_event_id = nil` on a failed signature — this story does not remove
   it, it closes the adjacent successful-signature gap.
4. Write the adversarial tamper matrix: body, timestamp, event-ID, key-ID, and
   signature-version, each independently manipulated, asserting the
   replay-window/dedup decision is unaffected by any field not covered by the
   signature.

### Expected files or components affected

`kernel/webhook/service.go` (`HandleInbound`, `dedupExtID`).

### Expected output

No security decision in `HandleInbound` reads a raw `InboundIn` field; the
adversarial tamper matrix passes for all 5 independently manipulated fields.

### Required artifacts

ART-W03-E03-S001-004 (rewired `HandleInbound`).

### Required evidence

EV-W03-E03-S001-003 (adversarial tamper-matrix test output).

### Related acceptance criteria

AC-W03-E03-S001-03.

### Completion criteria

The tamper matrix proves all 5 manipulated-field cases are inert to the
replay-window/dedup decision.

### Verification method

Direct test execution against the adversarial tamper matrix, logged output
retained as evidence.

### Risks

Moderate, per PLAN's own T3 risk note: "review against existing dedup-spoofing
mitigation" — this task must not regress the existing failed-signature
defense-in-depth measure while closing the successful-signature gap.

### Rollback or recovery considerations

If the rewire is found to have broken a legitimate existing consumer's dedup
behavior, the change is revertible as part of this story's single coordinated
commit — see `../plan.md` "Rollback strategy."

## Implementation Record

### What was actually implemented

- `HandleInbound` now calls `v.Verify(...)` and captures the returned
  `Envelope`.
- Replay-window check uses `env.OccurredAt` exclusively; `in.Timestamp` is no
  longer read on the success path.
- Dedup id uses `dedupExtID(env)`, which reads `env.EventID` or synthesizes a
  stable id from `env.CanonicalBody`; `in.ExternalEventID` is no longer used on
  the success path.
- `dedupExtID` signature changed from `func(InboundIn) string` to
  `func(Envelope) string`.
- Payload parsing uses `env.CanonicalBody` instead of `in.RawBody`.
- Existing defense-in-depth on signature failure (forcing
  `external_event_id = nil`) is preserved.
- Added `TestIntegrationHandleInbound_TamperMatrix` covering the five
  independently manipulated fields, plus an updated
  `TestIntegrationHandleInbound_TimestampOutOfWindow` using a deterministic test
  verifier to exercise the replay window with a stale authenticated timestamp.

### Components changed

`kernel/webhook` package: `service.go`, `internals_test.go`, `webhook_test.go`.

### Files changed

- `kernel/webhook/service.go`
- `kernel/webhook/internals_test.go`
- `kernel/webhook/webhook_test.go`

### Interfaces introduced or changed

None; consumes the interface from T001.

### Configuration changes

Not applicable.

### Schema or migration changes

Not applicable.

### Security changes

This task enforces the story's core security property: replay-window and dedup
decisions are bound exclusively to authenticated data returned by the verifier.

### Observability changes

Not applicable.

### Tests added or modified

- `TestIntegrationHandleInbound_TamperMatrix` in `kernel/webhook/webhook_test.go`
  (five sub-cases: body changed/re-signed, body tampered without re-signing,
  timestamp manipulated, external event id manipulated, key id header
  manipulated, signature version header manipulated).
- Updated `TestIntegrationHandleInbound_TimestampOutOfWindow` to use a
  deterministic `testVerifier` returning a stale `OccurredAt`.
- Updated `TestDedupExtID` in `kernel/webhook/internals_test.go` to use
  `Envelope`.

### Commits

TBD at story closure.

### Pull requests

TBD at story closure.

### Implementation dates

2026-07-13.

### Technical debt introduced

None.

### Known limitations

None beyond the documented `HMACVerifier` limitation.

### Follow-up items

None.

### Relationship to the approved plan

Implemented as planned.

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W03-E03-S001-03 | Run the adversarial tamper matrix: body, timestamp, event-ID, key-ID, signature-version, each independently manipulated, against `HandleInbound` | Local dev or CI | No security decision reads a raw `InboundIn` field; all 5 manipulated-field cases are inert to the replay-window/dedup decision | adversarial tamper-matrix test report | unassigned |

### Actual result

```
go test ./kernel/webhook -run 'TestIntegrationHandleInbound_TamperMatrix' -v
=== RUN   TestIntegrationHandleInbound_TamperMatrix
=== RUN   TestIntegrationHandleInbound_TamperMatrix/body_changed_and_re-signed
=== RUN   TestIntegrationHandleInbound_TamperMatrix/body_tampered_without_re-signing
=== RUN   TestIntegrationHandleInbound_TamperMatrix/timestamp_manipulated
=== RUN   TestIntegrationHandleInbound_TamperMatrix/external_event_id_manipulated
=== RUN   TestIntegrationHandleInbound_TamperMatrix/key_id_header_manipulated
=== RUN   TestIntegrationHandleInbound_TamperMatrix/signature_version_header_manipulated
--- PASS: TestIntegrationHandleInbound_TamperMatrix
    --- PASS: .../body_changed_and_re-signed
    --- PASS: .../body_tampered_without_re-signing
    --- PASS: .../timestamp_manipulated
    --- PASS: .../external_event_id_manipulated
    --- PASS: .../key_id_header_manipulated
    --- PASS: .../signature_version_header_manipulated
PASS
ok      github.com/qatoolist/wowapi/kernel/webhook
```

### Pass or fail

PASS.

### Evidence identifier

EV-W03-E03-S001-003.

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

Task complete; AC-W03-E03-S001-03 satisfied.

## Deviations Record

No deviations recorded.
