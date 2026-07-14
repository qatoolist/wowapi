---
id: W02-E03-S001
type: story
title: Version-allocation races and upload-blob GC
status: accepted
wave: W02
epic: W02-E03
owner: W02FKVerAgg
reviewer: W02ReviewGate
priority: high
created_at: 2026-07-12
updated_at: 2026-07-13
source_requirements:
  - DATA-05
depends_on: []
blocks: []
acceptance_criteria:
  - AC-W02-E03-S001-01
  - AC-W02-E03-S001-02
  - AC-W02-E03-S001-03
  - AC-W02-E03-S001-04
  - AC-W02-E03-S001-05
artifacts:
  - ART-W02-E03-S001-001
  - ART-W02-E03-S001-002
  - ART-W02-E03-S001-003
  - ART-W02-E03-S001-004
  - ART-W02-E03-S001-005
evidence:
  - EV-W02-E03-S001-001
  - EV-W02-E03-S001-002
  - EV-W02-E03-S001-003
  - EV-W02-E03-S001-004
  - EV-W02-E03-S001-005
decisions: []
risks:
  - RISK-W02-E03-001
---

# W02-E03-S001 — Version-allocation races and upload-blob GC

## Story ID

W02-E03-S001

## Title

Version-allocation races and upload-blob GC

## Objective

Replace the racy inline `MAX(version)+1` read in both `kernel/artifact.Generate` and
`kernel/document.InitiateUpload` with a locked parent counter or dedicated per-aggregate sequence
row; make `kernel/document`'s upload-session lifecycle durable, with atomic CAS confirmation of the
session and version together; and add a scheduled GC sweep that reclaims expired, unconfirmed upload
blobs without ever removing a referenced object.

## Value to the framework

Today, per PLAN's own evidence, two independent packages compute their next version number by
reading the current maximum and adding one — a pattern that is not safe under concurrent callers.
For `kernel/artifact`, this is a correctness defect: concurrent generation can produce duplicate or
conflicting version numbers. For `kernel/document`, the same race is compounded by an upload flow
that allocates a storage key and issues a presigned URL before the version race is resolved, so the
losing caller's blob is left in storage with no session record, no expiry, and — until this story —
no process that will ever reclaim it. This story closes both the correctness gap and the storage-
leak gap in one coherent fix, converting an unsafe read-then-write version allocation into a safe,
concurrency-proven mechanism, and converting an untracked, un-reclaimed orphan blob into a durably
tracked, conservatively garbage-collected one.

## Problem statement

`requirement-inventory.md` row DATA-05 states: "Version allocation races + blob GC (T1–T5) | IMPL |
P1 | planned | W02-E03-S001 |" with no dependency recorded in the Notes column. PLAN's own DATA-05
evidence: "`kernel/artifact.Generate` and `kernel/document.InitiateUpload` both compute
`MAX(version)+1` inline; the document path's loser leaves an orphaned, randomly-keyed blob with no
GC." PLAN's own T1–T5 task table (reproduced in full in `plan.md`) specifies: T1 replaces the
`MAX()+1` pattern with a locked counter or sequence row for both packages, tested with ≥20 concurrent
callers; T2 adds durable upload-session records for `kernel/document` (expiry, checksum/size, storage
key, status, cleanup ownership), persisted before the presigned URL returns; T3 makes confirmation
CAS the session and version atomically, so exactly one of two racing confirms succeeds; T4 adds a
scheduled GC sweep removing expired/unreferenced objects with metrics/audit, conservative enough
never to remove a referenced object; T5 applies T1's same counter fix specifically to
`kernel/artifact.Generate`, proven by its own dedicated mirror test.

## Source requirements

DATA-05 (T1, T2, T3, T4, T5) — all five tasks, per `impl/analysis/wave-allocation-detail.md`'s
canonical allocation: "E03 version-allocation-and-gc (DATA-05): S001 all T1–T5 (single reviewer
domain)."

## Current-state assessment

Per PLAN's own evidence for DATA-05 (to be re-confirmed at this story's own execution commit):
`kernel/artifact.Generate` and `kernel/document.InitiateUpload` both compute the next version number
via an inline `MAX(version)+1` read, with no locking or sequence primitive serializing concurrent
callers. `kernel/document`'s upload flow additionally allocates a storage key and returns a presigned
upload URL to the caller without first persisting a durable session record — so a caller that loses
the version race, or that never completes its upload, leaves a blob in storage with no tracking
record and no scheduled reclamation. This story's own re-confirmation step (per this programme's
fail-first convention applied elsewhere, e.g. W01-E01-S001, W02-E01-S001) is to re-read
`kernel/artifact.Generate` and `kernel/document.InitiateUpload` at this story's actual start commit
and confirm both still compute version via inline `MAX()+1` with no session-durability or GC
mechanism, before implementing the fix.

## Desired state

Both `kernel/artifact.Generate` and `kernel/document.InitiateUpload` allocate the next version number
via a locked parent counter or dedicated per-aggregate sequence row, proven race-free under at least
20 concurrent callers with zero unexpected conflicts. `kernel/document`'s upload-session lifecycle
persists a durable session record (expiry, checksum/size, storage key, status, cleanup ownership)
before the presigned URL is returned to the caller. Confirmation CASes the session and version
together atomically, so that of two racing confirmation calls, exactly one succeeds. A scheduled GC
sweep removes every past-expiry unconfirmed session's object and never removes a referenced object,
with metrics and audit recording each sweep's action.

## Scope

- T1 — the locked-counter/sequence-row version-allocation mechanism, applied to both
  `kernel/artifact.Generate` and `kernel/document.InitiateUpload`.
- T2 — durable upload-session records for `kernel/document`: expiry, checksum/size, storage key,
  status, cleanup ownership, persisted before the presigned URL is returned.
- T3 — atomic CAS confirmation of the session and version together.
- T4 — the scheduled GC sweep, with metrics and audit, conservative enough never to remove a
  referenced object.
- T5 — the mirrored counter fix and its own dedicated concurrency test for
  `kernel/artifact.Generate` specifically.

## Out of scope

- **DATA-06's aggregate write contract** (`kernel/resource`, W02-E04's scope) — a different package,
  a different failure mode (missing mirror write, not version-allocation races), no source-text
  linkage to DATA-05.
- **DATA-09's online-migration protocol tooling itself** (W02-E01's scope) — whether this story's own
  schema migrations (the sequence-row/counter table, the upload-session table) are authored through
  DATA-09's expand/backfill/validate/contract protocol once it exists, versus a conventional single
  forward-apply migration, is an implementation-time choice recorded in `plan.md`, not a scope
  dependency this story requires to close (see `dependencies.md` — DATA-05 has no recorded dependency
  on DATA-09).
- **Any broader redesign of `kernel/artifact` or `kernel/document`'s write paths** beyond the
  version-allocation, session-durability, confirmation-atomicity, and GC surface named by PLAN's own
  T1–T5 table.

## Assumptions

- The exact form of the "locked parent counter or dedicated per-aggregate sequence row" (a
  `SELECT ... FOR UPDATE` against a parent row, a PostgreSQL `SEQUENCE`, or a dedicated
  per-aggregate counter table) is not specified by the source beyond PLAN T1's own phrasing offering
  both options — this story's plan records the chosen mechanism and its rationale as an
  implementation-time decision, not one this document can pre-answer (mandate §8.5: "do not invent
  precise code changes where the repository does not yet provide enough information").
- The upload-session table's exact schema (column names, types) is not specified beyond PLAN T2's own
  field list (expiry, checksum/size, storage key, status, cleanup ownership) — the exact DDL is an
  implementation-time decision, to be authored per this programme's `<module>_<entity>` naming
  convention (PLAN T2's own risk note: "New table needs RLS + `<module>_<entity>` naming").
- The GC sweep's exact scheduling mechanism (a cron-style scheduled job, a periodic background
  worker, or a manually-triggered operational command) is not specified by the source — PLAN T4's own
  row says only "Scheduled GC," leaving the scheduling mechanism itself to be determined at
  implementation time.
- The GC sweep's grace window (how long past expiry an unconfirmed session must remain before its
  object is eligible for removal) is not given a specific figure by the source — PLAN T4's own risk
  note calls for "a conservative grace window" without stating a value; this story's plan records the
  chosen value and rationale as an implementation-time decision.

## Dependencies

None within W02-E03 at the story level (this is the epic's only story). Internal task dependencies
per PLAN DATA-05's own Depends-on column: T2 depends on T1; T3 depends on T1 and T2; T4 depends on T2
and T3; T5 depends on T1 (T5 does not depend on T2/T3/T4 — `kernel/artifact` has no upload-session or
GC surface). Depends on W00's exit gate at wave scope. No dependency on any other W02 epic, and no
epic or wave has been found in the source to depend on this story — see `dependencies.md`.

## Affected packages or components

`kernel/artifact` (the `Generate` version-allocation path); `kernel/document` (the
`InitiateUpload` version-allocation path, plus new upload-session persistence, confirmation, and GC
surfaces). A new sequence-row/counter mechanism and a new upload-session table (exact locations and
schemas TBD, see `plan.md` "Unresolved questions").

## Compatibility considerations

The version-allocation mechanism change (from inline `MAX()+1` to a locked counter/sequence) must
preserve the existing version-numbering contract from the caller's perspective — version numbers
remain monotonically increasing per aggregate, with no renumbering of already-issued versions. The
new upload-session table and GC sweep are additive; no existing `kernel/document` caller-facing
behavior is expected to change beyond the session record now existing durably before URL issuance
(previously implicit, now explicit and persisted).

## Security considerations

The GC sweep is the primary security/data-integrity-adjacent concern in this story: PLAN T4's own
risk note states plainly, "False-positive deletion is data loss — conservative grace window." The
sweep must never remove a referenced object; this is a required correctness property, not an optional
hardening measure, and AC-W02-E03-S001-04 is scoped specifically to prove it under a mixed
confirmed/expired/pending test.

## Performance considerations

RISK-W02-E03-001 (epic-scoped, reproduced in this story's front matter): the locked counter/sequence
row becomes the new serialization point for concurrent version allocation. T1 and T5's own
concurrency tests (≥20 concurrent callers) are required to measure lock wait, not merely prove
correctness, per this story's plan.

## Observability considerations

The GC sweep must emit metrics and an audit record for each sweep action (per PLAN T4's own
acceptance framing, "with metrics/audit") — sufficient for an operator to distinguish "the sweep
correctly reclaimed N expired sessions" from "the sweep silently did nothing" or "the sweep removed
something it should not have."

## Migration considerations

This story requires at least one schema migration: the sequence-row/counter mechanism (T1) and the
new upload-session table (T2). Both are new, additive schema elements — neither requires backfilling
or altering existing data in a way that risks the existing version history, per this story's own
"Compatibility considerations" above. Whether these migrations are authored through DATA-09's
protocol (once it exists) or as conventional forward-apply migrations is an implementation-time
choice — see "Out of scope."

## Documentation requirements

Document the counter/sequence mechanism and its concurrency guarantee; document the upload-session
schema and lifecycle (states, transitions, expiry); document the GC sweep's grace window, scheduling
mechanism, and audit/metrics output — so a future reader can understand the full upload lifecycle
without re-reading this story's own planning documents.

## Acceptance criteria

- **AC-W02-E03-S001-01**: N concurrent callers to the version-allocation path (both
  `kernel/artifact.Generate` and `kernel/document.InitiateUpload`) produce N unique, monotonic
  versions with zero unexpected conflicts, proven by a concurrency test with at least 20 concurrent
  callers (PLAN DATA-05 T1's own acceptance criterion and "Tests" column).
- **AC-W02-E03-S001-02**: A `kernel/document` upload-session record (expiry, checksum/size, storage
  key, status, cleanup ownership) is persisted before the presigned upload URL is returned to the
  caller, proven by a test that initiates an upload, simulates a crash, and asserts the session row
  exists with `status='pending'` and a set expiry (PLAN DATA-05 T2's own acceptance criterion and
  "Tests" column).
- **AC-W02-E03-S001-03**: Of two racing confirmation calls against the same upload session, exactly
  one succeeds — the session and version are CASed together atomically, proven by a concurrency test
  (PLAN DATA-05 T3's own acceptance criterion and "Tests" column).
- **AC-W02-E03-S001-04**: The scheduled GC sweep never removes a referenced object and removes every
  past-expiry unconfirmed session, proven by a test exercising a mixed set of confirmed, expired, and
  still-pending sessions (PLAN DATA-05 T4's own acceptance criterion and "Tests" column).
- **AC-W02-E03-S001-05**: `kernel/artifact.Generate`'s version allocation meets the same concurrency
  bar as AC-W02-E03-S001-01, proven by its own dedicated mirror test (PLAN DATA-05 T5's own
  acceptance criterion — "Same concurrency bar as T1" — and "Tests" column — "Mirror test").

## Required artifacts

- The locked-counter/sequence-row version-allocation mechanism (code, applied to both packages).
- The upload-session schema and table.
- The atomic CAS confirmation logic.
- The GC sweep job/mechanism.
- Documentation of the counter mechanism, session lifecycle, and GC sweep.
See `artifacts/index.md`.

## Required evidence

- Concurrency test output (≥20 concurrent callers) for `kernel/document`'s version allocation.
- Crash-simulation test output for upload-session durability.
- Racing-confirmation CAS test output.
- Mixed confirmed/expired/pending GC sweep test output.
- Mirror concurrency test output for `kernel/artifact.Generate`.
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`: `story.md`
and `plan.md` complete, acceptance criteria numbered and measurable, dependencies (none within this
epic) recorded, owner/reviewer assignment pending, unresolved questions (counter mechanism form,
upload-session schema, GC scheduling mechanism and grace window) explicitly recorded rather than
silently assumed.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; all five acceptance
criteria verified with evidence in `evidence/index.md`; `closure.md` completed; independent review
passed per mandate §14 (see `tasks/index.md` "Grouping rationale" for whether this P1 story warrants
a dedicated review task).

## Risks

RISK-W02-E03-001 (the new counter/sequence row becoming the serialization point for concurrent
version allocation — PLAN T1's own risk note: "Counter-row contention is the new serialization
point — measure lock wait") — see epic-level `risks.md` for full detail and mitigation/contingency.

## Residual-risk expectations

Some serialization on the counter row is an inherent, accepted trade-off of moving from an unsafe
`MAX()+1` read to a safe locked mechanism — this story's residual risk is expected to remain open
until T1's concurrency test actually measures lock wait, at which point it is expected to close as
"measured, acceptable" absent an unexpectedly severe contention result.

## Plan

See `plan.md`.
