---
id: W03-E03-ACCEPTANCE
type: epic-acceptance
epic: W03-E03
wave: W03
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W03-E03 — Epic-level acceptance

Reproduces and elaborates `epic.md`'s "Epic acceptance criteria" as a standalone,
independently-referenceable record, consistent with the wave-level `../../acceptance.md` pattern
(AC-W03-06 there maps onto this epic).

## AC-W03-E03-01 — Verifier interface change

`Verifier.Verify` returns `(Envelope, error)`; `Envelope{CanonicalBody, EventID, OccurredAt,
SignatureVersion, KeyID}` is defined; both `HMACVerifier` and `FakeVerifier` compile against and
satisfy the new interface; unit tests for both implementations pass.

## AC-W03-E03-02 — Authenticated envelope synthesis

`HMACVerifier` synthesizes `EventID`/`OccurredAt` from the authenticated body/receipt time only,
never from caller-supplied fields; a test proves `OccurredAt` is immune to a manipulated
`in.Timestamp`.

## AC-W03-E03-03 — HandleInbound rewired

No security decision in `HandleInbound` reads a raw `InboundIn` field; the adversarial tamper
matrix (body, timestamp, event-ID, key-ID, signature-version each independently manipulated) is
provably inert against the rewired path.

## AC-W03-E03-04 — Provider-verifier contract documented

A contract document exists describing what a provider-specific `Verifier` implementation must
guarantee, with a reference example.

## AC-W03-E03-05 — Independent review passed

W03-E03-S001 has passed independent review per mandate §14, with specific confirmation that the
breaking `Verifier` interface change is documented with explicit compatibility notes (no wowsociety
or other consumer silently broken without documentation).

## Acceptance authority

Product-security lead (PLAN §5.2's stated accountable role for PF-SEC).
