---
id: EV-W00-E02-S002-002
type: evidence
evidence_type: review report (cross-check table)
parent_story: W00-E02-S002
task: W00-E02-S002-T001
acceptance_criteria:
  - AC-W00-E02-S002-01
  - AC-W00-E02-S002-02
status: pass
created_at: 2026-07-13
---

# EV-W00-E02-S002-002 — Direct-dependency cross-check vs REVIEW §L/§M

- **Evidence ID:** EV-W00-E02-S002-002
- **Evidence type:** review report (cross-check table)
- **Story / task:** W00-E02-S002 / W00-E02-S002-T001
- **Acceptance criteria proven:** AC-W00-E02-S002-01, AC-W00-E02-S002-02
- **Execution command:** manual enumeration of `go.mod:8–20` (13 direct require lines) against
  `docs/implementation/fable5-final-architecture-review-2026-07-11.md` §L (lines 285–287) and §M
  (lines 289–294), supported by `go list -m all` / `go mod graph` / `go mod why -m` captures
  (EV-W00-E02-S002-001, `../logs/`).
- **Code revision / commit SHA:** `0a31186cada5c275a588c74081cf977adf346e61`
- **Branch:** `main`
- **Execution environment:** macOS 26.5.2 (Darwin 25.5.0), arm64, local workstation; concurrent
  sibling-worker test load present (non-timing evidence, unaffected).
- **Tool versions:** go1.26.5 darwin/arm64; git 2.x (read-only).
- **Date and time:** 2026-07-13
- **Result:** **PASS — zero unexplained drift.**
  - 13/13 direct dependencies dispositioned `approved` against §L; zero `undocumented drift`.
  - "10 vs 13" reconciled: §L's "otel×4" = go.mod lines 16–19; 10 logical = 13 lines.
  - New approvals: `cenkalti/backoff/v5` v5.0.3 present-indirect (go.mod:25);
    `hashicorp/golang-lru/v2` and `sony/gobreaker` absent from go.mod (expected).
  - §M rejected register: viper/envconfig, NATS/Kafka clients, password-hashing libs, custom
    crypto — all absent from go.mod and (except viper's unpruned-graph appearance via
    minio-go's own go.mod, not needed by the main module per `go mod why -m`) absent from the
    build list.
  - `sethvargo/go-retry` v0.3.0 re-confirmed present-indirect via goose/v3 (Stage-7 adjudication
    re-verified); `go.yaml.in/yaml/v3` v3.0.4 indirect (watch item unchanged).
- **File or URI:** full disposition tables in
  `../../artifacts/post-implementation/dependency-inventory.md`; raw captures in `../logs/`.
- **Checksum:** n/a (markdown, git-tracked).
- **Reviewer:** unassigned (conductor review gate pending).
- **Superseded evidence:** none.
