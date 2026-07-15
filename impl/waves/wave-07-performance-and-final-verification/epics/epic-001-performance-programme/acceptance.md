---
id: W07-E01-ACCEPTANCE
type: epic-acceptance
epic: W07-E01
wave: W07
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-14
---

# W07-E01 — Epic-level acceptance

Reproduces and elaborates `epic.md`'s "Epic acceptance criteria" section as a standalone,
independently-referenceable acceptance record, consistent with the wave-level `../../acceptance.md`
pattern (AC-W07-01 there maps onto this epic).

## AC-W07-E01-01 — PERF-02 relative evidence published; absolute SLOs conditional

Traces to W07-E01-S001, restated per `epic.md`.

**Status: satisfied.** `EV-W07-E01-S001-002` proves the six named profiles use real PostgreSQL;
`EV-W07-E01-S001-003` proves all 36 cold/warm × 1/10/100-tenant cells; `EV-W07-E01-S001-004`
proves all six attribution components in every cell; and `EV-W07-E01-S001-005` publishes the
relative/container result with `absolute_slo_status=conditional-on-DEC-Q9`.

## AC-W07-E01-02 — PERF-03 set-based query correctness proven

Traces to W07-E01-S002, restated per `epic.md`.

**Status: satisfied.** `EV-W07-E01-S002-002` proves exact precedence parity,
`EV-W07-E01-S002-003/-004` prove current and historical index access with four committed
`EXPLAIN` fixtures, `EV-W07-E01-S002-005` proves constant 8/8/8 total SQL count and one
rules-resolution statement at depths 3/10/50, and `EV-W07-E01-S002-006` proves live updates remain
visible on the next request.

## AC-W07-E01-03 — PERF-04 bounded-batch and leased-outbox correctness proven

Traces to W07-E01-S003, restated per `epic.md`.

**Status: satisfied.** `EV-W07-E01-S003-001/-002/-004` prove the 100-row materialization ceiling,
fixed statements, batched loads, and one endpoint query across the 10/1,000/100,000-row and
multi-endpoint cases. `EV-W07-E01-S003-005` proves the outbox claim commits before tenant handlers,
lease generations fence duplicate workers, inherited W04 chaos passes, and aggregate ordering holds.

## AC-W07-E01-04 — PERF-05 checksum-required behavior proven

Traces to W07-E01-S004, restated per `epic.md`.

**Status: satisfied.** `EV-W07-E01-S004-001/-002` prove canonical upload metadata, zero body
downloads on normal `Stat`, and fallback access only through a labeled byte/time-bounded repair.
`EV-W07-E01-S004-004` proves cursor-based interrupt/resume and idempotent restart without duplicate
repair work. `EV-W07-E01-S004-005` preserves the conditional DEC-Q9 publication boundary.

## AC-W07-E01-05 — CS-16 bench-coverage expansion complete

Traces to W07-E01-S004, restated per `epic.md`.

**Status: satisfied.** `EV-W07-E01-S004-006/-007` prove the exact seven hot paths execute and
`make bench-budget` exits 0. Current `Makefile` `BENCH_PKGS` and `bench-budgets.txt` contain
`kernel/database`, `jobs`, `outbox`, `workflow`, `auth`, `mfa`, and `httpclient`, each with its
corresponding benchmark and budget.

## AC-W07-E01-06 — Independent review passed, DEC-Q9 conditionality genuinely honored

Traces to all four stories, restated per `epic.md`.

**Status: satisfied.** S001 reviewer `W05ReviewGateFinal`, S002 reviewer
`W05ReviewGateFinal`, S003 reviewer `W07-Scoping-Dispatch.W07E01S003ReviewR`, and S004 reviewer
`W07-Scoping-Dispatch.W07E01S004ReviewR` each reported no open story-scope issue. Fresh epic reviewer
`W05ReviewGateRerun`, which authored no W07 change, then audited the aggregate closure and explicitly
reported no open findings. Every publication and closure keeps absolute SLO acceptance conditional on
the still-open DEC-Q9.

## Acceptance authority

Performance/SRE lead, per `../../wave.md`'s wave-level acceptance authority (PLAN §5.5's accountable
role for PF-PERF).
