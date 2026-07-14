---
id: W00-E01-S001
type: story
title: Verify workflow and boot composition slices at current HEAD
status: accepted
wave: W00
epic: W00-E01
owner: worker W00E01S001
reviewer: unassigned
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-13
source_requirements: [SEC-02, AR-04, AR-05, AR-06]
depends_on: []
blocks: []
acceptance_criteria: [AC-W00-E01-S001-01, AC-W00-E01-S001-02, AC-W00-E01-S001-03, AC-W00-E01-S001-04]
artifacts: [ART-W00-E01-S001-001, ART-W00-E01-S001-002, ART-W00-E01-S001-003, ART-W00-E01-S001-004]
evidence: [EV-W00-E01-S001-01, EV-W00-E01-S001-02, EV-W00-E01-S001-03, EV-W00-E01-S001-04]
decisions: []
risks: []
---

# W00-E01-S001 — Verify workflow and boot composition slices at current HEAD

## Story ID

W00-E01-S001

## Title

Verify workflow and boot composition slices at current HEAD

## Objective

Re-verify, at the current repository HEAD, four finding-slices that `requirement-inventory.md`
records as `partial` (executed-portion-only) — SEC-02 (T1-T3, workflow privileged-ops fail-closed),
AR-04 (T1, boot-time silent-behaviour elimination), AR-05 (T1/T2, composition/doc drift removal in
README and blueprint), and AR-06 (T1, hidden-constructor-bypass removal in `kernel/kernel.go`'s
`orgAncestry` composition) — by re-running the exact named test/inspection commands and registering
mandate-§10-conformant evidence for each. This story implements nothing new; it converts four prose
"EXECUTED" claims, cited against a prior commit SHA in the plan/review documents, into re-proven,
currently-pinned evidence inside `impl/`'s own traceability structure.

## Value to the framework

These three slices are load-bearing for later waves that build directly on top of them: W03's
SEC-02 T4/T5 ratification design assumes the T1-T3 fail-closed fix is genuinely intact today, not
merely intact at the SHA the review documents cite; W05-E03-S002's AR-04 T2-T5 work and
W05-E04-S001's AR-06 T2/T3 work both assume their respective T1 slices are still correct before
building the remainder on top. As a generic platform-kernel concern, this story validates
composition-root correctness (fail-closed authorization, deterministic boot-time config
validation, single-owner store composition) — properties every downstream product depends on the
kernel providing honestly, independent of any product-specific domain.

## Problem statement

`impl/analysis/requirement-inventory.md` §A records SEC-02, AR-04, and AR-06 each with disposition
`partial`: their executed portions (SEC-02 T1-T3; AR-04 T1; AR-06 T1) are described in the plan and
architecture-review documents as already implemented and tested, but the evidence backing that
claim exists only as prose test descriptions and `git log` commit-SHA citations in documents that
predate this program's own evidence register (mandate §10, `evidence-policy.md`). None of the three
slices has an evidence record identifying the execution command, the tested revision, the
environment, and a reviewer in the format `impl/` requires. Until that gap is closed, no downstream
wave can safely treat these slices as a proven "before" state.

## Source requirements

- **SEC-02** — Workflow privileged ops fail closed. `requirement-inventory.md` §A: "partial |
  W03-E05-S001 | T1-T3 EXECUTED (verified x2); T4 ratification design + T5 audit remain." This
  story covers only the T1-T3 executed portion.
- **AR-04** — Eliminate boot-time silent behaviour. `requirement-inventory.md` §A: "partial |
  W05-E03-S002 | T1 EXECUTED (verified x2); T2-T5 planned, dep AR-01; T5 waiver shared w/
  SEC-06/DX-07." This story covers only the T1 executed portion.
- **AR-05** — Composition/doc drift removal. `requirement-inventory.md` §A: "partial |
  W06-E04-S002 | T1/T2 EXECUTED; T3 doc-example CI gate (CS-22 spec), T4/T5 dep AR-03." This story
  covers only the T1/T2 executed portion — the W06-E04-S002 target applies to the *remaining*
  planned work (T3-T5), exactly as AR-04's and AR-06's W05 targets apply to their remainders while
  their executed slices are re-pinned here. AR-05 is one of the 8 executed finding-slices W00's
  charter names (PLAN §6: "8 findings ... have a real, independently-reviewed partial closure").
- **AR-06** — Remove hidden constructor bypasses. `requirement-inventory.md` §A: "partial |
  W05-E04-S001 | T1 EXECUTED; T2 lint + T3 audit planned." This story covers only the T1 executed
  portion.

## Current-state assessment

Confirmed facts, established by direct repository inspection at the time this story was drafted
(commit `0a31186`, 2026-07-12):

- `kernel/workflow/runtime.go` exists, alongside `runtime_extra_test.go`, `runtime_lifecycle_test.go`,
  and `runtime_test.go` in the same package, plus `testkit/workflowsim_cov_test.go`.
- `app/boot.go` exists, alongside `app/boot_extra_test.go`.
- `kernel/kernel.go` exists; `kernel/authz/caching_internal_test.go` and `kernel/kernel_rules_test.go`
  exist.

What is **not yet confirmed** (must be confirmed during story execution, not assumed): that
`NewRuntime` still panics on `ev == nil`; that `Override` still unconditionally checks authz with no
`if rt.authz != nil` skip; that `kernel/kernel.go`'s `orgAncestry` closure (previously cited at lines
252-254) still uses the composed `authzStore` instance rather than calling `authz.NewStore()` a
second time; and that `app/boot.go` still rejects unknown `modules.<typo>` namespaces with a
deterministic named error. This story's tasks exist specifically to confirm or refute these claims
against the current HEAD, not to assert them as already proven.

## Desired state

Three mandate-§10-conformant evidence records exist — one per finding-slice — each citing the exact
execution command, the commit SHA the command was run against, the environment, tool versions,
date/time, result, and reviewer. If any slice fails to re-verify, a `failed`-status evidence record
is preserved (not silently retried until green) and a follow-up remediation task is opened under
that finding's canonical target story (`W03-E05-S001` for SEC-02, `W05-E03-S002` for AR-04,
`W05-E04-S001` for AR-06) per `requirement-inventory.md`.

## Scope

- Re-running `go test ./kernel/workflow/... -race` and confirming the SEC-02 fail-closed behavior
  (nil-`ev` panic in `NewRuntime`; unconditional authz check in `Override`) is present and covered
  by the five named test files.
- Re-running `go test ./app/... -run Boot` (or the specific boot-namespace-rejection test) plus a
  full `go test ./...` green check, and confirming AR-04 T1's unknown-namespace rejection behavior
  in `app/boot.go`.
- Re-running `go test ./kernel/authz/... -race` and `go test ./kernel/... -run TestKernelRules -race`
  (or the equivalent covering `kernel_rules_test.go`), and confirming AR-06 T1's `orgAncestry`
  closure uses the composed `authzStore` instance via a sentinel-store-injection test.
- Registering one evidence record per acceptance criterion in this story's `evidence/index.md`.

## Out of scope

- Any code change, fix, or remediation to `kernel/workflow/runtime.go`, `app/boot.go`, or
  `kernel/kernel.go` — this story is verification-only (epic `epic.md` "Out of scope," inherited
  here). If a regression is found, a new task is opened under the finding's canonical target story
  listed above; it is not silently fixed inside this story.
- SEC-02 T4 (ratification design) and T5 (audit) — tracked at `W03-E05-S001`.
- AR-04 T2-T5 — tracked at `W05-E03-S002`, dependent on AR-01.
- AR-06 T2 (lint) and T3 (audit) — tracked at `W05-E04-S001`.
- Any other finding-slice covered by this epic's sibling stories (PERF-01, PERF-06, DATA-08,
  REL-04 — `W00-E01-S002`/`W00-E01-S003`). AR-05's executed T1/T2 (doc-drift fixes) ARE re-verified
  here (AC-04); its remaining T3-T5 are tracked at `W06-E04-S001`/`W06-E04-S002`.

## Assumptions

- The three slices described as EXECUTED in the plan/review documents are still present and
  unmodified at this story's working HEAD; if any has regressed, that is a finding this story's
  verification work surfaces, not an assumption the story proceeds under (inherited from `wave.md`
  "Assumptions").
- A working Go toolchain per `go.mod` is available in the execution environment. Whether
  `testkit/workflowsim_cov_test.go` requires a live Postgres instance (via `make ci-container`) is
  not yet confirmed and must be determined during task execution (see Task 001).
- `kernel/authz/caching_internal_test.go` and `kernel/kernel_rules_test.go`'s DB dependency (if any)
  is likewise not yet confirmed and must be determined during Task 003 execution.

## Dependencies

None. Per `epic.md` "Internal (cross-story) dependencies," S001 targets packages and test commands
disjoint from its sibling stories S002 and S003, and this story's own three tasks (T001/T002/T003)
target disjoint packages (`kernel/workflow`; `app`; `kernel`+`kernel/authz`) and can execute in any
order or in parallel — see `plan.md` "Implementation sequence."

## Affected packages or components

Verification-only; no production code is expected to change. The packages and files this story's
tasks read and re-test:

- `kernel/workflow/runtime.go`, `kernel/workflow/runtime_extra_test.go`,
  `kernel/workflow/runtime_lifecycle_test.go`, `kernel/workflow/runtime_test.go`,
  `testkit/workflowsim_cov_test.go` (SEC-02).
- `app/boot.go`, `app/boot_extra_test.go` (AR-04 T1).
- `kernel/kernel.go` (lines previously cited at 252-254, the `orgAncestry` closure),
  `kernel/authz/caching_internal_test.go`, `kernel/kernel_rules_test.go` (AR-06 T1).

## Compatibility considerations

Not applicable — this story makes no code change and therefore introduces no compatibility impact.
If a regression is found and a remediation task is opened under the finding's canonical target
story, compatibility considerations belong to that future story's own `story.md`.

## Security considerations

SEC-02 is itself a security-fail-closed finding: this story's Task 001 directly re-proves that
privileged workflow operations fail closed (deny-by-default) rather than silently succeeding when
authorization state is absent or misconfigured. A regression here would be a security regression,
not merely a functional one — per epic `risks.md` RISK-W00-001, a failed re-verification is treated
as high severity and blocks downstream SEC-02 ratification work (W03-E05-S001) until resolved.

## Performance considerations

None expected. `go test -race` re-runs of the named packages carry the normal race-detector runtime
overhead; no performance budget is defined or measured by this story (performance baselines are
`W00-E01-S002`'s scope).

## Observability considerations

None. This story does not add or change logging, metrics, or tracing; it re-runs existing tests
against existing instrumentation.

## Migration considerations

None. No data, schema, or configuration migration is involved in a verification-only story.

## Documentation requirements

None beyond this story's own governance documents (`story.md`, `plan.md`, `implementation.md`,
`verification.md`, `deviations.md`, `closure.md`, and the task/artifact/evidence indexes). No
external documentation (e.g. `docs/`) is expected to change as a result of this story.

## Acceptance criteria

- **AC-W00-E01-S001-01**: `go test ./kernel/workflow/... -race` exits 0 at the story's closing
  commit SHA; the test output confirms the presence and passing state of the `NewRuntime` nil-`ev`
  panic assertion and the `Override` unconditional-authz-check (fail-closed) assertion among the
  suite; evidence registered as an evidence ID in `evidence/index.md` (planned format
  `EV-W00-E01-S001-01`).
- **AC-W00-E01-S001-02**: `go test ./app/... -run Boot` (or the specific boot-namespace-rejection
  test) exits 0, AND a full `go test ./...` exits 0, at the story's closing commit SHA; the test
  output confirms `app/boot.go` rejects an unknown `modules.<typo>` config namespace with a
  deterministic named error; evidence registered as an evidence ID in `evidence/index.md` (planned
  format `EV-W00-E01-S001-02`).
- **AC-W00-E01-S001-03**: `go test ./kernel/authz/... -race` and the test(s) covering
  `kernel/kernel_rules_test.go` (`go test ./kernel/... -run TestKernelRules -race` or equivalent)
  exit 0 at the story's closing commit SHA; the test output confirms, via a sentinel-store-injection
  test, that `kernel/kernel.go`'s `orgAncestry` closure uses the composed `authzStore` instance
  rather than a second `authz.NewStore()` call; evidence registered as an evidence ID in
  `evidence/index.md` (planned format `EV-W00-E01-S001-03`).
- **AC-W00-E01-S001-04**: AR-05 T1/T2 doc-drift fixes still hold at the story's closing commit SHA:
  `grep -rn "RunAPI\|RunWorker\|RunMigrate" README.md docs/blueprint/` returns zero phantom-API
  hits, and `docs/blueprint/06-module-sdk.md`'s `Context` method listing matches the live
  `module/module.go` interface method-for-method (diff of documented vs `go doc`/source-extracted
  method sets is empty); evidence registered as `EV-W00-E01-S001-04` (grep + diff output log).

## Required artifacts

Test execution logs (one per acceptance criterion / task) — see `artifacts/index.md`. No other
artifact type is expected: this story produces no schema, interface, migration, or design-document
artifact, since no code changes.

## Required evidence

Four evidence records, one per acceptance criterion (three test-execution logs + one doc-drift grep/diff log), per `evidence/index.md`.
Evidence types: race-detector test-execution log (AC-01), test-execution log + full-suite green
check (AC-02), race-detector test-execution log (AC-03).

## Definition of ready

Per `governance/definition-of-ready.md` Story DoR: this story is specific (three named
finding-slices, not an aspirational theme); bounded (scope/out-of-scope both stated above);
implementable (exact commands and files are named — see `plan.md`); independently reviewable and
verifiable (does not depend on another story's completion); traceable to source requirements
(`source_requirements` front matter lists SEC-02, AR-04, AR-06); has measurable acceptance criteria
(AC-...-01/02/03 above, each a pass/fail test-exit-code result); dependencies identified (`none`,
stated above); assumptions recorded (see "Assumptions"); `plan.md` exists with task breakdown and
unresolved questions; required artifacts and evidence anticipated (above); compatibility/
security/performance/observability/migration considerations addressed (above, several explicitly
marked not applicable with reason). This story satisfies the DoR checklist and may move to `ready`
once an owner is assigned.

## Definition of done

Per `governance/definition-of-done.md`: this story reaches `accepted` only when, in addition to
each of the three tasks reaching `done` (own `implementation.md`/`verification.md` sections
recording actual result, evidence ID, execution date/revision, reviewer), every acceptance
criterion in this story has a corresponding `pass` entry with a valid evidence ID in
`verification.md`; required artifacts and evidence are registered per `artifact-policy.md`/
`evidence-policy.md`; `deviations.md` states "no deviations" or lists every deviation with reason,
impact, approval, and compensating controls; `closure.md` is complete; and the independent-review
checklist in `definition-of-done.md` has been run and passed clean — including its explicit check
that no source requirement (SEC-02, AR-04, AR-06) has been silently dropped. Mandate §7, applied
here verbatim: "A story must not be accepted solely because all tasks are marked complete."

## Risks

- RISK-W00-001 (wave/epic-level, inherited) — a claimed-executed slice fails to re-verify at current
  HEAD; high severity, would block W03-E05-S001 (SEC-02 ratification), W05-E03-S002 (AR-04
  remainder), and W05-E04-S001 (AR-06 remainder).
- RISK-W00-002 (wave/epic-level, inherited) — test infrastructure (Postgres via testkit) unavailable
  or misconfigured, producing a false-negative regression, if `testkit/workflowsim_cov_test.go` or
  the `kernel/authz`/`kernel` DB-backed tests require a live database; applicability to this story's
  specific tests is one of the "must be confirmed during execution" items above.

## Residual-risk expectations

Even after this story is accepted, some residual risk remains that a re-verified slice could regress
again later (e.g. an unrelated future change to `kernel/kernel.go` silently reintroducing a second
`authz.NewStore()` call) — this story proves current-HEAD correctness at a point in time, not a
permanent guarantee. Ongoing protection against regression is the concern of AR-06 T2 (lint rule,
tracked at `W05-E04-S001`), not this story. This residual risk is expected and accepted as normal
for a point-in-time re-verification story; it does not block acceptance.
