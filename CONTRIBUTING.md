# Contributing to wowapi

Thanks for your interest. wowapi is a **domain-neutral platform kernel** consumed as a versioned Go
dependency. Contributions must preserve that neutrality and the framework's security/isolation invariants.

By contributing you agree your work is licensed under the project's [Apache-2.0 license](LICENSE).

## Ground rules

- **Domain-neutrality is non-negotiable.** The kernel must not learn about any product domain
  (society/school/club/etc.). `make lint-boundaries` enforces the import law and a term denylist.
- **Follow the existing architecture.** Import law: `product modules → module → kernel`; `kernel` imports no
  `module`/`app`/`adapters`/`testkit`. See [`docs/SRS.md`](docs/SRS.md) §6 and `docs/blueprint/`.
- **Real tests over mocks.** Fakes only at process/network boundaries. New security controls need a structural
  enforcement **and** a test.
- **Conventional Commits** for commit messages *and* PR titles (a CI check enforces PR titles).

## Development setup

```bash
make setup     # installs golangci-lint, git hooks, downloads modules
make up        # starts Postgres + MinIO + Mailpit (docker compose)
```

The versioned git hooks (`.githooks/`, installed by `make setup`/`make hooks`) run fast format + lint checks on
commit and a fuller gate on push.

## The quality gates

Run before every push — this mirrors CI:

```bash
make check     # fmt-check + vet + lint-new + tidy-check + unit tests (fast pre-flight)
make ci        # authoritative correctness gate (vet, boundaries, unit, race, bench, build)
make ci-container  # the above against a real Postgres (WOWAPI_REQUIRE_DB=1) — what the CI gate runs
```

Additional local checks used by CI:

```bash
make actionlint       # lint GitHub Actions workflows
make govulncheck      # Go vulnerability scan
make goreleaser-check # validate the release config
```

See [`docs/working/quality-gates.md`](docs/working/quality-gates.md) and
[`docs/working/lint-backlog.md`](docs/working/lint-backlog.md) for the full culture and the linter burn-down rule
(the enforced gate blocks on **changed** code via `make lint-new`; the full backlog is burned down incrementally).

## Pull request flow

1. Branch from `main`.
2. Write the failing test first where applicable (TDD), then implement.
3. Keep the PR focused; no drive-by refactors outside the task.
4. Ensure `make check` is green and update docs (SRS/tracker/user-guide) if surface or behavior changed.
5. Open the PR with a Conventional-Commit title; fill in the template; link issues.
6. CI must be green and CODEOWNERS review approved before merge.

## CI/CD overview

| Workflow | Purpose |
|---|---|
| `ci.yml` | workflow-lint (actionlint), unit job (no DB), authoritative gate (real DB + race + fuzz), coverage |
| `codeql.yml` | CodeQL SAST |
| `vuln.yml` | govulncheck |
| `security-scan.yml` | gitleaks, Trivy, dependency-review |
| `scorecard.yml` | OpenSSF Scorecard |
| `pr.yml` | Conventional-Commit PR-title lint + path labeler |
| `release.yml` | GoReleaser: signed, attested, SBOM-bearing binaries + GHCR image on `v*` tags |

## Releasing (maintainers)

Releases are tag-driven. To cut `vX.Y.Z`:
```bash
git tag -s vX.Y.Z -m "vX.Y.Z"   # signed tag
git push origin vX.Y.Z          # triggers release.yml
```
`release.yml` runs GoReleaser (cross-platform CLI binaries, SBOMs, cosign keyless signatures), builds and pushes
the multi-arch `ghcr.io/qatoolist/wowapi` image, and attaches SLSA build-provenance attestations. No secrets are
required — it uses GitHub OIDC and `GITHUB_TOKEN`.

## Repository visibility & the scanning stack

This repo is currently **private without GitHub Advanced Security**. Some integrations only work on a public
repo (or with GHAS), so the affected workflows include a `visibility guard` job that detects visibility at
runtime and **skips** those steps while private — they do not fail. What runs where:

| Runs while private | Activates when public (or GHAS enabled) |
|---|---|
| `ci` (all jobs), `govulncheck`, gitleaks, Trivy (log-only), `pr` title+labeler, `release` | CodeQL, Trivy SARIF → Security tab, `dependency-review`, OpenSSF Scorecard + badge, pkg.go.dev badge |

**To go public** (unlocks everything with no code changes):
```bash
gh repo edit qatoolist/wowapi --visibility public --accept-visibility-change-consequences
```
Then uncomment the extra badges in `README.md` and add CodeQL to the required checks below.

## Repository administration (one-time, maintainer/admin)

These cannot be set from committed files and must be enabled in the repo settings/API:

- **Branch protection** on `main` — require PRs and these status checks (exact job names as shown in the Checks
  tab): `unit (no DB)`, `authoritative gate (ci-container + DB + fuzz seeds)`, `workflow lint (actionlint)`,
  `govulncheck`, and `conventional-commit title`; require CODEOWNERS review; require signed commits.
  Add `analyze (go)` (CodeQL) **only after the repo is public/GHAS** — while private it is skipped, and a
  required-but-skipped check would block every merge.
  ```bash
  gh api -X PUT repos/qatoolist/wowapi/branches/main/protection --input protection.json
  ```
- **Code security** (public/GHAS): enable Code scanning (CodeQL), Secret scanning + push protection, and
  Dependabot security updates (Settings → Code security and analysis).
- **Actions**: set workflow permissions to read-only by default (Settings → Actions → General).
- **Discussions**: enable if you want the issue-template "question" link to resolve.
