---
id: TRACK-DECISION-REGISTER
type: register
title: Decision register — ratified architecture decisions, open human decisions, provisional assumptions
status: active
created_at: 2026-07-12
updated_at: 2026-07-13
derived: true
---

# Decision register

DERIVED VIEW. Per mandate §11.8: record architectural and implementation decisions, including
unresolved decisions — not buried only in prose. Canonical source = REVIEW §U
(`fable5-final-architecture-review-2026-07-11.md` lines 384–387) for D-01..D-09, the three DEC-Qx
rows in `impl/analysis/requirement-inventory.md` §B, and the safe-default framing in REVIEW §F Q1
/ §7 for the provisional planning assumptions (PA-0x). No dedicated `planning-assumptions.md`
exists in the repository yet at the time of writing this register; the PA-0x rows below record
the safe defaults exactly as stated in the source documents (REVIEW §F Q1 and PLAN §7 items 1,
8, 10) rather than a not-yet-created secondary document.

## (a) Ratified architecture decisions (D-01..D-09)

Status `ratified`: each decision was ratified by Fable 5 in REVIEW §U; the formal ADR documents
were written and verified by story W00-E02-S003 on 2026-07-13. ADR paths below are relative to
`impl/waves/wave-00-baseline-and-verification/epics/epic-002-baseline-capture/stories/story-003-adr-ification/decisions/`.

| Decision ID | One-line decision | Status | Owner | ADR |
|---|---|---|---|---|
| D-01 | Framework owns grant validity (SEC-01 session-state authority split: framework owns validity/expiry/revocation) | ratified | product/security-lead (per REVIEW §U's own note — D-01 tuning is the exception to the default framework/Fable 5 ownership) | adr-001-framework-owns-grant-authority.md (ADR-W00-E02-S003-001) |
| D-02 | Single Registrar + typed keys (AR-01/AR-02 provider-graph design: one shared Registrar type with typed port keys, not per-subsystem types) | ratified | framework (Fable 5) | adr-002-single-registrar-typed-keys.md (ADR-W00-E02-S003-002) |
| D-03 | Post-seal error not panic in production (AR-01/AR-04 post-seal-mutation policy: error-only, not panic-in-prod) | ratified | framework (Fable 5) | adr-003-post-seal-mutation-error-not-panic.md (ADR-W00-E02-S003-003) |
| D-04 | Audit hash_version column (DATA-08 W6-T1: add a hash_version discriminator so historical audit rows remain verifiable after the hash contract widens) | ratified | framework (Fable 5) | adr-004-audit-hash-version-column.md (ADR-W00-E02-S003-004) |
| D-05 | GoReleaser split via --skip=publish (REL-01 T6: use GoReleaser's --skip=publish mode + separate publish invocation, not a hand-rolled pipeline) | ratified | framework (Fable 5) | adr-005-goreleaser-skip-publish-split.md (ADR-W00-E02-S003-005) |
| D-06 | Authz epoch table not message bus (SEC-04 T4: cross-pod cache invalidation via an epoch table, not a kernel message bus) | ratified | framework (Fable 5) | adr-006-authz-epoch-table-not-message-bus.md (ADR-W00-E02-S003-006) |
| D-07 | JWKS trusted-issuer config gate (SEC-06 T4: outbound-security escape-hatch governance via a trusted-issuer configuration gate) | ratified | framework (Fable 5) | adr-007-jwks-trusted-issuer-config-gate.md (ADR-W00-E02-S003-007) |
| D-08 | pgx query tracing via a thin in-kernel pgx.QueryTracer over the existing observability port; otelpgx rejected to keep vendor types out of kernel/database | ratified | framework (Fable 5) | adr-008-pgx-query-tracer-not-otelpgx.md (ADR-W00-E02-S003-008) |
| D-09 | Secrets: boot-time-once resolution + restart-based rotation is the documented v1 contract; file-provider is the next increment (DEF-01), no vault client in the kernel | ratified | framework (Fable 5) | adr-009-secrets-boot-time-rotation-contract.md (ADR-W00-E02-S003-009) |

## (b) Open human decisions (DEC-Q1, DEC-Q9, DEC-Q10)

Status `open-human`: cannot be resolved by an implementation agent; requires a named human with
the relevant authority/access.

| Decision ID | One-line description | Status | Owner | Target |
|---|---|---|---|---|
| DEC-Q1 | IdP grant_id claim contract — whether the production IdP will mint an opaque grant_id claim, and who approves break-glass grants (blocked, human — safe default unblocks build per REVIEW §F Q1) | open-human | human (IdP/security-lead decision) | W03-E01 (tracked) |
| DEC-Q9 | Reference-perf-env ownership — no owner or timeline exists for the dedicated Linux amd64 reference performance runner that PERF-02..05 absolute SLOs require (blocked, human — provisional default set: GH runner + reference json) | open-human | human (infra/programme-owner decision) | W07-E01 (tracked) |
| DEC-Q10 | Repo-admin activation of branch/tag/env protection — merge-queue rulesets unavailable on a user-owned repo; no protected release environment exists today (blocked, human — repo-admin action required) | open-human | human (repo-owner decision) | W06-E03 (tracked) |

## (c) Provisional planning assumptions (safe defaults, one per open decision)

Status `provisional — supersede when human decision lands`. Recorded so the programme can proceed
without waiting on the three human decisions above; each records the exact safe default the
programme is building against today.

| Decision ID | One-line description | Status | Owner | Target |
|---|---|---|---|---|
| PA-01 | Safe default for DEC-Q1: SEC-01 proceeds against a framework-owned grant table with conservative default expiry/revocation semantics, framework owns validity/expiry/revocation (per D-01/REVIEW §7 item 2 recommendation), pending the actual IdP grant_id claim contract | provisional — supersede when human decision lands | framework (Fable 5) | W03-E01 |
| PA-02 | Safe default for DEC-Q9: PERF-02..05 proceed against a relative/container benchmarking regime (GH runner + reference json), REVIEW §12 constraint, with absolute-SLO acceptance criteria explicitly deferred pending a dedicated reference runner | provisional — supersede when human decision lands | framework (Fable 5) | W07-E01 |
| PA-03 | Safe default for DEC-Q10: REL-01/REL-02 build the ~85% of release-gating logic that does not require repo-admin actions now, tracking admin-gated work separately ("PF-REL-ADMIN-01") so it does not silently block agent-completable work | provisional — supersede when human decision lands | framework (Fable 5) | W06-E03 |

## Summary

9 ratified decisions (D-01..D-09) + 3 open human decisions (DEC-Q1/Q9/Q10) + 3 provisional
planning assumptions (PA-01..PA-03) = **15 decision-register rows**. All nine D-01..D-09 rows
share the same target story (W00-E02-S003, ADR-ification) per `requirement-inventory.md`'s own
"D-01..D-09" row.
