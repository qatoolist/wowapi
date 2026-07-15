---
id: W00-E02-S002-T001
type: task
title: go.mod inventory and approved-register cross-check
status: done
parent_story: W00-E02-S002
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on: []
acceptance_criteria:
  - AC-W00-E02-S002-01
  - AC-W00-E02-S002-02
artifacts:
  - ART-W00-E02-S002-001
evidence:
  - EV-W00-E02-S002-001
  - EV-W00-E02-S002-002
---

# W00-E02-S002-T001 — go.mod inventory and approved-register cross-check

## Task Definition

*Per mandate §8.6. This section defines the task before work begins.*

### Task objective

Run `go list -m all`, `go mod graph`, and `go list -m -json all` (or equivalent) against the
current repository HEAD, capture the raw output, and produce a line-by-line cross-check of every
**direct** `go.mod` dependency against
`docs/implementation/fable5-final-architecture-review-2026-07-11.md` §L (approved register) and §M
(rejected register), with an explicit disposition for every entry and zero unaddressed drift.

### Parent story

W00-E02-S002 — Dependency and toolchain inventory.

### Owner

Unassigned.

### Status

`done` — executed 2026-07-13 (per `impl/governance/status-model.md`).

### Dependencies

None. This task does not depend on Task 002 or on any other story's completion.

### Detailed work

1. Record the exact commit SHA the task executes against.
2. Run `go list -m all` and capture full output.
3. Run `go mod graph` and capture full output.
4. Run `go list -m -json all` (or `go list -m -u all` if update-status detail is useful) to obtain
   per-module version/license-relevant detail where available via tooling; note that Go's tooling
   does not natively report license text, so license claims are cross-checked against REVIEW §L's
   own stated licenses (MIT/BSD/Apache/MPL-2.0) rather than re-derived from scratch, unless a
   license-scanning tool is already available in this repository's toolchain (confirm during Task
   002's tooling inspection; if one exists, prefer it for this cross-check too).
5. Enumerate every **direct** dependency listed in `go.mod`'s top `require` block (13 lines at
   story-authoring time: `validator/v10`, `jwt/v5`, `uuid`, `pgx/v5`, `minio-go/v7`, `goose/v3`,
   `prometheus/client_golang`, `shopspring/decimal`, `go.opentelemetry.io/otel`,
   `otel/exporters/otlp/otlptrace/otlptracehttp`, `otel/sdk`, `otel/trace`, `yaml.v3`) and, for
   each, record a disposition: `approved` (matches REVIEW §L's original-10 list), `newly-approved`
   (matches one of the three reuse-work approvals), or `undocumented drift` (not found in REVIEW §L
   at all — requires escalation, not silent inclusion).
6. Explicitly resolve the "10 vs 13" reconciliation question from `plan.md` "Unresolved questions"
   — state whether REVIEW §L's "otel×4" phrasing accounts for the count difference, or whether an
   actual new direct dependency exists that REVIEW §L never evaluated.
7. Confirm presence/absence and disposition of the three "new approvals for reuse work":
   `cenkalti/backoff/v5` (expected: present, indirect, per `go.mod` line 25 at story-authoring
   time), `hashicorp/golang-lru/v2` (expected: absent), `sony/gobreaker` (expected: absent). Record
   each explicitly, not just the ones that are present.
8. Confirm `github.com/sethvargo/go-retry v0.3.0` is present as an indirect dependency (re-confirms
   REVIEW's Stage-7 adjudication correcting an earlier auditor claim it was absent) and record it as
   present-and-unused.
9. Confirm the `yaml.v3` / `go.yaml.in/yaml` watch item: record that `go.yaml.in/yaml/v3` is present
   as an indirect dependency, consistent with REVIEW §L's "community fork already indirect" note; no
   action required, monitor-only per REVIEW §L.
10. Search the full `go list -m all` output for every REVIEW §M rejected dependency (`viper`,
    `envconfig` under any known module path, a NATS or Kafka client, any password-hashing library
    such as `bcrypt`-wrapping third-party modules beyond Go's own `golang.org/x/crypto/bcrypt`, if
    present) and confirm none appear; if any unexpectedly appears, flag it as drift requiring
    escalation rather than silently noting it.
11. Write the dependency-inventory document consolidating the above into a single reviewable
    artifact.

### Expected files or components affected

No existing repository file is modified. New file(s) created: a dependency-inventory document
under this story's `artifacts/` tree (exact path finalized at execution time, per
`artifact-policy.md`'s lifecycle-stage subdirectories created only on first real content).

### Expected output

A dependency-inventory document with: raw command output (or a pointer to where it is stored),
a complete direct-dependency disposition table, explicit confirmation of the three new-approval
packages' presence/absence, explicit confirmation of REVIEW §M's rejected dependencies' absence,
and the resolved "10 vs 13" reconciliation statement.

### Required artifacts

Dependency-inventory document (registered in `../artifacts/index.md` once produced).

### Required evidence

Raw `go list -m all` / `go mod graph` / `go list -m -json all` output; the cross-check
table/diff itself, as a distinct evidence item (registered in `../evidence/index.md` once
produced).

### Related acceptance criteria

AC-W00-E02-S002-01, AC-W00-E02-S002-02.

### Completion criteria

Every direct dependency in `go.mod` has exactly one disposition entry; the "10 vs 13" question is
explicitly resolved (not left ambiguous); REVIEW §M's rejected list and the three new-approval
packages are each explicitly addressed; the dependency-inventory document and its supporting
evidence are registered.

### Verification method

Independent reviewer re-runs `go list -m all` against the same commit SHA cited in the evidence
record and spot-confirms the disposition table matches; reviewer confirms no direct dependency was
omitted from the table by counting `go.mod` require-block lines against table rows.

### Risks

Inherits `RISK-W00-E02-001` (epic-level register): cross-check performed as a sample rather than a
full enumeration would miss drift. Mitigated by this task's explicit "enumerate every direct
dependency" instruction (step 5) rather than a spot-check.

### Rollback or recovery considerations

Not applicable — this task produces documentation only; if the produced document is later found
incorrect, it is corrected via a new evidence record marked to supersede the earlier one
(`evidence-policy.md`), not rolled back.

## Implementation Record

*Per mandate §8.7. Do not pre-populate implementation claims for work that has not yet occurred.*

### What was actually implemented

Executed 2026-07-13 at commit `0a31186cada5c275a588c74081cf977adf346e61`. Ran `go list -m all`
(340 lines), `go mod graph` (715 edges), `go list -m -json all`, and targeted `go mod why -m`
provenance checks; raw output stored in `../evidence/logs/`. Built the full 13-row
direct-dependency disposition table against REVIEW §L/§M — all 13 rows `approved`, zero drift;
resolved the "10 vs 13" question (§L's "otel×4" = go.mod:16–19, so 13 lines = 10 logical deps);
recorded the new-approval trio (backoff/v5 present-indirect v5.0.3; golang-lru/v2 and gobreaker
absent); re-confirmed sethvargo/go-retry v0.3.0 present-indirect via goose/v3; confirmed the
yaml.v3 / go.yaml.in/yaml/v3 watch state; confirmed all four §M rejected entries absent from
go.mod (viper appears only in the unpruned module graph via minio-go's own go.mod, not needed by
the main module). Wrote `../artifacts/post-implementation/dependency-inventory.md`.

### Components changed

None (documentation/evidence only, as planned).

### Files changed

New: `../artifacts/post-implementation/dependency-inventory.md`,
`../evidence/logs/go-list-m-all.txt`, `../evidence/logs/go-mod-graph.txt`,
`../evidence/logs/go-list-m-json-all.txt`, `../evidence/logs/go-mod-why.txt`,
`../evidence/reviews/dependency-crosscheck.md`. No existing repository file modified.

### Interfaces introduced or changed

None.

### Configuration changes

None.

### Schema or migration changes

None.

### Security changes

None (the §M absence check is verification of an existing control, not a new one).

### Observability changes

None.

### Tests added or modified

None.

### Commits

Executed at existing commit `0a31186cada5c275a588c74081cf977adf346e61`; commit of these
documentation files is handled by the conductor's normal PR process.

### Pull requests

None yet — documentation-only change, to be raised by the conductor.

### Implementation dates

2026-07-13 (single session).

### Technical debt introduced

None.

### Known limitations

License claims are cross-checked against REVIEW §L's stated licenses, not re-derived — no
license-scanning tool is pinned in this repository's toolchain (confirmed during Task 002's
inspection), matching the task-definition fallback in step 4.

### Follow-up items

None.

### Relationship to the approved plan

Matches `../plan.md` exactly (commands run, disposition table produced, unresolved question
resolved). No deviation.

## Verification Record

*Per mandate §8.8. Table below is planned before execution; fields after it are filled after
execution.*

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W00-E02-S002-01 | Re-run `go list -m all` at the cited commit SHA; confirm every direct dependency has a disposition row | Local dev or CI runner, network access to Go module proxy | Complete disposition table, zero unaddressed direct dependencies | Command-output + cross-check table | unassigned |
| AC-W00-E02-S002-02 | Search captured output for REVIEW §M rejected deps and the three new-approval packages | Same as above | Rejected deps confirmed absent (or flagged); new-approval packages' presence/absence explicitly recorded | Command-output + cross-check table | unassigned |

### Actual result

All 13 direct dependencies dispositioned `approved`; zero undocumented drift; §M rejected deps
confirmed absent; new-approval trio explicitly recorded; "10 vs 13" resolved (otel×4).

### Pass or fail

**Pass** (AC-W00-E02-S002-01 and AC-W00-E02-S002-02).

### Evidence identifier

EV-W00-E02-S002-001, EV-W00-E02-S002-002.

### Execution date

2026-07-13.

### Commit or revision

`0a31186cada5c275a588c74081cf977adf346e61` (branch `main`).

### Environment

macOS 26.5.2 (Darwin 25.5.0), arm64, go1.26.5 darwin/arm64, local workstation; concurrent
sibling-worker test load present (non-timing evidence, unaffected).

### Reviewer

Reviewer unassigned — conductor acceptance gate pending.

### Findings

No drift found. Notable (non-drift) observation: `spf13/viper` and `hashicorp/golang-lru/v2`
appear in the unpruned module graph solely via `minio-go/v7@v7.2.1`'s own go.mod requirements;
`go mod why -m` confirms the main module does not need either — recorded explicitly in the
artifact so a future reader does not mistake the graph appearance for go.mod drift.

### Retest status

Not required — first run passed.

### Final conclusion

Task complete; both related ACs pass with registered evidence.

## Deviations Record

*Per mandate §8.9. Initially state that deviations are not yet known. The approved plan must not
be silently altered to hide deviations.*

No deviations. Execution matched the task definition and `../plan.md` exactly.

### Deviation ID

Not applicable — no deviation recorded.

### Approved plan

Not applicable.

### Actual implementation

Not applicable.

### Reason

Not applicable.

### Impact

Not applicable.

### Risks

Not applicable.

### Approval

Not applicable.

### Compensating controls

Not applicable.

### Follow-up work

Not applicable.
