---
id: W03-E03-S001-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W03-E03-S001
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W03-E03-S001 — Artifacts index

Per mandate §9.2. Structure adaptation per `governance/naming-conventions.md` "Adaptation 2":
lifecycle subdirectories are created on first real content, not pre-populated empty. All entries
below are `not yet produced`.

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Path | Status |
|---|---|---|---|---|---|---|---|---|
| ART-W03-E03-S001-001 | `Envelope` type + changed `Verifier` interface | interface / source-code change | implementation | `Envelope{CanonicalBody, EventID, OccurredAt, SignatureVersion, KeyID}`; `Verify` returns `(Envelope, error)` | SEC-03 | W03-E03-S001-T001 | `foundation/webhook/verifier.go` | produced |
| ART-W03-E03-S001-002 | Updated `HMACVerifier`/`FakeVerifier` implementations | source-code change | implementation | Both implementations satisfy the new interface | SEC-03 | W03-E03-S001-T001 | `foundation/webhook/verifier.go` | produced |
| ART-W03-E03-S001-003 | `HMACVerifier` authenticated-data synthesis | source-code change | implementation | `EventID`/`OccurredAt` synthesized from authenticated body/receipt time only | SEC-03 | W03-E03-S001-T002 | `foundation/webhook/verifier.go` | produced |
| ART-W03-E03-S001-004 | Rewired `HandleInbound` | source-code change | implementation | Replay-window/dedup decisions sourced exclusively from `Envelope` | SEC-03 | W03-E03-S001-T003 | `foundation/webhook/service.go` | produced |
| ART-W03-E03-S001-005 | Provider-verifier contract document | documentation | post-implementation | Contract any `Verifier` implementation must guarantee, with a reference example | SEC-03 | W03-E03-S001-T004 | `artifacts/provider-verifier-contract.md` | produced |
