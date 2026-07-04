# Hardening H1 — review findings

Self-review of the H1 changes (architecture, security, boundaries, tests). No finding silently dropped.

| # | Sev | Finding | Resolution |
|---|---|---|---|
| 1 | med | `config.HTTP` gained a `[]string`, making it (and `Framework`) non-comparable — two tests used `!=` | Fixed: `reflect.DeepEqual` in `load_test.go` + `app/views_test.go`. No production `==` on these structs. |
| 2 | med | Signed-cursor envelope could be confused with a flat cursor whose columns are literally `__s`/`__v` | Envelope requires exactly two keys `__s`(string)+`__v`(object); scalar keyset values can't satisfy the object assertion, so it falls back to flat. DB column names never take `__` forms. |
| 3 | low | `Timeout` uses `http.TimeoutHandler`, whose 503 body is plain text, not RFC 7807 | Accepted for this phase: correctness (buffered writer, ctx cancel) matters more than the body shape on the rare timeout path; the reference proxy can normalize. Noted for a future problem-details timeout. |
| 4 | med | R6: is `FOR UPDATE` alone sufficient, or is the `legal_hold=false` UPDATE guard redundant? | Kept both — FOR UPDATE closes the race under READ COMMITTED (EvalPlanQual); the guard is defense-in-depth against a future isolation/refactor change. Revert-test proves the combination is load-bearing. |
| 5 | low | S5 sweep runs as `app_platform` cross-tenant — does it widen the platform role's blast radius? | Mirrors the existing `outbox_relay_all` pattern (00007); grant is SELECT+DELETE on one table via an explicit named policy. app_rt's tenant-scoped lifecycle is unchanged. |
| 6 | low | Fuzz seed corpus runs in CI but deep fuzzing does not | Intentional — deep fuzzing is non-deterministic and slow; `make test-fuzz` (FUZZTIME overridable) is the nightly/drill entry point. Documented in the Makefile target. |

## Boundaries

No import-law violations: edge middleware stays in `kernel/httpx`; the cursor signature lives in
`kernel/pagination` with the sort-aware mint/verify in `kernel/filtering` (which already depends on
pagination — no new edge). `SweepExpired` uses the existing `TxManager.Platform` seam. `make ci`
boundary lint passed.
