---
id: W04-E04-S002-EVIDENCE-INDEX
type: evidence-index
parent_story: W04-E04-S002
status: accepted
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W04-E04-S002 — Evidence index

Per mandate §10. Structure adaptation per `governance/naming-conventions.md` "Adaptation 2":
category subdirectories under `evidence/` are created on first real content, not pre-populated
empty. All entries below are `not yet produced`.

| Evidence ID | Type | Task | Acceptance criteria proven | Execution command | Commit SHA | Result | Status |
|---|---|---|---|---|---|---|---|
| EV-W04-E04-S002-001 | anchor-tamper-detection report | W04-E04-S002-T001 | AC-W04-E04-S002-01 | `go test ./kernel/audit/... -run TestIntegrationExternalAnchorTamperDetection -count=1 -v` | 733ef3e930cbb3f89f5bbc53d8f562c60e426513 | PASS | produced |
| EV-W04-E04-S002-002 | DSR export artifact-completion/checksum report | W04-E04-S002-T002 | AC-W04-E04-S002-02 | `go test ./kernel/retention/... -run 'Artifact|DSRExportArtifactWriteFailure' -count=1 -v` | 733ef3e930cbb3f89f5bbc53d8f562c60e426513 | PASS | produced |
| EV-W04-E04-S002-003 | legal-hold negative-test report | W04-E04-S002-T003 | AC-W04-E04-S002-03 | `go test ./kernel/retention/... -run TestIntegrationCentralLegalHoldBlocksDisposeErase -count=1 -v` | 733ef3e930cbb3f89f5bbc53d8f562c60e426513 | PASS | produced |
| EV-W04-E04-S002-004 | RecordClass enumeration record (both repos) | W04-E04-S002-T003 | AC-W04-E04-S002-04 (enumeration half) | `grep -R 'retention.NewRegistry().Register\|\.Register(retention.RecordClass' .` in wowapi; wowsociety has no registered classes per `dependencies.md` | 733ef3e930cbb3f89f5bbc53d8f562c60e426513 | wowapi: 0 product classes; wowsociety: none | produced |
| EV-W04-E04-S002-005 | explicit-status test report | W04-E04-S002-T004 | AC-W04-E04-S002-04 (status half) | `go test ./kernel/retention/... -run 'ExplicitPerClass' -count=1 -v` | 733ef3e930cbb3f89f5bbc53d8f562c60e426513 | PASS | produced |

Evidence status vocabulary (per mandate §10): `not yet produced` is this programme's pre-execution
state, outside the mandate's own failed/superseded/retested/resolved/accepted-exception vocabulary,
which applies only once an evidence item has actually been produced at least once.
