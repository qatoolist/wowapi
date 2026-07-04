# Hardening H2 — Async operability — proof bundle

Plan: [../../hardening-plan.md](../../hardening-plan.md). H2 closes the async-operability gaps. It is
delivered in coherent, individually QA-gated commits:

- **R4 — DLQ operability** — DONE (D-0062).
- **O2 — migration forward/down CI drill + expand/contract doc** — DONE (D-0063).
- **O5 — backup/restore runbook + drill** — DONE (D-0063).
- **E5 scheduler + R3 SLA-sweeper leader-safe scheduling** — DONE (D-0065).

## E5 + R3 — recurring scheduler + leader-safe kernel sweeps

| Verdict | Fix |
|---|---|
| real (P0) — SLA sweeper + idempotency sweep existed as methods but nothing ran them on a schedule, and N replicas would all fire | `jobs.Scheduler` over a new `schedules` table (migration 00014): fixed-interval tasks, each due tick claimed by an atomic `FOR UPDATE SKIP LOCKED` + `next_run_at<=now` conditional → exactly one replica per interval, no separate leader election. Wired in `StartWorker` (3rd loop) with two tasks: cross-tenant idempotency sweep + per-tenant workflow SLA sweep. Lag surfaced via `OnRun` hook (logged; wireable to a metric). |

Tests (`kernel/jobs/scheduler_test.go`): `TestIntegrationSchedulerLeaderSafe` — 6 concurrent replicas
ticking one due task run it exactly once; `TestIntegrationSchedulerRunsDueTask` — runs when due, does not
re-run after the claim advances `next_run_at`. The maintenance wiring (`app/maintenance.go`) is glue over
already-integration-tested pieces (`IdemStore.SweepExpired`, `Runtime.SweepSLA`).

## O2 — migration safety harness

`database.MigrateReset` (goose Down-to-0) + `TestIntegrationMigrationsReversible` run the full
forward→down→forward cycle on an isolated DB in `make ci-container`. **The drill caught a real defect:**
migration 00010 created `app_actor_id()` but its Down did not drop it, so re-apply failed
("function already exists") — fixed in the 00010 Down. Expand/contract guidance: `docs/operations/migrations.md`.

## O5 — backup/restore

`scripts/backup_restore_drill.sh` proves the dump→restore round-trip against a seeded instance (verified
locally: 43 tables restored, marker row intact; the verify step is authoritative over non-fatal
pg_restore version-skew warnings). Runbook: `docs/operations/backup-restore.md` (PITR + object-store
restore order).

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
