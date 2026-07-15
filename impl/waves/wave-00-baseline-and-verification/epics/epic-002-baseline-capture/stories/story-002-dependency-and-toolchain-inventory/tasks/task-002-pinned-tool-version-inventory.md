---
id: W00-E02-S002-T002
type: task
title: Pinned tool-version inventory
status: done
parent_story: W00-E02-S002
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on: []
acceptance_criteria:
  - AC-W00-E02-S002-03
artifacts:
  - ART-W00-E02-S002-002
evidence:
  - EV-W00-E02-S002-003
---

# W00-E02-S002-T002 — Pinned tool-version inventory

## Task Definition

*Per mandate §8.6. This section defines the task before work begins.*

### Task objective

Determine and record the pinned versions of `golangci-lint`, GoReleaser, Trivy, and `goose/v3` as
actually configured in this repository's own tooling (`Makefile`, CI workflow files, lint
configuration, `go.mod`), producing a tool-version-inventory document. Any version that cannot be
confirmed from this repository's own configuration must be recorded as unconfirmed/TBD rather than
invented or assumed from a secondhand citation.

### Parent story

W00-E02-S002 — Dependency and toolchain inventory.

### Owner

Unassigned.

### Status

`done` — executed 2026-07-13 (per `impl/governance/status-model.md`).

### Dependencies

None. Independent of Task 001.

### Detailed work

1. Inspect `Makefile` directly for a pinned `golangci-lint` version (do not trust the `v2.11.4`
   figure cited secondhand in `../../../wave.md` / `../../../dependencies.md` without re-confirming
   it against the `Makefile` at this task's own execution commit).
2. If a `golangci-lint` binary is available in the execution environment, run
   `golangci-lint --version` and cross-confirm it matches the `Makefile` pin; record both the
   configured pin and the actually-installed version, flagging any mismatch.
3. Inspect `.github/workflows/*.yml` (and any release-specific workflow, e.g. a `release.yml`) for
   a pinned GoReleaser version or action reference (e.g. a `goreleaser/goreleaser-action@vN` pin,
   or a `go install github.com/goreleaser/goreleaser@vN` line). Record whatever is actually found;
   if no explicit version pin exists anywhere in the repository's tooling, record that fact
   explicitly as "no pin found" rather than inventing a plausible-sounding version number. This
   directly resolves the TBD flagged in `../story.md` and `../plan.md`.
4. Inspect `.github/workflows/*.yml` and any dedicated security-scan configuration for a pinned
   Trivy version and its scanner configuration (e.g. which scan types are enabled — vuln, secret,
   misconfig — and severity thresholds). Record whatever is actually found, or "no pin found" if
   Trivy is invoked via an unpinned action/image tag (e.g. `:latest`), flagging that itself as a
   fact worth noting (not a judgment on whether it should be pinned — that is out of this task's
   scope, it only records current state).
5. Confirm `goose/v3`'s version directly from `go.mod` (`v3.27.2` at story-authoring time,
   re-confirm at this task's own execution commit) — this one is expected to be straightforward
   since it is a Go module dependency, not a standalone pinned CLI tool.
6. Write the tool-version-inventory document consolidating all four tools' confirmed (or
   explicitly-unconfirmed) versions with citations to the exact file (and line, where practical)
   each was found in.

### Expected files or components affected

No existing repository file is modified. New file created: a tool-version-inventory document under
this story's `artifacts/` tree (exact path finalized at execution time).

### Expected output

A tool-version-inventory document stating, for each of `golangci-lint`, GoReleaser, Trivy, and
`goose/v3`: the confirmed pinned version and its exact source citation, or an explicit
"unconfirmed/no pin found" statement if no pin exists in this repository's configuration.

### Required artifacts

Tool-version-inventory document (registered in `../artifacts/index.md` once produced).

### Required evidence

Command output (e.g. `golangci-lint --version`) where a binary is available; configuration-file
citation (file path + line reference) for each of the four tools' version source, whether a pin was
found or not (registered in `../evidence/index.md` once produced).

### Related acceptance criteria

AC-W00-E02-S002-03.

### Completion criteria

All four tool versions are addressed with an explicit outcome (confirmed-with-citation, or
explicitly unconfirmed/no-pin-found) — none silently omitted; the tool-version-inventory document
and its supporting evidence are registered.

### Verification method

Independent reviewer re-inspects the cited `Makefile`/CI-workflow lines directly and confirms the
recorded version (or "no pin found" claim) matches what is actually in those files at the cited
commit SHA.

### Risks

Low — the main risk is asserting a version without a source citation (mitigated by this task's
explicit citation requirement) or silently omitting GoReleaser/Trivy if no pin is found, rather than
recording that absence as a fact. This task's detailed-work steps 3 and 4 explicitly require
recording "no pin found" rather than skipping the tool.

### Rollback or recovery considerations

Not applicable — documentation-only output; corrections happen via a superseding evidence record,
not a rollback.

## Implementation Record

*Per mandate §8.7. Do not pre-populate implementation claims for work that has not yet occurred.*

### What was actually implemented

Executed 2026-07-13 at commit `0a31186cada5c275a588c74081cf977adf346e61`. Inspected `Makefile`,
`.github/workflows/{ci,release,security-scan,vuln}.yml`, `go.mod`, and `deployments/compose.yaml`
directly; ran local version-check commands (`golangci-lint version`, `goreleaser --version`,
`trivy --version`, `go version` — captured in `../evidence/logs/tool-versions.txt`). Findings:
golangci-lint pinned **v2.11.4** (`Makefile:16`, lockstep `ci.yml:62`; local binary 2.11.4
matches — no mismatch); GoReleaser — **no exact binary pin**: release path uses SHA-pinned
`goreleaser/goreleaser-action` v7.2.3 with floating `version: "~> v2"` (`release.yml:47–50`),
local Makefile targets deliberately `@latest` (`Makefile:344–362`); Trivy — SHA-pinned
`aquasecurity/trivy-action` v0.36.0 with no explicit binary-version input, scanner config
`fs` / `vuln,secret,misconfig` / `CRITICAL,HIGH` / `ignore-unfixed` / non-blocking `exit-code: 0`
(`security-scan.yml:68–75`); goose/v3 **v3.27.2** (`go.mod:13`). Wrote
`../artifacts/post-implementation/tool-version-inventory.md`.

### Components changed

None (documentation/evidence only, as planned).

### Files changed

New: `../artifacts/post-implementation/tool-version-inventory.md`,
`../evidence/logs/tool-versions.txt`. No existing repository file modified.

### Interfaces introduced or changed

None.

### Configuration changes

None.

### Schema or migration changes

None.

### Security changes

None.

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

Local `goreleaser`/`trivy` binary versions (v2.16.0 / 0.72.0) are recorded as informational
only — they are workstation state, not repository pins, and are labelled as such in the artifact.

### Follow-up items

GoReleaser and Trivy indeed have no exact binary pin (action-SHA + floating range instead) — per
this task's own scope note, that fact is recorded, not fixed; `W06-E03-S001` (REL-01 T6) /
`W06-E03-S003` (REL-02) are the candidate owners if a pin is later deemed necessary.

### Relationship to the approved plan

Matches `../plan.md` exactly (all four tools addressed with citations; TBDs resolved as
recorded-as-found facts). No deviation.

## Verification Record

*Per mandate §8.8. Table below is planned before execution; fields after it are filled after
execution.*

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W00-E02-S002-03 | Re-inspect cited `Makefile`/CI-workflow lines at the cited commit SHA; re-run version-check commands where available | Local dev or CI runner with relevant binaries, or read-only file inspection | Each of the four tool versions is either confirmed-with-citation or explicitly recorded as unconfirmed/no-pin-found — none silently omitted | Command-output + configuration-file citation evidence | unassigned |

### Actual result

All four tools addressed explicitly: golangci-lint confirmed pinned v2.11.4 (with matching local
binary); GoReleaser recorded as no-exact-binary-pin (SHA-pinned action, `~> v2`); Trivy recorded
as SHA-pinned trivy-action v0.36.0 with scanner config and no explicit binary pin; goose/v3
confirmed v3.27.2. None silently omitted.

### Pass or fail

**Pass** (AC-W00-E02-S002-03).

### Evidence identifier

EV-W00-E02-S002-003.

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

The secondhand `v2.11.4` citation in `wave.md`/`dependencies.md` is confirmed accurate at this
commit. The absence of exact GoReleaser/Trivy binary pins is a recorded fact, not drift — the
repository's own comments (`Makefile:344–348`, `security-scan.yml:64–67`) document the posture
deliberately.

### Retest status

Not required — first run passed.

### Final conclusion

Task complete; AC-W00-E02-S002-03 passes with registered evidence.

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
