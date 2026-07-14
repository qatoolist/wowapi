---
id: PLAN-W04-E04-S002
type: plan
parent_story: W04-E04-S002
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Plan — W04-E04-S002

Per mandate §8.5. Confirmed facts, planned changes, and implementation assumptions are distinguished
explicitly below; this plan does not invent precise code changes where the repository does not yet
provide enough information — PLAN's own risk notes for W6-T2 ("Genuinely new subsystem —
vendor/design decision needed") and W6-T3 ("New encryption-key-management dependency") both confirm
open design questions this plan records rather than pre-answers.

## Proposed architecture

Four largely-independent additions layered onto `kernel/audit` and `kernel/retention`: (1) an
external anchoring mechanism periodically publishing the audit chain's head hash to an
external, tamper-evident target; (2) a DSR export artifact writer replacing `retention/engine.go`'s
in-memory map with a durable, encrypted, checksummed file; (3) a central legal-hold enforcement
wrapper interposed between the DSR/retention orchestrator and every registered `Dispose`/`Erase`
callback; (4) explicit per-class status reporting integrated into the DSR result set, coordinated
with the artifact's own manifest shape from (2).

## Implementation strategy

1. Re-read `retention/engine.go` and the current `Dispose`/`Erase` callback registration mechanism at
   this story's actual start commit to confirm the current-state assessment holds.
2. **W6-T2**: Draft external-anchoring mechanism options (e.g. a public timestamping service, a
   separate append-only log, a third-party notarization service) with trade-offs; select one and
   document the rationale, given PLAN's own framing that this requires a genuine vendor/design
   decision.
3. Implement the anchoring mechanism: periodically publish the chain head externally; implement
   detection logic that cross-checks the local chain against the external anchor.
4. Write the anchor-then-tamper detection test: anchor the chain, tamper with a local row, confirm
   detection via the anchor.
5. **W6-T3**: Draft the DSR export artifact's exact format (manifest fields, per-class result schema,
   checksum algorithm, expiry semantics, access-policy model, download-audit schema) and the
   encryption-key-management design (custody, rotation, recovery), given PLAN's own framing that key
   management is a new dependency requiring explicit design.
6. Implement the artifact writer: replace `retention/engine.go`'s bare in-memory map return with a
   write path that produces the encrypted, checksummed artifact and gates export completion on write
   success.
7. Write the export-completion/checksum-verification test.
8. **W6-T4**: Enumerate every currently-registered `RecordClass` and its `Dispose`/`Erase` callback in
   both wowapi and wowsociety, per PLAN's own risk note ("enumerate every registered `RecordClass` in
   both repos first") — this enumeration must complete before the wrapper implementation begins, not
   in parallel with it.
9. Implement the central legal-hold enforcement wrapper, interposed so every registered callback
   passes through it regardless of the callback's own internal hold-check correctness.
10. Write the negative test: register a deliberately non-compliant callback (one that does not itself
    check for a hold) and confirm the wrapper still blocks it.
11. **W6-T5**: Coordinate with W6-T3's manifest shape (step 5) to add explicit per-class status
    reporting — every registered class appears in the DSR result set with a status (exported, erased,
    not-applicable, or partial), never a silent omission.
12. Write the explicit-status test.
13. Document all four mechanisms.

## Expected package or module changes

`kernel/audit` (external anchoring, extending the existing `Anchor`/`CheckAnchor` surface);
`kernel/retention` (`retention/engine.go`'s DSR export path, replaced with the artifact writer); the
`Dispose`/`Erase` callback registration mechanism (wrapped with the legal-hold enforcement layer);
new encryption-key-management code (exact package location TBD).

## Expected file changes where determinable

- `kernel/audit` — a new file or extension implementing the external anchoring mechanism (exact path
  TBD, dependent on the vendor/protocol decision in step 2).
- `kernel/retention/engine.go` — replaced/extended DSR export path producing the artifact instead of
  the bare in-memory map.
- The `Dispose`/`Erase` callback registration mechanism's own source file(s) — wrapped with the
  central legal-hold enforcement layer (exact location TBD, expected near the existing registration
  code).
- New test files for the anchor-then-tamper test, the export-completion/checksum test, the
  legal-hold negative test, and the explicit-status test.

## Contracts and interfaces

The `DisposeFunc`/`EraseFunc` contract changes: every registered callback now passes through the
central legal-hold wrapper. This is confirmed by the source as a **breaking change** (PLAN W6-T4's
own risk note) — the exact new contract shape (wrapper-injected parameter, a required registration-
time declaration, or a fully transparent interposition requiring no callback-code change) is TBD at
implementation time and must be chosen to minimize unnecessary breakage while still guaranteeing the
negative-test property (a non-compliant callback cannot bypass the wrapper).

## Data structures

The DSR export artifact's manifest structure (per-class results, checksum, expiry, access policy,
download-audit records) is a new data structure, exact shape TBD per step 5 above. The external
anchor's own record structure (whatever is published externally and however the local chain
cross-references it) is likewise a new data structure, exact shape TBD per step 2.

## APIs

None affected at the HTTP/API layer directly, though the DSR export flow's completion semantics
change (gated on artifact-write success) — any API surface that reports DSR export status must
reflect this new gating, to be confirmed at implementation time.

## Configuration changes

The external anchoring mechanism likely requires new configuration (target endpoint, credentials, or
similar, depending on the vendor/protocol chosen in step 2). The DSR export artifact's encryption-key
source likely requires new configuration (key-management-service reference, or equivalent). Exact
configuration surface TBD at implementation time, dependent on both open design decisions.

## Persistence changes

Possible new tables/columns for the DSR export artifact registry and/or the legal-hold wrapper's own
audit trail — exact schema TBD at implementation time. Any such change touching a live-production
table is expected to ship through W02-E01's protocol, consistent with S001's precedent, though this
story has no confirmed dependency on W02-E01 (to be resolved at implementation time if a schema
change proves necessary).

## Migration strategy

See "Persistence changes" above — deferred to implementation time pending the exact schema needs of
the artifact registry and legal-hold audit trail.

## Concurrency implications

The central legal-hold wrapper must correctly serialize or otherwise safely handle concurrent
`Dispose`/`Erase` requests against the same record without race conditions in the hold-check itself
— an implementation-time concern this plan flags but does not pre-resolve, since the exact concurrency
model depends on the wrapper's chosen implementation (step 9).

## Error-handling strategy

A DSR export must not report completion if the artifact write fails partway — the export-completion
gate (AC-W04-E04-S002-02) requires this explicitly. A `Dispose`/`Erase` callback blocked by the
legal-hold wrapper must produce a clear, hold-specific error (not a generic failure), distinguishing
"blocked by legal hold" from any other failure mode.

## Security controls

W6-T2's external anchoring, W6-T3's DSR export encryption, and W6-T4's central legal-hold wrapper are
each themselves the required security control for their respective task — not supplementary
hardening. The `RecordClass` enumeration step (W6-T4) is itself a required security-review-adjacent
control: it exists specifically to prevent the wrapper from silently missing a currently-registered
callback.

## Observability changes

Anchor-attempt success/failure, DSR export artifact-write success/failure, and legal-hold-wrapper
block events should each be logged, per `story.md` "Observability considerations" — implementation-
time additions, not separately mandated beyond each task's own acceptance criterion.

## Testing strategy

- W6-T2: anchor-then-tamper detection test — anchor the chain, tamper locally, confirm detection via
  the anchor.
- W6-T3: export-completion/checksum-verification test — confirm export completes only after
  successful artifact write, and the checksum verifies against the written artifact.
- W6-T4: negative test — a deliberately non-compliant callback (no internal hold check) is still
  blocked by the wrapper.
- W6-T5: explicit-status test — every registered class appears in the DSR result set with a status,
  none silently omitted.

## Regression strategy

Each of the four tests above becomes the regression guard for its own task's scope: any future change
that breaks anchor detection, artifact-write gating, legal-hold enforcement, or explicit-status
reporting would be caught by the corresponding test failing.

## Compatibility strategy

W6-T4's breaking `DisposeFunc`/`EraseFunc` contract change requires the enumeration step (step 8)
completing, and its results reviewed, before any callback code is required to change — this story's
own compatibility strategy is that enumeration precedes implementation, not the reverse. If the
enumeration surfaces more registered classes than anticipated, the task's own scope may need
revisiting before the wrapper lands — recorded as a plan-time constraint, not a silent assumption.

## Rollout strategy

Single story, landed as its own reviewable unit, sequenced after W04-E04-S001's acceptance (this
story's own dependency). W6-T2 through W6-T5 have no forced internal order beyond W6-T5's
coordination with W6-T3's manifest shape (step 11) and W6-T4's enumeration-before-implementation
constraint (step 8) — the four may otherwise be implemented in parallel within this story if the
owning implementer judges that efficient.

## Rollback strategy

W6-T4's central legal-hold wrapper, once landed, should be revertible without requiring a schema
rollback if it produces false positives blocking a legitimate callback — the wrapper is additive
interposition, not a destructive schema change, so a code-level revert is expected to be sufficient.
W6-T3's DSR export artifact, once artifacts have been produced under its encryption scheme, is harder
to revert without a compatibility plan for already-produced artifacts — this is recorded as a
rollback constraint to resolve at implementation time, not invented as a specific procedure here.

## Implementation sequence

As listed under "Implementation strategy" above (steps 1–13). Step 8 (the `RecordClass` enumeration)
must complete before step 9 (the wrapper implementation) begins, per PLAN's own risk note. Step 11
(explicit-status coordination) depends on step 5's manifest-shape design being settled first.

## Task breakdown

- **W04-E04-S002-T001** — External anchor verification for the audit chain (W6-T2; steps 2–4 above).
- **W04-E04-S002-T002** — Encrypted immutable DSR export artifact (W6-T3; steps 5–7 above).
- **W04-E04-S002-T003** — Central legal-hold enforcement wrapper (W6-T4; steps 8–10 above).
- **W04-E04-S002-T004** — Explicit partial/not-applicable per-class DSR status (W6-T5; steps 11–12
  above).
- **W04-E04-S002-T005** — Independent review (per mandate §14, scoped to this story).

## Expected artifacts

The external anchor mechanism; the DSR export artifact writer; the central legal-hold wrapper; the
explicit per-class status reporting mechanism; documentation of all four.

## Expected evidence

Anchor-then-tamper detection test output; export-completion/checksum-verification test output;
legal-hold negative test output; explicit-status test output; the `RecordClass` callback enumeration
record.

## Unresolved questions

- The external anchoring mechanism's exact vendor/protocol (W6-T2) — a genuinely open design question
  per PLAN's own risk note, to be decided at implementation time.
- The DSR export artifact's exact encryption-key-management scheme (W6-T3) — likewise open, to be
  decided at implementation time.
- The exact `DisposeFunc`/`EraseFunc` contract shape after the legal-hold wrapper lands (W6-T4) — TBD
  pending the enumeration step's results.
- The exact schema (if any) for the DSR export artifact registry and legal-hold audit trail — TBD
  pending each mechanism's own implementation-time design.

## Approval conditions

This plan is approved for implementation once: (a) W04-E04-S001 has reached `accepted` (this story's
upstream dependency), (b) the unresolved questions above — most centrally, the anchoring vendor/
protocol and the encryption-key-management scheme — are answered and documented, and (c) the owner
and reviewer are assigned.
