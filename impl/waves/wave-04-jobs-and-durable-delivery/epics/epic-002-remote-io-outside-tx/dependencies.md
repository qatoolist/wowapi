---
id: W04-E02-DEPS
type: epic-dependencies
epic: W04-E02
wave: W04
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W04-E02 — Dependencies

## Upstream (epics/waves this epic depends on)

- **W04-E01 (this wave) — hard dependency, not merely a sequencing convention.** DATA-03 T1's own
  dependency column states "DATA-02 T1" (the shared lease/fencing primitive, built in
  W04-E01-S001); `wave-allocation-detail.md`'s W04-E02 row confirms this exactly: "S001
  notify-and-webhook-three-stage (T1, T2, T3)." Concretely:
  - W04-E02-S001 depends on **W04-E01-S001** — T1 requires the shared primitive's lease columns
    ("Lease columns via shared primitive, not a bespoke copy") to exist before notify/webhook claim
    rows can be built against them.
  - W04-E02-S002's T8 (the 6-boundary chaos test) depends on **W04-E01-S003** — the shared chaos
    harness built there for DATA-02 T7 ("build as a reusable chaos harness shared with
    DATA-03/DATA-04"), per `dependencies.md` (wave-level): "W04-E02-S001/S002/S003's chaos work
    (DATA-03 T8) ... depend on W04-E01-S003's chaos harness ... they reuse it, not reimplement it."
  - W04-E02-S002's T6 (per-adapter idempotency-contract declaration) depends on **T2 and T3**
    (this epic's own S001), per PLAN DATA-03 T6's dependency column ("T2, T3").
- No dependency on W02 or W03 — `requirement-inventory.md`'s notes column for DATA-03 cites no W02
  dependency, and `wave.md`'s "Assumptions" confirms only W04-E04-S001 (DATA-08 W6-T1) carries the
  narrow W02-E01 dependency in this wave.

## Downstream (epics/waves that depend on this epic)

No epic or wave outside W04 is confirmed to depend on W04-E02 by name in `impl/index.md`'s wave map
or in `wave-allocation-detail.md`'s cross-wave sequencing notes. Within W04, this epic's own exit
(DATA-03 T1–T6, T8 satisfied) is one of four independent contributions to W04's wave-level exit
criteria, alongside W04-E01, W04-E03, and W04-E04 — no other W04 epic consumes this epic's own
output.

## Internal (within this epic)

- **S001 → S002 (partial).** S002's T4, T5 (inbound two-phase verification, failed-signature audit)
  have no task-level dependency on S001's T1–T3 beyond both consuming the same shared primitive
  from W04-E01 — T4's own dependency column is "T1" (this epic's T1, the shared-primitive reuse
  task), not T2/T3. S002's T6 (adapter idempotency contract) does depend on S001's T2 and T3
  directly, per PLAN DATA-03 T6's dependency column ("T2, T3") — the contract-declaration mechanism
  is validated against the concrete `Sender`/webhook-delivery adapters T2/T3 produce.
- **S002's T8 (chaos test) depends on T2, T3, T4** (its own dependency column: "T2-T4") — the chaos
  test exercises the three-stage protocol (T2, T3) and the inbound two-phase verification (T4)
  together, so T8 cannot start until all three are implemented.
- **S003 (FBL-04) has no dependency on S001 or S002.** FBL-04 is a small, independently-scoped
  retry-library adoption; `wave-allocation-detail.md` groups it into this epic ("S003
  retry-adoption") by shared-package proximity (both touch `kernel/notify`/`kernel/webhook`'s
  remote-I/O call sites), not by a task dependency. S003 may proceed in parallel with S001/S002,
  subject only to avoiding a merge conflict on the same call sites S001/S002 are simultaneously
  restructuring — a coordination note, not a blocking dependency, recorded in S003's own `plan.md`.

## Cross-wave dependencies

None beyond the W04-E01 dependency stated above. This epic does not depend on W01, W02, W03, W05,
W06, or W07 for any of its own exit criteria.

## External dependencies

`cenkalti/backoff/v5` (FBL-04, S003) — already present transitively in the module graph per
REVIEW §L's approved-dependency register ("New approvals for reuse work: `cenkalti/backoff/v5`
(MIT, already transitive)"); this epic's own action is to add it as a direct dependency and adopt
it, not to introduce a new external dependency into the framework's dependency surface for the
first time.

## Repository dependencies

None cross-repo for this epic's own closure. Per `wave.md`'s wowsociety-impact framing for
DATA-03: "Not affected today; conditionally breaking in the future. Zero `kernel/notify`/
`kernel/webhook` usage found. If wowsociety ever calls `webhook.HandleInbound` directly, T4's
transaction-ownership contract change would need integration review — flag for future, not now."
No coordination required for this epic's closure; T4's breaking change is recorded as a
forward-looking coordination note in S002's `story.md`/`plan.md`, not a blocking dependency today.

## Tooling dependencies

None beyond the `cenkalti/backoff/v5` module addition (already transitively present). The chaos-test
infrastructure this epic's S002-T8 reuses extends the existing Go test toolchain via W04-E01-S003's
harness; no new CI system is introduced by this epic.

## Decision dependencies

None. See `epic.md` "Required decisions."
