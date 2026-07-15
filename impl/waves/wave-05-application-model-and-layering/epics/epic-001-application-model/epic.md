---
id: W05-E01
type: epic
title: Application model
status: planned
wave: W05
owner: unassigned
reviewer: unassigned
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-12
source_requirements:
  - AR-01
depends_on: []
stories:
  - W05-E01-S001
  - W05-E01-S002
  - W05-E01-S003
  - W05-E01-S004
decisions:
  - D-02
  - D-03
risks:
  - RISK-W05-001
  - RISK-W05-002
---

# W05-E01 — Application model

## Epic objective

Replace the framework's mutable, unowned module-registration mega-context with an immutable
`ApplicationModel` compiled via `collect → validate → seal → expose read-only snapshot`, secured by
an owner-bound `Registrar` capability type that makes cross-module ownership claims structurally
impossible rather than string-compared — so that every later wave-05 epic (typed ports, the
authoritative manifest, kernel re-home) and every downstream module in wowapi and wowsociety
registers its declarations through a security boundary the framework itself enforces, not one a
module's own good behaviour merely happens to respect.

## Problem being solved

`requirement-inventory.md` row AR-01 states: "Ownership-bound `ApplicationModel` (T1–T11) | IMPL |
P1 | planned | W05-E01-S001..S004 | Core; D-02/D-03 resolve its open questions." PLAN's own AR-01
directive requirement frames the target: "immutable `ApplicationModel` compiled from ownership-bound
module declarations; registration APIs never accept an arbitrary owner string from module code;
every collector follows `collect → validate → seal → expose read-only snapshot`; post-seal mutation
errors/panics, never silently no-ops." Today, per PLAN's own evidence, the framework's
`authz.Registry.Register(p Permission)` "currently has no owner parameter at all" — the widest of
six registration surfaces PLAN identifies as string-compared or entirely unchecked, not
structurally secured. This epic closes that gap for all ~9+ declaration classes the framework
registers at boot.

## Scope

- The `ApplicationModel` type and its `collect → validate → seal → expose read-only snapshot`
  lifecycle skeleton, with post-seal mutation erroring (never silently no-op), panicking only under
  an explicit dev/test build tag (S001, PLAN AR-01 T1).
- The owner-bound `Registrar` capability type, mintable only by the compiler from
  `Manifest.ID`/`Module.Name()`, with an unexported seal method preventing module-code construction
  or type-assertion for another owner (S001, PLAN AR-01 T2; enacts D-02 and D-03).
- Owner-bound registrar wrappers for `resource.Registry`, `rules.Registry`, and — critically —
  `authz.Registry` permission registration, which today has zero ownership check at all (S002, PLAN
  AR-01 T3-T5).
- Owner-bound registrar wrappers for the remaining ~9+ declaration classes: events, jobs, workflow
  actions, providers, templates, health checks, migrations, seeds, OpenAPI (S002, PLAN AR-01 T6).
- Conversion of every exported registry reader to cloned/immutable data, so no exported reader
  returns a backing map/slice (S003, PLAN AR-01 T7).
- Rejection of Context retention after `Register()` returns — a module retaining a registrar
  post-boot gets an explicit error on mutation, never a silent no-op or a production panic (S003,
  PLAN AR-01 T8).
- A deterministic model hash emitted at startup/readiness, and race tests proving no runtime
  mutation of the sealed model (S003, PLAN AR-01 T9-T10).
- The legacy adapter wrapping the current `module.Module`/`Context` so existing modules (wowapi
  internal and wowsociety) compile and boot unchanged through the ownership-bound registration
  surface (S004, PLAN AR-01 T11).

## Out of scope

- **AR-02's typed port-key API and provider graph** — W05-E02's scope. This epic's `Registrar` type
  (T2) is the load-bearing prerequisite AR-02 reuses; this epic does not itself build the port
  system.
- **AR-03's manifest and derived-projection tooling** — W05-E03's scope. This epic's sealed
  `ApplicationModel` is the model AR-03's manifest consumes; this epic does not itself derive
  projections from declarations.
- **FBL-01's kernel package re-home** — W05-E05's scope, and explicitly sequenced after this epic
  per MATRIX CS-01 ("Dependencies: AR-01/02 first").
- **wowsociety's own cleanup of `internal/modules/policy/pack.go:334-338`'s dead retained-registrar
  field or `rulepoints.go:218`'s hardcoded owner-string literal** — PLAN's own wowsociety-impact note
  states this cleanup is "low-risk and can happen on wowsociety's own schedule," not required before
  or during this epic's landing. This is product-level code, out of framework scope per mandate
  §2.3.

## Source requirements

AR-01 (T1-T11). D-02 and D-03 (referenced, not authored — ratified in W00-E02-S003).

## Architectural context

AR-01 is, per this wave's own `wave.md` rationale, the load-bearing prerequisite for every other
W05 epic: PLAN's own PF-ARCH cross-cutting note states "AR-01 T1/T2 are the load-bearing
prerequisite for AR-02's `Registrar` reuse and AR-03's manifest-consumes-model dependency." Within
this epic, the eleven tasks form a dependency chain matching PLAN's own task table: T1 (the lifecycle
skeleton) and T2 (the `Registrar` capability type) are the foundation every other task depends on;
T3-T6 (the per-registry ownership wrappers) depend on T1+T2; T7 (snapshot immutability) depends on
T3-T6; T8 (post-seal rejection) depends on T1+T2 directly (parallel-safe with T3-T7); T9-T10 (model
hash, race tests) depend on the full T1-T9/T1-T8 surface; T11 (the legacy adapter) depends on
T1-T10 in full. This epic's four stories group these by phase-cluster per
`impl/analysis/wave-allocation-detail.md`'s canonical allocation: S001 (T1, T2) is the foundation;
S002 (T3, T4, T5, T6) is the per-registry ownership closure; S003 (T7, T9, T10, plus T8 post-seal
rejection) is the immutability/determinism/race-safety closure; S004 (T11) is the compatibility
story. Two unresolved design questions PLAN's own cross-cutting notes (5) and (6) flag — "do all
AR-01 per-subsystem registrars share one `Registrar` type... or does each get a distinct type" and
"should post-seal mutation panic in production builds, or only error" — are exactly the questions
D-02 and D-03 close, per REVIEW §F items 3 and 4. This epic's S001 accordingly carries a
`decisions/` directory referencing both.

## Included stories

- **W05-E01-S001 — model-and-registrar-capability** (PLAN AR-01 T1, T2; enacts D-02, D-03): the
  `ApplicationModel` lifecycle skeleton and the owner-bound `Registrar` capability type.
- **W05-E01-S002 — registry-ownership** (PLAN AR-01 T3, T4, T5, T6): owner-bound registrar wrappers
  for resource, rules, authz-permission, and the ~9+ remaining declaration classes.
- **W05-E01-S003 — snapshots-hash-race** (PLAN AR-01 T7, T9, T10, plus T8): snapshot immutability,
  post-seal mutation rejection, deterministic model hash, race safety.
- **W05-E01-S004 — legacy-adapter** (PLAN AR-01 T11): the compatibility adapter for existing modules.

## Dependencies

No dependency on any other W05 epic — this epic is W05's own foundation, matching AR-01's role at
wave scope (see `wave.md` "Dependencies"). Depends on this wave's own entry gate (W03-E01
acceptance). Downstream within this wave: W05-E02 and W05-E03 depend on this epic; W05-E05 depends
on this epic (and on W05-E02). See `dependencies.md`.

## Risks

RISK-W05-001 (AR-01 T5's authz-registration ownership gap — "the actual security boundary," "widest
gap of the six") and RISK-W05-002 (AR-01 T6's explicit under-scoping risk across ~9+ declaration
classes) both originate at wave scope and land entirely within this epic's S002. See `risks.md` for
the epic-scoped elaboration.

## Required decisions

D-02 (single `Registrar` type with per-subsystem typed keys) and D-03 (post-seal mutation errors in
production, panics only under an explicit dev/test build tag) — both already ratified in
W00-E02-S003, referenced (not re-decided) in S001's `decisions/index.md`.

## Epic acceptance criteria

- **AC-W05-E01-01**: The `ApplicationModel` compiles via `collect → validate → seal → expose
  read-only snapshot`; post-seal, further registration calls error in production builds (panic only
  under an explicit dev/test build tag, per D-03); the `Registrar` capability type is mintable only
  by the compiler and module code cannot construct or type-assert one for another owner, per D-02's
  single-type-with-typed-keys design.
- **AC-W05-E01-02**: Every declaration class in AR-01's own acceptance gate — resource, rules,
  authz-permission registration (previously zero-ownership-checked), and the ~9+ remaining classes —
  is ownership-checked, proven by a table-driven adversarial suite with one fixture per class.
- **AC-W05-E01-03**: No exported registry reader returns a backing map/slice; a module retaining a
  registrar post-boot receives an explicit error on mutation, never a silent no-op or production
  panic; two identical compiles emit a byte-identical model hash, one changed declaration emits a
  different hash; `go test -race` is clean on concurrent legitimate reads.
- **AC-W05-E01-04**: Existing modules (wowapi-internal and wowsociety) boot unchanged through the
  legacy adapter, proven by existing contract tests passing unmodified through the legacy path; the
  adapter itself derives owner from `Module.Name()` and routes through the same owner-bound
  registrars as the non-legacy path — it does not bypass T2-T6's ownership checks.
- **AC-W05-E01-05**: All four stories have passed independent review per mandate §14, with S001 and
  S002 specifically checked given their High-risk task content (T2's security-boundary type; T5's
  previously-zero-ownership-check gap; T6's under-scoping risk).

## Closure conditions

All four stories reach `accepted` (each satisfying its own `closure.md`); AC-W05-E01-01 through
AC-W05-E01-05 above are all satisfied; `closure-report.md` for this epic is completed with reviewer
conclusion and acceptance date; RISK-W05-001 and RISK-W05-002 are each resolved (adversarial test
genuinely proven; declaration-class enumeration genuinely complete) or explicitly recorded as
accepted residual risk — neither is silently dropped at closure.
