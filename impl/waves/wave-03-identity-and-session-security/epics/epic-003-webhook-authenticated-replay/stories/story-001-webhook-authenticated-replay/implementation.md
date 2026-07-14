---
id: IMPL-W03-E03-S001
type: implementation-record
parent_story: W03-E03-S001
status: produced
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Implementation record — W03-E03-S001

## What was actually implemented

- Added `Envelope{CanonicalBody, EventID, OccurredAt, SignatureVersion, KeyID}`
  to `kernel/webhook/verifier.go` with godoc describing the
  authenticated-fields-only contract.
- Changed the `Verifier` interface in `kernel/webhook/webhook.go` to return
  `(Envelope, error)` and documented the failure contract.
- Updated `HMACVerifier` to synthesize `EventID` from the authenticated body
  (`sha256:<hex(SHA-256(body))>`) and `OccurredAt` from receipt time
  (`time.Now()`), with explicit documentation that this is unsuitable for
  timestamped-provider protocols.
- Updated `FakeVerifier` to satisfy the new interface and return a valid
  `Envelope`.
- Rewired `HandleInbound` in `kernel/webhook/service.go` to source the
  replay-window check from `env.OccurredAt` and the dedup id from
  `dedupExtID(env)`, never reading `in.Timestamp` or `in.ExternalEventID` on
  the success path.
- Changed `dedupExtID` to accept `Envelope` instead of `InboundIn`.
- Preserved the existing defense-in-depth forcing `external_event_id = nil` on
  signature failure.
- Authored the provider-verifier contract document at
  `artifacts/provider-verifier-contract.md`.

## Components changed

`kernel/webhook` package and the story's `artifacts/` directory.

## Files changed

- `kernel/webhook/verifier.go`
- `kernel/webhook/webhook.go`
- `kernel/webhook/service.go`
- `kernel/webhook/internals_test.go`
- `kernel/webhook/coverage_test.go`
- `kernel/webhook/verifier_envelope_test.go` (new)
- `kernel/webhook/webhook_test.go`
- `impl/waves/wave-03-identity-and-session-security/epics/epic-003-webhook-authenticated-replay/stories/story-001-webhook-authenticated-replay/artifacts/provider-verifier-contract.md` (new)

## Interfaces introduced or changed

- New `kernel/webhook.Envelope` type.
- `kernel/webhook.Verifier.Verify` changed from
  `func(secret string, body []byte, headers map[string]string) error` to
  `func(secret string, body []byte, headers map[string]string) (Envelope, error)`.
- `kernel/webhook.dedupExtID` changed from `func(InboundIn) string` to
  `func(Envelope) string` (unexported helper).

## Configuration changes

None.

## Schema or migration changes

None.

## Security changes

Replay-window and dedup decisions in `HandleInbound` are now bound exclusively
to authenticated data returned by the verifier. Caller-supplied
`InboundIn.Timestamp` and `InboundIn.ExternalEventID` are no longer trusted on
the success path.

## Observability changes

None.

## Tests added or modified

- `kernel/webhook/verifier_envelope_test.go` (new):
  - `TestUnitHMACVerifier_Envelope`
  - `TestUnitHMACVerifier_OccurredAtIgnoresTimestampHeader`
  - `TestUnitHMACVerifier_BadSignature`
  - `TestUnitFakeVerifier_Envelope`
- `kernel/webhook/webhook_test.go`:
  - `TestIntegrationHandleInbound_TimestampManipulationImmune`
  - `TestIntegrationHandleInbound_TamperMatrix` (five independently manipulated
    fields)
  - Updated `TestIntegrationHandleInbound_TimestampOutOfWindow` to use a
    deterministic test verifier.
- `kernel/webhook/internals_test.go`:
  - Updated `TestDedupExtID` to use `Envelope`.
- `kernel/webhook/coverage_test.go`:
  - Updated `TestUnitFakeVerifier` and `TestUnitHMACVerifier_MissingHeader` for
    the new `(Envelope, error)` signature.

## Commits

TBD at story closure.

## Pull requests

TBD at story closure.

## Implementation dates

2026-07-13.

## Technical debt introduced

None.

## Known limitations

`HMACVerifier` is documented as unsuitable for provider protocols requiring a
provider-asserted timestamp. Such providers need a dedicated `Verifier`
implementation consistent with the provider-verifier contract.

## Follow-up items

None.

## Relationship to the approved plan

Implemented as planned; no deviations recorded.
