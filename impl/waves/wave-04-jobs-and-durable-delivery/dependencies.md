---
id: W04-DEPS
type: wave-dependencies
wave: W04
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W04 — Dependencies

## Upstream (waves this wave depends on)

- **W02 — narrow, story-scoped dependency, not a whole-wave dependency of every W04 epic.** Per
  `impl/index.md`'s wave map ("W04 | jobs-and-durable-delivery | ... | Depends on | W02 (DATA-09
  for W6-T1 migration) | Epics | 4"), the dependency is specifically on W02-E01 (the DATA-09
  online-migration protocol) and specifically for W04-E04-S001 (DATA-08 W6-T1's audit-hash
  migration). `impl/analysis/wave-allocation-detail.md`'s W04-E04 row states this exactly: "S001
  audit-hash-widening (W6-T1; D-04 enacted; dep W02-E01 protocol; ...)." No other W04 story — not
  E01 (DATA-02), not E02 (DATA-03), not E03 (DATA-04), not E04-S002/S003 (DATA-08 W6-T2..T5, DX-07)
  — has a dependency on W02's protocol or on any other W02 epic. This distinction is confirmed, not
  overstated: `requirement-inventory.md`'s notes column for DATA-02/DATA-03/DATA-04 cites no W02
  dependency; only DATA-08's row states "W0 slice EXECUTED (verified ×2); W6-T1 hash widening (D-04)
  + T2–T5 planned," with the W02 dependency recorded separately in
  `wave-allocation-detail.md`.
- The programme's strict W00→W07 entry ordering (`impl/index.md`: "Execution order: strictly
  W00→W07 for wave entry") sequences W04 after W02 and W03 by convention. W04 has no technical
  dependency on W03 (identity-and-session-security) at all — no W04 finding cites a SEC-0N or
  DATA-07 dependency.

## Downstream (waves that depend on this wave)

| Downstream item | Depends on (from W04) | Why |
|---|---|---|
| W05-E03-S002 (AR-04 T5 waiver mechanism) | — (inverse relationship) | W04-E04-S003 (DX-07 T4) depends on W05-E03-S002's waiver mechanism, not the reverse — recorded here for completeness since it is the one place this wave's own scope is gated on a *later* wave, per PLAN DX-07 T4's dependency column ("T1-T3, AR-04's waiver framework"). This is an explicit, approved exception to "a later wave must not start while a mandatory predecessor capability remains unaccepted" (mandate §15) because DX-07 T4 alone — not the whole DX-07 finding, not the whole W04-E04-S003 story — is deferred, and the deferral is recorded, not silent. |

No other wave is confirmed to depend on W04 by name in `impl/index.md`'s wave map or in
`wave-allocation-detail.md`'s cross-wave sequencing notes.

## Internal (within this wave, between epics)

- **W04-E02, W04-E03 depend on W04-E01 for the shared lease/fencing primitive and the shared chaos
  harness.** DATA-03 T1's own dependency column states "DATA-02 T1"; DATA-04 T2's states "DATA-02
  T1; T1 as interim." `wave-allocation-detail.md` states this exactly: "S003 idempotency-and-chaos
  (T5, T6, T7 chaos harness — harness shared with E02/E03)." Concretely:
  - W04-E02-S001 (notify/webhook three-stage protocol) depends on W04-E01-S001 (the primitive
    itself, PLAN DATA-03 T1: "Lease columns via shared primitive, not a bespoke copy").
  - W04-E02-S001/S002/S003's chaos work (DATA-03 T8) and W04-E03-S002's chaos work (DATA-04 T6)
    both depend on W04-E01-S003's chaos harness (DATA-02 T7: "build as a reusable chaos harness
    shared with DATA-03/DATA-04") — they reuse it, not reimplement it.
- **W04-E03-S001 (DATA-04 T1, the stopgap) has no dependency on W04-E01.** Per PLAN DATA-04 T1's own
  dependency column ("—") and `wave-allocation-detail.md` ("S001 T1 stopgap (can start at wave
  entry)"), the stopgap fix (correcting the false migration comment, enforcing single-processor via
  advisory lock/CAS) may land before the shared primitive exists. W04-E03-S002 (the full
  leased-claim rewrite, DATA-04 T2-T6) does depend on W04-E01-S001.
- **W04-E04 is internally sequenced E04-S001 → (no forced order between S002 and S003).** DATA-08
  W6-T2 through W6-T5 (S002) depend on W6-T1 (S001) per PLAN's own dependency column ("W6-T2 | W6-T1
  |"). DX-07 (S003) has no dependency on S001/S002 — it is an independent readiness/diagnostics
  concern grouped into the same epic by MATRIX CS-21's shared closure-spec framing, not by a task
  dependency.
- **W04-E01's own internal sequencing** mirrors W02-E01's phase-pipeline pattern: S001 (the
  primitive itself) → S002 (jobs lease/finalize/reclaim, consuming S001) → S003 (idempotency
  contract + chaos harness, consuming S002's finalize paths as the boundary the chaos test
  exercises).

## Cross-wave dependencies

None beyond the W02→W04-E04-S001 edge stated above and the inverse W05→W04-E04-S003(T4) edge. W04
does not depend on W01, W03, W05, W06, or W07 for any of its own exit criteria.

## External dependencies

`cenkalti/backoff/v5` (FBL-04, W04-E02-S003) — already present transitively in the module graph per
REVIEW §L's approved-dependency register ("New approvals for reuse work: `cenkalti/backoff/v5`
(MIT, already transitive)"); this wave's own action is to add it as a direct dependency and adopt
it, not to introduce a new external dependency into the framework's dependency surface for the
first time.

## Repository dependencies

None cross-repo for this wave's own framework-side closure. wowsociety impact is real but
non-blocking:

- **DATA-02** — PLAN's own wowsociety-impact note: "Not affected. Zero `kernel/jobs` import, zero
  job registration anywhere in wowsociety. Would become breaking (worker signature change, T5) the
  moment wowsociety registers a job — flag for roadmap." No coordination required for this wave's
  closure; T5's breaking signature change is recorded as a forward-looking coordination note, not a
  blocking dependency today.
- **DATA-03** — "Not affected today; conditionally breaking in the future." Same non-blocking
  posture.
- **DATA-04** — "Not affected. Zero `kernel/bulk` import anywhere in wowsociety."
- **DATA-08 W6-T1** — "Affected for `kernel/audit` — BREAKING for W6-T1." wowsociety produces real,
  live audit rows today (`identity/service.go`, `policy/service.go`, `impersonation.go`'s grant/
  revoke writes, `cmd/api/main.go`'s API-key audit wiring). Tracked as `PROD-05` in
  `requirement-inventory.md` §D ("DATA-08 W6 staging audit re-verification before version bump" —
  product compliance drill, hash_version branch verification) — product-level, excluded from this
  wave's framework-side closure per mandate §2.3, but recorded here per that same mandate section's
  requirement to identify the generic framework capability enabling it (the `hash_version` branch
  itself).
- **DX-07** — "Affected — wowsociety's already-generated `cmd/api/main.go` shows the identical gap"
  (readiness two-check shape). Not breaking — T1's fix changes the template only, not wowsociety's
  already-committed `main.go`. Tracked as `PROD-03` in `requirement-inventory.md` §D (readiness/
  timeout backports to committed main.go) — product-level hand-edit, out of this wave's scope.

## Tooling dependencies

None new beyond the `cenkalti/backoff/v5` module addition (already transitively present). DATA-02/
03/04's chaos-test infrastructure extends the existing Go test toolchain; no new CI system is
introduced.

## Decision dependencies

Only W04-E04-S001 depends on a decision — D-04 (audit `hash_version` discriminator), already
ratified in `impl/waves/wave-00-baseline-and-verification/epics/epic-002-baseline-capture/stories/
story-003-adr-ification/decisions/adr-004-audit-hash-version-column.md`. No other W04 story depends
on a D-0N decision — confirmed per `wave.md` "Assumptions."
