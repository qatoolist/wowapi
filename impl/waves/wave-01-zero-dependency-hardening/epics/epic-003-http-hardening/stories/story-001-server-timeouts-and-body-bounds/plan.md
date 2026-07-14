---
id: PLAN-W01-E03-S001
type: plan
parent_story: W01-E03-S001
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
---

# Plan — W01-E03-S001 — Server timeouts and body bounds

Per mandate §8.5: "Do not invent precise code changes where the repository does not yet provide
enough information. Clearly distinguish confirmed facts, planned changes, and implementation
assumptions." Confirmed facts below were verified directly against repository state on 2026-07-12;
they are marked `[CONFIRMED]`. Planned changes are marked `[PLANNED]`. Open implementation-time
judgment calls are marked `[ASSUMPTION]` and are also listed under "Unresolved questions."

## Proposed architecture

`[CONFIRMED]` The scaffold template (`internal/cli/templates/init/cmd_api_main.go.tmpl`) is rendered
by `wowapi init` into a product's `cmd/api/main.go`. It is not wowapi-internal runtime code — it is a
Go-template file containing Go source with `{{.Name}}`/`{{.Module}}` substitutions. The `http.Server{}`
literal it renders lives at lines 314-317 of the current template.

`[PLANNED]` Four new fields on `kernel/config.HTTP` (`ReadTimeout`, `WriteTimeout`, `IdleTimeout`,
`HeaderTimeout`), each wired into the rendered `http.Server{}` literal alongside the existing
`ReadHeaderTimeout` wiring. The architecture does not change — this is an additive extension of an
existing, already-correctly-shaped config struct and an already-correctly-shaped template literal.

## Implementation strategy

Three independent-but-sequenced steps, corresponding to this story's three tasks:

1. Add the four config keys with safe defaults and decide/implement their validation policy (T001,
   T002 combined at the config layer — see task breakdown for the T001/T002 split rationale).
2. Wire the new keys into the scaffold template's `http.Server{}` literal and add the fail-first
   template-render test (T001).
3. Fix the CSRF `FormValue` defensive-bound gap, independently of the timeout work (T003).

## Expected package or module changes

- `kernel/config` — `HTTP` struct gains 4 fields; `Defaults()` gains 4 default assignments;
  `Framework.Validate()` gains validation logic for the 4 new fields (see "Unresolved questions" for
  the unconditional-vs-prod-only policy choice).
- `internal/cli/templates/init` — `cmd_api_main.go.tmpl` gains 3 new struct-literal lines (the 4th,
  `ReadHeaderTimeout`, already exists); possibly `configs_base.yaml.tmpl`/`configs_local.yaml.tmpl` if
  example config values are added there (to be determined — see current-state note in `story.md`: not
  yet confirmed whether these example-config files enumerate every HTTP.* key or only a subset).
- `kernel/httpx` — `csrf.go`'s unsafe-method branch gains an `http.MaxBytesReader` wrap around the
  `r.FormValue` call site.

## Expected file changes where determinable

- `kernel/config/config.go` — `HTTP` struct (near line 104-114), `Defaults()` (near line 162-168),
  `Framework.Validate()` (near line 189-200).
- `internal/cli/templates/init/cmd_api_main.go.tmpl` — `http.Server{}` literal (near line 314-317).
- `kernel/httpx/csrf.go` — unsafe-method branch (near line 108-127, specifically the `FormValue` call
  at line 118).
- A new or extended template-render test file — exact location `[ASSUMPTION]`: DX-01 T5's scaffold
  test harness is referenced by `wave.md` as a shared primitive other W01 stories (DX-02's generator-
  output-boots test) build on; this story's template-render assertion should reuse that harness if it
  already provides a "render the init template set and inspect output" capability, rather than
  building a second, parallel harness. Whether that harness exists yet and in what form is not
  confirmed as of this plan's writing — T001 must locate it first.
- `kernel/config/config_test.go` or `unsafe_config_matrix_test.go` — the most direct existing
  precedent for "config value X causes Validate() to reject" tests is
  `kernel/config/unsafe_config_matrix_test.go` (confirmed to already contain table-driven mutate-and-
  expect-reject cases for `HTTP.ReadHeaderTimeout`, `HTTP.RequestTimeout`, `HTTP.MaxBodyBytes` at
  lines 106, 111, 116, 121). T002's new test cases should extend this existing table rather than
  create a new test file, following the codebase's existing pattern.

## Contracts and interfaces

No public interface changes. `config.HTTP` is a plain struct; adding fields is additive and does not
break any existing caller (Go's struct literal and field-access patterns tolerate new fields with
default zero values, and this story explicitly gives them safe non-zero defaults via `Defaults()`).

## Data structures

`[PLANNED]` `kernel/config.HTTP` gains:

```text
ReadTimeout   time.Duration `conf:"read_timeout"   default:"30s"  json:"read_timeout"   doc:"..."`
WriteTimeout  time.Duration `conf:"write_timeout"  default:"60s"  json:"write_timeout"  doc:"..."`
IdleTimeout   time.Duration `conf:"idle_timeout"   default:"120s" json:"idle_timeout"   doc:"..."`
HeaderTimeout time.Duration `conf:"header_timeout" default:"10s"  json:"header_timeout" doc:"..."`
```

`[ASSUMPTION]` The exact `conf:`/`json:` tag names above are illustrative, following the existing
`snake_case` convention visible in the same struct (`read_header_timeout`, `request_timeout`,
`max_body_bytes`) — the precise tag strings are an implementation-time detail, not invented as final
here per mandate §8.5. The `HeaderTimeout` name in particular needs implementation-time reconciliation
against the already-existing `ReadHeaderTimeout` field — see "Unresolved questions."

## APIs

No HTTP or gRPC API surface changes. This is server-configuration-only.

## Configuration changes

Four new `kernel/config.HTTP` keys as above, each with a safe non-zero default so unset config falls
through to the default rather than tripping any new rejection (this is the explicit RISK-W01-003
mitigation, already stated at wave level).

## Persistence changes

None.

## Migration strategy

None — no data or schema migration involved.

## Concurrency implications

None beyond what `http.Server`'s own documented concurrency model already provides — these are
passive connection-level timeout knobs, not new concurrent code paths.

## Error-handling strategy

A zero-value timeout in prod (or, per the unresolved validation-policy question, in any profile) is
surfaced through the existing `Framework.Validate()` joined-error-list mechanism
(`errors.Join(errs...)`, already used for every other validation failure in the same function) — no
new error-handling pattern is introduced.

## Security controls

Connection-level timeout enforcement (primary control of this story) and CSRF `FormValue` defensive
bounding (secondary, gosec G120 fix) — both described in `story.md` "Security considerations."

## Observability changes

None planned. A connection terminated by a new server-level timeout is visible through Go's standard
`net/http` server-level logging, unchanged by this story.

## Testing strategy

- **Unit**: `kernel/config` table-driven test cases (extending `unsafe_config_matrix_test.go`'s
  existing pattern) proving each of the 4 new keys defaults correctly and rejects an explicit zero
  value per the resolved validation policy.
- **Template-render / fail-first**: a test that renders the scaffold template's `cmd_api_main.go.tmpl`
  and asserts all four timeout fields appear in the `http.Server{}` literal, wired from `cfg.HTTP.*`.
  This test must be written to fail against the current template first (proving the gap is real, per
  mandate §13), then pass after the template fix.
- **CSRF defensive-bound**: a test proving `r.FormValue` in the unsafe-method CSRF branch no longer
  reads an unbounded body — e.g. a request with an oversized form body is rejected rather than fully
  buffered. Exact assertion mechanism (checking for a `MaxBytesError`-equivalent outcome, or a byte-
  count-bounded read) is an implementation-time detail.
- **Security/static-analysis**: a gosec re-run (once W01-E01-S002 enables it) confirming the
  `csrf.go:118` G120 hit is resolved. This story's own evidence should capture a scoped, ad hoc gosec
  run against `kernel/httpx/csrf.go` even before W01-E01-S002 lands, since S001 and S002 (W01-E01) may
  execute in parallel and this story should not depend on that epic's sequencing to prove its own AC.

## Regression strategy

The template-render test and config unit tests both run in the existing CI gate; no existing test is
expected to break since all four new keys are additive with safe defaults. The CSRF fix is a pure
defensive addition (bounding an already-present read) with no behavioral change for any request whose
form body is already within the framework's other body-size guardrails.

## Compatibility strategy

Additive-only. See `story.md` "Compatibility considerations."

## Rollout strategy

The fix ships in the scaffold template; it takes effect for any product generated by a `wowapi init`
run after this story lands. No rollout mechanism is needed beyond normal version-tagged release of
the wowapi module, since the template only affects future generation, not already-generated code.

## Rollback strategy

If the new timeout defaults prove too aggressive for a real deployment pattern discovered post-
release, the fix is reverted by reverting the template/config commit — no data or running-system
rollback is implicated, since this is generation-time and config-load-time behavior only.

## Implementation sequence

1. T001 — add the four config keys + wire into the scaffold template + fail-first template-render
   test (config-struct change and template change are tightly coupled — the template test cannot pass
   without the config fields existing, so they are one task).
2. T002 — prod-profile (or, per the resolved policy, unconditional) zero-timeout `config.Validate`
   rejection, extending the existing `unsafe_config_matrix_test.go` table.
3. T003 — CSRF `MaxBytesReader` defensive bound (independent of T001/T002; may run in parallel).

## Task breakdown

- **W01-E03-S001-T001** — scaffold-template four-timeout config keys + safe defaults + template-render
  test.
- **W01-E03-S001-T002** — prod-profile zero-timeout `config.Validate` rejection.
- **W01-E03-S001-T003** — CSRF `MaxBytesReader` defensive bound (gosec G120 fix).

## Expected artifacts

- Scaffold-template diff (`cmd_api_main.go.tmpl`).
- Config schema addition (`kernel/config.HTTP`).
- See `../artifacts/index.md`.

## Expected evidence

- Template-render assertion (fail-first log).
- Prod-profile zero-timeout rejection test output.
- gosec G120 resolution (scoped re-run).
- See `../evidence/index.md`.

## Unresolved questions

1. **Validation policy for the 4 new keys**: the task brief that produced this story specifies
   "`config.Validate` rejects ZERO values specifically in prod profile — same pattern as the existing
   SSRF-disable prod rejection at `kernel/config/config.go:261-263`." However, the three *already-
   existing* HTTP timeout keys (`ReadHeaderTimeout`, `RequestTimeout`, `MaxBodyBytes`) in the same
   `Framework.Validate()` function are rejected **unconditionally** (not prod-gated) at lines 192-200.
   Both patterns are real, existing precedents in the same file. This plan does not silently pick one
   — the implementer must confirm which pattern the reviewer/architecture-lead judges more consistent
   before T002 is implemented, and record the choice explicitly (in `implementation.md` if it matches
   the task brief's prod-only framing, or as a documented deviation in `deviations.md` if the
   unconditional pattern is chosen instead, since that would diverge from this story's own AC-
   W01-E03-S001-03 wording — see note on that AC if this occurs).
2. **`HeaderTimeout` vs. `ReadHeaderTimeout` naming**: the task brief specifies a new `HTTP.HeaderTimeout`
   config key with a 10s default, distinct from the already-existing `HTTP.ReadHeaderTimeout` field
   (also already defaulted to 5s and already wired into the scaffold template). Go's own
   `http.Server.ReadHeaderTimeout` field is the single field that governs header-read time — there is
   no separate `HeaderTimeout` field on `http.Server` itself. This plan flags that introducing a
   second, differently-named-but-same-purpose config key risks confusion (RISK-W01-E03-001) and that
   T001's implementation must resolve, at build time, whether: (a) MATRIX CS-09's "header 10s" default
   is intended to simply update the *existing* `ReadHeaderTimeout`'s default from 5s to 10s rather than
   add a new key, or (b) a genuinely distinct `HeaderTimeout` key is intended for some other purpose
   not yet clear from the source material. This plan does not invent the resolution; it is recorded as
   an open question for T001.
3. **Scaffold-template test harness location**: whether DX-01 T5's shared scaffold-test-harness
   primitive already exists in a form this story's template-render test can reuse is not confirmed as
   of this plan's writing (DX-01/DX-02 are W01-E04 scope, sequenced independently of this epic). T001
   must check for it before building a parallel harness.
4. **Example-config file updates**: whether `configs_base.yaml.tmpl`/`configs_local.yaml.tmpl`
   enumerate every `HTTP.*` key explicitly (requiring the 4 new keys to be added there too) or rely
   entirely on Go-level defaults (requiring no template change beyond `cmd_api_main.go.tmpl`) is not
   confirmed — T001 must inspect both files during implementation.

## Approval conditions

This plan is approved for implementation once: (a) the validation-policy question (unresolved
question 1) is confirmed with the framework architecture lead or resolved by an explicit, recorded
implementation-time decision; (b) the `HeaderTimeout`/`ReadHeaderTimeout` naming question (unresolved
question 2) is resolved the same way. Both are recorded in `implementation.md` once implementation
occurs, referencing this section.
