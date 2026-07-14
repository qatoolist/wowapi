---
id: VER-W06-E01-S002
type: verification-record
parent_story: W06-E01-S002
status: verified
created_at: 2026-07-12
updated_at: 2026-07-14
---

# Verification record — W06-E01-S002

## Verification matrix

| Acceptance criterion | Verification method | Required environment | Result | Evidence | Reviewer |
|---|---|---|---|---|---|
| AC-W06-E01-S002-01 | Versioned installed-binary fixture run and provenance/no-replace assertions | Go toolchain + local module proxy | PASS | EV-W06-E01-S002-007, EV-W06-E01-S002-010 | W06-E01-S002-Verify |
| AC-W06-E01-S002-02 | Generate all named subsystem types, assert every artifact and automatic wiring, build and boot | Go toolchain | PASS | EV-W06-E01-S002-007, EV-W06-E01-S002-010 | W06-E01-S002-Verify |
| AC-W06-E01-S002-03 | Direct real-infrastructure contract run | Postgres, MinIO, Mailpit, Jaeger/OTel | PASS | EV-W06-E01-S002-003, EV-W06-E01-S002-010 | W06-E01-S002-Verify |
| AC-W06-E01-S002-04 | Genuine N-1-to-candidate two-pass replay | Tagged v1.1.0, local candidate, real infrastructure | PASS | EV-W06-E01-S002-004, EV-W06-E01-S002-010 | W06-E01-S002-Verify |
| AC-W06-E01-S002-05 | Workflow/manifest validation plus deliberately incomplete fixture | CI configuration, actionlint, Python, Go | PASS | EV-W06-E01-S002-005, EV-W06-E01-S002-012 | W06-E01-S002-Verify |

## Post-execution record

### Actual result

- The installed candidate CLI generated and booted a versioned consumer with no checkout replace.
- `catalog` and `fulfillment` collectively contain resource, rule, workflow, event handler, recurring
  job, document flow, notification, and webhook output; the fixture asserts all expected files and
  generated module registration.
- Real-infrastructure execution passed authenticated CRUD, cross-tenant denial, outbox dispatch,
  worker downtime/restart recovery, and required service checks.
- The replay exercised build/boot at tagged `v1.1.0`, upgraded dependency and generated scaffold to
  local candidate `v1.2.0-w06e01s002.11`, then reran build/boot and real-infrastructure contracts.
- `make actionlint`, release-gate manifest validation, and the incomplete-fixture rejection test passed.
  The Wave-4 manifest entry invokes `make golden-consumer`, marks services required, and the exact-SHA
  runner now starts Jaeger alongside Postgres, MinIO, and Mailpit.

### Pass or fail

PASS: AC-01 through AC-05.

### Evidence identifiers

EV-W06-E01-S002-003 through EV-W06-E01-S002-005, EV-W06-E01-S002-007,
EV-W06-E01-S002-010, EV-W06-E01-S002-012, and final review EV-W06-E01-S002-014.

### Execution date

2026-07-13T20:31:25Z through 2026-07-13T20:33:48Z; independent final review on 2026-07-14.

### Commit or revision

Worktree snapshot based on `733ef3e930cbb3f89f5bbc53d8f562c60e426513`, with authoritative file
checksums in `artifacts/index.md`.

### Environment

Darwin arm64; Go 1.26.5; Docker Compose 5.3.1; real Postgres/MinIO/Mailpit/Jaeger services;
actionlint 1.7.12; Python 3.14.2.

### Reviewer

W06-E01-S002-Verify.

### Findings

The first independent review found stale artifact/evidence/lifecycle records, missing Jaeger
provisioning in the exact-SHA runner, and stale N-1/N prose. Those findings were fixed. Failed
environmental executions and the evidence-recorder assertion error are preserved as EV-008, EV-009,
and EV-011, with passing superseding records EV-010 and EV-012.

### Retest status

Retested. `make golden-consumer` passed after seeding the eval runtime's local offline module cache.
The CI checks passed after correcting the evidence recorder to parse the JSON-formatted manifest rather
than search for YAML syntax.

### Final conclusion

All five acceptance criteria have valid, content-pinned evidence. EV-W06-E01-S002-014 records the
passing fresh independent review; the story is verified and accepted.
