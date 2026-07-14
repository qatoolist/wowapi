---
id: TRACK-STATUS-REGISTER
type: register
title: Status register — every wave/epic/story target, planning-time snapshot
status: active
created_at: 2026-07-12
updated_at: 2026-07-14
derived: true
---

# Status register

**DERIVED VIEW — planning-time snapshot refreshed 2026-07-14.** Canonical status now lives in
each wave/epic/story's own front matter (`wave.md`, `epic.md`, `story.md`) per mandate §6; this
register is a generated roll-up and must be regenerated when those canonical statuses change. This
register does not itself set status — it reflects the canonical front-matter values.

Item IDs below are wave-epic-story **target strings** extracted verbatim from
`requirement-inventory.md`'s Target column across tables A, B, and C. Where an epic/story does not
yet have a descriptive title (no `waves/` directory exists), Title is marked
"TBD — see requirement-inventory.md row `<ID>`".

## W00 — baseline-and-verification

Objective: Pin verification of the 8 executed finding-slices at current HEAD; capture
coverage/lint/bench/CI baselines; ADR-ify D-01..09. Depends on: — (entry wave).

**Note (2026-07-13):** all six executed W00 stories are `accepted` as of 2026-07-13 —
W00-E01-S001, W00-E01-S002, W00-E01-S003, W00-E02-S001, W00-E02-S002, W00-E02-S003 (independent
review gate W00ReviewGate passed; conductor concurs). This register only lists the two W00
targets extracted from `requirement-inventory.md`; the other four stories' canonical status lives
in their `story.md` front matter.

| Item ID | Level | Title | Status | Source requirement(s) |
|---|---|---|---|---|
| W00-E01-S002 | story | TBD — see requirement-inventory.md rows PERF-01, PERF-06 | accepted (2026-07-13) | PERF-01 (INV), PERF-06 (INV) |
| W00-E02-S003 | story | ADR-ification of ratified decisions | accepted (2026-07-13) | D-01..D-09 |

## W01 — zero-dependency-hardening

Objective: Everything valuable with no upstream dependency: linter utilisation (FBL-05/07), OTel
correlation (FBL-06), HTTP hardening (FBL-08/09), generator+doc+test fixes (DX-01/02, T-DOC-01,
T-TEST-01, FBL-03). Depends on: W00.

**Note (2026-07-13):** all ten W01 stories are `accepted` as of 2026-07-13 (independent
review gate W01ReviewGate passed; conductor concurs; conductor applied the reviewer's fix list).
Canonical status lives in each story's `story.md` front matter.

| Item ID | Level | Title | Status | Source requirement(s) |
|---|---|---|---|---|
| W01-E01-S001 | story | TBD — see requirement-inventory.md row FBL-05 | accepted (2026-07-13) | FBL-05 |
| W01-E01-S002 | story | TBD — see requirement-inventory.md row FBL-07 | accepted (2026-07-13) | FBL-07 (partial) |
| W01-E01-S003 | story | TBD — see requirement-inventory.md row FBL-07 | accepted (2026-07-13) | FBL-07 (partial) |
| W01-E02-S001 | story | TBD — see requirement-inventory.md row FBL-06 | accepted (2026-07-13) | FBL-06 |
| W01-E02-S002 | story | TBD — see requirement-inventory.md row FBL-06 | accepted (2026-07-13) | FBL-06 |
| W01-E03-S001 | story | TBD — see requirement-inventory.md row FBL-09 | accepted (2026-07-13) | FBL-09 |
| W01-E03-S002 | story | TBD — see requirement-inventory.md row FBL-08 | accepted (2026-07-13) | FBL-08 |
| W01-E04-S001 | story | TBD — see requirement-inventory.md rows DX-01, DX-02 | accepted (2026-07-13) | DX-01, DX-02 |
| W01-E04-S002 | story | TBD — see requirement-inventory.md rows DX-05, FBL-03, T-DOC-01 | accepted (2026-07-13) | DX-05 (partial), FBL-03, T-DOC-01 |
| W01-E04-S003 | story | TBD — see requirement-inventory.md row T-TEST-01 | accepted (2026-07-13) | T-TEST-01 |

## W02 — data-safety-and-migration-tooling

Objective: DATA-09 online-migration protocol, then DATA-01 tenant FKs over it; DATA-05/06;
FBL-02 prod seed-sync. Depends on: W00.

| Item ID | Level | Title | Status | Source requirement(s) |
|---|---|---|---|---|
| W02-E01-S001 | story | TBD — see requirement-inventory.md row DATA-09 | planned | DATA-09 |
| W02-E01-S002 | story | TBD — see requirement-inventory.md row DATA-09 | planned | DATA-09 |
| W02-E01-S003 | story | TBD — see requirement-inventory.md row DATA-09 | planned | DATA-09 |
| W02-E02-S001 | story | TBD — see requirement-inventory.md row DATA-01 | planned | DATA-01 |
| W02-E02-S002 | story | TBD — see requirement-inventory.md row DATA-01 | planned | DATA-01 |
| W02-E03-S001 | story | TBD — see requirement-inventory.md row DATA-05 | planned | DATA-05 |
| W02-E04-S001 | story | TBD — see requirement-inventory.md row DATA-06 | planned | DATA-06 |
| W02-E05-S001 | story | TBD — see requirement-inventory.md row FBL-02 | planned | FBL-02 |

## W03 — identity-and-session-security

Objective: SEC-01 server-side session state (+D-01, DEC-Q1 safe default), SEC-06, SEC-03,
DATA-07 (dep SEC-01), SEC-02 remainder. Depends on: W01 (validation seam), W02 (grant-table
migration uses DATA-09).

**Note (2026-07-13):** all eight W03 stories are `ready` as of 2026-07-13 — story-level `story.md`
files and `plan.md` files are complete, task files exist under each story's `tasks/` directory, and
`progress.md` has been updated. W03-E04-S001 remains logically gated on W03-E01's acceptance; it is
not marked `blocked` at this planning stage.

| Item ID | Level | Title | Status | Source requirement(s) |
|---|---|---|---|---|
| W03-E01-S001 | story | TBD — see requirement-inventory.md row SEC-01 | ready | SEC-01 |
| W03-E01-S002 | story | TBD — see requirement-inventory.md row SEC-01 | ready | SEC-01 |
| W03-E01-S003 | story | TBD — see requirement-inventory.md row SEC-01 | ready | SEC-01 |
| W03-E01-S004 | story | TBD — see requirement-inventory.md row SEC-01 | ready | SEC-01 |
| W03-E02-S001 | story | TBD — see requirement-inventory.md row SEC-06 | ready | SEC-06 |
| W03-E03-S001 | story | TBD — see requirement-inventory.md row SEC-03 | ready | SEC-03 |
| W03-E04-S001 | story | TBD — see requirement-inventory.md row DATA-07 | ready | DATA-07 (blocked→planned) |
| W03-E05-S001 | story | TBD — see requirement-inventory.md row SEC-02 | ready | SEC-02 (partial) |

## W04 — jobs-and-durable-delivery

Objective: Shared lease/fencing primitive → DATA-02/03/04; FBL-04 retry adoption; DATA-08 W6
audit integrity (D-04); DX-07 readiness truthfulness. Depends on: W02 (DATA-09 for W6-T1
migration).

| Item ID | Level | Title | Status | Source requirement(s) |
|---|---|---|---|---|
| W04-E01-S001 | story | TBD — see requirement-inventory.md row DATA-02 | planned | DATA-02 |
| W04-E01-S002 | story | TBD — see requirement-inventory.md row DATA-02 | planned | DATA-02 |
| W04-E01-S003 | story | TBD — see requirement-inventory.md row DATA-02 | planned | DATA-02 |
| W04-E02-S001 | story | TBD — see requirement-inventory.md row DATA-03 | planned | DATA-03 |
| W04-E02-S002 | story | TBD — see requirement-inventory.md row DATA-03 | planned | DATA-03 |
| W04-E02-S003 | story | TBD — see requirement-inventory.md row FBL-04 | planned | FBL-04 |
| W04-E03-S001 | story | TBD — see requirement-inventory.md row DATA-04 | planned | DATA-04 |
| W04-E04-S001 | story | TBD — see requirement-inventory.md row DATA-08 | planned | DATA-08 (partial) |
| W04-E04-S002 | story | TBD — see requirement-inventory.md row DATA-08 | planned | DATA-08 (partial) |
| W04-E04-S003 | story | TBD — see requirement-inventory.md row DX-07 | planned | DX-07 |

## W05 — application-model-and-layering

Objective: AR-01/02 ownership model (+D-02/03), AR-03/AR-04 remainder, AR-06 remainder, SEC-04
cache (+D-06), FBL-01 kernel re-home. Depends on: W03 (SEC-01 actor model stabilises registrar
security assumptions).

| Item ID | Level | Title | Status | Source requirement(s) |
|---|---|---|---|---|
| W05-E01-S001 | story | TBD — see requirement-inventory.md row AR-01 | planned | AR-01 |
| W05-E01-S002 | story | TBD — see requirement-inventory.md row AR-01 | planned | AR-01 |
| W05-E01-S003 | story | TBD — see requirement-inventory.md row AR-01 | planned | AR-01 |
| W05-E01-S004 | story | TBD — see requirement-inventory.md row AR-01 | planned | AR-01 |
| W05-E02-S001 | story | TBD — see requirement-inventory.md row AR-02 | planned | AR-02 |
| W05-E02-S002 | story | TBD — see requirement-inventory.md row AR-02 | planned | AR-02 |
| W05-E02-S003 | story | TBD — see requirement-inventory.md row AR-02 | planned | AR-02 |
| W05-E03-S001 | story | TBD — see requirement-inventory.md row AR-03 | planned | AR-03 |
| W05-E03-S002 | story | TBD — see requirement-inventory.md row AR-03 | planned | AR-03 |
| W05-E03-S002 | story | TBD — see requirement-inventory.md row AR-04 | planned | AR-04 (partial) |
| W05-E04-S001 | story | TBD — see requirement-inventory.md row AR-06 | planned | AR-06 (partial) |
| W05-E04-S002 | story | TBD — see requirement-inventory.md row SEC-04 | planned | SEC-04 |
| W05-E05-S001 | story | TBD — see requirement-inventory.md row FBL-01 | planned | FBL-01 |
| W05-E05-S002 | story | TBD — see requirement-inventory.md row FBL-01 | planned | FBL-01 |

## W06 — contracts-compatibility-release

Objective: DX-03 design + DX-04 golden consumer; DX-06 merge + REL-03a/b diff gates; REL-01/REL-02
release gating (DEC-Q10 activation); doc-example gates (CS-22/AR-05). Depends on: W05 (AR-03
unblocks REL-03b legs).

| Item ID | Level | Title | Status | Source requirement(s) |
|---|---|---|---|---|
| W06-E01-S001 | story | TBD — see requirement-inventory.md row DX-03 | planned | DX-03 (deferred-design) |
| W06-E01-S002 | story | Golden consumer matrix — installed CLI, eight subsystems, real infrastructure, N-1/candidate replay, required CI gate | accepted (2026-07-14; EV-W06-E01-S002-014) | DX-04 |
| W06-E02-S001 | story | TBD — see requirement-inventory.md row DX-06 | planned | DX-06 |
| W06-E02-S002 | story | Compatibility gates buildable now | accepted (2026-07-14) | REL-03 (a-leg) |
| W06-E02-S003 | story | TBD — see requirement-inventory.md row REL-03 | planned | REL-03 (b-leg) |
| W06-E03-S001 | story | TBD — see requirement-inventory.md row REL-01 | planned | REL-01 |
| W06-E03-S002 | story | TBD — see requirement-inventory.md row REL-01 | planned | REL-01 |
| W06-E03-S003 | story | TBD — see requirement-inventory.md row REL-02 | planned | REL-02 |
| W06-E04-S002 | story | TBD — see requirement-inventory.md row AR-05 | planned | AR-05 (partial) |

## W07 — performance-and-final-verification

Objective: PERF-02..05 relative programme (+DEC-Q9), SEC-05 profile, REL-04 remainder (real
fuzz), product-alignment verification, programme closure gate. Depends on: all prior.

| Item ID | Level | Title | Status | Source requirement(s) |
|---|---|---|---|---|
| W07-E01-S001 | story | Request benchmarks against real PostgreSQL | accepted | PERF-02 |
| W07-E01-S002 | story | Rules resolution collapsed to bounded SQL | accepted | PERF-03 |
| W07-E01-S003 | story | Sweeper and worker bounded materialization | accepted | PERF-04 |
| W07-E01-S004 | story | Checksum behavior and seven-package bench coverage | accepted | PERF-05, CS-16 |
| W07-E02-S001 | story | TBD — see requirement-inventory.md row SEC-05 | planned | SEC-05 |
| W07-E02-S002 | story | TBD — see requirement-inventory.md rows REL-04, PERF-06 | planned | REL-04 (partial), PERF-06 (T3/T4 shared) |

## Tracked-only items (no dedicated wave-epic-story target yet)

The three open human decisions are tracked against their wave but have no story-level target of
their own (they gate stories, they are not stories):

| Item ID | Level | Title | Status | Source requirement(s) |
|---|---|---|---|---|
| W03-E01 (tracked) | decision | DEC-Q1 — IdP grant_id claim contract | blocked (human) | DEC-Q1 |
| W07-E01 (tracked) | decision | DEC-Q9 — reference-perf-env ownership | blocked (human) | DEC-Q9 |
| W06-E03 (tracked) | decision | DEC-Q10 — repo-admin activation | blocked (human) | DEC-Q10 |

## Summary counts

8 waves · 34 distinct story-level targets extracted from tables A/B · 2 decision-tracked items
(status blocked-human, not planned) · W00 stories `accepted` · W03 stories `ready` · all other
story-level items status=`planned`. Matrix-only rows
(table C: CS-03, CS-19, CS-24, CS-10, CS-25, K-RETAIN, K-P2, M-REJ, B11/B12/B13) carry no wave
target and are excluded here — see `deferred-items-register.md` and
`docs/implementation/framework-backlog-p2-decisions.md` for their disposition.
