---
id: W06-E02-ACCEPTANCE
type: epic-acceptance
epic: W06-E02
wave: W06
status: in-progress
created_at: 2026-07-12
updated_at: 2026-07-14
---

# W06-E02 — Epic-level acceptance

Reproduces and elaborates `epic.md`'s "Epic acceptance criteria" section as a standalone,
independently-referenceable acceptance record, consistent with the wave-level `../../acceptance.md`
pattern (AC-W06-03/04 there map onto this epic).

## AC-W06-E02-01 — OpenAPI merge complete-or-loud; AR-03 T2 duplicate closed

The merge struct covers every OpenAPI 3.1 top-level field and every `components.*` field with an
explicit per-field merge policy; the merged document validates against 3.1.1/2020-12; a seeded
breaking-API fixture fails the semantic-diff gate. AR-03's own target story (W05-E03) proceeds without
T2, per CONFLICT-01's resolution. Traces to W06-E02-S001.

## AC-W06-E02-02 — Compatibility gates buildable-now set complete

REL-03a's six tasks are complete and evidenced: Go public API diff via `apidiff`/`gorelease` wired as a
CI job; a module compile matrix with explicit version exclusions; config-schema compatibility with a
seeded breaking-config fixture; a migration upgrade-from-oldest-supported drill extending
`TestIntegrationMigrationsReversible`; container architecture smoke on every published architecture;
SBOM/provenance/signature verification folded in from REL-01 T8/T9. Traces to W06-E02-S002.

## AC-W06-E02-03 — Compatibility gates blocked set honestly recorded

REL-03b's three legs (OpenAPI semantic diff, event/schema compatibility, generated-consumer upgrade
check) are each recorded with explicit per-leg blocked-entry criteria naming the exact unblocking
story. Any leg that unblocks during this epic's execution window is completed and evidenced; any leg
still blocked at closure is recorded as deferred-with-restated-unblocking-condition. Traces to
W06-E02-S003.

## AC-W06-E02-04 — Independent review passed

S001 and S002 have passed independent review per mandate §14. S003's review (scoped to whichever legs
actually unblock and complete within this epic's window) specifically confirms the still-blocked legs'
entry criteria are honestly stated in `story.md`, not silently bypassed or falsely marked complete.

## Acceptance authority

Release/security-engineering lead, per PLAN §5.6's "Accountable role: release/security-engineering
lead" for PF-REL, applied to REL-03; MATRIX CS-15's shared DX-06/REL-03 scope confirms the same
accountable role covers DX-06 within this epic.
