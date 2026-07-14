---
id: W05-E05
type: epic
title: Kernel re-home
status: planned
wave: W05
owner: unassigned
reviewer: unassigned
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-12
source_requirements:
  - FBL-01
depends_on:
  - W05-E01
  - W05-E02
stories:
  - W05-E05-S001
  - W05-E05-S002
decisions: []
risks:
  - RISK-W05-003
  - RISK-W05-004
---

# W05-E05 — Kernel re-home

## Epic objective

Re-home the nine app-foundation/adapter packages (`webhook`, `notify`, `document`, `artifact`,
`attachment`, `comment`, `bulk`, `integration`, `mfa`) from `kernel/` to a new `foundation/` tree,
retaining `kernel/storage` as the correct port; extend `depguard` and `scripts/lint_boundaries.sh`
to enforce the corrected layering; ship a deprecated forwarding shim for `kernel/mfa`'s
wowsociety-facing surface; and verify wowsociety's identity/authz suite runs green on the new path
or the shim.

## Problem being solved

`requirement-inventory.md` row FBL-01: "Kernel re-home (9 pkgs → foundation/) | IMPL | P1 | planned
| W05-E05-S001..S002 | CS-01 mechanics; dep AR-01/02; wowsociety mfa migration story is
product-coordination (PROD-02)." MATRIX CS-01's own evidence: "`go list ./kernel/...` = 39
sub-packages (40 incl. root; personally verified). Nine are app-foundation/adapter concerns wearing
kernel paths: `webhook, notify, document, artifact, attachment, comment, bulk, integration, mfa`.
`kernel/storage` is a correct port and stays. wowsociety imports exactly one re-home candidate:
`kernel/mfa` (5 files, `internal/modules/identity/`) — grep-verified; the other 8 = 0 imports."
MATRIX CS-01's own defect framing: "delivery engines with network I/O, document services, and
feature subsystems live at kernel import paths; the kernel cannot honour 'small and stable' while
owning them... v1 stabilisation would freeze the wrong public surface." This is, per MATRIX CS-01's
own words, "the largest single architectural correction and must precede v1 stabilisation or the
wrong surface locks in" (REVIEW §J).

## Scope

- Creating the `foundation/` tree (S001).
- `git mv` for each of the 9 packages, updating import paths repo-wide — mechanical for 8 of 9
  packages (zero-consumer outside wowapi) (S001, MATRIX CS-01 step 2).
- `kernel/mfa` → `foundation/mfa` with a deprecated forwarding shim (type aliases + var forwarding)
  left at `kernel/mfa` for one minor version, so wowsociety migrates on its own schedule, then
  removed (S001, MATRIX CS-01 step 3).
- Extending `depguard` (`.golangci.yml` kernel rule) to deny `kernel → foundation` imports, and
  adding a `foundation` rule denying `foundation → app` (S001, MATRIX CS-01 step 4).
- Extending `scripts/lint_boundaries.sh`'s allowlist so a new kernel package addition fails CI
  without an explicit allowlist edit (review-forcing) (S001, MATRIX CS-01 step 5).
- Verifying `go list ./kernel/... | wc -l` is at or below the target-list count; depguard + boundaries
  lint green; wowsociety's identity/authz suite green on `foundation/mfa` or the shim (S002, MATRIX
  CS-01's acceptance bar).

## Out of scope

- **wowsociety's own code migration off `kernel/mfa` onto `foundation/mfa`** — this is PROD-02 in
  `requirement-inventory.md` §D, explicitly product-level coordination, out of this epic's
  framework-only scope per mandate §2.3. This epic ships the forwarding shim that makes wowsociety's
  own migration possible on its own schedule; it does not perform that migration.
- **`kernel/storage`'s own move** — MATRIX CS-01's own evidence: "`kernel/storage` is a correct port
  and stays." Not re-homed.
- **Any behavioral change to the nine re-homed packages' own logic** — MATRIX CS-01's own framing:
  "behaviour-preserving moves," "import-path churn only." This epic does not modify what any of the
  nine packages do, only where they live.

## Source requirements

FBL-01. No D-0N architecture-decision dependency in the source — confirmed by scanning
`requirement-inventory.md` §B: no D-0N row targets FBL-01 (D-02/D-03 are cited by MATRIX CS-01 only
as "sequencing context," not as a decision this epic itself enacts).

## Architectural context

FBL-01 depends on W05-E01 and W05-E02 (AR-01, AR-02) both completing first, per MATRIX CS-01's own
explicit "Dependencies: AR-01/02 first (re-homing mid-registration-rework causes double churn)." This
is why E05 is sequenced last among this wave's five epics. Within this epic, S001 (the mechanical
move and shim) must complete before S002 (verification) can meaningfully evaluate its own acceptance
bar.

## Included stories

- **W05-E05-S001 — foundation-move-and-shims** (MATRIX CS-01 mechanics: `git mv` 9 packages, `mfa`
  forwarding shim, depguard+boundaries extension).
- **W05-E05-S002 — re-home-verification** (kernel package-count AC, wowsociety identity-suite-green-
  on-shim AC — PROD-02 coordination).

## Dependencies

Depends on W05-E01 (full epic) and W05-E02 (full epic). No dependency on W05-E03 or W05-E04. This
epic is this wave's own last-sequenced epic per its own upstream dependency chain.

## Risks

RISK-W05-003 (FBL-01's own schedule risk, given its dependency on E01/E02 landing first, and its
status as "the largest single architectural correction") and RISK-W05-004 (the `kernel/mfa` shim's
auth-critical, security-sensitive migration coordination) both originate at wave scope and land
entirely within this epic's two stories. See `risks.md` for the epic-scoped elaboration.

## Required decisions

None. FBL-01 has no D-0N architecture-decision dependency in the source (confirmed — see "Source
requirements" above).

## Epic acceptance criteria

- **AC-W05-E05-01**: The `foundation/` tree exists with all nine packages moved (`git mv`, preserving
  history); import paths are updated repo-wide; the `kernel/mfa` deprecated forwarding shim is in
  place (type aliases + var forwarding).
- **AC-W05-E05-02**: The extended `depguard` rule denies `kernel → foundation` and `foundation → app`
  imports; `scripts/lint_boundaries.sh`'s extended allowlist fails CI on a new, un-allowlisted kernel
  package addition; the boundary-lint fixture that fails today against the nine packages is confirmed
  to pass after the re-home.
- **AC-W05-E05-03**: `go list ./kernel/... | wc -l` is at or below the target-list count; wowsociety's
  build and full identity/authz test suite run green against the new `foundation/mfa` path or the
  shim.
- **AC-W05-E05-04**: Both stories have passed independent review per mandate §14, given FBL-01's own
  status as "the largest single architectural correction" that "must precede v1 stabilisation," and
  given the `kernel/mfa` shim's auth-critical, security-sensitive nature.

## Closure conditions

Both stories reach `accepted`; AC-W05-E05-01 through AC-W05-E05-04 above are all satisfied;
`closure-report.md` for this epic is completed with reviewer conclusion and acceptance date;
wowsociety's `kernel/mfa` migration coordination (PROD-02) is recorded with a clear pointer for the
product-side migration, not silently treated as this epic's own responsibility to execute; the
mismatch between "framework-side re-home complete" and "wowsociety has actually migrated off the
shim" is explicitly not required for this epic's own closure (the shim's whole purpose is to decouple
these two events), and is recorded as such at closure.
