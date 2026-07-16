---
id: W03-E01-S004
type: story
title: Cross-repo cutover plan for the wowsociety impersonation-flow breaking change
status: implemented
wave: W03
epic: W03-E01
owner: unassigned
reviewer: unassigned
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-13
source_requirements:
  - SEC-01
depends_on:
  - W03-E01-S001
  - W03-E01-S002
blocks: []
acceptance_criteria:
  - AC-W03-E01-S004-01
  - AC-W03-E01-S004-02
  - AC-W03-E01-S004-03
artifacts: []
evidence: []
decisions: []
risks:
  - RISK-W03-002
---

# W03-E01-S004 — Cross-repo cutover plan for the wowsociety impersonation-flow breaking change

## Story ID

W03-E01-S004

## Title

Cross-repo cutover plan for the wowsociety impersonation-flow breaking change

## Objective

Produce three coordination-artifact documents — a sequencing plan, a staging-validation plan, and
a rollback plan — governing the two-repo coordinated cutover of wowsociety's impersonation flow
onto the framework's new `identity_grant` table and privileged-session resolver (S001/S002). This
is `requirement-inventory.md` row PROD-04's target story. **This is a documentation/verification
story. It produces no product code in either wowapi or wowsociety.**

## Value to the framework

PLAN §5.2's SEC-01 wowsociety-impact prose states the cutover cannot be a single-repo event:
"Sequencing: two-repo coordinated cutover — wowapi ships T1+T5, wowsociety's auth flow adopts
`grant_id`, only then cut over; validate T2 against wowsociety staging data before making it
unconditional." Without an explicit, reviewed plan for this sequencing, S001/S002's server-side
grant authority either ships unused (wowsociety keeps trusting its own unverified
`identity_impersonation_session` state) or wowsociety attempts an uncoordinated cutover that breaks
its live impersonation flow. This story's value is making the cutover coordinatable at all — it is
the bridge between a framework capability that exists and a product capability that safely adopts
it.

## Problem statement

SEC-01's T1 (`identity_grant` table) and T5 (privileged-session resolver) are, per PLAN's own
framing, "BREAKING for impersonation" in wowsociety. `internal/modules/identity/
impersonation.go:1-21` states explicitly: "What the framework does NOT provide: a session/grant
record... This file is that product-side layer" — wowsociety has already built its own workaround
(`identity_impersonation_session` table, `startImpersonation`/`stopImpersonation`, audited via
`kaudit.Entry`). `whoami.go:39,51` reads `actor.ImpersonatorUserID` directly off the framework
`authz.Actor`, "populated from the unverified claim, by explicit design (comment: trusts the claim
'without a DB re-check')." Test files `abac_test.go:52-94` and `whoami_impersonation_test.go:31-56`
construct `authz.Actor{ImpersonatorUserID: ...}` literals directly — PLAN's own words: "load-bearing
test surface that will need rewriting." REVIEW §P's impact-and-rework-matrix table confirms: "Add
grant-state columns to `identity_impersonation_session`; mint/reference framework `grant_id`;
rework whoami trust... Two-repo coordinated cutover; validate against staging data before framework
enforces." No plan for *how* this two-repo coordination actually happens — what ships first, how
staging validation is structured, what the rollback path is if the cutover breaks something —
exists today. This story creates that plan.

## Source requirements

SEC-01 (T1/T5's wowsociety-impact prose). Cross-referenced: `requirement-inventory.md` §D row
PROD-04 ("SEC-01 impersonation cutover (whoami/impersonation/tests) | Product auth flow rework |
SEC-01 T1/T5 grant contract + coordinated rollout plan"). REVIEW §P (wowapi→wowsociety impact
matrix, SEC-01 row).

## Current-state assessment

Confirmed facts (from PLAN §5.2 and REVIEW §P, cited above): wowsociety currently trusts
`actor.ImpersonatorUserID` directly off the unverified JWT claim, by explicit self-documented
design; wowsociety has its own `identity_impersonation_session` table and audit trail that is not
today reconciled with any framework-owned grant record (because no framework-owned grant record
exists yet — that is what S001/S002 build). No sequencing, staging-validation, or rollback plan for
the eventual cutover exists anywhere in the repository today; this story's three documents are
wholly new, not extensions of an existing plan.

## Desired state

Three reviewed documents exist: a sequencing plan stating which repo ships what and in what order
(framework T1+T5 first, then wowsociety's grant_id adoption, then the cutover itself — per PLAN's
own stated sequence); a staging-validation plan describing how T2's unconditional membership
enforcement is validated against wowsociety's staging data before being made unconditional in
wowsociety's production path; and a rollback plan describing how a failed or problematic cutover is
reverted without leaving either repo in an inconsistent state. None of the three documents contains
product code — they are coordination artifacts consumed by whoever executes the actual wowsociety-
side rework, which is out of this programme's scope.

## Scope

- A sequencing plan document: repo-by-repo ordering of the cutover (wowapi ships T1+T5 → wowsociety
  adopts `grant_id` in `identity_impersonation_session`/`startImpersonation`/`stopImpersonation` →
  coordinated cutover flips `whoami.go` to trust the framework's verified `Actor` fields).
- A staging-validation plan document: how T2's unconditional membership enforcement (from S001) and
  T5's resolver (from S002) are validated against wowsociety's staging environment and data before
  either is made unconditional/enforced in wowsociety's production path, per PLAN's explicit
  instruction: "validate T2 against wowsociety staging data before making it unconditional."
  Includes identification of the load-bearing test fixtures that need rewriting (`abac_test.go:
  52-94`, `whoami_impersonation_test.go:31-56`) as a staging-validation checklist item, per PLAN's
  own citation of them as "good regression coverage to re-run post-cutover."
- A rollback plan document: how to revert the cutover on either side if a problem is found —
  covering both "wowapi's grant-table enforcement causes a wowsociety regression" and "wowsociety's
  adoption of `grant_id` is itself found to be broken" failure modes.

## Out of scope

- Any wowsociety code change (`identity_impersonation_session` schema, `startImpersonation`/
  `stopImpersonation`, `whoami.go`, `abac_test.go`, `whoami_impersonation_test.go`) — this is
  product-level work (mandate §2.3), explicitly excluded from framework implementation and recorded
  as PROD-04.
- Any wowapi code change beyond what S001/S002 already deliver — this story does not implement or
  modify T1/T5, it plans their coordinated rollout.
- Deciding DEC-Q1 (the IdP `grant_id` claim contract) — the sequencing plan assumes S001/S002's
  safe-default behavior; if DEC-Q1 resolves differently, the sequencing plan may need revision, but
  that revision is out of this story's initial scope.
- Executing the actual cutover — this story produces the plan, not the cutover itself.

## Assumptions

- wowsociety's staging environment and its `identity_impersonation_session` data will be available
  to validate against at the time the cutover is actually executed. This story's staging-validation
  plan states the validation approach; it does not assume a specific availability timeline — see
  `plan.md`'s "Unresolved questions."
- The exact wowsociety-side engineering ownership and timeline for adopting `grant_id` is not known
  at this story's planning time and is explicitly not invented here, per mandate §18's instruction
  to state what must be determined rather than inventing specifics.
- S001 and S002's `identity_grant` schema and resolver contract are stable enough by this story's
  own execution time to plan a sequencing document against — if S001/S002 are still in flux, this
  story's plan documents should be drafted as living documents and revised as needed rather than
  treated as final on first pass.

## Dependencies

Depends on W03-E01-S001 (grant-table shape) and W03-E01-S002 (resolver contract) being at least
substantially planned, ideally implemented, so the sequencing/staging/rollback plans have a stable
target to sequence against. No story within W03 depends on this story's own output for its
technical implementation — S004 is a pure coordination artifact consumed outside this programme's
own execution graph, by whoever performs the eventual wowsociety-side rework.

## Affected packages or components

None — this story produces documentation only. No wowapi package or file is modified; no
wowsociety file is touched (this story does not have write access to, nor scope over, the
wowsociety repository).

## Compatibility considerations

This story's entire subject matter is a breaking-change compatibility plan — see "Problem
statement" above. The story itself introduces no new compatibility concern; it documents how to
manage the one SEC-01 T1/T5 already introduces.

## Security considerations

The staging-validation plan's own content is security-sensitive: it must not recommend validating
against production impersonation data without appropriate access controls, and must explicitly flag
that grant-authority migration (moving impersonation trust from client-claim to server-grant) is a
security-critical cutover that should not be executed silently or without a coordinated go/no-go
review — consistent with DATA-09's own "human sign-off strongly advisable" pattern for
safety-critical cutovers (PLAN §5.3 DATA-09 T8).

## Performance considerations

Not applicable — this story produces no runtime code.

## Observability considerations

The staging-validation plan should recommend what wowsociety-side observability (e.g. a dashboard
or log correlation confirming grant-table lookups are succeeding at the expected rate during the
cutover window) would give the coordinated rollout visibility — a recommendation within the plan
document, not an implementation this story performs itself.

## Migration considerations

Not applicable to this story directly (S001 owns the `identity_grant` migration itself); this
story's sequencing plan does reference DATA-09's online-migration discipline as the pattern
wowsociety's own `identity_impersonation_session` schema change (adding `grant_id` columns) should
follow, per REVIEW §P's own recommendation, but does not itself author that migration.

## Documentation requirements

This story's entire output is documentation: the sequencing plan, staging-validation plan, and
rollback plan. No additional documentation beyond the mandate §8 required story/task files and
these three plan documents is required.

## Acceptance criteria

- **AC-W03-E01-S004-01** — A sequencing plan document exists, stating the repo-by-repo order
  (wowapi T1+T5 → wowsociety `grant_id` adoption → coordinated cutover) and naming the specific
  wowsociety files/tests known to require rework (`whoami.go:39,51`, `impersonation.go`,
  `abac_test.go:52-94`, `whoami_impersonation_test.go:31-56`), reviewed by at least a wowapi-side
  reviewer.
- **AC-W03-E01-S004-02** — A staging-validation plan document exists, describing how S001's T2
  unconditional-membership enforcement and S002's T5 resolver are validated against wowsociety
  staging data before either is made unconditional/enforced in wowsociety's production path, and
  identifying the specific existing wowsociety test suites (`abac_test.go`,
  `whoami_impersonation_test.go`, `rls_test.go`) to re-run post-cutover per PLAN's own citation.
- **AC-W03-E01-S004-03** — A rollback plan document exists, covering both failure directions
  (wowapi-side enforcement causing a wowsociety regression; wowsociety-side `grant_id` adoption
  itself being broken), reviewed by at least a wowapi-side reviewer.

## Required artifacts

- Sequencing plan document.
- Staging-validation plan document.
- Rollback plan document.
See `artifacts/index.md`.

## Required evidence

- Review records for each of the three plan documents (review reports, per mandate §10's "review
  reports" evidence type — there is no executable test for a coordination-plan document).
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`:
`story.md` and `plan.md` complete, acceptance criteria numbered and measurable, dependencies
(S001/S002 substantially planned) recorded, owner/reviewer assignment pending.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`: all
three plan documents exist, match `plan.md`'s intended scope or have deviations recorded in
`deviations.md`; all three acceptance criteria verified with review-record evidence in
`evidence/index.md`; `closure.md` completed; independent review confirms no product code was
introduced by this story (a review check specific to this story's documentation-only nature).

## Risks

RISK-W03-002 (the wowsociety two-repo coordinated cutover cannot be completed unilaterally by this
wave) — see epic-level `risks.md` for full detail. This story's own scope is exactly the mitigation
for RISK-W03-002 (producing the coordination plan), not its full resolution (the plan's execution
remains outside this programme).

## Residual-risk expectations

Even after this story's three plan documents are accepted, residual risk remains that the actual
wowsociety-side execution diverges from the plan, is delayed, or surfaces a problem the plan did not
anticipate — this is accepted as a structural residual risk (see RISK-W03-002's "Residual risk"
framing at epic scope), not something this documentation-only story can eliminate.

## Plan

See `plan.md`.
