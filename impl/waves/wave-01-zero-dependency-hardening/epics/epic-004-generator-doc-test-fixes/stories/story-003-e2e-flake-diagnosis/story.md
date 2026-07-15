---
id: W01-E04-S003
type: story
title: E2E flake diagnosis — reproduction-first investigation of the intermittent internal/e2e full-suite failure
status: accepted
wave: W01
epic: W01-E04
owner: W01Flake
reviewer: unassigned
priority: P2
created_at: 2026-07-12
updated_at: 2026-07-13
source_requirements:
  - T-TEST-01
depends_on: []
blocks: []
acceptance_criteria:
  - AC-W01-E04-S003-01
  - AC-W01-E04-S003-02
artifacts:
  - ART-W01-E04-S003-001
  - ART-W01-E04-S003-002
  - ART-W01-E04-S003-003
evidence:
  - EV-W01-E04-S003-001
  - EV-W01-E04-S003-002
  - EV-W01-E04-S003-003
decisions: []
risks:
  - RISK-W01-004
---

# W01-E04-S003 — E2E flake diagnosis

## Story ID

W01-E04-S003.

## Title

E2E flake diagnosis — reproduction-first investigation of the intermittent `internal/e2e` full-suite
failure.

## Objective

Reproduce, or genuinely fail to reproduce, the previously-observed intermittent `internal/e2e`
full-suite test failure under a disciplined `-count`+parallel protocol; determine the actual DB-wiring
mechanism `internal/e2e` uses; and record an honest diagnosis — a confirmed cause with evidence, or a
documented non-reproduction downgraded to a monitoring item — without re-asserting the withdrawn
"shared-DB concurrency" cause attribution that was made without checking the facts.

## Value to the framework

Mandate §2.2's dependency-aware sequencing principle explicitly lists "test infrastructure before
coverage enforcement" as a required ordering. That principle only has force if the test infrastructure
itself is understood, not merely assumed reliable or merely assumed broken. An unexamined "known flaky
test" claim erodes trust in the entire `internal/e2e` suite's signal — every future story that relies on
that suite passing as acceptance evidence inherits the uncertainty this story is meant to resolve. Doing
this investigation honestly (reproduction-first, no pre-committed mechanism) also demonstrates, for the
rest of the programme, what mandate §18's "record assumptions explicitly" and "do not silently resolve
ambiguous architecture decisions" actually look like in practice for a test-reliability question, not
just an architecture question.

## Problem statement

A prior review asserted that one observed `internal/e2e` full-suite failure (which passed 4/4 when the
same tests were run in isolation) was caused by "shared-DB concurrency" — multiple tests racing against
a single shared database. That claim was made **without first checking** whether `testkit` already
provides per-test database isolation. It does: `testkit/db.go:83-144,313` clones a per-test database via
`CREATE DATABASE ... TEMPLATE` from a content-hashed, pre-migrated template database, and drops the
clone in `t.Cleanup`. The isolation mechanism the original diagnosis assumed missing **exists**. This
means the cause attribution is withdrawn — not because the underlying observation was wrong (one
full-suite failure that passed 4/4 in isolation is still a real, confirmed fact), but because the
explanation offered for it was never actually verified against the code. This story's job is to do the
verification the original diagnosis skipped: reproduce properly, check the actual mechanism
`internal/e2e` uses, and only then conclude anything about cause.

## Source requirements

T-TEST-01 (re-scoped per MATRIX CS-13).

## Current-state assessment

**Confirmed facts:**

- `testkit/db.go:83-144,313` implements per-test database cloning via `CREATE DATABASE ... TEMPLATE`
  from a content-hashed migrated template, with cleanup via `t.Cleanup`. This mechanism exists in the
  codebase today — it is not proposed or planned, it is already implemented and presumably already used
  by *some* tests in the repository.
- One full-suite run of `internal/e2e` produced a failure; the same failing test(s), when run in
  isolation, passed 4/4. This is the original observed fact and is not in dispute.

**Not confirmed — this is the central question this story exists to answer:**

- Whether `internal/e2e` specifically (as opposed to other test packages in the repository) actually
  uses `testkit.NewDB`'s cloning mechanism, or has its own, separate DB-wiring path that may or may not
  provide equivalent isolation. This determination has not yet been made — it is task T001's second
  step, not an assumption this story's planning makes in either direction.
- Whether the original single observed failure is reproducible at all under a disciplined protocol, or
  was a one-off (infrastructure hiccup, resource exhaustion, an unrelated transient condition) that will
  not recur.

## Desired state

A diagnosis exists that is either: (a) a confirmed root cause, backed by reproduction evidence and a
concrete understanding of `internal/e2e`'s actual DB-wiring mechanism, with a fix applied to address it
(task T002); or (b) an honest, evidence-backed statement that the failure could not be reproduced under
a reasonable investigation budget, downgraded to a monitored item rather than left as an open, vaguely-
attributed "known flake." In neither case does the story re-assert "shared-DB concurrency" without new
evidence that specifically supports it.

## Scope

- T-TEST-01's exact 3-step re-scoped protocol: (1) reproduce under `-count`+parallel full-suite runs;
  (2) determine whether `internal/e2e` actually uses `testkit.NewDB` cloning or its own DB wiring; (3)
  fix what the reproduction shows.
- Task T001 covers steps 1–2 (reproduce, investigate) and produces a decision about what step 3's fix
  should be.
- Task T002 implements step 3, strictly conditional on what T001's investigation actually finds.

## Out of scope

- **Hosted fuzzing never running real `-fuzz=` coverage-guided generation** — this gap is folded into
  FBL-07 by MATRIX CS-13, but is already covered by `W01-E01-S003`'s scope. This story does not
  duplicate it.
- **The pre-push hook's DB-silent-skip gap** — also folded into FBL-07 by MATRIX CS-13, also already
  covered by `W01-E01-S003`'s scope. This story does not duplicate it.
- **Any change to `testkit`'s cloning mechanism itself**, unless T001's investigation specifically
  determines the mechanism itself is the root cause (in which case T002's conditional fix would target
  it — but this is not assumed or pre-committed at planning time).
- **Any change to `internal/e2e`'s test *content*** (what the tests assert) — this story is about test
  *infrastructure* reliability, not test-content correctness.

## Assumptions

- Whether `internal/e2e` uses `testkit.NewDB` or its own DB wiring is explicitly **not** assumed in
  either direction — this is the central unresolved question T001 exists to answer, not a premise this
  story's planning takes as given.
- The reproduction budget (exact `-count`/`-parallel` values, number of repeated executions) is not
  fixed at planning time; `plan.md` records this as an implementation-time decision, bounded by
  reasonable CI/local time budgets, not an unbounded investigation.

## Dependencies

None. This story is independent of sibling stories W01-E04-S001 and W01-E04-S002, and independent of
the rest of W01's epics (E01/E02/E03) — its scope (`internal/e2e/`, `testkit/`) is disjoint from theirs.

## Affected packages or components

`internal/e2e/` (the test suite under investigation), `testkit/` (the DB-isolation mechanism whose
actual usage by `internal/e2e` is being determined).

## Compatibility considerations

Not applicable at the investigation stage (T001). If T002's conditional fix changes how `internal/e2e`
sets up its database, compatibility with any other consumer of the same test-setup code would need
assessment at that time — not predictable now.

## Security considerations

Not applicable.

## Performance considerations

`-count`+parallel full-suite reproduction runs are inherently more expensive in wall-clock and resource
terms than a single normal run. T001's plan should budget this explicitly (e.g., as a bounded number of
repeated executions, not an open-ended loop) to keep the investigation itself tractable within CI or
local development time constraints.

## Observability considerations

The reproduction run's own logs are the primary evidence artifact for this story — there is no
production observability change involved.

## Migration considerations

Not applicable.

## Documentation requirements

The diagnosis note itself (recording what was reproduced, what was determined about `internal/e2e`'s
DB-wiring mechanism, and the resulting conclusion) is this story's primary documentation deliverable,
stored per the evidence path convention `evidence/premier/T-TEST-01/`.

## Acceptance criteria

- **AC-W01-E04-S003-01**: A reproduction protocol is executed against `internal/e2e`'s full suite under
  `-count`+parallel settings (values to be fixed at implementation time within a reasonable budget); the
  result — reproduced-with-evidence, or genuinely-not-reproduced — is recorded; a determination is made
  and recorded of whether `internal/e2e` uses `testkit.NewDB` cloning or its own DB wiring; the withdrawn
  "shared-DB concurrency" cause is not re-asserted without new supporting evidence.
- **AC-W01-E04-S003-02**: T002's fix (or explicit no-code-fix, monitoring-only outcome) is implemented
  strictly according to what T001's investigation determined, with the actual branch taken recorded
  against the illustrative decision branches identified in `plan.md`/task-002.

## Required artifacts

See `artifacts/index.md`. Expected: reproduction-run log collection, the diagnosis note (functioning as
a task-level decision record), and — conditionally, depending on T001's findings — a code or test change
produced by T002.

## Required evidence

See `evidence/index.md`. Expected evidence at path `evidence/premier/T-TEST-01/`: reproduction-run
artifacts and the resulting diagnosis note.

## Definition of ready

Per `governance/definition-of-ready.md`: this story is specific (one named intermittent failure, one
re-scoped 3-step protocol), bounded (explicit out-of-scope list above, cross-referencing W01-E01-S003
for the two folded-in gaps this story does not duplicate), implementable as an investigation (T001's
protocol is fully specified even though its *outcome* is not — this satisfies mandate §8.5's guidance
for cases where implementation details cannot yet be known), independently reviewable/verifiable
(T001's reproduction output is itself the verification artifact), traceable (`source_requirements:
[T-TEST-01]`), measurable acceptance criteria (two ACs above — AC-01 is measurable as
"protocol executed, determination recorded," not "flake found," which correctly does not presuppose an
outcome), dependencies identified (none).

## Definition of done

Per `governance/definition-of-done.md`, with an explicit modification recorded here per mandate §18
("state what must be determined during the story rather than inventing specifics"): this story's DoD
does **not** require "the flake is fixed" as a hard gate. An honest, evidence-backed non-reproduction,
downgraded to a monitoring item, is an accepted, valid, complete outcome for this story — consistent
with RISK-W01-004's framing that an inconclusive result is a legitimate story outcome, not a story
failure. What IS required for `done`: T001's protocol was actually executed (not skipped or assumed),
its result is recorded with evidence, the DB-wiring determination is recorded, and T002's outcome
(whatever it is) is implemented consistently with what T001 actually found.

## Risks

RISK-W01-004 (the reproduction step fails to reproduce the failure at all, leaving the diagnosis
inconclusive) — see `../../risks.md` (epic level) and `../../../risks.md` (wave level) for full
elaboration. This risk is explicitly accepted as a legitimate outcome, not mitigated away.

## Residual-risk expectations

If T001 cannot reproduce the failure within the planned budget, the residual risk (an unexplained,
unreproduced single historical failure) is explicitly accepted and downgraded to an ongoing monitoring
item — tracked at programme level, not left as an open blocker against this story's closure. This is
the story's own definition of an acceptable terminal state, not a fallback of last resort.
