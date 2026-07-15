---
id: W06-E01-S002-T001
type: task
title: Golden-consumer scaffold job
status: done
parent_story: W06-E01-S002
owner: W06E01Impl
created_at: 2026-07-12
updated_at: 2026-07-14
depends_on: []
acceptance_criteria:
  - AC-W06-E01-S002-01
artifacts:
  - ART-W06-E01-S002-001
evidence:
  - EV-W06-E01-S002-001
  - EV-W06-E01-S002-007
  - EV-W06-E01-S002-010
---

# W06-E01-S002-T001 — Golden-consumer scaffold job

## Task Definition

### Task objective

Build the golden-consumer scaffold job: install the CLI via `go install`, reusing DX-01 T5's isolated-temp-dir subprocess-scaffold harness (W01-E04-S001) as the shared primitive.

### Parent story

W06-E01-S002

### Owner

W06E01Impl

### Status

done

### Dependencies

None (this story's own dependency on W01-E04-S001's harness is a story-level entry gate).

### Detailed work

1. Confirm W01-E04-S001's isolated-temp-dir harness is `accepted` and reusable as-is.
2. Build the scaffold job: install the CLI via `go install` into an isolated temp dir.
3. Confirm the installed binary is a standalone binary, not a repo-internal import.

### Expected files or components affected

A new golden-consumer fixture directory (exact path TBD); reuses W01-E04-S001's harness package.

### Expected output

A CI-job-ready scaffold that installs the CLI via `go install` in an isolated temp dir.

### Required artifacts

ART-W06-E01-S002-001 (golden-consumer scaffold job).

### Required evidence

EV-W06-E01-S002-001 (installation-log evidence).

### Related acceptance criteria

AC-W06-E01-S002-01.

### Completion criteria

The scaffold job installs the CLI via `go install`, confirmed not a repo-internal import.

### Verification method

Direct execution of the scaffold job, inspecting the installed binary's provenance.

### Risks

None beyond the general harness-reuse risk (mitigated by W01-E04-S001 already being `accepted`).

### Rollback or recovery considerations

Revert to a pre-scaffold state if the harness reuse proves incompatible; escalate rather than silently forking a second harness.

## Implementation Record

Implemented 2026-07-13 by W06E01Impl.

### What was actually implemented

`goldenConsumerScaffold` installs the CLI with real `go install` into an isolated `GOBIN`, stamps the
current binary, packages the checkout as a versioned module proxy, and calls W01's existing
`scaffoldPipeline`. The shared proxy helper was deepened only as required for current source: it now
includes the landed `foundation/` public tree and excludes nested test-fixture modules, which Go module
zips reject. The resulting consumer uses a versioned wowapi requirement and no checkout `replace`.

### Components changed

`internal/cli` test infrastructure.

### Files changed

- `internal/cli/golden_consumer_test.go`
- `internal/cli/e2e_scaffold_harness_test.go`
- `evidence/DX-04/t1-installed-two-module.log`
- `evidence/DX-04/t1-fail-first.log`

### Interfaces introduced or changed

Test-internal helper only.

### Configuration changes

None.

### Schema or migration changes

None in framework source.

### Security changes

None.

### Observability changes

None.

### Tests added or modified

Added `TestGoldenConsumerInstalledBinaryTwoModules`.

### Commits

None; base `733ef3e930cbb3f89f5bbc53d8f562c60e426513` plus uncommitted harness.

### Pull requests

None.

### Implementation dates

2026-07-13.

### Technical debt introduced

None.

### Known limitations

This task proves installation/scaffolding only; downstream task status records the generator gap.

### Follow-up items

None within T001.

### Relationship to the approved plan

Matched the plan and reused `scaffoldPipeline` rather than forking a second harness.
## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W06-E01-S002-01 | Run focused scaffold job | Supplied env + Go toolchain | CLI installs via go install; product resolves a version, no checkout replace | installation log | W06E01Impl |

### Actual result

The test logged successful go-install, init, tidy, download, build, boot smoke, two module generations,
two CRUD generations, final tidy, and final build. It asserted no wowapi checkout replace in go.mod.

### Pass or fail

Pass.

### Evidence identifier

EV-W06-E01-S002-001 (`evidence/DX-04/t1-installed-two-module.log`).

### Execution date

2026-07-13.

### Commit or revision

Base `733ef3e930cbb3f89f5bbc53d8f562c60e426513` plus uncommitted harness.

### Environment

Supplied DB/S3-required environment plus local S3/SMTP/OTLP endpoints.

### Reviewer

W06E01Impl.

### Findings

None for AC-01.

### Retest status

Pass: `go test ./internal/cli -run '^TestGoldenConsumerInstalledBinaryTwoModules$' -count=1 -v`.

### Final conclusion

T001 is done and AC-01 is verified.
## Deviations Record

*No task-local deviations.*

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
