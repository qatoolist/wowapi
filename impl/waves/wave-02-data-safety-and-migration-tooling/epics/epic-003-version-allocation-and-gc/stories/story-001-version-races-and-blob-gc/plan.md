---
id: PLAN-W02-E03-S001
type: plan
parent_story: W02-E03-S001
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
---

# Plan — W02-E03-S001

Per mandate §8.5. Confirmed facts, planned changes, and implementation assumptions are distinguished
explicitly below. This plan does not invent precise code changes where the repository does not yet
provide enough information — the exact counter mechanism, upload-session schema, and GC scheduling
approach are recorded as unresolved questions to be answered at implementation time, not invented
here.

## Source task table (PLAN DATA-05, reproduced verbatim for traceability)

| Task | Depends-on | Acceptance criteria | Tests | Evidence | Risk |
|---|---|---|---|---|---|
| T1. Replace `MAX()+1` with a locked parent counter or dedicated per-aggregate sequence row (both `kernel/artifact` and `kernel/document`) | — | N concurrent callers → N unique monotonic versions, zero unexpected conflicts | Concurrency test, ≥20 concurrent callers | `DATA-05/version-allocation/` | Counter-row contention is the new serialization point — measure lock wait |
| T2. Durable upload-session records for `kernel/document`: expiry, checksum/size, storage key, status, cleanup ownership | T1 | Session row persisted before the presigned URL returns | Test: initiate, simulate crash, assert `status='pending'` with expiry | `DATA-05/upload-session/` | New table needs RLS + `<module>_<entity>` naming |
| T3. Confirmation CASes the session and version atomically | T1, T2 | Two racing confirms: exactly one succeeds | Concurrency test | `DATA-05/confirm-cas/` | — |
| T4. Scheduled GC removing expired/unreferenced objects, with metrics/audit | T2, T3 | Never removes a referenced object; removes every past-expiry unconfirmed session | Mixed confirmed/expired/pending test | `DATA-05/gc-sweep/` | False-positive deletion is data loss — conservative grace window |
| T5. Same counter fix for `kernel/artifact.Generate` | T1 | Same concurrency bar as T1 | Mirror test | `DATA-05/artifact-version/` | — |

## Proposed architecture

A version-allocation mechanism (a locked parent-row read via `SELECT ... FOR UPDATE`, a PostgreSQL
`SEQUENCE`, or a dedicated per-aggregate counter table — see "Unresolved questions") that both
`kernel/artifact.Generate` and `kernel/document.InitiateUpload` call in place of their current inline
`MAX(version)+1` read, serializing concurrent version reads instead of racing on an unlocked read. A
new durable upload-session record type for `kernel/document`, persisted before a presigned upload URL
is ever returned to a caller, carrying enough state (expiry, checksum/size, storage key, status,
cleanup ownership) for both the confirmation path and the GC sweep to operate on without any other
source of truth. A confirmation path that CASes the session's status and the version allocation
together in one atomic operation, so a losing racer's confirmation is rejected outright rather than
partially applied. A scheduled GC sweep that reads session records only (never touches storage
directly without a session record backing the decision) and removes objects whose session is
past-expiry and unconfirmed, governed by a conservative grace window.

## Implementation strategy

1. Re-read `kernel/artifact.Generate` and `kernel/document.InitiateUpload` fresh at this story's
   actual start commit to confirm the current-state assessment (inline `MAX()+1`, no session
   durability, no GC) still holds.
2. Design the counter/sequence mechanism: draft options (locked parent-row read, PostgreSQL
   `SEQUENCE`, dedicated per-aggregate counter table) with trade-offs — particularly around lock-wait
   under concurrent load (RISK-W02-E03-001) — and select one, documenting the rationale.
3. Implement the chosen mechanism in `kernel/artifact.Generate`, replacing the inline `MAX()+1` read.
4. Write T1's concurrency test (≥20 concurrent callers) against `kernel/artifact.Generate` as the
   first proof of the mechanism, and measure lock wait as part of that test's evidence.
5. Implement the same mechanism in `kernel/document.InitiateUpload`, replacing its own inline
   `MAX()+1` read.
6. Extend T1's concurrency test (or add a parallel one) to also cover
   `kernel/document.InitiateUpload`, satisfying T1's full "both `kernel/artifact` and
   `kernel/document`" acceptance bar.
7. Design the upload-session schema (expiry, checksum/size, storage key, status, cleanup ownership),
   following this programme's `<module>_<entity>` table-naming convention and RLS requirements per
   PLAN T2's own risk note.
8. Implement session persistence in `kernel/document.InitiateUpload`, ensuring the session row is
   written before the presigned URL is constructed and returned.
9. Write T2's crash-simulation test: initiate an upload, simulate a crash before confirmation,
   assert the session row exists with `status='pending'` and a set expiry.
10. Implement the atomic CAS confirmation path: confirmation reads the session, checks status, and
    updates the session's status and the version allocation together in one transaction/CAS
    operation, so a losing racer's confirmation attempt is rejected, not partially applied.
11. Write T3's concurrency test: two racing confirmation calls against the same session, confirming
    exactly one succeeds.
12. Design and implement the scheduled GC sweep: reads session records for past-expiry, unconfirmed
    sessions within a conservative grace window, removes the corresponding storage object, and emits
    metrics/audit for each action.
13. Write T4's mixed-state test: a set of sessions in confirmed, expired-unconfirmed, and
    still-pending states, confirming the sweep removes only the expired-unconfirmed set's objects and
    never touches a confirmed or still-pending session's object.
14. Write T5's dedicated mirror test for `kernel/artifact.Generate`, confirming it independently meets
    the same concurrency bar as T1 (not merely relying on T1's own test coverage).
15. Document the counter mechanism, session lifecycle, and GC sweep.

## Expected package or module changes

`kernel/artifact` (version-allocation path in `Generate`); `kernel/document` (version-allocation path
in `InitiateUpload`, plus new session-persistence, confirmation, and GC surfaces). A new schema
migration for the counter/sequence mechanism and the upload-session table (exact package/file
locations TBD — see "Unresolved questions").

## Expected file changes where determinable

- `kernel/artifact`'s `Generate` version-allocation code path (exact file/line not yet confirmed by
  this plan — to be re-confirmed at this story's actual start commit per "Current-state assessment"
  re-confirmation step).
- `kernel/document`'s `InitiateUpload` version-allocation code path (same caveat).
- A new schema migration introducing the counter/sequence mechanism and the upload-session table.
- New confirmation-path code implementing the atomic CAS.
- A new GC sweep mechanism (scheduled job or command, exact location TBD).

## Contracts and interfaces

A version-allocation interface/function shared (or independently implemented in parallel, per
implementation-time decision) by `kernel/artifact` and `kernel/document`, replacing each package's own
inline `MAX()+1` call. An upload-session data contract (expiry, checksum/size, storage key, status,
cleanup ownership). A confirmation contract that accepts a session identifier and atomically
transitions its status while allocating/confirming the version.

## Data structures

The upload-session record (new). The counter/sequence-row or `SEQUENCE` object (new). No change to
the existing artifact/document version-history data model beyond how the next version number is
computed.

## APIs

`kernel/document.InitiateUpload`'s external contract is expected to remain stable from the caller's
perspective (it still returns a presigned upload URL); the internal addition of durable session
persistence before URL issuance is not expected to be a breaking API change, but this is an
implementation-time confirmation, not yet verified against the actual current signature.

## Configuration changes

The GC sweep's grace window and scheduling interval may be hardcoded constants or configuration keys
— to be determined at implementation time (see "Unresolved questions").

## Persistence changes

New: the counter/sequence mechanism (a new table, or a PostgreSQL `SEQUENCE` object, per the chosen
design) and the upload-session table. Both are additive; no existing table's schema is altered by
this story per its own "Compatibility considerations" in `story.md`.

## Migration strategy

New, additive schema migrations for the counter/sequence mechanism and the upload-session table.
Whether these are authored through DATA-09's online-migration protocol (once W02-E01 exists) or as
conventional forward-apply migrations is an implementation-time choice — this story has no dependency
on DATA-09 per `dependencies.md`, so it is not blocked either way.

## Concurrency implications

This is the central concern of the entire story. T1/T5's concurrency tests (≥20 concurrent callers)
must both prove correctness (N callers → N unique monotonic versions, zero unexpected conflicts) and
measure lock wait under the counter/sequence mechanism, per RISK-W02-E03-001. T3's concurrency test
must prove the confirmation CAS correctly rejects a losing racer rather than partially applying a
conflicting state.

## Error-handling strategy

A losing racer's confirmation attempt must fail cleanly (rejected by the CAS), not partially update
the session or version. The GC sweep must fail closed on any ambiguity about a session's true state —
per PLAN T4's own risk note, a false-positive deletion is data loss, so the sweep's error-handling
default must be "do not delete" when the session's state cannot be confidently determined as
past-expiry-and-unconfirmed.

## Security controls

RLS on the new upload-session table, per PLAN T2's own risk note: "New table needs RLS +
`<module>_<entity>` naming." The GC sweep's conservative grace window is itself a required control
against data loss, per PLAN T4's own risk note.

## Observability changes

The GC sweep emits metrics and an audit record for each sweep action (removal or explicit no-op),
per PLAN T4's own acceptance framing ("with metrics/audit").

## Testing strategy

- T1: concurrency test, ≥20 concurrent callers, against both `kernel/artifact.Generate` and
  `kernel/document.InitiateUpload`, proving zero unexpected conflicts and measuring lock wait.
- T2: crash-simulation test — initiate an upload, simulate a crash, assert session row exists with
  `status='pending'` and a set expiry.
- T3: concurrency test — two racing confirmation calls against the same session, exactly one
  succeeds.
- T4: mixed-state test — confirmed, expired-unconfirmed, and still-pending sessions in one test run,
  proving the sweep removes only the expired-unconfirmed set.
- T5: a dedicated mirror test for `kernel/artifact.Generate`, independent of T1's own test, at the
  same concurrency bar.

## Regression strategy

T1/T5's concurrency tests, once passing, become the ongoing regression guard against a future
reintroduction of an unsafe `MAX()+1` read in either package. T4's mixed-state test guards against a
future GC sweep regression that would remove a referenced or still-pending object.

## Compatibility strategy

The version-numbering contract (monotonically increasing per aggregate) is preserved from the
caller's perspective. No source guidance suggests a compatibility transition period is needed for
this story's change, unlike W02-E01-S001's CI-gate enforcement-timing question — this story's change
is internal to the allocation mechanism, not a new external contract callers must adapt to.

## Rollout strategy

Single story, landed as its own reviewable unit. T1–T5 are sequenced per their own dependency chain
(T2→T1; T3→T1,T2; T4→T2,T3; T5→T1), so implementation proceeds T1, T5 (parallel-safe once T1 lands),
T2, T3, T4 — matching `tasks/index.md`'s task ordering.

## Rollback strategy

Revert the counter/sequence mechanism and confirmation-CAS logic if either produces false-positive
allocation failures or false-positive confirmation rejections against legitimate, non-racing callers
under normal load. Revert (disable) the GC sweep immediately, without reverting the session-durability
mechanism itself, if any false-positive deletion is ever observed — per PLAN T4's own risk framing,
this is the single most consequential failure mode in this story and the rollback path for it must be
fast and independent of the other four tasks' own code.

## Implementation sequence

As listed under "Implementation strategy" above (steps 1–15), respecting T1→T5 (parallel-safe after
T1), T1+T2→T3, T2+T3→T4.

## Task breakdown

- **W02-E03-S001-T001** — Locked-counter/sequence-row version-allocation mechanism (steps 2–6 above,
  covering both `kernel/artifact` and `kernel/document`).
- **W02-E03-S001-T002** — Durable upload-session records for `kernel/document` (steps 7–9 above).
- **W02-E03-S001-T003** — Atomic CAS confirmation (steps 10–11 above).
- **W02-E03-S001-T004** — Scheduled GC sweep with metrics/audit (steps 12–13 above).
- **W02-E03-S001-T005** — Mirrored counter fix and dedicated concurrency test for
  `kernel/artifact.Generate` (step 14 above, largely completed as part of steps 3–4 but tracked as
  its own task per PLAN's own T5 row and its own dedicated acceptance criterion — see "Task sizing
  judgment" below).

## Task sizing judgment

DATA-05 is P1 per `requirement-inventory.md` ("Version allocation races + blob GC (T1–T5) | IMPL |
P1 | planned | W02-E03-S001 |"). Per this programme's convention (established in W02-E01-S001's
`tasks/index.md`), only P0 stories automatically receive a dedicated independent-review task; for a
P1 story, the default — consistent with wave-01's parsimony principle and mandate §12's warning
against "excessive fragmentation into trivial tasks that provide no tracking value" — is no separate
review task unless the task count or risk genuinely warrants one. This story's five tasks (T1–T5) are
a single, tightly-coupled persistence-correctness fix reviewed coherently as one unit (per
`epic.md`'s "single reviewer domain" framing, drawn from
`impl/analysis/wave-allocation-detail.md`'s own allocation note); no task introduces a risk profile
materially different from what T1–T5's own acceptance criteria and RISK-W02-E03-001 already capture.
Default applied: **5 tasks only, no separate independent-review task.** See
`tasks/index.md` "Grouping rationale" for the full statement of this judgment.

## Expected artifacts

The locked-counter/sequence-row mechanism (code, both packages); the upload-session schema and table;
the atomic CAS confirmation logic; the GC sweep mechanism; documentation of the counter mechanism,
session lifecycle, and GC sweep.

## Expected evidence

Concurrency test output (≥20 concurrent callers) for `kernel/document`'s version allocation;
crash-simulation test output for session durability; racing-confirmation CAS test output; mixed
confirmed/expired/pending GC sweep test output; the dedicated `kernel/artifact.Generate` mirror
concurrency test output.

## Unresolved questions

- Exact form of the counter/sequence mechanism (locked parent-row read, PostgreSQL `SEQUENCE`, or
  dedicated per-aggregate counter table) — to be decided at implementation time, with lock-wait
  measured as part of T1's own evidence per RISK-W02-E03-001.
- Exact upload-session table schema (column names, types) — to be decided at implementation time,
  following this programme's `<module>_<entity>` naming convention and RLS requirements.
- Exact GC sweep scheduling mechanism (cron-style scheduled job, periodic background worker, or
  manually-triggered operational command) — not specified by the source beyond "Scheduled GC."
- Exact GC sweep grace window value — PLAN T4's own risk note calls for "a conservative grace window"
  without stating a figure; the chosen value and rationale must be recorded at implementation time,
  not invented here.
- Whether the counter/sequence mechanism is implemented once and shared by both packages, or
  implemented independently in each — an implementation-time design choice with no source guidance
  either way.

## Approval conditions

This plan is approved for implementation once: (a) the counter/sequence mechanism's exact form is
selected (with lock-wait measurement built into T1's test plan), and (b) the owner and reviewer are
assigned.
