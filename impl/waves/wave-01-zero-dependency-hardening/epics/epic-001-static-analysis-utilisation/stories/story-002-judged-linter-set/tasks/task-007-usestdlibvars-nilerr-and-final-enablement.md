---
id: W01-E01-S002-T007
type: task
title: usestdlibvars fixes, nilerr annotation, and final judged-set enablement
status: done
parent_story: W01-E01-S002
owner: W01Lint
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on:
  - W01-E01-S002-T001
  - W01-E01-S002-T002
  - W01-E01-S002-T003
  - W01-E01-S002-T004
  - W01-E01-S002-T005
  - W01-E01-S002-T006
acceptance_criteria:
  - AC-W01-E01-S002-01
  - AC-W01-E01-S002-06
  - AC-W01-E01-S002-07
artifacts:
  - ART-W01-E01-S002-001
  - ART-W01-E01-S002-010
  - ART-W01-E01-S002-011
evidence:
  - EV-W01-E01-S002-008
  - EV-W01-E01-S002-009
  - EV-W01-E01-S002-010
---

# W01-E01-S002-T007 — usestdlibvars fixes, nilerr annotation, and final judged-set enablement

## Task Definition

### Task objective

Close out the story's remaining code and configuration work that no per-analyzer triage task owns:
mechanically fix every site the fresh `usestdlibvars` run enumerates; add the fail-closed-intent
annotation at `kernel/policy/policy.go:166` (the `nilerr` non-finding — annotate only, do not alter
the logic); permanently enable `gosec`, `errorlint`, `exhaustive`, `forcetypeassert`, `usestdlibvars`
in `.golangci.yml`; confirm `wrapcheck`/`revive` remain absent from the enabled set; and run the final
full-module-tree confirmation run.

### Parent story

W01-E01-S002 — Enable and triage the judged linter set.

### Owner

unassigned

### Status

todo

### Dependencies

W01-E01-S002-T001 through -T006 — the final `.golangci.yml` enablement and confirmation run can only
exit 0 once every per-analyzer fix/annotation from the six triage tasks has landed. The usestdlibvars
fix portion of this task consumes the site list produced by T001's story-wide fresh-run baseline
(EV-W01-E01-S002-001) and can begin as soon as that baseline exists; the enablement/confirmation
portion is strictly last.

### Detailed work

1. From T001's fresh-run baseline output, extract the `usestdlibvars` hit list (no sites are named in
   the source material — the list is this step's input from that run, not an invention).
2. Fix each enumerated usestdlibvars site mechanically: replace the flagged literal with the
   equivalent stdlib constant (behavior-preserving by construction, since the constant's value equals
   the literal). Record the per-site list in the story's triage record.
3. Add an inline comment at `kernel/policy/policy.go:166` explaining the deliberate fail-closed
   design: an unparseable runtime value evaluates the governing condition to `false` (deny), and
   malformed-policy errors are handled separately at line 161. Personally adjudicated by Fable 5 as
   not a bug during the source architecture-review pass — this task annotates only; changing the
   condition logic is explicitly out of scope and would be a deviation.
4. Update `.golangci.yml` to permanently enable `gosec`, `errorlint`, `exhaustive`,
   `forcetypeassert`, `usestdlibvars` (plan steps 12-13).
5. Confirm `wrapcheck` and `revive` are absent from the enabled analyzer set — their rejection
   (classification REJ, disposition rejected: ~50 hits each at MATRIX-pass time, noise-dominant
   without disproportionate tuning investment) is recorded in `story.md` "Scope"/"Out of scope" as
   the authoritative record; this step verifies the config matches that record.
6. Run the final full-module-tree `golangci-lint run ./...` against the updated `.golangci.yml` and
   confirm exit code 0 across all five newly-enabled analyzers (EV-W01-E01-S002-010).

### Expected files or components affected

`.golangci.yml`; `kernel/policy/policy.go`; whatever files the usestdlibvars site enumeration
identifies (TBD until T001's fresh run executes).

### Expected output

A permanently updated `.golangci.yml`; all usestdlibvars sites fixed; the nilerr non-finding
annotated in place; a passing final confirmation run log.

### Required artifacts

ART-W01-E01-S002-001 (updated `.golangci.yml`), ART-W01-E01-S002-010 (usestdlibvars site fixes),
ART-W01-E01-S002-011 (`kernel/policy/policy.go` nilerr annotation).

### Required evidence

EV-W01-E01-S002-008 (usestdlibvars fail-before/pass-after pair), EV-W01-E01-S002-009 (nilerr
annotation + wrapcheck/revive-absence review note), EV-W01-E01-S002-010 (final combined
confirmation run).

### Related acceptance criteria

AC-W01-E01-S002-01, AC-W01-E01-S002-06, AC-W01-E01-S002-07.

### Completion criteria

`usestdlibvars` exits 0 with every enumerated site fixed and recorded; `kernel/policy/policy.go:166`
carries the fail-closed-intent annotation with unchanged logic; `.golangci.yml` enables all five
judged analyzers and excludes wrapcheck/revive; the final full-module-tree run exits 0.

### Verification method

Direct command execution (`golangci-lint run --enable=usestdlibvars ./...`, final
`golangci-lint run ./...`), logged output retained as evidence; manual review of the policy.go
annotation text (accuracy of the fail-closed explanation) and of the final `.golangci.yml` `enable:`
list against the story's scope record.

### Risks

Low for the mechanical portions. The one judgment-bearing element is the nilerr annotation's wording
— it must make the fail-closed intent explicit enough that a future reader does not mistake the
suppressed warning for an oversight (see `story.md` "Security considerations"). The final
confirmation run carries the aggregate of RISK-W01-E01-002 (any drift the six triage tasks did not
fully dispose of surfaces here, as a failing final run — which is this task's designed detection
point, not a task failure).

### Rollback or recovery considerations

The `.golangci.yml` enablement is revertible independently of the code fixes. If the final
confirmation run fails on a hit no task disposed of, the gap is triaged back to the owning task
(reopened) rather than silently annotated away inside this task.

## Implementation Record

Implemented 2026-07-13 by W01Lint (working diff on HEAD `0a31186cada5c275a588c74081cf977adf346e61`; conductor owns commits).

usestdlibvars: 9 test-file sites fixed mechanically (http.MethodGet/MethodPut/StatusOK + needed imports). nilerr: policy.go:166 annotated with the fail-closed explanation + `//nolint:nilerr` marker, logic untouched; nilerr itself NOT enabled (DEV-003 — the 'already-enabled' story claim was drift). wrapcheck/revive REJECTED with fresh counts 464/231, absence machine-checked. Final enablement: all five judged analyzers in `.golangci.yml`; full-tree run exit 0.

## Verification Record

AC-W01-E01-S002-01/-06/-07: final full-tree run exit 0 (EV-010); usestdlibvars per-linter exit 0 with all 9 sites recorded (EV-008); wrapcheck-revive-absence.txt + policy.go comment (EV-009). **pass**

### Final conclusion

Judged set fully enabled; every disposition recorded.

## Deviations Record

None — see story-level `deviations.md` for story-wide drift records.
