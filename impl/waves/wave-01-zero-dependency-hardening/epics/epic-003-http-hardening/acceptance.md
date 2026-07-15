---
id: W01-E03-ACCEPTANCE
type: epic-acceptance
epic: W01-E03
wave: W01
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W01-E03 — Epic-level acceptance

Reproduced from `epic.md` "Epic acceptance criteria" as a standalone, independently-referenceable
file per the tracking-file set this epic ships (mirrors the wave-level `acceptance.md` pattern).

## AC-W01-E03-01 — Server timeouts and body bounds enforced

All four new HTTP timeout config keys (`ReadTimeout`, `WriteTimeout`, `IdleTimeout`, `HeaderTimeout`)
are present in the scaffold-rendered `http.Server{}` literal with the MATRIX CS-09 default values
(30s/60s/120s/10s). A template-render test asserts all four are present in generated output. A
prod-profile config with an explicit zero-value timeout fails `config.Validate`. `kernel/httpx/csrf.go`'s
`r.FormValue` call is wrapped in a defensive `http.MaxBytesReader`. Traces to W01-E03-S001.

## AC-W01-E03-02 — Central validation enforcement live

Boot rejects a fixture POST/PUT/PATCH route with no declared `RouteMeta.Request` contract, behind the
profile flag. An adversarial invalid-DTO POST to a route that does declare a contract, routed through
the new handler adaptor, returns 400 with field errors. The fixture-route boot-passes-today /
boot-fails-after-T1 fail-first sequence is evidenced (mandate §13 fail-first requirement). Traces to
W01-E03-S002.

## AC-W01-E03-03 — Independent review passed

Both stories have passed independent review per mandate §14 and `governance/definition-of-done.md`'s
independent-review checklist. S001 is specifically checked for the "safe defaults, not zero" framing
(RISK-W01-003) and for not silently duplicating the three already-existing HTTP timeout config keys
(RISK-W01-E03-001). S002 is specifically checked for profile-flag compat discipline (RISK-W01-002)
and for not silently building a waiver design that conflicts with the not-yet-built AR-04 T5.

## Acceptance authority

Framework architecture lead (role-based per `wave.md`'s ARCH-adjacent split — this epic's HTTP-layer
work is ARCH-adjacent, not DX-04's generator/doc-work track).

## Acceptance record — 2026-07-13

Satisfied 2026-07-13. All acceptance criteria for W01-E03 are met; independent review passed
(W01ReviewGate); accepted by conductor.
