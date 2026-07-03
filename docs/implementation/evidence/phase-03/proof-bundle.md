# Phase 3 — Proof Bundle

Scope (phase-plan row 3): `kernel/errors`, `kernel/httpx` (middleware, RouteMeta, response/problem,
decode, listing, etag), `kernel/validation`, `kernel/pagination`, `kernel/filtering`, idempotency
helpers, `module.Context` growth. Date: 2026-07-03.

## 1. Decision evidence
Pre/at-code: D-0031 (idempotency_keys migration 00003 pulled forward), D-0032 (Context gains
Routes/Validator). Post-review: D-0033 (database may emit taxonomy Kinds; depguard enforces the
import law), D-0034 (idempotency review-finding resolutions). All in
`docs/implementation/decisions.md`.

## 2. Discussion evidence
- Idempotency claim atomicity: the naive SELECT-FOR-UPDATE-then-upsert loses under concurrency
  (a FOR UPDATE cannot lock a non-existent row). Both review agents reproduced the double-execute
  independently; resolved with an atomic INSERT…ON CONFLICT DO NOTHING RETURNING claim.
- Response storage format: jsonb reformats whitespace and broke byte-exact replay → switched the
  column to bytea (caught by the first integration run, before review).
- RequestHash design: hash the caller's canonical command bytes (+ method/path/query), not the
  already-consumed raw body — format-independent and unknown-field-safe; documented in D-0034.
- KeysetClause: the blueprint (05 §2) listed it but Phase 3 initially missed it; the architecture
  review flagged that the Sort/Cursor types couldn't assemble a keyset query. Added with
  cursor-key allowlisting (also closes the SEC-22 injection concern).

## 3. Critique/review evidence
`review-findings.md`: 15 findings (1 critical reproduced, 4 medium, rest low/info). The critical
(SEC-16/ARCH-27) was reproduced on live Postgres by both agents, fixed, and now has an 8-goroutine
concurrency regression test passing ×5 under `-race`. Every finding fixed or accepted with
rationale + phase tracking.

## 4. Implementation evidence
New: `kernel/errors/`, `kernel/httpx/` (response, errors, decode, router, etag, idempotency,
listing, middleware, context), `kernel/validation/`, `kernel/pagination/` (pagination, cursor),
`kernel/filtering/` (filtering, sort, keyset), `kernel/database/idempotency.go`,
`migrations/00003_idempotency.sql`. Changed: `module/module.go`, `app/context.go`,
`.golangci.yml` (depguard). Deps: go-playground/validator/v10.
Team: 2 parallel implementation agents (validation; pagination+filtering) + lead
(errors, httpx, idempotency, all review fixes); 2 parallel review agents.

## 5. Verification evidence
`command-log.md`: per-package unit suites, idempotency store + WithIdempotency integration (live
PG, host + tools container), the SEC-16 concurrency regression ×5 `-race`, depguard import-law
check (0 violations), full `make ci` + `make test-integration` after fixes. Graphify updated.

## 6. Acceptance evidence
`acceptance-map.md`: all 16 Phase 3 exit criteria mapped to code + named tests; acceptance #3
(route metadata enforcement) and #11 (error contracts / list helpers / idempotency) each to
specific tests. Carried forward:
- `Router.Err()` boot enforcement + route→server wiring → Phase 5 (SEC-21/ARCH-28).
- `ScopeExtractor any` → `authz.Target` → Phase 4 (ARCH-35).
- First real `KeysetClause` use in a list endpoint → Phase 5.
- `IdempotencyConfig.ActorScope` real source (actor identity) → Phase 4.
- Graphify semantic `extract` still blocked on LLM key (R11).
