<!-- markdownlint-disable MD013 -->

# Framework Backlog P2 — Evidence-Backed Decisions (B11, B12, B13)

Date: 2026-07-10. Branch: `feat/backlog-p2` (off `feat/backlog-p1` @ 18c9589).

The P2 items in `framework-engineering-backlog.md` (completed B1–B10 programme record, archived 2026-07-11 to `wowapi2/archive/reports/`; each parked item's substance is restated in full below) are, by their own charter, "**only if data/need demands**". This document records the evidence gathered and the decision for each. Two items carry explicit gates (B11: benchmark; B13: operational need); B12 is gated on the "do not overbuild / defer until a bottleneck is proven" P2 rule. **All three are PARKED**, each with the concrete trigger that would reopen it. The corresponding delivery of value in this cycle is **W1** (wowsociety i18n migration), tracked separately.

## B11 — Radix-router strategy → PARKED

**Gate:** implement only if benchmark data shows `net/http.ServeMux` dispatch is a real cost; must preserve RouteMeta / boot-validation / OpenAPI contracts.

**Evidence (measured on this branch, `kernel/httpx` `BenchmarkDispatch`, full SecureHandler→gateRoute chain):**

| Routes | ns/op | allocs/op |
|---|---|---|
| 50 | 579.9 | 14 |
| 500 | 607.3 | 14 |
| 2000 | 656.2 | 14 |

Reproduces B5's recorded budget (577 / 604 / 666 ns/op) in `bench-budgets.txt`.

**Decision:** a radix tree optimizes route-**lookup scaling** and per-lookup allocations. Here dispatch cost rises only ~13% (580→656 ns) across a **40× route-count increase**, and allocations are **flat at 14** regardless of route count — i.e. the exact quantities a radix tree would improve are already effectively constant. Dispatch is 2.4–2.8× cheaper than a single in-memory authz decision (~1600 ns) and 3–4 orders of magnitude below the DB-backed budgets. There is no bottleneck to remove. **Keep `net/http.ServeMux`.**

**Reopen trigger:** a benchmark showing dispatch ns/op growing materially with route count (non-flat), OR dispatch exceeding a meaningful fraction of the middleware/authz/DB request budget at realistic route counts. If ever built, it must be a strategy interface behind the existing `RouteMeta` contract (boot validation, OpenAPI/seed sync, permission manifest) with these benches as the acceptance gate.

## B12 — Typed schema unification (validation / OpenAPI / codegen single source) → PARKED (narrowed)

**Gate (P2 rule):** deliver only if the duplication is a proven, present bottleneck; do not overbuild.

**Evidence (measured):**

- The duplication is real and specific: **Go request DTOs carry `validate:"…"` tags** (consumed by `httpx.BindAndValidate[T]`, `kernel/httpx/decode.go:55`) **while each module hand-authors an `openapi.json` fragment** (`module.Context.OpenAPI([]byte)`, `app/context.go:324`) describing the same endpoint shape. `wowapi openapi` (`internal/cli/openapi_cmd.go`) only **merges** fragments (`gatherFragments`/`mergeFragment`) — there is **no** type→schema generation today.
- Magnitude in the one real product (`wowsociety`): **2** hand-written fragments totalling **434** lines; **5** files with `validate:` tags. The scaffold ships a **1-line** `openapi.json` stub.
- `kernel/rules` `RuleValueSchema` (the strict grammar from B3) validates **rule configuration values**, a separate domain from API DTOs — explicitly **out of scope** for this unification.

**Decision:** a robust reflection-based generator (Go type + `validate` tags → OpenAPI schema + request/response `$ref`, nested types, enum/bounds mapping) is a large, edge-case-heavy kernel feature. Building it to de-duplicate **2 modules / 434 lines** is overbuild for a P2 item whose charter is "defer until bottlenecks proven." The design evidence does not justify it now. **Parked.**

**Recommended first increment when reopened (do this before any generator):** a drift-detection **contract test** that compares each module's `validate`-tagged request struct against its `openapi.json` fragment and fails on divergence. This targets the actual latent hazard — silent drift between the hand-written spec and the enforced validation — at a fraction of the cost of a generator, and defers the generator until it is clearly warranted.

**Reopen trigger:** the module count grows past hand-maintainability (≈5–6 modules with request bodies), OR a real defect is traced to OpenAPI-vs-handler drift. Rule-value schema stays separate.

## B13 — Hot-reloadable DB overlays for i18n/rules → PARKED

**Gate:** opt-in overlay only, gated on a **demonstrated operational need** (Decision 3), and only after immutability / validation / metrics / cache-invalidation semantics are defined.

**Evidence:**

- Decision 3 (ratified 2026-07-10) set i18n/rules catalogs to **freeze at boot** by default; hot-reload was explicitly deferred to this opt-in P2 item.
- No operational need has been demonstrated: there is no stated requirement to edit translations or rule values at runtime without a redeploy. i18n and rules are boot-loaded and stable. **W1** confirms boot-time loading via the B1 source layers is sufficient for the real product's localization.
- The freeze invariant is load-bearing for request-time safety (no read/write race); a DB overlay would have to re-introduce controlled mutability with a defined consistency + cache-invalidation contract.

**Decision:** **Parked.** No demonstrated need; building it would add cache-invalidation and consistency risk against the deliberate freeze invariant for no proven benefit.

**Reopen trigger:** a concrete operational requirement to change tenant/admin-editable i18n or rule values at runtime without redeploy. Prerequisites first: define immutability boundary, validation-on-write, reload metrics, and cache-invalidation semantics; the overlay applies **last** in the B1 precedence chain (embedded defaults → product overrides → product/module catalogs → Go bundles → **DB overlay**).

## Summary

| Item | Decision | Basis |
|---|---|---|
| B11 radix router | Parked | Dispatch flat (579→656 ns, 14 allocs) across 40× routes; nothing to optimize |
| B12 schema unification | Parked (narrowed) | Duplication real but small (2 modules/434 lines); generator = overbuild for P2; drift-check is the first increment when reopened |
| B13 hot-reload overlays | Parked | No operational need demonstrated (Decision 3); freeze invariant sufficient |

P2 delivery this cycle is **W1** (wowsociety consumes the B1 i18n loaders), which required and exercised real framework capability rather than speculative building.

## Re-verification — 2026-07-11 (see decisions.md D-0090)

An independent `/goal` pass re-derived all three items' evidence from live source (not from this doc) with
no code changes in between. Result: **all three decisions confirmed unchanged.**

- B11: live `BenchmarkDispatch` re-run measured 571.2/590.4/629.6 ns/op at 50/500/2000 routes, flat 14
  allocs/op — matches (slightly beats) the table above. RouteMeta boot-validation is confirmed live
  (`app/boot.go:254` permission-sync gate); OpenAPI generation stays hand-authored and unwired to
  `Route`/`Router.Routes()` (test-only callers) — see decisions.md D-0090 for the correction.
- B12: wowsociety recount matched exactly — 2 fragments/434 lines, 5 validate-tagged files, only 2 modules
  with request bodies (well under the ≈5-6 reopen trigger).
- B13: freeze-at-boot confirmed live in `app/boot.go`/`kernel/i18n`; no operational-need documentation found
  anywhere in the repo. New clarifying note: rule *values* already resolve live from `rule_versions` per
  request (`kernel/rules/resolver.go`), so the no-redeploy-mutation goal is already met for rules today —
  the B13 gap, if it ever materializes, is i18n-only.

Full citations and reasoning in `docs/implementation/decisions.md` D-0090. No reopen trigger was met for any
item; no code changed.
