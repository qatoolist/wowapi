---
id: W00-E02-S003-DECISIONS-INDEX
type: decisions-index
parent_story: W00-E02-S003
status: complete
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W00-E02-S003 — Decisions index

Per mandate §11.8: "Record architectural and implementation decisions, including unresolved
decisions. Do not bury decisions only in prose." This is the ONE exception among this epic's three
stories (S001, S002, S003) where a `decisions/` directory is created — S001 and S002 do not produce
architecture decisions of their own and so do not get one.

Each row's `status: ratified` asserts that the underlying decision was already made by Fable 5 in
`docs/implementation/fable5-final-architecture-review-2026-07-11.md` §F/§U — see each ADR's own
"Formalization note" for the explicit distinction between that and this programme's own
tracking-execution status (the owning task's `status: todo`/`done`).

| D-0N | ADR file | Title | Status | Owner |
|---|---|---|---|---|
| D-01 | [adr-001-framework-owns-grant-authority.md](adr-001-framework-owns-grant-authority.md) | Framework owns grant validity/expiry/revocation authority | ratified | Fable 5 (framework architecture lead role); product/security-lead (D-01 tuning — IdP claim shape only) |
| D-02 | [adr-002-single-registrar-typed-keys.md](adr-002-single-registrar-typed-keys.md) | One generic owner-bound Registrar type with per-subsystem typed keys | ratified | Fable 5 (framework architecture lead role) |
| D-03 | [adr-003-post-seal-mutation-error-not-panic.md](adr-003-post-seal-mutation-error-not-panic.md) | Post-seal mutation errors in production, panics only under an explicit dev/test build tag | ratified | Fable 5 (framework architecture lead role) |
| D-04 | [adr-004-audit-hash-version-column.md](adr-004-audit-hash-version-column.md) | Audit hash_version smallint column, version-branched verification | ratified | Fable 5 (framework architecture lead role) |
| D-05 | [adr-005-goreleaser-skip-publish-split.md](adr-005-goreleaser-skip-publish-split.md) | GoReleaser --skip=publish build-candidate + separate publish step | ratified | Fable 5 (framework architecture lead role) |
| D-06 | [adr-006-authz-epoch-table-not-message-bus.md](adr-006-authz-epoch-table-not-message-bus.md) | Per-tenant authz_epoch table, polled on the existing authz read path; LISTEN/NOTIFY optional only | ratified | Fable 5 (framework architecture lead role) |
| D-07 | [adr-007-jwks-trusted-issuer-config-gate.md](adr-007-jwks-trusted-issuer-config-gate.md) | Trusted-issuer/egress config as a declared fingerprinted field; custom JWKS client gated in prod | ratified | Fable 5 (framework architecture lead role) |
| D-08 | [adr-008-pgx-query-tracer-not-otelpgx.md](adr-008-pgx-query-tracer-not-otelpgx.md) | Thin in-kernel pgx.QueryTracer over the existing observability port; otelpgx rejected | ratified | Fable 5 (framework architecture lead role) |
| D-09 | [adr-009-secrets-boot-time-rotation-contract.md](adr-009-secrets-boot-time-rotation-contract.md) | Secrets: boot-time-once resolution + restart-based rotation is the documented v1 contract | ratified | Fable 5 (framework architecture lead role) |

## Consistency note

This index is cross-checked against each ADR file's own front matter (`id`, `title`, `status`,
`deciders`) as part of AC-W00-E02-S003-02's verification — see `../verification.md`. If this table
and an ADR's front matter ever disagree, the ADR file's own front matter is authoritative (same
canonical-source-of-truth rule as `impl/governance/status-model.md` applies to any per-item
document vs. its roll-up index) and this table must be corrected to match, not the reverse.
