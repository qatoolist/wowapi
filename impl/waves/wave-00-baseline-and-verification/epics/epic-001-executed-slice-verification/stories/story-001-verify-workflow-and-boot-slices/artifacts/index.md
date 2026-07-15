---
id: W00-E01-S001-ARTIFACTS-INDEX
type: artifact-index
parent_story: W00-E01-S001
status: executed
created_at: 2026-07-12
updated_at: 2026-07-13
derived: false
---

# Artifact index — W00-E01-S001

Per `artifact-policy.md` §9.2 required fields. All artifacts below were produced on 2026-07-13
against pinned commit `0a31186cada5c275a588c74081cf977adf346e61`. They are stored under this
story's `evidence/tests/` directory (the story's evidence tree, as anticipated at planning time);
each doubles as the raw file behind the corresponding evidence record in `evidence/index.md`.
Checksums are SHA-256 (first 16 hex chars).

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Repository path / storage location | Version | Checksum | Status | Reviewer | Retention requirement |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| ART-W00-E01-S001-001 | SEC-02 workflow race-test execution log | Execution log | post-implementation | Full `go test -v ./kernel/workflow/... -race` output: exit 0, no race warnings; nil-`ev`-panic (`TestNewRuntimePanicsOnNilDeps`) and fail-closed-`Override` (`TestIntegrationOverrideAuthzGate`, `TestIntegrationOverrideFailsClosedWithoutPermission`) assertions PASS | SEC-02 | W00-E01-S001-T001 | `evidence/tests/sec02-workflow-race.log` | commit `0a31186` | sha256:0a17e85ea35ecdce | produced | unassigned | Retain for the life of the SEC-02 finding-slice's traceability chain; do not delete on supersession |
| ART-W00-E01-S001-002 | AR-04 T1 boot-namespace-rejection test execution log + full-suite green check | Execution log | post-implementation | `go test -v ./app/... -run Boot` (exit 0, `TestBootFailsOnUnknownConfigNamespace` PASS) and full `go test ./...` (exit 0, 57 packages, 0 FAIL) | AR-04 | W00-E01-S001-T002 | `evidence/tests/ar04-boot-run-boot.log`; `evidence/tests/ar04-full-suite.log` | commit `0a31186` | sha256:d04aec5132af0008; sha256:91427e58ded80d82 | produced | unassigned | Same retention policy as ART-W00-E01-S001-001 |
| ART-W00-E01-S001-003 | AR-06 T1 authzStore sentinel-injection test execution log | Execution log | post-implementation | `go test -v ./kernel/authz/... -race` (sentinel test `TestCachingStoreOrgAncestorsRoutesToComposedInner` PASS) and `go test -v ./kernel/ -run 'TestIntegrationRulesResolverOrgAncestry' -race -count=1` (both org-ancestry integration tests PASS), both exit 0, no race warnings | AR-06 | W00-E01-S001-T003 | `evidence/tests/ar06-authz-race.log`; `evidence/tests/ar06-kernel-rules-race.log` | commit `0a31186` | sha256:b954cb0cbc1c15b0; sha256:97441fa6cb69364c | produced | unassigned | Same retention policy as ART-W00-E01-S001-001 |
| ART-W00-E01-S001-004 | AR-05 T1/T2 doc-drift grep + Context interface-diff log | Execution log | post-implementation | Phantom-API grep over `README.md` + `docs/blueprint/` (7 hits found in blueprint 04/06/10/12 — identical set already present at fix commit `345e4ce`; README and blueprint 11 clean; no such Go functions exist) plus method-set diff of blueprint 06's `Context` listing vs `module/module.go` (EMPTY — 40/40 match) | AR-05 | W00-E01-S001-T004 | `evidence/tests/ar05-doc-drift.log` | commit `0a31186` | sha256:3f0c10fa413d04f4 | produced | unassigned | Same retention policy as ART-W00-E01-S001-001 |

## Notes

No separate story-level summary artifact is registered — `verification.md`'s post-execution record
serves that role, as anticipated at planning time.
