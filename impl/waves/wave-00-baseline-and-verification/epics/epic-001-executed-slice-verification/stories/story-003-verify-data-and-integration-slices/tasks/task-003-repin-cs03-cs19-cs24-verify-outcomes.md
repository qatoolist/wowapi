---
id: W00-E01-S003-T003
type: task
title: Re-pin CS-03/CS-19/CS-24 matrix verify-outcomes
status: done
parent_story: W00-E01-S003
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on: []
acceptance_criteria: [AC-W00-E01-S003-03]
artifacts: [ART-W00-E01-S003-005]
evidence: [EV-W00-E01-S003-03]
---

# W00-E01-S003-T003 — Re-pin CS-03/CS-19/CS-24 matrix verify-outcomes

## Task Definition

*Per mandate §8.6. This file defines the task before work begins. The implementation record,
verification record, and deviations record for this task are the `##`-level sections below in this
same file, per `governance/naming-conventions.md` "Adaptation 1" (flat single-file tasks).*

### Task objective

Re-confirm, at the current repository HEAD, that the three MATRIX verify-outcome rows
`requirement-inventory.md` §C records with disposition `INV→verified` — CS-03 (config fail-closed +
fingerprint), CS-19 (i18n freeze + key-echo fallback), and CS-24 (SSRF dial-time guard) — still hold,
by locating and re-running (or re-inspecting, per each claim's original verification basis) the
test(s) or code path MATRIX cites for each, and register a mandate-§10-conformant evidence pointer
confirming each. This task exists in this story because `epic.md` "Scope" assigns the CS-03/CS-19/
CS-24 re-pinning to S003, on the grounds that "does the current repository state still match what the
documents claim" is the same kind of check whether the claim originates in PLAN (DATA-08, REL-04) or
in MATRIX (CS-03, CS-19, CS-24); this task is split out separately from Task 2 rather than folded
into it because Task 2 already combines four distinct verification methods (S3-gated test re-run,
TOTP determinism re-run, CI-workflow inspection for SD-01, CI-workflow inspection for SD-02) and
adding three more independent claims with their own separate verification bases would make Task 2
unreasonably broad and harder to review or fail independently — see `../../plan.md` "Approval
conditions" for the explicit note on this judgment call.

### Parent story

W00-E01-S003 — Verify data-durability and CI-integration slices at current HEAD.

### Owner

unassigned

### Status

`done` — executed and evidenced 2026-07-13; awaiting the conductor's story-level review gate.

### Dependencies

None hard. This task targets whichever source files implement CS-03, CS-19, and CS-24 (exact paths
to be confirmed during execution by following MATRIX's own citations for each), disjoint in scope
intent from Task 1's `kernel/attachment`/`kernel/notify` and Task 2's S3/TOTP/CI-pipeline focus. May
execute in any order relative to T001 and T002, including fully in parallel. A soft, non-blocking
convenience overlap exists with T002: if CS-24's SSRF dial-time-guard citation happens to touch
`.github/workflows/ci.yml`-adjacent config (e.g. a gosec G704 annotation named in
`requirement-inventory.md` §C's FBL-07 note), T002 may already have that file open — this is a
convenience, not a dependency; T003 does not require T002 to have run first.

### Detailed work

- For **CS-03** (config fail-closed + fingerprint): locate the specific test(s) or code path MATRIX
  cites as the basis for its `INV→verified` disposition; re-run or re-inspect per that basis; confirm
  the fail-closed behavior and fingerprint mechanism are still present and correct at current HEAD.
- For **CS-19** (i18n freeze + key-echo fallback): locate the specific test(s) or code path MATRIX
  cites; re-run or re-inspect per that basis; confirm the freeze behavior and key-echo fallback are
  still present and correct at current HEAD.
- For **CS-24** (SSRF dial-time guard): locate the specific test(s) or code path MATRIX cites; re-run
  or re-inspect per that basis; confirm the dial-time guard is still present and correct at current
  HEAD. Note `requirement-inventory.md` §C's associated remark that a "gosec G704 annotation task"
  lives inside FBL-07 — that annotation task itself is out of scope here (tracked at FBL-07's own
  target); this task only re-confirms CS-24's verify-outcome claim.
- For each of the three, record the exact commit SHA, the exact command or inspection method used
  (following, not inventing, the verification basis MATRIX originally used), the environment, tool
  versions, date/time, and result.
- If any of the three no longer holds, treat this as a regression in an already-`verified` security
  finding, not a routine re-verification failure — flag it immediately as a new finding per
  `../../story.md` "Risks," rather than opening a routine follow-up task silently.

### Expected files or components affected

None changed. Files read and re-tested/re-inspected: whichever source and test files MATRIX cites
for CS-03, CS-19, and CS-24 individually — exact paths to be confirmed during execution, not assumed
in advance (per `../../plan.md` "Unresolved questions").

### Expected output

A `pass`/`still holds` or `failed`/`regressed` result for each of CS-03, CS-19, and CS-24
individually, captured as a verify-outcome re-pin note citing the specific evidence pointer for each,
with a corresponding evidence record.

### Required artifacts

CS-03/CS-19/CS-24 verify-outcome re-pin note — see `../../artifacts/index.md`.

### Required evidence

One evidence record, planned ID `EV-W00-E01-S003-03`, evidence type "verify-outcome re-pin note with
evidence pointers" — see `../../evidence/index.md`.

### Related acceptance criteria

AC-W00-E01-S003-03.

### Completion criteria

This task is complete when: each of CS-03, CS-19, and CS-24 has been individually re-confirmed (not
merely asserted) against a confirmed commit SHA, following each claim's own original verification
basis as cited by MATRIX; the result for each is recorded in `verification.md` and in the story's
`verification.md`; the evidence record is registered in `../../evidence/index.md` with all required
fields per `evidence-policy.md`; and, if any claim no longer holds, it has been flagged as a new
finding (escalated, not silently absorbed into routine follow-up handling) per `../../story.md`
"Risks."

### Verification method

Locate MATRIX's own citation (test file, code path, or inspection method) for each of CS-03, CS-19,
and CS-24 individually, and re-execute or re-inspect exactly that basis at current HEAD — not a
newly invented verification method, since the claim to re-pin is specifically "the original basis
still holds," not "some other check now passes."

### Risks

- RISK-W00-001 (inherited, security-elevated for this task) — any of CS-03/CS-19/CS-24 fails to
  re-verify; unlike the epic's other re-verification tasks, a failure here is a regression in an
  already-`verified` security finding (fail-closed config, i18n freeze, SSRF guard), warranting
  immediate escalation rather than routine follow-up-task handling — see `../../story.md` "Risks"
  (the story-specific risk not yet assigned an ID).
- RISK-W00-002 (inherited) — if any of the three claims' original verification basis depends on
  infrastructure (DB, network) unavailable in the execution environment, a false-negative regression
  could result; must be ruled out before treating any failure as genuine.

### Rollback or recovery considerations

Not applicable in the code sense (this task changes no code). If a re-pin fails, no rollback occurs
— the failure is recorded as `failed`-status evidence (preserved, not deleted per
`evidence-policy.md`) and escalated immediately as a new finding given the security nature of all
three claims, rather than opened as a routine remediation task under a distant future-wave story.

## Implementation Record

*Per mandate §8.7.* Executed 2026-07-13 against commit
`0a31186cada5c275a588c74081cf977adf346e61` (branch `main`).

### What was actually implemented

Verification-only re-pin; no implementation. MATRIX
(`docs/implementation/fable5-closure-depth-matrix-2026-07-11.md`) records all three rows as
verify-outcomes whose "citations in the CS body are the record" (§2.1 lines 282/298/303) — the
original verification basis is code inspection at cited file:line ranges, so this task re-inspected
exactly those citations at HEAD (not a newly invented method), corroborated by re-running each
package's unit-test suite:

- **CS-03 (config fail-closed + fingerprint) — STILL HOLDS.** `kernel/config/load.go:132-139`
  (missing-environment and prod-flag-override errors, exact lines unchanged);
  `kernel/config/config.go:254-266` `Validate()` prod safety floor aggregated via
  `errors.Join`; `kernel/config/fingerprint.go:18,29-35` SHA-256 over canonical Secret-redacted
  JSON, with the documented rotation-limitation note intact.
- **CS-19 (i18n freeze + key-echo fallback) — STILL HOLDS.** `kernel/i18n/embed.go:11-28` layered
  embedded defaults; `kernel/i18n/catalog.go:29-41` precedence chain; `:98-103` real freeze seal;
  `:82-92` post-freeze `Add` no-op; `:108-134` missing-key fallback exact → default-locale →
  echoes the key, never erroring. All exact-line matches.
- **CS-24 (SSRF dial-time guard) — STILL HOLDS.** `kernel/httpclient/client.go:85-88` custom
  `DialContext` installed; `:178-209` resolve → check resolved IPs → dial the verified IP
  (DNS-rebinding TOCTOU closed; redirects re-enter the dialer per connection); `:71-84`
  `transport.Proxy = nil` (env-proxy bypass closed); `kernel/httpclient/guard.go:60-73`
  IPv6-embedded-v4 unwrapping, `:97-124` fail-closed `isBlockedIP`;
  `kernel/config/config.go:261-263` disable flag rejected in prod.

Corroboration: `go test kernel/config kernel/i18n kernel/httpclient -count=1 -v` — all `ok`,
199 top-level PASS, 0 FAIL, 0 SKIP, exit 0. Full re-pin note with the three single-line citation
drifts recorded: `../../evidence/logs/t003-cs-repin-note.md`.

### Components changed

None — verification-only task, as planned.

### Files changed

None. Only this story directory's own governance/evidence files were written.

### Interfaces introduced or changed

None.

### Configuration changes

None.

### Schema or migration changes

None.

### Security changes

None — the three existing security findings were re-confirmed without modification.

### Observability changes

None.

### Tests added or modified

None — existing tests/code paths re-inspected and re-run, not modified.

### Commits

None made by this task (read-only against `0a31186cada5c275a588c74081cf977adf346e61`).

### Pull requests

None.

### Implementation dates

2026-07-13 (single session).

### Technical debt introduced

None.

### Known limitations

Point-in-time re-pin at `0a31186`; three of the MATRIX line citations drifted by one line
(identical code) — recorded in the re-pin note, not silently normalised.

### Follow-up items

None — all three claims hold; no security escalation needed.

### Relationship to the approved plan

Executed per `../../plan.md` Task 3 strategy: MATRIX's own citation basis located and followed
(inspection), answering the plan's unresolved question about each claim's basis; the package-test
corroboration is additive, not a substitute basis.

## Verification Record

*Per mandate §8.8. Table below is planned before execution; fields after it are filled after
execution.*

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W00-E01-S003-03 | Locate and re-run/re-inspect the test(s)/code path MATRIX cites for CS-03, CS-19, and CS-24 individually; confirm each claim still holds at the story's closing commit | Environment per each claim's original verification basis (to be confirmed during execution) | All three claims re-confirmed with an evidence pointer; any regression flagged as a new finding | Verify-outcome re-pin note with evidence pointers | unassigned |

### Actual result

CS-03, CS-19, CS-24 each individually re-confirmed at `0a31186` against MATRIX's own citation
basis (re-inspection of the cited file:line ranges; majority verbatim, three single-line drifts
with identical code). Corroborating package suites (`kernel/config`, `kernel/i18n`,
`kernel/httpclient`): exit 0, 199 PASS / 0 FAIL / 0 SKIP.

### Pass or fail

**Pass — all three still hold.** No regression to escalate.

### Evidence identifier

`EV-W00-E01-S003-03` — registered in `../../evidence/index.md`; raw files
`../../evidence/logs/t003-cs-repin-note.md`, `t003-cs-repin-package-tests.log`.

### Execution date

2026-07-13 12:14 +0530.

### Commit or revision

`0a31186cada5c275a588c74081cf977adf346e61` (branch `main`).

### Environment

Local macOS host (darwin/arm64, macOS 26.5.2), go1.26.5. Inspection is environment-independent;
the three corroborating package suites need no DB/S3/network. Concurrent load present (sibling
W00 workers) — evidence is not timing-sensitive.

### Reviewer

Unassigned — acceptance is the conductor's review gate; not self-assigned.

### Findings

No regression in any of the three already-`verified` security findings. Minor: three MATRIX line
citations drifted by one line each (config.go Validate tail; client.go Proxy/DialContext block)
with identical code content — a citation-freshness note, not a code change.

### Retest status

Not applicable — first pass confirmed all claims; nothing retried.

### Final conclusion

AC-W00-E01-S003-03 **satisfied**: CS-03, CS-19, and CS-24 verify-outcome claims re-pinned at the
story's closing commit with mandate-§10-conformant evidence pointers.

## Deviations Record

*Per mandate §8.9.*

**No deviations.** The re-pin followed MATRIX's own citation basis exactly as the task definition
requires; the corroborating package-test runs are additional evidence, not a substituted method.

### Deviation ID

Not applicable — no deviations occurred.

### Approved plan

Not applicable — executed as planned.

### Actual implementation

Not applicable — matches the plan.

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
