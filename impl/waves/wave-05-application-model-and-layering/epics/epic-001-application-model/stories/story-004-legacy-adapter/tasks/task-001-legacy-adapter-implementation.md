---
id: W05-E01-S004-T001
type: task
title: Legacy adapter implementation
status: todo
parent_story: W05-E01-S004
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on: []
acceptance_criteria:
  - AC-W05-E01-S004-01
  - AC-W05-E01-S004-02
artifacts:
  - ART-W05-E01-S004-001
  - ART-W05-E01-S004-002
evidence:
  - EV-W05-E01-S004-001
  - EV-W05-E01-S004-002
---

# W05-E01-S004-T001 — Legacy adapter implementation

## Task Definition

### Task objective

Implement the legacy adapter wrapping `module.Module`/`Context`, deriving owner from
`Module.Name()` and routing every registration call through S002's owner-bound registrars, proven
both by existing contract tests passing unmodified and by S002's adversarial fixtures rejecting
cross-module claims identically through the legacy path.

### Parent story

W05-E01-S004 — Legacy module/context compatibility adapter.

### Owner

unassigned

### Status

todo

### Dependencies

None (depends on W05-E01-S003 at story scope — the full T1-T10 model surface).

### Detailed work

1. Re-read the current `module.Module`/`Context` implementation at this task's start commit.
2. Implement the adapter: for each registration call surface, mint a `Registrar` from the calling
   module's `Module.Name()` and route the call through S002's corresponding owner-bound wrapper.
3. Re-run S002's adversarial fixtures (resource, rules, authz, full declaration-class matrix)
   through the legacy path, confirming identical rejection behavior to the non-legacy path.
4. Run existing wowapi-internal and wowsociety module contract tests unmodified through the legacy
   path, capturing `AR-01/legacy_adapter_compat_test_output.txt`.
5. Document the adapter's owner-derivation mechanism and non-bypass guarantee.

### Expected files or components affected

A new adapter layer within or adjacent to `kernel/module` (exact path TBD per `plan.md`).

### Expected output

A legacy adapter proven both compatible (existing tests pass) and non-bypassing (adversarial
fixtures reject identically through the legacy path).

### Required artifacts

ART-W05-E01-S004-001 (adapter), ART-W05-E01-S004-002 (documentation).

### Required evidence

EV-W05-E01-S004-001 (compat-test output), EV-W05-E01-S004-002 (adversarial-fixtures-through-legacy-
path output).

### Related acceptance criteria

AC-W05-E01-S004-01, AC-W05-E01-S004-02.

### Completion criteria

Both proofs pass: existing contract tests unmodified through the legacy path; S002's adversarial
fixtures reject identically through the legacy path as through the non-legacy path.

### Verification method

Direct execution of both test suites; wowsociety's own build and contract-test suite run as part of
the compatibility proof.

### Risks

Medium, per PLAN T11's own risk column — "the adapter is itself a trust boundary." See
RISK-W05-E01-003 in epic-level `risks.md`.

### Rollback or recovery considerations

If the adversarial-fixtures-through-legacy-path proof reveals a bypass, do not ship — fix the
owner-derivation or routing logic before proceeding.

## Implementation Record

*Not yet implemented.*

### What was actually implemented

*Not yet implemented.*

### Components changed

*Not yet implemented.*

### Files changed

*Not yet implemented.*

### Interfaces introduced or changed

*Not yet implemented.*

### Configuration changes

*Not yet implemented.*

### Schema or migration changes

*Not applicable.*

### Security changes

*Not yet implemented — the non-bypass guarantee is the security property this task delivers;
recorded here once implemented.*

### Observability changes

*Not applicable.*

### Tests added or modified

*Not yet implemented.*

### Commits

*Not yet implemented.*

### Pull requests

*Not yet implemented.*

### Implementation dates

*Not yet implemented.*

### Technical debt introduced

*None anticipated.*

### Known limitations

*Not yet implemented.*

### Follow-up items

*Not yet implemented.*

### Relationship to the approved plan

*Not yet implemented.*

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W05-E01-S004-01 | Run existing contract tests through the legacy path | Local dev or CI + wowsociety build | Existing tests pass unmodified | integration-test report | unassigned |
| AC-W05-E01-S004-02 | Re-run S002's adversarial fixtures through the legacy path | Local dev or CI, Go toolchain | Identical rejection behavior — no bypass | adversarial-test report | unassigned |

### Actual result

*Not yet executed.*

### Pass or fail

*Not yet executed.*

### Evidence identifier

*Not yet executed.*

### Execution date

*Not yet executed.*

### Commit or revision

*Not yet executed.*

### Environment

*Not yet executed.*

### Reviewer

*Not yet executed.*

### Findings

*Not yet executed.*

### Retest status

*Not yet executed.*

### Final conclusion

*Not yet executed.*

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
