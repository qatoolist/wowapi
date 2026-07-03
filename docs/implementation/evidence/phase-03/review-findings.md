# Phase 3 — Review Findings

Two parallel critique agents reviewed the HTTP/error/validation/pagination/filtering/idempotency
slice (2026-07-03): **S** = hostile security reviewer (HTTP contracts, injection, secret leakage,
idempotency — ran live Postgres probes) and **A** = architecture/API reviewer. Both independently
found the idempotency concurrency race (SEC-16 = ARCH-27); it was reproduced against live Postgres
before fixing. All fixes carry regression tests; the race fix has a concurrency test passing ×5
under `-race`.

| ID | Sev | Finding | Resolution | Status |
|---|---|---|---|---|
| SEC-16 / ARCH-27 | critical | idempotency claim raced: `SELECT FOR UPDATE` can't lock a not-yet-existing row, so concurrent first-uses of a key both went Fresh and the unconditional upsert clobbered a completed response → operation runs twice, stored response destroyed — reproduced live | atomic `INSERT … ON CONFLICT DO NOTHING RETURNING` (only a real insert is Fresh) → else `SELECT … FOR UPDATE` + branch (completed/hash-mismatch/expired-reclaim/in-flight); `TestIntegrationIdempotencyConcurrent` (8 goroutines, exactly-once, ×5 `-race`) | **fixed** |
| SEC-18 | medium | `Recover` appended a problem body to already-written responses (corrupt body) and swallowed `http.ErrAbortHandler` — reproduced | write-tracking ResponseWriter (skip problem body once bytes are sent) + re-panic on ErrAbortHandler; `TestRecoverDoesNotCorruptWrittenResponse`, `TestRecoverPropagatesAbortHandler` | **fixed** |
| SEC-17 | medium | the in-flight/retry_later branch was effectively dead under the single-tx design | made real by the SEC-16 rewrite (explicit claim/read split); a live un-expired in_progress row now yields retry_later | **fixed (via SEC-16)** |
| SEC-19 | low | `RequestHash` omitted the query string → two requests differing only in query params shared a stored response | hash now includes `URL.RawQuery`; doc requires deterministic non-empty canonical | **fixed** |
| SEC-20 | low | duplicate JSON keys accepted; Content-Type unchecked | accepted (defense-in-depth): strict decode + `DisallowUnknownFields` + domain validation suffice; noted for a future WAF-alignment pass | **accepted (noted)** |
| SEC-21 / ARCH-28 | info | `Router.Err()` not enforced at any boot path yet | deferred to Phase 5 app wiring (route→server + boot Err() check); the metadata invariant itself is sound and unit-tested | **accepted (Phase 5)** |
| SEC-22 | info | a forged cursor could feed arbitrary keys into a future keyset query | `filtering.KeysetClause` validates cursor keys == sort columns exactly and pulls only VALUES (bound); `TestKeysetClauseRejectsMismatchedCursor`, `TestKeysetClauseColumnsAreAllowlisted` | **fixed (with ARCH-31)** |
| SEC-23 | info | `WithIdempotency` could store a non-2xx body; LIKE leading-wildcard perf | only 2xx stored (else Discard); LIKE perf noted for a Phase 11 query-budget pass | **fixed / noted** |
| ARCH-29 | medium | `DecodeJSON` accepted literal `null` as a zero-value struct | decode into `*T`; nil → "request body is required"; `TestDecodeJSONRejectsNull` | **fixed** |
| ARCH-30 | medium | no machine-checked import boundary; new database→errors edge undocumented | `depguard` rule added (kernel must not import module/app/adapters/testkit — passes clean); D-0033 authorizes database emitting taxonomy Kinds | **fixed** |
| ARCH-31 | medium | keyset pagination couldn't be assembled from these types (blueprint 05 §2 `KeysetClause` was missing) | added `filtering.KeysetClause` + `Sort.Terms()`; lexicographic mixed-direction predicate, placeholder-threaded; 4 tests | **fixed** |
| ARCH-32 | medium | non-2xx responses stored/replayed inconsistently | store only 2xx; `IdemStore.Discard` for the rest | **fixed** |
| ARCH-33 | low | `errors.E` silently drops unrecognized variadic args | accepted: documented "last error/op wins; other types ignored" is the contract; a debug guard is overkill | **accepted** |
| ARCH-34 | low | `RequireIfMatch` rejected `If-Match: *` | intentional — documented + explicit clearer error (optimistic concurrency requires a concrete version) | **fixed (doc + error)** |
| ARCH-35 | low | `ScopeExtractor` returns `any`; Phase 4 break | accepted: becomes `authz.Target` in Phase 4; rarely-set optional field, noted | **accepted (Phase 4)** |

Reviewer-confirmed solid: filtering/sort injection-proofness (columns only from allowlist, values
always $N — the `DROP TABLE` payload lands only in args); internal-error non-leakage (KindInternal
never sets Detail; no Op/cause/stack on the wire); validation redaction (never echoes field
values); cursor decode DoS-safe (4096 cap before base64, panic-free); cross-tenant idempotency
replay prevented by composite PK + FORCE RLS under the verified non-superuser app_rt.

Residual risk:
- `Router.Err()` boot enforcement is a Phase 5 must-do (SEC-21).
- `KeysetClause` is now implemented and reviewed against SEC-22; its first real use lands in Phase 5.
- `IdempotencyConfig.ActorScope` is a caller string until the Phase 4 actor identity supplies it.
