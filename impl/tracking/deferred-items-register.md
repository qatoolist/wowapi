---
id: TRACK-DEFERRED-ITEMS-REGISTER
type: register
title: Deferred items register — items explicitly deferred outside this implementation programme
status: active
created_at: 2026-07-12
updated_at: 2026-07-12
derived: true
---

# Deferred items register

DERIVED VIEW. Per mandate §11.10: every item deferred outside the current implementation
programme must include: Deferred ID | Source | Rationale | Prerequisites | Risk of deferral |
Intended future milestone | Approval. Canonical source = `impl/analysis/requirement-inventory.md`
deferred-disposition rows (§B DX-03; §C K-P2, B11/B12/B13) and
`docs/implementation/framework-backlog-p2-decisions.md`.

| Deferred ID | Source | Rationale | Prerequisites | Risk of deferral | Intended future milestone | Approval |
|---|---|---|---|---|---|---|
| DEF-01 | D-09 (REVIEW §U) | Secrets file-provider deferred as the "next increment" after v1's boot-time-once + restart-based-rotation contract; no vault client in the kernel for v1 | Demonstrated need for non-restart rotation | Low (documented, intentional v1 scope) | Post-programme, no wave assigned | Fable 5 (REVIEW §U) |
| DEF-02 | REVIEW §29 Q14 / K-P2 row (`requirement-inventory.md` §C) | gobreaker (circuit-breaker library) evaluation — P2 backlog item, no proven need yet | None blocking; P2 evaluation item | Low | Post-programme P2 backlog | REVIEW §K |
| DEF-03 | REVIEW §29 Q14 / K-P2 row (`requirement-inventory.md` §C) | jwx evaluation (JWKS library replacement) — same P2 framing as DEF-02, with a security-reviewed qualifier noted in REVIEW Q14 | None blocking; P2 evaluation item, security-reviewed | Low | Post-programme P2 backlog | REVIEW §K |
| DEF-04 | `framework-backlog-p2-decisions.md` (B11 — radix-router strategy) | Dispatch cost is flat (579.9→656.2 ns/op measured; re-verified 571.2→629.6 ns/op) across a 40× route-count increase; allocations flat at 14 regardless of route count — the exact quantities a radix tree would improve are already effectively constant. Dispatch is 2.4–2.8× cheaper than a single in-memory authz decision and 3–4 orders of magnitude below DB-backed budgets. Nothing to optimize | A benchmark showing dispatch ns/op growing materially with route count (non-flat), OR dispatch exceeding a meaningful fraction of the middleware/authz/DB request budget at realistic route counts | Low — evidence-backed, re-verified independently with no code changes between runs | Reopen only when the trigger benchmark is observed; no wave assigned | `framework-backlog-p2-decisions.md` decision + D-0090 re-verification |
| DEF-05 | `framework-backlog-p2-decisions.md` (B12 — typed schema unification: validation/OpenAPI/codegen single source) | Duplication is real but small — 2 hand-written OpenAPI fragments totalling 434 lines, 5 files with `validate:` tags, only 2 modules with request bodies, well under the ≈5–6-module reopen trigger. A reflection-based generator is overbuild for a P2 item whose charter is "defer until bottleneck proven" | Module count grows past hand-maintainability (≈5–6 modules with request bodies), OR a real defect is traced to OpenAPI-vs-handler drift. Recommended first increment when reopened: a drift-detection contract test comparing `validate`-tagged request structs against `openapi.json` fragments — cheaper than a full generator, targets the actual latent hazard (silent drift) directly | Medium if module count grows silently past the trigger without anyone noticing — mitigated by the drift-detection-test-first recommendation | Reopen when the module-count or drift-defect trigger is met; no wave assigned | `framework-backlog-p2-decisions.md` + D-0090 |
| DEF-06 | `framework-backlog-p2-decisions.md` (B13 — hot-reloadable DB overlays for i18n/rules) | No demonstrated operational need to edit translations or rule values at runtime without a redeploy; the freeze-at-boot invariant (Decision 3, ratified 2026-07-10) is load-bearing for request-time safety (no read/write race); W1 (wowsociety i18n migration) confirmed boot-time loading is sufficient for the real product's localization | A concrete operational requirement to change tenant/admin-editable i18n or rule values at runtime without redeploy. Prerequisites before building: immutability boundary, validation-on-write, reload metrics, and cache-invalidation semantics must be defined; the overlay applies last in the B1 precedence chain (embedded defaults → product overrides → product/module catalogs → Go bundles → DB overlay) | Low — freeze invariant is deliberate and load-bearing; building ahead of need would add cache-invalidation/consistency risk for no proven benefit | Reopen only on a concrete operational trigger; no wave assigned | `framework-backlog-p2-decisions.md` + D-0090 (2026-07-11 re-verification note: rule *values* already resolve live per-request via `kernel/rules/resolver.go`, so the B13 gap, if it ever materializes, is i18n-only) |
| DEF-07 | W03-E01-S003 (`review-gate-2026-07-16.md`) | Story's technical acceptance criteria (AC-W03-E01-S003-01/02) are independently re-verified and pass (DB-backed re-run, 2026-07-16); acceptance is blocked solely on a formal product-security-lead sign-off, a business approval not a technical re-verification gate | Product-security lead reviews and signs off on the assurance-freshness/credential-scheme-distinction change | Low — no surviving technical defect; purely a human-approval gate | Reopen (close) once sign-off is recorded; no wave reassignment needed | Conductor (Fable 5), pending product-security-lead sign-off |

## Summary

7 deferred items (DEF-01..DEF-07), each with an explicit reopen trigger or prerequisite, none
assigned to a wave in this programme (DEF-07 is W03-E01-S003's own human sign-off gate, tracked
here per the 2026-07-16 review-gate conditions). DX-03 (module DSL design) is a partial exception —
it has a nominal target `W06-E01-S001` for a design-investigation story only; the DSL build-out
itself remains deferred. See `requirement-traceability-matrix.md` exclusion notes.
