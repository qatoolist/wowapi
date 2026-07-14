---
id: W03-E03-S001
type: story
title: Bind webhook replay and dedup to provider-authenticated data
status: ready
wave: W03
epic: W03-E03
owner: unassigned
reviewer: unassigned
priority: high
created_at: 2026-07-12
updated_at: 2026-07-13
source_requirements:
  - SEC-03
depends_on: []
blocks: []
acceptance_criteria:
  - AC-W03-E03-S001-01
  - AC-W03-E03-S001-02
  - AC-W03-E03-S001-03
  - AC-W03-E03-S001-04
artifacts: []
evidence: []
decisions: []
risks:
  - RISK-W03-006
---

# W03-E03-S001 — Bind webhook replay and dedup to provider-authenticated data

## Story ID

W03-E03-S001

## Title

Bind webhook replay and dedup to provider-authenticated data

## Objective

Change the `Verifier` interface to return `(Envelope, error)` instead of a bare `error`; make
`HMACVerifier` synthesize `EventID`/`OccurredAt` from the authenticated body and receipt time only;
rewire `HandleInbound` to source replay-window and dedup decisions exclusively from `Envelope`; and
document the resulting provider-verifier contract. This is PLAN SEC-01 §5.2's SEC-03, T1 through
T4, in full.

## Value to the framework

Today a webhook request's replay-window and dedup checks are driven by caller-supplied,
unauthenticated fields — `InboundIn.Timestamp` and `InboundIn.ExternalEventID` — even when the
request's HMAC signature is valid. This means an attacker who can produce (or replay) a validly
signed body can still manipulate the timestamp or event-ID fields the security decision reads,
because those fields sit outside the signature's own authenticated scope in the current design.
This story closes that gap by making the `Verifier` responsible for producing an authenticated
`Envelope` — a set of fields the verifier itself derives from data the signature covers — and making
`HandleInbound` read exclusively from that envelope for every replay/dedup decision.

## Problem statement

PLAN §5.2's own evidence, cited verbatim: "`HMACVerifier.Verify` (`kernel/webhook/verifier.go:32`)
HMACs the body only, returns only `error`. `HandleInbound` (`service.go:22-114`) drives
replay-window/dedup off the caller-supplied, unauthenticated `InboundIn.Timestamp`/
`ExternalEventID`." PLAN also records a mitigating detail already present in the codebase:
"`service.go:60-63` already forces `external_event_id = nil` on a *failed* signature — good
defense-in-depth, doesn't address the successful-signature gap." This story's entire scope is
closing that successful-signature gap — the case where an attacker has a validly signed body but
can still manipulate the fields the replay/dedup logic reads.

## Source requirements

SEC-03 (T1, T2, T3, T4).

## Current-state assessment

Per PLAN §5.2's own evidence citation (to be re-confirmed at this story's own execution commit,
consistent with the wave-01 precedent of re-running fail-first checks rather than trusting a cited
snapshot blindly):

- `Verifier.Verify` (`kernel/webhook/verifier.go:32`) returns only `error` — no authenticated data
  is surfaced back to the caller beyond a pass/fail signal.
- `HMACVerifier`'s HMAC computation covers only the request body — no timestamp or event-ID field is
  part of the signed data today.
- `HandleInbound` (`kernel/webhook/service.go:22-114`) reads `InboundIn.Timestamp` and
  `InboundIn.ExternalEventID` directly — both caller-supplied, unauthenticated fields — to drive
  replay-window and dedup decisions.
- `service.go:60-63` already forces `external_event_id = nil` on a *failed* signature verification —
  this existing defense-in-depth measure does not address the case this story targets: a *valid*
  signature accompanying a manipulated timestamp or event-ID.

## Desired state

`Verifier.Verify` returns `(Envelope, error)`, where `Envelope{CanonicalBody, EventID, OccurredAt,
SignatureVersion, KeyID}` contains only fields the verifier itself derives from authenticated data.
`HMACVerifier` synthesizes `EventID` and `OccurredAt` from the authenticated body and/or receipt
time, never from caller-supplied request fields, and this behavior is documented as unsuitable for
timestamped-provider protocols requiring provider-asserted timestamps (a limitation the contract
document, T4, states explicitly). `HandleInbound` sources every replay-window and dedup decision
exclusively from `Envelope`, never from `InboundIn` directly. A provider-verifier contract document
exists describing what any `Verifier` implementation must guarantee.

## Scope

- T1 — the `Verifier` interface change to `(Envelope, error)`; `Envelope` type definition; updates
  to both `HMACVerifier` and `FakeVerifier`.
- T2 — `HMACVerifier`'s `EventID`/`OccurredAt` synthesis from authenticated data only, with the
  unsuitable-for-timestamped-providers limitation documented.
- T3 — the `HandleInbound` rewire to source replay-window/dedup exclusively from `Envelope`.
- T4 — the provider-verifier contract document with a reference example.

## Out of scope

- Any provider-specific `Verifier` implementation beyond `HMACVerifier`/`FakeVerifier` — this story
  changes the interface contract and updates its two existing implementations; per this epic's
  `epic.md`, a third provider-specific verifier is out of scope unless one already exists in the
  codebase (to be confirmed, not assumed, at implementation time).
- `kernel/webhook`'s outbound delivery path (`deliverToEndpoint`) — that is PLAN DATA-03 T3's scope
  (W04-E02), a different finding (remote I/O outside transactions).
- SEC-01, SEC-02, SEC-06, DATA-07 — separate findings in this wave with no shared file surface.

## Assumptions

- The exact `Envelope` struct's field types and naming conventions should follow existing
  `kernel/webhook` idioms — this story does not invent a naming convention divorced from the
  package's existing style; the precise types are confirmed by reading `kernel/webhook/verifier.go`
  directly at implementation time, not assumed here.
- PLAN's own citation of "zero custom `Verifier` implementation anywhere in wowsociety" is treated
  as a snapshot to re-confirm (fresh grep) at this story's own execution time, not blindly trusted —
  consistent with RISK-W03-006's framing.
- No `Verifier` implementation beyond `HMACVerifier`/`FakeVerifier` is assumed to exist in wowapi
  itself; this should also be confirmed fresh at implementation time rather than assumed exhaustive
  from PLAN's own citation, which only names these two.

## Dependencies

None within this story — T1 is the foundational interface change all three subsequent tasks build
on, but they are grouped into one story per `impl/analysis/wave-allocation-detail.md`. No dependency
on any other W03 epic; no cross-wave dependency beyond this wave's own entry criteria.

## Affected packages or components

`kernel/webhook/verifier.go` (`Verifier` interface, `Envelope` type, `HMACVerifier`,
`FakeVerifier`); `kernel/webhook/service.go` (`HandleInbound`); whatever documentation location
currently covers the webhook module's provider-integration contract (T4's target, to be identified
at implementation time).

## Compatibility considerations

**This is a breaking interface change.** PLAN's own risk column for T1 states plainly: "Breaking
interface change." Any type implementing the current `Verifier` interface (`Verify(...) error`)
will fail to compile against the new interface (`Verify(...) (Envelope, error)`) until updated. Per
PLAN's own wowsociety-impact note, immediately following the T1-T4 table: "**wowsociety impact —
SEC-03: Not affected.** Zero `kernel/webhook` import, zero custom `Verifier` implementation anywhere
in wowsociety." So while the interface change **is** breaking in the general sense, wowsociety's
current exposure to that break is zero, per PLAN's own citation. This claim should be re-confirmed
fresh (a current grep against wowsociety, not the cited snapshot alone) at this story's own
execution time, consistent with RISK-W03-006's mitigation — a breaking interface change is a safe
failure mode (a compile error, not a silent runtime regression) even if the re-confirmation finds an
unexpected consumer, but coordination would then be required before merge.

## Security considerations

This story directly closes the successful-signature replay/dedup gap described in "Problem
statement." The `Envelope`'s authenticated-fields-only guarantee (T2's acceptance criterion:
"Envelope never surfaces caller-supplied fields") is the core security property this story
establishes — any future `Verifier` implementation must uphold this guarantee for the security
benefit to hold, which is why T4's contract document is a required, not optional, part of this
story's scope: without a stated contract, a future provider-specific verifier could silently
reintroduce the gap by populating `Envelope` fields from unauthenticated data.

## Performance considerations

The interface change itself is not expected to add meaningful overhead — `Envelope` construction is
a small struct allocation per verified request, not a new I/O operation. No performance regression
is anticipated; not separately benchmarked as part of this story's acceptance criteria.

## Observability considerations

Not mandated by this story's acceptance criteria. A log or metric recording envelope-synthesis
outcomes is a reasonable implementation-time addition, not required scope.

## Migration considerations

None — no schema or data migration; this is a pure Go interface and application-logic change.

## Documentation requirements

T4's provider-verifier contract document is itself a required scope item, not merely a
"nice-to-have" — see "Scope" and "Security considerations" above.

## Acceptance criteria

- **AC-W03-E03-S001-01**: The `Verifier` interface returns `(Envelope, error)`;
  `Envelope{CanonicalBody, EventID, OccurredAt, SignatureVersion, KeyID}` is defined; `HMACVerifier`
  and `FakeVerifier` both compile against and satisfy the new interface; unit tests for both pass.
- **AC-W03-E03-S001-02**: `Envelope` never surfaces caller-supplied fields; a test proves
  `OccurredAt` is immune to a manipulated `in.Timestamp`.
- **AC-W03-E03-S001-03**: No security decision in `HandleInbound` reads a raw `InboundIn` field; the
  adversarial tamper matrix (body, timestamp, event-ID, key-ID, signature-version, each
  independently manipulated) passes.
- **AC-W03-E03-S001-04**: A provider-verifier contract document exists with a reference example.

## Required artifacts

- `Envelope` type definition and the changed `Verifier` interface.
- Updated `HMACVerifier` and `FakeVerifier` implementations.
- Rewired `HandleInbound`.
- Provider-verifier contract document.
See `artifacts/index.md`.

## Required evidence

- Unit test output for both `Verifier` implementations against the new interface.
- `OccurredAt`-immune-to-manipulated-timestamp test output.
- Adversarial tamper-matrix test output (5 independently manipulated fields).
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`:
`story.md` and `plan.md` complete, acceptance criteria numbered and measurable, dependencies (none)
recorded, owner/reviewer assignment pending.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; all four acceptance
criteria verified with evidence in `evidence/index.md`; `closure.md` completed; independent review
passed per mandate §14, with specific confirmation the breaking interface change is documented with
explicit compatibility notes.

## Risks

RISK-W03-006 (a future or undiscovered custom `Verifier` implementation breaking silently at
compile time only) — see epic-level `risks.md` for full detail and mitigation/contingency.

## Residual-risk expectations

Once the fresh re-confirmation of "zero custom `Verifier` implementation" (mitigating
RISK-W03-006) is executed and the tamper matrix genuinely proves all five manipulated-field cases
inert, no residual risk beyond the structural possibility of an undiscovered consumer requiring
coordination is expected to remain open against this story.

## Plan

See `plan.md`.
