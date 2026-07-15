---
id: W01-E01-S002-T001
type: task
title: gosec G704 annotation (JWKS taint, 2 sites)
status: done
parent_story: W01-E01-S002
owner: W01Lint
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on: []
acceptance_criteria:
  - AC-W01-E01-S002-02
artifacts:
  - ART-W01-E01-S002-002
evidence:
  - EV-W01-E01-S002-001
  - EV-W01-E01-S002-002
---

# W01-E01-S002-T001 — gosec G704 annotation (JWKS taint, 2 sites)

## Task Definition

### Task objective

Fresh-run gosec (and, as the shared baseline step for this story, errorlint/exhaustive/
forcetypeassert/usestdlibvars alongside it) at this story's actual start commit, then annotate the 2
named G704 hits at `kernel/auth/jwks.go:204,210` with an inline justification comment referencing
SEC-06. This is a governed, deliberate pattern — annotate, do not "fix" (change) it.

### Parent story

W01-E01-S002 — Enable and triage the judged linter set.

### Owner

unassigned

### Status

todo

### Dependencies

None — independent of T002-T006 (disjoint files/sites). This task's fresh-run step is shared
evidence infrastructure for the whole story (EV-W01-E01-S002-001), but does not block T002-T006 from
also being worked in parallel, since each of those tasks can independently re-run its own analyzer
scope if needed.

### Detailed work

1. Run `golangci-lint run --enable=gosec,errorlint,exhaustive,forcetypeassert,usestdlibvars ./...`
   against the current HEAD, before editing `.golangci.yml` — the fail-first evidence step for the
   whole story. Record the full output as EV-W01-E01-S002-001, regardless of whether it matches the
   cited "38 hits."
2. From that output, isolate the G704 hits. Confirm they are still at `kernel/auth/jwks.go:204,210`
   (or record the actual current line numbers if drifted).
3. Confirm the pinned gosec/golangci-lint v2.11.4 version's supported `#nosec` suppression-comment
   syntax (resolving `plan.md`'s "Unresolved questions" item on this point).
4. At each of the two sites, add an inline `#nosec` (with the confirmed rule-ID suffix if the pinned
   version supports/requires one) justification comment stating that this fetch is governed by SEC-06
   (outbound-security escape-hatch governance, D-07 ratified) — i.e., this is a deliberate,
   trusted-issuer JWKS fetch pattern, not an unreviewed taint.
5. Re-run `golangci-lint run --enable=gosec ./kernel/auth/...` to confirm the two sites no longer
   surface as unaddressed hits (gosec's own reporting behavior for annotated `#nosec` lines — exits 0
   or reports the site as suppressed, depending on the pinned version's output format — to be
   confirmed and recorded as the "after" state).

### Expected files or components affected

`kernel/auth/jwks.go`.

### Expected output

Both G704 sites carry an inline `#nosec` comment referencing SEC-06; the fresh-run baseline
(EV-W01-E01-S002-001) and the G704-specific fail-before/pass-after pair (EV-W01-E01-S002-002) are
both captured.

### Required artifacts

ART-W01-E01-S002-002 (updated `kernel/auth/jwks.go`, G704 annotation).

### Required evidence

EV-W01-E01-S002-001 (story-wide fresh-run baseline, produced by this task), EV-W01-E01-S002-002
(G704-specific fail-before/pass-after pair).

### Related acceptance criteria

AC-W01-E01-S002-02.

### Completion criteria

Both named G704 sites carry an accurate, SEC-06-referencing annotation; the fresh-run baseline is
captured and compared against the cited "38 hits"/named-site claims in `story.md`, with any drift
recorded in `deviations.md`.

### Verification method

Direct command execution (`golangci-lint run --enable=gosec`), logged output retained as evidence;
manual review confirming the annotation text accurately references SEC-06 and does not overstate or
understate the governed pattern.

### Risks

Low for the annotation itself (a single, already-adjudicated governed pattern). The task's fresh-run
step carries RISK-W01-E01-002 (the run may surface more than the cited "38 hits" story-wide) — see
epic-level `risks.md`.

### Rollback or recovery considerations

Revert the annotation if a reviewer determines the SEC-06 reference is inaccurate or the pattern is
not actually governed as described; escalate to re-open the G704 disposition as a real finding rather
than silently re-annotating without investigation.

## Implementation Record

Implemented 2026-07-13 by W01Lint (working diff on HEAD `0a31186cada5c275a588c74081cf977adf346e61`; conductor owns commits).

`kernel/auth/jwks.go` get(): both G704 sites annotated `#nosec G704` with an explanatory comment referencing SEC-06 (D-07) — trusted-issuer JWKS/discovery URI from boot config constrained by validateHTTPSURL. Pattern NOT changed. Also produced the shared fresh-run baseline (Phase-1 HEAD + Phase-2 enablement-state enumerations).

## Verification Record

AC-W01-E01-S002-02 (G704 slice): hits present in both fail-before runs; gosec per-linter run exit 0 after annotation (EV-002). **pass**

### Final conclusion

Governed pattern preserved and justified inline.

## Deviations Record

None — see story-level `deviations.md` for story-wide drift records.
