---
id: IMPL-W01-E03-S001
type: implementation-record
parent_story: W01-E03-S001
status: complete
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Implementation record — W01-E03-S001

Implemented 2026-07-13 by W01Http against HEAD 0a31186cada5c275a588c74081cf977adf346e61 (working tree; conductor owns the wave
commit). `git diff --stat` for this story's files:

```
kernel/config/config.go                            | 35 ++++++++--
kernel/config/load_test.go                         |  2 +-
kernel/config/unsafe_config_matrix_test.go         | 61 ++++++++++++++
kernel/httpx/csrf.go                               | 22 ++++++
kernel/httpx/csrf_test.go                          | 67 ++++++++++++++++
internal/cli/templates/init/cmd_api_main.go.tmpl   |  9 +++
internal/cli/templates/init/configs_base.yaml.tmpl |  9 ++-
internal/cli/scaffold_test.go                      | (2 new tests in the shared file)
docs/user-guide/configuration.md                   | (http key table + example updated)
```

## What was actually implemented

1. **Config schema (T001)** — `kernel/config.HTTP` gained `ReadTimeout` (30s), `WriteTimeout`
   (60s), `IdleTimeout` (120s); the EXISTING `ReadHeaderTimeout` default was bumped 5s → 10s.
   `Defaults()` updated to match. Resolution of plan.md unresolved question 2 (recorded here per
   the plan's approval conditions): **option (a)** — MATRIX CS-09's "header 10s" is delivered on
   the existing `ReadHeaderTimeout` key, NOT a new `HeaderTimeout` key. Rationale: Go's
   `http.Server` has exactly one header-read timeout field (`ReadHeaderTimeout`); a second,
   same-purpose config key would have nothing distinct to wire to and is precisely the
   RISK-W01-E03-001 confusion. Deviation from AC-01's literal wording recorded as DEV-001 in
   `deviations.md`.
2. **Prod-profile zero rejection (T002)** — `Framework.Validate()`'s existing prod block
   (`f.Environment.IsProd()`) gained `<= 0` rejections for the three new keys. Resolution of
   plan.md unresolved question 1: **prod-profile-only**, matching the task brief, the story AC,
   and the SSRF-disable precedent; `TestConnectionTimeoutZeroToleratedOutsideProd` pins the
   distinction against the unconditional pattern (an explicit zero stays a legal local/dev
   convenience — "0 = unlimited" is meaningful in dev, unlike the pre-existing three keys which
   gate mechanisms that must always exist). `ReadHeaderTimeout` remains under its stronger
   pre-existing unconditional `> 0` rule (unchanged).
3. **Scaffold template (T001)** — `cmd_api_main.go.tmpl`'s `http.Server{}` literal now sets all
   four timeouts from `cfg.HTTP.*` with a rationale comment; `configs_base.yaml.tmpl` enumerates
   the four keys with defaults and a comment block. Resolution of plan.md unresolved questions 3
   and 4: the existing `internal/cli/scaffold_test.go` harness (`callInit` +
   `assertFileMatches`/`assertFileContains`) was reused — no parallel harness built; the base
   config template DOES enumerate `http.*` keys, so the new keys were added there
   (`configs_local.yaml.tmpl` carries no http keys — untouched).
4. **CSRF defensive bound (T003)** — `CSRFProtect`'s form-field fallback wraps `r.Body` in
   `http.MaxBytesReader(w, r.Body, limit)` before `r.FormValue`; `CSRFPolicy` gained
   `MaxFormBytes int64` (0 → `csrfDefaultMaxFormBytes` = 1 MiB, matching the default
   `http.max_body_bytes`). Bound resolution per the task's sanctioned option: `HTTP.MaxBodyBytes`
   is not threaded through `SecurityChain(config.Security)`, so a policy-level knob with the
   1 MiB default constant was chosen over an additive SecurityChain signature change (see
   known limitations).

## Components changed

`kernel/config` (HTTP struct, Defaults, Validate), `kernel/httpx` (csrf.go),
`internal/cli/templates/init` (api main + base yaml templates), docs/user-guide.

## Interfaces introduced or changed

`config.HTTP`: 3 new fields (additive). `httpx.CSRFPolicy`: 1 new field (additive). No signature
changes; no existing caller affected.

## Configuration changes

`http.read_timeout`, `http.write_timeout`, `http.idle_timeout` (new, safe defaults);
`http.read_header_timeout` default 5s → 10s.

## Security changes

Closes the connection-level resource-exhaustion gap (Slowloris variants) for every future
`wowapi init` product; CSRF middleware now self-bounds its body read regardless of chain order.

## Tests added or modified

- `internal/cli/scaffold_test.go`: `TestInitAPIMainConfiguresAllServerTimeouts`,
  `TestInitConfigsBaseDocumentsServerTimeouts` (fail-first pair captured).
- `kernel/config/unsafe_config_matrix_test.go`: 3 prod-gated matrix rows,
  `TestHTTPTimeoutDefaultsMatchCS09`, `TestConnectionTimeoutZeroToleratedOutsideProd`
  (fail-first captured).
- `kernel/config/load_test.go`: default assertion 5s → 10s (consequence of the CS-09 bump).
- `kernel/httpx/csrf_test.go`: `TestCSRFOversizedFormBodyRejected` (fail-first captured),
  `TestCSRFCustomMaxFormBytesOverridesDefault`.

## Known limitations

- A product that raises `http.max_body_bytes` above 1 MiB and relies on the browser-profile CSRF
  FORM-FIELD fallback (not the header) for >1 MiB form posts must set `CSRFPolicy.MaxFormBytes`
  when constructing `CSRFProtect` directly; `SecurityChain` does not thread `HTTP.MaxBodyBytes`
  (kept out to avoid a signature change; candidate follow-up if a real product hits it).
- The wowsociety backport of the four template lines remains PROD-03 (out of scope, unchanged).

## Relationship to the approved plan

Matches plan.md with both flagged unresolved questions resolved as recorded above; the only
AC-wording divergence is DEV-001 (deviations.md).
