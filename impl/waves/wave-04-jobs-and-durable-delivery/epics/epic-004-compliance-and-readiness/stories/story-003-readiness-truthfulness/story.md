---
id: W04-E04-S003
type: story
title: Readiness and configuration diagnostics truthfulness
status: accepted
wave: W04
epic: W04-E04
owner: W04Compliance
reviewer: code-reviewer
priority: P1
created_at: 2026-07-12
updated_at: 2026-07-13
source_requirements:
  - DX-07
depends_on: []
blocks: []
acceptance_criteria:
  - AC-W04-E04-S003-01
  - AC-W04-E04-S003-02
  - AC-W04-E04-S003-03
artifacts: []
evidence: []
decisions: []
risks:
  - RISK-W04-004
---

# W04-E04-S003 — Readiness and configuration diagnostics truthfulness

## Story ID

W04-E04-S003

## Title

Readiness and configuration diagnostics truthfulness

## Objective

Add a migration-currency check to the generated readiness template so `/readyz` fails when the
applied-migration version lags expected (T1); add seed/rule/model-hash checks to readiness so it
reports migration version, seed/rule hash, and model hash (T2); and fix `config doctor`'s product-
root discovery to use `go env GOMOD`/`--project` instead of CWD-relative `os.Stat`, so delegation
works regardless of invocation directory and explicitly reports whether product validation ran (T3).
**DX-07 T4 (production-profile capacity/backpressure enforcement) is explicitly out of scope for this
story** — see "Out of scope" below.

## Value to the framework

The health contract's own documentation describes readiness as including "migration currency," but
the generated `cmd/api/main.go.tmpl` readiness map registers only `"db"` and `"seeds"` — a direct,
confirmed contradiction between documented contract and actual behavior. A framework whose readiness
endpoint can report healthy while running against a stale-migrated database, or while its own
product-checker discovery silently falls back to framework-only validation because it could not find
the product root from the current working directory, is a framework whose own operational-readiness
signal cannot be trusted. This story closes that specific, evidenced gap for the three tasks that are
buildable now, without waiting on the waiver-mechanism dependency that blocks T4.

## Problem statement

Per PLAN DX-07's own evidence: "the health contract's own doc describes readiness as including
'migration currency,' but the generated `cmd/api/main.go.tmpl` readiness map registers only `"db"`
and `"seeds"` — no migration-currency check exists, contradicting the documented contract.
`config_delegate.go`'s product-checker discovery is CWD-relative `os.Stat`, silently falling back to
framework-only validation if not found there." MATRIX CS-21 refines this evidence further: "the
framework readiness mechanism itself is correct and fail-closed — `kernel/httpx/health.go:52-79` runs
each check with a 3s timeout, 503 on any failure, reports `config_fingerprint`; `app/health.go:9-14`
documents DB/migration checks as a comment-only contract supplied via `extra`. The defect is thus
precisely located: contract-by-comment at the seam + template omission at the product end." The
mechanism is sound; the generated template simply never wires the migration-currency check the
contract promises.

## Source requirements

DX-07 (T1, T2, T3 only). **T4 is explicitly excluded** — see "Out of scope."

## Current-state assessment

Per PLAN DX-07's own evidence (to be re-confirmed at this story's actual start commit): the generated
`cmd/api/main.go.tmpl` readiness map registers exactly two checks, `"db"` and `"seeds"` — no
migration-currency check, no seed/rule/model-hash reporting. `config_delegate.go`'s product-checker
discovery uses CWD-relative `os.Stat`, silently falling back to framework-only validation if the
product root is not found relative to the current working directory. `CapacityMode` defaults to
`"advisory"` (never enforced); `HTTPMaxInFlight` defaults to `0` (backpressure fully disabled) — this
last pair of facts is the current-state context for T4, which this story does not implement (see "Out
of scope"). MATRIX CS-21's evidence refinement confirms the underlying readiness mechanism
(`kernel/httpx/health.go:52-79`) is itself correct and fail-closed (3s timeout per check, 503 on any
failure, reports `config_fingerprint`) — the defect is precisely the template's omission of the
migration-currency check the health contract's own comment (`app/health.go:9-14`) promises, not a
flaw in the underlying mechanism.

## Desired state

`/readyz` fails (503) when the applied-migration version lags the expected version, matching the
already-documented contract. Readiness reports migration version, seed/rule hash, and model hash (the
model-hash portion depending on AR-01's model hash existing, per PLAN T2's own dependency column).
`config doctor` discovers the product root via `go env GOMOD`/`--project`, working correctly
regardless of invocation directory (nested subdirectory, or outside-repo-with-`--project`), and
explicitly reports whether product validation ran, rather than silently falling back to
framework-only validation with no signal that it did so.

## Scope

- **T1** — migration-currency check added to the generated readiness template; `/readyz` fails when
  applied-migration version lags expected; proven by an integration test booting against a
  stale-migrated database and asserting a 503.
- **T2** — seed/rule/model-hash checks added to readiness reporting; readiness reports migration
  version, seed/rule hash, and model hash; the model-hash portion depends on AR-01's model hash, but
  T1 (and the rest of T2's non-model-hash reporting) can ship independently per PLAN's own dependency
  note ("AR-01's model hash for the model-hash portion; T1 can ship independently").
- **T3** — `config doctor` discovers the product root via `go env GOMOD`/`--project`, not
  CWD-relative `os.Stat`; delegation works regardless of invocation directory; explicitly reports
  whether product validation ran; proven by nested-subdirectory and outside-repo-with-`--project`
  test cases.

## Out of scope

- **DX-07 T4 — production-profile capacity/backpressure enforcement — is explicitly OUT OF SCOPE for
  this story.** T4's own dependency column states plainly: "T1-T3, AR-04's waiver framework." AR-04
  T5 (the shared waiver mechanism) is **W05-E03-S002**'s scope, which does not yet exist as of this
  story's planning — `impl/analysis/wave-allocation-detail.md`'s W05-E03-S002 row confirms it
  "builds the shared waiver mechanism consumed by SEC-06/DX-07." This is a forward reference by
  requirement ID (AR-04 T5) and by target story (W05-E03-S002) only — not a reference to a file path
  that may not yet exist, since a parallel effort may be building W05 concurrently and this story
  must not assume its files exist. No task is created for T4 anywhere in this story. This deferral is
  recorded consistently at wave level (RISK-W04-004, cited not re-derived) and at epic level
  (`epic.md` "Out of scope," `risks.md` RISK-W04-004) — see those documents for the full risk
  treatment.
- **DATA-08 W6-T1 through T5** — W04-E04-S001/S002's scope, unrelated to this story's DX-07 focus.
- **FBL-02** (production seed-sync path) — W02 scope, referenced only in MATRIX CS-21's shared
  closure-spec framing for context, not implemented here.
- **PROD-03** — wowsociety's own already-generated `cmd/api/main.go:240-243`'s manual backport of the
  migration-currency check pattern. PLAN DX-07's own wowsociety-impact note frames this as "a
  recommended follow-up (not blocking): manually backport the migration-currency check into
  wowsociety's own readiness map once wowapi's T1 pattern is established" — recorded as a
  non-blocking coordination note, not implemented here.

## Assumptions

- T4's exclusion is confirmed, not assumed, by the source: PLAN DX-07 T4's own dependency column
  ("T1-T3, AR-04's waiver framework") and `wave-allocation-detail.md`'s W05-E03-S002 row are both
  explicit that the waiver mechanism T4 requires is W05 scope, not yet built. This story treats that
  as a hard scope boundary, not a target to work around.
- T2's model-hash reporting portion is confirmed dependent on AR-01's model hash (PLAN T2's own
  dependency column: "AR-01's model hash for the model-hash portion; T1 can ship independently") — if
  AR-01's model hash is not yet available when this story is implemented, the model-hash portion of
  T2 may itself need to be sequenced or partially deferred; this is recorded as an implementation-time
  contingency in `plan.md`, not silently ignored.
- `config_delegate.go`'s exact current discovery logic (the CWD-relative `os.Stat` call) is treated as
  confirmed from PLAN's own evidence; this story's own re-confirmation step re-reads the file at
  actual start commit before implementing T3's fix.

## Dependencies

No dependency on W02-E01, W04-E04-S001, or W04-E04-S002 within this epic — per this epic's
`dependencies.md`: "DX-07 (S003) has no dependency on S001/S002 — it is an independent readiness/
diagnostics concern grouped into this epic by MATRIX CS-21's shared closure-spec framing... not by a
task dependency." T2's model-hash portion has an internal dependency on AR-01's model hash (W05
scope) — a partial, portion-level dependency, not a story-blocking one, per PLAN's own framing that
T1 can ship independently. No story within this epic depends on S003.

## Affected packages or components

The generated `cmd/api/main.go.tmpl` readiness template; `app/health.go`'s readiness-check
registration; `config_delegate.go`'s product-checker discovery logic; `kernel/httpx/health.go` only
if the underlying readiness-check execution mechanism itself requires extension (not expected — MATRIX
CS-21 confirms it is "correct and fail-closed" already).

## Compatibility considerations

T1's migration-currency check and T2's seed/rule/model-hash reporting change `/readyz`'s response
shape and its pass/fail behavior for any deployment currently running against a stale-migrated
database — a deployment that previously reported healthy (200) despite a migration lag will now
correctly report unhealthy (503) once this story lands, matching the already-documented but
previously unenforced contract. This is the intended, contract-restoring behavior change, not an
unintended regression, but it should be communicated to operators as a behavior change at rollout
time. T3's `config doctor` fix is confirmed **non-breaking for wowsociety**: PLAN's own wowsociety-
impact note states "Confirmed positive: `wowsociety/tools/configcheck/main.go` **exists**, so DX-07
T3's `config doctor` discovery fix is a non-issue for wowsociety — product-aware validation already
engages correctly today." T1's fix is also confirmed non-breaking for wowsociety's own already-
committed `cmd/api/main.go`: "T1's fix changes the template only; it does not retroactively alter
wowsociety's already-committed, non-regenerated `cmd/api/main.go`."

## Security considerations

None separately identified beyond the readiness-truthfulness property itself: a `/readyz` endpoint
that reports healthy while masking a real operational gap (stale migrations, missing product
validation) is itself a security-adjacent risk (an operator or automated deploy gate trusting a false
"healthy" signal), which is precisely what this story's three tasks correct.

## Performance considerations

None identified. The migration-currency check and hash-reporting additions are expected to execute
within the existing 3s-per-check timeout `kernel/httpx/health.go:52-79` already enforces; no separate
performance budget is introduced by this story.

## Observability considerations

Readiness reporting migration version, seed/rule hash, and model hash (T2) is itself an
observability improvement — these values become directly visible in the readiness payload rather than
requiring separate inspection. `config doctor`'s explicit reporting of whether product validation ran
(T3) is likewise an observability improvement over the current silent fallback.

## Documentation requirements

Document the migration-currency check's failure condition and expected-version source (T1);
document the readiness payload's new migration-version/seed-rule-hash/model-hash fields (T2);
document `config doctor`'s new discovery mechanism and its explicit product-validation-ran reporting
(T3); explicitly document that T4 is out of scope for this story and forward-reference AR-04 T5 /
W05-E03-S002.

## Acceptance criteria

- **AC-W04-E04-S003-01**: `/readyz` fails (503) when the applied-migration version lags the expected
  version, proven by an integration test that boots against a stale-migrated database and asserts the
  503 response, per PLAN DX-07 T1's own test column.
- **AC-W04-E04-S003-02**: Readiness reports migration version, seed/rule hash, and model hash, proven
  by an integration test inspecting the full readiness payload; the model-hash portion is proven
  contingent on AR-01's model hash being available (if unavailable at implementation time, this
  portion's status is recorded honestly in `deviations.md`, not silently claimed complete).
- **AC-W04-E04-S003-03**: `config doctor` discovers the product root via `go env GOMOD`/`--project`
  regardless of invocation directory, proven by nested-subdirectory and outside-repo-with-`--project`
  unit tests; delegation explicitly reports whether product validation ran, in both the success and
  fallback cases.

## Required artifacts

- The migration-currency readiness check (generated template change).
- The seed/rule/model-hash readiness reporting change.
- The `config doctor` `go env GOMOD`/`--project`-based discovery fix.
- Documentation of all three changes, including the explicit T4 out-of-scope note.
See `artifacts/index.md`.

## Required evidence

- Stale-migration 503 integration-test output (T1).
- Full-readiness-payload integration-test output (T2).
- Nested-subdirectory and outside-repo-`--project` `config doctor` discovery test output (T3).
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`: `story.md`
and `plan.md` complete, acceptance criteria numbered and measurable, no story-level dependency
recorded (T2's model-hash portion contingency on AR-01 recorded as a task-level note, not a story
blocker), owner/reviewer assignment pending, T4's exclusion and forward reference explicitly recorded
rather than silently dropped.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; all three acceptance
criteria verified with evidence in `evidence/index.md`; `closure.md` completed; independent review
passed per mandate §14, specifically confirming DX-07 T4 was correctly and explicitly scoped out —
no task silently attempting T4's capacity-enforcement behavior, and the forward reference to
W05-E03-S002 is not silently dropped.

## Risks

RISK-W04-004 (wave-level: DX-07 T4's forward dependency on W05-E03-S002's not-yet-built waiver
mechanism, leaving `CapacityMode`/`HTTPMaxInFlight`'s advisory/disabled defaults unresolved through
this epic's own closure) — see epic-level `risks.md` for full detail and mitigation/contingency. This
story does not attempt to mitigate RISK-W04-004 itself (T4 is out of scope); it only ensures the
deferral is honestly and traceably recorded.

## Residual-risk expectations

Once T1-T3 are implemented and evidenced, residual risk for this story's own scope is expected to be
low — a well-bounded, source-derived set of readiness/diagnostics fixes with no confirmed breaking-
change exposure. RISK-W04-004 remains open by design at this story's closure (and at the epic's) —
it is not this story's residual risk to resolve, only to record accurately.

## Plan

See `plan.md`.
