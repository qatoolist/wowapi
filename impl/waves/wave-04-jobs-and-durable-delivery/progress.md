---
id: W04-PROGRESS
type: wave-progress
wave: W04
status: accepted
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W04 progress (initial state)

Per mandate §16.2. Populated at programme-creation time; every item below is at its initial status.

## Epic status

| Epic | Title | Status | Stories | Story status breakdown |
|---|---|---|---|---|
| W04-E01 | lease-fencing-primitive-and-jobs | planned | 3 | 3 planned |
| W04-E02 | remote-io-outside-tx | planned | 3 | 3 planned |
| W04-E03 | bulk-multi-worker-safety | accepted | 2 | 2 accepted |
| W04-E04 | compliance-and-readiness | planned | 3 | 3 planned |

## Story status

| Story | Title | Status | Task count | Task status breakdown |
|---|---|---|---|---|
| W04-E01-S001 | shared-primitive | planned | 3 | 3 todo (incl. 1 independent-review task) |
| W04-E01-S002 | jobs-lease-and-finalize | planned | 4 | 4 todo (incl. 1 independent-review task) |
| W04-E01-S003 | idempotency-and-chaos | planned | 5 | 5 todo (incl. 1 evidence-aggregation task, 1 independent-review task) |
| W04-E02-S001 | notify-and-webhook-three-stage | planned | 4 | 4 todo (incl. 1 independent-review task) |
| W04-E02-S002 | inbound-two-phase-and-contracts | planned | 6 | 6 todo (incl. 1 evidence-aggregation task, 1 independent-review task) |
| W04-E02-S003 | retry-adoption | planned | 3 | 3 todo (incl. 1 lightweight independent-review task — rationale in its `tasks/index.md`) |
| W04-E03-S001 | stopgap | accepted | 1 | 1 done (no separate independent-review task — judgment documented in its `tasks/index.md`) |
| W04-E03-S002 | leased-claims-and-lifecycle | accepted | 6 | 6 done (incl. 1 independent-review task) |
| W04-E04-S001 | audit-hash-widening | planned | 2 | 2 todo (incl. 1 independent-review task) |
| W04-E04-S002 | anchor-dsr-hold | planned | 5 | 5 todo (incl. 1 independent-review task; no evidence-aggregation task — judgment documented in its `tasks/index.md`) |
| W04-E04-S003 | readiness-truthfulness | planned | 4 | 4 todo (incl. 1 independent-review task) |

## Blocked items

None yet — no story has entered `in-progress`. Note for future readers: W04-E04-S001's tasks are
recorded as gated on W02-E01's acceptance in `story.md`'s own `depends_on` and in
`epics/epic-004-compliance-and-readiness/dependencies.md` — this is a planned cross-wave
dependency, not a blocked item, until E04-S001 actually reaches `in-progress` without W02-E01
having accepted first.

## Critical dependencies

- W04-E04-S001 (DATA-08 W6-T1) depends on W02-E01 (DATA-09 online-migration protocol) — the audit
  hash-widening migration is expected to ship via that protocol, per
  `impl/analysis/wave-allocation-detail.md` and `impl/index.md`'s wave map.
- W04-E01-S001 (DATA-02 T1, shared primitive) supersedes W02-E01-S002's interim checkpoint lease —
  a planned, recorded supersession, not a silent replacement. See `risks.md` RISK-W04-001.
- W04-E02-S001/S002/S003 and W04-E03-S002 all reuse the chaos-test harness built by W04-E01-S003,
  per `wave-allocation-detail.md`'s explicit note ("harness shared with E02/E03") — they do not
  reimplement it.
- W04-E04-S003 (DX-07) T4 is deferred-linked to W05-E03-S002's AR-04 T5 waiver mechanism, which does
  not exist yet — recorded as an explicit forward reference, not implemented here.

## Open decisions

D-04 (audit `hash_version` discriminator) is enacted inside W04-E04-S001 — referenced from the
already-ratified ADR at `impl/waves/wave-00-baseline-and-verification/epics/epic-002-baseline-
capture/stories/story-003-adr-ification/`, not re-authored. No other W04 story enacts a D-0N
decision — confirmed by scanning `requirement-inventory.md` §B for any D-0N row targeting DATA-02,
DATA-03, DATA-04, or DX-07; none exists (see `wave.md` "Assumptions").

## Open risks

See `risks.md`.

## Artifact completeness

2/11 story-level artifact sets populated (W04-E03-S001, W04-E03-S002).

## Evidence completeness

8 evidence records registered (W04-E03-S001: 2; W04-E03-S002: 6).

## Review state

W04-E03 reviewed and accepted internally (review folded into task completion per tasks/index.md).

## Exit-gate readiness

Partial. 2 of 11 stories accepted (W04-E03-S001, W04-E03-S002).
