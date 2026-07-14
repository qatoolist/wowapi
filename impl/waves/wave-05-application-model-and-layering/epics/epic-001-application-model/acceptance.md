---
id: W05-E01-ACCEPTANCE
type: epic-acceptance
epic: W05-E01
wave: W05
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W05-E01 — Epic-level acceptance

Reproduces and elaborates `epic.md`'s "Epic acceptance criteria" section as a standalone,
independently-referenceable acceptance record, consistent with the wave-level `../../acceptance.md`
pattern (AC-W05-01 there maps onto this epic).

## AC-W05-E01-01 — Lifecycle skeleton and Registrar capability type enacted per D-02/D-03

The `ApplicationModel` compiles via `collect → validate → seal → expose read-only snapshot`;
post-seal registration calls error in production (panic only under an explicit dev/test build tag,
per D-03); the `Registrar` capability type is a single generic owner-bound type with per-subsystem
typed keys, mintable only by the compiler, per D-02. Traces to W05-E01-S001.

## AC-W05-E01-02 — Every declaration class ownership-checked

Resource, rules, authz-permission registration, and the ~9+ remaining declaration classes
(events, jobs, workflow actions, providers, templates, health checks, migrations, seeds, OpenAPI)
are each ownership-checked, proven by a table-driven adversarial suite with one fixture per class,
including the previously-zero-ownership-check `authz.Registry.Register(p Permission)` surface.
Traces to W05-E01-S002.

## AC-W05-E01-03 — Snapshot immutability, post-seal rejection, determinism, and race safety proven

No exported registry reader returns a backing map/slice; a retained registrar post-boot errors on
mutation, never a silent no-op or production panic; two identical compiles emit a byte-identical
model hash; `go test -race` is clean on concurrent legitimate reads. Traces to W05-E01-S003.

## AC-W05-E01-04 — Legacy adapter compatibility proven

Existing modules (wowapi-internal and wowsociety) boot unchanged through the legacy adapter;
existing contract tests pass unmodified through the legacy path; the adapter derives owner from
`Module.Name()` and does not bypass the ownership checks established by S002. Traces to
W05-E01-S004.

## AC-W05-E01-05 — Independent review passed

All four stories (S001, S002, S003, S004) have passed independent review per mandate §14. S001's
review specifically confirms D-02/D-03 are enacted as ratified, not re-interpreted. S002's review
specifically confirms T5's adversarial test and T6's declaration-class enumeration are both
genuinely proven, not merely implemented. S004's review specifically confirms the legacy adapter
does not bypass any ownership check established elsewhere in this epic.

## Acceptance authority

Framework architecture lead, per `../../wave.md`'s wave-level acceptance authority (PLAN §5.1's
accountable role for PF-ARCH).
