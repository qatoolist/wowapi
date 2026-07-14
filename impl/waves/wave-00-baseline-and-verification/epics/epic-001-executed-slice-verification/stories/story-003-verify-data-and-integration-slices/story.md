---
id: W00-E01-S003
type: story
title: Verify data-durability and CI-integration slices at current HEAD
status: accepted
wave: W00
epic: W00-E01
owner: unassigned
reviewer: unassigned
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-13
source_requirements: [DATA-08, REL-04, SD-01, SD-02]
depends_on: []
blocks: []
acceptance_criteria: [AC-W00-E01-S003-01, AC-W00-E01-S003-02, AC-W00-E01-S003-03]
artifacts: [ART-W00-E01-S003-001, ART-W00-E01-S003-002, ART-W00-E01-S003-003, ART-W00-E01-S003-004, ART-W00-E01-S003-005]
evidence: [EV-W00-E01-S003-01, EV-W00-E01-S003-02, EV-W00-E01-S003-03]
decisions: []
risks: []
---

# W00-E01-S003 — Verify data-durability and CI-integration slices at current HEAD

## Story ID

W00-E01-S003

## Title

Verify data-durability and CI-integration slices at current HEAD

## Objective

Re-verify, at the current repository HEAD, two finding-slices that `requirement-inventory.md`
records as `partial` — DATA-08 (W0-T1/W0-T2, attachment/notify durability) and REL-04 (T1-T4, S3/TOTP
wiring plus the parallel-CI pipeline state) — by re-running the exact named test commands and
inspecting the current CI configuration, then registering mandate-§10-conformant evidence for each.
This story also re-pins the three MATRIX verify-outcome rows (CS-03 config fail-closed, CS-19 i18n
freeze, CS-24 SSRF dial-time guard) that `epic.md` assigns to this story's scope, since re-confirming
"does the current repository state still match what the review/matrix documents claim" is the same
kind of check whether the claim comes from PLAN or from MATRIX. This story implements nothing new; it
converts prose "EXECUTED"/"verified" claims, cited against a prior commit SHA in the plan/review/
matrix documents, into re-proven, currently-pinned evidence inside `impl/`'s own traceability
structure.

## Value to the framework

DATA-08's W0 slice (attachment outbox-write fault-injection, legal-delivery audit durability) and
REL-04's T1-T4 slice (S3-gated test infrastructure, deterministic TOTP audit, parallel-CI pipeline)
are both load-bearing for later waves that build directly on top of them: W04-E04-S001..S002's
DATA-08 W6 hash-widening work assumes the W0 outbox-durability fix is genuinely intact today, and
W07-E02-S002's REL-04 T5-T8 fuzz work assumes the S3/TOTP/CI wiring this story re-verifies is still
the correct "before" state. As generic platform-kernel concerns, this story validates durable
outbox-write semantics (a property every downstream product depends on for compliance-evidence
integrity) and the truthfulness of the framework's own CI-integration coverage claims — neither is
specific to any downstream product's domain.

## Problem statement

`impl/analysis/requirement-inventory.md` §A records DATA-08 and REL-04 each with disposition
`partial`: their executed portions (DATA-08 W0-T1/W0-T2; REL-04 T1-T4) are described in the plan and
architecture-review documents as already implemented and tested, but the evidence backing that claim
exists only as prose test descriptions and `git log` commit-SHA citations in documents that predate
this program's own evidence register (mandate §10, `evidence-policy.md`). Separately,
`requirement-inventory.md` §C records CS-03, CS-19, and CS-24 as MATRIX verify-outcome rows already
re-earned by the closure-depth pass, with `wave.md`'s exit criteria requiring these three claims be
re-pinned with an evidence pointer at this wave's closing commit. `epic.md`'s "Scope" section
assigns that re-pinning to this story, on the grounds that REL-04's CI-pipeline-state confirmation
task is the natural home for a "does the current repository state still match what the documents
claim" check. None of DATA-08, REL-04, CS-03, CS-19, or CS-24 has an evidence record identifying the
execution command, the tested revision, the environment, and a reviewer in the format `impl/`
requires. Until that gap is closed, no downstream wave can safely treat these slices as a proven
"before" state.

## Source requirements

- **DATA-08** — Compliance evidence complete/durable. `requirement-inventory.md` §A: "partial |
  W04-E04-S001..S002 | W0 slice EXECUTED (verified x2); W6-T1 hash widening (D-04) + T2-T5 planned."
  This story covers only the W0-T1/W0-T2 executed portion — NOT the Wave-6 tasks.
- **REL-04** — Truthful integration coverage. `requirement-inventory.md` §A: "partial | W07-E02-S002
  | T1-T4 EXECUTED (verified x2); T5-T8 planned (T8 owns fuzz, shared w/ PERF-06 T3/T4)." This story
  covers only the T1-T4 executed portion.
- **SD-01** — `requirement-inventory.md` §E: "CI gate parallelized (3 legs), toolbox image
  GHA-cached, docs-only skip (#23) — Quality-gate baseline changed; W00 baseline captures new
  wall-clocks." This story confirms SD-01 is reflected in the current `.github/workflows/ci.yml` as
  part of the REL-04 CI-pipeline-state check.
- **SD-02** — `requirement-inventory.md` §E: "Bench path-scoped on PRs; nightly schedule;
  merge_group support (#24) — Partially advances FBL-07 (nightly exists); REL-04 T8 fuzz remains."
  This story confirms SD-02 is reflected in the current `.github/workflows/ci.yml`; the REL-04 T8
  fuzz remainder stays out of scope (tracked at `W07-E02-S002`).
- **CS-03/CS-19/CS-24** (`requirement-inventory.md` §C, disposition `INV→verified`) — re-pinned as
  part of this story per `epic.md` "Scope" (see Task 3). These are not new source requirements this
  story implements; they are matrix verify-outcome rows this story confirms still hold.

## Current-state assessment

Confirmed facts, established by direct repository inspection at the time this story was drafted
(commit `0a31186`, 2026-07-12):

- `kernel/attachment/attachment.go` and `kernel/attachment/coverage_test.go` exist.
- `kernel/notify/service.go` and `kernel/notify/notify_test.go` exist.
- A migration numbered `00011` exists in the migrations directory and is described in the source
  material as granting `events_outbox` INSERT permission used by the legal-delivery audit write.
- `.github/workflows/ci.yml`, `deployments/compose.yaml`, and `Makefile` exist and are the files
  named as carrying the REL-04 T1-T4 wiring (`WOWAPI_REQUIRE_S3`, `S3_TEST_ENDPOINT`, minio
  `service_healthy` condition, canonical `S3_ENDPOINT` variable) and the SD-01/SD-02 CI-pipeline
  state (3-leg parallel gate, toolbox image GHA-caching, docs-only skip, path-scoped/nightly/
  merge_group bench).

What is **not yet confirmed** (must be confirmed during story execution, not assumed): that
`kernel/attachment/attachment.go`'s outbox-write error path still propagates rather than discards the
error; that `kernel/attachment/coverage_test.go`'s fault-injection test still proves rollback on that
error path; that `kernel/notify/service.go`'s legal-delivery audit write still uses migration 00011's
grant and that `kernel/notify/notify_test.go` still covers it; that `Makefile`'s `ci-container` target
and the hosted `gate` job still wire `WOWAPI_REQUIRE_S3=1` and `S3_TEST_ENDPOINT`; that
`deployments/compose.yaml`'s minio service still declares a `service_healthy` condition; that the
canonical `S3_ENDPOINT` variable naming is still consistent across the wiring; that the TOTP audit is
still deterministic (not wall-clock-dependent) across at least two distinct mocked clock/timezone
settings; and that `.github/workflows/ci.yml` still reflects the SD-01 (3-leg parallel gate,
GHA-cached toolbox image, docs-only skip) and SD-02 (path-scoped/nightly/merge_group bench) state.
This story's tasks exist specifically to confirm or refute these claims against the current HEAD, not
to assert them as already proven. Whether CS-03/CS-19/CS-24's underlying implementation is still
intact at current HEAD is likewise not yet confirmed — Task 3 exists to re-pin (or flag regression
against) each.

## Desired state

Three mandate-§10-conformant evidence records exist — one per acceptance criterion — covering: (1)
DATA-08 W0's attachment/notify durability re-verification; (2) REL-04 T1-T4's S3/TOTP wiring and the
SD-01/SD-02 CI-pipeline-state confirmation; (3) CS-03/CS-19/CS-24's re-pinned verify-outcome evidence
pointers. Each evidence record cites the exact execution command (or inspection method for the
CI-state and CS-repin checks), the commit SHA it was run against, the environment, tool versions,
date/time, result, and reviewer. If any slice fails to re-verify, a `failed`-status evidence record is
preserved (not silently retried until green) and a follow-up remediation task is opened under that
finding's canonical target story (`W04-E04-S001..S002` for DATA-08, `W07-E02-S002` for REL-04) per
`requirement-inventory.md`; a CS-03/19/24 regression is flagged as a new finding, not silently
absorbed.

## Scope

- Re-running `go test ./kernel/attachment/... ./kernel/notify/...` (DB-gated, requires testkit
  Postgres) and confirming DATA-08 W0-T1 (outbox-write error no longer discarded, fault-injection
  test proves rollback) and W0-T2 (legal-delivery audit write via migration 00011's grant) are
  present and passing.
- Re-running the 20 S3-gated tests via `make ci-container` (or `docker compose` + `go test` with
  `WOWAPI_REQUIRE_S3=1`) against MinIO, and confirming REL-04 T1 (`Makefile`/hosted `gate` job
  wiring), T2 (minio `service_healthy` condition in `deployments/compose.yaml`), and T3 (canonical
  `S3_ENDPOINT` variable cleanup).
- Re-running the TOTP audit suite at two different mocked clock/timezone settings and confirming
  REL-04 T4 (deterministic, non-wall-clock-dependent audit path).
- Inspecting `.github/workflows/ci.yml` to confirm SD-01 (3-leg parallelized gate, GHA-cached
  toolbox image, docs-only skip) and SD-02 (bench path-scoped on PRs, nightly schedule, merge_group
  support) are reflected in the current pipeline configuration.
- Re-pinning CS-03 (config fail-closed + fingerprint), CS-19 (i18n freeze + key-echo fallback), and
  CS-24 (SSRF dial-time guard) verify-outcome claims with an evidence pointer confirming they still
  hold at this story's closing commit.
- Registering one evidence record per acceptance criterion in this story's `evidence/index.md`.

## Out of scope

- Any code change, fix, or remediation to `kernel/attachment/`, `kernel/notify/`,
  `.github/workflows/ci.yml`, `deployments/compose.yaml`, or `Makefile` — this story is
  verification-only (epic `epic.md` "Out of scope," inherited here). If a regression is found, a new
  task is opened under the finding's canonical target story listed above; it is not silently fixed
  inside this story.
- DATA-08 W6-T1 (hash widening, D-04) and W6-T2-T5 — tracked at `W04-E04-S001..S002`.
- REL-04 T5-T8 (T8 owns the fuzz work shared with PERF-06 T3/T4) — tracked at `W07-E02-S002`.
- FBL-07's nightly-fuzz remainder (SD-02 confirms the nightly *schedule* exists; it does not confirm
  the fuzz corpus itself has grown beyond seed-replay — that remains FBL-07's own disposition per
  `requirement-inventory.md` §B).
- Any other finding-slice covered by this epic's sibling stories (SEC-02, AR-04, AR-06 — `S001`;
  PERF-01, PERF-06 — `S002`) or by AR-05 (excluded from this epic entirely per `epic.md` "Out of
  scope").
- New implementation work for CS-03/CS-19/CS-24 — these are re-pinned as already-verified matrix
  outcomes, not re-implemented.

## Assumptions

- The two finding-slices (DATA-08 W0, REL-04 T1-T4) and the three matrix verify-outcomes (CS-03,
  CS-19, CS-24) described as EXECUTED/verified in the plan/review/matrix documents are still present
  and unmodified at this story's working HEAD; if any has regressed, that is a finding this story's
  verification work surfaces, not an assumption the story proceeds under (inherited from `wave.md`
  "Assumptions").
- `docker compose` / Postgres / MinIO test infrastructure referenced by the plan's evidence commands
  (`make ci-container`, the S3-gated test suite) is available in the execution environment — this is
  the one story in this epic that requires **both** Postgres and MinIO simultaneously (S001 needs at
  most Postgres; S002 needs neither).
- A working Go toolchain per `go.mod` is available in the execution environment.
- `.github/workflows/ci.yml` at the story's working HEAD is the authoritative record of CI pipeline
  state; no separate GitHub Actions run history needs to be queried to confirm SD-01/SD-02 — file
  inspection is sufficient, since the task is "does the config reflect the claimed state," not "did a
  specific run succeed."

## Dependencies

None. Per `epic.md` "Internal (cross-story) dependencies," S003 targets packages and test commands
disjoint from its sibling stories S001 and S002, and this story's own three tasks (T001/T002/T003)
target disjoint concerns (attachment/notify durability; S3/TOTP/CI-pipeline wiring; matrix
verify-outcome re-pinning) and can execute in any order — see `plan.md` "Implementation sequence" for
the one soft-ordering note (T003's CS-03 config-fail-closed check benefits from T002's CI-inspection
work already having the workflow file open, but does not require it).

## Affected packages or components

Verification-only; no production code is expected to change. The packages and files this story's
tasks read and re-test:

- `kernel/attachment/attachment.go`, `kernel/attachment/coverage_test.go` (DATA-08 W0-T1).
- `kernel/notify/service.go`, `kernel/notify/notify_test.go` (DATA-08 W0-T2), plus the migration
  numbered `00011` (events_outbox INSERT grant).
- `Makefile` (`ci-container` target, S3 env wiring), `deployments/compose.yaml` (minio service
  health-check condition), `.github/workflows/ci.yml` (hosted `gate` job S3 wiring, 3-leg
  parallelization, toolbox image caching, docs-only skip, bench path-scoping/nightly/merge_group
  support) (REL-04 T1-T3, SD-01, SD-02).
- The TOTP audit test suite (exact file path to be confirmed during Task 2 execution — not yet
  pinned in the source material) (REL-04 T4).
- Whatever source files implement CS-03 (config fail-closed + fingerprint), CS-19 (i18n freeze +
  key-echo fallback), and CS-24 (SSRF dial-time guard) — exact paths to be confirmed during Task 3
  execution by following the evidence pointers MATRIX cites for each.

## Compatibility considerations

Not applicable — this story makes no code change and therefore introduces no compatibility impact.
If a regression is found and a remediation task is opened under the finding's canonical target story,
compatibility considerations belong to that future story's own `story.md`.

## Security considerations

This story is security-load-bearing on two axes. First, DATA-08's outbox-write durability and
legal-delivery audit write are compliance-evidence integrity properties — a regression here means
compliance evidence could be silently lost, which this story's Task 1 directly re-proves does not
happen via fault injection. Second, REL-04 T4's TOTP-audit determinism is a security-verification
property: an audit path that is wall-clock-dependent could pass or fail non-deterministically in ways
that mask a real authentication-audit gap, which is why Task 2 re-runs the TOTP suite at two distinct
mocked clock/timezone settings rather than once. Third, Task 3 re-pins three matrix verify-outcomes
that are themselves security findings (CS-03 config fail-closed, CS-19 i18n freeze/key-echo fallback,
CS-24 SSRF dial-time guard) — a regression on any of these would be a security regression, not merely
a functional one, and is treated accordingly by `risks.md` inheritance from the epic/wave register.

## Performance considerations

None expected beyond the normal runtime cost of running the DB-gated and S3-gated test suites. No
performance budget is defined or measured by this story (performance baselines are `W00-E01-S002`'s
scope).

## Observability considerations

None. This story does not add or change logging, metrics, or tracing; it re-runs existing tests
against existing instrumentation and inspects existing CI configuration.

## Migration considerations

None new. DATA-08 W0-T2's legal-delivery audit write depends on migration `00011` already having been
applied in the test environment; this story does not create, modify, or re-run any migration — it
confirms the already-granted permission is exercised correctly by the existing code path.

## Documentation requirements

None beyond this story's own governance documents (`story.md`, `plan.md`, `implementation.md`,
`verification.md`, `deviations.md`, `closure.md`, and the task/artifact/evidence indexes). No
external documentation (e.g. `docs/`) is expected to change as a result of this story.

## Acceptance criteria

- **AC-W00-E01-S003-01**: `go test ./kernel/attachment/... ./kernel/notify/...` exits 0 at the
  story's closing commit SHA, against testkit Postgres; the test output confirms
  `kernel/attachment/coverage_test.go`'s fault-injection test proves rollback on the outbox-write
  error path (DATA-08 W0-T1), and confirms `kernel/notify/notify_test.go` proves the legal-delivery
  audit write succeeds via migration 00011's `events_outbox` INSERT grant (DATA-08 W0-T2); evidence
  registered as an evidence ID in `evidence/index.md` (planned format `EV-W00-E01-S003-01`).
- **AC-W00-E01-S003-02**: The 20 S3-gated tests exit 0 when run with `WOWAPI_REQUIRE_S3=1` against
  MinIO (via `make ci-container` or `docker compose` + `go test`), confirming REL-04 T1
  (`Makefile`/hosted `gate` job wiring) and T2 (minio `service_healthy` condition in
  `deployments/compose.yaml`) and T3 (canonical `S3_ENDPOINT` variable naming); the TOTP audit suite
  exits 0 at two distinct mocked clock/timezone settings, confirming REL-04 T4's determinism;
  inspection of `.github/workflows/ci.yml` confirms SD-01 (3-leg parallelized gate, GHA-cached
  toolbox image, docs-only skip) and SD-02 (bench path-scoped on PRs, nightly schedule, merge_group
  support) are reflected in the current pipeline configuration; evidence registered as an evidence ID
  in `evidence/index.md` (planned format `EV-W00-E01-S003-02`).
- **AC-W00-E01-S003-03**: CS-03 (config fail-closed + fingerprint), CS-19 (i18n freeze + key-echo
  fallback), and CS-24 (SSRF dial-time guard) verify-outcome claims are each re-confirmed at the
  story's closing commit SHA, with an evidence pointer citing the specific test(s) or inspection
  method that re-proves each; any regression is flagged as a new finding rather than silently
  absorbed; evidence registered as an evidence ID in `evidence/index.md` (planned format
  `EV-W00-E01-S003-03`).

## Required artifacts

Test execution logs for the DB-gated and S3-gated/TOTP suites, plus a CI-pipeline-state inspection
note and a CS-03/CS-19/CS-24 re-pin note (one artifact type per task) — see `artifacts/index.md`. No
other artifact type is expected: this story produces no schema, interface, migration, or
design-document artifact, since no code changes.

## Required evidence

Three evidence records, one per acceptance criterion, per `evidence/index.md`. Evidence types:
DB-gated test-execution log (AC-01); S3-gated test-execution log + TOTP determinism test log (2
clock/TZ settings) + CI-configuration inspection note (AC-02); verify-outcome re-pin note with
evidence pointers (AC-03).

## Definition of ready

Per `governance/definition-of-ready.md` Story DoR: this story is specific (two named finding-slices
plus one bounded matrix-outcome re-pin, not an aspirational theme); bounded (scope/out-of-scope both
stated above); implementable (exact commands and files are named — see `plan.md`); independently
reviewable and verifiable (does not depend on another story's completion); traceable to source
requirements (`source_requirements` front matter lists DATA-08, REL-04, SD-01, SD-02); has measurable
acceptance criteria (AC-...-01/02/03 above, each a pass/fail test-exit-code or inspection-confirmed
result); dependencies identified (`none`, stated above); assumptions recorded (see "Assumptions");
`plan.md` exists with task breakdown and unresolved questions; required artifacts and evidence
anticipated (above); compatibility/security/performance/observability/migration considerations
addressed (above, several explicitly marked not applicable with reason). This story satisfies the DoR
checklist and may move to `ready` once an owner is assigned.

## Definition of done

Per `governance/definition-of-done.md`: this story reaches `accepted` only when, in addition to each
of the three tasks reaching `done` (own implementation/verification records recording actual result,
evidence ID, execution date/revision, reviewer), every acceptance criterion in this story has a
corresponding `pass` entry with a valid evidence ID in `verification.md`; required artifacts and
evidence are registered per `artifact-policy.md`/`evidence-policy.md`; `deviations.md` states "no
deviations" or lists every deviation with reason, impact, approval, and compensating controls;
`closure.md` is complete; and the independent-review checklist in `definition-of-done.md` has been
run and passed clean — including its explicit check that no source requirement (DATA-08, REL-04,
SD-01, SD-02) or matrix verify-outcome (CS-03, CS-19, CS-24) has been silently dropped. Mandate §7,
applied here verbatim: "A story must not be accepted solely because all tasks are marked complete."

## Risks

- RISK-W00-001 (wave/epic-level, inherited) — a claimed-executed slice fails to re-verify at current
  HEAD; high severity, would block W04-E04-S001..S002 (DATA-08 W6 remainder) and W07-E02-S002 (REL-04
  T5-T8 remainder).
- RISK-W00-002 (wave/epic-level, inherited) — test infrastructure (Postgres and, uniquely for this
  story, MinIO) unavailable or misconfigured, producing a false-negative regression; this story
  carries the epic's highest exposure to this risk since it is the only story requiring both
  services simultaneously.
- A story-specific risk not yet assigned an ID: the CS-03/CS-19/CS-24 re-pin (Task 3) could surface
  that one of these matrix verify-outcomes no longer holds at current HEAD, which is not merely a
  re-verification failure but a regression in an already-`verified` security finding — this would
  need immediate escalation, not routine follow-up-task handling, given the security nature of all
  three (fail-closed config, i18n freeze, SSRF guard).

## Residual-risk expectations

Even after this story is accepted, some residual risk remains that a re-verified slice could regress
again later (e.g. a future change to the CI workflow silently dropping the docs-only skip, or a
future dependency bump reintroducing wall-clock dependence in the TOTP audit path) — this story
proves current-HEAD correctness at a point in time, not a permanent guarantee. Ongoing protection
against CI-configuration drift is not this story's concern; it is addressed, where applicable, by the
quality gates the CI pipeline itself enforces going forward. This residual risk is expected and
accepted as normal for a point-in-time re-verification story; it does not block acceptance.
