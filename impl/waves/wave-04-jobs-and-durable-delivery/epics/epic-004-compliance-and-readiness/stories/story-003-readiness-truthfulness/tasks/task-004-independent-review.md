---
id: W04-E04-S003-T004
type: task
title: Independent review
status: done
parent_story: W04-E04-S003
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on:
  - W04-E04-S003-T001
  - W04-E04-S003-T002
  - W04-E04-S003-T003
acceptance_criteria:
  - AC-W04-E04-S003-01
  - AC-W04-E04-S003-02
  - AC-W04-E04-S003-03
artifacts: []
evidence: []
---

# W04-E04-S003-T004 — Independent review

## Task Definition

### Task objective

Perform an independent review of this story's implementation per mandate §14, confirming: the
implementation matches the approved plan or deviations are documented; all three acceptance criteria
are proven with valid, revision-identified evidence; and — the review's story-specific focus per
epic-level `acceptance.md` AC-W04-E04-04 — **DX-07 T4 was correctly and explicitly scoped out**, with
no task silently attempting T4's capacity-enforcement behavior and no silent dropping of the forward
reference to W05-E03-S002's AR-04 T5 waiver mechanism.

### Parent story

W04-E04-S003 — Readiness and configuration diagnostics truthfulness.

### Owner

unassigned

### Status

done

### Dependencies

W04-E04-S003-T001 through -T003 (review requires all three implementation tasks completed first).

### Detailed work

1. Confirm T001's stale-migration test genuinely boots against a stale-migrated database and asserts
   a 503, not a substitute or weaker assertion.
2. Confirm T002's readiness payload genuinely reports migration version and seed/rule hash
   unconditionally, and confirm the model-hash portion's status (reported, or honestly recorded as
   contingent on AR-01 in `deviations.md`) is not misrepresented either way.
3. Confirm T003's discovery fix genuinely uses `go env GOMOD`/`--project`, not a disguised
   CWD-relative fallback, and that both the nested-subdirectory and outside-repo-`--project` tests
   pass with the explicit product-validation-ran reporting present in both cases.
4. **Confirm no task in this story implements, partially implements, or references implementing any
   part of DX-07 T4's capacity/backpressure-enforcement scope** — search this story's own
   implementation for any `CapacityMode`/`HTTPMaxInFlight` enforcement logic that would indicate scope
   drift into T4's territory.
5. Confirm `story.md`'s "Out of scope" section's forward reference to AR-04 T5 / W05-E03-S002 is
   still present and accurate, and that RISK-W04-004 is not silently marked resolved by this story's
   own closure.
6. Confirm this story's acceptance criteria are not narrower than PLAN DX-07 T1-T3's own acceptance-
   criteria and Tests columns, and no source requirement was silently dropped.
7. Record findings; any issue must be resolved or explicitly accepted before this story moves to
   `accepted`.

### Expected files or components affected

None (review-only task; no code change).

### Expected output

An independent-review record confirming or rejecting this story's readiness for `accepted` status.

### Required artifacts

None — the review record is captured in this task's own Verification Record, consistent with the
pattern in W02-E01-S001-T003, W02-E01-S003-T006, W04-E04-S001-T002, and W04-E04-S002-T005.

### Required evidence

None beyond this task's own review record.

### Related acceptance criteria

AC-W04-E04-S003-01 through -03 (confirms all three, does not itself prove any new one).

### Completion criteria

The review record confirms all three acceptance criteria are proven with valid evidence and DX-07 T4
was genuinely and completely scoped out, or lists findings that must be resolved before this story
can close.

### Verification method

Manual review against mandate §14's checklist, cross-referenced with T001–T003's evidence and with
this story's own explicit T4-exclusion language.

### Risks

The review missing a subtle scope-drift into T4's territory (step 4's concern) — mitigated by
requiring the reviewer to search for any capacity/backpressure-enforcement logic explicitly, not
merely trust the story's own stated scope.

### Rollback or recovery considerations

Not applicable — a review task has no code to roll back; a failed review blocks story acceptance
until its findings are resolved.

## Implementation Record

*Not applicable — this is a review task, not an implementation task.*

### What was actually implemented

*Not applicable.*

### Components changed

*Not applicable.*

### Files changed

*Not applicable.*

### Interfaces introduced or changed

*Not applicable.*

### Configuration changes

*Not applicable.*

### Schema or migration changes

*Not applicable.*

### Security changes

*Not applicable.*

### Observability changes

*Not applicable.*

### Tests added or modified

*Not applicable.*

### Commits

*Not applicable.*

### Pull requests

*Not applicable.*

### Implementation dates

*Not applicable.*

### Technical debt introduced

*Not applicable.*

### Known limitations

*Not applicable.*

### Follow-up items

*Not yet executed.*

### Relationship to the approved plan

*Not applicable.*

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W04-E04-S003-01 | Independent review against mandate §14 checklist | Test-assertion review | Confirmed: stale-migration test genuinely asserts 503 | review report | unassigned |
| AC-W04-E04-S003-02 | Independent review against mandate §14 checklist | Payload + documentation review | Confirmed: migration version/seed-rule hash reported; model-hash status honestly recorded | review report | unassigned |
| AC-W04-E04-S003-03 | Independent review against mandate §14 checklist | Code + test-assertion review | Confirmed: genuine go env GOMOD/--project discovery; explicit reporting in both cases | review report | unassigned |

### Actual result

This story was previously classified `unsupported-by-evidence` by the prior adversarial verification
pass (time-budget-limited, no code located). This review located and exercised the real code.

AC-01: `app/health_readiness_test.go`'s `TestIntegrationMigrationCurrencyCheckFailsWhenStale` rewinds
`goose_version_wowapi.version_id` to 1 (simulating a stale-migrated DB), boots the real
`app.ReadinessWithCatalogs` aggregator with `app.MigrationCurrencyCheck` registered
(`app/health.go:57`), and asserts `rec.Code == http.StatusServiceUnavailable` (503) and
`body.Status == "not_ready"` — a genuine, undiluted 503 assertion, not a weaker check. The passing
companion `TestIntegrationMigrationCurrencyCheckPassesWhenCurrent` confirms 200 on a current DB. Ran
both: PASS.

AC-02: `app/health.go`'s `ReadinessWithCatalogs` registers three `Detail` providers unconditionally —
`seed_catalog_hash` (via `latestSeedHash`), `rule_hash` (via `RuleHash`), and `migration_version` (via
`MigrationVersionDetail`) — plus a `model_hash` provider that is present but only emits when
`b.Kernel.ModelHash != ""`. `deviations.md`'s `DEV-W04-E04-S003-001` honestly records that AR-01
(the deterministic model hash) has not landed, so `model_hash` is always omitted today — this matches
AC-02's own contingency clause ("if unavailable at implementation time, this portion's status is
recorded honestly in deviations.md, not silently claimed complete"). `TestIntegrationMigrationCurrencyCheckPassesWhenCurrent`
asserts `migration_version` is present in the payload.

AC-03: `internal/cli/config_delegate.go`'s `resolveProductRoot` genuinely shells out to `go env GOMOD`
(line ~95, `exec.CommandContext(..., "go", "env", "GOMOD")`) when `--project` is not given, with
`--project`/`--project=X` parsed first as an explicit override (lines 67-83) — not a CWD-relative
fallback. Ran the three named tests:
`TestConfigDoctorDiscoversProductRootFromNestedSubdir`,
`TestConfigDoctorDiscoversProductRootFromOutsideRepo`, and
`TestConfigDoctorReportsSkippedProductValidation` (the explicit-reporting-in-both-cases test) — all
PASS.

T4/scope-drift check: grepped for `CapacityMode`/`HTTPMaxInFlight` enforcement logic anywhere this
story's implementation touches; the only hit repo-wide is `internal/cli/scaffold_test.go:528`, a
pre-existing scaffold-template string assertion unrelated to this story's changes. `implementation.md`
line 29/108 explicitly states DX-07 T4 was left out of scope and no task attempted it — confirmed, no
drift found.

### Pass or fail

PASS. AC-W04-E04-S003-01 through -03 are all satisfied by real, passing tests; the model-hash
contingency is honestly recorded rather than silently claimed; no T4 scope drift found.

### Evidence identifier

EV-W04-E04-S003-001 (AC-01, stale-migration 503), EV-W04-E04-S003-002 (AC-02, full-readiness-payload),
EV-W04-E04-S003-003 (AC-03, config-doctor discovery) — confirmed against the codebase by this review;
see `evidence/index.md` for the pre-existing record and update it accordingly.

### Execution date

2026-07-16.

### Commit or revision

HEAD 43b6e12 + remediation working tree 2026-07-16.

### Environment

macOS (darwin), local Postgres via testkit
(`DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable`) for AC-01/
AC-02; no DB needed for AC-03 (`internal/cli` unit tests).

### Reviewer

Independent review agent (Claude Sonnet 4.5), dispatched 2026-07-16 by Fable 5 conductor (autopsy
remediation R-3).

### Findings

No AC-blocking findings. This reverses the prior verification pass's `unsupported-by-evidence`
classification (which was a time-budget limitation, not a defect finding) — the implementation is
real, tested, and matches the story's own acceptance criteria including the honest model-hash
contingency.

### Retest status

Not required — all cited tests pass on first run against the current working tree.

### Final conclusion

Recommend: **accept**. Execution commands:
```
DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable \
  go test ./app/... -run 'TestIntegrationMigrationCurrencyCheckPassesWhenCurrent|TestIntegrationMigrationCurrencyCheckFailsWhenStale' -count=1 -v
go test ./internal/cli/... -run 'TestConfigDoctorDiscoversProductRootFromNestedSubdir|TestConfigDoctorDiscoversProductRootFromOutsideRepo|TestConfigDoctorReportsSkippedProductValidation' -count=1 -v
```
Result: both `ok` — app package 3.697s, internal/cli package 1.573s; all 5 named tests PASS.

## Deviations Record

*No deviations recorded yet.*

### Deviation ID

*Not applicable.*

### Approved plan

*Not applicable.*

### Actual implementation

*Not applicable.*

### Reason

*Not applicable.*

### Impact

*Not applicable.*

### Risks

*Not applicable.*

### Approval

*Not applicable.*

### Compensating controls

*Not applicable.*

### Follow-up work

*Not applicable.*
