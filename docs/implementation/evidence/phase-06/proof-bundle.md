# Phase 6 — Proof Bundle

Scope (phase-plan row 6): transactional outbox + relay/inbox, Postgres job runner (retries/DLQ),
worker process, migration 00007. Date: 2026-07-03.

## 1. Decision evidence
D-0047 (Postgres job runner, not River), D-0048 (relay reads cross-tenant as app_platform,
dispatches per-tenant), D-0049 (TenantDB/Context accessors), D-0050 (per-aggregate ordering + event
DLQ + job timeout/drain separation — review fixes).

## 2. Discussion evidence
- Runner engine choice: River is heavy with its own migrations; the module portability contract
  only depends on the jobs interfaces, so a focused Postgres queue behind them (D-0047) keeps the
  dependency surface small and the retry/DLQ semantics precisely ours.
- The relay's cross-tenant read vs tenant isolation: resolved with a role-scoped RLS policy
  (app_platform reads all, app_rt still isolated) + per-tenant re-entry for dispatch, mirroring the
  Phase 5 app_platform posture.
- Ordering: the review reproduced that per-aggregate order was only best-effort (advisory lock
  absent; retry reordered). Fixed with earliest-per-aggregate claim + advisory lock — now enforced
  and regression-tested, honestly documented (aggregate-less events unordered; external side
  effects at-least-once).

## 3. Critique/review evidence
`review-findings.md`: 7 findings (1 reproduced high ordering bug, 5 mediums on DLQ/cooldown/timeout/
shutdown/reclaim, 1 documented semantics). All fixed with regression tests or documented+enforced;
atomicity and tenant isolation verified solid by the reviewer.

## 4. Implementation evidence
New: `kernel/outbox/` (outbox, relay), `kernel/jobs/` (jobs, registry, runner), `app/worker.go`,
migration 00007. Changed: `module/module.go` + `app/context.go` + `app/boot.go` (Events/Outbox/Jobs
accessors + boot gates), `kernel/kernel.go` (Platform pool), testkit fixtures + worker test.
Team: 1 implementation agent (kernel/jobs) + lead (outbox, relay, migration, worker, all review
fixes); 1 reliability+security review agent.

## 5. Verification evidence
`command-log.md`: outbox integration (atomicity, dispatch, inbox, ordering-under-retry, DLQ,
isolation), jobs integration (enqueue-atomic, worker success/tenant-aware, retry→DLQ, backoff,
reclaim, isolation), the end-to-end worker test (dispatch + job + graceful shutdown), full
`make ci` + `make test-integration` host and in-container. Graphify updated.

## 6. Acceptance evidence
`acceptance-map.md`: all 16 Phase 6 exit criteria mapped; acceptance (atomic events / crash+retry /
tenant-aware jobs / graceful shutdown) each to named tests. Carried forward: optional job
idempotency inbox, claim-tx connection hold + durable event-DLQ admin API (Phase 11). Graphify
`extract` blocked on LLM key (R11).
