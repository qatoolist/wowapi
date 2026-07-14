---
id: IMPL-W01-E01-S002
type: implementation-record
parent_story: W01-E01-S002
status: complete
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Implementation record — W01-E01-S002

Implemented 2026-07-13 by W01Lint at HEAD `0a31186cada5c275a588c74081cf977adf346e61` (+ the W01 wave
working diff; conductor owns the commit). `.golangci.yml` now enables `gosec`, `errorlint`,
`exhaustive`, `forcetypeassert`, `usestdlibvars`. Fresh-run discipline: hits were enumerated twice —
at clean HEAD (Phase 1) and again at the Phase-2 enablement state with all sibling wave work in-tree
— and every hit at the enablement state has a disposition below. Config judgment recorded: gosec,
forcetypeassert, and errorlint are excluded for `_test.go` files (documented inline in the config:
test fixtures trip G404/G301/G306-style rules by design; a panicking assertion in a test IS the
failure signal; tests legitimately assert exact sentinel/type identity). Non-test code — including
the shipped `testkit` package — stays fully covered.

## Fresh-run drift vs the cited snapshot (recorded, per plan)

- gosec: cited "38 hits" → fresh runs report 111 total / 26 non-test at HEAD, 27 non-test at the
  Phase-2 state (new sibling file). The G115 class enumerated to 7 exact sites (below).
- errorlint: cited 1 site → 3 non-test at HEAD (+2 in post-#25 benchbudget code, per the W00
  baseline note) + 1 more in sibling-new `init_version.go` at Phase 2.
- exhaustive: cited 2 sites → 4 (2 new `reflect.Kind` switches in kernel/config).
- forcetypeassert: cited 2 sites → 3 (+ `kernel/httpclient/client.go:71`).
- usestdlibvars: no cited sites → 5 test-file hits at HEAD, 9 at Phase 2 (sibling test edits).
- wrapcheck/revive: cited "~50 each" → 464 / 231 at HEAD. Direction unchanged (noise-dominant);
  rejection stands a fortiori.

## Per-hit triage record (every non-test hit at the enablement state)

### gosec (27 non-test hits)

| Rule | Site | Disposition |
|---|---|---|
| G704 | `kernel/auth/jwks.go:204` (now :214) | **annotated** `#nosec G704` referencing SEC-06 (D-07): trusted-issuer JWKS/discovery URI from boot config constrained by `validateHTTPSURL`; governed pattern, not fixed |
| G704 | `kernel/auth/jwks.go:210` (now :220) | **annotated** — same SEC-06 justification |
| G115 | `kernel/audit/audit.go:164` | **annotated**: bijective int64→uint64 reinterpretation for hash-input encoding; chain seq never negative |
| G115 | `kernel/audit/audit.go:176` | **annotated**: length prefix over single Postgres row fields, bounded far below 4 GiB |
| G115 | `kernel/database/database.go:135` | **annotated**: bounded by prior validation — `config.Validate` rejects MaxConns outside [2,200] |
| G115 | `kernel/jobs/jobs.go:105` | **annotated**: splitmix64 seed reinterpretation (distribution-only); modulo result < span (positive duration) |
| G115 | `kernel/mfa/totp.go:116` | **annotated**: RFC 6238 counter; `t.Unix()` non-negative for any real clock, Step validated positive |
| G115 | `kernel/pagination/cursor.go:202` | **FIXED**: explicit bounds check added — uint > MaxInt64 now fails closed (`cursor value overflows int64`) instead of silently wrapping; no prior validation existed. Regression test `TestEncodeCursorUnsignedOverflow` (fails at HEAD, passes after) |
| G115 | `kernel/pagination/cursor.go:210` | **FIXED**: same bounds check for uint64 |
| G304 | `internal/buildinfo/buildinfo.go:70` | **annotated**: build-time diagnostic tooling reading the go.mod it discovered by walking up from its own cwd (the named "buildinfo file read" hit) |
| G304 | `internal/cli/openapi_cmd.go:128` | **annotated**: CLI reads the fragment paths its caller passed — the command's purpose |
| G304+G703 | `internal/tools/benchbudget/main.go:92` | **annotated** (G703 appeared once G304 was suppressed — both taint rules over the same site): tool reads the budgets file named on its own command line |
| G304 | `kernel/config/tree.go:19` (now :22) | **annotated**: boot-time config loader reads operator-configured YAML paths by design |
| G204 | `internal/cli/config_delegate.go:34` | **fixed+annotated**: `exec.CommandContext` (S001 scope) + `#nosec G204` — runs the repo's own `go run ./tools/configcheck` with the CLI caller's args |
| G204 | `internal/cli/lint_cmd.go:129` | **fixed+annotated**: `exec.CommandContext` + `#nosec G204` — fixed `go list` argv |
| G204 | `internal/cli/init_version.go:100` (sibling-new, W01Gen) | **fixed+annotated**: `exec.CommandContext` + `#nosec G204` — fixed `go list -m -json` argv, module@query is the caller's requested version |
| G306 | `internal/cli/init_version.go:97` (sibling-new) | **FIXED**: throwaway module stub written 0o600 (nothing needs wider perms) |
| G301 | `internal/cli/migrate_cmd.go:72` | **annotated**: migrations directory is project source, world-readable by design |
| G306 | `internal/cli/migrate_cmd.go:83` | **annotated**: generated migration stub is project source |
| G306 | `internal/cli/deploy_cmd.go:95` | **annotated** (+comment noting the manifest contains secretref names, not secrets) |
| G306 | `internal/cli/openapi_cmd.go:101` | **annotated**: merged OpenAPI spec is a world-readable build artifact |
| G301 | `internal/cli/scaffold.go:47,75` | **annotated**: scaffolded project directories are user source trees |
| G306 | `internal/cli/scaffold.go:64,78` | **annotated**: scaffolded files are the user's project source |
| G101 | `testkit/db.go:376,381` | **annotated**: local-test-only role passwords set via ALTER ROLE inside throwaway test databases; never committed production credentials |

**G120 (`kernel/httpx/csrf.go:118`, FBL-09)**: cross-reference only per story scope — the definitive
Phase-2 gosec re-run (routed to this story by W01Http) confirms the hit NO LONGER EXISTS: W01Http's
W01-E03 http-hardening work fixed the unbounded form parse before this story's enablement run.
Recorded as fixed-by-W01-E03, not by this story (see `deviations.md`).

The remaining 84–85 test-file gosec hits are covered by the documented `_test.go` exclusion, not
per-site annotations.

### errorlint (4 non-test)

- `kernel/httpx/middleware.go:54` — **fixed** (named site): `rec == http.ErrAbortHandler` →
  `err, ok := rec.(error); ok && errors.Is(err, http.ErrAbortHandler)` (recover() yields `any`, so
  the errors.Is adoption required the error-type guard; non-error panic values keep falling through
  to logging — behavior-preserving, wrapped-sentinel-tolerant).
- `internal/tools/benchbudget/main.go:114,118` — **fixed**: `%v` → `%w` on the wrapped ParseInt errors.
- `internal/cli/init_version.go:162` (sibling-new) — **fixed**: `%v` → `%w`.

### exhaustive (4)

All four are fail-closed switches, **annotated** with `//exhaustive:ignore` plus an explanatory
comment preserving and documenting the fail-closed design (per AC-04, NOT converted to
enumerations): `kernel/workflow/definition.go:313` (unknown/terminal step types have no outgoing
transitions), `kernel/workflow/runtime.go:170` (default: arm rejects unknown decisions with
invalid_decision — deny-by-default), `kernel/config/bind.go:326` and `kernel/config/schema.go:95`
(both fall through to a fail-closed return; drift sites recorded).

### forcetypeassert (3 non-test)

- `kernel/auth/jwks.go:112` — **fixed**: comma-ok on `http.DefaultTransport.(*http.Transport)` with
  a documented loud-panic false path (impossible under the stdlib contract; failing loudly at boot
  beats proceeding with an untamed transport).
- `kernel/httpclient/client.go:71` (drift site) — **fixed**: same pattern, consistent with the
  constructor's documented panic-on-bad-config behavior.
- `kernel/config/bind.go:150` — **fixed**: comma-ok with `b.errf` fail-closed binder error on the
  (unreachable) false path.

### usestdlibvars (9, all test files)

All **fixed mechanically**: `"GET"`→`http.MethodGet`, `"PUT"`→`http.MethodPut`,
`200`→`http.StatusOK` in kernel/document, kernel/httpx (4 sites incl. sibling-edited tests),
kernel/storage tests (+ the `net/http` imports those files then needed).

### nilerr non-finding (AC-07)

`kernel/policy/policy.go:166` (now :172) — **annotated, not fixed**: inline comment documenting the
deliberate fail-closed design (unparseable RUNTIME value → condition false/deny; malformed POLICY
errors handled loudly at :161) plus a `//nolint:nilerr` marker so the adjudication survives any
future nilerr enablement. The condition logic is unchanged. nilerr itself remains NOT enabled
(story-002's "already-enabled" framing was drift — it was never in the committed config; recorded in
`deviations.md`).

### wrapcheck / revive (AC-07)

**REJECTED** (classification REJ, disposition rejected, mandate §1.3/§1.4): 464 / 231 hits at the
fresh HEAD triage — noise-dominant without a per-project tuning investment disproportionate to the
benefit; staticcheck+errorlint cover the real classes. Neither enabled; absence proven in
`evidence/static-analysis/wrapcheck-revive-absence.txt`.

## Files changed

`.golangci.yml` (shared with S001); kernel: auth/jwks.go, httpclient/client.go, config/{bind,schema,
tree}.go, workflow/{definition,runtime}.go, policy/policy.go, audit/audit.go, database/database.go,
jobs/jobs.go, mfa/totp.go, pagination/cursor.go (+ pagination_test.go); httpx/middleware.go (+ 3
test files for usestdlibvars); document/service_test.go, storage/memory_test.go; internal:
buildinfo/buildinfo.go, cli/{config_delegate,lint_cmd,init_version,openapi_cmd,deploy_cmd,
migrate_cmd,scaffold}.go, tools/benchbudget/main.go; testkit/db.go.

## Tests added or modified

`kernel/pagination/pagination_test.go` `TestEncodeCursorUnsignedOverflow` — defends the new G115
bounds check (fail-first proven at HEAD).

## Commits

Conductor owns commits; delivered as the W01 wave working diff on `0a31186`.

## Known limitations

The `#nosec` annotations are accepted-risk records, not eliminated risk (per the story's
residual-risk expectations): G704/G304/G301/G306/G101/G115-bounded remain deliberate, reviewed
acceptances of linter-flagged patterns.

## Relationship to the approved plan

Matched `plan.md`; drift (counts, new sites, G120 already fixed, nilerr never-enabled) recorded in
`deviations.md`, not silently reconciled.
