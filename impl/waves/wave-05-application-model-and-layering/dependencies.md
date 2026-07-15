---
id: W05-DEPS
type: wave-dependencies
wave: W05
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W05 — Dependencies

## Upstream (waves this wave depends on)

- **W03** — full-wave dependency, specifically W03-E01 (server-side session/grant state, SEC-01)
  acceptance. Per `impl/analysis/wave-allocation-detail.md`'s "Cross-wave sequencing notes": "W05
  entry requires W03-E01 acceptance (actor model stability)." This wave's `Registrar` capability
  type (AR-01 T2) is the framework's security boundary for module registration, and the actor/
  session model W03-E01 stabilises is a load-bearing assumption for that boundary's correctness —
  starting W05's registration-model work against an actor model still in flux risks rework of the
  security-boundary type itself.
- **W00** (programme baseline) — D-02, D-03, D-06 are ratified in W00-E02-S003; this wave's
  W05-E01-S001 and W05-E04-S002 reference (not author) those ADRs. No other W05 story depends on a
  specific D-0N ADR — confirmed by scanning `requirement-inventory.md` §B for any D-0N row targeting
  AR-02, AR-03, AR-04, or FBL-01: none exists.

## Downstream (waves that depend on this wave)

| Downstream item | Depends on (from W05) | Why |
|---|---|---|
| W06-E02-S003 (REL-03b compatibility-gate legs) | W05-E03 (AR-03 manifest and projections) | `impl/index.md`'s wave map states W06 depends on "W05 (AR-03 unblocks REL-03b legs)" — the compatibility gates that assert manifest-derived projections match cannot run until AR-03's manifest/projection tooling exists. |
| W06-E04-S002 (AR-05 T4/T5 generated docs and labels) | W05-E03 (AR-03 manifest) | `impl/analysis/wave-allocation-detail.md`'s W06-E04-S002 row states "dep E02/W05-E03 manifest" — AR-05's generated-reference-docs projection is derived from AR-03's authoritative manifest, which this wave builds. |

## Cross-wave dependencies

None beyond the W03 entry dependency and the two downstream consumers (W06) stated above. W05 does
not depend on W01, W02, or W04.

## Internal (within this wave, between epics)

- **W05-E02 depends on W05-E01.** AR-02 T1/T2 reuse AR-01 T2's `Registrar` type directly (PLAN AR-02
  T1: "reuse AR-01 T2's `Registrar`") and AR-02 T2's registrar-forge safety claim depends on AR-01's
  capability type being in place first. `impl/analysis/wave-allocation-detail.md`'s own note: "AR-02
  ... Depends AR-01 T1/T2."
- **W05-E03 depends on W05-E01.** AR-03's manifest-derived-projection tooling (T3) explicitly lists
  `AR-01, AR-02` as a dependency in PLAN's own task table — the manifest cannot derive route/
  permission/resource projections from a model that does not yet exist in ownership-bound form.
- **W05-E05 depends on W05-E01 and W05-E02.** MATRIX CS-01's own "Dependencies" line: "AR-01/02
  first (re-homing mid-registration-rework causes double churn)." Re-homing the nine kernel packages
  while AR-01/AR-02's registration-model rework is still in flight would force the re-home to churn
  twice — once for the move, once again when the registration model lands. E05 is accordingly
  sequenced last among this wave's epics.
- **W05-E04 has no dependency on W05-E01/E02/E03.** AR-06's remaining tasks (T2, T3) concern only
  `kernel/kernel.go`'s constructor-bypass closure, a disjoint concern from the ownership-model
  rework. SEC-04's cache-bounding work (all tasks) is independent of AR-01/02/03's registration
  model — its only dependency is D-06's ratification (W00 baseline, already satisfied) and, for T4's
  epoch-bump wiring, awareness of DATA-07's grant-table mutation paths (W03 scope, already landed by
  this wave's own W03-E01 entry gate). E04's two stories may proceed in parallel with E01/E02/E03.
- **Within W05-E01**: S001 → S002 → S003 → S004 in strict dependency order, matching PLAN AR-01's
  own T-number dependency chain (T3-T6 depend on T1, T2; T7 depends on T3-T6; T8 depends on T1, T2;
  T9-T10 depend on T1-T9/T1-T8; T11 depends on T1-T10). S001 (T1, T2) is the epic's own foundation —
  every other AR-01 story, and AR-02/AR-03 at wave scope, depend on it.
- **Within W05-E02**: S001 → S002 → S003, matching PLAN AR-02's own chain (T3-T4 depend on T1-T2;
  T5 depends on T1-T4 and AR-03; T6 depends on T1-T5; T7 depends on T1-T6).
- **Within W05-E03**: S001 and S002 are independent of each other. AR-03 (S001) depends on W05-E01/
  E02 at wave scope, not on AR-04 (S002). AR-04 (S002) depends on AR-01 per PLAN's own AR-04 T2 row
  ("Depends-on: AR-01 T1"), i.e. on W05-E01, not on W05-E03-S001 itself.
- **Within W05-E04**: S001 (AR-06) and S002 (SEC-04) are independent of each other — disjoint code
  surface (`kernel/kernel.go` constructor closures vs. `kernel/authz/caching.go`).
- **Within W05-E05**: S001 (foundation-move-and-shims) → S002 (re-home-verification). S002's
  acceptance (kernel package-count AC, wowsociety identity-suite-green-on-shim AC) cannot be
  evaluated before S001's mechanical move and shim are in place.

## External dependencies

None new. FBL-01's re-home uses the existing Go toolchain and `depguard`/`scripts/lint_boundaries.sh`
tooling already present in the repository (MATRIX CS-01: "Reuse tier: fuller configuration of
existing tools ... this is a utilisation win, no new tooling"). SEC-04's T1 uses
`hashicorp/golang-lru/v2`, named as an "approved dep" in MATRIX CS-17.

## Repository dependencies

FBL-01 has a real, bounded wowsociety-facing dependency: wowsociety imports exactly one of the nine
re-homed packages (`kernel/mfa`, 5 files in `internal/modules/identity/`) — grep-verified per REVIEW
§J/§O. This is tracked as `PROD-02` in `requirement-inventory.md` §D. The other 8 re-homed packages
have zero wowsociety imports and move in a single mechanical commit with no cross-repo coordination
required. This wave's E05-S002 requires wowsociety's identity/authz suite to run green against the
new `foundation/mfa` path (or the shim during the grace window) as its own acceptance bar — the
product-side code migration itself is out of this wave's framework-only scope, but the verification
that the shim/new path does not break wowsociety is in scope.

## Tooling dependencies

None new beyond the already-present depguard/boundaries-lint infrastructure FBL-01 extends, and the
`golang-lru/v2` dependency SEC-04 T1 adds.

## Decision dependencies

- W05-E01-S001 depends on D-02 (single Registrar + typed keys) and D-03 (post-seal error not panic),
  both ratified in W00-E02-S003 — referenced, not authored, in this story's `decisions/index.md`.
- W05-E04-S002 depends on D-06 (per-tenant `authz_epoch` table, polled; not a message bus), ratified
  in W00-E02-S003 — referenced, not authored, in this story's `decisions/index.md`.
- No other W05 story has a decision dependency — confirmed against `requirement-inventory.md` §B (see
  `wave.md` "Assumptions").
