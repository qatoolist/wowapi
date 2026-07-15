---
id: PLAN-W03-E03-S001
type: plan
parent_story: W03-E03-S001
status: ready
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Plan — W03-E03-S001

Per mandate §8.5. Confirmed facts, planned changes, and implementation assumptions are distinguished
explicitly below.

## Proposed architecture

No new package. The `Verifier` interface in `kernel/webhook/verifier.go` changes its `Verify` method
signature to return `(Envelope, error)` instead of a bare `error`. A new `Envelope` type is defined,
carrying only fields the verifier itself derives from authenticated data. `HMACVerifier` and
`FakeVerifier` are both updated to satisfy the new interface. `HandleInbound`
(`kernel/webhook/service.go`) is rewired to read every replay-window and dedup decision from
`Envelope` rather than from the caller-supplied `InboundIn`.

## Implementation strategy

1. Re-confirm, at this story's actual start commit, the exact current state of
   `Verifier.Verify` (`kernel/webhook/verifier.go:32`), `HMACVerifier`, `FakeVerifier`, and
   `HandleInbound` (`kernel/webhook/service.go:22-114`) — re-read, don't trust PLAN's cited line
   numbers blindly (they may have drifted).
2. Fresh-confirm PLAN's own "zero custom `Verifier` implementation anywhere in wowsociety" claim via
   a current grep against wowsociety, per RISK-W03-006's mitigation — do not rely on the cited
   snapshot alone.
3. Define `Envelope{CanonicalBody, EventID, OccurredAt, SignatureVersion, KeyID}`, following existing
   `kernel/webhook` naming/type idioms.
4. Change the `Verifier` interface's `Verify` method to return `(Envelope, error)`.
5. Update `HMACVerifier` to synthesize `EventID` and `OccurredAt` from the authenticated body and/or
   receipt time only — never from caller-supplied request fields (`InboundIn.Timestamp`,
   `InboundIn.ExternalEventID`). Document this synthesis approach as unsuitable for
   timestamped-provider protocols requiring provider-asserted timestamps.
6. Update `FakeVerifier` to satisfy the new interface, preserving its existing test-double behavior
   as closely as possible while returning a valid `Envelope`.
7. Rewire `HandleInbound` so every replay-window and dedup decision is read exclusively from
   `Envelope`, never from `InboundIn` directly.
8. Write the test suite: unit tests for both `Verifier` implementations against the new interface; a
   test proving `OccurredAt` is immune to a manipulated `in.Timestamp`; the adversarial tamper matrix
   (body, timestamp, event-ID, key-ID, signature-version, each independently manipulated).
9. Author the provider-verifier contract document with a reference example, describing what any
   `Verifier` implementation must guarantee (in particular, the authenticated-fields-only property
   `Envelope` establishes).

## Expected package or module changes

`kernel/webhook` (`verifier.go`: `Verifier` interface, `Envelope` type, `HMACVerifier`,
`FakeVerifier`; `service.go`: `HandleInbound`); whatever documentation location currently covers the
webhook module's provider-integration contract (exact location TBD at implementation time).

## Expected file changes where determinable

- `kernel/webhook/verifier.go:32` — `Verifier.Verify`'s signature change; `Envelope` type
  definition; `HMACVerifier` and `FakeVerifier` updates.
- `kernel/webhook/service.go:22-114` — `HandleInbound`'s rewire to source replay-window/dedup
  exclusively from `Envelope`.
- A new or extended documentation file for the provider-verifier contract (exact path TBD).

## Contracts and interfaces

`Verifier.Verify(...) (Envelope, error)` replaces `Verifier.Verify(...) error` — a breaking interface
change, per PLAN's own T1 risk note. `Envelope{CanonicalBody, EventID, OccurredAt, SignatureVersion,
KeyID}` is a new, additive type. The provider-verifier contract document (T4) formalizes the
guarantee that `Envelope`'s fields are always derived from authenticated data, never from
caller-supplied request fields — any future `Verifier` implementation must uphold this.

## Data structures

`Envelope`'s exact field types and naming conventions follow existing `kernel/webhook` idioms,
confirmed by reading `kernel/webhook/verifier.go` directly at implementation time, not invented here
independent of the package's existing style.

## APIs

No public HTTP API surface is added or changed by this story — this is an internal Go interface and
application-logic change within `kernel/webhook`.

## Configuration changes

None anticipated.

## Persistence changes

None — no schema or data migration.

## Migration strategy

Not applicable — pure Go interface and application-logic change, no schema involved.

## Concurrency implications

None material beyond what already exists in the webhook inbound path — `Envelope` construction is a
per-request struct allocation, not a new concurrency-sensitive operation.

## Error-handling strategy

`Verifier.Verify` continues to return a non-nil `error` on verification failure, now alongside a
(possibly zero-value) `Envelope`; callers must continue to check `error` first, per Go convention, and
must not read `Envelope` fields when `error != nil`. This contract is stated explicitly in the T4
provider-verifier document so future implementations do not populate a partial or misleading
`Envelope` on failure.

## Security controls

`Envelope`'s authenticated-fields-only guarantee is the central security control this story
establishes: `HandleInbound` no longer trusts `InboundIn.Timestamp`/`InboundIn.ExternalEventID` for
any replay-window or dedup decision. The T4 contract document is itself a required security control,
not merely documentation — without a stated contract, a future provider-specific verifier could
silently reintroduce the gap.

## Observability changes

Not mandated by this story's acceptance criteria. A log or metric recording envelope-synthesis
outcomes is a reasonable implementation-time addition, not required scope.

## Testing strategy

- T1: unit tests for both `HMACVerifier` and `FakeVerifier` against the new `(Envelope, error)`
  interface.
- T2: a dedicated test proving `OccurredAt` is immune to a manipulated `in.Timestamp` — i.e. mutating
  the caller-supplied timestamp does not change the `Envelope`'s `OccurredAt`.
- T3: the adversarial tamper matrix — body, timestamp, event-ID, key-ID, and signature-version, each
  independently manipulated, asserting `HandleInbound`'s replay-window/dedup decision is unaffected
  by any field not covered by the signature.
- T4: no dedicated test (documentation task), but the contract document's reference example should
  be validated for accuracy against the actual `HMACVerifier` implementation.
- Fresh re-confirmation (per mandate's fail-first convention) that "zero custom `Verifier`
  implementation anywhere in wowsociety" still holds at this story's own execution commit.

## Regression strategy

The adversarial tamper matrix, run in CI, is the regression guard against a future change
reintroducing a caller-supplied-field dependency into `HandleInbound`'s security decisions.

## Compatibility strategy

This is a breaking interface change (PLAN's own T1 risk note, verbatim: "Breaking interface
change"). Per PLAN's own wowsociety-impact note, wowsociety has zero `kernel/webhook` import and zero
custom `Verifier` implementation today — re-confirmed fresh at this story's own execution time
(step 2 above), not merely trusted from the cited snapshot. A breaking interface change is a safe
failure mode (a compile error, not a silent runtime regression) even if the re-confirmation finds an
unexpected consumer, but coordination would then be required before merge — see RISK-W03-006.

## Rollout strategy

Single story, all four tasks land together since T2/T3 both depend on T1's interface change and T4
documents the resulting contract.

## Rollback strategy

The interface change is a single coordinated commit across `verifier.go` and `service.go`; if a
regression is found post-merge, reverting this story's commit(s) restores the prior `error`-only
interface and `InboundIn`-driven behavior cleanly, since no schema or persistent state is touched.

## Implementation sequence

T1 (interface change + both implementations) first, since T2 and T3 both build on it. T2
(`HMACVerifier` synthesis) and T3 (`HandleInbound` rewire) can proceed once T1 lands — PLAN's own
Depends-on column lists T2 as depending on T1, and T3 as depending on "T1, T2." T4 (contract document)
is written last, once the actual implementation shape of T1-T3 is final, so the document accurately
describes what was actually built.

## Task breakdown

- **W03-E03-S001-T001** — `Verifier` interface change to `(Envelope, error)`; `Envelope` type;
  `HMACVerifier`/`FakeVerifier` updates (SEC-03 T1).
- **W03-E03-S001-T002** — `HMACVerifier`'s `EventID`/`OccurredAt` synthesis from authenticated data
  only (SEC-03 T2).
- **W03-E03-S001-T003** — `HandleInbound` rewire to source replay-window/dedup exclusively from
  `Envelope` (SEC-03 T3).
- **W03-E03-S001-T004** — Provider-verifier contract document with reference example (SEC-03 T4).
- **W03-E03-S001-T005** — Independent review (mandate §14), with specific confirmation the breaking
  interface change is documented with explicit compatibility notes.

## Expected artifacts

`Envelope` type definition and the changed `Verifier` interface; updated `HMACVerifier` and
`FakeVerifier` implementations; rewired `HandleInbound`; the provider-verifier contract document.

## Expected evidence

Unit test output for both `Verifier` implementations; the `OccurredAt`-immune-to-manipulated-
timestamp test output; the adversarial tamper-matrix test output (5 independently manipulated
fields); the fresh wowsociety-consumer re-confirmation result.

## Unresolved questions

- The exact `Envelope` struct's field types and naming conventions — to be confirmed by reading
  `kernel/webhook/verifier.go` directly at implementation time, not invented here.
- Whether any provider-specific `Verifier` implementation beyond `HMACVerifier`/`FakeVerifier`
  already exists in wowapi itself — to be confirmed fresh at implementation time, not assumed
  exhaustive from PLAN's own citation.
- The exact target documentation location for T4's contract document — to be identified at
  implementation time.
- Whether the fresh wowsociety-consumer re-confirmation (step 2) finds any drift from PLAN's cited
  "zero" claim — genuinely unknown until re-run, per RISK-W03-006.

## Approval conditions

This plan is approved for implementation once: (a) the unresolved questions above are answered by a
first re-read of `kernel/webhook/verifier.go` and `service.go` at story start; (b) the fresh
wowsociety-consumer re-confirmation has run and its result is known; and (c) the owner and reviewer
are assigned.
