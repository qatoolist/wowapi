---
id: W01-E03-S001-T001
type: task
title: Scaffold-template four-timeout config keys + safe defaults + template-render test
status: done
parent_story: W01-E03-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on: []
acceptance_criteria:
  - AC-W01-E03-S001-01
  - AC-W01-E03-S001-02
artifacts: []
evidence: []
---

# W01-E03-S001-T001 — Scaffold-template four-timeout config keys + safe defaults + template-render test

## Task Definition

### Task objective

Add `ReadTimeout`, `WriteTimeout`, `IdleTimeout`, `HeaderTimeout` fields to `kernel/config.HTTP` with
MATRIX CS-09's safe defaults (30s/60s/120s/10s), wire all four into the scaffold template's
`http.Server{}` literal, and add a fail-first template-render test proving the gap and the fix.

### Parent story

W01-E03-S001 — Server timeouts and body bounds.

### Owner

Unassigned.

### Status

todo.

### Dependencies

None.

### Detailed work

1. Resolve `plan.md` unresolved question 2 (`HeaderTimeout` vs. `ReadHeaderTimeout` naming) before
   writing the field — confirm with review whether this is a new key or a default-value update to the
   existing `ReadHeaderTimeout` field.
2. Add the resolved field(s) to `kernel/config.HTTP` (`kernel/config/config.go`, near line 104-114),
   following the existing `conf:`/`json:`/`doc:` tag conventions visible on the sibling fields.
3. Add corresponding default assignments to `Defaults()` (near line 162-168).
4. Locate (or confirm absence of) the DX-01 T5 shared scaffold-test-harness primitive referenced at
   wave level before building a new template-render test harness.
5. Wire all four timeouts into the `http.Server{}` literal in
   `internal/cli/templates/init/cmd_api_main.go.tmpl` (near line 314-317).
6. Check `configs_base.yaml.tmpl`/`configs_local.yaml.tmpl` for whether HTTP.* keys are enumerated
   there; add the new keys if the existing pattern requires it.
7. Write the template-render test: first confirm it fails against the pre-fix template (evidence of
   the real gap, mandate §13 fail-first), then confirm it passes post-fix.

### Expected files or components affected

- `kernel/config/config.go`.
- `internal/cli/templates/init/cmd_api_main.go.tmpl`.
- Possibly `internal/cli/templates/init/configs_base.yaml.tmpl` / `configs_local.yaml.tmpl`.
- A template-render test file (location depends on whether an existing scaffold-test harness is
  reused — see `plan.md` unresolved question 3).

### Expected output

`kernel/config.HTTP` carries the resolved timeout field(s) with correct defaults; the scaffold
template's generated `http.Server{}` literal sets all four `net/http.Server` timeout fields from
`cfg.HTTP.*`; a passing template-render test exists, with its pre-fix failing run captured as
fail-first evidence.

### Required artifacts

Scaffold-template diff; config schema addition. See `../../artifacts/index.md`.

### Required evidence

Template-render assertion (fail-first pair: failing run + passing run). See
`../../evidence/index.md`.

### Related acceptance criteria

AC-W01-E03-S001-01, AC-W01-E03-S001-02.

### Completion criteria

The config fields exist with correct defaults; the template renders all four timeouts; the
template-render test's fail-first pair is captured; `plan.md` unresolved questions 2, 3, and 4 are
resolved and the resolution is recorded in this task's Implementation Record.

### Verification method

Run the template-render test locally and in CI; run `kernel/config` unit tests confirming the
defaults; diff the rendered template output against the expected `http.Server{}` literal shape.

### Risks

RISK-W01-E03-001 (naming collision with the existing `ReadHeaderTimeout` key) — directly addressed by
resolving unresolved question 2 before implementation, not after.

### Rollback or recovery considerations

Revert the config-struct and template commit; no running-system state is implicated (generation-time
and config-load-time only).

## Implementation Record

Implemented 2026-07-13 by W01Http at SHA 0a31186cada5c275a588c74081cf977adf346e61 (working tree; conductor owns the wave commit).

- `kernel/config.HTTP`: added `ReadTimeout` (30s) / `WriteTimeout` (60s) / `IdleTimeout` (120s);
  bumped existing `ReadHeaderTimeout` default 5s → 10s (plan Q2 resolved as option (a) — no new
  `HeaderTimeout` key; see story deviations.md DEV-001). `Defaults()` updated.
- `internal/cli/templates/init/cmd_api_main.go.tmpl`: `http.Server{}` literal sets all four
  timeouts from `cfg.HTTP.*` with rationale comment.
- `internal/cli/templates/init/configs_base.yaml.tmpl`: four timeout keys enumerated with
  defaults (plan Q4 resolved: base yaml does enumerate http keys; local yaml does not — untouched).
- Plan Q3 resolved: reused the existing `internal/cli/scaffold_test.go` harness (callInit +
  assertFileMatches); no parallel harness.
- Tests: `TestInitAPIMainConfiguresAllServerTimeouts`, `TestInitConfigsBaseDocumentsServerTimeouts`
  (fail-first pair captured), `TestHTTPTimeoutDefaultsMatchCS09`; `load_test.go` 5s→10s follow-through.
- Docs: `docs/user-guide/configuration.md` http table + example updated.

### Commits / pull requests

None yet — conductor owns the wave commit; working-tree diff recorded in the story's
implementation.md.

## Verification Record

| Acceptance criterion | Verification method | Environment | Result | Evidence | Reviewer |
|---|---|---|---|---|---|
| AC-W01-E03-S001-01 | `TestHTTPTimeoutDefaultsMatchCS09` | Local | PASS | EV-W01-E03-S001-002 | pending |
| AC-W01-E03-S001-02 | Template-render fail-first pair | Local | PASS (failed pre-fix, passes post-fix) | EV-W01-E03-S001-001 | pending |

Execution date 2026-07-13; revision 0a31186cada5c275a588c74081cf977adf346e61; environment local darwin/arm64 (go1.26.5);
reviewer pending (W01 wave review gate). Findings: none open. Retest: not required.
Final conclusion: task complete, ACs verified.

## Deviations Record

Story-level DEV-W01-E03-S001-001 (HeaderTimeout naming) originates in this task — see ../deviations.md.
