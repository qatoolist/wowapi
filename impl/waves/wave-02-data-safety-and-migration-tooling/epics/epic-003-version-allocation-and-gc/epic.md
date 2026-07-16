---
id: W02-E03
type: epic
title: Version allocation and GC
status: accepted
wave: W02
owner: W02FKVerAgg
reviewer: W02ReviewGate
priority: high
created_at: 2026-07-12
updated_at: 2026-07-13
source_requirements:
  - DATA-05
depends_on: []
stories:
  - W02-E03-S001
decisions: []
risks:
  - RISK-W02-E03-001
---

# W02-E03 — Version allocation and GC

## Epic objective

Replace the racy `MAX(version)+1` version-allocation pattern in `kernel/artifact` and
`kernel/document` with a locked parent counter or dedicated per-aggregate sequence row, make
`kernel/document`'s upload-session lifecycle durable and atomically confirmed, and add scheduled
garbage collection of orphaned upload blobs — so that concurrent version allocation never produces a
duplicate or skipped version, and a lost race in the document-upload path no longer leaves an
orphaned, randomly-keyed blob in storage with nothing tracking or reclaiming it.

## Problem being solved

`requirement-inventory.md` row DATA-05 records: "Version allocation races + blob GC (T1–T5) | IMPL |
P1 | planned | W02-E03-S001 |" — no Notes-column dependency is recorded. PLAN's own DATA-05 evidence
is exact: "`kernel/artifact.Generate` and `kernel/document.InitiateUpload` both compute
`MAX(version)+1` inline; the document path's loser leaves an orphaned, randomly-keyed blob with no
GC." The gap this epic closes is twofold: (1) both packages allocate the next version number by
reading the current maximum and adding one, a pattern that is not safe under concurrent callers —
two callers reading the same `MAX(version)` before either commits will both compute the same
"next" version, producing either a duplicate version or a serialization conflict depending on
isolation level; and (2) `kernel/document`'s upload flow generates a storage key and issues a
presigned upload URL before the version race is resolved, so the caller that loses the version race
has already had a blob key allocated in storage with no session record, no expiry, and no scheduled
process to ever reclaim it.

## Scope

- Replacing `MAX(version)+1` with a locked parent counter or dedicated per-aggregate sequence row in
  both `kernel/artifact.Generate` and `kernel/document.InitiateUpload`, proven under concurrent load
  (PLAN DATA-05 T1).
- Durable upload-session records for `kernel/document`: expiry, checksum/size, storage key, status,
  and cleanup ownership, persisted before the presigned URL is returned to the caller (PLAN DATA-05
  T2).
- Atomic confirmation: the session and version are CASed together so that of two racing confirms,
  exactly one succeeds (PLAN DATA-05 T3).
- Scheduled garbage collection removing expired/unreferenced upload objects, with metrics and audit,
  conservative enough to never remove a referenced object (PLAN DATA-05 T4).
- The same counter-race fix applied specifically to `kernel/artifact.Generate`, with its own mirror
  concurrency test at the same bar as T1 (PLAN DATA-05 T5).

## Out of scope

- **Any change to `kernel/artifact` or `kernel/document`'s other write paths** not implicated by the
  version-allocation or upload-session races — this epic's scope is bounded to what PLAN's own
  DATA-05 evidence and T1–T5 task table name: version allocation, upload-session durability,
  confirmation atomicity, and GC. It does not revisit unrelated aspects of either package (e.g.
  artifact content validation, document access-grant semantics).
- **DATA-06's aggregate write contract** (the resource-mirror atomicity helper, W02-E04's scope) —
  DATA-05's version-allocation fix and DATA-06's mirror-write contract are independent findings
  targeting different packages (`kernel/artifact`/`kernel/document` here; `kernel/resource` there);
  no source text ties them together, and this epic does not attempt to unify them.
- **DATA-09's online-migration protocol** (W02-E01's scope) — this epic's own schema changes (the new
  upload-session table, any sequence-row/counter table) are implementation details of T1–T2, not a
  consumer relationship recorded in the source; DATA-05's Notes column in `requirement-inventory.md`
  cites no dependency on DATA-09, and this epic does not assume one. Whether this epic's own
  migrations are authored using DATA-09's protocol once it exists is an implementation-time choice,
  not a scope dependency (see "Dependencies" below).

## Source requirements

DATA-05 (T1, T2, T3, T4, T5). Confirmed to have no D-0N architecture-decision dependency — a scan of
`requirement-inventory.md` §B and REVIEW §F/§U finds no D-0N row citing DATA-05 as its consumer (see
`wave.md` "Assumptions" for the wave-level confirmation covering DATA-05 by name).

## Architectural context

DATA-05 targets two independent packages (`kernel/artifact`, `kernel/document`) that share the same
underlying defect pattern — inline `MAX(version)+1` computation with no locking or sequence
primitive — but differ in blast radius: `kernel/artifact.Generate`'s race (T5) produces a version-
number correctness defect only, while `kernel/document.InitiateUpload`'s race (T1–T4) additionally
drives an upload-session and storage-blob lifecycle whose losing caller leaves orphaned state behind
in external storage, not just in the database. This is why PLAN's own task table gives
`kernel/document` four tasks (T1–T4: the counter fix, session durability, atomic confirmation, and
GC) while `kernel/artifact` needs only one (T5: the same counter fix, mirrored). T1's own row states
the counter fix applies to "both `kernel/artifact` and `kernel/document`," and T5 is not a duplicate
of that work — T5 exists to hold `kernel/artifact.Generate` to "the same concurrency bar as T1" with
its own dedicated mirror test, because `kernel/artifact` has no upload-session or GC surface for T2–
T4's fixes to also exercise. Read together: T1 establishes the counter/sequence mechanism and applies
it to both packages' version-read path; T5 is the acceptance gate confirming `kernel/artifact`'s
application of that same mechanism is independently proven, not merely assumed correct because T1's
own test happened to cover `kernel/document`.

## Included stories

- **W02-E03-S001 — version-races-and-blob-gc** (PLAN DATA-05 T1–T5, all five tasks): the locked-
  counter/sequence-row version-allocation fix for both packages; durable upload-session records;
  atomic CAS confirmation; scheduled GC; the mirrored `kernel/artifact` concurrency proof.
  `impl/analysis/wave-allocation-detail.md`'s canonical allocation states this exactly: "E03
  version-allocation-and-gc (DATA-05): S001 all T1–T5 (single reviewer domain)" — all five tasks are
  grouped into one story because they form a single, tightly-coupled persistence-correctness fix
  (one counter mechanism, one session lifecycle it enables, one confirmation path that CASes both,
  one GC sweep that depends on the session records existing, and one mirrored proof of the same
  mechanism applied to the sibling package), reviewable coherently by a single reviewer with
  domain expertise in this exact persistence path, rather than split across artificial story
  boundaries that would separate a fix from its own proof (T5) or a session mechanism from the
  confirmation logic that CASes it (T2/T3). This is consistent with mandate §12's decomposition
  guidance: a story should be split when it "affects several unrelated framework capabilities" or
  "require[s] different reviewers" — T1–T5 affect one capability (correct, durable version/upload
  allocation) and do not require different reviewers.

## Dependencies

No dependency on any other W02 epic — this epic's scope (`kernel/artifact`, `kernel/document`) is
disjoint from W02-E01/E02's migration-protocol and tenant-FK work, W02-E04's `kernel/resource` scope,
and W02-E05's seed-sync scope. This epic depends only on W00's exit gate per `wave.md`'s entry
criteria. Confirmed from the source: PLAN's own DATA-05 T1 row has an empty Depends-on column ("—"),
and `requirement-inventory.md`'s DATA-05 Notes column cites no dependency on DATA-09 or DATA-01 (or
any other finding). This is unlike W02-E02, which is explicitly gated on W02-E01; W02-E03 carries no
equivalent gate. See `dependencies.md`.

## Risks

RISK-W02-E03-001 (the new counter-row becoming the new serialization point under concurrent load,
per PLAN T1's own risk note) — see `risks.md` for full detail and mitigation/contingency.

## Required decisions

None. DATA-05 has no D-0N architecture-decision dependency in the source — confirmed by scanning
`requirement-inventory.md` §B and REVIEW §F/§U for any D-0N row citing DATA-05; none exists. This
epic's story accordingly carries no `decisions/` directory.

## Epic acceptance criteria

- **AC-W02-E03-01**: N concurrent callers to the version-allocation path (both `kernel/artifact` and
  `kernel/document`) produce N unique, monotonic versions with zero unexpected conflicts, proven by a
  concurrency test with at least 20 concurrent callers.
- **AC-W02-E03-02**: A `kernel/document` upload-session record (expiry, checksum/size, storage key,
  status, cleanup ownership) is persisted before the presigned upload URL is returned to the caller,
  proven by a test that initiates an upload, simulates a crash, and asserts the session row exists
  with `status='pending'` and a set expiry.
- **AC-W02-E03-03**: Of two racing confirmation calls against the same upload session, exactly one
  succeeds — the session and version are CASed together atomically, proven by a concurrency test.
- **AC-W02-E03-04**: The scheduled GC sweep never removes a referenced object and removes every
  past-expiry unconfirmed session, proven by a test exercising a mixed set of confirmed, expired, and
  still-pending sessions.
- **AC-W02-E03-05**: `kernel/artifact.Generate`'s version allocation meets the same concurrency bar
  as AC-W02-E03-01, proven by its own dedicated mirror test (not merely covered incidentally by the
  `kernel/document` test).
- **AC-W02-E03-06**: The story has passed independent review per mandate §14 (see "Required
  decisions" in `stories/story-001-version-races-and-blob-gc/plan.md` for whether a dedicated review
  task is warranted for this P1 story).

## Closure conditions

The story reaches `accepted` (satisfying its own `closure.md`); AC-W02-E03-01 through AC-W02-E03-06
above are all satisfied; `closure-report.md` for this epic is completed with reviewer conclusion and
acceptance date; RISK-W02-E03-001 (counter-row contention as the new serialization point) is recorded
with its measured lock-wait evidence, not silently dropped at closure.

## Status update (2026-07-16)

`status: accepted` (reconfirmed). Independent review executed 2026-07-16 superseded the prior
uncorroborated `W02ReviewGate` citation and the story's false "T006 complete" claim — no dedicated
review task ever existed for this story by design.

— dated 2026-07-16, conductor adjudication (Fable 5), per review-gate-2026-07-16.md records
