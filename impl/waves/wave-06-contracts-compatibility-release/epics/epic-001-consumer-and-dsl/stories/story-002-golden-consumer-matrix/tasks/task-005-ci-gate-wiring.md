---
id: W06-E01-S002-T005
type: task
title: CI gate wiring
status: done
parent_story: W06-E01-S002
owner: W06E01Impl
created_at: 2026-07-12
updated_at: 2026-07-14
depends_on:
  - W06-E01-S002-T004
acceptance_criteria:
  - AC-W06-E01-S002-05
artifacts:
  - ART-W06-E01-S002-005
evidence:
  - EV-W06-E01-S002-005
  - EV-W06-E01-S002-012
---

# W06-E01-S002-T005 — CI gate wiring

## Task Definition

### Task objective

Wire the golden-consumer fixture into CI as a required gate, appearing in `ci/release-gates.yaml` at its Wave-4 boundary.

### Parent story

W06-E01-S002

### Owner
W06E01Impl

### Status

done

### Dependencies

W06-E01-S002-T004 (the fixture must be fully working end-to-end before it can be made a required gate).

### Detailed work

1. Coordinate with W06-E03-S001's manifest-schema work (may not yet exist depending on sequencing —
   see `plan.md` "Unresolved questions").
2. Add this fixture's entry to `ci/release-gates.yaml` at its Wave-4 boundary, or a forward-compatible
   placeholder entry if the schema does not yet exist, to be reconciled once it does.
3. Confirm the CI workflow actually invokes this fixture and treats its failure as a required-gate
   failure, not merely advisory.

### Expected files or components affected

CI workflow configuration changes; a `ci/release-gates.yaml` entry.

### Expected output

The fixture runs as a required CI gate.

### Required artifacts

ART-W06-E01-S002-005 (CI gate wiring).

### Required evidence

EV-W06-E01-S002-005 (CI config test).

### Related acceptance criteria

AC-W06-E01-S002-05.

### Completion criteria

The manifest entry exists at the correct Wave-4 boundary and CI treats fixture failure as blocking.

### Verification method

CI config inspection plus a deliberately-failing-fixture test confirming the gate actually blocks.

### Risks

Sequencing risk against W06-E03-S001's manifest schema — mitigated by the placeholder-entry fallback.

### Rollback or recovery considerations

If the manifest schema changes shape after this task lands a placeholder entry, update the entry to match — record the update as a deviation if it changes this task's own original entry shape.

## Implementation Record

Completed 2026-07-14.

- `Makefile` defines `make golden-consumer` with DB/S3 fail-closed requirements and the five focused
  golden-consumer/RLS-census tests.
- `.github/workflows/ci.yml` runs the fixture in a dedicated job after starting Postgres, MinIO,
  Mailpit, and Jaeger.
- `ci/release-gates.yaml` registers `golden-consumer` at `required_from_wave: 4`, invokes
  `make golden-consumer`, requires services, and emits a dedicated gate-evidence artifact.
- `.github/workflows/required-gates.yml` executes the manifest matrix at the exact requested SHA and now
  starts Jaeger as well as Postgres, MinIO, and Mailpit for service-requiring gates.
- `TestGoldenConsumerFailingFixture` proves an incomplete fixture is rejected.

### Configuration changes

Golden consumer is a required Wave-4 release gate.

### Tests added or modified

`TestGoldenConsumerFailingFixture`; actionlint and release-gate schema validation cover the wiring.

### Implementation dates

2026-07-14.

### Known limitations

No hosted GitHub run was created from this uncommitted shared worktree. The local command, manifest
contract, workflow syntax, exact-SHA checkout logic, service set, and failure injection are verified.

### Relationship to the approved plan

Matches T005. The manifest schema existed, so no placeholder entry was needed.
## Verification Record

### Actual result

PASS. Actionlint and manifest schema validation passed. The manifest entry invokes
`make golden-consumer`, begins at Wave 4, and requires services. The exact-SHA runner provisions
Postgres, MinIO, Mailpit, and Jaeger. The deliberately incomplete fixture test passed by rejecting the
fixture.

### Evidence identifier

EV-W06-E01-S002-005 and EV-W06-E01-S002-012.

### Execution date

2026-07-13T20:33:32Z through 2026-07-13T20:33:48Z.

### Commit or revision

Worktree snapshot based on `733ef3e930cbb3f89f5bbc53d8f562c60e426513`, content-pinned in
`artifacts/index.md`.

### Environment

actionlint 1.7.12; Python 3.14.2; Go 1.26.5.

### Reviewer

W06-E01-S002-Verify.

### Retest status

Retested after correcting the evidence recorder to parse the JSON-formatted manifest.

### Final conclusion

T005 is done and AC-05 is verified.
## Deviations Record

No task-local deviation.
