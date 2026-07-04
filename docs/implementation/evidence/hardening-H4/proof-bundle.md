# Hardening H4 — Durable audit + tamper-evidence — proof bundle

Plan: [../../hardening-plan.md](../../hardening-plan.md). H4 builds the compliance audit layer:

- **E1 — durable field-level audit trail + query API** — DONE (D-0069).
- **S6 — cryptographic tamper-evidence (hash-chaining + verification)** — DONE (D-0070). **H4 complete.**

## S6 — audit tamper-evidence (hash-chaining)

| Verdict | Fix |
|---|---|
| real (P0) — audit rows were append-only (grant-enforced) but had no cryptographic proof against an owner/DBA mutation | Migration 00018 adds `seq`/`row_hash`/`prev_hash` to `audit_logs` + a per-tenant `audit_chain(next_seq, head_hash)`. `Record` now chains: it locks the tenant's chain head, assigns a gap-free `seq`, computes `row_hash = sha256(prev_hash ‖ length-prefixed canonical row)`, and advances the head — all in the caller's tx. `Verify` walks the chain recomputing hashes and checking prev-links + seq continuity; `Anchor` exports the head (seq + hash) for external notarization. |

Correctness details: the timestamp is truncated to microseconds (Postgres `timestamptz` precision) so
Record's hash input matches what Verify reads back; `metadata` (jsonb, which reformats on round-trip) is
deliberately excluded from the hash — the audited change (entity/field/before/after/actor/action) is
what the chain protects. Length-prefixed field encoding prevents value-boundary collisions.

Tests (`kernel/audit/audit_test.go`): chain verifies clean (Count/HeadSeq/Anchor); **mutation detected**
— an admin-pool UPDATE of a committed row's action (bypassing app_rt append-only) → Verify fails at that
seq ("row_hash does not match"); **deletion detected** — an admin DELETE → Verify fails at the seq gap.
Gate: 0 FAIL, 0 SKIP, 80 packages; boundary lint + 00018 reversibility pass.

## E1 — durable field-level audit trail

| Verdict | Fix |
|---|---|
| real (P0) — the only audit was `authz.AuditSink.AuthzDenial` (denial logging via a nil-safe log sink); the kernel explicitly stubbed "durable audit_logs writer replaces it". No durable, field-level, queryable audit. | `kernel/audit.Writer` over an `audit_logs` table (migration 00017). `Record` appends an entry (entity, field, before/after, actor, actor-kind, impersonator, request-id, action, reason, metadata) **inside the caller's tenant tx** — so an audit row commits iff the change does. `Query(Filter)` reads it back (by entity/actor/action, newest-first). A `Redactor` hook masks sensitive field values before persistence (per-module redaction). |

Append-only is grant-enforced: `app_rt` has `SELECT, INSERT` on `audit_logs` but **no UPDATE/DELETE**, so
the runtime cannot rewrite history — proven by a test that asserts app_rt UPDATE and DELETE are denied.
This gives integrity; cryptographic tamper-*evidence* (detecting an owner/DBA mutation) is S6's
hash-chaining, which builds on this table.

Tests (`kernel/audit/audit_test.go`, real Postgres + RLS): record + query with field-level before/after
and actor/request-id capture (newest-first via UUIDv7 tiebreaker); **append-only** (app_rt UPDATE +
DELETE both denied); redaction (an `ssn` redactor masks the stored values); tenant isolation (tenant 2
cannot see tenant 1's rows). Gate: 0 FAIL, 0 SKIP, 80 packages; boundary lint + 00017 reversibility pass.

Usable today: a service records with its own `database.TenantDB` (`audit.New(idgen, redactor).Record(...)`).
Follow-ups: automatic trigger-based field capture; a `module.Context` accessor.
**Closed in the post-hardening review (D-0077):** bridging `authz` denials to durable audit rows — the
`kernel.durableAudit` sink writes an `authz.denied` row in its own tenant tx (the read-only eval tx
cannot). S6 hash-chaining is delivered above.
</content>
