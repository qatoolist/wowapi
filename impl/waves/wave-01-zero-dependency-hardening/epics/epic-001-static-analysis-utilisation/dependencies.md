---
id: W01-E01-DEPS
type: epic-dependencies
epic: W01-E01
wave: W01
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W01-E01 — Dependencies

## Upstream (epics/waves this epic depends on)

- **W00** (full wave) — per `../../dependencies.md` (wave-level), W01 depends on W00's exit gate:
  the 8 executed finding-slices re-verified at current HEAD, baselines captured (including the
  current lint hit-count baseline this epic's stories re-confirm rather than blindly trust — see
  RISK-W01-001), D-01..D-09 ratified as ADRs. This epic has no additional upstream dependency beyond
  the wave-level W00 gate — none of FBL-05/FBL-07's scope touches AR-01/AR-02/SEC-01/DATA-09.

## Downstream (epics/waves that depend on this epic)

| Downstream item | Depends on (from this epic) | Why |
|---|---|---|
| W06-E02 (DX-06/REL-03) | W01-E01-S001 | REL-03's compatibility gates assume the raw `pgx.Rows` public-contract decision (CS-10) is settled, which S001 finalizes mechanically by enabling sqlclosecheck/rowserrcheck against it. |
| All later waves' CI runs | W01-E01-S001/S002/S003 | Every later wave's PR/CI gate runs against the linter/supply-chain configuration this epic lands — a later wave's evidence is only comparable if this epic's gate state is the stable baseline (wave-level `dependencies.md` table, reproduced here at epic scope since this epic is the entirety of that table's linter/supply-chain row). |
| W01-E03-S001 (FBL-09, `http-hardening` epic) | W01-E01-S002 (cross-reference only) | S002's gosec triage explicitly excludes G120 (`kernel/httpx/csrf.go:118`) and cross-references it to FBL-09's own fix — FBL-09 must not silently assume G120 is already handled by this epic. |

## Internal (within this epic)

S001, S002, and S003 target disjoint files and configuration surfaces:

- S001 — `.golangci.yml` (zero-cost analyzer block), `internal/cli/config_delegate.go`,
  `internal/cli/lint_cmd.go`, `app/maintenance.go`, `kernel/config/config.go`,
  `kernel/database/txmanager.go`-adjacent pool-config wiring.
- S002 — `.golangci.yml` (judged analyzer block), `kernel/auth/jwks.go`, `kernel/config/bind.go`,
  `kernel/workflow/definition.go`, `kernel/workflow/runtime.go`, `kernel/httpx/middleware.go`,
  plus per-site annotations across audit/database/jobs/mfa/pagination packages (G115 set, enumerated
  at implementation time).
- S003 — `.github/workflows/ci.yml`, `.github/workflows/security-scan.yml`, the pre-push hook script
  (path to be located at implementation time — distinct from `.githooks/pre-commit`).

No file-level or ordering dependency exists between the three stories; they may execute in any order
or in parallel. The only shared surface is `.golangci.yml`, and S001/S002 edit disjoint blocks within
it (zero-cost analyzer list vs. judged analyzer list) — a merge conflict is possible if both stories
edit the file concurrently without rebasing, but there is no semantic ordering dependency.

## Cross-wave dependencies

None beyond the W00→W01 entry dependency stated above.

## External dependencies

- `golangci-lint` v2.11.4 pinned toolchain (already installed per W00's toolchain inventory) — all
  25 analyzers this epic enables or triages already ship in this binary; no new tool install required.
- Trivy (already wired in `security-scan.yml`) for S003's license-scanning-signal choice, if that
  choice is Trivy's license scanner rather than `go-licenses`.

## Repository dependencies

None cross-repo. All three stories are wowapi-internal (lint config, CI config, hook script,
kernel/CLI code). wowsociety impact is optional/additive only: the new
`MaxConnLifetime`/`MaxConnIdleTime` config keys (S001) are available for wowsociety to adopt but not
required.

## Tooling dependencies

None beyond the already-installed golangci-lint/Trivy/Go-toolchain dependencies listed above.

## Decision dependencies

None. See `epic.md` "Required decisions" — CS-10 is already a decided, closed question this epic
enforces, not re-litigates.
