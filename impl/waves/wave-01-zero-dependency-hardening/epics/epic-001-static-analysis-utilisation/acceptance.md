---
id: W01-E01-ACCEPTANCE
type: epic-acceptance
epic: W01-E01
wave: W01
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W01-E01 — Epic-level acceptance

Reproduces and elaborates `epic.md`'s "Epic acceptance criteria" section as a standalone,
independently-referenceable acceptance record, consistent with the wave-level `../../acceptance.md`
pattern (AC-W01-01 through AC-W01-03 there map onto this epic).

## AC-W01-E01-01 — Zero-cost linter set enabled

`sqlclosecheck`, `rowserrcheck`, `bodyclose`, `wastedassign`, `makezero`, `musttag`, `testifylint`
are enabled in `.golangci.yml`; a full-module-tree `golangci-lint run` against them exits 0.
`noctx`'s 2 named prod hits (`internal/cli/config_delegate.go:34`, `internal/cli/lint_cmd.go:129`)
and `copyloopvar`'s 1 named prod hit (`app/maintenance.go:148`) are fixed. Traces to
W01-E01-S001 and its acceptance criteria AC-W01-E01-S001-01 through -04.

## AC-W01-E01-02 — Judged linter set enabled and triaged

`gosec` (with the named triage list: G704 JWKS annotation, G115 conversion set reviewed and
annotated-or-bounded per site, G304 buildinfo annotation), `errorlint` (mechanical `errors.Is` fix
at `kernel/httpx/middleware.go:54`), `exhaustive` (fail-closed default arms annotated at
`kernel/workflow/definition.go:313` and `runtime.go:170`), `forcetypeassert` (checked assertions at
`kernel/auth/jwks.go:112` and `kernel/config/bind.go:150`), `usestdlibvars` (mechanical fixes as
found) are enabled in `.golangci.yml` with every hit resolved (fixed or annotated with an inline
justification comment traceable to this epic's triage record). `wrapcheck`/`revive` are explicitly
recorded as a rejected recommendation (REJ classification, disposition rejected, with rationale) —
not enabled. `kernel/policy/policy.go:166`'s `nilerr` hit is recorded as an explicit non-finding
(deliberate fail-closed design, annotated not fixed). Traces to W01-E01-S002.

## AC-W01-E01-03 — Supply-chain and hook hygiene

`go mod verify` runs as a step in `ci.yml`. A license-scanning signal is enabled (Trivy license
scanner or `go-licenses`, with the choice documented and the alternative's rejection rationale
recorded). The nightly fuzz-schedule wiring is confirmed to exist (per session-delta SD-02 / PR #24)
and correctly invokes seed-corpus replay; the remaining coverage-guided `-fuzz=` gap is explicitly
recorded as out of scope (W07 / REL-04 T8 / PERF-06 T3-T4), not silently closed nor silently
duplicated. The pre-push hook no longer silently skips DB tests without `WOWAPI_REQUIRE_DB` set — it
either requires the variable or fails loudly if the DB is unavailable. Traces to W01-E01-S003.

## AC-W01-E01-04 — Independent review passed

All three stories (S001, S002, S003) have passed independent review per mandate §14. S002's review
specifically confirms no gosec/errorlint/exhaustive/forcetypeassert hit was silently dropped from the
triage record (every hit has a fix commit or an annotation with justification, traceable in
`evidence/index.md`). S003's review specifically confirms the nightly-fuzz scope boundary is stated
honestly — neither silently closed as "done" nor silently duplicated against W07's REL-04 T8 work.

## Acceptance authority

Framework architecture lead (role-based per `../../wave.md`'s split — the static-analysis epic sits
on the "ARCH-adjacent linter/observability/HTTP work" side of the wave's owner split, not the
DX-04/generator side).

## Acceptance record — 2026-07-13

Satisfied 2026-07-13. All acceptance criteria for W01-E01 are met; independent review passed
(W01ReviewGate); accepted by conductor.
