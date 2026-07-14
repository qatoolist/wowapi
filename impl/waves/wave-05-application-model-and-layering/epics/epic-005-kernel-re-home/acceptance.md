---
id: W05-E05-ACCEPTANCE
type: epic-acceptance
epic: W05-E05
wave: W05
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W05-E05 — Epic-level acceptance

## AC-W05-E05-01 — Foundation tree and forwarding shim in place

The `foundation/` tree exists with all nine packages moved via `git mv` (preserving history); import
paths are updated repo-wide; the `kernel/mfa` deprecated forwarding shim (type aliases + var
forwarding) is in place. Traces to W05-E05-S001.

## AC-W05-E05-02 — Layering enforcement extended

The extended `depguard` rule denies `kernel → foundation` and `foundation → app` imports;
`scripts/lint_boundaries.sh`'s extended allowlist fails CI on a new, un-allowlisted kernel package;
the boundary-lint fixture that fails today against the nine packages is confirmed to pass after the
re-home. Traces to W05-E05-S001.

## AC-W05-E05-03 — Package-count and wowsociety-suite acceptance bars proven

`go list ./kernel/... | wc -l` is at or below the target-list count; wowsociety's build and full
identity/authz test suite run green against the new `foundation/mfa` path or the shim. Traces to
W05-E05-S002.

## AC-W05-E05-04 — Independent review passed

Both stories have passed independent review per mandate §14, given FBL-01's status as "the largest
single architectural correction" and the `kernel/mfa` shim's auth-critical nature.

## Acceptance authority

Framework architecture lead, per `../../wave.md`'s wave-level acceptance authority.
