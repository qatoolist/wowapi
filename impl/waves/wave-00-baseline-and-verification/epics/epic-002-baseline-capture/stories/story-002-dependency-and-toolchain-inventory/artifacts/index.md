---
id: W00-E02-S002-ARTIFACTS-INDEX
type: artifact-index
parent_story: W00-E02-S002
status: complete
created_at: 2026-07-12
updated_at: 2026-07-13
derived: false
---

# Artifact index — W00-E02-S002

Per mandate §9.2. Both artifacts were produced 2026-07-13 at commit
`0a31186cada5c275a588c74081cf977adf346e61`; the `post-implementation/` subdirectory was created on
first real content per Adaptation 2 (`impl/governance/naming-conventions.md`).

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Repository path or storage location | Version | Checksum | Status | Reviewer | Retention requirement |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| ART-W00-E02-S002-001 | Dependency inventory | design document | post-implementation | `go.mod` direct/indirect dependency list with per-direct-dependency disposition against REVIEW §L/§M; "10 vs 13" reconciliation; §M rejected-register absence check; new-approval trio presence/absence | AC-W00-04, AC-W00-E02-03 | W00-E02-S002-T001 | `artifacts/post-implementation/dependency-inventory.md` | 1.0 (commit 0a31186) | n/a (markdown, git-tracked) | produced | unassigned | Retain for the life of the programme; superseding versions must not overwrite — see `evidence-policy.md` preservation rule, applied here by analogy for the artifact's own revision history |
| ART-W00-E02-S002-002 | Tool-version inventory | design document | post-implementation | Pinned versions of `golangci-lint` (v2.11.4), GoReleaser (no exact binary pin; action SHA-pinned `~> v2`), Trivy (trivy-action v0.36.0 SHA-pinned, scanner config recorded), `goose/v3` (v3.27.2) with file:line citations | AC-W00-04, AC-W00-E02-03 | W00-E02-S002-T002 | `artifacts/post-implementation/tool-version-inventory.md` | 1.0 (commit 0a31186) | n/a (markdown, git-tracked) | produced | unassigned | Retain for the life of the programme |

Both artifacts exist and are registered. Reviewer sign-off pending the conductor's acceptance
gate.
