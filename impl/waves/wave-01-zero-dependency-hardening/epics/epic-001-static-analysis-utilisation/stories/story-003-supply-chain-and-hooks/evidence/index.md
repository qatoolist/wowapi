---
id: W01-E01-S003-EVIDENCE-INDEX
type: evidence-index
parent_story: W01-E01-S003
status: produced
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W01-E01-S003 — Evidence index

Per mandate §10. Structure adaptation per `governance/naming-conventions.md` "Adaptation 2": category
subdirectories under `evidence/` (e.g. `security/`, `logs/`, `reviews/`) are created on first real
content, not pre-populated empty. All entries below are `not yet produced`.

**Shared evidence-root citation.** `requirement-inventory.md`/MATRIX CS-23's own evidence-register row
cites a shared historical evidence path convention of `evidence/premier/FBL-05/` and
`evidence/premier/FBL-07/` across the source documents that originally recorded FBL-05's and FBL-07's
findings — since S002 (the other FBL-07 half, `W01-E01-S002`) is expected to cite the same historical
root for its own FBL-07 evidence lineage. This is a citation of the *source documents'* historical
evidence-path convention, not an instruction that this story's own `evidence/index.md` physically shares
a file with S002's — each story keeps its own separate, physical `evidence/index.md` under its own
`story-<NNN>/` directory, exactly as this file is structured.

All four evidence items were produced 2026-07-13 by W01Lint. Shared record fields (mandate §10):
story W01-E01-S003; commit SHA `0a31186cada5c275a588c74081cf977adf346e61` (HEAD; changes are an
uncommitted working diff on top — `.githooks/pre-push | 26 ++`, `ci.yml | 2 +`,
`security-scan.yml | 9 ++` — conductor owns the wave commit); branch `main`; environment
darwin/arm64 dev workstation (hook runs in pristine `git archive HEAD` copy), GitHub Actions for the
observed scheduled run; tool versions Go 1.26.5, golangci-lint v2.11.4, actionlint v1.7.12 (pin),
Trivy (local homebrew; CI pins trivy-action v0.36.0); reviewer: pending independent review
(mandate §14). Files under `logs/`.

| Evidence ID | Type | Task | AC proven | Execution command | Result | Status | File |
|---|---|---|---|---|---|---|---|
| EV-W01-E01-S003-001 | command-execution log | T001 | AC-…-01 | `go mod verify` + `actionlint ci.yml security-scan.yml` | `all modules verified`, exit 0; actionlint clean | produced (superseded-by-CI-run planned as `retested` after conductor push) | `logs/gomodverify-and-actionlint.log` |
| EV-W01-E01-S003-002 | security-scan report | T002 | AC-…-02 | `trivy fs --scanners license .` (pristine HEAD copy; CI job adds severity CRITICAL,HIGH) | 70 dep licenses enumerated, 0 CRITICAL/HIGH | produced (same CI-run supersession plan) | `logs/trivy-license-local-report.txt` |
| EV-W01-E01-S003-003 | CI execution record + audit note | T004 | AC-…-03 | `gh run list/view --workflow=ci.yml --event=schedule` + file-chain inspection | Scheduled run 29229288699 success; seed-replay step observed executing | produced | `logs/nightly-fuzz-observed-run.log` + `../artifacts/nightly-fuzz-confirmation.md` |
| EV-W01-E01-S003-004 | execution log (fail-before/pass-after) | T003 | AC-…-04 | `.githooks/pre-push` under 4 env scenarios (see `verification.md`) | before: silent pass w/ skips; after: loud fail w/o DB, pass w/ DB, loud opt-out | produced | `logs/prepush-fail-before-silent-pass.log`, `logs/prepush-after-nodb-loud-fail.log`, `logs/prepush-after-withdb-pass.log`, `logs/prepush-after-optout-pass.log` |

## Reviewer completion addendum — 2026-07-16

**Reviewer**: Independent review agent (Claude Sonnet 4.5), dispatched 2026-07-16 by Fable 5 conductor (autopsy remediation R-3).
**Review date**: 2026-07-16.
**Commit revision reviewed against**: HEAD 43b6e12 + remediation working tree 2026-07-16.
**Disposition**: Spot-checked; gap CONFIRMED still open. Re-examined `.github/workflows/ci.yml` (go mod verify step present, line 189-190) and this repo's live state: the working tree remains uncommitted/unpushed as of 2026-07-16 (git status shows the wave's closure-report.md itself still modified, never pushed), so the go-mod-verify CI step and Trivy license scanner have STILL never executed in actual CI — the same gap the autopsy found on 2026-07-13 persists. EV-W01-E01-S003-001/002's 'produced (superseded-by-CI-run planned as retested after conductor push)' status is honest and explicit (satisfies the evidence-policy carry-forward rationale-note requirement — this is not a silent carry-forward), but AC-W01-03's CI-execution leg is not yet proven. Recommend accept-with-conditions: accept the local-run + actionlint evidence as sufficient for the non-CI-execution acceptance criteria, but keep AC-W01-03 in a 'retested-pending' state until an actual CI run is recorded as a new evidence record referencing this one as superseded, per evidence-policy.md's revision-pinning rule.

This addendum retroactively fills the evidence-policy-mandated "reviewer" field. The original
record above (including any "Pending — conductor acceptance gate" line) is left unmodified per
the failed-evidence preservation convention — this is an appended addendum, not a rewrite.
