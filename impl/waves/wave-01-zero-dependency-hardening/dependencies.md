---
id: W01-DEPS
type: wave-dependencies
wave: W01
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W01 — Dependencies

## Upstream (waves this wave depends on)

- **W00** — full wave dependency per the strict W00→W07 entry ordering. Specifically depends on
  W00-E02-S003's D-08 ADR (pgx query-tracer approach) for W01-E02-S002, and on W00-E01's
  re-verification of AR-04 T1/AR-06 T1 (this wave's AR-04/AR-06-remainder-adjacent work is out of
  scope for W01 itself — those land in W05 — but W01-E04-S002's FBL-03 upstream-register work should
  reflect W00's confirmed-current state, not the review documents' possibly-stale claims).

## Downstream (waves that depend on this wave)

| Downstream item | Depends on (from W01) | Why |
|---|---|---|
| W03-E01 (SEC-01) | W01-E03-S002 (central validation seam) | `impl/index.md` wave map notes W03 depends on "W01 (validation seam)" — SEC-01's new grant-table endpoints should be built against the RouteMeta contract-enforcement pattern this wave establishes |
| W05-E03 (AR-03) | W01-E03-S002 (RouteMeta.Request contract) | AR-03 derives projections from RouteMeta; the T1 contract-declaration field this wave adds must already exist as a stable input |
| W06-E02 (DX-06/REL-03) | W01-E01-S001 (FBL-05 pgx contract decision) | REL-03's compatibility gates assume the raw `pgx.Rows` public contract decision (CS-10) is settled, which this wave's FBL-05 story finalizes mechanically |
| All later waves' CI runs | W01-E01-S001/S002/S003 | Every later wave's PR/CI gate runs against the linter/supply-chain configuration this wave lands — a later wave's evidence is only comparable if this wave's gate state is the stable baseline |

## Cross-wave dependencies

None beyond W00→W01 entry and the downstream table above. W01 does not depend on W02-W07.

## External dependencies

- OTel adapter (`adapters/tracing/otel`) already present — W01-E02 extends it, does not introduce a
  new external service dependency.
- `golangci-lint` v2.11.4 pinned toolchain (already installed per W00's toolchain inventory).

## Repository dependencies

None cross-repo — all W01 work is wowapi-internal. wowsociety impact is additive/optional (FBL-06
correlation, FBL-09 backport recommendation) or zero (FBL-05/07/08 internal to wowapi tooling and
boot-time contract, DX-02 generator fix wowsociety already avoided via governance discipline).

## Tooling dependencies

- `internal/tools/docexamples` (new, small, per AR-05 T3 / MATRIX CS-22) is referenced by
  W01-E04-S002's documentation-reconciliation work only if the doc-example CI gate is pulled forward;
  per the story-grouping in this batch, the full doc-example gate (CS-22) is W06-scoped (AR-05 T3/T4),
  not W01 — W01-E04-S002 is limited to T-DOC-01 + DX-05 residual + FBL-03 register updates.

## Decision dependencies

- D-08 (pgx query tracer design) — from W00-E02-S003, consumed by W01-E02-S002.
