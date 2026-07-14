---
id: W01-E03-DEPS
type: epic-dependencies
epic: W01-E03
wave: W01
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W01-E03 — Dependencies

## Upstream (epics/waves this epic depends on)

- **W00** — full wave-level dependency via `wave.md`'s entry criteria (W00 exit gate satisfied: the
  8 executed finding-slices re-verified at current HEAD, baselines captured, D-01..D-09 ratified).
  No W00-E02-S003 (D-08 pgx-tracer ADR) dependency applies to this epic specifically — that
  dependency belongs to W01-E02, not W01-E03.
- No dependency on any other W01 epic (W01-E01 static-analysis-utilisation, W01-E02
  observability-correlation, W01-E04 generator-doc-test-fixes) — this epic's two stories touch
  `kernel/httpx`, `kernel/config`, `kernel/validation`, and the scaffold/crud templates, disjoint
  from those epics' files.

## Internal (between this epic's own stories)

- **S001 ↔ S002**: none. S001 touches the scaffold template's `http.Server{}` construction and
  `kernel/httpx/csrf.go`; S002 touches `kernel/httpx/router.go` (`RouteMeta`), `kernel/httpx/decode.go`,
  and crud/scaffold templates for the handler adaptor. Disjoint file sets; either may start first or
  both may run in parallel.

## Downstream (epics/waves that depend on this epic)

Per `../../dependencies.md` (wave-level downstream table):

| Downstream item | Depends on (from this epic) | Why |
|---|---|---|
| W03-E01 (SEC-01) | W01-E03-S002 (central validation seam) | SEC-01's new grant-table endpoints should be built against the `RouteMeta` contract-enforcement pattern this story establishes |
| W05-E03 (AR-03) | W01-E03-S002 (`RouteMeta.Request` contract field) | AR-03 derives projections from `RouteMeta`; the T1 contract-declaration field this story adds must already exist as a stable input |

## Forward-compatibility coordination notes (explicitly NOT `depends_on` relationships)

These are prose coordination notes only. Per the task brief that produced this epic and per mandate
§2.2/§2.6, a note that a later wave's design should remain compatible with work landing now is not the
same as a blocking dependency — recording it as `depends_on` would incorrectly imply this epic cannot
proceed until the later wave exists, which is false; the sequencing is the other way around (this
epic lands first and constrains itself not to conflict with what comes later).

- **AR-04 T5 (W05-E03-S002, waiver mechanism)** — not yet built. S002's T1 introduces a waiver field
  for genuinely body-less mutations. `story-002-central-validation-enforcement/plan.md` records that
  this waiver field must be designed additively/forward-compatible with AR-04 T5's future general
  boot-time-silent-behaviour waiver mechanism — i.e. S002 must not invent a second, conflicting
  one-off waiver shape that AR-04 T5 later has to reconcile or replace. This is a design constraint on
  S002's own T1, not a dependency that blocks S002 from starting or finishing before AR-04 T5 exists.
- **AR-03 (W05-E03-S001..S002, RouteMeta-derived projections)** — not yet built. MATRIX CS-08's own
  note states this story "coordinates with AR-03 (RouteMeta is a projection input) — build T1
  compatibly, don't wait." S002 does not depend on AR-03; it builds the `RouteMeta.Request` field now
  in a shape AR-03 can later consume without rework. Recorded here as the precise coordination note,
  not as a blocking dependency in any `story.md` `depends_on` front-matter field.

## Cross-wave dependencies

None beyond the W00→W01 entry and the downstream table above.

## External dependencies

None beyond what `../../dependencies.md` already records at wave level (`golangci-lint` v2.11.4
pinned toolchain — not directly consumed by this epic's stories, but part of the shared CI gate both
stories' evidence runs against).

## Repository dependencies

- **S001**: wowsociety impact is a required backport (PROD-03, `requirement-inventory.md` §D) —
  wowsociety's already-committed, hand-edited `cmd/api/main.go` does not automatically pick up the
  scaffold-template fix. This story enables PROD-03 (fixes the template) but does not itself perform
  the wowsociety-side backport, per the framework/product boundary (mandate §2.3) — the backport is
  tracked and executed as wowsociety's own repository's work, out of scope here.
- **S002**: wowsociety impact is additive at first (ships behind a profile flag). wowsociety auditing
  its existing handlers for missing `BindAndValidate` calls before flipping the flag to
  enforced-by-default is a downstream coordination note recorded in `story-002.../story.md`'s
  compatibility considerations — not an execution step of this story.

## Tooling dependencies

None beyond the existing `golangci-lint`/CI gate (S001's gosec G120 fix is verified by re-running the
gosec analyzer W01-E01-S002 enables; this epic does not itself enable gosec, only fixes the one named
hit that story's scope excludes).

## Decision dependencies

None. Both stories proceed without a blocking ADR (see `epic.md` "Required decisions").
