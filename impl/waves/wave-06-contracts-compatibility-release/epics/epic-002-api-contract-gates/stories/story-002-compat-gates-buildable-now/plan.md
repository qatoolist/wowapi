---
id: PLAN-W06-E02-S002
type: plan
parent_story: W06-E02-S002
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-14
---

# Plan — W06-E02-S002

Per mandate §8.5. Confirmed facts, planned changes, and implementation assumptions are distinguished
explicitly below.

## Proposed architecture

Six independent CI-gate mechanisms, each wired into the CI pipeline as its own job or test extension:
a Go API diff job, a compile-matrix job, a config-compat gate, an extended migration-reversibility test,
an architecture-smoke job (against the REL-01 candidate image), and a cross-reference to REL-01's
existing SBOM/provenance evidence. None of the six shares implementation surface with any other beyond
the shared CI infrastructure they all run within.

## Implementation strategy

1. **T1** — Wire `golang.org/x/exp/apidiff`/`gorelease` as a CI job comparing the current public API
   surface against the previous release, classifying changes per DX-05's ratified v1/N-1 policy; write
   a seeded breaking-API fixture test.
2. **T2** — Build a compile matrix across supported Go/dependency versions; document any excluded
   version explicitly in CI configuration.
3. **T4** — Build a config schema compatibility gate against `kernel/config/schema.go`; write a seeded
   breaking-config fixture (field removal/type change) and a generated fixture migration test for
   additive optional fields.
4. **T6** — Extend `TestIntegrationMigrationsReversible` into an upgrade-from-oldest-supported drill:
   seed at the oldest supported version, migrate forward, reverse on disposable data.
5. **T8** — Build a container architecture smoke CI job that runs against the candidate image produced
   by REL-01's `build-candidate` stage (W06-E03-S001), on every published architecture.
6. **T9** — Cross-reference REL-01 T8/T9's existing SBOM/provenance/signature verification evidence as
   satisfying REL-03's own naming of the same property; no separate implementation.

## Expected package or module changes

New CI job configurations for T1, T2, T4, T8; an extension to the existing migration-reversibility test
file for T6; no new package for T9 (cross-reference only).

## Expected file changes where determinable

- New CI workflow configuration for the API diff job (T1).
- New CI workflow configuration for the compile matrix (T2).
- New CI workflow configuration and fixture files for the config-compat gate (T4).
- Extension to the existing `TestIntegrationMigrationsReversible` test file (T6).
- New CI workflow configuration for the architecture-smoke job (T8).
- No new file for T9 beyond a documentation cross-reference.

## Contracts and interfaces

T1's API diff operates against the Go public API surface (exported symbols) — no new contract is
defined, the diff tool observes existing contracts. T4's config-compat gate operates against
`kernel/config/schema.go` as the existing source of truth.

## Data structures

None new.

## APIs

None affected — these are CI-time gates observing existing APIs, not changing them.

## Configuration changes

None to runtime configuration; CI configuration changes only.

## Persistence changes

None.

## Migration strategy

T6 is itself a migration-testing extension; no new migration strategy beyond what already exists.

## Concurrency implications

None.

## Error-handling strategy

Each gate must fail with a clear, specific error identifying exactly what changed/broke, per `story.md`
"Observability considerations."

## Security controls

T9's SBOM/provenance/signature verification is a supply-chain security control, shared with REL-01;
no new control beyond what REL-01 T8/T9 already provides.

## Observability changes

None beyond each gate's own clear-failure-reporting requirement.

## Testing strategy

- T1: seeded breaking-API fixture, confirming the diff gate fails it.
- T2: compile-matrix CI run, confirming explicit (not silent) version exclusions.
- T4: seeded breaking-config fixture (fail) and additive-optional-field fixture (pass).
- T6: migration upgrade-drill test at oldest-supported version, forward then reverse on disposable data.
- T8: architecture-smoke test per published architecture against the candidate image.
- T9: no new test — cross-reference REL-01 T8/T9's existing golden-failure tests.

## Regression strategy

Once wired into CI, all six gates become ongoing regression guards against their respective
compatibility classes (API surface, compile-version support, config schema, migration reversibility,
architecture bootability, supply-chain provenance).

## Compatibility strategy

T1, T4, T6 are themselves compatibility-enforcement mechanisms, each consuming DX-05's already-ratified
v1/N-1 policy where relevant (T1 explicitly; T6 implicitly via "oldest supported version").

## Rollout strategy

Six independent CI-gate additions; each may land and be made required independently as it is proven
working, rather than requiring all six to land simultaneously.

## Rollback strategy

If any gate produces false positives in CI, it can be temporarily demoted to advisory (non-blocking)
while the flake/false-positive is diagnosed, per standard CI-gate rollback pattern.

## Implementation sequence

T1, T2, T4, T6 have no ordering dependency on each other and may proceed in parallel. T8 should be
sequenced after or in parallel with W06-E03-S001's own REL-01 T6/T7 work (see `story.md`
"Assumptions"). T9 requires no new implementation sequencing — it is a documentation cross-reference,
addable at any point once REL-01 T8/T9's evidence exists.

## Task breakdown

- **W06-E02-S002-T001** — Go public API diff (REL-03 T1).
- **W06-E02-S002-T002** — Module compile matrix (REL-03 T2).
- **W06-E02-S002-T003** — Config schema compatibility (REL-03 T4).
- **W06-E02-S002-T004** — Migration upgrade-from-oldest-supported drill (REL-03 T6).
- **W06-E02-S002-T005** — Container architecture smoke (REL-03 T8).
- **W06-E02-S002-T006** — SBOM/provenance/signature verification fold-in (REL-03 T9).
- **W06-E02-S002-T007** — Independent review.

## Expected artifacts

The Go API diff CI job; the module compile matrix CI configuration; the config schema compatibility
gate; the extended migration upgrade-drill test; the container architecture smoke CI job; the REL-03 T9
cross-reference.

## Expected evidence

Seeded breaking-API-fixture test output; compile-matrix CI run output; seeded breaking-config-fixture
test output; migration upgrade-drill test output; architecture-smoke test output; the shared REL-01
T8/T9 evidence reference.

## Unresolved questions

- Exact coordination point between this story's T005 (architecture smoke) and W06-E03-S001's own
  build-candidate delivery — whether T005 waits for W06-E03-S001's `accepted` status or develops against
  a stub candidate image reconciled later.
- Exact set of "supported Go/dependency versions" for T002's compile matrix — not specified by any
  source document beyond "supported," to be determined at implementation time consulting the
  framework's own stated Go-version support policy (if one exists) or established at this task's own
  implementation time if not.

## Approval conditions

This plan is approved for implementation once the owner and reviewer are assigned; T005's specific
coordination point with W06-E03-S001 should be resolved before T005 itself begins, though T001-T004,
T006 may begin immediately.
