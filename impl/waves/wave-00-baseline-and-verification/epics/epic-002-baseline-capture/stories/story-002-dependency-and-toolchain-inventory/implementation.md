---
id: IMPL-W00-E02-S002
type: implementation-record
parent_story: W00-E02-S002
status: complete
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Implementation record — W00-E02-S002

*This record aggregates the implementation reality of the story across both of its tasks.
Executed 2026-07-13 at commit `0a31186cada5c275a588c74081cf977adf346e61` (branch `main`) by the
wave-00 worker; details per task in `tasks/task-001-*.md` and `tasks/task-002-*.md`.*

## What was actually implemented

**T001 (dependency inventory):** captured `go list -m all` (340-line build list), `go mod graph`
(715 edges), `go list -m -json all`, and targeted `go mod why -m` provenance checks; produced the
13-row direct-dependency disposition table against REVIEW §L/§M (all `approved`, zero drift);
resolved the "10 vs 13" question (§L "otel×4" = go.mod:16–19); recorded new-approval trio state
(backoff/v5 present-indirect v5.0.3; golang-lru/v2, gobreaker absent); confirmed §M rejected deps
absent from go.mod; re-confirmed go-retry present-indirect and the yaml.v3 watch item.
**T002 (tool versions):** confirmed golangci-lint pinned v2.11.4 (`Makefile:16`, `ci.yml:62`,
local binary matches); GoReleaser has no exact binary pin (SHA-pinned goreleaser-action v7.2.3,
`version: "~> v2"`, `release.yml:47–50`); Trivy via SHA-pinned trivy-action v0.36.0 with scanners
vuln/secret/misconfig, CRITICAL/HIGH, non-blocking (`security-scan.yml:68–75`); goose/v3 v3.27.2
(`go.mod:13`).

## Components changed

None (documentation artifacts only, as expected).

## Files changed

No existing repository file modified. New files, all inside this story directory:
`artifacts/post-implementation/dependency-inventory.md`,
`artifacts/post-implementation/tool-version-inventory.md`, `evidence/logs/go-list-m-all.txt`,
`evidence/logs/go-mod-graph.txt`, `evidence/logs/go-list-m-json-all.txt`,
`evidence/logs/go-mod-why.txt`, `evidence/logs/tool-versions.txt`,
`evidence/reviews/dependency-crosscheck.md`; plus status/record updates to this story's own
scaffolding files (indexes, task files, verification/deviations/closure records, story.md).

## Interfaces introduced or changed

None.

## Configuration changes

None.

## Schema or migration changes

None.

## Security changes

None (the §M absence cross-check verifies an existing control; nothing new introduced).

## Observability changes

None.

## Tests added or modified

None (no testable code produced).

## Commits

Executed against existing commit `0a31186cada5c275a588c74081cf977adf346e61`; committing these
documentation files is the conductor's PR process.

## Pull requests

None yet — raised by the conductor.

## Implementation dates

2026-07-13 (single session, both tasks).

## Technical debt introduced

None.

## Known limitations

License dispositions rely on REVIEW §L's stated licenses (no license-scanning tool is pinned in
this repository's toolchain); local goreleaser/trivy binary versions are informational
workstation state, not repository pins.

## Follow-up items

None required. Candidate noted for later waves: GoReleaser/Trivy have no exact binary pin
(deliberate per in-repo comments) — owners if that changes: `W06-E03-S001` / `W06-E03-S003`.

## Relationship to the approved plan

Matches `plan.md` exactly — same commands, same two artifacts at the planned paths, all four
unresolved questions answered by measurement. No deviations.
