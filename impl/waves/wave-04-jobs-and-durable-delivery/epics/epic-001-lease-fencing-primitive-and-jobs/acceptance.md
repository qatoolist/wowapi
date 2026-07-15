---
id: W04-E01-ACCEPTANCE
type: epic-acceptance
epic: W04-E01
wave: W04
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W04-E01 — Epic-level acceptance

Reproduces and elaborates `epic.md`'s "Epic acceptance criteria" section as a standalone,
independently-referenceable acceptance record, consistent with the wave-level `../../acceptance.md`
pattern (AC-W04-01 there maps onto this epic).

## AC-W04-E01-01 — Shared primitive built once, as a reusable kernel building block

The shared lease/fencing primitive (`lease_token`, monotonic `lease_generation`,
`lease_expires_at`, optional heartbeat) is implemented as one reusable kernel building block, not a
`kernel/jobs`-only type — proven by this epic's own jobs-queue application (S002) consuming it
through a shared interface/package boundary, and structurally validated against W04-E02's (DATA-03)
and W04-E03's (DATA-04) own stated field needs before being treated as locked (per
RISK-W04-E01-001's mitigation). Traces to W04-E01-S001.

## AC-W04-E01-02 — Jobs queue fenced end-to-end

`jobs_queue` carries lease columns; claim SQL assigns a fresh `lease_token` and
`lease_generation+1` per claim; `complete`/`fail` finalize paths compare token/generation and
reject a mismatch (a stale finalize provably affects zero rows); `ReclaimStalled` bumps
`lease_generation` on reclaim, producing a provably new lease epoch, evidenced by a test asserting
the generation delta. Traces to W04-E01-S002.

## AC-W04-E01-03 — Idempotency contract and chaos proof

Every job worker declares exactly one of: inbox/effect-ledger unique on `(job_id, effect_name)`,
domain CAS, or provider idempotency key, and cannot register without doing so. A test proves fencing
the queue row alone does not undo an already-committed stale-worker domain transaction — the effect
ledger, not the queue row, is what catches an idempotency-ignoring worker. The named chaos test
`DATA-02/chaos/duplicate_worker_lease_expiry_test.go` proves exactly one logical effect recorded and
worker A's writes rejected at all three named boundaries (domain, external, finalize), and the
harness is built reusably for W04-E02/W04-E03 to consume without reimplementation. Traces to
W04-E01-S003.

## AC-W04-E01-04 — Independent review passed

All three stories (S001, S002, S003) have passed independent review per mandate §14. S001's review
specifically confirms the W02-E01-S002 interim-checkpoint-lease migration step is genuinely executed
and evidenced (not silently skipped), and that the primitive's field set was reviewed against
DATA-03/DATA-04's stated needs before being treated as locked. S003's review specifically confirms
the T5 worker-signature breaking change is recorded honestly as an open coordination note (not
silently resolved or hidden), and that the chaos test genuinely exercises all three named boundaries
rather than a subset.

## Acceptance authority

Data/reliability lead, per `../../wave.md`'s wave-level acceptance authority (PLAN §5.3's
accountable role for PF-DATA).
