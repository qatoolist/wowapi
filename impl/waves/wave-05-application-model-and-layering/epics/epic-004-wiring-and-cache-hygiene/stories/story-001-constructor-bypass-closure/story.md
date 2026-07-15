---
id: W05-E04-S001
type: story
title: Constructor-boundary lint and kernel.go audit
status: ready-for-review
wave: W05
epic: W05-E04
owner: task
reviewer: unassigned
priority: medium
created_at: 2026-07-12
updated_at: 2026-07-13
source_requirements:
  - AR-06
depends_on: []
blocks: []
acceptance_criteria:
  - AC-W05-E04-S001-01
  - AC-W05-E04-S001-02
artifacts:
  - ART-W05-E04-S001-001
  - ART-W05-E04-S001-002
evidence:
  - EV-W05-E04-S001-001
  - EV-W05-E04-S001-002
decisions: []
risks: []
---

# W05-E04-S001 — Constructor-boundary lint and kernel.go audit

## Story ID

W05-E04-S001

## Title

Constructor-boundary lint and kernel.go audit

## Objective

Implement an AST/lifecycle lint forbidding ad hoc infrastructure constructors outside composition
packages, and audit `kernel/kernel.go` to confirm or refute whether the fixed
closure-captures-a-fresh-instance pattern (already fixed at T1) is isolated to the one cited line.

## Value to the framework

T1's fix (already executed) closed one confirmed instance of a pattern that undermines the
framework's composition-root discipline — the `orgAncestry` closure calling `authz.NewStore()` a
second time instead of using the composed instance. This story ensures the pattern cannot silently
recur (T2's lint) and confirms it was not present elsewhere undetected (T3's audit).

## Problem statement

`requirement-inventory.md` row AR-06 states: "Remove hidden constructor bypasses | IMPL | P1 |
partial | W05-E04-S001 | T1 EXECUTED; T2 lint + T3 audit planned." PLAN's own AR-06 task table: T2 —
"AST/lifecycle lint forbidding ad-hoc infrastructure constructors outside composition packages | T1,
may share tooling with AR-02 T6 | Lint fails on a reintroduced ad-hoc constructor outside `kernel/`'s
composition root | Adversarial lint fixture | `AR-06/constructor_boundary_lint_test.go` | Medium —
new `go/analysis`-based tooling." T3 — "Audit `kernel/kernel.go` for any other instance of the same
'closure captures a fresh instance instead of the composed one' pattern | T1, T2 | Explicit audit
confirming/refuting the pattern is isolated to the one cited line | Audit report |
`AR-06/kernel_constructor_audit.md` | Low — mostly investigative; risk is under-scoping to just the
one cited line."

## Source requirements

AR-06 (T2, T3). T1 is already executed — see below.

## Current-state assessment

Per `requirement-inventory.md`'s own AR-06 row: "T1 EXECUTED." The `orgAncestry` closure fix
(`kernel/kernel.go:252-254`) is already in place per PLAN's own evidence. No constructor-boundary
lint and no formal audit report exist yet. This story does not re-plan or re-implement T1.

## Desired state

A lint rule fails CI on any reintroduced ad hoc infrastructure constructor outside composition
packages, proven by an adversarial fixture. An explicit audit report confirms or refutes whether the
closure-captures-a-fresh-instance pattern exists anywhere else in `kernel/kernel.go` beyond the
already-fixed line — with PLAN's own risk note explicit that under-scoping the audit to just the one
cited line is the main risk to guard against.

## Scope

- The constructor-boundary lint tool (T2), possibly sharing tooling with AR-02 T6's own
  `go/analysis`-based approach.
- The `kernel/kernel.go` audit for any other instance of the closure-captures-a-fresh-instance
  pattern (T3).

## Out of scope

- **AR-06 T1 (the `orgAncestry` closure fix)** — already executed. Not re-planned here.
- **Any other package's own constructor-bypass patterns beyond `kernel/kernel.go`** — T3's own scope
  is explicitly `kernel/kernel.go`, not a framework-wide audit; PLAN's own acceptance criterion names
  this one file.

## Assumptions

- T2's "may share tooling with AR-02 T6" note (PLAN's own dependency column) is recorded as a
  coordination opportunity, not a hard dependency — AR-02 T6 (W05-E02-S003) and this story are
  independent in their own scheduling; if AR-02 T6's tooling exists first, this task should reuse it;
  if not, this task builds its own, without blocking on AR-02 T6's own story completing first.

## Dependencies

None within W05-E04 (S001 is independent of S002). No dependency on any other W05 epic.

## Affected packages or components

A new lint tool (exact location TBD, likely `go/analysis`-based); `kernel/kernel.go` (read/audited,
not necessarily modified unless the audit finds a new instance of the pattern).

## Compatibility considerations

None material — a new lint rule and an audit report, not a runtime behavior change (unless the
audit finds a new instance requiring a fix, in which case that fix would follow T1's own pattern).

## Security considerations

The constructor-bypass pattern this lint guards against is itself a correctness/composition-
discipline concern (ensuring the composed, potentially-decorated instance is always used, not a
fresh one bypassing any configured cache/decorator) — adjacent to, but not itself, a
capability-security concern in the AR-01/AR-02 sense.

## Performance considerations

None material.

## Observability considerations

None beyond the lint's own CI failure reporting and the audit report's own documentation.

## Migration considerations

None.

## Documentation requirements

Document the lint rule's scope and the audit's findings (`AR-06/kernel_constructor_audit.md`).

## Acceptance criteria

- **AC-W05-E04-S001-01**: A lint rule fails CI on a reintroduced ad hoc infrastructure constructor
  outside composition packages — proven by `AR-06/constructor_boundary_lint_test.go`.
- **AC-W05-E04-S001-02**: An explicit audit report confirms or refutes whether the
  closure-captures-a-fresh-instance pattern exists anywhere else in `kernel/kernel.go` beyond the
  already-fixed line — `AR-06/kernel_constructor_audit.md`.

## Required artifacts

- The constructor-boundary lint tool (code).
- The `kernel/kernel.go` audit report.
See `artifacts/index.md`.

## Required evidence

- `AR-06/constructor_boundary_lint_test.go` output.
- `AR-06/kernel_constructor_audit.md`.
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`: `story.md`
and `plan.md` complete, acceptance criteria numbered and measurable, AR-06 T1's already-executed
status explicitly recorded, owner/reviewer assignment pending.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; both acceptance
criteria verified with evidence in `evidence/index.md`; `closure.md` completed.

## Risks

None material beyond ordinary implementation risk — PLAN's own risk column values (Medium, Low) are
moderate-to-low for this story's two tasks; T3's own risk note ("mostly investigative; risk is
under-scoping to just the one cited line") is addressed by the audit's own explicit
confirm-or-refute requirement.

## Residual-risk expectations

Residual risk is expected to be low once the audit report explicitly documents its search scope and
findings, addressing PLAN's own under-scoping concern directly.

## Plan

See `plan.md`.
