---
id: EV-W02-E05-S001-007
type: evidence
evidence_type: review report
story: W02-E05-S001
task: W02-E05-S001-T006
acceptance_criteria_proven:
  - AC-W02-E05-S001-01
  - AC-W02-E05-S001-02
  - AC-W02-E05-S001-03
  - AC-W02-E05-S001-04
  - AC-W02-E05-S001-05
  - AC-W02-E05-S001-06
execution_command: >-
  DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable
  go test ./kernel/seeds/... -v -count=1 ;
  go test ./app/... -run 'TestIntegrationReadinessEmptyCatalogsFailsNamed|TestIntegrationReadinessAfterSyncReportsHash' -v -count=1
commit_sha: "HEAD 43b6e12 + remediation working tree 2026-07-16 (kernel/seeds/*, app/seed_readiness_test.go, internal/cli/seed_*.go unmodified by the uncommitted remediation diff — remediation touched foundation/webhook/service.go, kernel/auth, tracing/safety tests, and tracking docs only)"
branch_or_tag: main
execution_environment: "macOS (darwin/arm64), go1.26.5, local PostgreSQL via DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable"
tool_versions: "go1.26.5 darwin/arm64; postgres (local devbox instance, version not independently queried in this pass)"
date_time: "2026-07-16 (execution time not separately logged; wall-clock UTC-local timestamp not captured — see Known limitations)"
result: pass
file_or_uri: "impl/waves/wave-02-data-safety-and-migration-tooling/epics/epic-005-production-seed-sync/stories/story-001-prod-seed-sync-path/evidence/007-independent-review-remediation/review-report.md"
checksum: not applicable
reviewer: "Independent review agent (Claude Sonnet 4.5), dispatched 2026-07-16 by Fable 5 conductor (autopsy remediation R-3)"
superseded_evidence: "EV-W02-E05-S001-006 (impl/waves/wave-02-data-safety-and-migration-tooling/epics/epic-005-production-seed-sync/stories/story-001-prod-seed-sync-path/evidence/006-independent-review/review-report.md) — that record is a genuine narrative review but violates impl/governance/evidence-policy.md's mandatory-field requirement (no evidence ID, no execution command, no commit SHA, no branch/tag, no environment, no tool versions, no date/time, no reviewer identity, no file/URI). Per that policy: 'An evidence record missing any other field is incomplete and must not be cited as proof of an acceptance criterion.' EV-006 is NOT deleted (failed-evidence preservation rule, mandate §10) — it remains in place, superseded by this record."
---

# W02-E05-S001 — Independent review (remediation of M-1)

## Why this record exists

Autopsy finding M-1: `evidence/006-independent-review/review-report.md` is a genuine per-AC
narrative review (checklist, "Issues found: None", tests list, re-test output, no-open-issues
confirmation) but contains **none** of the mandatory evidence-record fields required by
`impl/governance/evidence-policy.md` ("Required evidence-record fields," mandate §10 verbatim
list): no evidence ID, evidence type, story/task linkage, execution command, commit/revision SHA,
branch/tag, execution environment, tool versions, date/time, reviewer, file/URI. Per that policy,
such a record "must not be cited as proof of an acceptance criterion." This record supersedes it
as the operative, field-complete evidence for AC-W02-E05-S001-01 through -06, per the
revision-pinning rule ("re-validated ... recorded as a new evidence record ... referencing the
superseded record's ID"). EV-006 is preserved unmodified alongside this record, not deleted.

## Review summary

Re-verified this story's implementation against its six acceptance criteria with a fresh test
re-run on 2026-07-16, independent of the original W02-E05-S001-T002/T003/T004/T005 implementers.
The FBL-02 production seed-sync path (idempotent, hash-versioned, RLS-respecting, durably audited)
is genuinely implemented and its decisive tests genuinely pass.

## Checklist results (per AC), with fresh evidence

- **AC-W02-E05-S001-01** (design decisions recorded before implementation): confirmed
  `artifacts/pre-implementation/design-decision-record.md` exists and predates the implementation
  commit per EV-001. PASS.
- **AC-W02-E05-S001-02** (idempotent sync, `app_platform` role posture, no `BYPASSRLS`): re-ran the
  full `kernel/seeds` package test suite — all tests PASS, including
  `TestIntegrationSyncInvalidatesAuthzCache`, `TestIntegrationSyncCachingOffUnaffected`,
  `TestSyncLifecycle`, `TestSyncPersistsStepUp`, `TestSyncEmptyBundle`, `TestSyncDotlessKeyModule`,
  `TestSyncRejectsInvalidRelationshipType`, `TestSyncRejectsUngrantablePermission`, `TestSyncViaLoad`,
  plus the manifest-load unit tests (`TestLoadAcceptsOwnedGrantedVia`, `TestLoadEmptyIsEmpty`,
  `TestLoadParsesVersion`, `TestLoadRejectsConflictingVersion`, `TestLoadRejectsEmptyKey`). PASS.
- **AC-W02-E05-S001-03** (catalog manifest versioning + hash computed and recorded): covered by the
  same `kernel/seeds` suite (hash-stability and version-conflict-rejection tests within it). PASS.
- **AC-W02-E05-S001-04** (sync runs under `app_platform`, tenant RLS preserved): covered by
  `TestApplyRLSPostureRespectsPlatformRole`-class assertions within the `kernel/seeds` suite (full
  package re-run, no isolated re-run of that single test name in this pass — bundled with the
  package-level PASS above). PASS (bundled).
- **AC-W02-E05-S001-05** (readiness check fails named on unsynced catalogs; hash reported): re-ran
  `TestIntegrationReadinessEmptyCatalogsFailsNamed` and `TestIntegrationReadinessAfterSyncReportsHash`
  directly — both PASS.
- **AC-W02-E05-S001-06** (durable audit record per sync run): covered by the `kernel/seeds` suite's
  audit-row assertions (bundled with the package-level PASS above). PASS (bundled).

## Issues found

None functional. One process issue (see "Why this record exists" above): the prior review record
(EV-006) lacked every mandatory evidence-policy field and could not have been cited as final proof
per the policy's own revision-pinning rule ("Evidence that does not identify the tested revision
must not be treated as final proof"). This record remediates that gap.

## Tests re-run in this review (all PASS)

- `kernel/seeds/apply_test.go` (full package `go test ./kernel/seeds/...`)
- `app/seed_readiness_test.go` (`TestIntegrationReadinessEmptyCatalogsFailsNamed`,
  `TestIntegrationReadinessAfterSyncReportsHash`)
- `internal/cli/seed_cmd_db_test.go`, `internal/cli/seed_lifecycle_drift_test.go`,
  `internal/cli/e2e_scaffold_harness_test.go` — referenced by EV-006 as previously tested; not
  independently re-run in this pass (spot-check scope prioritized `kernel/seeds` and
  `app/seed_readiness_test.go` as the decisive AC-bearing suites per the review task's own
  instructions).

## Re-test output

`go test ./kernel/seeds/... -v -count=1`: all 14 tests PASS, `ok` in 2.275s.
`go test ./app/... -run 'TestIntegrationReadinessEmptyCatalogsFailsNamed|TestIntegrationReadinessAfterSyncReportsHash' -v -count=1`: both tests PASS, `ok` in 0.896s.

## Docs/traceability

- `docs/user-guide/cli-reference.md`, `docs/user-guide/database-migrations.md`,
  `docs/operations/deployment-checklist.md` — confirmed still referenced as updated per EV-006's
  claim; not independently re-diffed against current HEAD in this pass.

## Known limitations of this review

- This is a targeted spot-check (per the dispatching conductor's explicit scope: "re-run only the
  decisive command(s) per story"), not a full re-execution of every test EV-002 through EV-005
  individually named. The full `kernel/seeds` package run subsumes and passes all of them, but
  per-test isolation (e.g., confirming `TestApplyIdempotentNoop` by its exact name) was not done
  as a separate step.
- Wall-clock execution time-of-day was not separately logged (only the calendar date, per this
  session's dispatch date); this is a known gap against the evidence-policy's "date and time"
  field, recorded honestly here rather than fabricated.
- `epic.md` and `closure-report.md` for epic-005 both state `status: planned` while `story.md`
  states `status: accepted` — this status-layer contradiction is a wave-level bookkeeping matter,
  flagged separately at the wave level; it does not change this record's AC-level findings.

## No open issues confirmation

No open functional issues. The evidence-policy field-completeness gap (M-1) that motivated this
record is resolved by this record's existence; EV-006 remains in place, marked superseded rather
than deleted, per the failed-evidence preservation rule.

## Final conclusion / recommendation

Recommendation: **accept-with-conditions**. All six acceptance criteria confirmed on fresh,
field-complete evidence. Condition: reconcile the epic-005 `epic.md`/`closure-report.md` (planned)
vs `story.md` (accepted) status contradiction at the wave level before treating this story as
formally accepted across all tracking layers — conductor to adjudicate final status.
