---
id: W04-EPICS-INDEX
type: epics-index
wave: W04
status: accepted
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W04 — Epics index

| Epic | Title | Status | Stories | Objective |
|---|---|---|---|---|
| [W04-E01](epic-001-lease-fencing-primitive-and-jobs/epic.md) | lease-fencing-primitive-and-jobs | planned | 3 | Build the DATA-02 shared lease/fencing primitive (this wave's keystone build) and apply it to the jobs queue's claim, finalize, and reclaim paths; establish the idempotency-declaration contract and the chaos harness shared with E02/E03 |
| [W04-E02](epic-002-remote-io-outside-tx/epic.md) | remote-io-outside-tx | planned | 3 | Move `kernel/notify`/`kernel/webhook` remote I/O outside database transactions via a three-stage claim→effect→finalize protocol (DATA-03); adopt `cenkalti/backoff/v5` for retry (FBL-04) |
| [W04-E03](epic-003-bulk-multi-worker-safety/epic.md) | bulk-multi-worker-safety | planned | 2 | Correct the false "replica-safe" migration claim and make bulk multi-worker processing actually safe via a leased, `SKIP LOCKED`-honest claim path (DATA-04) |
| [W04-E04](epic-004-compliance-and-readiness/epic.md) | compliance-and-readiness | planned | 3 | Widen the audit hash chain to cover every persisted field (DATA-08 Wave-6, enacting D-04); make readiness and configuration diagnostics truthful (DX-07 T1-T3; T4 deferred-linked to W05) |
