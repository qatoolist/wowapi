# Phase 3 — Acceptance Map

Phase 3 exit criteria (Goal 2 Phase 3 + phase-plan row 3 + blueprint 04 §4–5 / 05 §1–2) → proof.

| # | Criterion | Proof |
|---|---|---|
| 1 | `kernel/errors` closed taxonomy → HTTP mapping | `kernel/errors/errors.go`; `TestKindMappingClosedSet` (all 13 kinds), `TestTenantIsolationMaskedAs404`, `TestUnknownKindIsInternal` |
| 2 | RFC 9457 problem-details is the only error body | `kernel/httpx/errors.go` `ProblemError` + `WriteError`; `TestWriteErrorMapsKinds`, `TestWriteErrorValidationCarriesFields` |
| 3 | Internal errors never leak cause/Op/stack | `TestWriteErrorInternalNeverLeaks` (plain err + KindInternal → opaque 500, empty Detail); `Recover` no-leak test |
| 4 | **Routes without metadata fail registration** (acceptance #3) | `RouteMeta.validate` + `Router`; `TestRouterRejectsMissingMetadata`, `TestRouterRejectsPublicWithPermission`, `TestRouterAccumulatesErrors`, `TestRouterRejectsDuplicate`, `TestRouterValidRoutesAndPermissions` (boot-path enforcement wired Phase 5, SEC-21) |
| 5 | Strict JSON decode (unknown fields, size cap, single value, no null) | `kernel/httpx/decode.go`; `TestDecodeJSONStrict`, `TestDecodeJSONRejectsNull` |
| 6 | Shape validation with field paths, no value echo | `kernel/validation`; 11 tests incl. json field paths + value-absent-from-message |
| 7 | **Error contracts + list helpers + idempotency tested** (acceptance #11) | error tests above; `kernel/pagination` + `kernel/filtering` suites (injection→placeholder, clamping, keyset); idempotency integration tests |
| 8 | Pagination: opaque cursor, offset/cursor envelopes, clamping | `kernel/pagination`; round-trip, malformed/oversized cursor → KindValidation, per_page clamping |
| 9 | Filtering/sorting: allowlist-driven, injection-proof | `kernel/filtering`; `DROP TABLE` payload lands only in args, unknown field/op rejected, placeholder numbering; `KeysetClause` (05 §2) with cursor-key allowlisting |
| 10 | Idempotency: at-most-once, replay, conflict, in-flight, cross-tenant | `kernel/database/idempotency.go` + `WithIdempotency`; `TestIntegrationIdempotencyStore/InFlight/Concurrent`, `TestIntegrationWithIdempotency/Conflict` — byte-exact replay, RLS-scoped, concurrency-safe |
| 11 | ETag / If-Match optimistic concurrency helpers | `kernel/httpx/etag.go`; `TestETagRoundTrip`, `TestRequireIfMatchMissing`, `*`-rejection |
| 12 | Middleware: request id, panic recovery without leak/corruption | `kernel/httpx/middleware.go`; `TestRecoverMiddlewareReturns500WithoutLeaking`, `TestRecoverDoesNotCorruptWrittenResponse`, `TestRecoverPropagatesAbortHandler`, `TestRequestIDHonorsInbound` |
| 13 | `module.Context` grows Routes()/Validator() (D-0032) | `module/module.go`, `app/context.go`; app wiring to a server lands Phase 5 |
| 14 | Import law machine-checked | `.golangci.yml` depguard (0 violations) + `scripts/lint_boundaries.sh` OK |
| 15 | Container-first verification | host `make ci` + `make test-integration`; tools-container `make test-integration` green |
| 16 | Evidence bundle + reviews | this directory; review-findings.md (security + architecture, 1 reproduced critical fixed) |
