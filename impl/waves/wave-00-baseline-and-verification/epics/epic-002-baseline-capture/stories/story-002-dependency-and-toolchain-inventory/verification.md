---
id: VER-W00-E02-S002
type: verification-record
parent_story: W00-E02-S002
status: complete
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Verification record — W00-E02-S002

## Planned verification procedure

Per mandate §8.8. One row per acceptance criterion for this story.

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W00-E02-S002-01 | Run `go list -m all`, `go mod graph`, `go list -m -json all` at a named commit SHA; manually enumerate every direct `go.mod` dependency and confirm each has an explicit disposition entry (approved / newly-approved / undocumented drift) against REVIEW §L; zero entries left unaddressed. | Local dev environment or CI runner with network access to the Go module proxy, checked out at the story's execution commit. | Every direct dependency in `go.mod` has exactly one disposition row; zero dependencies are left without a stated disposition; any drift found is explicitly flagged, not silently absorbed. | Command-output evidence (raw `go list`/`go mod graph` output) + cross-check table/diff evidence. | Reviewer (unassigned) |
| AC-W00-E02-S002-02 | Search the captured `go list -m all` output for each of REVIEW §M's named rejected dependencies (viper, envconfig, a NATS/Kafka message-bus client, any password-hashing library) and confirm none are present, direct or indirect; separately confirm presence/absence of `cenkalti/backoff/v5`, `hashicorp/golang-lru/v2`, `sony/gobreaker`, and the `yaml.v3`/`go.yaml.in/yaml` watch item. | Same as above. | REVIEW §M's rejected dependencies are confirmed absent (or, if any is found present, it is flagged as drift requiring escalation, not silently noted); presence/absence of each of the three new-approval packages and the yaml watch item is explicitly recorded either way. | Command-output evidence + cross-check table. | Reviewer (unassigned) |
| AC-W00-E02-S002-03 | Inspect `Makefile`, `.github/workflows/*.yml`, and lint configuration file(s) directly (not via secondhand citation) to confirm pinned versions of `golangci-lint`, GoReleaser, Trivy, `goose/v3`; run `golangci-lint --version` (and equivalent for other tools where a local binary is available) to cross-confirm against the configuration-file pin. | Local dev environment or CI runner with the relevant tool binaries installed, or read-only file inspection where a binary is unavailable. | Each of the four tool versions is either confirmed with a citation to the exact file/line it is pinned at, or explicitly recorded as "no pin found / TBD" if that is what is actually the case — no version is asserted without a source. | Command-output evidence + configuration-file citation evidence. | Reviewer (unassigned) |

## Post-execution record

*Executed 2026-07-13 at commit `0a31186cada5c275a588c74081cf977adf346e61` (branch `main`).*

### Per-AC results

| Acceptance criterion | Actual result | Pass/fail | Evidence IDs |
|---|---|---|---|
| AC-W00-E02-S002-01 | `go list -m all` (340 lines), `go mod graph` (715 edges), `go list -m -json all` captured at commit `0a31186`; all 13 direct require lines (go.mod:8–20) dispositioned individually — 13/13 `approved`, zero `undocumented drift`, zero unaddressed; "10 vs 13" reconciled (§L "otel×4" = go.mod:16–19 → 10 logical deps) | **pass** | EV-W00-E02-S002-001, EV-W00-E02-S002-002 |
| AC-W00-E02-S002-02 | §M rejected deps all explicitly addressed: viper/envconfig absent from go.mod (viper appears only in the unpruned graph via minio-go's own go.mod; `go mod why -m` → "main module does not need"); no NATS/Kafka client; no password-hashing lib (x/crypto needed only for sha3 via validator); no custom crypto module. New-approval trio recorded: backoff/v5 v5.0.3 present-indirect (go.mod:25), golang-lru/v2 absent from go.mod, gobreaker absent entirely; yaml watch item: go.yaml.in/yaml/v3 v3.0.4 indirect, unchanged | **pass** | EV-W00-E02-S002-001, EV-W00-E02-S002-002 |
| AC-W00-E02-S002-03 | All four tools confirmed from this repository's own files: golangci-lint **v2.11.4** (`Makefile:16`, `ci.yml:62`; local binary 2.11.4 matches); GoReleaser **no exact binary pin** — SHA-pinned goreleaser-action v7.2.3 + `version: "~> v2"` (`release.yml:47–50`), Makefile targets deliberately `@latest`; Trivy **no explicit binary pin** — SHA-pinned trivy-action v0.36.0, scanners vuln/secret/misconfig, CRITICAL/HIGH, ignore-unfixed, exit-code 0 (`security-scan.yml:68–75`); goose/v3 **v3.27.2** (`go.mod:13`). None omitted; unpinned states recorded as facts, not invented versions | **pass** | EV-W00-E02-S002-003 |

### Actual result

See per-AC table above; full detail in `artifacts/post-implementation/dependency-inventory.md`
and `artifacts/post-implementation/tool-version-inventory.md`.

### Pass or fail

**Pass — all three acceptance criteria.**

### Evidence identifier

EV-W00-E02-S002-001, EV-W00-E02-S002-002, EV-W00-E02-S002-003 (`evidence/index.md`).

### Execution date

2026-07-13.

### Commit or revision

`0a31186cada5c275a588c74081cf977adf346e61` (branch `main`).

### Environment

macOS 26.5.2 (Darwin 25.5.0), arm64, go1.26.5 darwin/arm64, local workstation; concurrent
sibling-worker test load present (non-timing evidence, unaffected).

### Reviewer

Reviewer unassigned — conductor acceptance gate pending.

### Findings

Zero drift vs REVIEW §L/§M. Non-drift observations recorded: viper/golang-lru appear only in the
unpruned module graph via minio-go's own go.mod (not needed by the main module); GoReleaser/Trivy
deliberately have no exact binary pin per in-repo comments.

### Retest status

Not required — first run passed.

### Final conclusion

All three ACs pass with commit-pinned, registered evidence; story ready for independent review.
