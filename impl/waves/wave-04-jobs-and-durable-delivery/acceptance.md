---
id: W04-ACCEPTANCE
type: wave-acceptance
wave: W04
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W04 — Wave-level acceptance

## AC-W04-01 — Shared lease/fencing primitive built once, reused three times

The shared lease/fencing primitive (`lease_token`, monotonic `lease_generation`,
`lease_expires_at`, optional heartbeat) exists as one reusable kernel building block and is
consumed — not copied — by jobs (E01), notify/webhook (E02), and bulk processing (E03); the named
chaos test `DATA-02/chaos/duplicate_worker_lease_expiry_test.go` proves exactly one logical effect
recorded and the stale worker's writes rejected at all three named boundaries (domain, external,
finalize). Traces to W04-E01-S001, W04-E01-S002, W04-E01-S003.

## AC-W04-02 — Remote I/O moved outside database transactions; retry unified

`kernel/notify` and `kernel/webhook.deliverToEndpoint` run no network/secret-resolution call while
a database transaction is open, proven by the three-stage claim→effect→finalize protocol and its
boundary-matrix test; inbound webhook verification is immune to a secret rotation/deactivation race
between snapshot and verification; the 6-boundary chaos test (`DATA-03/chaos/`) proves zero
duplicate external effects across all 6 fault points on both notify and webhook; `cenkalti/
backoff/v5` replaces both hand-rolled retry implementations with retry-schedule parity proven under
fault injection. Traces to W04-E02-S001, W04-E02-S002, W04-E02-S003.

## AC-W04-03 — Bulk multi-worker processing actually safe, not documented-safe

The false "replica-safe" migration-comment claim is removed and a second concurrent processor is
mechanically rejected via the stopgap (advisory lock/CAS) proven by a 2-processor concurrency test;
the leased claim SQL provably uses `SKIP LOCKED` with a bounded batch (`EXPLAIN`-plan assertion);
the named chaos test `DATA-04/chaos/duplicate_worker_test.go` proves ≥2 concurrent processors
claim/retry/pause/resume/cancel the same operation without duplicate effects or stale finalization,
matching the Wave-3 exit-gate wording verbatim. Traces to W04-E03-S001, W04-E03-S002.

## AC-W04-04 — Audit hash chain covers every persisted field; readiness diagnostics are truthful

The widened `chainHash` (D-04's `hash_version smallint` discriminator) breaks verification when
`metadata`, `tx_id`, or any other declared field is mutated, proven by an independent per-field
tamper test; historical rows verify under the `hash_version=1` branch; external anchor verification
detects tamper even if the local `head_hash` were compromised; DSR export is a checksummed encrypted
immutable artifact; every registered record class without an export/erase callback receives an
explicit not-applicable status, never a silent omission; the readiness endpoint returns 503 when
applied-migration version lags expected and reports seed/rule/model-hash once healthy; `config
doctor` discovers the product root via `go env GOMOD`/`--project` regardless of invocation
directory. DX-07 T4 is explicitly out of scope for this wave, deferred-linked to W05-E03-S002's
AR-04 T5 waiver mechanism — its absence does not block this AC. Traces to W04-E04-S001,
W04-E04-S002, W04-E04-S003.

## AC-W04-05 — Independent review passed

Every W04 story has passed independent review per mandate §14. E01-S001 (the keystone primitive)
and E04-S001 (DATA-08 W6-T1, confirmed the single highest-risk task in this wave's scope) are
specifically checked for: no silent scope reduction against PLAN's own T-row acceptance criteria;
the RISK-W04-001 interim-lease-migration correctly executed and evidenced (not silently skipped);
the D-04 version-branch verification genuinely proven on both historical and new-row branches (not
merely claimed); DX-07 T4's deferral to W05 correctly recorded, not silently absorbed as "done" by
omission.

## Acceptance authority

Data/reliability lead, per `wave.md`'s "Acceptance authority."
