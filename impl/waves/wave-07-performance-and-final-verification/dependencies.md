---
id: W07-DEPS
type: wave-dependencies
wave: W07
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W07 — Dependencies

## Upstream (waves this wave depends on)

- **W00, W01, W02, W03, W04, W05, W06** — full dependency on all seven prior waves, per `impl/index.md`'s
  wave map row for W07 ("Depends on: all prior"). This is the only wave in the programme with this
  breadth of entry dependency, reflecting its own closure/verification purpose.
- **W03, W05** (specifically) — SEC-05 (W07-E02-S001) hard-depends on SEC-01 (W03-E01), SEC-03
  (W03-E03), SEC-06 (W03-E02), and SEC-04 (W05-E04-S002) all being substantially complete, per PLAN
  SEC-05 T1's own dependency row.
- **W04** (specifically) — PERF-04's own T5 (this wave's W07-E01-S003) has a hard dependency on
  DATA-02/DATA-03's lease primitives (W04-E01, W04-E02), already built and accepted by that wave — this
  wave's own S003 consumes those primitives, it does not re-derive fencing logic.

## Downstream (waves that depend on this wave)

None — W07 is the programme's final wave; no later wave exists to depend on it.

## Internal (within this wave, between epics)

- **W07-E04 depends on W07-E01, W07-E02, W07-E03** (and transitively on every prior wave) — the final
  verification gate (E04-S001) cannot meaningfully re-run the REVIEW §30-style gate until every other
  epic in this wave has reached its own closure state, since the gate's own scope is the whole programme,
  not merely this wave's other three epics in isolation.
- **W07-E01 and W07-E02 are independent of each other** — the performance programme (PERF-02..05, CS-16)
  and the verification-hardening epic (SEC-05, REL-04) target disjoint scope and may proceed in
  parallel, subject each to its own upstream dependency.
- **W07-E03 is independent of W07-E01 and W07-E02** — PROD-01..05's own coordination-artifact
  verification depends only on the underlying framework capabilities each PROD-0N row names (DATA-01
  T1+DATA-09 protocol for PROD-01; FBL-01's shim for PROD-02; DX-07 T1+FBL-09 for PROD-03; SEC-01 T1/T5
  for PROD-04; D-04's hash_version branch for PROD-05) — all already built in earlier waves, not this
  wave's own E01/E02 work.

## Cross-wave dependencies

W03-E01 (SEC-01), W03-E02 (SEC-06), W03-E03 (SEC-03), W05-E04-S002 (SEC-04) for W07-E02-S001. W04-E01
(DATA-02), W04-E02 (DATA-03) for W07-E01-S003's T5. W02-E01 (DATA-09 protocol, transitively via DATA-01)
for W07-E03-S001's PROD-01 verification. W05-E05 (FBL-01) for W07-E03-S001's PROD-02 verification.
W01-E03 (DX-07 T1 via W04-E04-S003, FBL-09 via W01-E03-S001) for W07-E03-S001's PROD-03 verification.
W03-E01 (SEC-01 T1/T5) for W07-E03-S001's PROD-04 verification. W04-E04-S001 (D-04, hash_version) for
W07-E03-S001's PROD-05 verification.

## External dependencies

None new for W07-E02/E03/E04. W07-E01's own reference-environment work (PERF-02 T1) introduces a
provisional external dependency: a dedicated Linux amd64 GitHub Actions runner, per DEC-Q9's own
provisional default — this is infrastructure, not a code dependency, and is tracked at epic level.

## Repository dependencies

W07-E03's own scope is explicitly framework-side-only per mandate §2.3 — this wave verifies that
wowsociety-coordination artifacts exist, it does not create a repository dependency on wowsociety's own
codebase (no wowsociety code is read, modified, or required to exist for this wave's own closure).

## Tooling dependencies

`benchstat` (PERF-06 T2, already integrated per W00-E01-S002's own closure) for W07-E01's statistical
regression-gate consumption. No new tooling dependency introduced by this wave beyond the provisional
reference-runner infrastructure (DEC-Q9).

## Decision dependencies

DEC-Q9 (reference-performance-environment ownership) — open, tracked at W07-E01 epic level, with REVIEW
§F row 9's own provisional default already unblocking relative/container work. No other new D-0N or
DEC-Q decision opens in this wave.
