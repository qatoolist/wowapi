---
id: W03-E03
type: epic
title: Webhook authenticated replay
status: planned
wave: W03
owner: unassigned
reviewer: unassigned
priority: high
created_at: 2026-07-12
updated_at: 2026-07-12
source_requirements:
  - SEC-03
depends_on: []
stories:
  - W03-E03-S001
decisions: []
risks:
  - RISK-W03-006
---

# W03-E03 — Webhook authenticated replay

## Epic objective

Bind webhook replay-window and dedup decisions exclusively to provider-authenticated data by
changing the `Verifier` interface to return `(Envelope, error)` instead of a bare `error`, updating
both implementations (`HMACVerifier`, `FakeVerifier`), and rewiring `HandleInbound` so no security
decision reads a raw, caller-supplied `InboundIn` field. This is PLAN §5.2's SEC-03, T1 through T4
in full (single-story epic per `impl/analysis/wave-allocation-detail.md`).

## Problem being solved

`requirement-inventory.md` row SEC-03 records: "Webhook replay bound to authenticated data (T1–T4)"
— class IMPL, priority P1, disposition `planned`, target `W03-E03-S001`, notes "Breaking Verifier
interface." PLAN §5.2's own evidence: "`HMACVerifier.Verify` (`kernel/webhook/verifier.go:32`)
HMACs the body only, returns only `error`. `HandleInbound` (`service.go:22-114`) drives
replay-window/dedup off the caller-supplied, unauthenticated `InboundIn.Timestamp`/
`ExternalEventID`." PLAN notes a mitigating detail already present: "`service.go:60-63` already
forces `external_event_id = nil` on a *failed* signature — good defense-in-depth, doesn't address
the successful-signature gap." The gap this epic closes is specifically the successful-signature
case: even with a valid HMAC signature, the timestamp and event-ID used for replay-window and dedup
decisions are today read from the unauthenticated request body/headers rather than derived from
data the signature itself covers.

## Scope

- T1 — change the `Verifier` interface to return `(Envelope, error)`, where `Envelope{CanonicalBody,
  EventID, OccurredAt, SignatureVersion, KeyID}`; update `HMACVerifier` and `FakeVerifier`.
- T2 — `HMACVerifier` synthesizes `EventID`/`OccurredAt` from the authenticated body/receipt time
  only, documented as unsuitable for timestamped-provider protocols otherwise.
- T3 — rewire `HandleInbound` to source replay-window/dedup exclusively from `Envelope`.
- T4 — document the provider-verifier contract.

## Out of scope

- SEC-01, SEC-06, DATA-07, SEC-02 — separate epics in this wave with no shared file surface.
- Any change to `kernel/webhook`'s outbound delivery path (`deliverToEndpoint`) — that is PLAN
  DATA-03 T3's scope (W04-E02), a different finding entirely (remote I/O outside transactions, not
  replay/dedup authentication).
- Provider-specific verifier implementations beyond `HMACVerifier`/`FakeVerifier` — this epic
  changes the interface contract and its two existing implementations; a hypothetical third
  provider-specific verifier is out of scope unless one already exists in the codebase (to be
  confirmed at implementation time).

## Source requirements

SEC-03 (T1–T4).

## Architectural context

This epic makes the webhook inbound path's security decisions (replay-window checks, dedup checks)
strictly a function of data the cryptographic signature covers, closing the class of attack where a
validly-signed webhook body is replayed with a manipulated timestamp or event-ID to bypass
replay/dedup protection. The `Verifier` interface change (T1) is the load-bearing contract change:
today `Verify` returns only `error`; after this epic, it returns `(Envelope, error)`, and
`HandleInbound`'s security-relevant reads move from `InboundIn` (caller-supplied) to `Envelope`
(verifier-derived, authenticated). This is a breaking interface change to any custom `Verifier`
implementation, though PLAN's own wowsociety-impact note states there are none in wowsociety today.

The affected layers are `kernel/webhook/verifier.go` (`Verifier` interface, `HMACVerifier`),
`kernel/webhook/service.go` (`HandleInbound`), and — for T4 — whatever documentation currently
covers the webhook module's provider-integration contract.

## Included stories

- **W03-E03-S001 — webhook-authenticated-replay** (SEC-03 T1–T4, single story per
  `impl/analysis/wave-allocation-detail.md`: "S001 T1–T4 (breaking Verifier interface — compat notes
  mandatory)").

## Dependencies

No dependency on any other W03 epic — SEC-03 is architecturally independent, touching only
`kernel/webhook`. No cross-wave dependency beyond this wave's own entry criteria (W01 validation
seam, W02 DATA-09 protocol — though this epic's own schema footprint is zero, so DATA-09 is not
directly consumed here).

## Risks

RISK-W03-006 (a future or undiscovered custom `Verifier` implementation breaking silently at
compile time only) — inherited from `../../risks.md` (wave-level); see `risks.md` (epic-level) for
elaboration.

## Required decisions

None. The `Envelope` shape and the interface-change approach are specified directly by PLAN §5.2's
T1 row, not an open architecture question requiring a new ADR.

## Epic acceptance criteria

- **AC-W03-E03-01**: `Verifier` returns `(Envelope, error)`; `HMACVerifier` and `FakeVerifier` both
  compile against and satisfy the new interface; unit tests for both pass.
- **AC-W03-E03-02**: `Envelope` never surfaces caller-supplied fields; `OccurredAt` is immune to a
  manipulated `in.Timestamp`, proven by a dedicated test.
- **AC-W03-E03-03**: No security decision in `HandleInbound` reads a raw `InboundIn` field; the
  adversarial tamper matrix (body/timestamp/event-ID/key-ID/signature-version independently
  manipulated) passes.
- **AC-W03-E03-04**: A provider-verifier contract document exists with a reference example.
- **AC-W03-E03-05**: The story has passed independent review per mandate §14, with specific
  confirmation the breaking interface change is documented with explicit compatibility notes.

## Closure conditions

W03-E03-S001 reaches `accepted`; AC-W03-E03-01 through AC-W03-E03-05 above are all satisfied;
`closure-report.md` for this epic is completed with reviewer conclusion and acceptance date.
