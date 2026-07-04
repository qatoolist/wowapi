# Hardening H2 — Async operability — proof bundle

Plan: [../../hardening-plan.md](../../hardening-plan.md). H2 closes the async-operability gaps. It is
delivered in coherent, individually QA-gated commits:

- **R4 — DLQ operability** — DONE (D-0062). This bundle.
- R3 — SLA sweeper registration + leader-safe scheduling — pending (paired with the E5 scheduler).
- O2 — migration forward/down CI drill + expand/contract doc — pending.
- O5 — backup/restore runbook + drill — pending.

## R4 — DLQ operability

| Verdict | Fix |
|---|---|
| real (P0) — dead-lettering worked (`jobs_queue.status='discarded'`, `events_outbox.dispatch_status='dead'`) but there was no way to inspect, replay, or purge DLQ entries | Admin functions `jobs.{ListDead,ReplayDead,DiscardDead}` + `outbox.{ListDeadEvents,ReplayDeadEvent,DiscardDeadEvent}`; `wowapi dlq <jobs\|events> <list\|inspect\|replay\|discard>` CLI; migration 00013 grants `DELETE` to app_platform |

Replay safety (roadmap "replay is idempotent-safe"): jobs are at-least-once with idempotent workers;
events dedup via the `processed_events` inbox on re-dispatch — so replay cannot double-apply a correct
handler/worker.

## Implementation inventory

New: `kernel/jobs/dlq.go` (+`dlq_test.go`), `kernel/outbox/dlq.go` (+`dlq_test.go`),
`internal/cli/dlq_cmd.go` (+`dlq_cmd_test.go`), `migrations/00013_dlq_admin.sql`.
Changed: `internal/cli/cli.go` (dispatch + help), `migrations/migrations_test.go`.

## Acceptance

`make ci` + `make ci-container` green: **0 FAIL, 0 SKIP, 74 packages**, DB tests forced. DLQ list/replay/
discard proven against real Postgres for both jobs and events, incl. `KindNotFound` on wrong/again ids;
CLI arg-validation unit-tested (exit 2 for bad domain/action/id before any DB connect).

Note on coverage: the `wowapi dlq` DB path (`dlqPool` → kernel funcs) is thin; the kernel functions
carry the integration coverage. An end-to-end CLI-through-`Run()` DB test was dropped because
`testkit.NewDB` isolates each test in its own database while the CLI connects via `DATABASE_URL` (the
base DB) — wiring the per-test DSN into the CLI was judged not worth the testkit surface.
</content>
