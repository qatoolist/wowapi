---
id: W05-E01-DEPS
type: epic-dependencies
epic: W05-E01
wave: W05
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W05-E01 — Dependencies

## Upstream (epics/waves this epic depends on)

- **W03-E01** (full-wave entry gate) — per `../../dependencies.md` (wave-level), W05 depends on
  W03-E01 acceptance. This epic has no additional upstream dependency beyond that gate.

## Downstream (epics/waves that depend on this epic)

| Downstream item | Depends on (from this epic) | Why |
|---|---|---|
| W05-E02 (AR-02, this wave) | W05-E01-S001 (T1, T2) | PLAN AR-02 T1's own dependency row: "Depends-on: AR-01 T1, T2" — AR-02 reuses AR-01 T2's `Registrar` type directly. |
| W05-E03 (AR-03, this wave) | W05-E01-S001 through S004 (full epic) | PLAN AR-03 T3's own dependency row: "Depends-on: T1, AR-01, AR-02" — the manifest-derived-projection tooling requires the ownership-bound model to exist. |
| W05-E05 (FBL-01, this wave) | W05-E01 (full epic) | MATRIX CS-01's own "Dependencies: AR-01/02 first (re-homing mid-registration-rework causes double churn)." |

## Internal (within this epic)

S001 → S002 → S003 → S004 in strict dependency order, matching PLAN AR-01's own T-number
dependency chain:

- S002 depends on S001 (T3-T6 depend on T1, T2 — the `ApplicationModel` skeleton and `Registrar`
  capability type must exist before any per-registry ownership wrapper can be built against them).
- S003 depends on S002 (T7 depends on T3-T6; T9-T10 depend on the full preceding task surface). T8
  (post-seal rejection, grouped into S003 per `impl/analysis/wave-allocation-detail.md`) depends
  only on T1, T2 directly and is parallel-safe with S002's own tasks, but is sequenced into S003 for
  story-grouping coherence with T7/T9/T10's immutability/determinism/race-safety theme, not because
  it has a hard dependency on S002 completing first.
- S004 depends on S003 (T11 depends on T1-T10 in full — the legacy adapter wraps the complete,
  race-safe, deterministic-hash model, not a partially-built one).

## Cross-wave dependencies

None beyond the W03-E01 entry dependency and the downstream table above (which includes both
within-wave and no cross-wave targets beyond what `../../dependencies.md` already states at wave
scope for W06).

## External dependencies

None new. This epic's work is internal Go type/API design over the existing `kernel/module`,
`kernel/resource`, `kernel/rules`, `kernel/authz` packages.

## Repository dependencies

wowsociety's module-contract tests are a required regression check for S004 (AR-01 T11's own
acceptance criterion: "existing contract tests pass unmodified through the legacy path") — this is
a verification dependency, not a code dependency; no wowsociety code change is required for this
epic's own closure per REVIEW's own AR-01 wowsociety-impact note ("No wowsociety change required
before/during Wave 1 landing").

## Tooling dependencies

None new.

## Decision dependencies

D-02 and D-03, both ratified in W00-E02-S003, referenced by S001. See `epic.md` "Required
decisions."
