# Hardening H4 — Durable audit + tamper-evidence — proof bundle

Plan: [../../hardening-plan.md](../../hardening-plan.md). H4 builds the compliance audit layer:

- **E1 — durable field-level audit trail + query API** — DONE (D-0069). This bundle.
- S6 — cryptographic tamper-evidence (per-tenant-per-period hash-chaining + verification tool) —
  pending. It layers on the append-only `audit_logs` table this phase creates.

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
Follow-ups (documented, not this pass): a `module.Context` accessor; bridging `authz` denials to durable
rows (the `AuthzDenial` sink lacks a tx handle); automatic trigger-based field capture; S6 hash-chaining.
</content>
