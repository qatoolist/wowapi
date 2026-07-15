---
id: IMPL-W01-E03-S002
type: implementation-record
parent_story: W01-E03-S002
status: complete
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Implementation record — W01-E03-S002

Implemented 2026-07-13 by W01Http against HEAD 0a31186cada5c275a588c74081cf977adf346e61 (working tree; conductor owns the wave
commit). `git diff --stat` for this story's files:

```
app/boot.go                                        |  7 ++
kernel/config/security.go                          |  9 +++
kernel/config/security_test.go                     | 22 ++++++
kernel/httpx/decode.go                             | 19 +++++
kernel/httpx/router.go                             | 66 ++++++++++++++++
kernel/httpx/route_contract_test.go                | (new file)
internal/cli/templates/crud/resource.go.tmpl       | 49 +++++++++++--
internal/cli/scaffold_test.go                      | (1 new test in the shared file)
docs/user-guide/configuration.md                   | (security key table row)
```

## What was actually implemented

1. **RouteMeta seam (T001)** — `RouteMeta` gained `Request any` (a zero-value prototype of the
   request DTO, e.g. `Request: CreateOrderRequest{}`) and `NoRequestBody bool` (the waiver,
   doc-commented as forward-compatible with AR-04 T5, referencing requirement-inventory row
   AR-04). **Design resolutions recorded per plan.md approval conditions:**
   - *Candidate A vs. B (plan Q1)*: **Candidate A** (type-token prototype). Decisive factor: the
     MATRIX CS-08 / AR-03 coordination note — `RouteMeta` is a projection input AR-03 (W05) will
     derive OpenAPI request schemas from, so the concrete declared type must be reachable from
     the route table; Candidate B's boolean marker discards it. The A-drift risk (declared type
     vs. adaptor type parameter) is accepted and documented on the field; AR-03 can later add a
     reflect-based consistency check without any registration-API change.
   - *Check placement (plan Q4)*: in `Router.Handle` via a new `checkRequestContract(method,
     meta)` helper on Router — `RouteMeta.validate()` has no method access (confirmed structural
     fact); the helper plugs into the same `r.errs` accumulation as the existing invariant.
   - *Flag location (plan Q2)*: `config.Security.EnforceRouteContracts` (framework-level) — the
     prod-safety-floor flags already live on Framework, and `Security` is the posture section.
     `app.Boot` reads it and calls `router.RequireRequestContracts()` BEFORE modules register.
   - *Waiver shape (plan Q3)*: minimal bool (`NoRequestBody`), per the plan's recommendation.
2. **Boot-time check (T001)** — POST/PUT/PATCH only (`mutatingMethods` set; DELETE deliberately
   body-less-by-convention). Two invariants: (a) `Request` + `NoRequestBody` together is
   rejected UNCONDITIONALLY (both fields are new — nothing can depend on the combination;
   mirrors the Public/Permission contradiction rule); (b) missing both is rejected only under
   `RequireRequestContracts` (compat: profile-flag first, RISK-W01-002). Error text names the
   route (method + pattern prefix from Handle) and the remedy.
3. **Adaptor (T002)** — `httpx.ValidatedHandler[T](v, maxBytes, fn)` in `decode.go`, composing
   the UNCHANGED `BindAndValidate[T]` and routing failures through the existing
   `WriteError` KindValidation path (byte-identical 400 shape to a direct caller).
4. **crud template migration (T003)** — generated modules now emit `Create<R>Request` /
   `Update<R>Request` contract structs (fields carry `validate:"required"` starter tags with a
   comment telling authors to adjust), declare them on `RouteMeta.Request`, and wire create/update
   through `ValidatedHandler(v, 1<<20, ...)` with `v := mc.Validator()`. W01Gen's `.deactivate`
   DELETE permission change (same file, disjoint line) preserved — coordinated via irc.

## Interfaces introduced or changed

`httpx.RouteMeta`: 2 new fields (additive; existing literals compile and boot unchanged).
`httpx.Router.RequireRequestContracts()` (new). `httpx.ValidatedHandler[T]` (new).
`config.Security.EnforceRouteContracts` (new, default false). No existing signature changed;
`BindAndValidate` untouched.

## Security changes

Closes the "handler forgot BindAndValidate ⇒ zero validation, silently" gap for any product that
opts in, and makes the safe pattern the generated default.

## Tests added or modified

- `kernel/httpx/route_contract_test.go` (new): compat-default test, 3-verb rejection test,
  declared-contract pass, waiver exemption, contradiction guard, non-mutating exemption,
  adversarial 400-with-field-errors, valid-DTO passthrough.
- `kernel/config/security_test.go`: `TestEnforceRouteContractsDefaultsOff`.
- `internal/cli/scaffold_test.go`: `TestGenCRUDMutatingRoutesDeclareContractsAndUseValidatedHandler`.

## Known limitations

- Candidate A's declared-type/bound-type drift is not statically enforced (Go generics cannot
  express it); accepted and documented, AR-03 hook noted above.
- Enforcement applies to routes registered AFTER the mode is set (app.Boot sets it before any
  module registers, so the ordering is safe in the real path; hand-built routers must call
  `RequireRequestContracts()` first, as documented).
- Flipping the flag for wowsociety + auditing its routes is downstream work (out of scope per
  the story).

## Relationship to the approved plan

Matches plan.md; all four unresolved questions resolved and recorded above; no AC deviation.
