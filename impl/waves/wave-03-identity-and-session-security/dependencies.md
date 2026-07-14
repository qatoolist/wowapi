---
id: W03-DEPS
type: wave-dependencies
wave: W03
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W03 — Dependencies

## Upstream (waves this wave depends on)

- **W01** — depends specifically on W01-E03-S002's central-validation (`RouteMeta`) seam.
  `impl/index.md`'s wave map states W03 depends on "W01 (validation seam)": SEC-01's new
  grant-table endpoints (T1/T2) and SEC-03's `Envelope`-consuming webhook endpoints should be built
  against the `RouteMeta` contract-enforcement pattern W01 establishes, not against a moving target.
- **W02** — depends specifically on W02-E01's DATA-09 online expand/backfill/validate/contract
  protocol. `impl/index.md`'s wave map states W03 depends on "W02 (grant-table migration uses
  DATA-09)": PLAN SEC-01 T1's own risk column states "Schema is genuinely new — get security-lead
  sign-off before merge," and DATA-09's own cross-cutting note ranks the `identity_grant` migration
  among the risky migrations that should follow the online protocol rather than a one-off unsafe
  migration.
- **W00-E02-S003** — the ADR-ification story. This wave's E01-S001 references `ADR-W00-E02-S003-001`
  (D-01: framework owns grant validity/expiry/revocation) and E02-S001 references
  `ADR-W00-E02-S003-007` (D-07: JWKS trusted-issuer config gate) as already-ratified design premises,
  not decisions this wave makes.

## Downstream (waves/epics that depend on this wave)

| Downstream item | Depends on (from W03) | Why |
|---|---|---|
| W05 (entire wave entry) | W03-E01 acceptance | `impl/analysis/wave-allocation-detail.md` "Cross-wave sequencing notes": "W05 entry requires W03-E01 acceptance (actor model stability)." AR-01/AR-02's ApplicationModel and registrar-capability work assumes a stable, server-side-verified actor model — building it against SEC-01's still-in-flux claim-copy behavior would require rework. |
| W05-E01 (AR-01) | W03-E01 acceptance | Same actor-model-stability reasoning as above, at epic scope. |
| W05-E04-S002 (SEC-04) | soft — W03-E04-S001-T4 references it | The relationship is bidirectional in practice: W03-E04-S001-T4's cache-invalidation acceptance criterion is deferred-linked to W05-E04-S002's epoch-table work landing, but W05-E04-S002 does not itself require W03 to be complete first — see "Cross-wave dependencies" below for the precise direction. |

## Internal epic dependencies (within W03)

| Dependent | Depends on | Type | Notes |
|---|---|---|---|
| W03-E04-S001 (DATA-07) | W03-E01 (SEC-01), **accepted**, not merely started | Hard | PLAN §5.3 DATA-07 T1 row, verbatim: "Hard dependency on PF-SEC's SEC-01 — do not schedule before it lands." `requirement-inventory.md` row DATA-07 records disposition "blocked→planned" for exactly this reason. This wave's internal sequencing must not start E04-S001's implementation work before E01's `closure.md` records `accepted`. |
| W03-E05-S001-T5 | W03-E01-S001's grant-ID field | Hard | `impl/analysis/wave-allocation-detail.md` E05 row: "T5 durable audit (grant-ID field dep on E01 S001)." PLAN SEC-02 T5's own Depends-on column: "T1, T3, T4; benefits from SEC-01 T1." |
| W03-E04-S001-T4 | W05-E04-S002 (SEC-04 epoch table, D-06) | Soft, deferred-linked | `impl/analysis/wave-allocation-detail.md` E04 row: "SEC-04 epoch dep noted soft — cache-invalidation AC deferred-linked to W05-E04-S002." PLAN DATA-07 T4's own Depends-on column: "T1-T3; also depends on SEC-04's cache-epoch work," with the explicit warning "do not assume PF-SEC delivers on PF-DATA's timeline" (read here as: do not assume W05 delivers the epoch table on W03's timeline). |
| W03-E02-S001 | none (within W03) | — | SEC-06 is independent of E01/E03/E04/E05 — no shared file surface, no design premise dependency beyond D-07 (upstream, not intra-wave). |
| W03-E03-S001 | none (within W03) | — | SEC-03 is independent of E01/E02/E04/E05 — its own `Verifier` interface change touches `kernel/webhook` only. |

## Cross-wave dependencies

- W03-E01 → W05 entry (see "Downstream" above).
- W03-E04-S001-T4 ↔ W05-E04-S002 (see "Internal epic dependencies" above) — the acceptance-criterion
  deferral runs from W03 toward W05, not the reverse; W03-E04's other three acceptance criteria
  (T1, T2, T4's non-cache-invalidation portions) do not wait on W05.
- W03-E01-S004 (cross-repo-cutover-plan) references PROD-04 and, indirectly, the eventual wowsociety
  auth-flow rework — that rework itself is out of this programme's framework-implementation scope
  (mandate §2.3: product-level items are recorded, not implemented) and has no formal cross-wave
  entry in this programme; it is tracked as a coordination artifact only.

## External dependencies

- wowsociety staging environment and its `identity_impersonation_session` data — required for
  E01-S004's staging-validation plan and for the SEC-01 T2 rollout note ("validate T2 against
  wowsociety staging data before making it unconditional"). Availability/timeline is not assumed;
  see E01-S004's `plan.md` "Unresolved questions."
- IdP claim-contract confirmation (DEC-Q1) — human-blocked, tracked but not assumed resolved within
  this wave; see `wave.md` "Assumptions."

## Repository dependencies

wowapi-internal for all implementation work (E01–E03, E05). E04 (DATA-07) is wowapi-internal
(`kernel/relationship`); `requirement-inventory.md` row DATA-07 records "No confirmed direct usage"
in wowsociety, to be re-verified at ship time per PLAN's own note. E01-S004 is the sole
cross-repository-facing artifact in this wave, and it is documentation/coordination only — no
wowsociety code is written by this wave (mandate §2.3 framework/product boundary).

## Tooling dependencies

None beyond what W00/W01/W02 already establish (pgx, jwt/v5, the DATA-09 migration tooling, the
`RouteMeta` validation seam).

## Decision dependencies

- D-01 (`ADR-W00-E02-S003-001`) — from W00-E02-S003, consumed (referenced, not re-authored) by
  W03-E01-S001.
- D-07 (`ADR-W00-E02-S003-007`) — from W00-E02-S003, consumed (referenced, not re-authored) by
  W03-E02-S001.
- DEC-Q1 (IdP `grant_id` claim contract) — human-blocked, tracked at `requirement-inventory.md` §B;
  W03-E01-S001 proceeds against its documented safe default rather than waiting for resolution; see
  `wave.md` "Assumptions."
