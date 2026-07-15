---
id: W00-E02-S003-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W00-E02-S003
status: complete
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W00-E02-S003 — Artifacts index

Per mandate §9.2. Nine ADR files are the sole artifact type this story produces, all at lifecycle
stage "implementation" (mandate §9.3 lists "architecture decisions" as a named implementation-stage
example). No `pre-implementation/` or `post-implementation/` subdirectory content applies to this
story. Per this repository's Adaptation 2 (`naming-conventions.md`), no `implementation/`
subdirectory is pre-created — each ADR file lives directly under `../decisions/` (the story's
`decisions/` directory doubles as the artifact's authoritative repository location; this index
records that path rather than duplicating the file under `artifacts/implementation/`).

All nine were produced (authored 2026-07-12 by the story authoring pass; verified and
status-vocabulary-corrected 2026-07-13 at commit 0a31186cada5c275a588c74081cf977adf346e61) and are
`current`.

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Repository path | Version | Checksum | Status | Reviewer | Retention requirement |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| ART-W00-E02-S003-001 | ADR — Framework owns grant authority | architecture decision / design document | implementation | Formalizes D-01: framework owns grant validity/expiry/revocation; wowsociety keeps its table for product UX/audit only | D-01 | W00-E02-S003-T001 | `decisions/adr-001-framework-owns-grant-authority.md` | 1 | not applicable (text document, repo file is authoritative) | current | W00-E02-S003 execution worker + reviewer subagent (2026-07-13) | Permanent — durable decision record, retained for the life of the programme |
| ART-W00-E02-S003-002 | ADR — Single Registrar, typed keys | architecture decision / design document | implementation | Formalizes D-02: one generic owner-bound `Registrar` type with per-subsystem typed keys | D-02 | W00-E02-S003-T001 | `decisions/adr-002-single-registrar-typed-keys.md` | 1 | not applicable | current | W00-E02-S003 execution worker + reviewer subagent (2026-07-13) | Permanent |
| ART-W00-E02-S003-003 | ADR — Post-seal mutation error, not panic | architecture decision / design document | implementation | Formalizes D-03: post-seal mutation errors in prod builds, panics only under dev/test build tag | D-03 | W00-E02-S003-T001 | `decisions/adr-003-post-seal-mutation-error-not-panic.md` | 1 | not applicable | current | W00-E02-S003 execution worker + reviewer subagent (2026-07-13) | Permanent |
| ART-W00-E02-S003-004 | ADR — Audit hash_version column | architecture decision / design document | implementation | Formalizes D-04: `hash_version smallint NOT NULL DEFAULT 1` column, version-branched verification | D-04 | W00-E02-S003-T002 | `decisions/adr-004-audit-hash-version-column.md` | 1 | not applicable | current | W00-E02-S003 execution worker + reviewer subagent (2026-07-13) | Permanent |
| ART-W00-E02-S003-005 | ADR — GoReleaser skip-publish split | architecture decision / design document | implementation | Formalizes D-05: `release --skip=publish` build-candidate + separate `goreleaser publish` step | D-05 | W00-E02-S003-T002 | `decisions/adr-005-goreleaser-skip-publish-split.md` | 1 | not applicable | current | W00-E02-S003 execution worker + reviewer subagent (2026-07-13) | Permanent |
| ART-W00-E02-S003-006 | ADR — authz_epoch table, not message bus | architecture decision / design document | implementation | Formalizes D-06: per-tenant epoch table polled on existing authz read path; LISTEN/NOTIFY optional only | D-06 | W00-E02-S003-T002 | `decisions/adr-006-authz-epoch-table-not-message-bus.md` | 1 | not applicable | current | W00-E02-S003 execution worker + reviewer subagent (2026-07-13) | Permanent |
| ART-W00-E02-S003-007 | ADR — JWKS trusted-issuer config gate | architecture decision / design document | implementation | Formalizes D-07: trusted-issuer/egress config as a declared fingerprinted field; custom JWKS client gated in prod | D-07 | W00-E02-S003-T002 | `decisions/adr-007-jwks-trusted-issuer-config-gate.md` | 1 | not applicable | current | W00-E02-S003 execution worker + reviewer subagent (2026-07-13) | Permanent |
| ART-W00-E02-S003-008 | ADR — pgx query tracer, not otelpgx | architecture decision / design document | implementation | Formalizes D-08: thin in-kernel `pgx.QueryTracer` over existing observability `Tracer` port; `otelpgx` rejected | D-08 | W00-E02-S003-T003 | `decisions/adr-008-pgx-query-tracer-not-otelpgx.md` | 1 | not applicable | current | W00-E02-S003 execution worker + reviewer subagent (2026-07-13) | Permanent |
| ART-W00-E02-S003-009 | ADR — Secrets boot-time rotation contract | architecture decision / design document | implementation | Formalizes D-09: boot-time-once resolution + restart-based rotation as v1 contract; no vault client in kernel | D-09 | W00-E02-S003-T003 | `decisions/adr-009-secrets-boot-time-rotation-contract.md` | 1 | not applicable | current | W00-E02-S003 execution worker + reviewer subagent (2026-07-13) | Permanent |

## Cross-reference

`decisions/index.md` is the decision-register view of the same nine files (D-0N ID → ADR file →
title → status → owner). This `artifacts/index.md` is the artifact-management view (mandate §9.2
fields: type, lifecycle stage, producing task, retention). Both describe the same nine underlying
files; neither duplicates the other's content, only its indexing purpose.
