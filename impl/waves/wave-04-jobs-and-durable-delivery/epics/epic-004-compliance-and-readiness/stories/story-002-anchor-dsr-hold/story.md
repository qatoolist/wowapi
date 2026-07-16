---
id: W04-E04-S002
type: story
title: External anchoring, DSR export artifact, central legal-hold, and explicit per-class status
status: accepted
wave: W04
epic: W04-E04
owner: W03-E02-E03-E04-E05-Rerun
reviewer: unassigned
priority: P1
created_at: 2026-07-12
updated_at: 2026-07-13
source_requirements:
  - DATA-08
depends_on:
  - W04-E04-S001
blocks: []
acceptance_criteria:
  - AC-W04-E04-S002-01
  - AC-W04-E04-S002-02
  - AC-W04-E04-S002-03
  - AC-W04-E04-S002-04
artifacts: []
evidence: []
decisions: []
risks:
  - RISK-W04-E04-001
  - RISK-W04-E04-002
---

# W04-E04-S002 — External anchoring, DSR export artifact, central legal-hold, and explicit per-class status

## Story ID

W04-E04-S002

## Title

External anchoring, DSR export artifact, central legal-hold, and explicit per-class status

## Objective

Build external anchor verification for the audit chain so tampering is detectable even if the local
`head_hash` were compromised (W6-T2); persist DSR exports as encrypted immutable artifacts with a
manifest, per-class results, checksum, expiry, access policy, and download audit, replacing
`retention/engine.go`'s bare in-memory map return (W6-T3); build a central legal-hold enforcement
wrapper every `Dispose`/`Erase` callback must pass through, replacing today's per-callback
responsibility (W6-T4); and produce explicit partial/not-applicable results for record classes
without export/erase callbacks so the result set never silently omits a registered class (W6-T5).

## Value to the framework

This story completes DATA-08's Wave-6 compliance-evidence scope, building directly on S001's widened,
version-discriminated audit chain. Without external anchoring (W6-T2), the tamper-evidence chain
S001 hardens is still only as trustworthy as the local database it lives in — an attacker with
sufficient database access could still compromise the chain's own `head_hash`. Without an encrypted,
durable DSR export artifact (W6-T3), `retention/engine.go`'s current bare in-memory map return means
a completed DSR request has no durable, auditable proof of what was actually exported. Without a
central legal-hold wrapper (W6-T4), compliance depends on every individual `Dispose`/`Erase`
callback correctly implementing its own hold check — a single non-compliant callback silently
defeats the guarantee. Without explicit per-class status (W6-T5), a DSR result that omits a
registered record class is indistinguishable from one that correctly found nothing to export for
that class — a silent gap masquerading as a clean result.

## Problem statement

PLAN DATA-08's own task rows state the target state directly: W6-T2 — "Chain-head periodically
anchored externally; tamper detectable even if local `head_hash` were compromised." W6-T3 —
"Replaces `retention/engine.go`'s bare in-memory map return" with "an encrypted immutable artifact
with manifest, per-class results, checksum, expiry, access policy, download audit." W6-T4 — "Central
legal-hold enforcement wrapper every `Dispose`/`Erase` callback must pass through, replacing today's
per-callback responsibility," with the explicit acceptance bar: "Negative test: a deliberately
non-compliant callback is still blocked by the framework wrapper." W6-T5 — "Explicit partial/not-
applicable results for record classes without export/erase callbacks," with the acceptance bar:
"Result set explicitly lists every registered class with a status, never a silent omission."

## Source requirements

DATA-08 (Wave-6 tasks W6-T2, W6-T3, W6-T4, W6-T5).

## Current-state assessment

Per PLAN DATA-08's own evidence for these four tasks (to be re-confirmed at this story's actual start
commit): no external anchor mechanism exists for the audit chain beyond the existing `Anchor`/
`CheckAnchor` tail-truncation guard (which S001 leaves largely as-is beyond potential version-
awareness); `retention/engine.go` returns a bare in-memory map with no durable artifact, no
encryption, no manifest, no checksum, no expiry, no access policy, and no download audit; legal-hold
enforcement is per-callback today, with no central wrapper every `Dispose`/`Erase` callback must pass
through; DSR results have no confirmed explicit-status mechanism for record classes lacking an
export/erase callback. This story's own re-confirmation step is to read `retention/engine.go` and
the current `Dispose`/`Erase` callback registration mechanism at this story's actual start commit
before building the wrapper and artifact-writer described below.

## Desired state

The audit chain's head is periodically anchored externally (mechanism TBD per `plan.md`'s
"Unresolved questions" — a genuinely new subsystem per PLAN's own risk note, requiring a vendor/
design decision), with tamper detectable via the anchor even if the local `head_hash` were
compromised. DSR exports complete only after successfully writing an encrypted, immutable artifact
(manifest, per-class results, checksum, expiry, access policy, download audit) — `retention/
engine.go`'s bare in-memory map return is fully replaced. Every registered `Dispose`/`Erase` callback
passes through a central legal-hold enforcement wrapper; a deliberately non-compliant callback is
still blocked. The DSR result set explicitly lists every registered record class with a status
(exported, erased, not-applicable, or partial), never a silent omission.

## Scope

- External anchor verification for the audit chain (W6-T2): the anchoring mechanism's own design
  (vendor/protocol choice), the anchor-then-tamper detection test.
- DSR export as an encrypted immutable artifact (W6-T3): manifest, per-class results, checksum,
  expiry, access policy, download audit; the test confirming export completes only after artifact
  write succeeds and the checksum verifies; access-gated download audited.
- Central legal-hold enforcement wrapper (W6-T4): the wrapper every `Dispose`/`Erase` callback must
  pass through; the enumeration of every currently-registered `RecordClass` in both wowapi and
  wowsociety, per PLAN's own risk note, before implementing the wrapper; the negative test proving a
  deliberately non-compliant callback is still blocked.
- Explicit partial/not-applicable status reporting (W6-T5): coordinated with W6-T3's manifest shape
  per PLAN's own dependency note, so the result set explicitly lists every registered class with a
  status.

## Out of scope

- **DATA-08 W6-T1** (the audit hash-chain widening and `hash_version` migration) — W04-E04-S001's
  scope. This story's external anchoring builds on S001's widened, versioned hash chain but does not
  itself modify `chainHash` or `Verify`'s version-branch logic.
- **DATA-08 W0-T1, W0-T2** — already executed elsewhere; not re-implemented or re-verified here.
- **DX-07's readiness/config diagnostics scope** — W04-E04-S003's scope, unrelated to this story.
- **PROD-05** — the wowsociety-side staging audit re-verification drill; product-level, not
  implemented here (it concerns S001's hash-widening migration, not this story's own scope, but is
  noted here for completeness given both stories share DATA-08 as their source requirement).
- **Choosing the specific external anchoring vendor/protocol** beyond identifying it as a required
  design decision — recorded as an unresolved question in `plan.md`, per PLAN W6-T2's own risk note:
  "Genuinely new subsystem — vendor/design decision needed."
- **The exact encryption-key-management scheme for the DSR export artifact** beyond identifying it as
  a required design decision — recorded as an unresolved question in `plan.md`, per PLAN W6-T3's own
  risk note: "New encryption-key-management dependency."

## Assumptions

- The external anchoring mechanism's exact vendor/protocol is not yet determined by the source —
  PLAN W6-T2's own risk note confirms this is a genuinely open design question this story's plan must
  record, not invent.
- The DSR export artifact's exact encryption-key-management scheme is likewise not yet determined by
  the source, per PLAN W6-T3's own risk note — recorded as an unresolved question, not invented.
- W6-T4's `RecordClass` callback enumeration (both repos) is treated as a required precondition step,
  not an optional nicety, per PLAN's own risk note: "enumerate every registered `RecordClass` in both
  repos first."
- W6-T5's explicit-status work is coordinated with W6-T3's manifest shape, per PLAN's own dependency
  column for W6-T5 ("W6-T3, W6-T4") — this story treats the manifest shape (from W6-T3) as the
  natural home for per-class status reporting, to be confirmed at implementation time.

## Dependencies

Depends on W04-E04-S001 (DATA-08 W6-T1's widened, version-discriminated audit chain) — per this
epic's `dependencies.md`: "S002 ... depends on S001 (W6-T1's widened, versioned hash chain)... per
`wave.md`'s own dependency framing." No dependency on W02-E01 (this story's tasks do not touch the
migration-protocol surface S001 required). No downstream story within this epic depends on S002.

## Affected packages or components

`kernel/audit` (external anchor mechanism, extending `Anchor`/`CheckAnchor`); `kernel/retention`
(specifically `retention/engine.go`'s DSR export path, replaced with an artifact-writing
implementation); the `Dispose`/`Erase` callback registration mechanism (wrapped with the central
legal-hold enforcement layer); new encryption-key-management code for the DSR export artifact (exact
location TBD).

## Compatibility considerations

**W6-T4 is a breaking change to the `DisposeFunc`/`EraseFunc` contract.** Per PLAN's own risk note:
"Breaking change to the `DisposeFunc`/`EraseFunc` contract — enumerate every registered `RecordClass`
in both repos first." Every currently-registered callback (in both wowapi and any consuming product,
including wowsociety if it has registered any) must be enumerated and confirmed compatible with the
new wrapper-mediated contract before the wrapper lands — this is not an incidental implementation
detail, it is a required precondition step this story's plan and task must record explicitly. No
`kernel/attachment`/`kernel/notify`/`kernel/retention` usage was found in wowsociety per this wave's
`dependencies.md` wowsociety-impact note at the time PLAN's evidence was gathered, so W6-T3/T4/T5
land on wowsociety's future DSR roadmap rather than its current code — non-blocking for this story's
own closure, but the enumeration step must still be performed against wowapi's own registered classes
regardless.

## Security considerations

W6-T2's external anchoring exists specifically to protect against a local-database-compromise
scenario where an attacker could otherwise alter `head_hash` itself — this is the acceptance-defining
security property, not a supplementary hardening measure. W6-T3's encrypted DSR export artifact
introduces a new encryption-key-management dependency, per PLAN's own risk note — the key-custody,
rotation, and recovery design must be treated as a required security control, not an incidental
implementation detail. W6-T4's central legal-hold wrapper is itself a required security/compliance
control: PLAN's own acceptance criterion is explicit that "a deliberately non-compliant callback is
still blocked by the framework wrapper" — the wrapper must fail closed, not rely on callback
cooperation.

## Performance considerations

None separately identified beyond the DSR export artifact's own write/encryption cost, which is
inherent to the artifact-durability requirement itself (W6-T3), not a separate performance concern
this story must additionally address.

## Observability considerations

External anchor failures (W6-T2), DSR export artifact write failures (W6-T3), and legal-hold-wrapper
blocks of non-compliant callbacks (W6-T4) should each be observable — logged at minimum — so an
operator can distinguish a legitimate hold-block from a misconfigured callback, or a failed anchor
attempt from a successful one. Reasonable implementation-time additions given each task's own
acceptance bar, though not separately mandated by the source beyond the acceptance criteria
themselves.

## Migration considerations

W6-T3's DSR export artifact and W6-T4's legal-hold wrapper may require schema changes (e.g. an
artifact-registry table, an audit-download table) — exact schema TBD at implementation time. Any such
migration is expected to ship through W02-E01's online-migration protocol if it touches a
live-production table, consistent with this epic's own precedent in S001, though this story has no
confirmed dependency on W02-E01 the way S001 does — to be resolved at implementation time if a
schema change proves necessary.

## Documentation requirements

Document the external anchoring mechanism and its verification procedure; the DSR export artifact's
format (manifest, per-class results, checksum, expiry, access policy, download audit) and its
encryption-key-management scheme; the central legal-hold wrapper's contract and how a `Dispose`/
`Erase` callback registers against it; the explicit per-class status vocabulary (exported, erased,
not-applicable, partial).

## Acceptance criteria

- **AC-W04-E04-S002-01**: The audit chain's head is periodically anchored externally; a test that
  anchors, then tampers with the local chain, confirms the tampering is detectable via the anchor
  even where local `head_hash` alone would not reveal it.
- **AC-W04-E04-S002-02**: A DSR export completes only after successfully writing an encrypted,
  immutable artifact containing a manifest, per-class results, checksum, expiry, and access policy,
  with downloads audited; a test confirms export completion is gated on artifact-write success and
  the checksum verifies against the written artifact.
- **AC-W04-E04-S002-03**: Every registered `Dispose`/`Erase` callback passes through a central
  legal-hold enforcement wrapper; a negative test with a deliberately non-compliant callback confirms
  the callback is still blocked by the framework wrapper, not merely by its own (absent or incorrect)
  internal check.
- **AC-W04-E04-S002-04**: The DSR result set explicitly lists every registered record class with a
  status (never a silent omission), coordinated with W6-T3's manifest shape; every currently-
  registered `RecordClass` in both wowapi and wowsociety has been enumerated before the legal-hold
  wrapper (AC-03) was implemented, per PLAN W6-T4's own precondition.

## Required artifacts

- The external anchor mechanism and its verification code.
- The DSR export artifact writer (manifest, checksum, encryption, access policy, download-audit
  code).
- The central legal-hold enforcement wrapper.
- The explicit per-class status reporting mechanism.
- Documentation of all four mechanisms.
See `artifacts/index.md`.

## Required evidence

- Anchor-then-tamper detection test output.
- DSR export artifact-completion and checksum-verification test output.
- Central legal-hold negative test output (non-compliant callback blocked).
- Explicit-status test output (every registered class listed with a status).
- The `RecordClass` callback enumeration record (both repos), predating the legal-hold wrapper's
  implementation.
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`: `story.md`
and `plan.md` complete, acceptance criteria numbered and measurable, dependency on W04-E04-S001
recorded, owner/reviewer assignment pending, unresolved questions (external anchoring vendor/
protocol, DSR export encryption-key-management scheme) explicitly recorded rather than silently
assumed.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; all four acceptance
criteria verified with evidence in `evidence/index.md`; `closure.md` completed; independent review
passed per mandate §14, specifically confirming the `RecordClass` enumeration genuinely predates the
legal-hold wrapper's implementation and the DSR export artifact is genuinely gated on write success.

## Risks

RISK-W04-E04-001 (W6-T4's breaking change to the `DisposeFunc`/`EraseFunc` contract, requiring
complete callback enumeration before the wrapper lands) and RISK-W04-E04-002 (W6-T3's new
encryption-key-management dependency) — see epic-level `risks.md` for full detail and mitigation/
contingency.

## Residual-risk expectations

Once the `RecordClass` enumeration step (W6-T4) and the key-management design decision (W6-T3) are
both executed as planned, residual risk is expected to reduce to low/low-medium — this story's scope
is well-bounded by PLAN's own task rows, with no confirmed live-production breaking-change exposure
comparable to S001's.

## Plan

See `plan.md`.
