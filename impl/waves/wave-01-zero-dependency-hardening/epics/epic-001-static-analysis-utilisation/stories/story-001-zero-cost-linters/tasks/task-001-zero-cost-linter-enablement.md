---
id: W01-E01-S001-T001
type: task
title: Zero-cost linter enablement
status: done
parent_story: W01-E01-S001
owner: W01Lint
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on: []
acceptance_criteria:
  - AC-W01-E01-S001-01
artifacts:
  - ART-W01-E01-S001-001
evidence:
  - EV-W01-E01-S001-001
---

# W01-E01-S001-T001 — Zero-cost linter enablement

## Task Definition

### Task objective

Enable `sqlclosecheck`, `rowserrcheck`, `bodyclose`, `wastedassign`, `makezero`, `musttag`, and
`testifylint` in `.golangci.yml`, and prove — via a fail-first fresh re-run, not by trusting the cited
MATRIX CS-23 snapshot — that the module tree is clean against all seven.

### Parent story

W01-E01-S001 — Enable the zero-cost leak-detection linter set.

### Owner

unassigned

### Status

todo

### Dependencies

None.

### Detailed work

1. Read `.golangci.yml` to confirm the current `enable:`/`disable:` state of all seven analyzers
   (resolving `plan.md`'s "Unresolved questions" item on this point).
2. Run `golangci-lint run --enable=sqlclosecheck,rowserrcheck,bodyclose,wastedassign,makezero,
   musttag,testifylint ./...` against the current HEAD, before editing `.golangci.yml` — this is the
   fail-first evidence step (confirms or corrects the cited "26 sites clean, zero violations" claim).
3. If the fresh run confirms zero hits: edit `.golangci.yml` to permanently enable all seven
   analyzers.
4. If the fresh run surfaces a new hit not covered by the cited snapshot: record it in
   `deviations.md`, fix it (small, in-scope fix) or escalate per RISK-W01-E01-002's contingency (a
   5th task) if the fix is non-trivial, then proceed.
5. Run `golangci-lint run ./...` against the full module tree with the updated `.golangci.yml` to
   confirm exit code 0.

### Expected files or components affected

`.golangci.yml`.

### Expected output

An updated `.golangci.yml` with all seven zero-cost analyzers enabled, and two run logs (fail-first
pre-enablement forced-run, and post-enablement confirmation run) both showing the expected result.

### Required artifacts

ART-W01-E01-S001-001 (updated `.golangci.yml`).

### Required evidence

EV-W01-E01-S001-001 (static-analysis report, fail-first + confirmation pair).

### Related acceptance criteria

AC-W01-E01-S001-01.

### Completion criteria

`golangci-lint run` with all seven analyzers enabled exits 0 across the full module tree, evidenced
by a logged run against a named commit SHA.

### Verification method

Direct command execution (`golangci-lint run`), logged output retained as evidence per
`evidence/index.md`.

### Risks

RISK-W01-E01-002 (fresh re-run surfaces an unrecorded hit) — see epic-level `risks.md`.

### Rollback or recovery considerations

Revert `.golangci.yml` if enablement produces an unexpected volume of new hits that cannot be
resolved within this task's bounded scope; escalate per the risk's contingency rather than silently
disabling an analyzer again without recording why.

## Implementation Record

Implemented 2026-07-13 by W01Lint (working diff on HEAD `0a31186cada5c275a588c74081cf977adf346e61`; conductor owns commits).

Enabled sqlclosecheck/rowserrcheck/bodyclose/wastedassign/makezero/musttag/testifylint in `.golangci.yml`. Fresh triage at HEAD confirmed 0 hits for all seven; one Phase-2 `musttag` hit in sibling-new `internal/cli/init_version.go` fixed (json tag). Files: `.golangci.yml`, `internal/cli/init_version.go`.

## Verification Record

AC-W01-E01-S001-01: per-linter `--enable-only` runs and the full-tree run all exit 0 (EV-001; `evidence/static-analysis/per-linter-enablement-pass-after.txt`). **pass**

### Final conclusion

Zero-cost set enabled and pinned at zero hits.

## Deviations Record

None — see story-level `deviations.md` for story-wide drift records.
