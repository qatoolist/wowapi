---
id: IMPL-W01-E01-S003
type: implementation-record
parent_story: W01-E01-S003
status: complete
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Implementation record — W01-E01-S003

Implemented 2026-07-13 by W01Lint at HEAD `0a31186cada5c275a588c74081cf977adf346e61`. Note on
revision pinning: the conductor owns commits for this wave, so the changes below exist as a working
change on top of that SHA (`git diff --stat`: `.githooks/pre-push | 26 ++`, `.github/workflows/ci.yml
| 2 +`, `.github/workflows/security-scan.yml | 9 ++` — 33 insertions, 4 deletions); the evidence
records cite the HEAD SHA plus this diff-stat per the wave's evidence convention.

## What was actually implemented

1. **T001 — `go mod verify` CI step**: added as a distinct step in `ci.yml`'s `unit` job, directly
   after the `go.mod tidy check` step (same job that already owns module-hygiene checks). A failing
   step fails the `unit` job — no `continue-on-error` anywhere in the workflow.
2. **T002 — license-scanning signal**: the planned choice (Trivy `license` scanner) was carried
   through **unchanged** — `license` appended to the `trivy` job's `scanners:` list in
   `security-scan.yml`, with the step comment documenting the signal-not-gate posture. Choice
   validated before committing to it: a local Trivy run against a pristine HEAD copy enumerated all
   70 Go dependency licenses (MIT/Apache-2.0/BSD — all LOW severity), confirming the scanner is not
   hollow for `gomod` targets; the job's existing `severity: CRITICAL,HIGH` filter means the job log
   reports exactly the forbidden/restricted classes (currently zero) — the desired signal shape.
   Deliberately a **signal, not a gate** (job keeps `exit-code: "0"`, matching its other scanners and
   `story.md`'s residual-risk boundary); `dependency-review`'s `license-check: true` continues to
   cover `pull_request` events, Trivy now covers pushes and the weekly schedule.
3. **T003 — pre-push hook DB-silent-skip fix**: `.githooks/pre-push`'s bare `go test ./...` (which
   let DB-gated tests self-skip while the hook printed `pre-push: OK`) replaced with a required-DB
   invocation: `DATABASE_URL="${DATABASE_URL:-<compose default, mirrors Makefile TEST_DSN>}"
   WOWAPI_REQUIRE_DB=1 go test ./...`. `WOWAPI_REQUIRE_DB` is the repository's established convention
   (`testkit.RequireDB`, `testkit/db.go:161-171`; Makefile `test-integration`/`bench`/`coverage`/
   `ci-container-*`) — reused, not invented. On failure the hook prints an actionable block (start
   `make up` / export `DATABASE_URL` / explicit skip / `--no-verify`) and exits 1. An **explicit,
   loudly-announced opt-out** (`WOWAPI_PREPUSH_SKIP_DB=1`) preserves the hook's lighter-than-CI
   design for machines with no local DB — permitted by the task's own step-3 language ("how to
   explicitly opt out"); the *silent* skip is gone. `.githooks/pre-commit` untouched (verified:
   `git diff --quiet .githooks/pre-commit`).
4. **T004 — nightly fuzz-schedule confirmation**: no code change needed (wiring is correct as
   found). Audit note at `artifacts/nightly-fuzz-confirmation.md`: cron `17 3 * * *` →
   `changes` fail-safe (`code=true`) → `gate` test leg → `make ci-container-test` →
   `go test ./kernel/filtering/ ./kernel/pagination/ -run "^Fuzz" -count=1` (seed replay, no
   `-fuzz=`), plus an **observed** scheduled run (29229288699, 2026-07-13, success) whose log shows
   the seed-replay invocation executing. `-fuzz=` coverage-guided gap explicitly restated as W07
   scope (REL-04 T8 / PERF-06 T3/T4).

## Components changed

CI workflows (`ci.yml` unit job; `security-scan.yml` trivy job); git pre-push hook. No Go packages.

## Files changed

- `.github/workflows/ci.yml` (+2)
- `.github/workflows/security-scan.yml` (+7/−2, step rename + comment + scanner list)
- `.githooks/pre-push` (+24/−2)

## Interfaces introduced or changed

None.

## Configuration changes

`WOWAPI_PREPUSH_SKIP_DB` introduced as the pre-push hook's explicit DB-skip opt-out (hook-local, not
read by any Go code). `WOWAPI_REQUIRE_DB` reused unchanged.

## Schema or migration changes

None.

## Security changes

`go mod verify` (module-cache-vs-go.sum integrity) now runs on every CI `unit` job execution;
dependency-license detection now runs on every push/schedule via Trivy. Both are read-only checks.

## Observability changes

None.

## Tests added or modified

None (per plan — hook verified by direct execution, both CI steps by direct execution; see
`verification.md`).

## Commits

Conductor owns commits for this wave; working change delivered on top of
`0a31186cada5c275a588c74081cf977adf346e61`.

## Pull requests

None (conductor owns the wave-level integration).

## Implementation dates

2026-07-13.

## Technical debt introduced

None.

## Known limitations

- The license signal is detection-only (job `exit-code: "0"`); converting it into a blocking gate is
  a future policy decision, per `story.md` "Residual-risk expectations".
- The hook's required-DB posture covers DB-gated tests only; S3-gated tests (`WOWAPI_REQUIRE_S3`)
  can still self-skip locally — outside FBL-07's named scope, recorded as a follow-up candidate.
- The two CI-step additions are proven by local execution of the identical commands plus a clean
  `actionlint` pass; an in-CI run log will exist only after the conductor pushes the wave commit
  (recorded as an explicit evidence limitation, not silently claimed).

## Follow-up items

- Document the new hook behavior in `docs/user-guide/build-deploy.md` at wave documentation pass
  (routed to W01Docs; see `deviations.md` note if unrouted at closure).
- Optional: mirror the required-DB posture for S3-gated tests.

## Relationship to the approved plan

Matched `plan.md` on all four tasks. The "License signal decision" (Trivy) and the "Task breakdown
grouping decision" (T004 separate) were both carried through unchanged. No plan text was edited.
