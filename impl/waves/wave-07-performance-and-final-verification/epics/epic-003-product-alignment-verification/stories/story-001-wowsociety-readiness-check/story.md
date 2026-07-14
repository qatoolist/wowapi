---
id: W07-E03-S001
type: story
title: wowsociety readiness check — framework-side PROD-01..05 coordination-artifact verification
status: blocked
wave: W07
epic: W07-E03
owner: W07-Phase-A-Execution.W07E03S001
reviewer: W05ReviewGateFinal
priority: P2
created_at: 2026-07-12
updated_at: 2026-07-14
source_requirements: []
depends_on:
  - W02-E01
  - W02-E02
  - W05-E05
  - W04-E04
  - W01-E03
  - W03-E01
  - W00-E02
blocks: []
acceptance_criteria:
  - AC-W07-E03-S001-01
  - AC-W07-E03-S001-02
  - AC-W07-E03-S001-03
  - AC-W07-E03-S001-04
  - AC-W07-E03-S001-05
artifacts:
  - ART-W07-E03-S001-001
evidence:
  - EV-W07-E03-S001-001
  - EV-W07-E03-S001-002
  - EV-W07-E03-S001-003
  - EV-W07-E03-S001-004
  - EV-W07-E03-S001-005
decisions: []
risks:
  - RISK-W07-E03-001
---

# W07-E03-S001 — wowsociety readiness check — framework-side PROD-01..05 coordination-artifact verification

## Story ID

W07-E03-S001

## Title

wowsociety readiness check — framework-side PROD-01..05 coordination-artifact verification

## Objective

Verify, for each of the five `requirement-inventory.md` §D product-coordination items (PROD-01..05),
that the named enabling framework capability genuinely exists and is documented as consumable by
wowsociety, and that a product upgrade path is documented. **This story is explicitly framework-side
only — it performs zero wowsociety-repository code change under any circumstance**, per mandate §2.3's
framework/product boundary.

## Value to the framework

Without this story, the five PROD-01..05 coordination items — each recorded in `requirement-
inventory.md` §D specifically because mandate §2.3 excludes them from framework implementation — would
have no closing verification step anywhere in this programme. This story closes that gap the only way
mandate §2.3 permits: by confirming the framework's own side of the coordination is genuinely ready
(the enabling capability exists, documented, consumable), while leaving the wowsociety-side action
itself entirely to wowsociety's own repository and its own maintainers.

## Problem statement

`requirement-inventory.md` §D, quoted in full:

- "PROD-01 | wowsociety `policy_override` composite FK | Product schema fix | DATA-01 T1 (parent unique
  index) + DATA-09 protocol"
- "PROD-02 | wowsociety `kernel/mfa` import migration (5 identity files) | Product code migration |
  FBL-01 re-home ships deprecated forwarding shim"
- "PROD-03 | wowsociety readiness/timeout backports to committed main.go | Product hand-edit | DX-07 T1
  + FBL-09 fix the templates"
- "PROD-04 | SEC-01 impersonation cutover (whoami/impersonation/tests) | Product auth flow rework |
  SEC-01 T1/T5 grant contract + coordinated rollout plan"
- "PROD-05 | DATA-08 W6 staging audit re-verification before version bump | Product compliance drill |
  hash_version branch verification (D-04)"

Each row's own "Enabling framework capability" column names a wowapi-side deliverable this programme has
already built and accepted in an earlier wave. This story verifies each of those five deliverables is
genuinely present and documented, not merely claimed complete by its originating wave's own closure
record.

## Source requirements

None directly source-cited (this story verifies existing deliverables). PROD-01..05 themselves.

## Current-state assessment

Direct verification at revision `733ef3e930cbb3f89f5bbc53d8f562c60e426513` supersedes the
originating waves' closure claims for this coordination decision. DATA-09's protocol, FBL-01's
forwarding shim, DX-07/FBL-09's generated readiness and timeout wiring, SEC-01's grant table/resolver,
and D-04's versioned audit verification are present and pass focused tests. Two prerequisites are not
ready: `rule_versions` lacks DATA-01's required `UNIQUE (tenant_id, id)` parent key, and
W03-E01-S004's SEC-01 rollout documents contradict the current grant schema and resolver authority
model. The exact findings and consumer paths are recorded in `ART-W07-E03-S001-001`.

## Desired state

For each of PROD-01..05: a documented coordination artifact exists confirming (a) the enabling framework
capability is genuinely present (re-verified directly, not merely cited from the originating wave's own
closure claim), and (b) a documented product-side upgrade path exists describing what wowsociety would
need to do to consume the capability — without this story itself performing any part of that consumption.

## Scope

- Re-verify DATA-01 T1 + DATA-09's protocol exist, for PROD-01.
- Re-verify FBL-01's deprecated forwarding shim exists, for PROD-02.
- Re-verify DX-07 T1 + FBL-09's template fixes exist, for PROD-03.
- Re-verify SEC-01 T1/T5's grant contract exists and document (or confirm existing documentation of) a
  coordinated rollout plan, for PROD-04.
- Re-verify D-04's `hash_version` branch verification exists, for PROD-05.
- Produce one consolidated coordination-artifact record covering all five.

## Out of scope

- **Any wowsociety-repository code change** — explicitly and absolutely out of scope, per mandate §2.3.
- **Performing the wowsociety-side action itself** (e.g. actually migrating wowsociety's 5 identity
  files for PROD-02, actually running wowsociety's own staging audit re-verification for PROD-05) — these
  remain wowsociety-repository actions, this story only confirms the framework-side prerequisite is
  ready and documents the path.
- **PROD-04's own "coordinated rollout plan"** in the sense of actually scheduling or executing a
  cross-repo cutover — this story confirms such a plan is documented (per SEC-01's own W03-E01-S004
  cross-repo-cutover-plan story, if that story produced one) or documents the gap if it does not yet
  exist, it does not itself author a net-new rollout plan beyond what SEC-01's own story already
  produced.

## Assumptions

- SEC-01's own W03-E01-S004 (cross-repo-cutover-plan, per `impl/analysis/wave-allocation-detail.md`'s
  own W03-E01 grouping: "S004 cross-repo-cutover-plan (coordination artifact for PROD-04: sequencing,
  staging validation, rollback — documentation/verification story, no product code)") is assumed to have
  already produced the PROD-04 coordination artifact this story consumes — this story's own PROD-04
  verification task re-confirms that artifact exists and is current, it does not author a new one from
  scratch unless W03-E01-S004's own artifact is found missing or stale.
- The exact format of the "documented coordination artifact" this story produces for each PROD-0N item
  (a single consolidated document, or five separate ones) is not specified by any source document — this
  story's own implementation determines the format, favoring a single consolidated record per the
  story's own "Required artifacts" framing below.

## Dependencies

Depends cross-wave on W02-E01/E02 (DATA-09, DATA-01), W05-E05 (FBL-01), W04-E04 (DX-07 T1, D-04),
W01-E03 (FBL-09), W03-E01 (SEC-01). Although those earlier records reached their own accepted
states, this story's required direct re-verification found the DATA-01 parent-key and SEC-01 rollout-
artifact gaps above. Those findings block this story rather than being waived by the entry gate.

## Affected packages or components

None — this is a documentation/verification story with zero code change of any kind, in either
repository.

## Compatibility considerations

Not applicable.

## Security considerations

PROD-04's own verification (SEC-01's impersonation cutover) is the one item in this set with genuine
security stakes — this story's own re-verification of SEC-01 T1/T5's grant contract and the coordinated
rollout plan's existence is itself a security-relevant confirmation, though the story performs no
security-relevant code change of its own.

## Performance considerations

Not applicable.

## Observability considerations

Not applicable.

## Migration considerations

Not applicable — no schema or data migration is performed by this story (PROD-01's own composite FK and
PROD-05's own `hash_version` verification are both wowsociety-side actions this story documents the
readiness for, not performs).

## Documentation requirements

This story's entire output is documentation: the consolidated PROD-01..05 coordination-artifact record.

## Acceptance criteria

- **AC-W07-E03-S001-01**: PROD-01's coordination artifact confirms DATA-01 T1's `UNIQUE(tenant_id, id)` on
  `rule_versions` and DATA-09's online-migration protocol both genuinely exist (re-verified directly),
  and documents the product upgrade path for wowsociety's own `policy_override` composite FK.
- **AC-W07-E03-S001-02**: PROD-02's coordination artifact confirms FBL-01's deprecated forwarding shim at
  `kernel/mfa` genuinely exists, and documents the product upgrade path for wowsociety's own 5-file
  import migration.
- **AC-W07-E03-S001-03**: PROD-03's coordination artifact confirms DX-07 T1's migration-currency readiness check
  and FBL-09's server-timeout template fixes genuinely exist, and documents the product backport path for
  wowsociety's own committed `main.go`.
- **AC-W07-E03-S001-04**: PROD-04's coordination artifact confirms SEC-01 T1/T5's grant contract genuinely
  exists and that a coordinated rollout plan is documented (consuming W03-E01-S004's own artifact if it
  exists, or documenting the gap if it does not).
- **AC-W07-E03-S001-05**: PROD-05's coordination artifact confirms D-04's `hash_version` branch-verification
  logic genuinely exists, and documents the product staging-drill re-verification path. Zero
  wowsociety-repository code change is performed anywhere in this story's own execution.

## Required artifacts

- A consolidated PROD-01..05 coordination-artifact record.
See `artifacts/index.md`.

## Required evidence

- Direct re-verification output for each of the five enabling framework capabilities (not merely a
  citation of the originating wave's own closure claim).
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`: `story.md`
and `plan.md` complete, all five acceptance criteria numbered and measurable, cross-wave dependencies
recorded, owner/reviewer assignment pending.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; all five acceptance
criteria verified with evidence in `evidence/index.md`; `closure.md` completed; independent review
passed per mandate §14, specifically confirming zero wowsociety-repository code change occurred anywhere
in this story's own execution, and that each of the five re-verifications was genuinely performed
directly, not merely copied from the originating wave's own closure claim.

## Risks

RISK-W07-E03-001 (a documentation gap found in one of the five enabling capabilities' own consumability
documentation) — see epic-level `risks.md` for full detail and mitigation/contingency.

## Residual-risk expectations

Once all five acceptance criteria are verified and the zero-wowsociety-code-change constraint is
confirmed honored, residual risk is expected to be negligible.

## Plan

See `plan.md`.
