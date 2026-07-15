---
id: DEV-W06-E03-S001
type: deviations-record
parent_story: W06-E03-S001
status: accepted-exception
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Deviations record — W06-E03-S001

## DEV-W06-E03-S001-001 — ADR-005 publisher command incompatibility

- **Approved plan:** use GoReleaser `release --skip=publish`, then a separate GoReleaser publish invocation so publication cannot rebuild.
- **Observed fact:** focused execution of `go run github.com/goreleaser/goreleaser/v2@v2.17.0 --help` showed that OSS v2.17.0 has no `publish` command. Official Split & Merge is GoReleaser Pro-only; its documented split phase can push images, violating this story's non-publishing candidate boundary.
- **Decision:** Main/release-security authority explicitly authorized a manifest-verified `gh`/ORAS publisher; GoReleaser still creates the candidate once with `--skip=publish`.
- **Compensating controls:** publish re-verifies the GitHub attestation and every manifested hash; it creates a draft release, uploads only attested bytes, copies the already-built OCI layout by immutable digest, runs clean verification, and only then publishes the release and moves aliases. Artifact and image Trivy reports are mandatory manifest subjects.
- **Evidence:** `evidence/tests/goreleaser-split-caveat.txt`, `evidence/tests/release-contracts.txt`, and `.github/workflows/release.yml`.
- **Impact:** deliberate ADR-005 mechanism deviation with the same no-rebuild security invariant; no GoReleaser Pro dependency was introduced.
