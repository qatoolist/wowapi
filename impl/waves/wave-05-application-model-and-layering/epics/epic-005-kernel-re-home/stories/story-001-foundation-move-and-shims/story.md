---
id: W05-E05-S001
type: story
title: Foundation tree, package moves, and mfa forwarding shim
status: planned
wave: W05
epic: W05-E05
owner: unassigned
reviewer: unassigned
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-12
source_requirements:
  - FBL-01
depends_on: [W05-E01-S001, W05-E01-S002, W05-E01-S003, W05-E01-S004, W05-E02-S001, W05-E02-S002, W05-E02-S003]
blocks:
  - W05-E05-S002
acceptance_criteria:
  - AC-W05-E05-S001-01
  - AC-W05-E05-S001-02
  - AC-W05-E05-S001-03
artifacts: []
evidence: []
decisions: []
risks:
  - RISK-W05-004
---

# W05-E05-S001 — Foundation tree, package moves, and mfa forwarding shim

## Story ID

W05-E05-S001

## Title

Foundation tree, package moves, and mfa forwarding shim

## Objective

Create the `foundation/` tree; `git mv` all nine app-foundation/adapter packages
(`webhook, notify, document, artifact, attachment, comment, bulk, integration, mfa`) out of
`kernel/`, updating import paths repo-wide; ship a deprecated forwarding shim (type aliases + var
forwarding) at `kernel/mfa` for one minor version; and extend `depguard` and
`scripts/lint_boundaries.sh` to enforce the corrected layering going forward.

## Value to the framework

This story is FBL-01's own mechanics — the actual move, per MATRIX CS-01's own 5-step "Fix
(mechanics, not just intent)" list. Without it, the kernel's public surface remains polluted by
nine unrelated subsystems, and every future kernel-version bump would continue dragging their churn
with it — the exact problem MATRIX CS-01 names as the reason this correction "must precede v1
stabilisation."

## Problem statement

MATRIX CS-01's own "Fix (mechanics, not just intent)" list: "(1) create `foundation/` tree; (2)
`git mv` each package, update import paths repo-wide (mechanical, 8 of 9 are zero-consumer outside
wowapi); (3) `kernel/mfa` → `foundation/mfa` with a deprecated forwarding shim (type aliases + var
forwarding) left at `kernel/mfa` for one minor version so wowsociety migrates on its own schedule,
then remove; (4) extend `depguard` (`.golangci.yml` kernel rule) to deny `kernel → foundation`
imports and add a `foundation` rule denying `foundation → app`; (5) extend
`scripts/lint_boundaries.sh` allowlist so a new kernel package addition fails CI without an explicit
allowlist edit (review-forcing)." MATRIX CS-01's own current evidence: "`go list ./kernel/...` = 39
sub-packages (40 incl. root; personally verified)... wowsociety imports exactly one re-home
candidate: `kernel/mfa` (5 files, `internal/modules/identity/`) — grep-verified; the other 8 = 0
imports."

## Source requirements

FBL-01 (CS-01 mechanics steps 1-5).

## Current-state assessment

Per MATRIX CS-01's own evidence, all nine packages currently live at `kernel/` import paths;
`kernel/storage` (a correct port) also lives there and is not re-homed. No `foundation/` tree
exists yet. No forwarding shim exists. The current `depguard`/`lint_boundaries.sh` configuration
does not yet deny `kernel → foundation` imports (since `foundation/` does not yet exist) and does
not yet enforce a review-forcing allowlist for new kernel packages. This story's own
re-confirmation step is to re-run `go list ./kernel/...` and the repo-wide `kernel/mfa` import
search (across both wowapi-internal and wowsociety) at this story's actual start commit, confirming
MATRIX CS-01's own evidence still holds.

## Desired state

The `foundation/` tree exists with all nine packages moved via `git mv` (preserving history); every
import path referencing the old `kernel/<pkg>` location is updated to `foundation/<pkg>` repo-wide.
`kernel/mfa` retains a deprecated forwarding shim (type aliases + var forwarding onto
`foundation/mfa`) for one minor version. The extended `depguard` rule denies `kernel → foundation`
imports and a new `foundation` rule denies `foundation → app` imports. The extended
`scripts/lint_boundaries.sh` allowlist fails CI on a new, un-allowlisted kernel package addition.

## Scope

- Creating the `foundation/` tree.
- `git mv` for all nine packages, with repo-wide import-path updates.
- The `kernel/mfa` deprecated forwarding shim.
- The `depguard` extension (deny `kernel → foundation`; add `foundation → app` denial).
- The `scripts/lint_boundaries.sh` allowlist extension (review-forcing for new kernel packages).

## Out of scope

- **The verification acceptance bar itself (package count, wowsociety suite green)** — S002's own
  scope, evaluated after this story's move and shim are in place.
- **`kernel/storage`'s own move** — stays, per MATRIX CS-01's own explicit statement.
- **Any behavioral change to the nine packages' own logic** — this is a behaviour-preserving move,
  per MATRIX CS-01's own framing.
- **wowsociety's own migration off `kernel/mfa`** — PROD-02, product-level coordination, out of
  framework scope.

## Assumptions

- MATRIX CS-01's own "8 of 9 are zero-consumer outside wowapi" (i.e. wowapi-internal-only, aside
  from `kernel/mfa`) is taken as confirmed evidence requiring this story's own re-confirmation via
  repo-wide search at the actual start commit, not as a permanently-fixed fact.
- The forwarding shim's exact mechanics (type aliases + var forwarding) is taken directly from
  MATRIX CS-01's own step-3 language — this story's own implementation follows that pattern, not an
  invented alternative shim mechanism.

## Dependencies

Depends on W05-E01 (full epic) and W05-E02 (full epic) at epic scope. No dependency within W05-E05
(this is the epic's first story). Blocks W05-E05-S002 (verification cannot meaningfully evaluate its
own acceptance bar before this story's move and shim exist).

## Affected packages or components

All nine re-homed packages and every file repo-wide that imports any of them; `.golangci.yml`
(depguard rule); `scripts/lint_boundaries.sh`.

## Compatibility considerations

The `kernel/mfa` forwarding shim is this story's central compatibility mechanism — it exists
specifically so wowsociety's own migration timing is decoupled from this story's own landing. The
other 8 packages have zero external consumers and require no compatibility shim of their own.

## Security considerations

The `kernel/mfa` shim is security-sensitive per REVIEW §P's own framing: "TOTP/OTP identity code on
wowsociety's auth path." The shim's correctness (forwarding calls transparently, not silently
altering behavior) is a required property, not merely a convenience.

## Performance considerations

None material — this is a compile-time/import-path change, not a runtime behavior change.

## Observability considerations

None material beyond standard CI/build reporting for the lint extensions.

## Migration considerations

This story is itself a large-scale code migration (import-path level), though it is explicitly not
a database/schema migration. `git mv`'s history-preservation is a required property, per the
mandate's own general preference for traceable history (not explicitly named in the source, but
consistent with good practice for a move of this scale).

## Documentation requirements

Document the `foundation/` tree's existence and its four-level-architecture rationale (referencing
REVIEW §J); document the `kernel/mfa` shim's deprecation timeline (one minor version); document the
extended depguard/boundaries-lint rules.

## Acceptance criteria

- **AC-W05-E05-S001-01**: The `foundation/` tree exists with all nine packages moved via `git mv`
  (history preserved); every import path is updated repo-wide; a full build succeeds.
- **AC-W05-E05-S001-02**: The `kernel/mfa` deprecated forwarding shim is in place (type aliases + var
  forwarding), proven by a test confirming calls through the shim behave identically to calls
  through `foundation/mfa` directly.
- **AC-W05-E05-S001-03**: The extended `depguard` rule denies `kernel → foundation` and
  `foundation → app` imports (proven by an adversarial fixture attempting each); the extended
  `scripts/lint_boundaries.sh` allowlist fails CI on a new, un-allowlisted kernel package addition
  (proven by an adversarial fixture).

## Required artifacts

- The `foundation/` tree (code, moved).
- The `kernel/mfa` forwarding shim (code).
- The extended `depguard` configuration.
- The extended `scripts/lint_boundaries.sh` allowlist.
See `artifacts/index.md`.

## Required evidence

- Full build success output post-move.
- The forwarding-shim behavioral-equivalence test output.
- The depguard adversarial-fixture output (both denial rules).
- The boundaries-lint adversarial-fixture output.
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`: `story.md`
and `plan.md` complete, acceptance criteria numbered and measurable, dependency on W05-E01/E02
recorded, owner/reviewer assignment pending.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; all three acceptance
criteria verified with evidence in `evidence/index.md`; `closure.md` completed; independent review
passed per mandate §14, specifically confirming the `kernel/mfa` shim's behavioral equivalence is
genuinely proven, given its auth-critical, security-sensitive status.

## Risks

RISK-W05-004 (the `kernel/mfa` re-home's auth-critical, security-sensitive nature) — see epic-level
`risks.md` for full detail and mitigation/contingency.

## Residual-risk expectations

Residual risk is expected to be low for the 8 mechanical package moves once the full build passes;
medium for the `kernel/mfa` shim until its behavioral-equivalence test is confirmed and independently
re-checked by this story's own review task.

## Plan

See `plan.md`.

## Note (autopsy remediation R-1, 2026-07-16)

Status is unchanged — this story remains genuinely unexecuted as tracked (`planned`, all tasks
`todo`). However, the implementation-autopsy report
(`impl/reports/implementation-autopsy-report-2026-07-16.md`, §4 row W05-E05-S001, independent
verdict **contradictory**) found that the FBL-01 kernel re-home this story describes is, in
substance, ALREADY DONE AND WIRED on `main`: all nine packages are under `foundation/` with shims
in place, executed in commit `e8cda6b` entirely outside this story's tracked execution (autopsy
H-7). See deviation **DEV-PROG-002** in `impl/tracking/programme-deviations.md` for the full
record. — autopsy remediation R-1, 2026-07-16.
