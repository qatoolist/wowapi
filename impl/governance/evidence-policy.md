---
id: GOV-EVIDENCE-POLICY
type: governance
title: Evidence policy — required fields, revision pinning, and preservation rules
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
derived: false
---

# Evidence policy

Mandate §10, "Evidence-management requirements." Evidence proves that implementation or
acceptance criteria are satisfied — it is distinct from artifacts (`artifact-policy.md`), which
are the things produced or consumed, not the proof that they behave correctly.

## Evidence examples (mandate §10, non-exhaustive)

unit-test reports · functional-test reports · integration-test reports · race-detector reports ·
coverage reports · benchmark results · static-analysis reports · security scans · dependency
scans · compatibility results · CI execution records · screenshots · execution logs · migration
logs · review reports · acceptance approvals · regression reports.

## Required evidence-record fields (mandate §10, verbatim list)

Every evidence record must identify:

- evidence ID;
- evidence type;
- story and task;
- acceptance criteria proven;
- execution command;
- code revision or commit SHA;
- branch or tag;
- execution environment;
- relevant tool versions;
- date and time;
- result;
- file or URI;
- checksum where appropriate;
- reviewer;
- superseded evidence where applicable.

All fields are mandatory except "checksum" (only where appropriate, e.g. binary/log artifacts)
and "superseded evidence" (only when this record supersedes a prior one). An evidence record
missing any other field is incomplete and must not be cited as proof of an acceptance criterion.

## Revision-pinning rule

Mandate §10, verbatim:

> Evidence that does not identify the tested revision must not be treated as final proof.

This programme applies that rule strictly:

- Every evidence record **must cite the exact commit SHA** it was captured against — never
  "current HEAD," "latest main," or any other moving-target reference. HEAD moves; a SHA does
  not.
- If the codebase advances past that pinned SHA **before the story is accepted**, the evidence is
  stale for acceptance purposes and must be handled one of two ways:
  1. **Re-validated** — re-run against the new HEAD and recorded as a new evidence record with
     status `retested`, referencing the superseded record's ID in "superseded evidence where
     applicable."
  2. **Explicitly carried forward** — if re-running is unnecessary because nothing material
     changed between the pinned SHA and current HEAD for the specific acceptance criterion in
     question, this must be stated as an explicit rationale note (what changed in between, why it
     does not affect this AC) attached to the evidence record or the story's `verification.md`.
     Silent carry-forward without a rationale note is not permitted.

## Failed-evidence preservation rule

Mandate §10, verbatim:

> Do not delete earlier failed verification merely because a later run passes.

Evidence records use exactly this status vocabulary (mandate §10):

| Status | Meaning |
|---|---|
| `failed` | Verification ran and did not meet the expected result. |
| `superseded` | Replaced by a later evidence record for the same AC (link recorded both ways). |
| `retested` | A new run against a newer revision, following a `failed` or stale prior record. |
| `resolved` | The underlying issue the `failed` record exposed has since been fixed and re-proven. |
| `accepted exception` | The failure is known and intentionally accepted as a residual risk rather than fixed (requires acceptance-authority sign-off, recorded alongside). |

A failed run stays in the evidence record with `failed` status permanently; it is never deleted
or overwritten. The passing re-run is a separate record referencing the failed one via
"superseded evidence where applicable," so the full history — including what broke and when it
was fixed — remains reconstructable.

## Evidence-bundle convention (precedent this policy generalizes)

`docs/implementation/evidence/README.md` establishes the existing bundle pattern for this repo:
each bundle directory contains `proof-bundle.md` (decisions, discussions, implementation
inventory, acceptance checklist status), `review-findings.md` (finding/severity/file:line/
resolution, no finding silently dropped), `command-log.md` (exact commands, exit codes,
summarized output — including commands that could not run, with reason and residual risk), and
`acceptance-map.md` (acceptance criteria → code/test/command evidence).

This policy generalizes that same four-file bundle shape into each story's `evidence/` directory
under `impl/waves/.../stories/story-NNN-.../evidence/`, scoped per acceptance criterion rather
than per phase, and governed by the field list and status vocabulary above. The
`docs/implementation/evidence/README.md` rule that applies unchanged here: "Reviewed/tested/
verified" claims without a corresponding evidence record are treated as not done.

## Where evidence lives

`evidence/index.md` at story creation time declares the categories of evidence this story expects
to produce (per `story.md` "required evidence"). Category subdirectories
(`baselines/tests/coverage/logs/screenshots/benchmarks/security/static-analysis/compatibility/
regression/reviews/acceptance`) are created only on first real content — see `naming-conventions.md`
Adaptation 2 for the rationale.
