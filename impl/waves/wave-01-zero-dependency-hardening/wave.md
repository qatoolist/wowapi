---
id: W01
type: wave
title: Zero-dependency hardening
status: accepted
owner: unassigned
reviewer: unassigned
priority: high
created_at: 2026-07-12
updated_at: 2026-07-13
included_epics:
  - W01-E01
  - W01-E02
  - W01-E03
  - W01-E04
depends_on:
  - W00
blocks:
  - W02
  - W03
  - W04
  - W05
  - W06
  - W07
source_requirements:
  - FBL-05
  - FBL-07
  - FBL-06
  - FBL-08
  - FBL-09
  - FBL-03
  - DX-01
  - DX-02
  - DX-05
  - T-DOC-01
  - T-TEST-01
  - CS-10
  - CS-23
  - CS-05
  - CS-08
  - CS-09
  - CS-14
  - CS-13
  - D-08
---

# W01 — Zero-dependency hardening

## Objective

Land every finding that has **no upstream dependency on the AR-01/AR-02 application-model work, the
SEC-01 session-security work, or the DATA-09 migration-tooling work** — the largest batch of
independently-shippable value in the programme. This wave covers static-analysis utilisation
(zero-cost and judged linter sets), supply-chain and pre-push hook hygiene, OTel trace/log
correlation and pgx query tracing, HTTP transport hardening, central validation enforcement,
generator correctness fixes, and a cluster of documentation/test-diagnosis cleanups.

## Rationale

`impl/index.md`'s wave map explicitly derives W01 as "everything valuable with no upstream
dependency," constructed by scanning `requirement-inventory.md` for `planned`/`partial` items whose
`Notes` column cites no blocking dependency on AR-01/AR-02/SEC-01/DATA-09. This is dependency-aware
sequencing per mandate §2.2 ("configuration foundations before configuration-dependent modules...
observability primitives before module-specific instrumentation... test infrastructure before
coverage enforcement") applied literally: FBL-05/07 (linters), FBL-06 (observability), FBL-08/09
(HTTP layer), and DX-01/02/T-DOC-01/T-TEST-01 (generator/docs/test diagnosis) are all leaf work —
each closes a real capability gap identified in MATRIX CS-05/CS-08/CS-09/CS-10/CS-13/CS-14/CS-23
without requiring the ApplicationModel, the session-security redesign, or the online-migration
protocol to exist first.

## Framework capabilities delivered

- Machine-enforced resource-leak and security-linter discipline (zero-cost set) plus a judged set
  (gosec triage, errorlint, exhaustive annotations) — MATRIX CS-10/CS-23.
- Supply-chain hygiene: `go mod verify` in CI, a license-scanning signal while dependency-review is
  visibility-dormant, and a fixed pre-push hook DB-silent-skip gap.
- End-to-end observability: a log record inside a traced request carries `trace_id`/`span_id`; pgx
  query spans appear in the trace tree — MATRIX CS-05, decision D-08.
- HTTP transport hardening: configured (not infinite-default) server timeouts, prod-profile
  zero-timeout rejection, CSRF `MaxBytesReader` bound — MATRIX CS-09.
- Central validation enforcement at the `RouteMeta` seam: a mutating route cannot boot without a
  declared request contract — MATRIX CS-08.
- A generator that emits a valid, boot-passing CRUD module (fixes the `.delete`-verb defect that
  currently makes every `gen crud` invocation dead-on-arrival) — MATRIX CS-14.
- Corrected documentation traceability (DX-05 §6-vs-§9 fix, wowsociety upstream register
  reconciliation) and a diagnosed (not merely re-labelled) e2e flake — MATRIX CS-13.

## Included epics

- **W01-E01 — static-analysis-utilisation**: zero-cost linters (FBL-05), judged linter set (FBL-07
  part), supply-chain and hooks (FBL-07 remainder).
- **W01-E02 — observability-correlation**: trace/log correlation (FBL-06 T1/T2), pgx query tracer
  (FBL-06 T3, decision D-08).
- **W01-E03 — http-hardening**: server timeouts and body bounds (FBL-09), central validation
  enforcement (FBL-08).
- **W01-E04 — generator-doc-test-fixes**: generator correctness (DX-01/DX-02), documentation
  reconciliation (T-DOC-01, DX-05 residual, FBL-03), e2e flake diagnosis (T-TEST-01).

## Entry criteria

- W00 exit gate satisfied: the 8 executed finding-slices are re-verified at current HEAD, baselines
  captured, D-01..D-09 ratified as ADRs. In particular, D-08 (pgx query tracing approach) must be
  ratified before W01-E02-S002 can implement it as specified rather than as an open question.

## Exit criteria

- `.golangci.yml` enables the full zero-cost set (sqlclosecheck, rowserrcheck, bodyclose,
  wastedassign, makezero, musttag, testifylint) plus noctx/copyloopvar fixes, with zero unexplained
  new hits at wave close.
- The judged set (gosec with the CS-23 triage list, errorlint, exhaustive w/ fail-closed annotations,
  forcetypeassert, usestdlibvars) is enabled with every hit triaged (fixed or annotated with
  justification), plus `go mod verify` in CI and a license signal.
- Correlation attrs (`trace_id`/`span_id`) present on log records inside an active span, absent
  (not empty-string noise) without one; pgx spans appear in the trace tree.
- HTTP server timeouts are config-driven with safe defaults; prod profile rejects zero-value
  timeouts; CSRF middleware applies `MaxBytesReader`.
- Boot rejects a mutating route with no declared request contract (behind a profile flag for
  compatibility, per FBL-08's "compat: profile-flag first" note); an adversarial invalid-DTO POST
  returns 400 with field errors on a route the contract does declare.
- Generator-output-boots CI test passes: `gen crud` output compiles and boots without a closed-verb-
  set rejection.
- T-DOC-01's plan §6-vs-§9 inconsistency is fixed; DX-05's residual reconciliation items land; FBL-03
  marks the resolved wowsociety upstream findings closed.
- T-TEST-01's reproduction run either confirms or refutes the intermittent e2e failure and records a
  diagnosis — not a re-assertion of the withdrawn "shared-DB concurrency" cause.

## Dependencies

Depends on W00 (see above). Internally, W01-E03-S002 (central validation) coordinates with the
future AR-03 (W05) since `RouteMeta` is a projection input AR-03 will later derive — built compatibly
now per MATRIX CS-08's explicit note, not blocked on AR-03.

## Assumptions

- D-08's decision (thin in-kernel `pgx.QueryTracer`, not `otelpgx`) is ratified by W00-E02-S003
  before W01-E02-S002 begins; if not yet ratified, W01-E02-S002 documents this as a blocking gap
  rather than re-deciding it.
- The judged linter set's triage list (gosec G704/G115/G304, errorlint, exhaustive annotations) from
  MATRIX CS-23 is current at W01's start; if HEAD has drifted since the matrix was written, the
  triage step re-confirms rather than blindly applying the matrix's line citations.

## Risks

See `risks.md`. Headline risks: judged-linter enablement surfacing more hits than MATRIX CS-23's
snapshot recorded (drift since the matrix pass); FBL-08's boot-time enforcement breaking an
undeclared-contract route that currently works by accident (mitigated by the profile-flag-first
compat strategy).

## Quality gates

- Every linter enablement is its own fail-first artifact (mandate §13): the enablement run itself, at
  zero/near-zero hits per CS-23's inventory, is the evidence.
- FBL-08's boot rejection and FBL-09's timeout rejection are proven with adversarial fixtures, not
  merely "the happy path still works."
- Generator fix is proven fail-first: the generator-output-boots test must fail before the `.delete`→
  `.deactivate` fix and pass after.

## Required artifacts

- Updated `.golangci.yml` (linter config changes).
- `RouteMeta.Request` contract type + binding adaptor (FBL-08).
- Config schema additions for HTTP timeouts (FBL-09).
- `slog.Handler` wrapper + `pgx.QueryTracer` implementation (FBL-06).
- Fixed generator template (`resource.go.tmpl`) + generator-output-boots test harness (DX-02, shares
  DX-01 T5's scaffold primitive).
- Corrected plan doc §6/§9, corrected wowsociety upstream register entries (documentation artifacts).

## Required evidence

- Per-linter enablement run logs (zero/near-zero hit state).
- Correlation test output (trace_id attr present/absent matrix) + exported trace tree with pgx child
  spans.
- Boot-rejection test output (undeclared mutating route) + adversarial invalid-DTO 400 test.
- Template-render assertion + prod-profile zero-timeout rejection test.
- Generator-output-boots test log (fail before, pass after).
- T-TEST-01 reproduction-run artifact and resulting diagnosis note.

## Expected implementation outcome

A framework that mechanically enforces the resource-leak/security hygiene its own tooling already
ships but doesn't enable; a working, joined-up observability story (log↔trace↔DB-span correlation);
an HTTP layer that fails closed on both slow-connection exhaustion and unvalidated mutating routes;
and a generator whose output is provably boot-safe rather than merely hand-verified once.

## Acceptance authority

Framework architecture lead / developer-experience lead (role-based, split by epic — DX-04's
generator/doc work vs. the ARCH-adjacent linter/observability/HTTP work).

## Closure conditions

All exit criteria satisfied; all four epics' `closure-report.md` accepted; `waves/index.md`'s W01 row
updated to reflect `accepted` status; no unresolved regression from the judged-linter or FBL-08/09
enforcement work.
