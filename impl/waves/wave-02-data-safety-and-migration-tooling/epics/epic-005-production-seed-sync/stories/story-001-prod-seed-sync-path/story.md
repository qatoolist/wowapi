---
id: W02-E05-S001
type: story
title: Production catalog seed-sync path
status: accepted
wave: W02
epic: W02-E05
owner: W02SeedRerun
reviewer: W02ReviewGate
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-13
source_requirements:
  - FBL-02
depends_on: []
blocks: []
acceptance_criteria:
  - AC-W02-E05-S001-01
  - AC-W02-E05-S001-02
  - AC-W02-E05-S001-03
  - AC-W02-E05-S001-04
  - AC-W02-E05-S001-05
  - AC-W02-E05-S001-06
artifacts:
  - ART-W02-E05-S001-001
  - ART-W02-E05-S001-002
  - ART-W02-E05-S001-003
  - ART-W02-E05-S001-004
  - ART-W02-E05-S001-005
  - ART-W02-E05-S001-006
  - ART-W02-E05-S001-007
evidence:
  - EV-W02-E05-S001-001
  - EV-W02-E05-S001-002
  - EV-W02-E05-S001-003
  - EV-W02-E05-S001-004
  - EV-W02-E05-S001-005
  - EV-W02-E05-S001-006
decisions:
  - ART-W02-E05-S001-001
risks:
  - RISK-W02-004
---

# W02-E05-S001 — Production catalog seed-sync path

## Story ID

W02-E05-S001

## Title

Production catalog seed-sync path

## Objective

Design (investigation-first) and implement a `wowapi seed sync --env prod`-shaped path — or
whatever command shape the design investigation concludes — that is idempotent, RLS-respecting,
driven by versioned catalog manifests, supports dry-run, and produces an audit record; and wire
readiness so that a production-profile boot against an empty catalog database fails readiness with
a named check until seed-sync has run, with the readiness payload reporting the seed/catalog hash
once it has.

## Value to the framework

A framework that silently boots a production deployment into a deny-everything state — every
permission check failing against empty catalogs, with readiness reporting healthy — is not
production-deployable, whatever its other qualities. MATRIX CS-21 grades this P0-prod for exactly
that reason: the gap is "prod-blocking, never in the original 38." Closing it gives the framework a
first-class production bootstrap story: catalog content is declared in versioned manifests, synced
idempotently under an explicit and audited role posture, previewable via dry-run, and readiness
truthfully refuses to advertise a deployment whose catalogs have never been populated. The
capability is generic (any product built on the framework needs its catalogs populated in prod),
which is why it is framework scope rather than a wowsociety-only fix — wowsociety's PF-9 finding is
the evidence, but the mechanism belongs in the kernel/CLI layer both repos consume.

## Problem statement

`requirement-inventory.md` §B row FBL-02 records: "Production seed-sync path (PF-9) | IMPL |
P0-prod | planned | W02-E05-S001 | CS-21 acceptance bar fixed; design detail = story investigation
task." MATRIX CS-21's evidence: "**no production seed-sync path at all** (wowsociety PF-9: prod
boots with deny-everything catalogs — prod-blocking, never in the original 38)." CS-21's fail-first
section names the observable defect: "prod boot with empty catalogs → currently silently
deny-everything, after: named readiness failure." There is no mechanism today to populate a
production catalog database, and nothing in the readiness path knows whether population has
happened.

## Source requirements

FBL-02. Cross-referenced closure spec: MATRIX CS-21 (FBL-02 portion only — CS-21 also covers DX-07,
which is W04-E04-S003's scope, out of scope here). FBL-02 has no PLAN §5 T-row table; CS-21's prose
"Fix"/"Fail-first" sections and its italicized fixed acceptance bar are this story's task-level
source of record.

## Current-state assessment

Per MATRIX CS-21's evidence (to be re-confirmed at this story's own execution commit):

- **No production seed-sync path at all** — this is the FBL-02 gap itself, verbatim from CS-21's
  evidence line. There is no command, job, or boot hook that populates production catalog data.
- The readiness template "registers only `\"db\"`+`\"seeds\"` checks — no migration-currency check
  despite the health contract's own doc claiming it" (CS-21 evidence — this sentence is DX-07's
  adjacent context; the migration-currency half is W04-E04-S003's scope, cited here only to locate
  the seam this story's readiness wiring touches).
- CS-21's evidence refinement, locating the seam precisely: "framework readiness mechanism itself is
  correct and fail-closed — `kernel/httpx/health.go:52-79` runs each check with a 3s timeout, 503 on
  any failure, reports `config_fingerprint`; `app/health.go:9-14` documents DB/migration checks as a
  *comment-only contract* supplied via `extra`." The readiness plumbing this story wires into is
  sound; what is missing is the seed-sync capability and a check that consults it.
- The observable failure mode: "prod boot with empty catalogs → currently silently deny-everything"
  (CS-21 fail-first) — the deployment passes readiness and serves traffic while every catalog-backed
  authorization decision denies.

**This assessment reflects the state cited in MATRIX CS-21 at the time it was written.** Per this
programme's fail-first convention, this story's first implementation step re-confirms the absence
(no seed-sync path, readiness silent on empty catalogs) at the actual start commit, and the
readiness fail-first test (boot prod-profile against empty catalogs, observe today's silent-ready
behavior) is captured as the "before" evidence prior to any fix landing.

## Desired state

CS-21's fixed acceptance bar, verbatim: "a prod-profile boot on an empty catalog DB reaches
readiness only after seed-sync has run, and the readiness payload reports the seed/catalog hash."
Concretely: a `wowapi seed sync --env prod`-shaped path exists that is "idempotent, RLS-respecting,
versioned catalog manifests, dry-run + audit" (CS-21's fix sketch); a prod-profile boot against
empty catalogs returns a named readiness failure (503, per the existing fail-closed readiness
mechanics) until sync has run; after sync, readiness passes and its payload reports the seed/catalog
hash.

## Scope

- The design investigation resolving every question CS-21 defers ("design detail to be ratified in
  Phase 5"): catalog manifest format, versioning scheme, CLI command shape, idempotency mechanism,
  seed/catalog-hash computation and payload placement, RLS/role posture, dry-run output format,
  audit-record integration. Sequenced first; all implementation tasks gated on its documented
  decisions (T001).
- The seed-sync core path: idempotent, RLS-respecting sync of catalog data with a dry-run mode
  (T002).
- The versioned catalog manifest schema and its parsing/validation, per T001's ratified format
  (T003).
- Readiness wiring: the empty-catalog named readiness failure, the sync-completion gate, and the
  seed/catalog-hash reporting in the readiness payload (T004).
- Audit-record production for each sync run, per T001's documented integration decision (T005).
- Independent review (P0-prod story, mandate §14) (T006).

## Out of scope

- **DX-07 T1–T4** (migration-currency readiness check, seed/rule/model-hash checks beyond this
  story's own seed/catalog hash, `config doctor`, prod-profile capacity/backpressure enforcement) —
  W04-E04-S003's scope, per `requirement-inventory.md` row DX-07. This story touches the readiness
  payload only to add its own acceptance bar's hash reporting.
- **AR-04 T5's waiver mechanism** (CS-21 cross-reference for DX-07's enforcement half) —
  W05-E03-S002's scope.
- **The wowsociety backport** (its `cmd/api/main.go:240-243` readiness gap — PROD-03) and **the PF-9
  register closure** (FBL-03, W01-E04-S002's documentation scope) — product-level coordination,
  recorded as pointers at closure, not this story's deliverables.
- **Production seed content** — what the correct catalog entries are for a given deployment is an
  operations concern; this story delivers the mechanism, not the content.
- **Any new D-0N-caliber architecture decision** — if T001 concludes one is needed, it is escalated
  per `../../epic.md`'s "Required decisions" process safeguard, not resolved silently in-story.

## Assumptions

- **The exact catalog manifest format, versioning scheme, and CLI command shape are NOT yet
  determined by the source** — MATRIX CS-21 is explicit: "design detail to be ratified in Phase 5,
  but the acceptance bar is fixed now." This is the explicit subject of this story's
  design-investigation task (T001); treating any specific format as decided before that task
  completes would violate mandate §18 ("Where implementation details cannot yet be known, state
  what must be determined during the story rather than inventing specifics"). The `wowapi seed sync
  --env prod` shape in CS-21 is a sketch, assumed indicative but not binding on T001's outcome.
- The existing readiness mechanism (`kernel/httpx/health.go:52-79`) is assumed to remain the
  fail-closed check runner this story's readiness wiring plugs into, per CS-21's evidence
  refinement — to be re-confirmed at implementation time.
- The audit record is assumed likely to integrate with the existing `kernel/audit` infrastructure
  rather than a bespoke log — but this is T001's decision to make and document, not a premise.
- "Seeds" as they exist today are assumed to be a dev/test-time concern with no prod-safe execution
  path — the precise current seed mechanism and its reuse potential is a T001 investigation input,
  to be read from the repository at story start rather than asserted here.

## Dependencies

None (`depends_on: []`). This story — and its epic — is independent of every other W02 epic,
including W02-E01's migration protocol (different manifest concepts; see epic-level
`dependencies.md`). Depends only on W00's exit gate at wave scope. Internal sequencing: T002–T005
are gated on T001's documented design decisions; T006 gates acceptance.

## Affected packages or components

New: the seed-sync command path (expected in the CLI layer, `internal/cli/`-adjacent — exact
location per T001) and the catalog manifest schema/parser. Extended: the readiness check
registration seam (`app/health.go`'s `extra`-supplied checks and/or the readiness template —
exact wiring point per T001, informed by CS-21's "contract-by-comment at the seam + template
omission at the product end" diagnosis), and (per T001's audit decision) `kernel/audit` usage. All
locations are investigation outputs, not confirmed file-level claims.

## Compatibility considerations

Additive for existing deployments: a deployment whose catalogs are already populated must not be
broken by the new readiness check — the check gates on "seed-sync has run / catalogs populated,"
which an existing healthy deployment satisfies (the exact satisfaction predicate — hash presence vs.
populated-catalog detection for pre-existing databases — is a T001 design question). The dev/test
profile boot path must remain unaffected (the named readiness failure is prod-profile behavior per
the acceptance bar). wowsociety's generated main.go has the identical gap and gains the fix by
backport (PROD-03), not automatically — same template-delivery model as FBL-09/DX-07.

## Security considerations

The RLS/role posture is this story's central security question (RISK-W02-E05-002): seed-sync
populates the catalogs that access control presupposes, on a database where they are empty. The
sync must be "RLS-respecting" (CS-21's own requirement) — a silent superuser bypass would satisfy
the mechanics while violating the intent. T001 must document which role the sync runs as and why
that posture does not undermine tenancy controls; the audit record and dry-run mode are the
compensating controls CS-21's own fix sketch builds in. Independent review (T006) specifically
checks this decision.

## Performance considerations

Catalog data is small (configuration-scale, not tenant-data-scale); no performance budget applies
beyond the existing 3s-per-readiness-check timeout (`kernel/httpx/health.go:52-79`), which the new
readiness check must respect — the check must consult a cheap indicator (e.g. the recorded
seed/catalog hash), not re-scan catalog tables expensively on every probe. Exact mechanism per
T001.

## Observability considerations

The readiness payload gains the seed/catalog hash (the acceptance bar's own requirement). The named
readiness failure for unsynced catalogs is itself the observability fix — today's failure mode is
silent. Sync runs (and dry-runs) should log their manifest version and outcome; the audit record is
the durable trail.

## Migration considerations

No schema migration is confirmed as required. If T001's design concludes a table is needed (e.g. a
sync-state/hash record), that migration is authored within this story per the repository's existing
migration conventions — and, this being W02, classified against DATA-09's manifest schema if
W02-E01-S001 has landed by then (a convenience of shared wave timing, not a dependency; this story
does not wait for it).

## Documentation requirements

- The T001 design document (manifest format, versioning, CLI shape, idempotency, RLS posture,
  dry-run format, audit integration — each with rationale) — itself a tracked artifact.
- Operator-facing documentation for the seed-sync command: how to author/version a catalog
  manifest, run a dry-run, interpret the audit record, and read the readiness payload's hash.

## Acceptance criteria

- **AC-W02-E05-S001-01**: Every design question named in `plan.md`'s "Unresolved questions" has a
  documented decision with rationale in T001's design document, recorded before any implementation
  task (T002–T005) began; any D-0N-caliber decision was escalated per the epic's process safeguard.
- **AC-W02-E05-S001-02**: The seed-sync path is idempotent and dry-run-capable: a second run against
  an already-synced database converges with no spurious writes (proven by a repeat-run test), and a
  dry-run against an unsynced database produces a change plan without writing (proven by a
  no-writes assertion).
- **AC-W02-E05-S001-03**: The sync consumes versioned catalog manifests per T001's ratified format:
  a manifest failing schema validation is rejected before any write; the applied manifest version is
  recorded and retrievable.
- **AC-W02-E05-S001-04**: The sync is RLS-respecting per T001's documented role posture: a test
  verifies the sync runs under the documented role (not an undocumented superuser bypass) and that
  tenant-scoped RLS enforcement is preserved on any tenant table the posture touches.
- **AC-W02-E05-S001-05**: CS-21's fixed acceptance bar, both halves: fail-first — a prod-profile
  boot against an empty catalog DB, before the fix, silently reaches ready (today's defect,
  captured); after the fix, the same boot returns a named readiness failure until seed-sync has
  run, and once sync has run, readiness passes with the payload reporting the seed/catalog hash.
- **AC-W02-E05-S001-06**: Each sync run (including dry-run, per T001's decision on dry-run
  auditing) produces a durable audit record per T001's documented integration decision, proven by
  an audit-row assertion test.

## Required artifacts

- The T001 design document (catalog manifest format et al.) — pre-implementation artifact.
- The seed-sync command implementation.
- The catalog manifest schema and parser/validator.
- The readiness wiring change (named check + hash reporting).
- Operator documentation.
See `artifacts/index.md`.

## Required evidence

- Design-decision record (T001).
- Idempotency repeat-run test output; dry-run no-writes test output.
- Manifest schema-validation test output (accept/reject pair).
- RLS-posture verification test output.
- Empty-catalog readiness fail-first/pass-after test output, including the hash-reporting assertion.
- Audit-record assertion test output.
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`: `story.md`
and `plan.md` complete, acceptance criteria numbered and measurable, dependencies (none) recorded,
owner/reviewer assignment pending, and — centrally for this story — the unresolved design questions
explicitly recorded in `plan.md` as T001's mandate rather than silently assumed answered.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` (as revised by T001's documented decisions — a recorded plan
revision, not a silent rewrite) or deviations are recorded in `deviations.md`; all six acceptance
criteria verified with evidence in `evidence/index.md`; `closure.md` completed; independent review
passed per mandate §14, specifically confirming T001's decisions predate implementation and the RLS
posture is genuinely justified.

## Risks

RISK-W02-004 (design investigation surfaces unanticipated infrastructure needs) and
RISK-W02-E05-002 (RLS-posture tension in the bootstrap context) — see epic-level `risks.md` for
full detail, mitigation, and contingency.

## Residual-risk expectations

Once T001's decisions land with documented rationale and T006's review passes, residual risk
concentrates in the inherent privilege of any production bootstrap path — accepted with the audit
record and dry-run mode as compensating controls (CS-21's own fix sketch includes both for this
reason). No other residual risk is expected to remain open at acceptance.

## Plan

See `plan.md`.
