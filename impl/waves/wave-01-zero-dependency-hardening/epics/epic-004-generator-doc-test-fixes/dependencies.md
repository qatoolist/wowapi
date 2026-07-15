---
id: W01-E04-DEPS
type: epic-dependencies
epic: W01-E04
wave: W01
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W01-E04 — Dependencies

## Upstream (epics/waves this epic depends on)

- **W00** (full wave) — per `../../dependencies.md` (wave-level), W01 depends on W00's exit gate: the
  8 executed finding-slices re-verified at current HEAD, baselines captured, D-01..D-09 ratified as
  ADRs. This epic has no additional upstream dependency beyond the wave-level W00 gate — none of
  DX-01/DX-02/DX-05/T-DOC-01/FBL-03/T-TEST-01's scope touches AR-01/AR-02/SEC-01/DATA-09.

## Downstream (epics/waves that depend on this epic)

| Downstream item | Depends on (from this epic) | Why |
|---|---|---|
| W06-E01-S002 (DX-04, golden consumer) | W01-E04-S001 (DX-01 T5 harness) | `requirement-inventory.md` row DX-04 states "Dep DX-01 T5" — the golden-consumer/upgrade-matrix story reuses the isolated-temp-dir generate→build→boot→smoke scaffold this epic's S001 builds, rather than reimplementing it. |
| wowsociety upstream register (PF-2 entry) | W01-E04-S001 (DX-02 verb fix) | S002's FBL-03 task cannot mark PF-2 closed until S001's DX-02 fix actually lands — see "Internal" below for the precise cross-story form of this dependency. |
| W06-E02/E03 (DX-06/REL-03) | W01-E04-S002 (DX-05 T5 deferral) | DX-05 T5 is explicitly "shared with REL-03" per the plan; this epic's S002 does not implement it, only records the deferral with the cross-reference, so REL-03's W06 story is the actual owner of that shared plumbing. |

## Internal (within this epic)

- **S002 → S001, story-level** (`W01-E04-S002.depends_on: ["W01-E04-S001"]`): S002's FBL-03 task item
  for PF-2's closure is contingent on S001's DX-02 task (permission-verb fix) landing — the upstream
  register entry cannot honestly be marked closed before the fix it cites exists. This is recorded as
  a genuine story-level `depends_on` entry in S002's front matter, not merely a note, because the
  specific sub-task (PF-2 closure) cannot produce valid evidence without S001's artifact.
- **S002's DX-05 T4 → S001's DX-01 T1–T4 (task-level, soft/plumbing dependency)**: DX-05 T4 ("`wowapi
  version` fails mutating generator commands on incompatible major/minor pairing") reuses the
  version-verification plumbing (the `go list -m` resolution check, the version-comparison logic) that
  S001's T001 builds for DX-01 T1–T4. This is not a hard story-level blocking dependency — S002's
  `story.md`/`plan.md` can be authored in parallel with S001 — but S002's task implementing DX-05 T4
  should not begin *implementation* before S001's T001 has landed, to avoid duplicating or diverging
  from the version-comparison logic S001 establishes. Recorded at task level in S002's task file for
  the DX-05-T3/T4 task, not escalated to a story-level `depends_on` entry, since S002's other two tasks
  (T-DOC-01 fix, FBL-03 register reconciliation minus the PF-2 sub-item) have no such dependency.
- **S003 has no dependency on S001 or S002**: T-TEST-01's reproduction-and-diagnosis investigation
  targets `internal/e2e`/`testkit`, entirely disjoint from S001's `internal/cli/` scope and S002's
  documentation scope. S003 may proceed in parallel with either.

## Cross-wave dependencies

None beyond the W00→W01 entry dependency stated above, and the W06 downstream consumers noted in the
table above (DX-04 golden consumer, DX-05 T5/REL-03).

## External dependencies

None. This epic introduces no new external tool, service, or dependency — DX-01's version resolution
uses Go's own `go list -m` and VCS metadata (`git`) tooling already present in any Go development
environment; DX-02's fix is a one-token template change; T-TEST-01's investigation uses the existing
`go test -count=N -parallel=N` tooling and `testkit`'s existing DB-cloning mechanism.

## Repository dependencies

- **FBL-03's PF-2/PF-6/RFF-001 register entries live in the `wowsociety` repository, not `wowapi`.**
  Per mandate §2.3's framework/product boundary discipline, this epic's S002 can only plan/recommend
  the wowsociety-side register edit (a PROD-level coordination note), not execute it directly — this
  is a genuine cross-repository dependency, not merely a documentation nuance, and is recorded as such
  in S002's `story.md` "Out of scope" and `plan.md`.
- **wowsociety impact assessment for DX-01/DX-02**: both findings were independently confirmed as
  NOT affecting wowsociety (DX-01: `replace => ../wowapi` never touches the CLI-generated dependency
  line; DX-02: `docs/CONVENTIONS.md:10` governance kept existing wowsociety modules immune) — this is
  informational, not a blocking cross-repository dependency, but is recorded here because it was an
  explicit check performed during scoping, not an assumption.

## Tooling dependencies

None beyond the Go toolchain (`go`, `go list -m`, `git`) already required for any wowapi development.

## Decision dependencies

None. See `epic.md` "Required decisions" — this epic requires no new ADR.
