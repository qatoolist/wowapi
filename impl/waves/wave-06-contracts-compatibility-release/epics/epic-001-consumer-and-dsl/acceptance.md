---
id: W06-E01-ACCEPTANCE
type: epic-acceptance
epic: W06-E01
wave: W06
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W06-E01 — Epic-level acceptance

Reproduces and elaborates `epic.md`'s "Epic acceptance criteria" section as a standalone,
independently-referenceable acceptance record, consistent with the wave-level `../../acceptance.md`
pattern (AC-W06-01/02 there map onto this epic).

## AC-W06-E01-01 — Module DSL design recorded, not implemented

A module-DSL design doc and an ADR-style decision record exist, covering `port`, `Manifest[T]`, and
`Operation[Request,Response]` at design depth, explicitly labeled "target, not implemented" per AR-05's
labeling discipline. No DX-03 implementation code, compiler, or runtime type-system change is produced.
Traces to W06-E01-S001.

## AC-W06-E01-02 — Golden consumer installs and exercises the full subsystem set

The fixture installs via `go install`, not a repo-internal import. Resource, rule, workflow, event
handler, recurring job, document flow, notification, and webhook generation are each exercised at least
once, across at least two modules. Traces to W06-E01-S002.

## AC-W06-E01-03 — Golden consumer boots against real infrastructure and survives an upgrade

The fixture boots API and worker processes against real Postgres/MinIO/Mailpit/OTel with authenticated
CRUD, async delivery, restart/retry, and RLS isolation all passing. An upgrade-from-previous-version
replay (fixture at N-1, upgraded to N, contracts rerun) passes. The fixture is wired into CI as a
required gate. Traces to W06-E01-S002.

## AC-W06-E01-04 — Independent review passed (S002 only)

W06-E01-S002 has passed independent review per mandate §14, specifically confirming the upgrade replay
is a genuine two-pass integration test and the subsystem-coverage claim matches what was actually
exercised, not merely claimed. W06-E01-S001, as a design-investigation story with no code produced,
does not carry an independent-review task — see S001's own `tasks/index.md` for the rationale.

## Acceptance authority

Developer-experience lead, per PLAN §5.4's "Accountable role: developer-experience lead" for PF-DX.
