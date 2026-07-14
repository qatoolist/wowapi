---
id: W05-E05-DEPS
type: epic-dependencies
epic: W05-E05
wave: W05
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W05-E05 — Dependencies

## Upstream (epics/waves this epic depends on)

- **W05-E01** (full epic) and **W05-E02** (full epic) — MATRIX CS-01's own explicit "Dependencies:
  AR-01/02 first (re-homing mid-registration-rework causes double churn)."

## Downstream (epics/waves that depend on this epic)

None identified within this wave or beyond — FBL-01's own re-home is not itself a named dependency
of any other W05 epic or later wave's own stories in `impl/analysis/wave-allocation-detail.md`.

## Internal (within this epic)

S001 (the mechanical move and shim) → S002 (verification). S002's own acceptance bar (package-count
AC, wowsociety identity-suite-green-on-shim AC) cannot be meaningfully evaluated before S001's move
and shim are in place.

## Cross-wave dependencies

None beyond the upstream W05-E01/E02 dependency.

## External dependencies

None new. This epic reuses the existing `depguard` and `scripts/lint_boundaries.sh` tooling — MATRIX
CS-01's own "Reuse tier: fuller configuration of existing tools ... this is a utilisation win, no new
tooling."

## Repository dependencies

wowsociety imports exactly one of the nine re-homed packages: `kernel/mfa` (5 files,
`internal/modules/identity/`) — grep-verified per REVIEW §J/§O. This is `PROD-02` in
`requirement-inventory.md` §D. The other 8 packages have zero wowsociety imports and move
mechanically. This epic's own S002 requires wowsociety's identity/authz suite to run green against
the new path or the shim — a verification dependency, not a code dependency this epic modifies in
wowsociety's own repository.

## Tooling dependencies

None beyond the already-present depguard/boundaries-lint infrastructure this epic extends.

## Decision dependencies

None. See `epic.md` "Required decisions."
