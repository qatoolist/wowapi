---
id: W02-E05
type: epic
title: Production seed-sync
status: planned
wave: W02
owner: unassigned
reviewer: unassigned
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-12
source_requirements:
  - FBL-02
depends_on: []
stories:
  - W02-E05-S001
decisions: []
risks:
  - RISK-W02-004
---

# W02-E05 — Production seed-sync

## Epic objective

Close FBL-02: build the production catalog seed-sync path that does not exist today — an
idempotent, RLS-respecting, versioned-manifest-driven, dry-run-capable, audited mechanism for
populating a production catalog database — and wire readiness so that a production-profile boot
against an empty catalog database no longer silently produces a deny-everything system, but instead
fails readiness with a named check until seed-sync has run, and reports the seed/catalog hash once
it has.

## Problem being solved

`requirement-inventory.md` §B records: "FBL-02 | Production seed-sync path (PF-9) | IMPL | P0-prod |
planned | W02-E05-S001 | CS-21 acceptance bar fixed; design detail = story investigation task."
MATRIX CS-21's evidence states the gap exactly: "**no production seed-sync path at all** (wowsociety
PF-9: prod boots with deny-everything catalogs — prod-blocking, never in the original 38)." This is
a production-blocking gap, not a theoretical one: a fresh production deployment boots, passes
readiness, and serves traffic against empty authorization/policy catalogs — every permission check
silently denies, and nothing in the boot or readiness path names the cause. CS-21's evidence
refinement locates the defect precisely: "framework readiness mechanism itself is correct and
fail-closed — `kernel/httpx/health.go:52-79` runs each check with a 3s timeout, 503 on any failure,
reports `config_fingerprint`; `app/health.go:9-14` documents DB/migration checks as a *comment-only
contract* supplied via `extra`. The defect is thus precisely located: contract-by-comment at the
seam + template omission at the product end" — and, for this epic's FBL-02 half specifically, the
total absence of any sync path that could populate the catalogs in the first place.

CS-21 fixes the acceptance bar while explicitly deferring the design: "**FBL-02**: a `wowapi seed
sync --env prod` path (idempotent, RLS-respecting, versioned catalog manifests, dry-run + audit) —
design detail to be ratified in Phase 5, but the acceptance bar is fixed now: *a prod-profile boot
on an empty catalog DB reaches readiness only after seed-sync has run, and the readiness payload
reports the seed/catalog hash*."

## Scope

- Design investigation resolving the catalog manifest format, versioning scheme, CLI command shape,
  idempotency mechanism, RLS/role posture, dry-run output format, and audit-record integration —
  the design detail CS-21 explicitly defers (S001's first task).
- The seed-sync path itself: idempotent, RLS-respecting sync of catalog data from versioned
  manifests, with a dry-run mode and an audit record per run (S001).
- Readiness wiring: a prod-profile boot against an empty catalog database fails readiness with a
  named check until seed-sync has run; the readiness payload reports the seed/catalog hash (S001).

## Out of scope

- **DX-07's readiness-truthfulness tasks (T1–T4)** — migration-currency readiness checks, seed/rule/
  model-hash checks beyond the seed/catalog hash this epic's acceptance bar names, `config doctor`,
  and prod-profile capacity/backpressure enforcement. CS-21 covers both findings, but DX-07 is
  targeted at `W04-E04-S003` per `requirement-inventory.md`; this epic implements only the FBL-02
  half. Where the two halves touch (the readiness payload), this epic adds exactly the seed/catalog-
  hash reporting its own acceptance bar requires, built compatibly with — not blocking on — DX-07's
  later work.
- **AR-04 T5's waiver mechanism** — cross-referenced by CS-21 for DX-07's enforcement half; not
  consumed by FBL-02's scope. W05-E03-S002's concern.
- **wowsociety's identical readiness gap** (`cmd/api/main.go:240-243`, per CS-21: "backport after
  T1; PF-9 is *its* finding, closing it closes the register entry (FBL-03)") — product-level,
  tracked as PROD-03/FBL-03 coordination items in `requirement-inventory.md` §D and W01-E04-S002's
  register-reconciliation scope. This epic delivers the framework capability; the product backport
  and register closure are not this epic's closure conditions.
- **Seed *content* authoring** — what the correct production catalog entries for any given
  deployment are is a deployment-operations concern; this epic delivers the mechanism that syncs
  declared content, not the content itself.

## Source requirements

FBL-02. Cross-referenced closure spec: MATRIX CS-21 (the FBL-02 portion; DX-07's portion is
W04-E04-S003's). FBL-02 is a REVIEW finding with no PLAN §5 T-row table — CS-21's prose "Fix" and
fail-first sections are the task-level source of record for this epic, and the acceptance bar is
fixed verbatim there.

## Architectural context

The framework's readiness mechanism (`kernel/httpx/health.go`) is already correct and fail-closed
per CS-21's evidence refinement — each check runs with a 3s timeout, any failure returns 503, and
the payload reports `config_fingerprint`. What is missing is (a) any mechanism at all for populating
production catalog data (seeds today are a dev/test concern), and (b) a readiness check that knows
whether that population has happened. The seed-sync path is expected to live alongside the existing
CLI surface (CS-21's own sketch names `wowapi seed sync --env prod`), consume versioned catalog
manifests whose format is this epic's central design question, respect RLS (the sync must not
casually run as a superuser bypassing the very tenancy controls the catalogs configure), and write
an audit record — plausibly through the existing `kernel/audit` infrastructure that DATA-06/DATA-08
already exercise, though that integration choice is explicitly a design-investigation output, not a
pre-made decision.

A naming caution recorded here so no reader conflates two manifest concepts inside this same wave:
DATA-09's *migration manifest* (W02-E01, classifying schema migrations) and FBL-02's *catalog
manifest* (this epic, declaring versioned seed content) are different artifacts with different
schemas, consumers, and lifecycles. The source draws no dependency between them, and this epic does
not depend on W02-E01 — see "Dependencies."

## Included stories

- **W02-E05-S001 — prod-seed-sync-path** (FBL-02, design + implement per CS-21's fixed acceptance
  bar): the design-investigation task (catalog manifest format and its sibling questions) sequenced
  first, then the seed-sync path, manifest schema, readiness wiring, and audit-record production,
  closed by independent review (P0-prod).

## Dependencies

None upstream within W02. This epic is explicitly independent of W02-E01's online-migration
protocol: although both epics introduce a "manifest" concept, DATA-09's migration manifest
(classifying schema-change risk) and FBL-02's catalog manifest (declaring versioned seed content)
are unrelated designs, and neither `requirement-inventory.md`, MATRIX CS-21, nor
`wave-allocation-detail.md` records any dependency from FBL-02 onto DATA-09 — this epic may execute
before, after, or in parallel with E01–E04 (per `../../dependencies.md`: "W02-E03, W02-E04, W02-E05
are independent of W02-E01, W02-E02, and of each other"). Depends only on W00's exit gate at wave
scope. Downstream: the wowsociety readiness backport (PROD-03) and PF-9 register closure (FBL-03,
W01-E04-S002) both consume this epic's delivered capability but are not gated inside this wave.

## Risks

RISK-W02-004 (the design investigation may surface a need for new infrastructure, a new dependency,
or a design pattern not anticipated by this wave's planning) originates at wave scope and lands
entirely within this epic's single story. See `risks.md` for the epic-scoped elaboration and one
additional epic-specific risk (RLS-posture tension in a prod bootstrap context).

## Required decisions

None in this programme's D-0N ADR sense — no D-01..D-09 decision targets FBL-02 (confirmed against
`requirement-inventory.md` §B and REVIEW §F/§U; see `../../wave.md` "Assumptions"). However, the
*design content* of the catalog manifest format is explicitly unresolved by the source — MATRIX
CS-21: "design detail to be ratified in Phase 5" — and S001's design-investigation task exists
specifically to resolve it. That resolution is a story-internal design-investigation output, not a
new ADR. **Process safeguard, stated explicitly:** if the investigation concludes that a decision of
D-0N-caliber significance is required (for example, a new framework-wide manifest/versioning
convention intended to outlive this story, or a new external dependency), that decision must be
escalated for ADR treatment through the programme's decision register rather than silently decided
inside this story — per mandate §11.8 ("Do not bury decisions only in prose") and §18 ("Do not
silently resolve ambiguous architecture decisions").

## Epic acceptance criteria

- **AC-W02-E05-01**: The design investigation is complete before implementation: every design
  question named in S001's `plan.md` "Unresolved questions" (manifest format, versioning scheme,
  CLI shape, idempotency mechanism, RLS posture, dry-run format, audit integration) has a
  documented decision with rationale, recorded before any implementation task began.
- **AC-W02-E05-02**: The seed-sync path exists and is idempotent, RLS-respecting, versioned-
  manifest-driven, dry-run-capable, and audited — each property proven by its own test per S001's
  acceptance criteria, not asserted from code review.
- **AC-W02-E05-03**: CS-21's fixed acceptance bar holds verbatim: a prod-profile boot on an empty
  catalog DB reaches readiness only after seed-sync has run, and the readiness payload reports the
  seed/catalog hash. The fail-first half is proven too: before the fix, the same boot silently
  reaches a deny-everything ready state; after, it returns a named readiness failure until sync.
- **AC-W02-E05-04**: S001 has passed independent review per mandate §14 (P0-prod story),
  specifically confirming the design-investigation decisions were documented before implementation
  began (not backfilled), and that no design question was silently resolved without a recorded
  rationale.

## Closure conditions

S001 reaches `accepted` (satisfying its own `closure.md`); AC-W02-E05-01 through AC-W02-E05-04 are
all satisfied; `closure-report.md` for this epic is completed with reviewer conclusion and
acceptance date; RISK-W02-004's outcome is recorded (design landed within anticipated scope, or the
scope expansion was split out per its contingency — either way documented, not silent); the
wowsociety backport (PROD-03) and PF-9 register closure (FBL-03) pointers are recorded for their
owning tracks, not silently dropped.
