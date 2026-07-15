---
id: W05
type: wave
title: Application model and layering
status: planned
owner: unassigned
reviewer: unassigned
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-12
included_epics:
  - W05-E01
  - W05-E02
  - W05-E03
  - W05-E04
  - W05-E05
depends_on:
  - W03
blocks:
  - W06
source_requirements:
  - AR-01
  - AR-02
  - AR-03
  - AR-04
  - AR-06
  - SEC-04
  - FBL-01
  - CS-01
  - CS-17
---

# W05 — Application model and layering

## Objective

Replace the framework's mutable, unowned module-registration surface with an ownership-bound,
immutable `ApplicationModel` (AR-01), a typed compiled provider graph (AR-02), a single
authoritative manifest with derived projections (AR-03 remainder), boot-time strictness with an
explicit waiver mechanism (AR-04 remainder), a closed constructor-bypass surface with a bounded
authz cache (AR-06 remainder + SEC-04), and a corrected four-level package layering with the nine
misplaced kernel packages re-homed to `foundation/` (FBL-01). This wave converts "the kernel
surface is unowned, unbounded, and wrongly layered" into "the kernel surface is a small, stable,
capability-secured, and correctly-layered contract" — the load-bearing architectural correction
this programme's own risk register names as its top item: "kernel surface locks in before FBL-01."

## Rationale

`impl/index.md`'s wave map assigns W05 "AR-01/AR-02 ownership model (+D-02/D-03), AR-03/AR-04
remainder, AR-06 remainder, SEC-04 cache (+D-06), FBL-01 kernel re-home," depending on W03 ("actor
model stabilises registrar security assumptions"). `impl/analysis/wave-allocation-detail.md`'s
"Cross-wave sequencing notes" state this exactly: "W05 entry requires W03-E01 acceptance (actor
model stability)." The PLAN's own PF-ARCH cross-cutting note makes the internal sequencing
argument directly: "AR-01 T1/T2 are the load-bearing prerequisite for AR-02's `Registrar` reuse and
AR-03's manifest-consumes-model dependency." MATRIX CS-01 states FBL-01's own dependency
explicitly: "Dependencies: AR-01/02 first (re-homing mid-registration-rework causes double churn)."
SEC-04 is grouped into this wave's E04 alongside AR-06 because both are wiring-and-cache-hygiene
concerns sharing no dependency on AR-01/02/03/04's registration-model work, and because SEC-04's
T4 cross-pod cache-invalidation question is resolved by D-06 (ratified in W00-E02-S003), unblocking
what MATRIX CS-17 calls the "highest-risk task" in that finding.

## Framework capabilities delivered

- An immutable `ApplicationModel` compiled via `collect → validate → seal → expose read-only
  snapshot`, with an owner-bound `Registrar` capability type that makes cross-module ownership
  claims structurally impossible rather than string-compared (AR-01 T1-T11).
- A typed port-key API (`port.Key[T]`) and a compiled, boot-time-validated provider graph replacing
  ad hoc wiring, with zero reflection on request hot paths (AR-02 T1-T7).
- A single authoritative module manifest from which routes, permission/resource catalogs, and other
  projections are deterministically derived, closing the golden-delta acceptance gate (AR-03 T1,
  T3, T4, T5).
- Boot-time rejection of duplicate collectors, empty required fragments, and unwaived no-op
  adapters in `prod`, built on a shared waiver mechanism consumed by SEC-06 and DX-07 (AR-04 T2-T5).
- A closed constructor-bypass surface in `kernel/kernel.go`, enforced by a lint rule forbidding
  ad hoc infrastructure construction outside composition packages (AR-06 T2-T3).
- A bounded, epoch-invalidated authorization cache replacing the unbounded map, closing SEC-04's
  cross-pod staleness gap via the D-06 epoch-table decision and closing DATA-07 T4's
  cache-invalidation acceptance criterion (SEC-04 T1-T6).
- A corrected kernel package layer: nine app-foundation/adapter packages re-homed out of `kernel/`
  to `foundation/`, with a deprecated forwarding shim for `kernel/mfa`'s wowsociety-facing surface
  and an extended depguard/boundaries lint that fails on any future kernel-layer violation (FBL-01).

## Included epics

- **W05-E01 — application-model (AR-01)**: the `ApplicationModel` lifecycle skeleton and
  owner-bound `Registrar` capability type; per-registry ownership wrappers; snapshot immutability,
  deterministic model hash, and post-seal race safety; the legacy compatibility adapter.
- **W05-E02 — typed-ports (AR-02)**: the typed port-key API and compiled provider graph; boot-time
  graph validation and the three-profile projection; retirement of the hand-maintained
  `kernel/lifecycle` manifest with a legacy shim.
- **W05-E03 — authoritative-declarations (AR-03 + AR-04 remainder)**: the manifest schema and its
  derived projections, including the golden-delta acceptance gate; boot-time strictness (duplicate
  collectors, empty fragments, post-seal config rejection) and the shared no-op-adapter waiver
  mechanism.
- **W05-E04 — wiring-and-cache-hygiene**: closure of the remaining `kernel/kernel.go`
  constructor-bypass surface and its lint enforcement; the bounded, epoch-invalidated authz cache.
- **W05-E05 — kernel-re-home (FBL-01)**: the mechanical re-home of nine packages to `foundation/`
  with the `kernel/mfa` forwarding shim, extended depguard/boundaries enforcement, and
  wowsociety-facing re-home verification.

## Entry criteria

- W03's exit gate satisfied, specifically W03-E01 (server-side session/grant state, SEC-01)
  accepted — per `impl/analysis/wave-allocation-detail.md`'s explicit cross-wave sequencing note:
  "W05 entry requires W03-E01 acceptance (actor model stability)." This wave's `Registrar`
  capability type (AR-01 T2) and the security-boundary work built on it assume the actor/session
  model W03-E01 stabilises; starting W05 registration-model work against an unstable actor model
  would risk rework.
- W00's exit gate satisfied at programme scope (baseline/coverage/lint state, D-01..D-09 ratified as
  ADRs) — this wave specifically consumes D-02, D-03 (enacted in W05-E01-S001) and D-06 (enacted in
  W05-E04-S002), all ratified in W00-E02-S003.

## Exit criteria

- AR-01's `ApplicationModel` lifecycle, owner-bound `Registrar`, per-registry ownership wrappers
  (including the previously-zero-ownership-check `authz.Registry` permission registration), snapshot
  immutability, deterministic model hash, race safety, and the legacy compatibility adapter are all
  in place and evidenced — PLAN AR-01 T1-T11's acceptance criteria satisfied in full.
- AR-02's typed port-key API, compiled provider graph with zero hot-path reflection, boot-time graph
  validation, three-profile projection, retirement of the hand-maintained lifecycle manifest, and the
  legacy port adapter are in place and evidenced — PLAN AR-02 T1-T7 satisfied.
- AR-03's manifest schema and derived-projection tooling (T1, T3, T4, T5) pass the golden-delta
  acceptance gate — PLAN AR-03's own framing: "this test IS the acceptance gate." T2 (OpenAPI merge)
  remains explicitly out of this wave's scope, single-owned by DX-06 (W06).
- AR-04's remaining boot-strictness tasks (T2-T5) are in place, with the shared no-op-adapter waiver
  mechanism built and consumed correctly by this wave's own T5 acceptance test — PLAN AR-04 T2-T5
  satisfied. T1 (unknown-namespace rejection) remains correctly recorded as already executed, not
  re-planned.
- AR-06's remaining tasks (T2: constructor-boundary lint; T3: `kernel/kernel.go` audit) are in
  place and evidenced — PLAN AR-06 T2-T3 satisfied. T1 (the `orgAncestry` closure fix) remains
  correctly recorded as already executed, not re-planned.
- SEC-04's bounded LRU cache, epoch-based cross-pod invalidation (D-06), singleflight
  miss-collapsing, decision provenance metadata, and prod-config gating are in place and evidenced —
  PLAN SEC-04 T1-T6 satisfied, and DATA-07 T4's cache-invalidation acceptance criterion is closed by
  this wave's own AC per `impl/analysis/wave-allocation-detail.md`'s explicit note.
- FBL-01's nine-package re-home to `foundation/` is complete, the `kernel/mfa` forwarding shim is in
  place, depguard and `scripts/lint_boundaries.sh` are extended to reject the re-homed import paths,
  and wowsociety's identity/authz suite is green on the new `foundation/mfa` path (or the shim during
  the grace window) — MATRIX CS-01's acceptance bar satisfied.

## Dependencies

Depends on W03 (full-wave entry gate, specifically W03-E01 acceptance) and, at programme baseline,
on W00's D-01..D-09 ADR ratification. Internally, W05-E02 and W05-E03 depend on W05-E01 (AR-02's
`Registrar` reuse and AR-03's manifest-consumes-model dependency both require AR-01's T1/T2 to have
landed); W05-E05 depends on both W05-E01 and W05-E02 completing first, per MATRIX CS-01's own
"Dependencies: AR-01/02 first (re-homing mid-registration-rework causes double churn)." See
`dependencies.md` for the full upstream/downstream/internal detail.

## Assumptions

- No W05 story other than W05-E01-S001 (D-02, D-03) and W05-E04-S002 (D-06) enacts a D-0N
  architecture decision — confirmed by scanning `requirement-inventory.md` §B for any D-0N row
  targeting AR-02, AR-03, AR-04, or FBL-01 specifically: none exists. Only AR-01 (D-02, D-03) and
  SEC-04 (D-06) have D-0N rows. This is confirmed, not assumed, from the source text.
- AR-04 T1 (unknown-namespace rejection at boot) and AR-06 T1 (the `orgAncestry` closure fix) are
  already executed and verified twice per `requirement-inventory.md`'s AR-04 and AR-06 rows. This
  wave's E03-S002 and E04-S001 scope only the remainder (AR-04 T2-T5; AR-06 T2-T3) and do not
  re-plan or re-implement the executed tasks.
- AR-03 T2 (the OpenAPI merge fix) is explicitly single-owned by DX-06 per `requirement-inventory.md`
  ("T2 = DX-06 duplicate → single owner DX-06"), itself W06 scope. This wave's E03-S001 covers AR-03
  T1, T3, T4, T5 only and records T2 as an out-of-scope cross-reference, not silently drops it.
- AR-04 T5's waiver mechanism is a forward-shared primitive: `impl/analysis/wave-allocation-detail.md`
  states "T5 builds the shared waiver mechanism consumed by SEC-06/DX-07." This wave's E03-S002
  builds the mechanism and records the forward dependency by ID; it does not require SEC-06's or
  DX-07's own stories to exist yet.
- FBL-01's wowsociety impact is real but bounded to `kernel/mfa` (5 identity-module files) per
  REVIEW §J/§O/§P — recorded as PROD-02 coordination in `requirement-inventory.md` §D. This wave's
  E05 delivers the framework-side re-home and the forwarding shim; it does not perform wowsociety's
  own code migration (product-level, out of framework scope per mandate §2.3), but does require
  wowsociety's identity/authz suite to run green against the new path or the shim as its own
  acceptance bar.

## Risks

See `risks.md`. Headline risks: FBL-01 is explicitly "the largest single architectural correction"
and "must precede v1 stabilisation" (MATRIX CS-01) — any slip here delays the whole programme's
kernel-surface freeze; AR-01 T5 (authz permission registration) is flagged High risk as "the
actual security boundary" and "only registry with zero existing ownership check"; AR-01 T6 carries
an explicit under-scoping risk ("easy to under-scope") across ~9+ declaration classes; SEC-04 T4 was
"Highest-risk task" with an open architecture decision, now resolved by D-06 but still carrying
implementation risk around correctly wiring epoch bumps into every framework-side mutation path.

## Quality gates

- Every AR-01 task's acceptance criterion that names an adversarial test (T2's compile-fail
  fixture, T3-T6's per-class adversarial suites, T8's post-seal mutation rejection test, T9's
  hash-determinism test, T10's race test) is proven with that named test as evidence, not asserted
  from code review alone.
- AR-02 T3's zero-reflection-on-hot-path claim is proven by benchmark and static lint, not
  inspection.
- AR-03 T3's golden-delta test is treated as the acceptance gate itself, per PLAN's own framing —
  no AR-03 story may close without it passing.
- AR-04 T5's readiness-waiver integration matrix (profile × waiver × adapter-real/no-op) is proven
  by the named integration test, and the waiver mechanism's shared-consumer contract (SEC-06/DX-07)
  is documented, not merely implemented ad hoc.
- SEC-04 T4's cross-pod epoch-bump claim is proven by the named simulated cross-pod test, and the
  DATA-07 T4 cache-invalidation AC closure is recorded explicitly, not implied.
- FBL-01's acceptance bar (MATRIX CS-01) is proven by the named boundary-lint fixture (fails today
  against the nine packages; passes after re-home) and by wowsociety's identity suite running green
  on the new `foundation/mfa` path or the shim.

## Required artifacts

- AR-01: the `ApplicationModel` type and lifecycle skeleton; the `Registrar` capability type; the
  per-registry ownership wrappers (resource, rules, authz, and the ~9+ remaining declaration
  classes); the snapshot-immutability conversion; the model-hash function; the legacy adapter.
- AR-02: the `port.Key[T]` API and four generic free functions; the compiled provider graph; the
  boot-time graph validator; the three-profile projection compiler; the legacy port adapter.
- AR-03: the manifest schema definition; the projection-derivation tooling; the duplicate-identity/
  omitted-projection lint rule; the documentation/test/manifest export projections.
- AR-04: the duplicate-collector rejection; the empty-required-fragment rejection; the post-seal
  config/namespace/collector rejection; the explicit optional-capability waiver mechanism.
- AR-06: the constructor-boundary lint tool; the `kernel/kernel.go` audit report.
- SEC-04: the `golang-lru`-backed bounded cache; the `authz_epoch` table and epoch-bump wiring;
  the singleflight miss-collapse; the decision-provenance metadata; the prod-config gate.
- FBL-01: the `foundation/` package tree (9 packages `git mv`'d); the `kernel/mfa` forwarding shim;
  the extended depguard rule; the extended `scripts/lint_boundaries.sh` allowlist.

## Required evidence

- AR-01: per-task adversarial-fixture test output for T2-T6, T8; hash-determinism and
  hash-sensitivity test output (T9); race-detector output (T10); legacy-adapter integration-test
  output (T11); wowsociety module-contract-test regression output.
- AR-02: port-API unit-test output; registrar-forge compile-fail fixture output; hot-path
  no-reflection benchmark output; boot-graph-validation adversarial-suite output; three-profile
  projection integration-test output; lifecycle-lint regression output; legacy-port-adapter
  integration-test output.
- AR-03: manifest-schema-fixture round-trip test output; golden-declaration-delta test output;
  duplicate-identity/omitted-projection lint fixture output; full-projection golden test output.
- AR-04: per-collector adversarial-fixture output; empty-required-fragment fixture output; post-seal
  config-rejection regression output; prod/no-op-adapter/waiver integration-matrix test output.
- AR-06: constructor-boundary lint adversarial-fixture output; the `kernel/kernel.go` audit report.
- SEC-04: bounded-cache insert/race test output; eviction-metrics test output; singleflight test
  output; simulated cross-pod epoch test output; decision-provenance test output; prod-config
  negative-test output.
- FBL-01: the boundary-lint fixture's before/after run output (fails against the nine packages
  today, passes after); wowsociety build and identity/authz suite run output on the new
  `foundation/mfa` path or the shim.

## Expected implementation outcome

A framework whose module-registration surface is capability-secured rather than string-compared;
whose port wiring is typed, boot-validated, and free of hot-path reflection; whose route/permission/
schema surface is derived from one authoritative manifest instead of hand-duplicated; whose boot
sequence fails loudly on any unowned or unwaived configuration rather than silently proceeding; whose
authorization cache is bounded and correctly invalidated across pods instead of unbounded and
convention-dependent; and whose kernel package surface is finally the small, stable set the
framework's own architecture intends — with the nine misplaced packages correctly re-homed before
that surface locks in for v1 stabilisation.

## Acceptance authority

Framework architecture lead — per PLAN §5.1's own "Accountable role: framework architecture lead"
for PF-ARCH, applied uniformly across AR-01/02/03/04/06; SEC-04 (PF-SEC, product-security lead per
§5.2) and FBL-01 (REVIEW, no PLAN §5.x accountable-role table entry) are co-accepted alongside the
framework architecture lead given their tight coupling to this wave's registration-model and
layering work — recorded here as an explicit joint-acceptance arrangement for this wave, not a
silent substitution of PLAN's own SEC-04 accountability.

## Closure conditions

All exit criteria satisfied; all five epics' `closure-report.md` accepted; `waves/index.md`'s W05
row updated to reflect `accepted` status; no unresolved regression from the ownership-model rework
or the kernel re-home; FBL-01's wowsociety `kernel/mfa` migration coordination (PROD-02) is recorded
with a clear pointer for the product-side migration, not silently treated as this wave's own
responsibility to execute; the AR-01 T6 under-scoping risk and the SEC-04 T4 epoch-wiring risk are
each explicitly resolved or recorded as accepted residual risk before this wave is marked `accepted`.
