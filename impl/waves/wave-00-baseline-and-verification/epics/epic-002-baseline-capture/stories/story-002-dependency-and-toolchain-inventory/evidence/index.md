---
id: W00-E02-S002-EVIDENCE-INDEX
type: evidence-index
parent_story: W00-E02-S002
status: complete
created_at: 2026-07-12
updated_at: 2026-07-13
derived: false
---

# Evidence index — W00-E02-S002

Per mandate §10. Evidence produced 2026-07-13 at commit
`0a31186cada5c275a588c74081cf977adf346e61` (branch `main`). Category subdirectories `logs/` and
`reviews/` were created on first real content per Adaptation 2
(`impl/governance/naming-conventions.md`).

| Evidence ID | Evidence type | Story and task | Acceptance criteria proven | Execution command | Code revision or commit SHA | Branch or tag | Execution environment | Relevant tool versions | Date and time | Result | File or URI | Checksum | Reviewer | Superseded evidence |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| EV-W00-E02-S002-001 | dependency-scan / command-output | W00-E02-S002 / T001 | AC-W00-E02-S002-01, AC-W00-E02-S002-02 | `go list -m all`; `go mod graph`; `go list -m -json all`; `go mod why -m <pkg>` | 0a31186cada5c275a588c74081cf977adf346e61 | main | macOS 26.5.2 (Darwin 25.5.0), arm64, local workstation; concurrent sibling-worker test load present (non-timing evidence) | go1.26.5 darwin/arm64 | 2026-07-13 | pass — 340-line build list, 715-edge graph captured; targeted provenance checks recorded | `evidence/logs/go-list-m-all.txt`, `evidence/logs/go-mod-graph.txt`, `evidence/logs/go-list-m-json-all.txt`, `evidence/logs/go-mod-why.txt` | n/a | W00ReviewGate (independent reviewer agent); accepted by conductor 2026-07-13 | none |
| EV-W00-E02-S002-002 | review report (cross-check table) | W00-E02-S002 / T001 | AC-W00-E02-S002-01, AC-W00-E02-S002-02 | manual enumeration of go.mod:8–20 against REVIEW §L (lines 285–287) / §M (lines 289–294) | 0a31186cada5c275a588c74081cf977adf346e61 | main | macOS 26.5.2 (Darwin 25.5.0), arm64, local workstation | go1.26.5 darwin/arm64 | 2026-07-13 | pass — 13/13 direct deps `approved`, zero drift; §M rejected deps absent; new-approval trio explicitly recorded | `evidence/reviews/dependency-crosscheck.md` (tables in `artifacts/post-implementation/dependency-inventory.md`) | n/a | W00ReviewGate (independent reviewer agent); accepted by conductor 2026-07-13 | none |
| EV-W00-E02-S002-003 | command-output / static-analysis | W00-E02-S002 / T002 | AC-W00-E02-S002-03 | `golangci-lint version`; `goreleaser --version`; `trivy --version`; `go version`; inspection of `Makefile:16,19,344–362`, `.github/workflows/ci.yml:58–65,168`, `.github/workflows/release.yml:38–50`, `.github/workflows/security-scan.yml:68–75`, `.github/workflows/vuln.yml:28–31`, `go.mod:13`, `deployments/compose.yaml` | 0a31186cada5c275a588c74081cf977adf346e61 | main | macOS 26.5.2 (Darwin 25.5.0), arm64, local workstation | go1.26.5 darwin/arm64; golangci-lint 2.11.4; goreleaser v2.16.0 (local); trivy 0.72.0 (local) | 2026-07-13 | pass — golangci-lint pin v2.11.4 confirmed (local binary matches); GoReleaser: no exact binary pin (action SHA-pinned v7.2.3, `~> v2`); Trivy: trivy-action SHA-pinned v0.36.0, no explicit binary pin, scanners vuln/secret/misconfig CRITICAL/HIGH non-blocking; goose/v3 v3.27.2 | `evidence/logs/tool-versions.txt` (analysis in `artifacts/post-implementation/tool-version-inventory.md`) | n/a | W00ReviewGate (independent reviewer agent); accepted by conductor 2026-07-13 | none |

All three evidence records above are complete per `evidence-policy.md` (all mandatory fields
populated, commit-pinned). Reviewer sign-off pending the conductor's acceptance gate.
