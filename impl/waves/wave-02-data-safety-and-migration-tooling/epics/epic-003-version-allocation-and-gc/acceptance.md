---
id: W02-E03-ACCEPTANCE
type: epic-acceptance
epic: W02-E03
wave: W02
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W02-E03 — Epic-level acceptance

Reproduces and elaborates `epic.md`'s "Epic acceptance criteria" section as a standalone,
independently-referenceable acceptance record, consistent with the wave-level `../../acceptance.md`
pattern.

## AC-W02-E03-01 — Race-free version allocation under concurrent load

N concurrent callers to the version-allocation path in both `kernel/artifact.Generate` and
`kernel/document.InitiateUpload` produce N unique, monotonic versions with zero unexpected
conflicts. Proven by a concurrency test with at least 20 concurrent callers, per PLAN DATA-05 T1's
own "Tests" column. Traces to W02-E03-S001 (T1).

## AC-W02-E03-02 — Durable upload sessions persisted before URL issuance

A `kernel/document` upload-session record (expiry, checksum/size, storage key, status, cleanup
ownership) is persisted before the presigned upload URL is returned to the caller. Proven by a test
that initiates an upload, simulates a crash, and asserts the session row exists with
`status='pending'` and a set expiry, per PLAN DATA-05 T2's own "Tests" column. Traces to
W02-E03-S001 (T2).

## AC-W02-E03-03 — Atomic confirmation under racing confirms

Of two racing confirmation calls against the same upload session, exactly one succeeds — the session
and version are CASed together atomically. Proven by a concurrency test, per PLAN DATA-05 T3's own
"Tests" column. Traces to W02-E03-S001 (T3).

## AC-W02-E03-04 — Conservative, correct scheduled GC

The scheduled GC sweep never removes a referenced object and removes every past-expiry unconfirmed
session. Proven by a test exercising a mixed set of confirmed, expired, and still-pending sessions,
per PLAN DATA-05 T4's own "Tests" column. Traces to W02-E03-S001 (T4).

## AC-W02-E03-05 — kernel/artifact mirrors the same concurrency bar

`kernel/artifact.Generate`'s version allocation meets the same concurrency bar as AC-W02-E03-01,
proven by its own dedicated mirror test, per PLAN DATA-05 T5's own "Tests" column ("Mirror test") —
not merely covered incidentally by the `kernel/document` test written for T1. Traces to
W02-E03-S001 (T5).

## AC-W02-E03-06 — Independent review passed

The story (S001) has passed independent review per mandate §14, confirming: the counter/sequence
mechanism is genuinely applied to both `kernel/artifact` and `kernel/document`, not merely one; the
upload-session and GC mechanisms genuinely satisfy their own acceptance criteria with valid evidence
identifying the tested commit SHA; and RISK-W02-E03-001's lock-wait measurement is genuinely
recorded, not silently omitted.

## Acceptance authority

Data/reliability lead, per `../../wave.md`'s wave-level acceptance authority (PLAN §5.3's
accountable role for PF-DATA).
