# Design: Enterprise-grade CI/CD & GitHub Actions for wowapi

- **Date:** 2026-07-05
- **Status:** Implemented — `.github/workflows/` ships the workflows described here (validated by `actionlint` + green CI).
- **Author:** framework maintainer (qatoolist)
- **Scope:** Replace the single "newbie" `ci.yml` with an enterprise CI/CD posture: security scanning,
  supply-chain integrity, signed releases, dependency hygiene, and repo governance.

## Goal
Make wowapi's automation match its production ambitions: every PR is scanned and gated; every dependency and
action is tracked and pinned; every tagged version ships as signed, attested, SBOM-bearing artifacts; and the
repo carries the governance files an enterprise consumer expects from a framework they adopt as a dependency.

## Decisions (from brainstorming)
- **Pillars:** all four — security & scanning, release & artifacts, dependency & CI hardening, repo governance.
- **License:** Apache-2.0 (+ NOTICE).
- **Release artifacts:** full supply chain — cross-platform CLI binaries + GHCR container image, each with SBOM,
  cosign **keyless** signature, and SLSA provenance attestation; auto-generated release notes.
- **Registry:** `ghcr.io/qatoolist/wowapi`.
- **Secrets:** none. All signing/attestation/registry auth uses GitHub OIDC + `GITHUB_TOKEN`.
- **Branch protection / required checks / code-scanning enablement:** documented as manual repo-admin steps, not
  applied by this change.

## Non-goals (YAGNI)
- Stale bot (active solo repo; low value).
- Renovate (Dependabot covers it).
- release-please (GoReleaser + conventional commits covers changelog/notes).
- Go version matrix (module requires Go 1.26; older toolchains fail on 1.26 features). GoReleaser still
  cross-compiles all target OS/arch.

## Architecture

### Cross-cutting conventions (every workflow)
- Top-level least-privilege `permissions:` (default `contents: read`; widen per-job only where needed).
- `concurrency` group with `cancel-in-progress` on PR-ish triggers.
- Per-job `timeout-minutes`.
- Every third-party action pinned to a **full commit SHA** with a `# vX.Y.Z` comment (Dependabot `github-actions`
  ecosystem keeps them current).
- Go set up via `actions/setup-go` with `go-version: '1.26'` and build/module cache enabled.

### Workflows (`.github/workflows/`)
| File | Trigger | Jobs / purpose |
|---|---|---|
| `ci.yml` (refactor) | push `main`, PR | `workflow-lint` (actionlint) · `unit` (fmt-check, vet, lint-new, tidy-check, boundaries, unit tests, build) · `gate` (ci-container: real-DB suite + race + fuzz seeds) · `coverage` (profile artifact + threshold) |
| `codeql.yml` | push `main`, PR, weekly cron | CodeQL SAST for Go → Security tab |
| `vuln.yml` | push `main`, PR, daily cron | `govulncheck ./...` (Go vuln DB, call-graph aware) |
| `security-scan.yml` | PR, push `main`, weekly cron | `gitleaks` (secret scan, full history on schedule) · `trivy` fs/config scan → SARIF · `dependency-review` (PR-only, blocks vulnerable/incompatible deps) |
| `scorecard.yml` | push `main`, weekly cron | OpenSSF Scorecard → SARIF + badge |
| `pr.yml` | PR opened/edited/synchronize | conventional-commit **PR-title lint** · path-based `labeler` |
| `release.yml` | push tag `v*` | supply-chain release (below) |

Split rather than one mega-file so each has independent required-check status, schedule, and blast radius.

### Release & supply-chain pipeline (`release.yml` + `.goreleaser.yaml`)
On a `v*` tag:
1. **GoReleaser** builds `wowapi` (`./cmd/wowapi`) for `linux|darwin|windows × amd64|arm64`, `CGO_ENABLED=0`,
   `-trimpath -ldflags "-s -w -X github.com/qatoolist/wowapi/internal/buildinfo.version={{.Version}}"`;
   produces archives, `checksums.txt`, **per-artifact SBOM (syft)**, and **cosign keyless** signatures of the
   checksums + artifacts (Fulcio cert + Rekor log; no private keys).
2. **Container image:** `docker buildx` multi-arch (amd64/arm64) from the existing distroless `cli` stage →
   `ghcr.io/qatoolist/wowapi:{version}` + `:latest`, `VERSION` build-arg wired to the tag.
3. **SLSA provenance:** GitHub-native `actions/attest-build-provenance` for the binary archives and for the image
   digest; `actions/attest-sbom` attaches the image SBOM. Attestations land in the Sigstore transparency log.
4. **GitHub Release** with auto-notes grouped by conventional-commit type; archives, checksums, `.sig`/`.pem`,
   SBOMs attached.

Permissions: `contents: write`, `packages: write`, `id-token: write`, `attestations: write` — all `GITHUB_TOKEN`.

### Governance & dependency hygiene
- `LICENSE` (Apache-2.0) + `NOTICE`.
- `.github/dependabot.yml` — `gomod`, `github-actions`, `docker`; weekly; grouped minor/patch.
- `CODEOWNERS` (@qatoolist as global owner).
- `SECURITY.md` — private reporting (GitHub Security Advisories), supported-versions table, response SLA.
- `CONTRIBUTING.md` — wires to existing `make check` / `docs/working/quality-gates.md`; conventional-commit rule;
  documents the manual branch-protection / required-checks / code-scanning enablement steps.
- `CODE_OF_CONDUCT.md` — Contributor Covenant 2.1.
- `.github/ISSUE_TEMPLATE/` — bug + feature + config issue forms (+ `config.yml` linking Security/Discussions).
- `.github/PULL_REQUEST_TEMPLATE.md`.
- `.github/labeler.yml` — path globs (kernel, module, cli, docs, ci, deps).
- **Makefile:** add `actionlint`, `govulncheck`, `goreleaser-check` targets; fold `actionlint` + `govulncheck`
  into `make check` where cheap.

## Validation (the "tests")
- `actionlint` passes on all workflows (local + a CI job — dogfooded).
- `goreleaser check` validates config; `goreleaser release --snapshot --clean` dry-runs the full build (no
  publish) to prove binaries + SBOM + image build before any real tag exists.
- Existing `make ci` still green.
- Independent third-party review gate before declaring done.

## Manual follow-ups (documented in CONTRIBUTING.md, not applied here)
- Enable branch protection on `main` with required checks: `unit`, `gate`, `workflow-lint`, `CodeQL`,
  `govulncheck`, PR-title lint.
- Enable GitHub code scanning + secret scanning + push protection.
- Enable Dependabot security updates.
- First tag (`v0.2.0`) exercises `release.yml` end-to-end.

## Post-review decisions (2026-07-05)
An independent review flagged that the repo is **private without GHAS**, so CodeQL, SARIF upload, Scorecard
publish, dependency-review, and public attestation verification would fail at runtime. Decision: **keep the repo
private and add guards** rather than go public or buy GHAS. Each affected workflow gained a `visibility guard`
job (runtime `gh api` visibility check) that skips the GHAS-only steps while private and activates them
automatically if the repo becomes public. Trivy runs in log-only (`exit-code: 0`) mode without SARIF upload;
gitleaks + govulncheck gate directly. README shows only CI + License badges while private (others staged in a
comment). SECURITY.md verification corrected to the `checksums.txt` + `gh attestation verify` flow (GoReleaser
signs only the checksums file, not each archive).

## Risks
- Wrong pinned SHA breaks a workflow → mitigate by resolving each SHA from `gh api` at implementation time and
  running `actionlint`.
- GoReleaser multi-arch image build needs `buildx`/QEMU in the runner → include the setup steps.
- Keyless signing requires `id-token: write` + public repo or GHAS → note in SECURITY.md.
