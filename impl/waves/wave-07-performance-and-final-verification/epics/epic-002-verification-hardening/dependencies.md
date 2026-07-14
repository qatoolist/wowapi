---
id: W07-E02-DEPS
type: epic-dependencies
epic: W07-E02
wave: W07
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W07-E02 — Dependencies

## Upstream (epics/waves this epic depends on)

- **W03-E01 (SEC-01), W03-E02 (SEC-06), W03-E03 (SEC-03), W05-E04-S002 (SEC-04)** — S001's own hard
  dependency: "SEC-01–04 substantially complete," per PLAN SEC-05 T1's own dependency row. Already
  satisfied by this wave's own all-prior-waves entry gate.
- **W00-E01-S002 (PERF-06 T1/T2, REL-04 T1-T4)** — already `EXECUTED`; S002 consumes their already-
  verified state as its own starting point.

## Downstream (epics/waves that depend on this epic)

| Downstream item | Depends on (from this epic) | Why |
|---|---|---|
| W07-E04-S001 (final verification gate, this wave) | This epic (both stories) | The final gate's own re-run scope includes this epic's own SEC-05/REL-04 closure state as one of many inputs. |

## Internal (within this epic)

S001 and S002 are independent of each other — SEC-05's control-map work and REL-04's coverage-
truthfulness work target disjoint scope (a security-standards mapping exercise vs. CI-gate/fuzz-
infrastructure work) and may proceed in parallel, subject each to its own upstream dependency.

## Cross-wave dependencies

W03-E01, W03-E02, W03-E03, W05-E04-S002 (SEC-01/03/04/06) for S001, as stated above.

## External dependencies

S001's own external assessment is, by definition, performed by an external party (not a coding agent) —
this is a genuine external dependency, though not a human-blocked one in the DEC-Q sense (it is a
professional-services engagement, not a repo-admin action). S002's T8 real-fuzz work uses the existing
Go native fuzz testing infrastructure, already partially wired (per `requirement-inventory.md`'s own
FBL-07 row: "Nightly ci schedule EXISTS since #24 (fuzz portion still seed-replay only)").

## Repository dependencies

None cross-repo for this epic's own closure.

## Tooling dependencies

Go's native fuzz testing (`go test -fuzz`) for S002's own T8. No new tooling dependency for S001 beyond
the external assessment itself (a professional-services engagement, not a software tool).

## Decision dependencies

None. See `epic.md` "Required decisions."
