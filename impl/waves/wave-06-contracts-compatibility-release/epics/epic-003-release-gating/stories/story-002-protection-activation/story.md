---
id: W06-E03-S002
type: story
title: Protection activation — branch/tag/environment protection (human-gated, DEC-Q10)
status: blocked
wave: W06
epic: W06-E03
owner: repo-administrator
reviewer: release/security-engineering-lead
priority: P0
created_at: 2026-07-12
updated_at: 2026-07-13
source_requirements:
  - REL-01
depends_on:
  - W06-E03-S001
blocks: []
acceptance_criteria:
  - AC-W06-E03-S002-01
  - AC-W06-E03-S002-02
artifacts: []
evidence:
  - EV-W06-E03-S002-001
decisions: []
risks:
  - RISK-W06-001
---

# W06-E03-S002 — Protection activation — branch/tag/environment protection (human-gated, DEC-Q10)

## Story ID

W06-E03-S002

## Title

Protection activation — branch/tag/environment protection (human-gated, DEC-Q10)

## Objective

Activate the final, admin-only layer of REL-01's release-gating trust boundary: branch protection on
`main`, a protected `release` GitHub Environment with required reviewers, and a tag protection ruleset.

**THIS STORY IS HUMAN-GATED.** It cannot enter `ready` or `in-progress` status until DEC-Q10 (a
repo-administrator action) is resolved by a human with repo-admin access — no coding agent can create a
protected GitHub Environment, set branch protection, or configure a tag protection ruleset. This is not
a scope-reduction or a quality shortcut; it is a structural fact about GitHub's own permission model,
confirmed via live `gh api` calls in the source review.

## Value to the framework

Without this story's own activation, W06-E03-S001's release pipeline runs its `publish` job unprotected
in scratch — proven correct in mechanics, but not enforced by GitHub's own platform-level controls. This
story is the final step that converts "a pipeline whose logic is correct" into "a pipeline whose logic
is correct AND cannot be bypassed by a compromised or careless direct push/tag/environment action,"
because GitHub's own branch/tag/environment protection is the actual enforcement backstop no amount of
workflow YAML can substitute for.

## Problem statement

PLAN's own REL-01 evidence, verified via live `gh api` calls: "`gh api repos/qatoolist/wowapi/branches/
main/protection` → 404, no branch protection exists. `gh api repos/qatoolist/wowapi/environments` →
`{"total_count":0}`, no GitHub Environments exist. `security_and_analysis` fields all `disabled` — no
GHAS license active." PLAN's own "Human-required blockers" section: "**Protected `release` GitHub
Environment does not exist** (`total_count: 0`) — creating one with required reviewers is a
repo-admin-console-only action. **T7 cannot be end-to-end proven until this exists.** **No branch
protection on `main`, no tag protection rules exist** — the directive's own closing sentence ('Protect
release tags and the environment at the repository/organization level') requires repo-admin console
action, independent of how well `required-gates.yml` is built." REVIEW §F row 10 confirms: "GitHub
org-admin actions (branch/tag/env protection) for REL-01/REL-02 | **Genuine repo-administration** | Only
the *final activation* needs admin. All workflow YAML, gate manifest, and verification script are
authorable + testable now against a scratch repo. See §G. | **No** for implementation; **Yes** for
rollout enforcement only." `requirement-inventory.md` §B confirms: "DEC-Q10 | Repo-admin activation
(branch/tag/env protection) | OPS | P0 | blocked (human) | W06-E03 (tracked) | Merge-queue rulesets
unavailable on user-owned repo (session fact)."

## Source requirements

REL-01 (T9's own protection-activation remainder); DEC-Q10.

## Current-state assessment

Confirmed via live `gh api` calls at the source review's own execution time (2026-07-11): no branch
protection exists on `main`; no GitHub Environments exist at all (`total_count: 0`); no tag protection
ruleset exists; `security_and_analysis` fields are all `disabled`. This story's own re-confirmation step
is to re-run these exact `gh api` calls at this story's own actual start commit/date, since repository
settings can change independently of code — this re-confirmation is itself the first action a human
performing this story's work must take, per this programme's fail-first re-confirmation convention.

## Desired state

Branch protection is configured on `main` (exact rule set — required status checks referencing this
wave's own manifest-driven gates, required reviews, etc. — TBD by the repo administrator performing this
activation, informed by W06-E03-S001's own gate set). A protected `release` GitHub Environment exists
with required reviewers, such that W06-E03-S001's `publish` job (which already targets this environment
in its own YAML) now actually runs against real platform-level protection instead of unprotected-in-
scratch. A tag protection ruleset exists, protecting release tags from unauthorized creation/deletion,
consistent with the directive's own closing instruction: "Protect release tags and the environment at
the repository/organization level." **REVIEW's own session-fact note applies:** "Merge-queue rulesets
unavailable on user-owned repo" — this story's own scope must work within that platform constraint, not
assume a merge-queue-ruleset feature this specific repository does not have access to.

## Scope

- Repo-admin action: configure branch protection on `main` (required status checks, required reviews,
  as determined appropriate by the repo administrator at activation time).
- Repo-admin action: create the protected `release` GitHub Environment with required reviewers.
- Repo-admin action: configure a tag protection ruleset for release tags.
- Post-activation validation: confirm W06-E03-S001's `publish` job now runs against the real protected
  environment (not the stub used during S001's own development/testing), and that T7's own unmanifested-
  artifact rejection test still passes against the real environment.

## Out of scope

- **Any of W06-E03-S001's own pipeline-mechanics work** — that story's T1-T8 are fully buildable and
  testable without this story's activation; this story does not re-implement or re-test that mechanics
  work, it only activates the platform-level enforcement layer around it.
- **Choosing the exact branch-protection rule set's specific parameters** (number of required reviews,
  which status checks are required) beyond what W06-E03-S001's own gate manifest already names as
  required checks — the repo administrator's own judgment at activation time governs any parameter this
  planning document does not pre-specify, consistent with mandate §18's instruction not to invent
  specifics the source does not give.

## Assumptions

- This story's status may legitimately remain `planned` (blocked-entry) indefinitely if no repo
  administrator acts — this is expected and must be recorded honestly in this wave's and this epic's own
  `closure-report.md` as a deferred, tracked, non-silent open item, per REVIEW's own recommendation to
  track this as a distinct ticket ("PF-REL-ADMIN-01: configure release environment + tag/branch
  protection... tracked separately from and blocking T7/T9's *full* closure, so agent-completable
  YAML/script work isn't silently gated on an unstaffed admin task").
- The exact branch-protection rule parameters (required review count, specific required status checks)
  are not specified by any source document beyond "required status checks referencing this wave's own
  manifest-driven gates" — the repo administrator determines the exact parameters at activation time,
  informed by but not dictated by this planning document.

## Dependencies

Depends on W06-E03-S001 (the release pipeline this activation protects) reaching a state where its own
gate manifest exists to reference as required status checks — S001 need not be fully `accepted` before
this story's *entry criterion* (DEC-Q10 resolution) can be satisfied independently by a human, but the
activation's own required-status-check configuration is more meaningful once S001's manifest is stable.
Depends on DEC-Q10 (human, repo-admin) as this story's own explicit blocked-entry criterion — this is
the primary, structural dependency this story exists to track.

## Affected packages or components

None in the code sense — this story's own "implementation" is a set of GitHub repository/organization
settings changes performed via the GitHub web console or `gh api`/`gh` CLI by a human with repo-admin
access. No Go package, no CI workflow file, is modified by this story itself (W06-E03-S001's own
`publish` job YAML already targets the `release` environment by name — creating that environment is
this story's action, not a YAML change).

## Compatibility considerations

Once branch protection is active, any existing direct-push-to-`main` workflow (if any exists today)
would need to go through a PR instead — this is the intended, correct behavior change this story exists
to produce, not an unintended regression.

## Security considerations

This entire story IS the security control — branch protection, protected environment, and tag
protection are the platform-level enforcement backstop for everything W06-E03-S001 built in workflow
logic. There is no security consideration separate from this story's own objective.

## Performance considerations

Not applicable.

## Observability considerations

Once active, GitHub's own audit log records every branch-protection-bypass attempt, environment-
approval action, and tag-protection-rule violation — this is platform-provided observability, not
something this story must additionally build.

## Migration considerations

Not applicable.

## Documentation requirements

Document the exact branch-protection rule set, environment-reviewer list, and tag-protection-ruleset
configuration once activated, so a future reader (including a future audit) can see exactly what
platform-level controls are in place without needing repo-admin access to inspect them directly.

## Acceptance criteria

- **AC-W06-E03-S002-01**: Branch protection is active on `main`, a protected `release` GitHub Environment
  exists with required reviewers, and a tag protection ruleset is active — each confirmed via a live
  `gh api` call re-run (not merely a console screenshot), mirroring the exact verification method the
  source review itself used to confirm the *absence* of these controls.
- **AC-W06-E03-S002-02**: W06-E03-S001's `publish` job, previously tested against a stub environment, is
  re-verified against the real protected `release` environment, and its unmanifested-artifact rejection
  test still passes.

## Required artifacts

- A record of the activated branch-protection configuration.
- A record of the activated `release` environment's required-reviewer configuration.
- A record of the activated tag-protection ruleset.
See `artifacts/index.md`. All entries are "not yet produced" until a human repo administrator performs
the activation.

## Required evidence

- Live `gh api` call output confirming branch protection is active (mirroring the exact call used to
  confirm its prior absence).
- Live `gh api` call output confirming the `release` environment exists with required reviewers.
- Live `gh api` call output confirming the tag protection ruleset is active.
- Re-verification output of W06-E03-S001's `publish` job against the real environment.
See `evidence/index.md`.

## Definition of ready

**This story has a non-standard, explicitly stated readiness gate: it CANNOT move to `ready` or
`in-progress` until DEC-Q10 is resolved by a human with repo-admin access.** Per
`governance/definition-of-ready.md`'s "dependencies identified" requirement, this story's DoR is
satisfied for *planning* purposes (its front matter and this `story.md` correctly state the blocking
dependency) while remaining structurally unable to proceed to `ready` until that human action occurs.
This is a deliberate, source-confirmed application of the mandate's own allowance for explicit `blocked`
status with documented entry criteria (per `requirement-inventory.md`'s own framing of DEC-Q1/DEC-Q9/
DEC-Q10 as "blocked (human)").

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`: both
acceptance criteria verified with live-`gh-api`-call evidence in `evidence/index.md` (not a console
screenshot or an unverified claim); `closure.md` completed; independent review passed per mandate §14,
specifically confirming the activation was genuinely performed by a human with repo-admin access (via
live API verification) and not fabricated or assumed.

## Risks

RISK-W06-001 (this story cannot enter `ready`/`in-progress` until DEC-Q10 is resolved by a human with
repo-admin access) — see epic-level `risks.md` for full detail and mitigation/contingency.

## Residual-risk expectations

This story's residual risk cannot be reduced by any action within this programme's own execution
capacity — it is a genuine, irreducible human-administration dependency. The mitigation is honest
tracking (this story's own explicit blocked-entry framing, and REVIEW's own recommended separate
"PF-REL-ADMIN-01" ticket), not risk elimination.

## Plan

See `plan.md`.
