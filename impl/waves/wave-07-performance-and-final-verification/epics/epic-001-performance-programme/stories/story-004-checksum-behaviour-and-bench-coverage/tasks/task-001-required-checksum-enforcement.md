---
id: W07-E01-S004-T001
type: task
title: Required checksum enforcement and call-site audit
status: done
parent_story: W07-E01-S004
owner: W07-Scoping-Dispatch.W07E01S004
created_at: 2026-07-12
updated_at: 2026-07-16
depends_on: []
acceptance_criteria:
  - AC-W07-E01-S004-01
artifacts:
  - ART-W07-E01-S004-001
evidence:
  - EV-W07-E01-S004-001
---

# W07-E01-S004-T001 — Required checksum enforcement and call-site audit

## Task Definition

### Task objective

Require framework uploads to always persist canonical checksum metadata; audit every upload call path for universality.

### Parent story

W07-E01-S004

### Owner

W07-Scoping-Dispatch.W07E01S004

### Status

complete

### Dependencies

None.

### Detailed work

1. Enumerate every current upload call site.
2. Require checksum-metadata persistence at each site.
3. Write an integration test: upload via the framework path, call Stat, assert no GetObject call
   occurs.

### Expected files or components affected

adapters/storage/s3's implementation files.

### Expected output

Every upload call path persists checksum metadata; normal Stat performs no body download.

### Required artifacts

ART-W07-E01-S004-001 (checksum-required enforcement + call-site audit).

### Required evidence

EV-W07-E01-S004-001 (integration test output).

### Related acceptance criteria

AC-W07-E01-S004-01.

### Completion criteria

No GetObject call occurs on a normal Stat.

### Verification method

Direct execution of the integration test.

### Risks

Medium — enumerate every current upload call site; scope expands if any bypasses checksum-signing, per PLAN T1's own risk note.

### Rollback or recovery considerations

If a call site is found bypassing checksum-signing after this task lands, add enforcement there immediately and record the gap in `deviations.md`.

## Implementation Record

Implemented a checksum-required framework upload path. `Service.InitiateUploadChecksum`
validates lowercase-hex SHA-256, persists it in the pending upload session, and
requires the optional `storage.ChecksumUploader` capability before issuing a URL.
The former checksum-free `InitiateUpload` entry point now fails closed. Memory and
S3 implement the capability; S3 signs SHA-256 headers and normal `Stat` reads
checksum metadata with HEAD semantics only.

Files: `foundation/document/service.go`, `kernel/storage/storage.go`,
`kernel/storage/memory.go`, `adapters/storage/s3/s3.go`, upload call-site tests,
and `perf/results/perf-05-checksum-inventory-v1.json`.

No schema, migration, or runtime configuration change was required. The security
posture changed from implicit checksum fallback to required canonical metadata.
Implemented on 2026-07-14 in the working tree based on `733ef3e`; no PR was
created. No technical debt or deviation from `plan.md` was introduced.
## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W07-E01-S004-01 | `WOWAPI_REQUIRE_S3=1 go test ./adapters/storage/s3 -count=1` plus focused document/storage tests | Local MinIO | Framework upload persists canonical SHA-256 and normal Stat performs zero GetObject calls | integration test report | story reviewer |

Result: **PASS** on 2026-07-14 against local MinIO and the working tree based on
`733ef3e`. Evidence: EV-W07-E01-S004-001. No findings remained in the focused
tests; story-level independent review is recorded by T007.
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
