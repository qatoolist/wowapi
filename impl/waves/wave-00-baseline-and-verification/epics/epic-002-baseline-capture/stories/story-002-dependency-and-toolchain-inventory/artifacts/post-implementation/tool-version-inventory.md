---
id: ART-W00-E02-S002-002
type: artifact
title: Tool-version inventory — golangci-lint, GoReleaser, Trivy, goose/v3
lifecycle_stage: post-implementation
parent_story: W00-E02-S002
producing_task: W00-E02-S002-T002
status: produced
created_at: 2026-07-13
updated_at: 2026-07-13
commit_sha: 0a31186cada5c275a588c74081cf977adf346e61
---

# Tool-version inventory — W00-E02-S002-T002

Captured 2026-07-13 at commit `0a31186cada5c275a588c74081cf977adf346e61` (branch `main`),
environment: macOS 26.5.2 (Darwin 25.5.0), arm64, `go version go1.26.5 darwin/arm64`. All
versions read directly from this repository's own configuration files at this commit (citations
below), not copied from any secondhand document. Local-binary outputs in
`../../evidence/logs/tool-versions.txt` (EV-W00-E02-S002-003).

## Primary tools (the four required by AC-W00-E02-S002-03)

### 1. golangci-lint — **CONFIRMED: v2.11.4 (pinned)**

- Configured pin: `Makefile:16` — `GOLANGCI_VERSION ?= v2.11.4` (declared "single source of
  truth" at Makefile:12–15).
- CI lockstep pin: `.github/workflows/ci.yml:62` — `GOLANGCI_VERSION: "v2.11.4"`; installed at
  ci.yml:168 via `go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$GOLANGCI_VERSION`.
- Locally installed binary cross-check: `golangci-lint version` → `2.11.4 built with go1.26.1`
  — **matches the pin, no mismatch** (`tool-versions.txt`).
- The secondhand `v2.11.4` citation in `wave.md`/`dependencies.md` is hereby independently
  re-confirmed against the Makefile at this commit (plan.md unresolved question resolved).

### 2. GoReleaser — **CONFIRMED: no exact version pin; action SHA-pinned with floating `~> v2` range (recorded as found)**

- Authoritative release path: `.github/workflows/release.yml:47–50` —
  `goreleaser/goreleaser-action@f06c13b6b1a9625abc9e6e439d9c05a8f2190e94 # v7.2.3` with
  `distribution: goreleaser`, `version: "~> v2"`. The **action** is SHA-pinned (v7.2.3) but the
  GoReleaser **binary** version is a floating major-2 range (`~> v2`), i.e. no exact binary pin
  exists in this repository — recorded as a fact, per task step 3, not a judgment.
- Local convenience targets deliberately unpinned: `Makefile:344–348` (comment) and
  `Makefile:356`, `Makefile:361` — `go install github.com/goreleaser/goreleaser/v2@latest`,
  with the rationale "the authoritative release build is pinned in release.yml (SHA-pinned
  goreleaser-action + version ~> v2)".
- Locally installed binary (informational only, not a repo pin): `GitVersion: v2.16.0`
  (`tool-versions.txt`).
- Resolves the "GoReleaser pinned version TBD" question from story.md/plan.md; downstream
  consumer `W06-E03-S001` (REL-01 T6) should cite this finding.

### 3. Trivy — **CONFIRMED: action pinned at trivy-action v0.36.0 (SHA-pinned); no explicit Trivy binary version pin**

- `.github/workflows/security-scan.yml:68` —
  `aquasecurity/trivy-action@ed142fd0673e97e23eac54620cfb913e5ce36c25 # v0.36.0`. The action is
  SHA-pinned; no explicit `version:`/`trivy-version` input is set, so the Trivy binary version is
  whatever trivy-action v0.36.0 bundles/defaults to — recorded as found.
- Scanner configuration (security-scan.yml:70–75): `scan-type: fs`;
  `scanners: vuln,secret,misconfig`; `format: table`; `severity: CRITICAL,HIGH`;
  `ignore-unfixed: true`; `exit-code: "0"` (informational, non-blocking — rationale comment at
  lines 64–67: Go vulns gated by govulncheck, secrets by gitleaks; flip to blocking once public).
- Locally installed binary (informational only, not a repo pin): `Version: 0.72.0`
  (`tool-versions.txt`).
- Resolves the "Trivy pinned version + scanner configuration TBD" question from story.md/plan.md.

### 4. goose/v3 — **CONFIRMED: v3.27.2**

- `go.mod:13` — `github.com/pressly/goose/v3 v3.27.2` (Go module dependency; also row 6 of the
  dependency-inventory disposition table, disposition `approved`).

## Supporting toolchain pins observed during inspection (context, same commit)

| Tool | Version | Source citation |
|---|---|---|
| Go toolchain | `go 1.26` / `toolchain go1.26.5` | `go.mod:3,5`; `ci.yml:58` `GO_VERSION: "1.26.5"`; `release.yml:40`, `vuln.yml:28` `go-version: "1.26.5"` |
| actionlint | v1.7.12 (pinned) | `Makefile:19`; `ci.yml:65` |
| govulncheck | `@latest` — deliberately unpinned (tracks newest vuln DB) | `Makefile:344–346`; `vuln.yml:31` |
| docker compose images | postgres:16-alpine · minio/minio:latest · axllent/mailpit:latest · jaegertracing/all-in-one:1.57 · neo4j:5-community | `deployments/compose.yaml:7,23,40,49,59` |

## Conclusion

All four required tools addressed with an explicit outcome — none omitted:
golangci-lint **confirmed pinned v2.11.4**; GoReleaser **no exact binary pin (action SHA-pinned,
`~> v2` floating range) — recorded as found**; Trivy **no explicit binary pin (trivy-action
SHA-pinned v0.36.0, scanners vuln/secret/misconfig, CRITICAL/HIGH, non-blocking) — recorded as
found**; goose/v3 **confirmed v3.27.2**.
